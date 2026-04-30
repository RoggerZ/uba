package runner

import (
	"testing"

	"github.com/1340691923/xwl_bi/platform-basic-libs/util"
)

func TestDiffWorkerPoolStats(t *testing.T) {
	prev := util.WorkerPoolStats{
		GoroutinesTotal: 10,
		Running:         1,
		Idle:            3,
		Capacity:        4,
		Queued:          0,
		SubmittedTotal:  2,
		CompletedTotal:  2,
		RejectedTotal:   0,
	}
	current := util.WorkerPoolStats{
		GoroutinesTotal: 11,
		Running:         2,
		Idle:            2,
		Capacity:        4,
		Queued:          8,
		SubmittedTotal:  10,
		CompletedTotal:  4,
		RejectedTotal:   1,
	}

	fields := diffWorkerPoolStats(prev, current)
	if len(fields) == 0 {
		t.Fatal("expected changed fields, got none")
	}
}

func TestShouldLogWorkerPoolPressure(t *testing.T) {
	if shouldLogWorkerPoolPressure(util.WorkerPoolStats{Queued: 0, BusyRatio: 0.95}) {
		t.Fatal("queued=0 should not trigger pressure log")
	}
	if shouldLogWorkerPoolPressure(util.WorkerPoolStats{Queued: 3, BusyRatio: 0.5}) {
		t.Fatal("busy_ratio < 0.8 should not trigger pressure log")
	}
	if !shouldLogWorkerPoolPressure(util.WorkerPoolStats{Queued: 3, BusyRatio: 0.9}) {
		t.Fatal("expected pressure log when queued>0 and busy_ratio>=0.8")
	}
}
