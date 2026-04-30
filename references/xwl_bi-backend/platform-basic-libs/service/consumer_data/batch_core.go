package consumer_data

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/1340691923/xwl_bi/engine/logs"
	"github.com/1340691923/xwl_bi/model"
	"github.com/1340691923/xwl_bi/platform-basic-libs/util"
	"go.uber.org/zap"
)

type batchFlushFunc[T any] func(batch []T) (remaining []T, err error)
type batchBeforeFlushFunc func(bufferLength int) error

type AsyncExecutor interface {
	Submit(task func()) error
}

// batchCore 抽象了 sinker 里几类批量器的公共行为：
// 1. 统一维护 buffer / swap / restore。
// 2. 统一处理并发 Flush 互斥。
// 3. 统一处理“按大小触发”和“按定时器触发”两种刷盘方式。
// 4. 统一处理失败回滚，避免未落库数据直接丢失。
//
// 示例：
// 1. 当前批次有 [A, B, C]
// 2. Flush 时先 swapBuffer，把 [A, B, C] 摘出来
// 3. 如果 flushFn 失败，就把失败批次 restore 回 buffer 头部
// 4. 后续 Add 进来的新数据会排在它们后面，而不是覆盖它们
type batchCore[T any] struct {
	buffer        []T
	bufferMutex   *sync.RWMutex
	flushMutex    *sync.Mutex
	batchSize     int
	flushInterval int
	flushName     string
	flushFn       batchFlushFunc[T]
	beforeFlush   batchBeforeFlushFunc
	traceDebug    bool
	asyncExecutor AsyncExecutor
	flushQueued   uint32
}

// newBatchCore 创建一个最小可用的批量器基座。
//
// 这个构造函数不关心具体业务类型，只关心：
// 1. batch 大小
// 2. 定时器间隔
// 3. 真正的刷盘实现 flushFn
func newBatchCore[T any](config model.BatchConfig, flushName string, flushFn batchFlushFunc[T]) *batchCore[T] {
	config = config.Normalize()
	return &batchCore[T]{
		buffer:        make([]T, 0, config.BufferSize),
		bufferMutex:   new(sync.RWMutex),
		flushMutex:    new(sync.Mutex),
		batchSize:     config.BufferSize,
		flushInterval: config.FlushInterval,
		flushName:     flushName,
		flushFn:       flushFn,
	}
}

func (t *batchCore[T]) enableTraceDebug() {
	t.traceDebug = true
}

func (t *batchCore[T]) setBeforeFlush(fn batchBeforeFlushFunc) {
	t.beforeFlush = fn
}

func (t *batchCore[T]) setAsyncExecutor(executor AsyncExecutor) {
	t.asyncExecutor = executor
}

// swapBuffer 把当前 buffer 整体摘出来交给 Flush，在锁外执行慢操作。
//
// 这么做的原因是：
// 1. Add 仍然可以继续接收新数据。
// 2. 数据库事务、网络 I/O 不会长期占着 bufferMutex。
func (t *batchCore[T]) swapBuffer() []T {
	t.bufferMutex.Lock()
	defer t.bufferMutex.Unlock()

	if len(t.buffer) == 0 {
		return nil
	}

	batch := append([]T(nil), t.buffer...)
	t.buffer = make([]T, 0, t.batchSize)
	return batch
}

// restoreBuffer 用于 Flush 失败后的回滚。
//
// 如果这一批数据没真正落库，就必须把它们放回缓冲区头部，
// 避免出现“调用方以为只是失败，实际数据已经丢了”的情况。
func (t *batchCore[T]) restoreBuffer(batch []T) {
	if len(batch) == 0 {
		return
	}

	t.bufferMutex.Lock()
	defer t.bufferMutex.Unlock()

	restored := make([]T, 0, len(batch)+len(t.buffer))
	restored = append(restored, batch...)
	t.buffer = append(restored, t.buffer...)
}

// Add 只负责两件事：
// 1. 把一条数据追加进 buffer。
// 2. 当 buffer 达到 batchSize 时触发主动 Flush。
func (t *batchCore[T]) Add(data T) (err error) {
	begin := time.Now()

	t.bufferMutex.Lock()
	t.buffer = append(t.buffer, data)
	bufferLength := len(t.buffer)
	t.bufferMutex.Unlock()

	if bufferLength >= t.batchSize {
		err = t.requestFlush()
		if t.traceDebug && time.Since(begin) >= 100*time.Millisecond {
			logs.Logger.Debug(
				"batch add slow",
				zap.String("flush_name", t.flushName),
				zap.Duration("cost", time.Since(begin)),
				zap.Int("buffer_length", bufferLength),
				zap.Int("batch_size", t.batchSize),
				zap.Bool("triggered_flush", true),
				zap.Error(err),
			)
		}
		return err
	}

	if t.traceDebug && time.Since(begin) >= 100*time.Millisecond {
		logs.Logger.Debug(
			"batch add slow",
			zap.String("flush_name", t.flushName),
			zap.Duration("cost", time.Since(begin)),
			zap.Int("buffer_length", bufferLength),
			zap.Int("batch_size", t.batchSize),
			zap.Bool("triggered_flush", false),
		)
	}
	return nil
}

func (t *batchCore[T]) requestFlush() error {
	if t.asyncExecutor == nil {
		return t.Flush()
	}

	// flushQueued 用来做一个非常小的“去重阀门”：
	// 1. 多个 Add 同时命中 batchSize 时，只排一个 flush 任务
	// 2. 当前 flush 结束后，如果 buffer 里又重新堆满，再补发下一轮 flush
	//
	// 这样能避免因为高并发命中边界，把大量重复 flush 任务塞满 reportPool。
	if !atomic.CompareAndSwapUint32(&t.flushQueued, 0, 1) {
		return nil
	}

	if err := t.asyncExecutor.Submit(func() {
		defer func() {
			atomic.StoreUint32(&t.flushQueued, 0)
			if t.getBufferLength() >= t.batchSize {
				_ = t.requestFlush()
			}
		}()

		for {
			if err := t.Flush(); err != nil {
				if IsDeferredFlushError(err) {
					return
				}
				logs.Logger.Error(t.flushName, zap.Error(err))
				return
			}

			if t.getBufferLength() < t.batchSize {
				return
			}
		}
	}); err != nil {
		atomic.StoreUint32(&t.flushQueued, 0)
		return err
	}

	return nil
}

// Flush 负责执行一次完整刷盘。
//
// 这里的关键点是：
// 1. 先用 flushMutex 避免并发 Flush。
// 2. 再 swapBuffer，把批次摘出来。
// 3. 把真正的 I/O 委托给 flushFn。
// 4. 如果 flushFn 返回错误，就把 remaining 恢复回去。
func (t *batchCore[T]) Flush() (err error) {
	waitBegin := time.Now()
	t.flushMutex.Lock()
	defer t.flushMutex.Unlock()
	waitCost := time.Since(waitBegin)

	if t.traceDebug && waitCost >= 100*time.Millisecond {
		logs.Logger.Debug(
			"batch flush waited for lock",
			zap.String("flush_name", t.flushName),
			zap.Duration("wait_cost", waitCost),
			zap.Int("buffer_length", t.getBufferLength()),
		)
	}

	if t.beforeFlush != nil {
		if err := t.beforeFlush(t.getBufferLength()); err != nil {
			util.RecordPersistenceError("flush_deferred", err)
			return err
		}
	}

	batch := t.swapBuffer()
	if len(batch) == 0 {
		return nil
	}

	flushBegin := time.Now()
	remaining, err := t.flushFn(batch)
	flushCost := time.Since(flushBegin)
	if t.traceDebug && (flushCost >= 100*time.Millisecond || err != nil) {
		logs.Logger.Debug(
			"batch flush timing",
			zap.String("flush_name", t.flushName),
			zap.Duration("wait_cost", waitCost),
			zap.Duration("flush_cost", flushCost),
			zap.Int("batch_size", len(batch)),
			zap.Int("remaining_size", len(remaining)),
			zap.Error(err),
		)
	}
	if err != nil {
		if len(remaining) == 0 {
			remaining = batch
		}
		t.restoreBuffer(remaining)
	}

	return err
}

func (t *batchCore[T]) getBufferLength() int {
	t.bufferMutex.RLock()
	defer t.bufferMutex.RUnlock()
	return len(t.buffer)
}

// FlushAll 常用于进程退出阶段。
// 它会一直尝试刷到 buffer 为空，尽量减少退出时的残留数据。
func (t *batchCore[T]) FlushAll() error {
	for t.getBufferLength() > 0 {
		if err := t.Flush(); err != nil {
			return err
		}
	}
	return nil
}

// RegularFlushing 是兜底定时器。
//
// 即使流量很低、始终达不到 batchSize，
// 也能依靠定时器把零散数据稳定刷出去。
func (t *batchCore[T]) RegularFlushing() {
	go func() {
		ticker := time.NewTicker(time.Duration(t.flushInterval) * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			if err := t.requestFlush(); err != nil {
				if IsDeferredFlushError(err) {
					continue
				}
				logs.Logger.Error(t.flushName, zap.Error(err))
			}
		}
	}()
}
