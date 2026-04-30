package util

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestDynamicWorkerPoolSubmitAndClose(t *testing.T) {
	pool, err := NewDynamicWorkerPool(DynamicWorkerPoolConfig{
		MinWorkers:   1,
		MaxWorkers:   2,
		QueueSize:    4,
		TuneInterval: 10 * time.Millisecond,
		DrainTimeout: 2 * time.Second,
	})
	if err != nil {
		t.Fatalf("NewDynamicWorkerPool returned error: %v", err)
	}
	if pool.funcPool == nil {
		t.Fatal("expected funcgeneric worker pool to be initialized")
	}

	var done int64
	for i := 0; i < 3; i++ {
		if err := pool.Submit(func() {
			time.Sleep(20 * time.Millisecond)
			atomic.AddInt64(&done, 1)
		}); err != nil {
			t.Fatalf("Submit returned error: %v", err)
		}
	}

	if err := pool.Close(); err != nil {
		t.Fatalf("Close returned error: %v", err)
	}

	if got := atomic.LoadInt64(&done); got != 3 {
		t.Fatalf("completed tasks = %d, want 3", got)
	}
	if err := pool.Submit(func() {}); err != ErrWorkerPoolClosed {
		t.Fatalf("submit after close = %v, want ErrWorkerPoolClosed", err)
	}

	stats := pool.Stats()
	if stats.SubmittedTotal != 3 {
		t.Fatalf("submitted_total = %d, want 3", stats.SubmittedTotal)
	}
	if stats.CompletedTotal != 3 {
		t.Fatalf("completed_total = %d, want 3", stats.CompletedTotal)
	}
	if stats.RejectedTotal != 1 {
		t.Fatalf("rejected_total = %d, want 1", stats.RejectedTotal)
	}
	if !stats.Closed {
		t.Fatal("expected pool to be marked closed")
	}
}

func TestDynamicWorkerPoolTuneWithinBounds(t *testing.T) {
	pool, err := NewDynamicWorkerPool(DynamicWorkerPoolConfig{
		MinWorkers:   1,
		MaxWorkers:   4,
		QueueSize:    8,
		TuneInterval: 10 * time.Millisecond,
		DrainTimeout: 2 * time.Second,
	})
	if err != nil {
		t.Fatalf("NewDynamicWorkerPool returned error: %v", err)
	}
	if pool.funcPool == nil {
		t.Fatal("expected funcgeneric worker pool to be initialized")
	}
	defer pool.Close()

	for i := 0; i < 4; i++ {
		if err := pool.Submit(func() {
			time.Sleep(50 * time.Millisecond)
		}); err != nil {
			t.Fatalf("Submit returned error: %v", err)
		}
	}

	time.Sleep(80 * time.Millisecond)
	stats := pool.Stats()
	if stats.Capacity < 1 || stats.Capacity > 4 {
		t.Fatalf("capacity = %d, want within [1,4]", stats.Capacity)
	}
	if stats.QueueCapacity != 8 {
		t.Fatalf("queue_capacity = %d, want 8", stats.QueueCapacity)
	}
	if stats.Running < 0 || stats.Idle < 0 {
		t.Fatalf("running/idle should not be negative, got running=%d idle=%d", stats.Running, stats.Idle)
	}
	if stats.QueueUsageRatio < 0 || stats.QueueUsageRatio > 1 {
		t.Fatalf("queue_usage_ratio = %f, want within [0,1]", stats.QueueUsageRatio)
	}
	if stats.BusyRatio < 0 || stats.BusyRatio > 1 {
		t.Fatalf("busy_ratio = %f, want within [0,1]", stats.BusyRatio)
	}
	if stats.GoroutinesTotal <= 0 {
		t.Fatalf("goroutines_total = %d, want > 0", stats.GoroutinesTotal)
	}
}
