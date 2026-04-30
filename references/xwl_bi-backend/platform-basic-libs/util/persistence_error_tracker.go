package util

import (
	"strings"
	"sync"
	"time"
)

const persistenceErrorWindow = time.Minute

type PersistenceErrorClassification struct {
	ErrorClass              string `json:"error_class"`
	Retriable               bool   `json:"retriable"`
	Severity                string `json:"severity"`
	SuggestedAction         string `json:"suggested_action"`
	CountTowardCircuitBreak bool   `json:"count_toward_circuit_break"`
}

type PersistenceErrorSnapshot struct {
	CountLastMinute int       `json:"countLastMinute"`
	LastClass       string    `json:"lastClass"`
	LastError       string    `json:"lastError"`
	LastOccurredAt  time.Time `json:"lastOccurredAt"`
}

type persistenceErrorEvent struct {
	at    time.Time
	class string
}

type persistenceErrorTracker struct {
	mutex          sync.Mutex
	window         time.Duration
	events         []persistenceErrorEvent
	lastClass      string
	lastError      string
	lastOccurredAt time.Time
}

var globalPersistenceErrorTracker = &persistenceErrorTracker{
	window: persistenceErrorWindow,
}

// ClassifyPersistenceError 把持久化错误统一折叠成保护状态机可消费的错误类别。
//
// 设计目的：
// 1. 让 ReportData2CK / ReportAcceptStatus / RealTimeWarehousing 不必各自实现一套熔断分类。
// 2. 让保护判定看到的是“错误类别”和“是否计入熔断”，而不是零散字符串。
//
// 示例：
//  1. `dial tcp ... connectex` + `clickhouse_prepare_failed`
//     -> `clickhouse_prepare_failed`，计入熔断
//  2. `Too many parts`
//     -> `too_many_parts`，计入熔断
//  3. `converting int to Int32 is unsupported`
//     -> `db_type_conversion`，默认不计入熔断
func ClassifyPersistenceError(stage string, err error) PersistenceErrorClassification {
	classification := PersistenceErrorClassification{
		ErrorClass:              stage,
		Retriable:               true,
		Severity:                "warn",
		SuggestedAction:         "observe",
		CountTowardCircuitBreak: true,
	}
	if err == nil {
		return classification
	}

	message := strings.ToLower(err.Error())
	switch {
	case strings.Contains(message, "too many parts") || strings.Contains(message, "code: 252"):
		classification.ErrorClass = "too_many_parts"
		classification.Severity = "high"
		classification.SuggestedAction = "pause_or_defer_flush"
	case strings.Contains(message, "flush deferred"):
		classification.ErrorClass = "flush_deferred"
		classification.Severity = "medium"
		classification.SuggestedAction = "backoff_flush"
	case strings.Contains(message, "bad connection"):
		classification.ErrorClass = "db_bad_connection"
		classification.Severity = "high"
		classification.SuggestedAction = "reconnect"
	case strings.Contains(message, "deadline exceeded") || strings.Contains(message, "timeout"):
		classification.ErrorClass = "db_timeout"
		classification.Severity = "high"
		classification.SuggestedAction = "pause_and_retry"
	case strings.Contains(message, "broken pipe"):
		classification.ErrorClass = "db_broken_pipe"
		classification.Severity = "high"
		classification.SuggestedAction = "reconnect"
	case strings.Contains(message, "connection reset"):
		classification.ErrorClass = "db_connection_reset"
		classification.Severity = "high"
		classification.SuggestedAction = "reconnect"
	case strings.Contains(message, "unknown setting"):
		classification.ErrorClass = "db_unknown_setting"
		classification.Retriable = false
		classification.Severity = "high"
		classification.SuggestedAction = "fix_sql_or_driver_config"
	case strings.Contains(message, "converting") || strings.Contains(message, "unsupported"):
		classification.ErrorClass = "db_type_conversion"
		classification.Retriable = false
		classification.Severity = "medium"
		classification.SuggestedAction = "fix_row_mapping"
		classification.CountTowardCircuitBreak = false
	case strings.Contains(message, "transaction"):
		classification.ErrorClass = "db_transaction_state"
		classification.Retriable = false
		classification.Severity = "high"
		classification.SuggestedAction = "fix_transaction_state"
	case strings.Contains(message, "connectex") || strings.Contains(message, "connection refused") || strings.Contains(message, "eof"):
		classification.Severity = "high"
		classification.SuggestedAction = "pause_and_retry"
	}

	switch stage {
	case "mysql_unhealthy":
		classification.ErrorClass = "mysql_unhealthy"
		classification.Severity = "high"
		classification.SuggestedAction = "wait_mysql_recover"
	case "clickhouse_prepare_failed":
		classification.ErrorClass = "clickhouse_prepare_failed"
		classification.Severity = "high"
		classification.SuggestedAction = "pause_or_retry_prepare"
	case "clickhouse_append_failed":
		classification.ErrorClass = "clickhouse_append_failed"
		classification.Severity = "high"
		classification.SuggestedAction = "check_row_mapping"
	case "clickhouse_send_failed":
		classification.ErrorClass = "clickhouse_send_failed"
		classification.Severity = "high"
		classification.SuggestedAction = "pause_or_retry_send"
	case "clickhouse_sql_driver_error":
		classification.ErrorClass = "clickhouse_sql_driver_error"
		classification.Severity = "high"
		classification.SuggestedAction = "pause_or_retry_sql_path"
	case "clickhouse_schema_change_failed":
		classification.ErrorClass = "clickhouse_schema_change_failed"
		classification.Severity = "medium"
		classification.SuggestedAction = "fix_schema_change"
		classification.CountTowardCircuitBreak = false
	}

	return classification
}

// RecordPersistenceError 记录一次持久化错误，并把最近 1 分钟内“计入熔断”的错误数暴露给保护状态机。
//
// 示例：
//  1. 连续 10 次 `clickhouse_prepare_failed`
//     -> `CountLastMinute` 会快速增长，保护状态机可直接进入 `hard_paused`
//  2. 连续 10 次 `db_type_conversion`
//     -> 仍会更新 `LastClass/LastError`，但默认不计入熔断计数
func RecordPersistenceError(stage string, err error) PersistenceErrorClassification {
	classification := ClassifyPersistenceError(stage, err)
	globalPersistenceErrorTracker.record(classification, err)
	return classification
}

func GetPersistenceErrorSnapshot() PersistenceErrorSnapshot {
	return globalPersistenceErrorTracker.snapshot()
}

func ResetPersistenceErrorTrackerForTest() {
	globalPersistenceErrorTracker.reset()
}

func (t *persistenceErrorTracker) record(classification PersistenceErrorClassification, err error) {
	if t == nil {
		return
	}

	t.mutex.Lock()
	defer t.mutex.Unlock()

	now := time.Now()
	t.pruneLocked(now)
	if classification.CountTowardCircuitBreak {
		t.events = append(t.events, persistenceErrorEvent{
			at:    now,
			class: classification.ErrorClass,
		})
	}
	t.lastClass = classification.ErrorClass
	if err != nil {
		t.lastError = err.Error()
	}
	t.lastOccurredAt = now
}

func (t *persistenceErrorTracker) snapshot() PersistenceErrorSnapshot {
	if t == nil {
		return PersistenceErrorSnapshot{}
	}

	t.mutex.Lock()
	defer t.mutex.Unlock()

	now := time.Now()
	t.pruneLocked(now)
	return PersistenceErrorSnapshot{
		CountLastMinute: len(t.events),
		LastClass:       t.lastClass,
		LastError:       t.lastError,
		LastOccurredAt:  t.lastOccurredAt,
	}
}

func (t *persistenceErrorTracker) reset() {
	if t == nil {
		return
	}

	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.events = nil
	t.lastClass = ""
	t.lastError = ""
	t.lastOccurredAt = time.Time{}
}

func (t *persistenceErrorTracker) pruneLocked(now time.Time) {
	if t == nil || len(t.events) == 0 {
		return
	}

	cutoff := now.Add(-t.window)
	pruned := t.events[:0]
	for _, event := range t.events {
		if event.at.Before(cutoff) {
			continue
		}
		pruned = append(pruned, event)
	}
	t.events = pruned
}
