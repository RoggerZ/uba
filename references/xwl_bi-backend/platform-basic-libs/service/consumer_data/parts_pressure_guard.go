package consumer_data

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/1340691923/xwl_bi/engine/logs"
	"github.com/1340691923/xwl_bi/platform-basic-libs/util"
	"go.uber.org/zap"
)

const (
	partsPressureThresholdRatio = 0.85
	partsPressureCooldown       = 30 * time.Second
	tooManyPartsCooldown        = 120 * time.Second
)

type PartsPressureTopPartition struct {
	PartitionID string `json:"partition_id"`
	Parts       int64  `json:"parts"`
}

type PartsPressureSnapshot struct {
	Table              string                      `json:"table"`
	ActiveParts        int64                       `json:"active_parts"`
	MaxPartsInTotal    int64                       `json:"max_parts_in_total"`
	PartsToThrowInsert int64                       `json:"parts_to_throw_insert"`
	TopPartitions      []PartsPressureTopPartition `json:"top_partitions"`
	SampledAt          time.Time                   `json:"sampled_at"`
}

type deferredFlushError struct {
	table  string
	until  time.Time
	reason string
}

func (e *deferredFlushError) Error() string {
	return fmt.Sprintf("%s flush deferred until %s: %s", e.table, e.until.Format(time.RFC3339), e.reason)
}

func IsDeferredFlushError(err error) bool {
	var target *deferredFlushError
	return errors.As(err, &target)
}

// IsTooManyPartsError 用于识别 ClickHouse 返回的 parts 过载错误。
//
// 当前同时兼容：
// 1. 文本包含 `Too many parts`
// 2. 错误码包含 `code: 252`
func IsTooManyPartsError(err error) bool {
	if err == nil {
		return false
	}

	message := err.Error()
	return strings.Contains(message, "Too many parts") || strings.Contains(message, "code: 252")
}

// PartsPressureGuard 负责把 ClickHouse parts 观测结果转成批量器冷却保护。
//
// 设计目标：
//  1. 采样线程负责查 system.parts / settings。
//  2. Flush 热路径只读取最近一次快照，不直接查系统表。
//  3. 一旦 active parts 已接近阈值，或者已经收到 Too many parts 错误，
//     就对对应表进入一段冷却期，避免继续放大 parts 压力。
type PartsPressureGuard struct {
	tableName     string
	mutex         sync.RWMutex
	snapshot      PartsPressureSnapshot
	cooldownUntil time.Time
	lastError     string
	lastErrorAt   time.Time
	bypass        bool
}

func NewPartsPressureGuard(tableName string) *PartsPressureGuard {
	return &PartsPressureGuard{tableName: tableName}
}

// UpdateSnapshot 用最新采样结果刷新 guard 视角里的 parts 压力快照。
func (g *PartsPressureGuard) UpdateSnapshot(snapshot PartsPressureSnapshot) {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	g.snapshot = snapshot
}

// SetBypass 用于进程退出阶段临时绕过 guard。
//
// 示例：
// 1. 正常运行时 guard 生效，避免继续打爆 ClickHouse。
// 2. 程序收到退出信号后，FlushAll 需要尽量冲刷剩余内存数据，此时可以短暂 bypass。
func (g *PartsPressureGuard) SetBypass(enabled bool) {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	g.bypass = enabled
}

// BeforeFlush 在真正执行 flush 之前，根据最近一次快照决定是否需要进入/保持冷却。
//
// 关键规则：
// 1. active_parts 达到 `85% * max_parts_in_total` 时，进入短冷却。
// 2. 冷却窗口内直接返回 deferredFlushError，不开事务、不访问 ClickHouse。
// 3. 如果已经 bypass，则完全跳过 guard。
//
// 这里返回 error 而不是 bool，是为了让上层批量器能复用现有错误通道，
// 同时又能通过 `IsDeferredFlushError` 区分“延后写入”和“真实失败”。
func (g *PartsPressureGuard) BeforeFlush(bufferLength int) error {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	if g.bypass {
		return nil
	}

	now := time.Now()
	if g.snapshot.MaxPartsInTotal > 0 && float64(g.snapshot.ActiveParts) >= float64(g.snapshot.MaxPartsInTotal)*partsPressureThresholdRatio {
		if !now.Before(g.cooldownUntil) {
			g.cooldownUntil = now.Add(partsPressureCooldown)
			logs.Logger.Warn(
				"clickhouse parts pressure guard entered cooldown",
				zap.String("table", g.tableName),
				zap.Int("buffer_length", bufferLength),
				zap.Int64("active_parts", g.snapshot.ActiveParts),
				zap.Int64("max_parts_in_total", g.snapshot.MaxPartsInTotal),
				zap.Int64("parts_to_throw_insert", g.snapshot.PartsToThrowInsert),
				zap.Time("cooldown_until", g.cooldownUntil),
				zap.Any("top_partitions", g.snapshot.TopPartitions),
			)
		}
	}

	if now.Before(g.cooldownUntil) {
		reason := "parts pressure cooldown"
		if g.lastError != "" {
			reason = g.lastError
		}
		return &deferredFlushError{
			table:  g.tableName,
			until:  g.cooldownUntil,
			reason: reason,
		}
	}

	return nil
}

// ObserveFlushError 在 flush 已经真正触碰 ClickHouse 且返回 Too many parts 后调用。
//
// 这和 BeforeFlush 的区别是：
// 1. BeforeFlush 处理“采样已显示高压”的预防性冷却。
// 2. ObserveFlushError 处理“数据库已经明确拒绝插入”的反应式冷却。
//
// 一旦命中 Too many parts，这里会把冷却期直接拉长到 120 秒，
// 给后台 merge 和人工排查留出恢复窗口。
func (g *PartsPressureGuard) ObserveFlushError(err error, bufferLength int) {
	if !IsTooManyPartsError(err) {
		return
	}

	util.RecordPersistenceError("too_many_parts", err)

	g.mutex.Lock()
	defer g.mutex.Unlock()

	now := time.Now()
	newCooldownUntil := now.Add(tooManyPartsCooldown)
	if newCooldownUntil.After(g.cooldownUntil) {
		g.cooldownUntil = newCooldownUntil
	}
	g.lastError = err.Error()
	g.lastErrorAt = now

	logs.Logger.Warn(
		"clickhouse too many parts guard extended cooldown",
		zap.String("table", g.tableName),
		zap.Int("buffer_length", bufferLength),
		zap.Int64("active_parts", g.snapshot.ActiveParts),
		zap.Int64("max_parts_in_total", g.snapshot.MaxPartsInTotal),
		zap.Int64("parts_to_throw_insert", g.snapshot.PartsToThrowInsert),
		zap.Time("cooldown_until", g.cooldownUntil),
		zap.String("error", g.lastError),
		zap.Any("top_partitions", g.snapshot.TopPartitions),
	)
}
