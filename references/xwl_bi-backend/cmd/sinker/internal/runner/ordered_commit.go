package runner

import (
	"fmt"
	"sync"

	"github.com/1340691923/xwl_bi/engine/logs"
	"github.com/1340691923/xwl_bi/platform-basic-libs/util"
	"go.uber.org/zap"
)

type commitState struct {
	done         bool
	generationID int32
	markFn       func()
}

// partitionOrderedCommitter 负责单个 topic-partition 的顺序提交。
//
// 这套结构专门解决一个问题：
// Kafka 分区内消息虽然被 worker 并发处理，但 offset 提交仍然必须按原顺序推进。
//
// 这里采用的规则是：
// 1. 每条消息先 Register，注册自己的 offset 和 markFn。
// 2. 任务真正跑完以后再调用 Complete。
// 3. Complete 不区分业务成功还是失败，只表示“这条 offset 已经处理完毕，可以参与顺序推进”。
// 4. 只有从当前最小未提交 offset 开始形成连续完成区间，才会依次执行 markFn。
//
// 示例：
// 1. 先 Register(10)、Register(11)
// 2. 如果 11 先完成，Complete(11) 后不会立刻提交，因为 10 还没完成
// 3. 当 10 完成后，再 Complete(10)
// 4. 此时 10 和 11 会一起按顺序提交
type partitionOrderedCommitter struct {
	mu                   sync.Mutex
	initialized          bool
	nextOffset           int64
	states               map[int64]*commitState
	doneCount            int
	lastBlocked          int64
	lastRegisteredOffset int64
	largestRegisterGap   int64
}

// newPartitionOrderedCommitter 创建一个新的分区顺序提交器。
//
// 初始状态说明：
// 1. initialized=false，表示当前还没见到任何 offset
// 2. nextOffset 未初始化
// 3. states 为空
//
// 一旦第一条消息 Register，nextOffset 就会被设置成那条消息的 offset。
func newPartitionOrderedCommitter() *partitionOrderedCommitter {
	return &partitionOrderedCommitter{
		states: make(map[int64]*commitState),
	}
}

// Register 注册一条待顺序提交的消息。
//
// 逐步解释：
// 1. 先加锁，保证同一个分区内并发 Register 不会互相覆盖。
// 2. 如果这是第一条消息，就把 nextOffset 初始化成当前 offset。
// 3. 如果这条 offset 之前没注册过，就把它放进 states。
//
// 示例：
// 1. 第一次 Register(10)
//   - initialized 变成 true
//   - nextOffset = 10
//   - states[10] = {done:false, markFn:...}
//
// 2. 第二次 Register(11)
//   - nextOffset 仍然是 10
//   - states[11] = {done:false, markFn:...}
func (c *partitionOrderedCommitter) Register(offset int64, generationID int32, markFn func()) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.initialized {
		c.initialized = true
		c.nextOffset = offset
	}

	if _, ok := c.states[offset]; !ok {
		if len(c.states) == 0 {
			c.lastRegisteredOffset = offset
			c.largestRegisterGap = 0
		} else if offset > c.lastRegisteredOffset {
			gap := offset - c.lastRegisteredOffset
			if gap > c.largestRegisterGap {
				c.largestRegisterGap = gap
			}
			c.lastRegisteredOffset = offset
		}
		c.states[offset] = &commitState{generationID: generationID, markFn: markFn}
	}
}

// Complete 标记某条 offset 已经处理完成，并尽可能顺序推进提交。
//
// 逐步解释：
// 1. 先加锁，避免多个 worker 同时修改 states。
// 2. 找到当前 offset 对应的状态；如果根本没注册过，就直接返回。
// 3. 把这条 offset 标记为 done=true。
// 4. 从 nextOffset 开始循环检查：
//   - 如果 nextOffset 对应的消息还没完成，就停止推进
//   - 如果已经完成，就执行 markFn，删除该状态，并让 nextOffset++
//
// 5. 这样就能保证只有连续完成的 offset 才会被顺序提交。
//
// 示例：
// 1. 已注册 10 和 11，nextOffset=10
// 2. 先 Complete(11)
//   - states[11].done = true
//   - 但 nextOffset=10 还没完成，所以不提交
//
// 3. 再 Complete(10)
//   - states[10].done = true
//   - 发现 10 已完成，执行 markFn(10)，nextOffset=11
//   - 继续发现 11 已完成，执行 markFn(11)，nextOffset=12
func (c *partitionOrderedCommitter) Complete(offset int64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	state, ok := c.states[offset]
	if !ok {
		return
	}
	if !state.done {
		state.done = true
		c.doneCount++
	}

	nextOffsetBefore := c.nextOffset
	pendingCountBefore := len(c.states)
	doneCountBefore := c.doneCount
	advancedCount := 0
	firstAdvancedOffset := int64(0)
	lastAdvancedOffset := int64(0)
	firstGenerationID := int32(0)
	lastGenerationID := int32(0)

	for {
		current, ok := c.states[c.nextOffset]
		if !ok || !current.done {
			break
		}

		if advancedCount == 0 {
			firstAdvancedOffset = c.nextOffset
			firstGenerationID = current.generationID
		}
		lastAdvancedOffset = c.nextOffset
		lastGenerationID = current.generationID
		current.markFn()
		delete(c.states, c.nextOffset)
		c.doneCount--
		c.nextOffset++
		advancedCount++
	}

	if advancedCount == 0 {
		if util.IsSinkerDiagnosticLogEnabled() && offset > nextOffsetBefore && c.lastBlocked != nextOffsetBefore {
			c.lastBlocked = nextOffsetBefore
			logs.Logger.Info(
				"ordered commit blocked by earlier pending offset",
				zap.Int64("complete_offset", offset),
				zap.Int32("complete_generation_id", state.generationID),
				zap.Int64("next_offset", nextOffsetBefore),
				zap.Int("pending_count", pendingCountBefore),
				zap.Int("done_count", doneCountBefore),
				zap.Bool("executed_mark", false),
			)
		}
		return
	}

	pendingCountAfter := len(c.states)
	doneCountAfter := c.doneCount
	if pendingCountAfter == 0 {
		c.lastRegisteredOffset = 0
		c.largestRegisterGap = 0
	} else if pendingCountAfter == 1 {
		c.largestRegisterGap = 0
	}
	if util.IsSinkerDiagnosticLogEnabled() && (c.lastBlocked == nextOffsetBefore || advancedCount > 1) {
		logs.Logger.Info(
			"ordered commit advanced offsets",
			zap.Int64("complete_offset", offset),
			zap.Int32("complete_generation_id", state.generationID),
			zap.Int64("next_offset_before", nextOffsetBefore),
			zap.Int64("next_offset_after", c.nextOffset),
			zap.Int("pending_count_before", pendingCountBefore),
			zap.Int("pending_count_after", pendingCountAfter),
			zap.Int("done_count_before", doneCountBefore),
			zap.Int("done_count_after", doneCountAfter),
			zap.Int("advanced_count", advancedCount),
			zap.Int64("first_mark_offset", firstAdvancedOffset),
			zap.Int64("last_mark_offset", lastAdvancedOffset),
			zap.Int32("first_mark_generation_id", firstGenerationID),
			zap.Int32("last_mark_generation_id", lastGenerationID),
			zap.Bool("executed_mark", true),
		)
	}
	if c.lastBlocked == nextOffsetBefore {
		c.lastBlocked = 0
	}
}

type orderedCommitManager struct {
	committers sync.Map
}

func newOrderedCommitManager() *orderedCommitManager {
	return &orderedCommitManager{}
}

func (m *orderedCommitManager) Get(topic string, partition int) *partitionOrderedCommitter {
	key := fmt.Sprintf("%s:%d", topic, partition)
	if committer, ok := m.committers.Load(key); ok {
		return committer.(*partitionOrderedCommitter)
	}

	created := newPartitionOrderedCommitter()
	actual, _ := m.committers.LoadOrStore(key, created)
	return actual.(*partitionOrderedCommitter)
}

type orderedCommitSnapshot struct {
	key                 string
	initialized         bool
	nextOffset          int64
	pendingCount        int
	doneCount           int
	oldestPendingOffset int64
	largestPendingGap   int64
}

func (m *orderedCommitManager) Snapshots() []orderedCommitSnapshot {
	snapshots := make([]orderedCommitSnapshot, 0)
	m.committers.Range(func(key, value any) bool {
		snapshot := value.(*partitionOrderedCommitter).snapshot()
		snapshot.key = key.(string)
		snapshots = append(snapshots, snapshot)
		return true
	})
	return snapshots
}

func (c *partitionOrderedCommitter) snapshot() orderedCommitSnapshot {
	c.mu.Lock()
	defer c.mu.Unlock()

	var (
		oldestPendingOffset = int64(0)
		largestPendingGap   = int64(0)
	)
	if len(c.states) > 0 {
		oldestPendingOffset = c.nextOffset
		if len(c.states) > 1 {
			largestPendingGap = c.largestRegisterGap
		}
	}

	return orderedCommitSnapshot{
		initialized:         c.initialized,
		nextOffset:          c.nextOffset,
		pendingCount:        len(c.states),
		doneCount:           c.doneCount,
		oldestPendingOffset: oldestPendingOffset,
		largestPendingGap:   largestPendingGap,
	}
}
