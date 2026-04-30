package util

import (
	"errors"
	"log"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/panjf2000/ants/v2"
)

var (
	ErrWorkerPoolClosed = errors.New("worker pool is closed")
)

type DynamicWorkerPoolConfig struct {
	Name         string
	MinWorkers   int
	MaxWorkers   int
	QueueSize    int
	TuneInterval time.Duration
	DrainTimeout time.Duration
}

type WorkerPoolStats struct {
	GoroutinesTotal int
	Queued          int64
	QueueCapacity   int
	QueueUsageRatio float64
	Running         int
	Capacity        int
	Idle            int
	BusyRatio       float64
	MinWorkers      int
	MaxWorkers      int
	SubmittedTotal  int64
	CompletedTotal  int64
	RejectedTotal   int64
	Closed          bool
}

// workerPoolTask 把任务执行所需的最小诊断信息收口到协程池内部。
//
// 这样无论任务最终有没有真正被 ants worker 执行，我们都能从同一个地方输出：
// 1. 提交来源
// 2. 进入 ants 执行的时刻
// 3. 执行结束或 panic
type workerPoolTask struct {
	id          int64
	fn          func()
	submittedAt time.Time
	callerFile  string
	callerLine  int
}

// DynamicWorkerPool 是一个面向当前项目的轻量协程池封装。
//
// 它负责：
// 1. 有界排队
// 2. 动态调节 ants 容量
// 3. 暴露池统计
// 4. 把任务真正开始执行/结束执行的日志统一打在池内部
type DynamicWorkerPool struct {
	funcPool *ants.PoolWithFuncGeneric[*workerPoolTask]

	taskQueue chan *workerPoolTask
	stopCh    chan struct{}
	name      string

	minWorkers   int
	maxWorkers   int
	runtimeMin   int
	runtimeMax   int
	tuneInterval time.Duration
	drainTimeout time.Duration

	boundsMutex sync.RWMutex
	queued         int64
	submittedTotal int64
	completedTotal int64
	rejectedTotal  int64
	closed         uint32
	taskSeq        int64

	wg sync.WaitGroup
}

func (c DynamicWorkerPoolConfig) Normalize() DynamicWorkerPoolConfig {
	defaultWorkers := maxInt(1, runtime.GOMAXPROCS(0))
	if c.MinWorkers <= 0 {
		c.MinWorkers = defaultWorkers
	}
	if c.MaxWorkers <= 0 {
		c.MaxWorkers = minInt(64, defaultWorkers*4)
	}
	if c.MaxWorkers < c.MinWorkers {
		c.MaxWorkers = c.MinWorkers
	}
	if c.QueueSize <= 0 {
		c.QueueSize = 4096
	}
	if c.TuneInterval <= 0 {
		c.TuneInterval = 2 * time.Second
	}
	if c.DrainTimeout <= 0 {
		c.DrainTimeout = 30 * time.Second
	}
	return c
}

func NewDynamicWorkerPool(cfg DynamicWorkerPoolConfig) (*DynamicWorkerPool, error) {
	cfg = cfg.Normalize()
	wp := &DynamicWorkerPool{
		taskQueue:    make(chan *workerPoolTask, cfg.QueueSize),
		stopCh:       make(chan struct{}),
		name:         cfg.Name,
		minWorkers:   cfg.MinWorkers,
		maxWorkers:   cfg.MaxWorkers,
		runtimeMin:   cfg.MinWorkers,
		runtimeMax:   cfg.MaxWorkers,
		tuneInterval: cfg.TuneInterval,
		drainTimeout: cfg.DrainTimeout,
	}
	funcPool, err := ants.NewPoolWithFuncGeneric[*workerPoolTask](cfg.MinWorkers, wp.runTask)
	if err != nil {
		return nil, err
	}
	wp.funcPool = funcPool

	if IsSinkerDiagnosticLogEnabled() {
		log.Printf("worker pool implementation selected pool_name=%s impl=funcgeneric", wp.name)
	}

	wp.wg.Add(2)
	go wp.dispatchLoop()
	go wp.tuneLoop()
	return wp, nil
}

// Submit 只负责把任务放入有界队列。
//
// 真正是否进入 ants worker、何时开始执行，由 dispatchLoop + runTask 继续记录。
func (p *DynamicWorkerPool) Submit(task func()) error {
	if atomic.LoadUint32(&p.closed) == 1 {
		atomic.AddInt64(&p.rejectedTotal, 1)
		return ErrWorkerPoolClosed
	}

	taskID := atomic.AddInt64(&p.taskSeq, 1)
	callerFile, callerLine := workerPoolCaller()
	envelope := &workerPoolTask{
		id:          taskID,
		fn:          task,
		submittedAt: time.Now(),
		callerFile:  callerFile,
		callerLine:  callerLine,
	}

	atomic.AddInt64(&p.queued, 1)
	select {
	case p.taskQueue <- envelope:
		atomic.AddInt64(&p.submittedTotal, 1)
		if IsSinkerDiagnosticLogEnabled() && p.shouldTraceTask(taskID) {
			log.Printf("worker pool task enqueued %s", p.traceTaskFields(envelope))
		}
		return nil
	case <-p.stopCh:
		atomic.AddInt64(&p.queued, -1)
		atomic.AddInt64(&p.rejectedTotal, 1)
		return ErrWorkerPoolClosed
	}
}

func (p *DynamicWorkerPool) Stats() WorkerPoolStats {
	running := p.running()
	capacity := p.capacity()
	idle := maxInt(0, capacity-running)
	queued := atomic.LoadInt64(&p.queued)
	queueCapacity := cap(p.taskQueue)

	var queueUsageRatio float64
	if queueCapacity > 0 {
		queueUsageRatio = float64(queued) / float64(queueCapacity)
	}

	var busyRatio float64
	if capacity > 0 {
		busyRatio = float64(running) / float64(capacity)
	}

	return WorkerPoolStats{
		GoroutinesTotal: runtime.NumGoroutine(),
		Queued:          queued,
		QueueCapacity:   queueCapacity,
		QueueUsageRatio: queueUsageRatio,
		Running:         running,
		Capacity:        capacity,
		Idle:            idle,
		BusyRatio:       busyRatio,
		MinWorkers:      p.runtimeMinWorkers(),
		MaxWorkers:      p.runtimeMaxWorkers(),
		SubmittedTotal:  atomic.LoadInt64(&p.submittedTotal),
		CompletedTotal:  atomic.LoadInt64(&p.completedTotal),
		RejectedTotal:   atomic.LoadInt64(&p.rejectedTotal),
		Closed:          atomic.LoadUint32(&p.closed) == 1,
	}
}

func (p *DynamicWorkerPool) Name() string {
	return p.name
}

func (p *DynamicWorkerPool) Close() error {
	if !atomic.CompareAndSwapUint32(&p.closed, 0, 1) {
		return nil
	}

	close(p.stopCh)
	close(p.taskQueue)
	p.wg.Wait()
	return p.releaseTimeout(p.drainTimeout)
}

func (p *DynamicWorkerPool) dispatchLoop() {
	defer p.wg.Done()

	for task := range p.taskQueue {
		submitErr := p.invokeTask(task)
		if submitErr != nil {
			atomic.AddInt64(&p.queued, -1)
			atomic.AddInt64(&p.rejectedTotal, 1)
			if IsSinkerDiagnosticLogEnabled() {
				log.Printf("worker pool ants invoke failed %s err=%v", p.traceTaskFields(task), submitErr)
			}
		}
	}
}

func (p *DynamicWorkerPool) runTask(task *workerPoolTask) {
	atomic.AddInt64(&p.queued, -1)
	startedAt := time.Now()
	if IsSinkerDiagnosticLogEnabled() && p.shouldTraceTask(task.id) {
		log.Printf(
			"worker pool task started in ants worker %s queue_wait_cost=%s",
			p.traceTaskFields(task),
			startedAt.Sub(task.submittedAt),
		)
	}

	defer func() {
		if r := recover(); r != nil {
			if IsSinkerDiagnosticLogEnabled() {
				log.Printf(
					"worker pool task panic in ants worker %s queue_wait_cost=%s run_cost=%s panic=%v stack=%s",
					p.traceTaskFields(task),
					startedAt.Sub(task.submittedAt),
					time.Since(startedAt),
					r,
					string(debug.Stack()),
				)
			}
		}

		atomic.AddInt64(&p.completedTotal, 1)
		if IsSinkerDiagnosticLogEnabled() && p.shouldTraceTask(task.id) {
			log.Printf(
				"worker pool task finished in ants worker %s queue_wait_cost=%s run_cost=%s",
				p.traceTaskFields(task),
				startedAt.Sub(task.submittedAt),
				time.Since(startedAt),
			)
		}
	}()

	task.fn()
}

func (p *DynamicWorkerPool) tuneLoop() {
	defer p.wg.Done()

	ticker := time.NewTicker(p.tuneInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			p.tune()
		case <-p.stopCh:
			return
		}
	}
}

func (p *DynamicWorkerPool) tune() {
	stats := p.Stats()
	currentCap := stats.Capacity
	desiredCap := currentCap
	runtimeMin := p.runtimeMinWorkers()
	runtimeMax := p.runtimeMaxWorkers()

	switch {
	case stats.Queued > int64(currentCap) && stats.Running >= currentCap-1:
		desiredCap = minInt(runtimeMax, currentCap*2)
	case stats.Queued == 0 && currentCap > runtimeMin && stats.Running < maxInt(1, currentCap/2):
		desiredCap = maxInt(runtimeMin, currentCap/2)
	}

	if desiredCap != currentCap {
		p.tuneCapacity(desiredCap)
	}
}

func (p *DynamicWorkerPool) SetRuntimeBounds(minWorkers, maxWorkers int) {
	if minWorkers <= 0 {
		minWorkers = p.minWorkers
	}
	if maxWorkers <= 0 {
		maxWorkers = p.maxWorkers
	}
	if minWorkers < p.minWorkers {
		minWorkers = p.minWorkers
	}
	if maxWorkers > p.maxWorkers {
		maxWorkers = p.maxWorkers
	}
	if maxWorkers < minWorkers {
		maxWorkers = minWorkers
	}

	p.boundsMutex.Lock()
	p.runtimeMin = minWorkers
	p.runtimeMax = maxWorkers
	p.boundsMutex.Unlock()

	currentCap := p.capacity()
	switch {
	case currentCap < minWorkers:
		p.tuneCapacity(minWorkers)
	case currentCap > maxWorkers:
		p.tuneCapacity(maxWorkers)
	}
}

func (p *DynamicWorkerPool) ResetRuntimeBounds() {
	p.boundsMutex.Lock()
	p.runtimeMin = p.minWorkers
	p.runtimeMax = p.maxWorkers
	p.boundsMutex.Unlock()
}

func (p *DynamicWorkerPool) runtimeMinWorkers() int {
	p.boundsMutex.RLock()
	defer p.boundsMutex.RUnlock()
	return p.runtimeMin
}

func (p *DynamicWorkerPool) runtimeMaxWorkers() int {
	p.boundsMutex.RLock()
	defer p.boundsMutex.RUnlock()
	return p.runtimeMax
}

func (p *DynamicWorkerPool) shouldTraceTask(taskID int64) bool {
	if p.name != "sinker-report-consumer" {
		return false
	}
	return taskID <= 10 || taskID%1000 == 0
}

func (p *DynamicWorkerPool) traceTaskFields(task *workerPoolTask) string {
	stats := p.Stats()
	return "pool_name=" + p.name +
		" task_id=" + formatInt64(task.id) +
		" caller_file=" + task.callerFile +
		" caller_line=" + formatInt(task.callerLine) +
		" tasks_queued=" + formatInt64(stats.Queued) +
		" workers_running=" + formatInt(stats.Running) +
		" workers_capacity=" + formatInt(stats.Capacity) +
		" submitted_total=" + formatInt64(stats.SubmittedTotal) +
		" completed_total=" + formatInt64(stats.CompletedTotal)
}

func workerPoolCaller() (string, int) {
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		return "", 0
	}
	return filepath.Base(file), line
}

func formatInt(value int) string {
	return strconv.Itoa(value)
}

func formatInt64(value int64) string {
	return strconv.FormatInt(value, 10)
}

func (p *DynamicWorkerPool) invokeTask(task *workerPoolTask) error {
	return p.funcPool.Invoke(task)
}

func (p *DynamicWorkerPool) running() int {
	return p.funcPool.Running()
}

func (p *DynamicWorkerPool) capacity() int {
	return p.funcPool.Cap()
}

func (p *DynamicWorkerPool) tuneCapacity(size int) {
	p.funcPool.Tune(size)
}

func (p *DynamicWorkerPool) releaseTimeout(timeout time.Duration) error {
	return p.funcPool.ReleaseTimeout(timeout)
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
