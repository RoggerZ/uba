package runner

import (
	"sync"
	"time"

	"github.com/1340691923/xwl_bi/engine/logs"
	"github.com/1340691923/xwl_bi/model"
	"github.com/1340691923/xwl_bi/platform-basic-libs/sinker"
	"github.com/1340691923/xwl_bi/platform-basic-libs/util"
	"github.com/IBM/sarama"
	"go.uber.org/zap"
)

type consumerRateSnapshot struct {
	Group          string    `json:"group"`
	Topic          string    `json:"topic"`
	PartitionCount int       `json:"partitionCount"`
	CurrentOffset  int64     `json:"currentOffset"`
	LogEndOffset   int64     `json:"logEndOffset"`
	Lag            int64     `json:"lag"`
	DeltaLag       int64     `json:"deltaLag"`
	SpeedPerSecond float64   `json:"speedPerSecond"`
	SampledAt      time.Time `json:"sampledAt"`
	SamplingMode   string    `json:"samplingMode"`
}

type consumerOffsetReader interface {
	CurrentMarkedOffset() int64
	CurrentMarkedOffsets() map[int32]int64
}

type consumerGroupOffsetAdmin interface {
	ListConsumerGroupOffsets(group string, topicPartitions map[string][]int32) (*sarama.OffsetFetchResponse, error)
	Close() error
}

type consumerRateSampler struct {
	group    string
	topic    string
	consumer consumerOffsetReader
	client   sarama.Client
	admin    consumerGroupOffsetAdmin
	// sampleInterval 控制底层 offset/lag 采样频率。
	// 保护状态机、速率趋势和 status 接口快照都基于这个窗口更新。
	// 例如 sampleInterval=5s 时，内部每 5 秒刷新一次 currentOffset/logEndOffset/lag。
	sampleInterval time.Duration
	// normalLogInterval 控制常驻速率日志输出节奏，不影响内部采样频率。
	// 例如 sampleInterval=5s、normalLogInterval=3m 时，快照仍每 5 秒更新，
	// 但平时只会每 3 分钟输出一条 sinker consumer rate 日志。
	normalLogInterval time.Duration
	// diagnosticInterval 是诊断窗口开启后的默认日志间隔。
	// 当诊断会话没有显式指定 rateLogInterval 时，采样器回退到这个值。
	// 例如 diagnosticInterval=5s 时，diagnostic enable 后会把速率日志提频到每 5 秒。
	diagnosticInterval time.Duration

	mutex      sync.RWMutex
	lastSample consumerRateSnapshot
	lastLogAt  time.Time
}

func (s *consumerRateSampler) SetLogIntervals(normal time.Duration, diagnostic time.Duration) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if normal > 0 {
		s.normalLogInterval = normal
	}
	if diagnostic > 0 {
		s.diagnosticInterval = diagnostic
	}
}

func newConsumerRateSampler(group, topic string, consumer *sinker.KafkaSarama, cfg model.KafkaCfg, protection model.ProtectionConfig) (*consumerRateSampler, error) {
	saramaConfig, err := sinker.GetSaramaConfig(cfg)
	if err != nil {
		return nil, err
	}

	client, err := sarama.NewClient(cfg.Addresses, saramaConfig)
	if err != nil {
		return nil, err
	}

	admin, err := sarama.NewClusterAdmin(cfg.Addresses, saramaConfig)
	if err != nil {
		logs.Logger.Warn("consumer rate sampler create cluster admin failed",
			zap.String("group", group),
			zap.String("topic", topic),
			zap.Error(err),
		)
	}

	return &consumerRateSampler{
		group:              group,
		topic:              topic,
		consumer:           consumer,
		client:             client,
		admin:              admin,
		sampleInterval:     time.Duration(protection.SampleIntervalSeconds) * time.Second,
		normalLogInterval:  time.Duration(protection.NormalRateLogIntervalSeconds) * time.Second,
		diagnosticInterval: time.Duration(protection.DefaultDiagnosticRateLogIntervalSeconds) * time.Second,
	}, nil
}

func (s *consumerRateSampler) Close() error {
	if s == nil {
		return nil
	}

	var closeErr error
	if s.admin != nil {
		if err := s.admin.Close(); err != nil {
			closeErr = err
		}
	}
	if s.client != nil {
		if err := s.client.Close(); err != nil && closeErr == nil {
			closeErr = err
		}
	}
	return closeErr
}

func (s *consumerRateSampler) Snapshot() consumerRateSnapshot {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.lastSample
}

func (s *consumerRateSampler) Start(stop <-chan struct{}) {
	if s == nil || stop == nil {
		return
	}

	go func() {
		ticker := time.NewTicker(s.sampleInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.sampleAndLog()
			case <-stop:
				return
			}
		}
	}()
}

func (s *consumerRateSampler) sampleAndLog() {
	now := time.Now()
	partitions, err := s.client.Partitions(s.topic)
	if err != nil {
		logs.Logger.Warn("consumer rate sample query partitions failed",
			zap.String("group", s.group),
			zap.String("topic", s.topic),
			zap.Error(err),
		)
		return
	}

	var logEndOffset int64
	for _, partition := range partitions {
		offset, err := s.client.GetOffset(s.topic, partition, sarama.OffsetNewest)
		if err != nil {
			logs.Logger.Warn("consumer rate sample query offset failed",
				zap.String("group", s.group),
				zap.String("topic", s.topic),
				zap.Int32("partition", partition),
				zap.Error(err),
			)
			return
		}
		logEndOffset += offset
	}

	currentOffset := s.currentOffset(partitions)
	lag := logEndOffset - currentOffset
	if lag < 0 {
		lag = 0
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	snapshot := consumerRateSnapshot{
		Group:          s.group,
		Topic:          s.topic,
		PartitionCount: len(partitions),
		CurrentOffset:  currentOffset,
		LogEndOffset:   logEndOffset,
		Lag:            lag,
		SampledAt:      now,
	}
	if !s.lastSample.SampledAt.IsZero() {
		snapshot.DeltaLag = s.lastSample.Lag - lag
		deltaSeconds := now.Sub(s.lastSample.SampledAt).Seconds()
		if deltaSeconds > 0 {
			snapshot.SpeedPerSecond = float64(snapshot.DeltaLag) / deltaSeconds
		}
	}

	snapshot.SamplingMode = s.currentSamplingMode()
	s.lastSample = snapshot

	if s.shouldLog(now, snapshot.SamplingMode) {
		s.lastLogAt = now
		logs.Logger.Info("sinker consumer rate",
			zap.String("group", snapshot.Group),
			zap.String("topic", snapshot.Topic),
			zap.Int("partition_count", snapshot.PartitionCount),
			zap.Int64("current_offset", snapshot.CurrentOffset),
			zap.Int64("log_end_offset", snapshot.LogEndOffset),
			zap.Int64("lag", snapshot.Lag),
			zap.Int64("delta_lag", snapshot.DeltaLag),
			zap.Float64("speed_per_sec", snapshot.SpeedPerSecond),
			zap.String("sampling_mode", snapshot.SamplingMode),
		)
	}
}

// currentOffset 返回当前采样窗口下更接近真实消费进度的 offset。
//
// 优先级：
// 1. 优先读取 consumer group 已提交的 committed offset。
// 2. 如果当前进程内已经 mark 了更靠前的 offset，则按 partition 取更大的值。
// 3. 如果 admin 查询失败或没有有效 committed offset，就回退到进程内 mark 快照。
//
// 这样可以同时解决两个问题：
// - 冷启动后尚未再次消费时，CurrentMarkedOffsets 为空，但 group committed offset 实际已经很高。
// - 正常运行中，进程内 mark 可能暂时领先于 coordinator 上已提交的 offset。
func (s *consumerRateSampler) currentOffset(partitions []int32) int64 {
	markedOffsets := s.consumer.CurrentMarkedOffsets()
	committedOffsets, ok := s.currentCommittedOffsets(partitions)
	if !ok {
		return sumPartitionOffsets(markedOffsets, partitions)
	}

	var current int64
	for _, partition := range partitions {
		committedOffset := committedOffsets[partition]
		markedOffset := markedOffsets[partition]
		if markedOffset > committedOffset {
			current += markedOffset
			continue
		}
		current += committedOffset
	}
	return current
}

func (s *consumerRateSampler) currentCommittedOffsets(partitions []int32) (map[int32]int64, bool) {
	if s.admin == nil {
		return nil, false
	}

	response, err := s.admin.ListConsumerGroupOffsets(s.group, map[string][]int32{
		s.topic: partitions,
	})
	if err != nil {
		logs.Logger.Warn("consumer rate sample query committed offsets failed",
			zap.String("group", s.group),
			zap.String("topic", s.topic),
			zap.Error(err),
		)
		return nil, false
	}

	offsets := make(map[int32]int64, len(partitions))
	found := false
	for _, partition := range partitions {
		block := response.GetBlock(s.topic, partition)
		if block == nil {
			continue
		}
		if block.Err != sarama.ErrNoError {
			logs.Logger.Warn("consumer rate sample query committed offset block failed",
				zap.String("group", s.group),
				zap.String("topic", s.topic),
				zap.Int32("partition", partition),
				zap.String("error_code", block.Err.Error()),
			)
			return nil, false
		}
		if block.Offset < 0 {
			continue
		}
		offsets[partition] = block.Offset
		found = true
	}
	return offsets, found
}

func sumPartitionOffsets(offsets map[int32]int64, partitions []int32) int64 {
	var total int64
	for _, partition := range partitions {
		total += offsets[partition]
	}
	return total
}

func (s *consumerRateSampler) currentSamplingMode() string {
	session := util.CurrentSinkerDiagnosticSession()
	if !session.Enabled {
		return "normal"
	}
	return "diagnostic"
}

// effectiveLogInterval 返回当前应当使用的速率日志输出间隔。
// 规则分三层：
// 1. 诊断会话显式指定了 rateLogInterval，则优先使用会话值。
// 2. 诊断会话已开启但未指定速率间隔，则使用 diagnosticInterval。
// 3. 平时模式回落到 normalLogInterval。
// 示例：
// - 平时：normal=3m, diagnostic=5s => 返回 3m
// - 诊断开启且 request.rateLogInterval=10s => 返回 10s
// - 诊断开启但未传 rateLogInterval => 返回 5s
func (s *consumerRateSampler) effectiveLogInterval() time.Duration {
	session := util.CurrentSinkerDiagnosticSession()
	if session.Enabled {
		if session.RateLogInterval > 0 {
			return session.RateLogInterval
		}
		return s.diagnosticInterval
	}
	return s.normalLogInterval
}

// shouldLog 基于上次输出时间和当前生效间隔判断这次采样是否需要打印日志。
// 它只决定“是否输出”，不影响 sampleAndLog 对 lastSample 的刷新。
// 示例：
// - 首次采样时 lastLogAt 为空，一定返回 true
// - 上次输出距今 2m、effectiveLogInterval=3m，则返回 false
// - 诊断模式下上次输出距今 6s、effectiveLogInterval=5s，则返回 true
func (s *consumerRateSampler) shouldLog(now time.Time, samplingMode string) bool {
	if s.lastLogAt.IsZero() {
		return true
	}
	return now.Sub(s.lastLogAt) >= s.effectiveLogInterval()
}
