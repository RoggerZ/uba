package util

import (
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	SinkerDiagnosticLogEnv = "SINKER_DIAGNOSTIC_LOG"
	SinkerTraceOffsetEnv   = "SINKER_TRACE_OFFSET"
)

// SinkerDiagnosticSession 描述当前生效的 sinker 运行态诊断窗口。
// 它把 trace offset、report_handler 慢阶段阈值和消费速率日志提频放在同一个 TTL 会话里管理。
// 示例：
//   - Enabled=true, TraceOffsetEnabled=true, TraceOffset=37029664
//     表示只追踪这一条 offset 的完整链路。
//   - Enabled=true, RateLogInterval=5s, ExpiresAt=10:05:00
//     表示在 10:05:00 之前，每 5 秒输出一次消费速率日志。
type SinkerDiagnosticSession struct {
	// Enabled 表示当前是否处于诊断窗口内。
	Enabled bool
	// TraceOffsetEnabled 表示是否开启了单条 offset 跟踪。
	TraceOffsetEnabled bool
	// TraceOffset 是需要额外打印链路细节的目标 offset。
	TraceOffset int64
	// ReportHandlerStageTimingThreshold 控制 report_handler 分阶段耗时日志的阈值。
	ReportHandlerStageTimingThreshold time.Duration
	// RateLogInterval 控制诊断窗口内的消费速率日志间隔；为 0 时回退到默认诊断值。
	RateLogInterval time.Duration
	// ExpiresAt 为诊断窗口到期时间；零值表示常驻或未设置 TTL。
	ExpiresAt time.Time
	// LastUpdatedAt 记录最近一次 enable/disable/expire 变更时间。
	LastUpdatedAt time.Time
	// Source 记录当前会话来源，例如 env、admin_http、expired。
	Source string
}

var sinkerDiagnosticSessionStore = struct {
	once  sync.Once
	mutex sync.RWMutex
	state SinkerDiagnosticSession
}{}

func IsTruthyEnvValue(rawValue string) bool {
	switch strings.TrimSpace(strings.ToLower(rawValue)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func IsSinkerDiagnosticLogEnabled() bool {
	return CurrentSinkerDiagnosticSession().Enabled
}

// CurrentSinkerDiagnosticSession 返回当前生效的 sinker 运行态诊断会话。
// 这个会话统一管理三类诊断能力：
// 1. report_handler 分阶段耗时打点
// 2. trace offset 单条跟踪
// 3. 消费速率日志提频
// 如果 ExpiresAt 已经过期，这里会在读取时自动回落到常驻模式。
// 示例：
//   - Enabled=true, TraceOffsetEnabled=true, TraceOffset=37029664
//     表示只对 offset=37029664 打开单条链路跟踪。
//   - Enabled=true, RateLogInterval=5s
//     表示诊断窗口内把消费速率日志提频到每 5 秒输出一次。
func CurrentSinkerDiagnosticSession() SinkerDiagnosticSession {
	sinkerDiagnosticSessionStore.once.Do(loadSinkerDiagnosticSessionFromEnv)
	sinkerDiagnosticSessionStore.mutex.Lock()
	defer sinkerDiagnosticSessionStore.mutex.Unlock()

	expireSinkerDiagnosticSessionLocked(time.Now())
	return sinkerDiagnosticSessionStore.state
}

// EnableSinkerDiagnosticSession 用新的 TTL 窗口覆盖当前诊断会话。
// 传入 nil 表示关闭对应子能力，例如 traceOffset=nil 代表不启用单条 offset 跟踪。
// duration>0 时会写入 ExpiresAt，过期后由 CurrentSinkerDiagnosticSession 自动回收。
func EnableSinkerDiagnosticSession(duration time.Duration, traceOffset *int64, reportHandlerStageTimingThreshold *time.Duration, rateLogInterval *time.Duration, source string, now time.Time) SinkerDiagnosticSession {
	sinkerDiagnosticSessionStore.once.Do(loadSinkerDiagnosticSessionFromEnv)
	sinkerDiagnosticSessionStore.mutex.Lock()
	defer sinkerDiagnosticSessionStore.mutex.Unlock()

	sinkerDiagnosticSessionStore.state.Enabled = true
	sinkerDiagnosticSessionStore.state.TraceOffsetEnabled = traceOffset != nil
	sinkerDiagnosticSessionStore.state.TraceOffset = 0
	if traceOffset != nil {
		sinkerDiagnosticSessionStore.state.TraceOffset = *traceOffset
	}
	sinkerDiagnosticSessionStore.state.ReportHandlerStageTimingThreshold = 0
	if reportHandlerStageTimingThreshold != nil {
		sinkerDiagnosticSessionStore.state.ReportHandlerStageTimingThreshold = *reportHandlerStageTimingThreshold
	}
	sinkerDiagnosticSessionStore.state.RateLogInterval = 0
	if rateLogInterval != nil {
		sinkerDiagnosticSessionStore.state.RateLogInterval = *rateLogInterval
	}
	sinkerDiagnosticSessionStore.state.LastUpdatedAt = now
	sinkerDiagnosticSessionStore.state.Source = source
	if duration > 0 {
		sinkerDiagnosticSessionStore.state.ExpiresAt = now.Add(duration)
	} else {
		sinkerDiagnosticSessionStore.state.ExpiresAt = time.Time{}
	}

	return sinkerDiagnosticSessionStore.state
}

// DisableSinkerDiagnosticSession 主动关闭当前诊断窗口，并清空 trace/threshold/rateLog 等运行态覆盖值。
func DisableSinkerDiagnosticSession(source string, now time.Time) SinkerDiagnosticSession {
	sinkerDiagnosticSessionStore.once.Do(loadSinkerDiagnosticSessionFromEnv)
	sinkerDiagnosticSessionStore.mutex.Lock()
	defer sinkerDiagnosticSessionStore.mutex.Unlock()

	sinkerDiagnosticSessionStore.state.Enabled = false
	sinkerDiagnosticSessionStore.state.TraceOffsetEnabled = false
	sinkerDiagnosticSessionStore.state.TraceOffset = 0
	sinkerDiagnosticSessionStore.state.ReportHandlerStageTimingThreshold = 0
	sinkerDiagnosticSessionStore.state.RateLogInterval = 0
	sinkerDiagnosticSessionStore.state.ExpiresAt = time.Time{}
	sinkerDiagnosticSessionStore.state.LastUpdatedAt = now
	sinkerDiagnosticSessionStore.state.Source = source
	return sinkerDiagnosticSessionStore.state
}

// ShouldTraceSinkerOffset 判断指定 offset 是否命中了当前诊断窗口的单条跟踪条件。
// 只有会话启用、TraceOffsetEnabled=true 且 offset 精确匹配时才返回 true。
func ShouldTraceSinkerOffset(offset int64) bool {
	session := CurrentSinkerDiagnosticSession()
	return session.Enabled && session.TraceOffsetEnabled && session.TraceOffset == offset
}

func loadSinkerDiagnosticSessionFromEnv() {
	now := time.Now()
	state := SinkerDiagnosticSession{
		Enabled:       IsTruthyEnvValue(os.Getenv(SinkerDiagnosticLogEnv)),
		LastUpdatedAt: now,
		Source:        "env",
	}

	rawTraceOffset := strings.TrimSpace(os.Getenv(SinkerTraceOffsetEnv))
	if rawTraceOffset != "" {
		if parsedOffset, err := strconv.ParseInt(rawTraceOffset, 10, 64); err == nil && parsedOffset >= 0 {
			state.TraceOffsetEnabled = true
			state.TraceOffset = parsedOffset
		}
	}

	sinkerDiagnosticSessionStore.state = state
}

func expireSinkerDiagnosticSessionLocked(now time.Time) {
	if !sinkerDiagnosticSessionStore.state.Enabled {
		return
	}
	if sinkerDiagnosticSessionStore.state.ExpiresAt.IsZero() {
		return
	}
	if now.Before(sinkerDiagnosticSessionStore.state.ExpiresAt) {
		return
	}

	sinkerDiagnosticSessionStore.state.Enabled = false
	sinkerDiagnosticSessionStore.state.TraceOffsetEnabled = false
	sinkerDiagnosticSessionStore.state.TraceOffset = 0
	sinkerDiagnosticSessionStore.state.ReportHandlerStageTimingThreshold = 0
	sinkerDiagnosticSessionStore.state.RateLogInterval = 0
	sinkerDiagnosticSessionStore.state.ExpiresAt = time.Time{}
	sinkerDiagnosticSessionStore.state.LastUpdatedAt = now
	sinkerDiagnosticSessionStore.state.Source = "expired"
}

// ResetSinkerDiagnosticSessionForTest 重置全局诊断状态，避免测试之间相互污染。
func ResetSinkerDiagnosticSessionForTest() {
	sinkerDiagnosticSessionStore = struct {
		once  sync.Once
		mutex sync.RWMutex
		state SinkerDiagnosticSession
	}{}
}
