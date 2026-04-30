package runner

import "sync/atomic"

type reportCompletionGateSnapshot struct {
	InFlightMessages int64 `json:"inFlightMessages"`
	WaitingTasks     int64 `json:"waitingTasks"`
	CompletedMessages int64 `json:"completedMessages"`
}

type reportCompletionGateTracker struct {
	inFlightMessages int64
	waitingTasks     int64
	completedMessages int64
}

func (t *reportCompletionGateTracker) TrackGateCreated() {
	if t == nil {
		return
	}
	atomic.AddInt64(&t.inFlightMessages, 1)
}

func (t *reportCompletionGateTracker) TrackTaskAdded() {
	if t == nil {
		return
	}
	atomic.AddInt64(&t.waitingTasks, 1)
}

func (t *reportCompletionGateTracker) TrackTaskDone() {
	if t == nil {
		return
	}
	atomic.AddInt64(&t.waitingTasks, -1)
}

func (t *reportCompletionGateTracker) TrackCompleted() {
	if t == nil {
		return
	}
	atomic.AddInt64(&t.inFlightMessages, -1)
	atomic.AddInt64(&t.completedMessages, 1)
}

func (t *reportCompletionGateTracker) Snapshot() reportCompletionGateSnapshot {
	if t == nil {
		return reportCompletionGateSnapshot{}
	}
	return reportCompletionGateSnapshot{
		InFlightMessages: atomic.LoadInt64(&t.inFlightMessages),
		WaitingTasks:     atomic.LoadInt64(&t.waitingTasks),
		CompletedMessages: atomic.LoadInt64(&t.completedMessages),
	}
}
