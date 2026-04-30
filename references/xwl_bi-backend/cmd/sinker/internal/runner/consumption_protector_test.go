package runner

import (
	"runtime"
	"testing"
	"time"

	"github.com/1340691923/xwl_bi/engine/db"
	"github.com/1340691923/xwl_bi/model"
	"github.com/1340691923/xwl_bi/platform-basic-libs/util"
)

func intPtr(v int) *int             { return &v }
func int64Ptr(v int64) *int64       { return &v }
func float64Ptr(v float64) *float64 { return &v }
func uint64Ptr(v uint64) *uint64    { return &v }
func stringPtr(v string) *string    { return &v }
func boolPtr(v bool) *bool          { return &v }

type fakeRegulatedConsumer struct {
	paused       bool
	pauseCalls   int
	resumeCalls  int
	regulatorSet bool
}

func (f *fakeRegulatedConsumer) PauseConsumption() {
	f.paused = true
	f.pauseCalls++
}

func (f *fakeRegulatedConsumer) ResumeConsumption() {
	f.paused = false
	f.resumeCalls++
}

func (f *fakeRegulatedConsumer) IsPaused() bool {
	return f.paused
}

func (f *fakeRegulatedConsumer) SetConsumeRegulator(regulator interface{ Wait() }) {
	f.regulatorSet = regulator != nil
}

type fakeConsumerSnapshotReader struct {
	snapshot consumerRateSnapshot
}

func (f *fakeConsumerSnapshotReader) Snapshot() consumerRateSnapshot {
	return f.snapshot
}

func (f *fakeConsumerSnapshotReader) SetLogIntervals(normal, diagnostic time.Duration) {}

type fakePipelineSnapshotReader struct {
	snapshot reportConsumerPipelineSnapshot
}

func (f *fakePipelineSnapshotReader) Snapshot() reportConsumerPipelineSnapshot {
	return f.snapshot
}

type fakeWorkerPoolController struct {
	stats      util.WorkerPoolStats
	setCalls   int
	resetCalls int
	lastSetMin int
	lastSetMax int
}

func (f *fakeWorkerPoolController) Stats() util.WorkerPoolStats {
	return f.stats
}

func (f *fakeWorkerPoolController) SetRuntimeBounds(minWorkers, maxWorkers int) {
	f.setCalls++
	f.lastSetMin = minWorkers
	f.lastSetMax = maxWorkers
}

func (f *fakeWorkerPoolController) ResetRuntimeBounds() {
	f.resetCalls++
}

func TestReportConsumptionProtectorEntersSoftLimitedAfterThreeSoftWindows(t *testing.T) {
	reportConsumer := &fakeRegulatedConsumer{}
	realTimeConsumer := &fakeRegulatedConsumer{}
	reportRate := &fakeConsumerSnapshotReader{}
	realTimeRate := &fakeConsumerSnapshotReader{}
	pipeline := &fakePipelineSnapshotReader{
		snapshot: reportConsumerPipelineSnapshot{
			PendingCount: 25000,
			Gate: reportCompletionGateSnapshot{
				InFlightMessages: 25000,
				WaitingTasks:     45000,
			},
		},
	}
	reportPool := &fakeWorkerPoolController{
		stats: util.WorkerPoolStats{
			MinWorkers:      4,
			MaxWorkers:      20,
			QueueUsageRatio: 0.6,
			BusyRatio:       0.95,
		},
	}
	persistPool := &fakeWorkerPoolController{
		stats: util.WorkerPoolStats{
			MinWorkers: 2,
			MaxWorkers: 8,
		},
	}

	oldHealth := getMySQLHealthState
	oldMem := readRuntimeMemStats
	oldGoroutines := currentGoroutineNum
	oldPersistence := readPersistenceErrorSnapshot
	defer func() {
		getMySQLHealthState = oldHealth
		readRuntimeMemStats = oldMem
		currentGoroutineNum = oldGoroutines
		readPersistenceErrorSnapshot = oldPersistence
	}()

	getMySQLHealthState = func(name string) (db.DBHealthState, bool) {
		return db.DBHealthState{Name: name, Status: "healthy"}, true
	}
	readRuntimeMemStats = func(stats *runtime.MemStats) {
		stats.HeapAlloc = 256 << 20
	}
	currentGoroutineNum = func() int {
		return 100
	}
	readPersistenceErrorSnapshot = func() util.PersistenceErrorSnapshot {
		return util.PersistenceErrorSnapshot{}
	}

	protector := newReportConsumptionProtector(
		model.ProtectionConfig{}.Normalize(),
		reportConsumer,
		realTimeConsumer,
		reportRate,
		realTimeRate,
		pipeline,
		reportPool,
		persistPool,
	)

	for _, speed := range []float64{1000, 900, 800, 700, 600} {
		reportRate.snapshot = consumerRateSnapshot{SpeedPerSecond: speed}
		protector.sampleAndAct()
	}

	if protector.state != protectionStateSoftLimited {
		t.Fatalf("state = %s, want %s", protector.state, protectionStateSoftLimited)
	}
	if reportPool.setCalls == 0 || persistPool.setCalls == 0 {
		t.Fatalf("expected both worker pools to be limited, got report=%d persist=%d", reportPool.setCalls, persistPool.setCalls)
	}
	if reportConsumer.pauseCalls != 0 {
		t.Fatalf("soft limited should not pause report consumer, got %d", reportConsumer.pauseCalls)
	}
}

func TestReportConsumptionProtectorEntersHardPausedAfterThreeHardWindows(t *testing.T) {
	reportConsumer := &fakeRegulatedConsumer{}
	realTimeConsumer := &fakeRegulatedConsumer{}
	reportRate := &fakeConsumerSnapshotReader{}
	realTimeRate := &fakeConsumerSnapshotReader{}
	pipeline := &fakePipelineSnapshotReader{
		snapshot: reportConsumerPipelineSnapshot{
			PendingCount: 120000,
			Gate: reportCompletionGateSnapshot{
				InFlightMessages: 120000,
				WaitingTasks:     250000,
			},
		},
	}
	reportPool := &fakeWorkerPoolController{stats: util.WorkerPoolStats{MinWorkers: 4, MaxWorkers: 20}}
	persistPool := &fakeWorkerPoolController{stats: util.WorkerPoolStats{MinWorkers: 2, MaxWorkers: 8}}

	oldHealth := getMySQLHealthState
	oldMem := readRuntimeMemStats
	oldGoroutines := currentGoroutineNum
	oldPersistence := readPersistenceErrorSnapshot
	defer func() {
		getMySQLHealthState = oldHealth
		readRuntimeMemStats = oldMem
		currentGoroutineNum = oldGoroutines
		readPersistenceErrorSnapshot = oldPersistence
	}()

	getMySQLHealthState = func(name string) (db.DBHealthState, bool) {
		return db.DBHealthState{Name: name, Status: "healthy"}, true
	}
	readRuntimeMemStats = func(stats *runtime.MemStats) {
		stats.HeapAlloc = 3 << 30
	}
	currentGoroutineNum = func() int {
		return 5000
	}
	readPersistenceErrorSnapshot = func() util.PersistenceErrorSnapshot {
		return util.PersistenceErrorSnapshot{}
	}

	protector := newReportConsumptionProtector(
		model.ProtectionConfig{}.Normalize(),
		reportConsumer,
		realTimeConsumer,
		reportRate,
		realTimeRate,
		pipeline,
		reportPool,
		persistPool,
	)

	for i := 0; i < 3; i++ {
		protector.sampleAndAct()
	}

	if protector.state != protectionStateHardPaused {
		t.Fatalf("state = %s, want %s", protector.state, protectionStateHardPaused)
	}
	if reportConsumer.pauseCalls == 0 {
		t.Fatal("expected report consumer to be paused")
	}
	if !reportConsumer.paused {
		t.Fatal("report consumer should remain paused")
	}
	if realTimeConsumer.pauseCalls != 0 {
		t.Fatalf("real time consumer should not pause on first hard transition, got %d", realTimeConsumer.pauseCalls)
	}

	protector.sampleAndAct()
	if realTimeConsumer.pauseCalls != 0 {
		t.Fatalf("real time consumer should not pause after only one sustained hard window, got %d", realTimeConsumer.pauseCalls)
	}

	for i := 0; i < 2; i++ {
		protector.sampleAndAct()
	}
	if realTimeConsumer.pauseCalls == 0 {
		t.Fatal("expected real time consumer to be paused after hard escalation")
	}
}

func TestReportConsumptionProtectorRateTrendIsSignalButNotOnlySignal(t *testing.T) {
	reportConsumer := &fakeRegulatedConsumer{}
	realTimeConsumer := &fakeRegulatedConsumer{}
	reportRate := &fakeConsumerSnapshotReader{}
	realTimeRate := &fakeConsumerSnapshotReader{}
	pipeline := &fakePipelineSnapshotReader{
		snapshot: reportConsumerPipelineSnapshot{
			PendingCount: 21000,
		},
	}
	reportPool := &fakeWorkerPoolController{
		stats: util.WorkerPoolStats{
			MinWorkers:      4,
			MaxWorkers:      20,
			QueueUsageRatio: 0.6,
			BusyRatio:       0.91,
		},
	}
	persistPool := &fakeWorkerPoolController{stats: util.WorkerPoolStats{MinWorkers: 2, MaxWorkers: 8}}

	oldHealth := getMySQLHealthState
	oldMem := readRuntimeMemStats
	oldGoroutines := currentGoroutineNum
	oldPersistence := readPersistenceErrorSnapshot
	defer func() {
		getMySQLHealthState = oldHealth
		readRuntimeMemStats = oldMem
		currentGoroutineNum = oldGoroutines
		readPersistenceErrorSnapshot = oldPersistence
	}()

	getMySQLHealthState = func(name string) (db.DBHealthState, bool) {
		return db.DBHealthState{Name: name, Status: "healthy"}, true
	}
	readRuntimeMemStats = func(stats *runtime.MemStats) {
		stats.HeapAlloc = 256 << 20
	}
	currentGoroutineNum = func() int {
		return 100
	}
	readPersistenceErrorSnapshot = func() util.PersistenceErrorSnapshot {
		return util.PersistenceErrorSnapshot{}
	}

	protector := newReportConsumptionProtector(
		model.ProtectionConfig{}.Normalize(),
		reportConsumer,
		realTimeConsumer,
		reportRate,
		realTimeRate,
		pipeline,
		reportPool,
		persistPool,
	)

	for _, speed := range []float64{3000, 2000, 1000} {
		reportRate.snapshot = consumerRateSnapshot{SpeedPerSecond: speed}
		protector.sampleAndAct()
	}

	if protector.state != protectionStateSoftLimited {
		t.Fatalf("state = %s, want %s", protector.state, protectionStateSoftLimited)
	}
}

func TestReportConsumptionProtectorRecoversStepByStep(t *testing.T) {
	reportConsumer := &fakeRegulatedConsumer{paused: true}
	realTimeConsumer := &fakeRegulatedConsumer{paused: true}
	reportRate := &fakeConsumerSnapshotReader{snapshot: consumerRateSnapshot{Lag: 0, SpeedPerSecond: 0}}
	realTimeRate := &fakeConsumerSnapshotReader{snapshot: consumerRateSnapshot{Lag: 0, SpeedPerSecond: 0}}
	pipeline := &fakePipelineSnapshotReader{
		snapshot: reportConsumerPipelineSnapshot{
			PendingCount: 100,
			Gate: reportCompletionGateSnapshot{
				InFlightMessages: 100,
				WaitingTasks:     100,
			},
		},
	}
	reportPool := &fakeWorkerPoolController{stats: util.WorkerPoolStats{MinWorkers: 4, MaxWorkers: 20, QueueUsageRatio: 0.01}}
	persistPool := &fakeWorkerPoolController{stats: util.WorkerPoolStats{MinWorkers: 2, MaxWorkers: 8, QueueUsageRatio: 0.01}}

	oldHealth := getMySQLHealthState
	oldMem := readRuntimeMemStats
	oldGoroutines := currentGoroutineNum
	oldPersistence := readPersistenceErrorSnapshot
	defer func() {
		getMySQLHealthState = oldHealth
		readRuntimeMemStats = oldMem
		currentGoroutineNum = oldGoroutines
		readPersistenceErrorSnapshot = oldPersistence
	}()

	getMySQLHealthState = func(name string) (db.DBHealthState, bool) {
		return db.DBHealthState{Name: name, Status: "healthy"}, true
	}
	readRuntimeMemStats = func(stats *runtime.MemStats) {
		stats.HeapAlloc = 128 << 20
	}
	currentGoroutineNum = func() int {
		return 100
	}
	readPersistenceErrorSnapshot = func() util.PersistenceErrorSnapshot {
		return util.PersistenceErrorSnapshot{}
	}

	protector := newReportConsumptionProtector(
		model.ProtectionConfig{}.Normalize(),
		reportConsumer,
		realTimeConsumer,
		reportRate,
		realTimeRate,
		pipeline,
		reportPool,
		persistPool,
	)
	protector.state = protectionStateHardPaused
	protector.hardHoldUntil = time.Now().Add(-time.Second)
	protector.recentSpeeds = []float64{1000, 1100, 1200}

	for i := 0; i < 3; i++ {
		protector.sampleAndAct()
	}
	if protector.state != protectionStateSoftLimited {
		t.Fatalf("state after hard recovery = %s, want %s", protector.state, protectionStateSoftLimited)
	}
	if realTimeConsumer.resumeCalls == 0 {
		t.Fatal("expected real time consumer to resume when leaving hard pause")
	}

	protector.softHoldUntil = time.Now().Add(-time.Second)
	protector.recentSpeeds = []float64{1200, 1300, 1400}
	for i := 0; i < 3; i++ {
		protector.sampleAndAct()
	}
	if protector.state != protectionStateNormal {
		t.Fatalf("state after soft recovery = %s, want %s", protector.state, protectionStateNormal)
	}
	if reportConsumer.resumeCalls == 0 {
		t.Fatal("expected report consumer to resume")
	}
	if reportPool.resetCalls == 0 || persistPool.resetCalls == 0 {
		t.Fatalf("expected pools to reset bounds, got report=%d persist=%d", reportPool.resetCalls, persistPool.resetCalls)
	}
}

func TestReportConsumptionProtectorHardPausedByPersistenceErrors(t *testing.T) {
	reportConsumer := &fakeRegulatedConsumer{}
	realTimeConsumer := &fakeRegulatedConsumer{}
	reportRate := &fakeConsumerSnapshotReader{}
	realTimeRate := &fakeConsumerSnapshotReader{}
	pipeline := &fakePipelineSnapshotReader{}
	reportPool := &fakeWorkerPoolController{stats: util.WorkerPoolStats{MinWorkers: 4, MaxWorkers: 20}}
	persistPool := &fakeWorkerPoolController{stats: util.WorkerPoolStats{MinWorkers: 2, MaxWorkers: 8}}

	oldHealth := getMySQLHealthState
	oldMem := readRuntimeMemStats
	oldGoroutines := currentGoroutineNum
	oldPersistence := readPersistenceErrorSnapshot
	defer func() {
		getMySQLHealthState = oldHealth
		readRuntimeMemStats = oldMem
		currentGoroutineNum = oldGoroutines
		readPersistenceErrorSnapshot = oldPersistence
	}()

	getMySQLHealthState = func(name string) (db.DBHealthState, bool) {
		return db.DBHealthState{Name: name, Status: "healthy"}, true
	}
	readRuntimeMemStats = func(stats *runtime.MemStats) {
		stats.HeapAlloc = 3 << 29
	}
	currentGoroutineNum = func() int {
		return 100
	}
	readPersistenceErrorSnapshot = func() util.PersistenceErrorSnapshot {
		return util.PersistenceErrorSnapshot{
			CountLastMinute: 20,
			LastClass:       "clickhouse_send_failed",
			LastError:       "dial tcp 127.0.0.1:9000: connectex",
		}
	}

	protector := newReportConsumptionProtector(
		model.ProtectionConfig{}.Normalize(),
		reportConsumer,
		realTimeConsumer,
		reportRate,
		realTimeRate,
		pipeline,
		reportPool,
		persistPool,
	)

	for i := 0; i < 3; i++ {
		protector.sampleAndAct()
	}

	if protector.state != protectionStateHardPaused {
		t.Fatalf("state = %s, want %s", protector.state, protectionStateHardPaused)
	}
	if reportConsumer.pauseCalls == 0 {
		t.Fatal("expected report consumer to pause on persistence errors")
	}
}

func TestReportConsumptionProtectorSetRoundsSubSecondSampleInterval(t *testing.T) {
	protector := newReportConsumptionProtector(
		model.ProtectionConfig{}.Normalize(),
		&fakeRegulatedConsumer{},
		&fakeRegulatedConsumer{},
		&fakeConsumerSnapshotReader{},
		&fakeConsumerSnapshotReader{},
		&fakePipelineSnapshotReader{},
		&fakeWorkerPoolController{stats: util.WorkerPoolStats{MinWorkers: 1, MaxWorkers: 1}},
		&fakeWorkerPoolController{stats: util.WorkerPoolStats{MinWorkers: 1, MaxWorkers: 1}},
	)

	if err := protector.Set(protectSetRequest{SampleInterval: "500ms"}); err != nil {
		t.Fatalf("Set returned error: %v", err)
	}
	if protector.config.SampleIntervalSeconds != 1 {
		t.Fatalf("SampleIntervalSeconds = %d, want 1", protector.config.SampleIntervalSeconds)
	}
	if got := protector.currentSampleInterval(); got != time.Second {
		t.Fatalf("currentSampleInterval = %s, want 1s", got)
	}
}

func TestReportConsumptionProtectorMockOverridesStatus(t *testing.T) {
	reportRate := &fakeConsumerSnapshotReader{snapshot: consumerRateSnapshot{Lag: 10, SpeedPerSecond: 1}}
	realTimeRate := &fakeConsumerSnapshotReader{snapshot: consumerRateSnapshot{Lag: 20, SpeedPerSecond: 2}}
	pipeline := &fakePipelineSnapshotReader{
		snapshot: reportConsumerPipelineSnapshot{
			PendingCount: 1,
			Gate: reportCompletionGateSnapshot{
				InFlightMessages: 1,
				WaitingTasks:     1,
			},
		},
	}

	oldHealth := getMySQLHealthState
	oldMem := readRuntimeMemStats
	oldGoroutines := currentGoroutineNum
	oldPersistence := readPersistenceErrorSnapshot
	defer func() {
		getMySQLHealthState = oldHealth
		readRuntimeMemStats = oldMem
		currentGoroutineNum = oldGoroutines
		readPersistenceErrorSnapshot = oldPersistence
	}()

	getMySQLHealthState = func(name string) (db.DBHealthState, bool) {
		return db.DBHealthState{Name: name, Status: "healthy"}, true
	}
	readRuntimeMemStats = func(stats *runtime.MemStats) {
		stats.HeapAlloc = 128 << 20
	}
	currentGoroutineNum = func() int {
		return 100
	}
	readPersistenceErrorSnapshot = func() util.PersistenceErrorSnapshot {
		return util.PersistenceErrorSnapshot{}
	}

	protector := newReportConsumptionProtector(
		model.ProtectionConfig{
			Enabled: boolPtr(true),
			Mock: model.ProtectionMockConfig{
				Enabled:                          true,
				ReportLag:                        int64Ptr(12345),
				PipelinePending:                  intPtr(23456),
				GateInFlight:                     int64Ptr(34567),
				GateWaitingTasks:                 int64Ptr(45678),
				ReportConsumerQueueUsagePermille: intPtr(600),
				ReportConsumerBusyPermille:       intPtr(950),
				PersistenceErrorCount:            intPtr(7),
				MySQLStatus:                      stringPtr("degraded"),
				MySQLConsecutiveFailures:         intPtr(4),
			},
		}.Normalize(),
		&fakeRegulatedConsumer{},
		&fakeRegulatedConsumer{},
		reportRate,
		realTimeRate,
		pipeline,
		&fakeWorkerPoolController{stats: util.WorkerPoolStats{MinWorkers: 1, MaxWorkers: 1}},
		&fakeWorkerPoolController{stats: util.WorkerPoolStats{MinWorkers: 1, MaxWorkers: 1}},
	)

	status := protector.Status()
	if status.ReportRate.Lag != 12345 {
		t.Fatalf("ReportRate.Lag = %d, want 12345", status.ReportRate.Lag)
	}
	if status.ReportPipeline.PendingCount != 23456 {
		t.Fatalf("PendingCount = %d, want 23456", status.ReportPipeline.PendingCount)
	}
	if status.ReportPipeline.Gate.InFlightMessages != 34567 {
		t.Fatalf("Gate.InFlightMessages = %d, want 34567", status.ReportPipeline.Gate.InFlightMessages)
	}
	if status.PersistenceErrors.CountLastMinute != 7 {
		t.Fatalf("PersistenceErrors.CountLastMinute = %d, want 7", status.PersistenceErrors.CountLastMinute)
	}
	if status.MySQLHealth.Status != "degraded" {
		t.Fatalf("MySQLHealth.Status = %s, want degraded", status.MySQLHealth.Status)
	}
}

func TestReportConsumptionProtectorMockCanTriggerSoftLimited(t *testing.T) {
	oldHealth := getMySQLHealthState
	oldMem := readRuntimeMemStats
	oldGoroutines := currentGoroutineNum
	oldPersistence := readPersistenceErrorSnapshot
	defer func() {
		getMySQLHealthState = oldHealth
		readRuntimeMemStats = oldMem
		currentGoroutineNum = oldGoroutines
		readPersistenceErrorSnapshot = oldPersistence
	}()

	getMySQLHealthState = func(name string) (db.DBHealthState, bool) {
		return db.DBHealthState{Name: name, Status: "healthy"}, true
	}
	readRuntimeMemStats = func(stats *runtime.MemStats) {
		stats.HeapAlloc = 128 << 20
	}
	currentGoroutineNum = func() int {
		return 100
	}
	readPersistenceErrorSnapshot = func() util.PersistenceErrorSnapshot {
		return util.PersistenceErrorSnapshot{}
	}

	protector := newReportConsumptionProtector(
		model.ProtectionConfig{
			Enabled: boolPtr(true),
			Mock: model.ProtectionMockConfig{
				Enabled:                          true,
				PipelinePending:                  intPtr(25000),
				GateInFlight:                     int64Ptr(25000),
				ReportConsumerQueueUsagePermille: intPtr(600),
				ReportConsumerBusyPermille:       intPtr(950),
			},
		}.Normalize(),
		&fakeRegulatedConsumer{},
		&fakeRegulatedConsumer{},
		&fakeConsumerSnapshotReader{},
		&fakeConsumerSnapshotReader{},
		&fakePipelineSnapshotReader{},
		&fakeWorkerPoolController{stats: util.WorkerPoolStats{MinWorkers: 1, MaxWorkers: 1}},
		&fakeWorkerPoolController{stats: util.WorkerPoolStats{MinWorkers: 1, MaxWorkers: 1}},
	)

	for i := 0; i < 3; i++ {
		protector.sampleAndAct()
	}
	if protector.state != protectionStateSoftLimited {
		t.Fatalf("state = %s, want %s", protector.state, protectionStateSoftLimited)
	}
}

func TestReportConsumptionProtectorStatusMarksConsumerPoolSaturatedWithoutQueueBuildUp(t *testing.T) {
	reportRate := &fakeConsumerSnapshotReader{}
	realTimeRate := &fakeConsumerSnapshotReader{}
	pipeline := &fakePipelineSnapshotReader{
		snapshot: reportConsumerPipelineSnapshot{
			PendingCount: 2500,
			Gate: reportCompletionGateSnapshot{
				InFlightMessages: 2500,
				WaitingTasks:     2500,
			},
		},
	}
	reportPool := &fakeWorkerPoolController{
		stats: util.WorkerPoolStats{
			MinWorkers:      1,
			MaxWorkers:      1,
			QueueUsageRatio: 0,
			BusyRatio:       0.95,
			Capacity:        1,
		},
	}

	oldHealth := getMySQLHealthState
	oldMem := readRuntimeMemStats
	oldGoroutines := currentGoroutineNum
	oldPersistence := readPersistenceErrorSnapshot
	defer func() {
		getMySQLHealthState = oldHealth
		readRuntimeMemStats = oldMem
		currentGoroutineNum = oldGoroutines
		readPersistenceErrorSnapshot = oldPersistence
	}()

	getMySQLHealthState = func(name string) (db.DBHealthState, bool) {
		return db.DBHealthState{Name: name, Status: "healthy"}, true
	}
	readRuntimeMemStats = func(stats *runtime.MemStats) {
		stats.HeapAlloc = 128 << 20
	}
	currentGoroutineNum = func() int {
		return 100
	}
	readPersistenceErrorSnapshot = func() util.PersistenceErrorSnapshot {
		return util.PersistenceErrorSnapshot{}
	}

	protector := newReportConsumptionProtector(
		model.ProtectionConfig{}.Normalize(),
		&fakeRegulatedConsumer{},
		&fakeRegulatedConsumer{},
		reportRate,
		realTimeRate,
		pipeline,
		reportPool,
		&fakeWorkerPoolController{stats: util.WorkerPoolStats{MinWorkers: 1, MaxWorkers: 1}},
	)

	status := protector.Status()
	found := false
	for _, signal := range status.CurrentSoftSignals {
		if signal == "report_consumer_pool_saturated" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("CurrentSoftSignals = %#v, want report_consumer_pool_saturated", status.CurrentSoftSignals)
	}
}

func TestReportConsumptionProtectorStatusDoesNotMarkConsumerPoolSaturatedForSameBacklogWithLargerCapacity(t *testing.T) {
	reportRate := &fakeConsumerSnapshotReader{}
	realTimeRate := &fakeConsumerSnapshotReader{}
	pipeline := &fakePipelineSnapshotReader{
		snapshot: reportConsumerPipelineSnapshot{
			PendingCount: 600,
			Gate: reportCompletionGateSnapshot{
				InFlightMessages: 600,
				WaitingTasks:     600,
			},
		},
	}
	reportPool := &fakeWorkerPoolController{
		stats: util.WorkerPoolStats{
			MinWorkers:      4,
			MaxWorkers:      20,
			QueueUsageRatio: 0,
			BusyRatio:       0.95,
			Capacity:        20,
		},
	}

	oldHealth := getMySQLHealthState
	oldMem := readRuntimeMemStats
	oldGoroutines := currentGoroutineNum
	oldPersistence := readPersistenceErrorSnapshot
	defer func() {
		getMySQLHealthState = oldHealth
		readRuntimeMemStats = oldMem
		currentGoroutineNum = oldGoroutines
		readPersistenceErrorSnapshot = oldPersistence
	}()

	getMySQLHealthState = func(name string) (db.DBHealthState, bool) {
		return db.DBHealthState{Name: name, Status: "healthy"}, true
	}
	readRuntimeMemStats = func(stats *runtime.MemStats) {
		stats.HeapAlloc = 128 << 20
	}
	currentGoroutineNum = func() int {
		return 100
	}
	readPersistenceErrorSnapshot = func() util.PersistenceErrorSnapshot {
		return util.PersistenceErrorSnapshot{}
	}

	protector := newReportConsumptionProtector(
		model.ProtectionConfig{}.Normalize(),
		&fakeRegulatedConsumer{},
		&fakeRegulatedConsumer{},
		reportRate,
		realTimeRate,
		pipeline,
		reportPool,
		&fakeWorkerPoolController{stats: util.WorkerPoolStats{MinWorkers: 1, MaxWorkers: 1}},
	)

	status := protector.Status()
	for _, signal := range status.CurrentSoftSignals {
		if signal == "report_consumer_pool_saturated" {
			t.Fatalf("CurrentSoftSignals = %#v, want no report_consumer_pool_saturated", status.CurrentSoftSignals)
		}
	}
}

func TestReportConsumptionProtectorStatusDoesNotMarkConsumerPoolSaturatedWhenBusyBelowThreshold(t *testing.T) {
	reportRate := &fakeConsumerSnapshotReader{}
	realTimeRate := &fakeConsumerSnapshotReader{}
	pipeline := &fakePipelineSnapshotReader{
		snapshot: reportConsumerPipelineSnapshot{
			PendingCount: 5000,
			Gate: reportCompletionGateSnapshot{
				InFlightMessages: 5000,
				WaitingTasks:     5000,
			},
		},
	}
	reportPool := &fakeWorkerPoolController{
		stats: util.WorkerPoolStats{
			MinWorkers:      1,
			MaxWorkers:      1,
			QueueUsageRatio: 0,
			BusyRatio:       0.89,
			Capacity:        1,
		},
	}

	oldHealth := getMySQLHealthState
	oldMem := readRuntimeMemStats
	oldGoroutines := currentGoroutineNum
	oldPersistence := readPersistenceErrorSnapshot
	defer func() {
		getMySQLHealthState = oldHealth
		readRuntimeMemStats = oldMem
		currentGoroutineNum = oldGoroutines
		readPersistenceErrorSnapshot = oldPersistence
	}()

	getMySQLHealthState = func(name string) (db.DBHealthState, bool) {
		return db.DBHealthState{Name: name, Status: "healthy"}, true
	}
	readRuntimeMemStats = func(stats *runtime.MemStats) {
		stats.HeapAlloc = 128 << 20
	}
	currentGoroutineNum = func() int {
		return 100
	}
	readPersistenceErrorSnapshot = func() util.PersistenceErrorSnapshot {
		return util.PersistenceErrorSnapshot{}
	}

	protector := newReportConsumptionProtector(
		model.ProtectionConfig{}.Normalize(),
		&fakeRegulatedConsumer{},
		&fakeRegulatedConsumer{},
		reportRate,
		realTimeRate,
		pipeline,
		reportPool,
		&fakeWorkerPoolController{stats: util.WorkerPoolStats{MinWorkers: 1, MaxWorkers: 1}},
	)

	status := protector.Status()
	for _, signal := range status.CurrentSoftSignals {
		if signal == "report_consumer_pool_saturated" {
			t.Fatalf("CurrentSoftSignals = %#v, want no report_consumer_pool_saturated", status.CurrentSoftSignals)
		}
	}
}

func TestReportConsumptionProtectorStatusDoesNotMarkConsumerPoolSaturatedWhenOnlyOneBacklogSideCrossesThreshold(t *testing.T) {
	reportRate := &fakeConsumerSnapshotReader{}
	realTimeRate := &fakeConsumerSnapshotReader{}
	pipeline := &fakePipelineSnapshotReader{
		snapshot: reportConsumerPipelineSnapshot{
			PendingCount: 2500,
			Gate: reportCompletionGateSnapshot{
				InFlightMessages: 100,
				WaitingTasks:     2500,
			},
		},
	}
	reportPool := &fakeWorkerPoolController{
		stats: util.WorkerPoolStats{
			MinWorkers:      1,
			MaxWorkers:      1,
			QueueUsageRatio: 0,
			BusyRatio:       0.95,
			Capacity:        1,
		},
	}

	oldHealth := getMySQLHealthState
	oldMem := readRuntimeMemStats
	oldGoroutines := currentGoroutineNum
	oldPersistence := readPersistenceErrorSnapshot
	defer func() {
		getMySQLHealthState = oldHealth
		readRuntimeMemStats = oldMem
		currentGoroutineNum = oldGoroutines
		readPersistenceErrorSnapshot = oldPersistence
	}()

	getMySQLHealthState = func(name string) (db.DBHealthState, bool) {
		return db.DBHealthState{Name: name, Status: "healthy"}, true
	}
	readRuntimeMemStats = func(stats *runtime.MemStats) {
		stats.HeapAlloc = 128 << 20
	}
	currentGoroutineNum = func() int {
		return 100
	}
	readPersistenceErrorSnapshot = func() util.PersistenceErrorSnapshot {
		return util.PersistenceErrorSnapshot{}
	}

	protector := newReportConsumptionProtector(
		model.ProtectionConfig{}.Normalize(),
		&fakeRegulatedConsumer{},
		&fakeRegulatedConsumer{},
		reportRate,
		realTimeRate,
		pipeline,
		reportPool,
		&fakeWorkerPoolController{stats: util.WorkerPoolStats{MinWorkers: 1, MaxWorkers: 1}},
	)

	status := protector.Status()
	for _, signal := range status.CurrentSoftSignals {
		if signal == "report_consumer_pool_saturated" {
			t.Fatalf("CurrentSoftSignals = %#v, want no report_consumer_pool_saturated", status.CurrentSoftSignals)
		}
	}
}

func TestReportConsumptionProtectorStatusMarksConsumerPoolSaturatedAtExactDynamicThreshold(t *testing.T) {
	reportRate := &fakeConsumerSnapshotReader{}
	realTimeRate := &fakeConsumerSnapshotReader{}
	pipeline := &fakePipelineSnapshotReader{
		snapshot: reportConsumerPipelineSnapshot{
			PendingCount: 500,
			Gate: reportCompletionGateSnapshot{
				InFlightMessages: 500,
				WaitingTasks:     500,
			},
		},
	}
	reportPool := &fakeWorkerPoolController{
		stats: util.WorkerPoolStats{
			MinWorkers:      1,
			MaxWorkers:      1,
			QueueUsageRatio: 0,
			BusyRatio:       0.90,
			Capacity:        1,
		},
	}

	oldHealth := getMySQLHealthState
	oldMem := readRuntimeMemStats
	oldGoroutines := currentGoroutineNum
	oldPersistence := readPersistenceErrorSnapshot
	defer func() {
		getMySQLHealthState = oldHealth
		readRuntimeMemStats = oldMem
		currentGoroutineNum = oldGoroutines
		readPersistenceErrorSnapshot = oldPersistence
	}()

	getMySQLHealthState = func(name string) (db.DBHealthState, bool) {
		return db.DBHealthState{Name: name, Status: "healthy"}, true
	}
	readRuntimeMemStats = func(stats *runtime.MemStats) {
		stats.HeapAlloc = 128 << 20
	}
	currentGoroutineNum = func() int {
		return 100
	}
	readPersistenceErrorSnapshot = func() util.PersistenceErrorSnapshot {
		return util.PersistenceErrorSnapshot{}
	}

	protector := newReportConsumptionProtector(
		model.ProtectionConfig{}.Normalize(),
		&fakeRegulatedConsumer{},
		&fakeRegulatedConsumer{},
		reportRate,
		realTimeRate,
		pipeline,
		reportPool,
		&fakeWorkerPoolController{stats: util.WorkerPoolStats{MinWorkers: 1, MaxWorkers: 1}},
	)

	status := protector.Status()
	found := false
	for _, signal := range status.CurrentSoftSignals {
		if signal == "report_consumer_pool_saturated" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("CurrentSoftSignals = %#v, want report_consumer_pool_saturated at exact threshold", status.CurrentSoftSignals)
	}
}

func TestReportConsumptionProtectorStatusAddsPoolSaturationSignalWithoutQueueBuildUp(t *testing.T) {
	reportConsumer := &fakeRegulatedConsumer{}
	realTimeConsumer := &fakeRegulatedConsumer{}
	reportRate := &fakeConsumerSnapshotReader{}
	realTimeRate := &fakeConsumerSnapshotReader{}
	pipeline := &fakePipelineSnapshotReader{
		snapshot: reportConsumerPipelineSnapshot{
			PendingCount: 2500,
			Gate: reportCompletionGateSnapshot{
				InFlightMessages: 2500,
				WaitingTasks:     2500,
			},
		},
	}
	reportPool := &fakeWorkerPoolController{
		stats: util.WorkerPoolStats{
			MinWorkers:      4,
			MaxWorkers:      20,
			QueueUsageRatio: 0,
			BusyRatio:       0.95,
		},
	}

	oldHealth := getMySQLHealthState
	oldMem := readRuntimeMemStats
	oldGoroutines := currentGoroutineNum
	oldPersistence := readPersistenceErrorSnapshot
	defer func() {
		getMySQLHealthState = oldHealth
		readRuntimeMemStats = oldMem
		currentGoroutineNum = oldGoroutines
		readPersistenceErrorSnapshot = oldPersistence
	}()

	getMySQLHealthState = func(name string) (db.DBHealthState, bool) {
		return db.DBHealthState{Name: name, Status: "healthy"}, true
	}
	readRuntimeMemStats = func(stats *runtime.MemStats) {
		stats.HeapAlloc = 128 << 20
	}
	currentGoroutineNum = func() int {
		return 100
	}
	readPersistenceErrorSnapshot = func() util.PersistenceErrorSnapshot {
		return util.PersistenceErrorSnapshot{}
	}

	protector := newReportConsumptionProtector(
		model.ProtectionConfig{}.Normalize(),
		reportConsumer,
		realTimeConsumer,
		reportRate,
		realTimeRate,
		pipeline,
		reportPool,
		&fakeWorkerPoolController{stats: util.WorkerPoolStats{MinWorkers: 1, MaxWorkers: 1}},
	)

	status := protector.Status()
	found := false
	for _, signal := range status.CurrentSoftSignals {
		if signal == "report_consumer_pool_saturated" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("CurrentSoftSignals = %#v, want report_consumer_pool_saturated", status.CurrentSoftSignals)
	}
}

func TestReportConsumptionProtectorMockCanOverrideToZero(t *testing.T) {
	reportRate := &fakeConsumerSnapshotReader{snapshot: consumerRateSnapshot{Lag: 100, SpeedPerSecond: 1}}
	realTimeRate := &fakeConsumerSnapshotReader{snapshot: consumerRateSnapshot{Lag: 100, SpeedPerSecond: 2}}
	pipeline := &fakePipelineSnapshotReader{
		snapshot: reportConsumerPipelineSnapshot{
			PendingCount: 100,
			Gate: reportCompletionGateSnapshot{
				InFlightMessages: 100,
				WaitingTasks:     100,
			},
		},
	}

	oldHealth := getMySQLHealthState
	oldMem := readRuntimeMemStats
	oldGoroutines := currentGoroutineNum
	oldPersistence := readPersistenceErrorSnapshot
	defer func() {
		getMySQLHealthState = oldHealth
		readRuntimeMemStats = oldMem
		currentGoroutineNum = oldGoroutines
		readPersistenceErrorSnapshot = oldPersistence
	}()

	getMySQLHealthState = func(name string) (db.DBHealthState, bool) {
		return db.DBHealthState{Name: name, Status: "healthy"}, true
	}
	readRuntimeMemStats = func(stats *runtime.MemStats) {
		stats.HeapAlloc = 256 << 20
	}
	currentGoroutineNum = func() int {
		return 100
	}
	readPersistenceErrorSnapshot = func() util.PersistenceErrorSnapshot {
		return util.PersistenceErrorSnapshot{
			CountLastMinute: 5,
			LastClass:       "real_error",
		}
	}

	protector := newReportConsumptionProtector(
		model.ProtectionConfig{
			Enabled: boolPtr(true),
			Mock: model.ProtectionMockConfig{
				Enabled:               true,
				ReportLag:             int64Ptr(0),
				RealTimeLag:           int64Ptr(0),
				PipelinePending:       intPtr(0),
				GateInFlight:          int64Ptr(0),
				GateWaitingTasks:      int64Ptr(0),
				PersistenceErrorCount: intPtr(0),
			},
		}.Normalize(),
		&fakeRegulatedConsumer{},
		&fakeRegulatedConsumer{},
		reportRate,
		realTimeRate,
		pipeline,
		&fakeWorkerPoolController{stats: util.WorkerPoolStats{MinWorkers: 1, MaxWorkers: 1}},
		&fakeWorkerPoolController{stats: util.WorkerPoolStats{MinWorkers: 1, MaxWorkers: 1}},
	)

	status := protector.Status()
	if status.ReportRate.Lag != 0 || status.RealTimeRate.Lag != 0 {
		t.Fatalf("mock zero lag override failed: report=%d realtime=%d", status.ReportRate.Lag, status.RealTimeRate.Lag)
	}
	if status.ReportPipeline.PendingCount != 0 || status.ReportPipeline.Gate.InFlightMessages != 0 || status.ReportPipeline.Gate.WaitingTasks != 0 {
		t.Fatalf("mock zero pipeline override failed: %+v", status.ReportPipeline)
	}
	if status.PersistenceErrors.CountLastMinute != 0 {
		t.Fatalf("mock zero persistence error override failed: %d", status.PersistenceErrors.CountLastMinute)
	}
	if status.PersistenceErrors.LastClass != "" || status.PersistenceErrors.LastError != "" || !status.PersistenceErrors.LastOccurredAt.IsZero() {
		t.Fatalf("mock zero persistence metadata override failed: %+v", status.PersistenceErrors)
	}
}

func TestReportConsumptionProtectorMockSpeedDoesNotPolluteTrendHistory(t *testing.T) {
	oldHealth := getMySQLHealthState
	oldMem := readRuntimeMemStats
	oldGoroutines := currentGoroutineNum
	oldPersistence := readPersistenceErrorSnapshot
	defer func() {
		getMySQLHealthState = oldHealth
		readRuntimeMemStats = oldMem
		currentGoroutineNum = oldGoroutines
		readPersistenceErrorSnapshot = oldPersistence
	}()

	getMySQLHealthState = func(name string) (db.DBHealthState, bool) {
		return db.DBHealthState{Name: name, Status: "healthy"}, true
	}
	readRuntimeMemStats = func(stats *runtime.MemStats) {
		stats.HeapAlloc = 128 << 20
	}
	currentGoroutineNum = func() int {
		return 100
	}
	readPersistenceErrorSnapshot = func() util.PersistenceErrorSnapshot {
		return util.PersistenceErrorSnapshot{}
	}

	protector := newReportConsumptionProtector(
		model.ProtectionConfig{
			Enabled: boolPtr(true),
			Mock: model.ProtectionMockConfig{
				Enabled:              true,
				ReportSpeedPerSecond: float64Ptr(10),
			},
		}.Normalize(),
		&fakeRegulatedConsumer{},
		&fakeRegulatedConsumer{},
		&fakeConsumerSnapshotReader{snapshot: consumerRateSnapshot{SpeedPerSecond: 1000}},
		&fakeConsumerSnapshotReader{},
		&fakePipelineSnapshotReader{},
		&fakeWorkerPoolController{stats: util.WorkerPoolStats{MinWorkers: 1, MaxWorkers: 1}},
		&fakeWorkerPoolController{stats: util.WorkerPoolStats{MinWorkers: 1, MaxWorkers: 1}},
	)
	protector.recentSpeeds = []float64{1000, 900, 800}

	protector.sampleAndAct()
	if len(protector.recentSpeeds) != 0 {
		t.Fatalf("recentSpeeds = %#v, want cleared when mock speed is enabled", protector.recentSpeeds)
	}
}

func TestReportConsumptionProtectorRecoveryBlockedByMySQLDegraded(t *testing.T) {
	reportConsumer := &fakeRegulatedConsumer{paused: true}
	realTimeConsumer := &fakeRegulatedConsumer{paused: true}
	reportRate := &fakeConsumerSnapshotReader{snapshot: consumerRateSnapshot{Lag: 0, SpeedPerSecond: 0}}
	realTimeRate := &fakeConsumerSnapshotReader{snapshot: consumerRateSnapshot{Lag: 0, SpeedPerSecond: 0}}
	pipeline := &fakePipelineSnapshotReader{}

	oldHealth := getMySQLHealthState
	oldMem := readRuntimeMemStats
	oldGoroutines := currentGoroutineNum
	oldPersistence := readPersistenceErrorSnapshot
	defer func() {
		getMySQLHealthState = oldHealth
		readRuntimeMemStats = oldMem
		currentGoroutineNum = oldGoroutines
		readPersistenceErrorSnapshot = oldPersistence
	}()

	getMySQLHealthState = func(name string) (db.DBHealthState, bool) {
		return db.DBHealthState{Name: name, Status: "degraded", ConsecutiveFailures: 5}, true
	}
	readRuntimeMemStats = func(stats *runtime.MemStats) {
		stats.HeapAlloc = 128 << 20
	}
	currentGoroutineNum = func() int {
		return 100
	}
	readPersistenceErrorSnapshot = func() util.PersistenceErrorSnapshot {
		return util.PersistenceErrorSnapshot{}
	}

	protector := newReportConsumptionProtector(
		model.ProtectionConfig{}.Normalize(),
		reportConsumer,
		realTimeConsumer,
		reportRate,
		realTimeRate,
		pipeline,
		&fakeWorkerPoolController{stats: util.WorkerPoolStats{MinWorkers: 1, MaxWorkers: 1}},
		&fakeWorkerPoolController{stats: util.WorkerPoolStats{MinWorkers: 1, MaxWorkers: 1}},
	)
	protector.state = protectionStateHardPaused
	protector.hardHoldUntil = time.Now().Add(-time.Second)
	protector.recentSpeeds = []float64{1000, 1100, 1200}

	for i := 0; i < 3; i++ {
		protector.sampleAndAct()
	}
	if protector.state != protectionStateHardPaused {
		t.Fatalf("state = %s, want remain hard_paused", protector.state)
	}
}

func TestReportConsumptionProtectorDoesNotPauseWhenOnlyGoroutinesHighAndPipelineIdle(t *testing.T) {
	reportConsumer := &fakeRegulatedConsumer{}
	realTimeConsumer := &fakeRegulatedConsumer{}
	reportRate := &fakeConsumerSnapshotReader{}
	realTimeRate := &fakeConsumerSnapshotReader{}
	pipeline := &fakePipelineSnapshotReader{}
	reportPool := &fakeWorkerPoolController{stats: util.WorkerPoolStats{MinWorkers: 4, MaxWorkers: 16}}
	persistPool := &fakeWorkerPoolController{stats: util.WorkerPoolStats{MinWorkers: 2, MaxWorkers: 8}}

	oldHealth := getMySQLHealthState
	oldMem := readRuntimeMemStats
	oldGoroutines := currentGoroutineNum
	oldPersistence := readPersistenceErrorSnapshot
	defer func() {
		getMySQLHealthState = oldHealth
		readRuntimeMemStats = oldMem
		currentGoroutineNum = oldGoroutines
		readPersistenceErrorSnapshot = oldPersistence
	}()

	getMySQLHealthState = func(name string) (db.DBHealthState, bool) {
		return db.DBHealthState{Name: name, Status: "healthy"}, true
	}
	readRuntimeMemStats = func(stats *runtime.MemStats) {
		stats.HeapAlloc = 128 << 20
	}
	currentGoroutineNum = func() int {
		return 5000
	}
	readPersistenceErrorSnapshot = func() util.PersistenceErrorSnapshot {
		return util.PersistenceErrorSnapshot{}
	}

	protector := newReportConsumptionProtector(
		model.ProtectionConfig{}.Normalize(),
		reportConsumer,
		realTimeConsumer,
		reportRate,
		realTimeRate,
		pipeline,
		reportPool,
		persistPool,
	)

	for i := 0; i < 3; i++ {
		protector.sampleAndAct()
	}

	if protector.state != protectionStateNormal {
		t.Fatalf("state = %s, want remain normal", protector.state)
	}
	if reportConsumer.pauseCalls != 0 || realTimeConsumer.pauseCalls != 0 {
		t.Fatalf("unexpected pause calls report=%d realtime=%d", reportConsumer.pauseCalls, realTimeConsumer.pauseCalls)
	}
}

func TestReportConsumptionProtectorRecoversFromHardPauseWhenOnlyGoroutinesRemainHigh(t *testing.T) {
	reportConsumer := &fakeRegulatedConsumer{paused: true}
	realTimeConsumer := &fakeRegulatedConsumer{paused: true}
	reportRate := &fakeConsumerSnapshotReader{snapshot: consumerRateSnapshot{Lag: 0, SpeedPerSecond: 0}}
	realTimeRate := &fakeConsumerSnapshotReader{snapshot: consumerRateSnapshot{Lag: 0, SpeedPerSecond: 0}}
	pipeline := &fakePipelineSnapshotReader{}
	reportPool := &fakeWorkerPoolController{stats: util.WorkerPoolStats{MinWorkers: 4, MaxWorkers: 16}}
	persistPool := &fakeWorkerPoolController{stats: util.WorkerPoolStats{MinWorkers: 2, MaxWorkers: 8}}

	oldHealth := getMySQLHealthState
	oldMem := readRuntimeMemStats
	oldGoroutines := currentGoroutineNum
	oldPersistence := readPersistenceErrorSnapshot
	defer func() {
		getMySQLHealthState = oldHealth
		readRuntimeMemStats = oldMem
		currentGoroutineNum = oldGoroutines
		readPersistenceErrorSnapshot = oldPersistence
	}()

	getMySQLHealthState = func(name string) (db.DBHealthState, bool) {
		return db.DBHealthState{Name: name, Status: "healthy"}, true
	}
	readRuntimeMemStats = func(stats *runtime.MemStats) {
		stats.HeapAlloc = 128 << 20
	}
	currentGoroutineNum = func() int {
		return 5000
	}
	readPersistenceErrorSnapshot = func() util.PersistenceErrorSnapshot {
		return util.PersistenceErrorSnapshot{}
	}

	protector := newReportConsumptionProtector(
		model.ProtectionConfig{}.Normalize(),
		reportConsumer,
		realTimeConsumer,
		reportRate,
		realTimeRate,
		pipeline,
		reportPool,
		persistPool,
	)
	protector.state = protectionStateHardPaused
	protector.hardHoldUntil = time.Now().Add(-time.Second)
	protector.recentSpeeds = []float64{1000, 1100, 1200}

	for i := 0; i < 3; i++ {
		protector.sampleAndAct()
	}

	if protector.state != protectionStateSoftLimited {
		t.Fatalf("state after recovery = %s, want soft_limited", protector.state)
	}
	if reportConsumer.resumeCalls == 0 || realTimeConsumer.resumeCalls == 0 {
		t.Fatalf("expected both consumers to resume, got report=%d realtime=%d", reportConsumer.resumeCalls, realTimeConsumer.resumeCalls)
	}
}

func TestReportConsumptionProtectorDoesNotPauseWhenGoroutinesMatchPartitionBaseline(t *testing.T) {
	reportConsumer := &fakeRegulatedConsumer{}
	realTimeConsumer := &fakeRegulatedConsumer{}
	reportRate := &fakeConsumerSnapshotReader{snapshot: consumerRateSnapshot{PartitionCount: 300}}
	realTimeRate := &fakeConsumerSnapshotReader{snapshot: consumerRateSnapshot{PartitionCount: 300}}
	pipeline := &fakePipelineSnapshotReader{
		snapshot: reportConsumerPipelineSnapshot{
			PendingCount: 600,
			Gate: reportCompletionGateSnapshot{
				InFlightMessages: 600,
				WaitingTasks:     600,
			},
		},
	}
	reportPool := &fakeWorkerPoolController{stats: util.WorkerPoolStats{MinWorkers: 4, MaxWorkers: 16}}
	persistPool := &fakeWorkerPoolController{stats: util.WorkerPoolStats{MinWorkers: 2, MaxWorkers: 8}}

	oldHealth := getMySQLHealthState
	oldMem := readRuntimeMemStats
	oldGoroutines := currentGoroutineNum
	oldPersistence := readPersistenceErrorSnapshot
	defer func() {
		getMySQLHealthState = oldHealth
		readRuntimeMemStats = oldMem
		currentGoroutineNum = oldGoroutines
		readPersistenceErrorSnapshot = oldPersistence
	}()

	getMySQLHealthState = func(name string) (db.DBHealthState, bool) {
		return db.DBHealthState{Name: name, Status: "healthy"}, true
	}
	readRuntimeMemStats = func(stats *runtime.MemStats) {
		stats.HeapAlloc = 128 << 20
	}
	currentGoroutineNum = func() int {
		return 4265
	}
	readPersistenceErrorSnapshot = func() util.PersistenceErrorSnapshot {
		return util.PersistenceErrorSnapshot{}
	}

	protector := newReportConsumptionProtector(
		model.ProtectionConfig{}.Normalize(),
		reportConsumer,
		realTimeConsumer,
		reportRate,
		realTimeRate,
		pipeline,
		reportPool,
		persistPool,
	)

	for i := 0; i < 3; i++ {
		protector.sampleAndAct()
	}

	if protector.state != protectionStateNormal {
		t.Fatalf("state = %s, want remain normal", protector.state)
	}
}

func TestReportConsumptionProtectorPausesWhenGoroutinesExceedDerivedBaselineUnderPressure(t *testing.T) {
	reportConsumer := &fakeRegulatedConsumer{}
	realTimeConsumer := &fakeRegulatedConsumer{}
	reportRate := &fakeConsumerSnapshotReader{snapshot: consumerRateSnapshot{PartitionCount: 300}}
	realTimeRate := &fakeConsumerSnapshotReader{snapshot: consumerRateSnapshot{PartitionCount: 300}}
	pipeline := &fakePipelineSnapshotReader{
		snapshot: reportConsumerPipelineSnapshot{
			PendingCount: 600,
			Gate: reportCompletionGateSnapshot{
				InFlightMessages: 600,
				WaitingTasks:     600,
			},
		},
	}
	reportPool := &fakeWorkerPoolController{stats: util.WorkerPoolStats{MinWorkers: 4, MaxWorkers: 16}}
	persistPool := &fakeWorkerPoolController{stats: util.WorkerPoolStats{MinWorkers: 2, MaxWorkers: 8}}

	oldHealth := getMySQLHealthState
	oldMem := readRuntimeMemStats
	oldGoroutines := currentGoroutineNum
	oldPersistence := readPersistenceErrorSnapshot
	defer func() {
		getMySQLHealthState = oldHealth
		readRuntimeMemStats = oldMem
		currentGoroutineNum = oldGoroutines
		readPersistenceErrorSnapshot = oldPersistence
	}()

	getMySQLHealthState = func(name string) (db.DBHealthState, bool) {
		return db.DBHealthState{Name: name, Status: "healthy"}, true
	}
	readRuntimeMemStats = func(stats *runtime.MemStats) {
		stats.HeapAlloc = 128 << 20
	}
	currentGoroutineNum = func() int {
		return 6000
	}
	readPersistenceErrorSnapshot = func() util.PersistenceErrorSnapshot {
		return util.PersistenceErrorSnapshot{}
	}

	protector := newReportConsumptionProtector(
		model.ProtectionConfig{}.Normalize(),
		reportConsumer,
		realTimeConsumer,
		reportRate,
		realTimeRate,
		pipeline,
		reportPool,
		persistPool,
	)

	for i := 0; i < 3; i++ {
		protector.sampleAndAct()
	}

	if protector.state != protectionStateHardPaused {
		t.Fatalf("state = %s, want hard_paused", protector.state)
	}
}

func TestReportConsumptionProtectorRecoveryBlockedByConsumerPoolSaturation(t *testing.T) {
	reportConsumer := &fakeRegulatedConsumer{paused: true}
	realTimeConsumer := &fakeRegulatedConsumer{paused: true}
	reportRate := &fakeConsumerSnapshotReader{snapshot: consumerRateSnapshot{Lag: 0, SpeedPerSecond: 0}}
	realTimeRate := &fakeConsumerSnapshotReader{snapshot: consumerRateSnapshot{Lag: 0, SpeedPerSecond: 0}}
	pipeline := &fakePipelineSnapshotReader{
		snapshot: reportConsumerPipelineSnapshot{
			PendingCount: 2500,
			Gate: reportCompletionGateSnapshot{
				InFlightMessages: 2500,
				WaitingTasks:     2500,
			},
		},
	}
	reportPool := &fakeWorkerPoolController{
		stats: util.WorkerPoolStats{
			MinWorkers:      1,
			MaxWorkers:      1,
			Capacity:        1,
			QueueUsageRatio: 0,
			BusyRatio:       0.95,
		},
	}

	oldHealth := getMySQLHealthState
	oldMem := readRuntimeMemStats
	oldGoroutines := currentGoroutineNum
	oldPersistence := readPersistenceErrorSnapshot
	defer func() {
		getMySQLHealthState = oldHealth
		readRuntimeMemStats = oldMem
		currentGoroutineNum = oldGoroutines
		readPersistenceErrorSnapshot = oldPersistence
	}()

	getMySQLHealthState = func(name string) (db.DBHealthState, bool) {
		return db.DBHealthState{Name: name, Status: "healthy"}, true
	}
	readRuntimeMemStats = func(stats *runtime.MemStats) {
		stats.HeapAlloc = 128 << 20
	}
	currentGoroutineNum = func() int {
		return 100
	}
	readPersistenceErrorSnapshot = func() util.PersistenceErrorSnapshot {
		return util.PersistenceErrorSnapshot{}
	}

	protector := newReportConsumptionProtector(
		model.ProtectionConfig{}.Normalize(),
		reportConsumer,
		realTimeConsumer,
		reportRate,
		realTimeRate,
		pipeline,
		reportPool,
		&fakeWorkerPoolController{stats: util.WorkerPoolStats{MinWorkers: 1, MaxWorkers: 1}},
	)
	protector.state = protectionStateSoftLimited
	protector.softHoldUntil = time.Now().Add(-time.Second)

	for i := 0; i < 3; i++ {
		protector.sampleAndAct()
	}
	if protector.state != protectionStateSoftLimited {
		t.Fatalf("state = %s, want remain soft_limited", protector.state)
	}
}

func TestReportConsumptionProtectorHardPauseEscalationDoesNotAccumulateAcrossRecoveries(t *testing.T) {
	reportConsumer := &fakeRegulatedConsumer{}
	realTimeConsumer := &fakeRegulatedConsumer{}
	reportRate := &fakeConsumerSnapshotReader{}
	realTimeRate := &fakeConsumerSnapshotReader{}
	pipeline := &fakePipelineSnapshotReader{
		snapshot: reportConsumerPipelineSnapshot{
			PendingCount: 120000,
			Gate: reportCompletionGateSnapshot{
				InFlightMessages: 120000,
				WaitingTasks:     250000,
			},
		},
	}

	oldHealth := getMySQLHealthState
	oldMem := readRuntimeMemStats
	oldGoroutines := currentGoroutineNum
	oldPersistence := readPersistenceErrorSnapshot
	defer func() {
		getMySQLHealthState = oldHealth
		readRuntimeMemStats = oldMem
		currentGoroutineNum = oldGoroutines
		readPersistenceErrorSnapshot = oldPersistence
	}()

	getMySQLHealthState = func(name string) (db.DBHealthState, bool) {
		return db.DBHealthState{Name: name, Status: "healthy"}, true
	}
	readRuntimeMemStats = func(stats *runtime.MemStats) {
		stats.HeapAlloc = 128 << 20
	}
	currentGoroutineNum = func() int {
		return 100
	}
	readPersistenceErrorSnapshot = func() util.PersistenceErrorSnapshot {
		return util.PersistenceErrorSnapshot{}
	}

	protector := newReportConsumptionProtector(
		model.ProtectionConfig{}.Normalize(),
		reportConsumer,
		realTimeConsumer,
		reportRate,
		realTimeRate,
		pipeline,
		&fakeWorkerPoolController{stats: util.WorkerPoolStats{MinWorkers: 1, MaxWorkers: 1}},
		&fakeWorkerPoolController{stats: util.WorkerPoolStats{MinWorkers: 1, MaxWorkers: 1}},
	)

	for i := 0; i < 3; i++ {
		protector.sampleAndAct()
	}
	if protector.state != protectionStateHardPaused {
		t.Fatalf("first state = %s, want hard_paused", protector.state)
	}

	protector.transitionLocked(protectionStateSoftLimited, time.Now(), "recover to soft")
	protector.transitionLocked(protectionStateNormal, time.Now(), "recover to normal")
	if protector.hardPauseWindows != 0 {
		t.Fatalf("hardPauseWindows = %d, want 0 after recovery", protector.hardPauseWindows)
	}

	for i := 0; i < 3; i++ {
		protector.sampleAndAct()
	}
	if realTimeConsumer.pauseCalls != 0 {
		t.Fatalf("realTimeConsumer pauseCalls = %d, want 0 before sustained hard pause windows", realTimeConsumer.pauseCalls)
	}
}

func TestReportConsumptionProtectorTransitionToNormalClearsHoldMetadata(t *testing.T) {
	protector := newReportConsumptionProtector(
		model.ProtectionConfig{}.Normalize(),
		&fakeRegulatedConsumer{},
		&fakeRegulatedConsumer{},
		&fakeConsumerSnapshotReader{},
		&fakeConsumerSnapshotReader{},
		&fakePipelineSnapshotReader{},
		&fakeWorkerPoolController{stats: util.WorkerPoolStats{MinWorkers: 1, MaxWorkers: 1}},
		&fakeWorkerPoolController{stats: util.WorkerPoolStats{MinWorkers: 1, MaxWorkers: 1}},
	)

	start := time.Date(2026, 4, 17, 11, 0, 0, 0, time.FixedZone("CST", 8*3600))
	protector.transitionLocked(protectionStateSoftLimited, start, "soft thresholds reached")
	if protector.softHoldUntil.IsZero() {
		t.Fatal("expected softHoldUntil to be set")
	}

	recoveredAt := start.Add(2 * time.Minute)
	protector.transitionLocked(protectionStateNormal, recoveredAt, "healthy windows recovered to normal")

	if !protector.softHoldUntil.IsZero() {
		t.Fatalf("softHoldUntil = %s, want zero", protector.softHoldUntil)
	}
	if !protector.hardHoldUntil.IsZero() {
		t.Fatalf("hardHoldUntil = %s, want zero", protector.hardHoldUntil)
	}
	if protector.lastTransitionAt != recoveredAt {
		t.Fatalf("lastTransitionAt = %s, want %s", protector.lastTransitionAt, recoveredAt)
	}
	if protector.lastTransitionReason != "healthy windows recovered to normal" {
		t.Fatalf("lastTransitionReason = %q", protector.lastTransitionReason)
	}
}

func TestReportConsumptionProtectorRapidRelapseExtendsHardHold(t *testing.T) {
	protector := newReportConsumptionProtector(
		model.ProtectionConfig{}.Normalize(),
		&fakeRegulatedConsumer{},
		&fakeRegulatedConsumer{},
		&fakeConsumerSnapshotReader{},
		&fakeConsumerSnapshotReader{},
		&fakePipelineSnapshotReader{},
		&fakeWorkerPoolController{stats: util.WorkerPoolStats{MinWorkers: 1, MaxWorkers: 1}},
		&fakeWorkerPoolController{stats: util.WorkerPoolStats{MinWorkers: 1, MaxWorkers: 1}},
	)

	start := time.Date(2026, 4, 17, 11, 0, 0, 0, time.FixedZone("CST", 8*3600))
	protector.transitionLocked(protectionStateHardPaused, start, "hard thresholds reached")
	firstHold := protector.hardHoldUntil.Sub(start)

	protector.transitionLocked(protectionStateSoftLimited, start.Add(3*time.Minute), "healthy windows recovered from hard pause")
	protector.transitionLocked(protectionStateNormal, start.Add(4*time.Minute), "healthy windows recovered to normal")
	relapseAt := start.Add(4*time.Minute + 5*time.Second)
	protector.transitionLocked(protectionStateHardPaused, relapseAt, "hard thresholds reached")
	secondHold := protector.hardHoldUntil.Sub(relapseAt)

	if secondHold <= firstHold {
		t.Fatalf("secondHold = %s, want > firstHold %s", secondHold, firstHold)
	}
	if secondHold != firstHold*2 {
		t.Fatalf("secondHold = %s, want %s", secondHold, firstHold*2)
	}
}
