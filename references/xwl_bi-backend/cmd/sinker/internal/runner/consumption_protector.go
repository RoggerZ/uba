package runner

import (
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/1340691923/xwl_bi/engine/db"
	"github.com/1340691923/xwl_bi/engine/logs"
	"github.com/1340691923/xwl_bi/model"
	"github.com/1340691923/xwl_bi/platform-basic-libs/util"
	"go.uber.org/zap"
)

type protectionState string

const (
	protectionStateNormal                   protectionState = "normal"
	protectionStateSoftLimited              protectionState = "soft_limited"
	protectionStateHardPaused               protectionState = "hard_paused"
	kafkaClaimGoroutinesPerPartitionGroup                   = 7
	kafkaClaimGoroutineSoftHeadroomPermille                 = 100
	kafkaClaimGoroutineHardHeadroomPermille                 = 250
	kafkaClaimGoroutineSoftHeadroomMin                      = 256
	kafkaClaimGoroutineHardHeadroomMin                      = 512
)

var (
	getMySQLHealthState          = db.GetDBHealthState
	readRuntimeMemStats          = runtime.ReadMemStats
	currentGoroutineNum          = runtime.NumGoroutine
	readPersistenceErrorSnapshot = util.GetPersistenceErrorSnapshot
)

type hostResourceSnapshot struct{}

type HostResourceSampler interface {
	Sample() hostResourceSnapshot
}

type noopHostResourceSampler struct{}

func (noopHostResourceSampler) Sample() hostResourceSnapshot {
	return hostResourceSnapshot{}
}

type consumerSnapshotReader interface {
	Snapshot() consumerRateSnapshot
	SetLogIntervals(normal, diagnostic time.Duration)
}

type pipelineSnapshotReader interface {
	Snapshot() reportConsumerPipelineSnapshot
}

type workerPoolController interface {
	Stats() util.WorkerPoolStats
	SetRuntimeBounds(minWorkers, maxWorkers int)
	ResetRuntimeBounds()
}

type regulatedConsumer interface {
	PauseConsumption()
	ResumeConsumption()
	IsPaused() bool
	SetConsumeRegulator(regulator interface{ Wait() })
}

type protectionStatusResponse struct {
	Enabled                 bool                           `json:"enabled"`
	ObserveOnly             bool                           `json:"observeOnly"`
	State                   protectionState                `json:"state"`
	SoftHoldUntil           time.Time                      `json:"softHoldUntil,omitempty"`
	HardHoldUntil           time.Time                      `json:"hardHoldUntil,omitempty"`
	LastTransitionAt        time.Time                      `json:"lastTransitionAt,omitempty"`
	LastTransitionReason    string                         `json:"lastTransitionReason,omitempty"`
	CurrentSoftSignals      []string                       `json:"currentSoftSignals,omitempty"`
	CurrentHardSignals      []string                       `json:"currentHardSignals,omitempty"`
	CurrentRecoveryBlockers []string                       `json:"currentRecoveryBlockers,omitempty"`
	ReportRate              consumerRateSnapshot           `json:"reportRate"`
	RealTimeRate            consumerRateSnapshot           `json:"realTimeRate"`
	ReportPipeline          reportConsumerPipelineSnapshot `json:"reportPipeline"`
	ReportConsumerPool      util.WorkerPoolStats           `json:"reportConsumerPool"`
	ReportPersistPool       util.WorkerPoolStats           `json:"reportPersistPool"`
	PersistenceErrors       util.PersistenceErrorSnapshot  `json:"persistenceErrors"`
	MySQLHealth             db.DBHealthState               `json:"mysqlHealth"`
	Goroutines              int                            `json:"goroutines"`
	HeapAllocBytes          uint64                         `json:"heapAllocBytes"`
}

type protectSetRequest struct {
	Enabled                   *bool  `json:"enabled,omitempty"`
	ObserveOnly               *bool  `json:"observeOnly,omitempty"`
	SampleInterval            string `json:"sampleInterval,omitempty"`
	NormalRateLogInterval     string `json:"normalRateLogInterval,omitempty"`
	DiagnosticRateLogInterval string `json:"diagnosticRateLogInterval,omitempty"`
	SoftTargetRatePerSecond   *int   `json:"softTargetRatePerSecond,omitempty"`
}

type sleepRateLimiter struct {
	mutex         sync.Mutex
	enabled       bool
	interval      time.Duration
	nextAllowedAt time.Time
}

func (l *sleepRateLimiter) SetRate(targetRate int) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	if targetRate <= 0 {
		l.enabled = false
		l.interval = 0
		l.nextAllowedAt = time.Time{}
		return
	}

	l.enabled = true
	l.interval = time.Second / time.Duration(targetRate)
	if l.interval <= 0 {
		l.interval = time.Millisecond
	}
}

func (l *sleepRateLimiter) Disable() {
	l.SetRate(0)
}

func (l *sleepRateLimiter) Wait() {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if !l.enabled {
		return
	}

	now := time.Now()
	if l.nextAllowedAt.IsZero() || !now.Before(l.nextAllowedAt) {
		l.nextAllowedAt = now.Add(l.interval)
		return
	}

	sleepDuration := l.nextAllowedAt.Sub(now)
	l.nextAllowedAt = l.nextAllowedAt.Add(l.interval)
	time.Sleep(sleepDuration)
}

type reportConsumptionProtector struct {
	config              model.ProtectionConfig
	reportConsumer      regulatedConsumer
	realTimeConsumer    regulatedConsumer
	reportRateSampler   consumerSnapshotReader
	realTimeRateSampler consumerSnapshotReader
	reportPipeline      pipelineSnapshotReader
	reportConsumerPool  workerPoolController
	reportPersistPool   workerPoolController
	hostSampler         HostResourceSampler
	rateLimiter         *sleepRateLimiter

	mutex                sync.RWMutex
	enabled              bool
	observeOnly          bool
	state                protectionState
	softHoldUntil        time.Time
	hardHoldUntil        time.Time
	hardPauseWindows     int
	recentSpeeds         []float64
	softHitWindows       int
	hardHitWindows       int
	healthyWindows       int
	lastTransitionAt     time.Time
	lastTransitionReason string
	lastRecoveryAt       time.Time
	softHoldDuration     time.Duration
	hardHoldDuration     time.Duration
	hardPausedWindows    int
}

func newReportConsumptionProtector(
	config model.ProtectionConfig, reportConsumer regulatedConsumer, realTimeConsumer regulatedConsumer,
	reportRateSampler consumerSnapshotReader, realTimeRateSampler consumerSnapshotReader, reportPipeline pipelineSnapshotReader,
	reportConsumerPool workerPoolController, reportPersistPool workerPoolController,
) *reportConsumptionProtector {
	protectionConfig := config.Normalize()
	rateLimiter := &sleepRateLimiter{}
	rateLimiter.Disable()
	reportConsumer.SetConsumeRegulator(rateLimiter)

	return &reportConsumptionProtector{
		config:              protectionConfig,
		reportConsumer:      reportConsumer,
		realTimeConsumer:    realTimeConsumer,
		reportRateSampler:   reportRateSampler,
		realTimeRateSampler: realTimeRateSampler,
		reportPipeline:      reportPipeline,
		reportConsumerPool:  reportConsumerPool,
		reportPersistPool:   reportPersistPool,
		hostSampler:         noopHostResourceSampler{},
		rateLimiter:         rateLimiter,
		enabled:             protectionConfig.Enabled != nil && *protectionConfig.Enabled,
		observeOnly:         protectionConfig.ObserveOnly,
		state:               protectionStateNormal,
	}
}

func (p *reportConsumptionProtector) Start(stop <-chan struct{}) {
	if p == nil || stop == nil {
		return
	}

	go func() {
		for {
			timer := time.NewTimer(p.currentSampleInterval())
			select {
			case <-timer.C:
				p.sampleAndAct()
			case <-stop:
				if !timer.Stop() {
					select {
					case <-timer.C:
					default:
					}
				}
				return
			}
		}
	}()
}

func (p *reportConsumptionProtector) Enable() {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.enabled = true
}

func (p *reportConsumptionProtector) Disable() {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.enabled = false
	p.transitionLocked(protectionStateNormal, time.Now(), "disabled")
}

func (p *reportConsumptionProtector) Set(request protectSetRequest) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if request.Enabled != nil {
		p.enabled = *request.Enabled
	}
	if request.ObserveOnly != nil {
		p.observeOnly = *request.ObserveOnly
	}
	if request.SoftTargetRatePerSecond != nil {
		p.config.SoftTargetRatePerSecond = *request.SoftTargetRatePerSecond
	}
	if request.SampleInterval != "" {
		duration, err := time.ParseDuration(request.SampleInterval)
		if err != nil {
			return fmt.Errorf("sample interval 格式错误: %w", err)
		}
		p.config.SampleIntervalSeconds = durationToIntervalSeconds(duration)
	}
	if request.NormalRateLogInterval != "" {
		duration, err := time.ParseDuration(request.NormalRateLogInterval)
		if err != nil {
			return fmt.Errorf("normal rate log interval 格式错误: %w", err)
		}
		if duration <= 0 {
			return fmt.Errorf("normal rate log interval 必须大于 0")
		}
		p.reportRateSampler.SetLogIntervals(duration, 0)
		p.realTimeRateSampler.SetLogIntervals(duration, 0)
	}
	if request.DiagnosticRateLogInterval != "" {
		duration, err := time.ParseDuration(request.DiagnosticRateLogInterval)
		if err != nil {
			return fmt.Errorf("diagnostic rate log interval 格式错误: %w", err)
		}
		if duration <= 0 {
			return fmt.Errorf("diagnostic rate log interval 必须大于 0")
		}
		p.reportRateSampler.SetLogIntervals(0, duration)
		p.realTimeRateSampler.SetLogIntervals(0, duration)
	}
	return nil
}

func (p *reportConsumptionProtector) Status() protectionStatusResponse {
	var mem runtime.MemStats
	readRuntimeMemStats(&mem)

	var (
		mysqlHealth       = db.DBHealthState{}
		persistenceErrors = readPersistenceErrorSnapshot()
	)
	mysqlHealth, _ = getMySQLHealthState("mysql")

	p.mutex.RLock()
	defer p.mutex.RUnlock()

	status := protectionStatusResponse{
		Enabled:              p.enabled,
		ObserveOnly:          p.observeOnly,
		State:                p.state,
		SoftHoldUntil:        p.softHoldUntil,
		HardHoldUntil:        p.hardHoldUntil,
		LastTransitionAt:     p.lastTransitionAt,
		LastTransitionReason: p.lastTransitionReason,
		ReportRate:           p.reportRateSampler.Snapshot(),
		RealTimeRate:         p.realTimeRateSampler.Snapshot(),
		ReportPipeline:       p.reportPipeline.Snapshot(),
		ReportConsumerPool:   p.reportConsumerPool.Stats(),
		ReportPersistPool:    p.reportPersistPool.Stats(),
		PersistenceErrors:    persistenceErrors,
		MySQLHealth:          mysqlHealth,
		Goroutines:           currentGoroutineNum(),
		HeapAllocBytes:       mem.HeapAlloc,
	}
	p.applyMockLocked(&status)
	status.CurrentSoftSignals = p.softSignalsLocked(status)
	status.CurrentHardSignals = p.hardSignalsLocked(status)
	status.CurrentRecoveryBlockers = p.recoveryBlockersLocked(status, time.Now())
	return status
}

func (p *reportConsumptionProtector) currentSampleInterval() time.Duration {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	return p.currentSampleIntervalLocked()
}

func (p *reportConsumptionProtector) currentSampleIntervalLocked() time.Duration {
	interval := time.Duration(p.config.SampleIntervalSeconds) * time.Second
	if interval <= 0 {
		return time.Second
	}
	return interval
}

func (p *reportConsumptionProtector) sampleAndAct() {
	if p == nil {
		return
	}

	status := p.Status()

	p.mutex.Lock()
	defer p.mutex.Unlock()

	if !p.enabled {
		return
	}

	if p.config.Mock.Enabled && p.config.Mock.ReportSpeedPerSecond != nil {
		p.recentSpeeds = nil
	} else {
		p.recentSpeeds = append(p.recentSpeeds, status.ReportRate.SpeedPerSecond)
		if len(p.recentSpeeds) > 3 {
			p.recentSpeeds = p.recentSpeeds[len(p.recentSpeeds)-3:]
		}
	}

	now := time.Now()
	hardSignals := p.hardSignalsLocked(status)
	softSignals := p.softSignalsLocked(status)
	recoveryBlockers := p.recoveryBlockersLocked(status, now)
	if len(hardSignals) > 0 {
		p.hardHitWindows++
		p.softHitWindows = 0
		p.healthyWindows = 0
		if p.state == protectionStateHardPaused {
			p.hardPausedWindows++
		}
	} else if len(softSignals) >= 2 {
		p.softHitWindows++
		p.hardHitWindows = 0
		p.healthyWindows = 0
		p.hardPausedWindows = 0
	} else if len(recoveryBlockers) == 0 {
		p.healthyWindows++
		p.softHitWindows = 0
		p.hardHitWindows = 0
		p.hardPausedWindows = 0
	} else {
		p.softHitWindows = 0
		p.hardHitWindows = 0
		p.healthyWindows = 0
		p.hardPausedWindows = 0
	}

	if p.hardHitWindows >= 3 {
		if p.state == protectionStateHardPaused && !p.observeOnly && p.hardPausedWindows >= p.config.HardEscalationWindows {
			p.realTimeConsumer.PauseConsumption()
		}
		p.transitionLocked(protectionStateHardPaused, now, "hard thresholds reached: "+strings.Join(hardSignals, ","))
		return
	}
	if p.softHitWindows >= 3 {
		p.transitionLocked(protectionStateSoftLimited, now, "soft thresholds reached: "+strings.Join(softSignals, ","))
		return
	}
	if p.healthyWindows >= p.config.RecoveryHealthyWindows {
		switch p.state {
		case protectionStateHardPaused:
			p.transitionLocked(protectionStateSoftLimited, now, "healthy windows recovered from hard pause")
		case protectionStateSoftLimited:
			p.transitionLocked(protectionStateNormal, now, "healthy windows recovered to normal")
		}
	}
}

func (p *reportConsumptionProtector) softSignalsLocked(status protectionStatusResponse) []string {
	signals := make([]string, 0, 8)
	if int64(status.ReportPipeline.PendingCount) >= p.config.SoftThresholds.OrderedCommitPendingCount {
		signals = append(signals, "ordered_commit_pending")
	}
	if status.ReportPipeline.Gate.InFlightMessages >= p.config.SoftThresholds.GateInFlightMessages {
		signals = append(signals, "gate_in_flight")
	}
	if status.ReportPipeline.Gate.WaitingTasks >= p.config.SoftThresholds.GateWaitingTasks {
		signals = append(signals, "gate_waiting_tasks")
	}
	if p.isReportConsumerPoolSaturatedLocked(status) {
		signals = append(signals, "report_consumer_pool_saturated")
	}
	if status.ReportPersistPool.QueueUsageRatio >= float64(p.config.SoftThresholds.WorkerQueueUsagePermille)/1000 &&
		status.ReportPersistPool.BusyRatio >= float64(p.config.SoftThresholds.WorkerBusyPermille)/1000 {
		signals = append(signals, "report_persist_pool_saturated")
	}
	if status.HeapAllocBytes >= p.config.SoftThresholds.HeapAllocBytes {
		signals = append(signals, "heap_alloc_high")
	}
	if status.Goroutines >= p.effectiveSoftGoroutineThresholdLocked(status) && p.shouldTreatHighGoroutinesAsPressureLocked(status) {
		signals = append(signals, "goroutines_high")
	}
	if status.PersistenceErrors.CountLastMinute >= p.config.SoftThresholds.PersistenceErrors {
		signals = append(signals, "persistence_errors")
	}
	if p.isRateTrendWorseningLocked() {
		signals = append(signals, "rate_trend_worsening")
	}
	return signals
}

// shouldTreatHighGoroutinesAsPressureLocked 用于避免“goroutine 总数高，但消费链路实际上空闲”时误触发保护暂停。
//
// 线上真实场景：
//  1. Kafka topic 分区很多时，consumer group 自身就会常驻大量 goroutine。
//  2. 如果此时 pipeline/pool/persistence 都是空闲的，仅凭 goroutine 数高就 hard_paused，
//     会把消费永久冻住，lag 只能在重启后的短暂窗口内下降。
//  3. 因此这里要求 goroutines_high 必须和“当前确实存在处理压力”同时出现，才参与状态机判定。
func (p *reportConsumptionProtector) shouldTreatHighGoroutinesAsPressureLocked(status protectionStatusResponse) bool {
	var (
		pendingThreshold  = maxInt(500, int(p.config.SoftThresholds.OrderedCommitPendingCount/40))
		inFlightThreshold = maxInt64(500, p.config.SoftThresholds.GateInFlightMessages/40)
		waitingThreshold  = maxInt64(1000, p.config.SoftThresholds.GateWaitingTasks/40)
	)

	return status.ReportPipeline.PendingCount >= pendingThreshold ||
		status.ReportPipeline.Gate.InFlightMessages >= inFlightThreshold ||
		status.ReportPipeline.Gate.WaitingTasks >= waitingThreshold ||
		status.ReportConsumerPool.QueueUsageRatio >= 0.1 ||
		status.ReportPersistPool.QueueUsageRatio >= 0.1 ||
		(status.ReportConsumerPool.BusyRatio >= 0.8 && status.ReportPipeline.PendingCount > 0) ||
		(status.ReportPersistPool.BusyRatio >= 0.8 && status.ReportPipeline.Gate.WaitingTasks > 0) ||
		status.PersistenceErrors.CountLastMinute > 0
}

// estimatedKafkaClaimGoroutineBaselineLocked 估算当前部署形态下 Kafka claim 常驻 goroutine 的正常基线。
//
// 估算方式：
// 1. 直接读取 Kafka SDK 采样得到的 topic 实际分区数；
// 2. 对当前 sinker 同时运行的 consumer group 分别累加；
// 3. 再乘以 Sarama 每个 partition-group claim 常驻的 goroutine 组数。
//
// 例如 300 分区、2 个 group 时，基线大约就是 300 * 2 * 7 = 4200。
func (p *reportConsumptionProtector) estimatedKafkaClaimGoroutineBaselineLocked(status protectionStatusResponse) int {
	totalPartitions := 0
	if status.ReportRate.PartitionCount > 0 {
		totalPartitions += status.ReportRate.PartitionCount
	}
	if status.RealTimeRate.PartitionCount > 0 {
		totalPartitions += status.RealTimeRate.PartitionCount
	}
	if totalPartitions <= 0 {
		// 如果 Kafka 采样窗口尚未刷新到有效分区数，退回到当前运行态里已经注册过的 report committer 数。
		totalPartitions = status.ReportPipeline.CommitterCount
	}
	if totalPartitions <= 0 {
		return 0
	}

	return totalPartitions * kafkaClaimGoroutinesPerPartitionGroup
}

func (p *reportConsumptionProtector) effectiveSoftGoroutineThresholdLocked(status protectionStatusResponse) int {
	baseline := p.estimatedKafkaClaimGoroutineBaselineLocked(status)
	if baseline <= 0 {
		return p.config.SoftThresholds.Goroutines
	}

	headroom := maxInt(kafkaClaimGoroutineSoftHeadroomMin, baseline*kafkaClaimGoroutineSoftHeadroomPermille/1000)
	return maxInt(p.config.SoftThresholds.Goroutines, baseline+headroom)
}

func (p *reportConsumptionProtector) effectiveHardGoroutineThresholdLocked(status protectionStatusResponse) int {
	baseline := p.estimatedKafkaClaimGoroutineBaselineLocked(status)
	if baseline <= 0 {
		return p.config.HardThresholds.Goroutines
	}

	headroom := maxInt(kafkaClaimGoroutineHardHeadroomMin, baseline*kafkaClaimGoroutineHardHeadroomPermille/1000)
	return maxInt(p.config.HardThresholds.Goroutines, baseline+headroom)
}

// isReportConsumerPoolSaturatedLocked 把“消费池饱和”定义成更贴近真实高压的状态：
// 1. worker 忙碌度已经很高
// 2. 并且要么队列真的堆起来，要么 pending/gate 已经在明显积压
//
// 示例：
// 1. busy=0.95, queue=0.60 -> 直接视为饱和
// 2. busy=1.00, queue=0, pending=1500, gate=1500 -> 也视为饱和
// 3. busy=0.40, queue=0, pending=1500 -> 不算饱和
func (p *reportConsumptionProtector) isReportConsumerPoolSaturatedLocked(status protectionStatusResponse) bool {
	if status.ReportConsumerPool.BusyRatio < float64(p.config.SoftThresholds.WorkerBusyPermille)/1000 {
		return false
	}

	if status.ReportConsumerPool.QueueUsageRatio >= float64(p.config.SoftThresholds.WorkerQueueUsagePermille)/1000 {
		return true
	}

	var (
		consumerCapacity = maxInt(1, status.ReportConsumerPool.Capacity)
		pendingThreshold = maxInt(500, minLocalInt(int(p.config.SoftThresholds.OrderedCommitPendingCount/10), consumerCapacity*200))
		gateThreshold    = maxInt64(500, minLocalInt64(p.config.SoftThresholds.GateInFlightMessages/10, int64(consumerCapacity*200)))
	)
	return status.ReportPipeline.PendingCount >= pendingThreshold &&
		status.ReportPipeline.Gate.InFlightMessages >= gateThreshold
}

func (p *reportConsumptionProtector) hardSignalsLocked(status protectionStatusResponse) []string {
	signals := make([]string, 0, 8)
	if int64(status.ReportPipeline.PendingCount) >= p.config.HardThresholds.OrderedCommitPendingCount {
		signals = append(signals, "ordered_commit_pending")
	}
	if status.ReportPipeline.Gate.InFlightMessages >= p.config.HardThresholds.GateInFlightMessages {
		signals = append(signals, "gate_in_flight")
	}
	if status.ReportPipeline.Gate.WaitingTasks >= p.config.HardThresholds.GateWaitingTasks {
		signals = append(signals, "gate_waiting_tasks")
	}
	if status.ReportConsumerPool.QueueUsageRatio >= float64(p.config.HardThresholds.WorkerQueueUsagePermille)/1000 {
		signals = append(signals, "report_consumer_pool_queue")
	}
	if status.ReportPersistPool.QueueUsageRatio >= float64(p.config.HardThresholds.WorkerQueueUsagePermille)/1000 &&
		status.ReportPersistPool.BusyRatio >= float64(p.config.HardThresholds.WorkerBusyPermille)/1000 {
		signals = append(signals, "report_persist_pool_saturated")
	}
	if status.HeapAllocBytes >= p.config.HardThresholds.HeapAllocBytes {
		signals = append(signals, "heap_alloc_high")
	}
	if status.Goroutines >= p.effectiveHardGoroutineThresholdLocked(status) && p.shouldTreatHighGoroutinesAsPressureLocked(status) {
		signals = append(signals, "goroutines_high")
	}
	if status.PersistenceErrors.CountLastMinute >= p.config.HardThresholds.PersistenceErrors {
		signals = append(signals, "persistence_errors")
	}
	if status.MySQLHealth.Status == "degraded" && status.MySQLHealth.ConsecutiveFailures >= p.config.HardThresholds.DBConsecutiveFailures {
		signals = append(signals, "mysql_degraded")
	}
	return signals
}

func (p *reportConsumptionProtector) recoveryBlockersLocked(status protectionStatusResponse, now time.Time) []string {
	blockers := make([]string, 0, 8)
	switch p.state {
	case protectionStateNormal:
		return blockers
	case protectionStateSoftLimited:
		if now.Before(p.softHoldUntil) {
			blockers = append(blockers, "soft_hold")
		}
	case protectionStateHardPaused:
		if now.Before(p.hardHoldUntil) {
			blockers = append(blockers, "hard_hold")
		}
	}

	if int64(status.ReportPipeline.PendingCount) >= 5000 {
		blockers = append(blockers, "ordered_commit_pending")
	}
	if status.ReportPipeline.Gate.InFlightMessages >= 5000 {
		blockers = append(blockers, "gate_in_flight")
	}
	if status.ReportPipeline.Gate.WaitingTasks >= 10000 {
		blockers = append(blockers, "gate_waiting_tasks")
	}
	if status.ReportConsumerPool.QueueUsageRatio >= 0.1 {
		blockers = append(blockers, "report_consumer_pool_queue")
	}
	if status.ReportPersistPool.QueueUsageRatio >= 0.1 {
		blockers = append(blockers, "report_persist_pool_queue")
	}
	if p.isReportConsumerPoolSaturatedLocked(status) {
		blockers = append(blockers, "report_consumer_pool_saturated")
	}
	if status.HeapAllocBytes >= 512<<20 {
		blockers = append(blockers, "heap_alloc_high")
	}
	if status.Goroutines >= p.effectiveSoftGoroutineThresholdLocked(status) && p.shouldTreatHighGoroutinesAsPressureLocked(status) {
		blockers = append(blockers, "goroutines_high")
	}
	if status.PersistenceErrors.CountLastMinute > 0 {
		blockers = append(blockers, "persistence_errors")
	}
	if status.MySQLHealth.Status == "degraded" {
		blockers = append(blockers, "mysql_degraded")
	}
	if p.isRateTrendWorseningLocked() {
		blockers = append(blockers, "rate_trend_worsening")
	}
	return blockers
}

func (p *reportConsumptionProtector) isRateTrendWorseningLocked() bool {
	if len(p.recentSpeeds) < 3 {
		return false
	}
	return p.recentSpeeds[0] > p.recentSpeeds[1] && p.recentSpeeds[1] > p.recentSpeeds[2]
}

func (p *reportConsumptionProtector) transitionLocked(next protectionState, now time.Time, reason string) {
	if p.state == next {
		return
	}

	prev := p.state
	p.state = next
	switch next {
	case protectionStateNormal:
		p.softHoldUntil = time.Time{}
		p.hardHoldUntil = time.Time{}
		p.softHoldDuration = 0
		p.hardHoldDuration = 0
		p.hardPauseWindows = 0
		p.hardPausedWindows = 0
		p.lastRecoveryAt = now
		p.rateLimiter.Disable()
		if !p.observeOnly {
			p.reportConsumer.ResumeConsumption()
			p.realTimeConsumer.ResumeConsumption()
			p.reportConsumerPool.ResetRuntimeBounds()
			p.reportPersistPool.ResetRuntimeBounds()
		}
	case protectionStateSoftLimited:
		p.softHoldDuration = p.nextHoldDurationLocked(time.Duration(p.config.SoftMinHoldSeconds)*time.Second, p.softHoldDuration, now)
		p.softHoldUntil = now.Add(p.softHoldDuration)
		p.hardHoldUntil = time.Time{}
		p.hardPauseWindows = 0
		p.hardPausedWindows = 0
		p.rateLimiter.SetRate(p.config.SoftTargetRatePerSecond)
		if prev == protectionStateHardPaused {
			p.lastRecoveryAt = now
		}
		if !p.observeOnly {
			p.reportConsumer.ResumeConsumption()
			p.realTimeConsumer.ResumeConsumption()
			p.reportConsumerPool.SetRuntimeBounds(p.reportConsumerPool.Stats().MinWorkers, p.reportConsumerPool.Stats().MinWorkers)
			p.reportPersistPool.SetRuntimeBounds(p.reportPersistPool.Stats().MinWorkers, p.reportPersistPool.Stats().MinWorkers)
		}
	case protectionStateHardPaused:
		p.hardPauseWindows++
		p.hardPausedWindows = 0
		p.hardHoldDuration = p.nextHoldDurationLocked(time.Duration(p.config.HardMinHoldSeconds)*time.Second, p.hardHoldDuration, now)
		p.hardHoldUntil = now.Add(p.hardHoldDuration)
		p.rateLimiter.Disable()
		if !p.observeOnly {
			p.reportConsumer.PauseConsumption()
			if p.hardPauseWindows >= p.config.HardEscalationWindows {
				p.realTimeConsumer.PauseConsumption()
			}
		}
	}
	p.lastTransitionAt = now
	p.lastTransitionReason = reason

	logs.Logger.Warn("report consumption protector state changed",
		zap.String("from", string(prev)),
		zap.String("to", string(next)),
		zap.Bool("observe_only", p.observeOnly),
		zap.String("reason", reason),
	)
}

func (p *reportConsumptionProtector) nextHoldDurationLocked(base, previous time.Duration, now time.Time) time.Duration {
	hold := base
	if base <= 0 {
		return 0
	}

	if !p.lastRecoveryAt.IsZero() && now.Sub(p.lastRecoveryAt) <= 2*p.currentSampleIntervalLocked() {
		hold = previous
		if hold < base {
			hold = base
		}
		hold *= 2
	}

	maxAdaptive := time.Duration(p.config.MaxAdaptiveHoldSeconds) * time.Second
	if maxAdaptive > 0 && hold > maxAdaptive {
		hold = maxAdaptive
	}
	return hold
}

func durationToIntervalSeconds(duration time.Duration) int {
	if duration <= 0 {
		return 1
	}

	return int((duration + time.Second - 1) / time.Second)
}

func maxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func minLocalInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func minLocalInt64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func (p *reportConsumptionProtector) applyMockLocked(status *protectionStatusResponse) {
	if p == nil || status == nil {
		return
	}

	mock := p.config.Mock
	if !mock.Enabled {
		return
	}

	if mock.ReportLag != nil {
		status.ReportRate.Lag = *mock.ReportLag
	}
	if mock.ReportSpeedPerSecond != nil {
		status.ReportRate.SpeedPerSecond = *mock.ReportSpeedPerSecond
	}
	if mock.RealTimeLag != nil {
		status.RealTimeRate.Lag = *mock.RealTimeLag
	}
	if mock.RealTimeSpeedPerSecond != nil {
		status.RealTimeRate.SpeedPerSecond = *mock.RealTimeSpeedPerSecond
	}
	if mock.PipelinePending != nil {
		status.ReportPipeline.PendingCount = *mock.PipelinePending
	}
	if mock.GateInFlight != nil {
		status.ReportPipeline.Gate.InFlightMessages = *mock.GateInFlight
	}
	if mock.GateWaitingTasks != nil {
		status.ReportPipeline.Gate.WaitingTasks = *mock.GateWaitingTasks
	}
	if mock.ReportConsumerQueueUsagePermille != nil {
		status.ReportConsumerPool.QueueUsageRatio = float64(*mock.ReportConsumerQueueUsagePermille) / 1000
	}
	if mock.ReportConsumerBusyPermille != nil {
		status.ReportConsumerPool.BusyRatio = float64(*mock.ReportConsumerBusyPermille) / 1000
	}
	if mock.ReportPersistQueueUsagePermille != nil {
		status.ReportPersistPool.QueueUsageRatio = float64(*mock.ReportPersistQueueUsagePermille) / 1000
	}
	if mock.ReportPersistBusyPermille != nil {
		status.ReportPersistPool.BusyRatio = float64(*mock.ReportPersistBusyPermille) / 1000
	}
	if mock.HeapAllocBytes != nil {
		status.HeapAllocBytes = *mock.HeapAllocBytes
	}
	if mock.Goroutines != nil {
		status.Goroutines = *mock.Goroutines
	}
	if mock.PersistenceErrorCount != nil {
		status.PersistenceErrors.CountLastMinute = *mock.PersistenceErrorCount
		if *mock.PersistenceErrorCount == 0 {
			status.PersistenceErrors.LastClass = ""
			status.PersistenceErrors.LastError = ""
			status.PersistenceErrors.LastOccurredAt = time.Time{}
		} else if status.PersistenceErrors.LastClass == "" {
			status.PersistenceErrors.LastClass = "mock_persistence_error"
		}
		if *mock.PersistenceErrorCount > 0 && status.PersistenceErrors.LastError == "" {
			status.PersistenceErrors.LastError = "mock persistence error injected"
		}
		if *mock.PersistenceErrorCount > 0 && status.PersistenceErrors.LastOccurredAt.IsZero() {
			status.PersistenceErrors.LastOccurredAt = time.Now()
		}
	}
	if mock.MySQLStatus != nil {
		status.MySQLHealth.Status = *mock.MySQLStatus
	}
	if mock.MySQLConsecutiveFailures != nil {
		status.MySQLHealth.ConsecutiveFailures = *mock.MySQLConsecutiveFailures
	}
}
