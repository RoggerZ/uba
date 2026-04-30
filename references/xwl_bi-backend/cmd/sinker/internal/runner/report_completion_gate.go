package runner

import (
	"sync"
	"sync/atomic"
)

type reportCompletionGate struct {
	offset         int64
	committer      *partitionOrderedCommitter
	tracker        *reportCompletionGateTracker
	remainingTasks int32
	completedOnce  sync.Once
}

// reportCompletionGate 负责把“单条消息”和“这条消息自己关联的异步持久化任务”重新绑定起来。
//
// 这次修复里最重要的边界是：
// 1. 不是等全局所有 batch 都结束
// 2. 也不是等整个 worker pool 空了
// 3. 而是只等“这一条消息自己挂上的任务”全部结束
// 4. 然后只对这一条消息执行一次 `committer.Complete(offset)`
//
// 这个 gate 不负责分区内顺序，它只负责回答一个问题：
// “这条消息本身，现在是不是已经可以去参与顺序提交了？”
//
// 示例一：成功消息
// 1. offset=100 会登记两个异步任务：
//   - success status 入批
//   - metric 明细入批
//
// 2. 因此先调用两次 `AddTask()`，remainingTasks=2
// 3. status 对应的 batch flush 完成，调用一次 `TaskDone()`，remainingTasks=1
// 4. metric 对应的 batch flush 完成，再调用一次 `TaskDone()`，remainingTasks=0
// 5. 此时 gate 会触发一次 `committer.Complete(100)`
//
// 示例二：只写状态的丢弃消息
// 1. offset=101 缺少 distinct_id，只登记一个 fail status 任务
// 2. 只调用一次 `AddTask()`
// 3. status batch flush 完成后调用一次 `TaskDone()`
// 4. 这时就可以直接 `Complete(101)`
//
// 示例三：没有异步任务的路径
// 1. 某条消息提前返回，但没有真正挂上任何异步持久化任务
// 2. callback defer 阶段调用 `NoAsyncTaskCompleteNow()`
// 3. 因为 remainingTasks 仍然是 0，所以直接 `Complete(offset)`
func newReportCompletionGate(offset int64, committer *partitionOrderedCommitter, tracker *reportCompletionGateTracker) *reportCompletionGate {
	gate := &reportCompletionGate{
		offset:    offset,
		committer: committer,
		tracker:   tracker,
	}
	if tracker != nil {
		tracker.TrackGateCreated()
	}
	return gate
}

// AddTask 表示“当前这条消息又登记了一个新的异步任务”。
//
// 它只增加当前消息自己的计数，不影响其他消息。
//
// 示例：
// 1. 先登记 status 任务 -> remainingTasks=1
// 2. 再登记 metric 任务 -> remainingTasks=2
func (g *reportCompletionGate) AddTask() {
	if g == nil {
		return
	}
	atomic.AddInt32(&g.remainingTasks, 1)
	g.tracker.TrackTaskAdded()
}

// TaskDone 表示“当前这条消息关联的某一个异步任务已经结束”。
//
// 注意这里说的是“这一条消息自己的某个任务结束”，
// 不是“全局 flush 线程空闲”，也不是“这一轮所有 batch 都结束”。
//
// 只有 remainingTasks 从正数减到 0 的那一次，才会真正触发完成。
//
// 示例：
// 1. remainingTasks=2
// 2. status batch 完成 -> `TaskDone()` 后 remainingTasks=1
// 3. metric batch 完成 -> `TaskDone()` 后 remainingTasks=0
// 4. 第 3 步才会触发 `Complete(offset)`
func (g *reportCompletionGate) TaskDone() {
	if g == nil {
		return
	}

	if atomic.AddInt32(&g.remainingTasks, -1) == 0 {
		g.tracker.TrackTaskDone()
		g.complete()
		return
	}
	g.tracker.TrackTaskDone()
}

// NoAsyncTaskCompleteNow 专门处理“这条消息没有登记任何异步任务”的场景。
//
// 这类路径仍然保持当前语义：
// callback 尾部发现 remainingTasks 还是 0 时，允许立刻完成。
//
// 示例：
// 1. 某条消息走了一条不需要真正入批的路径
// 2. 整个 handler 过程中都没有调用 `AddTask()`
// 3. callback defer 阶段调用 `NoAsyncTaskCompleteNow()`
// 4. 这时会直接触发一次 `Complete(offset)`
func (g *reportCompletionGate) NoAsyncTaskCompleteNow() {
	if g == nil {
		return
	}

	if atomic.LoadInt32(&g.remainingTasks) == 0 {
		g.complete()
	}
}

// complete 负责把“最多只完成一次”这个约束集中收口。
//
// 即使出现下面这种并发场景，也只会推进一次：
// 1. 多个异步任务几乎同时结束
// 2. callback defer 又额外调用了 `NoAsyncTaskCompleteNow()`
//
// `sync.Once` 保证同一条消息最多只会触发一次 `committer.Complete(offset)`。
func (g *reportCompletionGate) complete() {
	g.completedOnce.Do(func() {
		g.tracker.TrackCompleted()
		g.committer.Complete(g.offset)
	})
}
