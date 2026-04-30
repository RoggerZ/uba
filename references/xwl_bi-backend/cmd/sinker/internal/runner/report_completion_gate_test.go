package runner

import "testing"

func TestReportCompletionGateCompletesAfterAllTasksDone(t *testing.T) {
	committer := newPartitionOrderedCommitter()
	calls := 0
	committer.Register(10, 1, func() { calls++ })

	gate := newReportCompletionGate(10, committer, &reportCompletionGateTracker{})
	gate.AddTask()
	gate.AddTask()
	gate.NoAsyncTaskCompleteNow()
	if calls != 0 {
		t.Fatalf("complete before tasks done = %d, want 0", calls)
	}

	gate.TaskDone()
	if calls != 0 {
		t.Fatalf("complete after first task = %d, want 0", calls)
	}

	gate.TaskDone()
	if calls != 1 {
		t.Fatalf("complete after all tasks = %d, want 1", calls)
	}
}

func TestReportCompletionGateCompletesImmediatelyWithoutAsyncTasks(t *testing.T) {
	committer := newPartitionOrderedCommitter()
	calls := 0
	committer.Register(20, 1, func() { calls++ })

	gate := newReportCompletionGate(20, committer, &reportCompletionGateTracker{})
	gate.NoAsyncTaskCompleteNow()
	if calls != 1 {
		t.Fatalf("complete without tasks = %d, want 1", calls)
	}
}
