package runner

import "testing"

func TestPartitionOrderedCommitterSequentialSuccess(t *testing.T) {
	committer := newPartitionOrderedCommitter()
	calls := []int64{}

	committer.Register(10, 1, func() { calls = append(calls, 10) })
	committer.Register(11, 1, func() { calls = append(calls, 11) })

	committer.Complete(11)
	if len(calls) != 0 {
		t.Fatalf("unexpected commits before offset 10: %+v", calls)
	}

	committer.Complete(10)
	if len(calls) != 2 || calls[0] != 10 || calls[1] != 11 {
		t.Fatalf("unexpected commit order: %+v", calls)
	}
}

func TestPartitionOrderedCommitterAllowsLaterOffsetsAfterEarlierFailureCompleted(t *testing.T) {
	committer := newPartitionOrderedCommitter()
	calls := []int64{}

	committer.Register(20, 1, func() { calls = append(calls, 20) })
	committer.Register(21, 1, func() { calls = append(calls, 21) })

	// 按当前策略，“失败”不再阻塞顺序提交；
	// 只要 offset=20 这条消息已经处理完成，就应该允许 20 和 21 一起顺序推进。
	committer.Complete(20)
	committer.Complete(21)

	if len(calls) != 2 || calls[0] != 20 || calls[1] != 21 {
		t.Fatalf("unexpected commit order: %+v", calls)
	}
}

func TestPartitionOrderedCommitterSnapshotTracksDoneCountWithoutScanRegression(t *testing.T) {
	committer := newPartitionOrderedCommitter()

	committer.Register(30, 1, func() {})
	committer.Register(31, 1, func() {})
	committer.Register(32, 1, func() {})

	committer.Complete(31)
	snapshot := committer.snapshot()
	if snapshot.pendingCount != 3 {
		t.Fatalf("pendingCount = %d, want 3", snapshot.pendingCount)
	}
	if snapshot.doneCount != 1 {
		t.Fatalf("doneCount = %d, want 1", snapshot.doneCount)
	}

	committer.Complete(30)
	snapshot = committer.snapshot()
	if snapshot.pendingCount != 1 {
		t.Fatalf("pendingCount after advance = %d, want 1", snapshot.pendingCount)
	}
	if snapshot.doneCount != 0 {
		t.Fatalf("doneCount after advance = %d, want 0", snapshot.doneCount)
	}
	if snapshot.nextOffset != 32 {
		t.Fatalf("nextOffset after advance = %d, want 32", snapshot.nextOffset)
	}
}
