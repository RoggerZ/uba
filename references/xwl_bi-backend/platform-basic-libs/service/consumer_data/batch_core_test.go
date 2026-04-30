package consumer_data

import (
	"sync/atomic"
	"testing"

	"github.com/1340691923/xwl_bi/model"
)

type fakeAsyncExecutor struct {
	tasks []func()
}

func (f *fakeAsyncExecutor) Submit(task func()) error {
	f.tasks = append(f.tasks, task)
	return nil
}

func TestBatchCoreAddAtBatchSizeRequestsAsyncFlush(t *testing.T) {
	var flushCount int32
	executor := &fakeAsyncExecutor{}
	core := newBatchCore[int](model.BatchConfig{BufferSize: 2, FlushInterval: 0}, "test-batch", func(batch []int) ([]int, error) {
		atomic.AddInt32(&flushCount, 1)
		return nil, nil
	})
	core.setAsyncExecutor(executor)

	if err := core.Add(1); err != nil {
		t.Fatalf("Add(1) returned error: %v", err)
	}
	if got := atomic.LoadInt32(&flushCount); got != 0 {
		t.Fatalf("flush count after first add = %d, want 0", got)
	}

	if err := core.Add(2); err != nil {
		t.Fatalf("Add(2) returned error: %v", err)
	}
	if got := atomic.LoadInt32(&flushCount); got != 0 {
		t.Fatalf("flush should not run synchronously, got %d", got)
	}
	if len(executor.tasks) != 1 {
		t.Fatalf("queued async tasks = %d, want 1", len(executor.tasks))
	}

	executor.tasks[0]()

	if got := atomic.LoadInt32(&flushCount); got != 1 {
		t.Fatalf("flush count after async task = %d, want 1", got)
	}
	if got := core.getBufferLength(); got != 0 {
		t.Fatalf("buffer length after async flush = %d, want 0", got)
	}
}
