package runner

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestClickHousePartsMonitorStartsOnWriteAndStopsAfterIdle(t *testing.T) {
	stop := make(chan struct{})
	defer close(stop)

	var sampleCount int32
	monitor := newClickHousePartsMonitorWithOptions(
		clickHouseInsertPressureSettings{},
		nil,
		stop,
		20*time.Millisecond,
		40*time.Millisecond,
		func() {
			atomic.AddInt32(&sampleCount, 1)
		},
	)

	monitor.NotifyWrite("xwl_real_time_warehousing")
	time.Sleep(35 * time.Millisecond)

	if atomic.LoadInt32(&sampleCount) == 0 {
		t.Fatal("expected monitor to sample after write activity")
	}

	time.Sleep(90 * time.Millisecond)

	monitor.mutex.Lock()
	running := monitor.running
	monitor.mutex.Unlock()
	if running {
		t.Fatal("expected monitor to stop after idle timeout")
	}
}

func TestClickHousePartsMonitorSparseWritesDoNotCauseImmediateStop(t *testing.T) {
	stop := make(chan struct{})
	defer close(stop)

	var sampleCount int32
	monitor := newClickHousePartsMonitorWithOptions(
		clickHouseInsertPressureSettings{},
		nil,
		stop,
		20*time.Millisecond,
		80*time.Millisecond,
		func() {
			atomic.AddInt32(&sampleCount, 1)
		},
	)

	monitor.NotifyWrite("xwl_acceptance_status")
	time.Sleep(40 * time.Millisecond)
	monitor.NotifyWrite("xwl_acceptance_status")
	time.Sleep(40 * time.Millisecond)

	monitor.mutex.Lock()
	running := monitor.running
	monitor.mutex.Unlock()
	if !running {
		t.Fatal("expected monitor to stay running while sparse writes continue")
	}
}
