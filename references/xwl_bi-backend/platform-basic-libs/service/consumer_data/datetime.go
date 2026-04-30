package consumer_data

import (
	"fmt"
	"strings"
	"time"

	"github.com/1340691923/xwl_bi/platform-basic-libs/util"
)

// normalizeDateTimeForClickHouse 负责把批量写库里常见的时间字符串统一转换成 ClickHouse 能直接接受的 time.Time。
//
// 设计目标：
// 1. 兼容完整时间，例如 "2026-04-08 16:14:53"。
// 2. 兼容只有日期的历史值，例如 "2026-04-08"。
// 3. 避免因为单条脏时间把整个批次永久卡死。
//
// 示例：
// 1. "2026-04-08 16:14:53" -> time.Time("2026-04-08 16:14:53")
// 2. "2026-04-08" -> time.Time("2026-04-08 00:00:00")
func normalizeDateTimeForClickHouse(raw string) (time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, fmt.Errorf("empty datetime")
	}

	for _, layout := range []string{util.TimeFormat, util.TimeFormatDay2} {
		if parsed, err := time.ParseInLocation(layout, raw, time.Local); err == nil {
			return parsed, nil
		}
	}

	return time.Time{}, fmt.Errorf("unsupported datetime format %q", raw)
}

// NormalizeDateTimeString 是给其他模块复用的字符串版本包装。
//
// 它和 normalizeDateTimeForClickHouse 使用同一套解析规则，
// 只是把结果重新格式化回 "2006-01-02 15:04:05"。
func NormalizeDateTimeString(raw string) (string, error) {
	parsed, err := normalizeDateTimeForClickHouse(raw)
	if err != nil {
		return "", err
	}
	return parsed.Format(util.TimeFormat), nil
}

// ParseDateTimeString 暴露统一的 ClickHouse 兼容时间解析规则。
func ParseDateTimeString(raw string) (time.Time, error) {
	return normalizeDateTimeForClickHouse(raw)
}

// NormalizeReportTime 用于把上报链路里的 report_time 整理成统一标准格式。
//
// 为什么放在 consumer_data：
// 1. 当前真正共用它的是 `report_server` 和 `sinker`。
// 2. 这两个模块都已经依赖 consumer_data 的时间规则，不需要再为了时间 helper 人为引入一个新目录。
// 3. 它内部复用的就是 ClickHouse 兼容时间解析规则，放在这里语义最直接。
//
// 示例：
// 1. "2026-04-08" -> "2026-04-08 00:00:00"
// 2. "2026-04-08 16:14:53" -> "2026-04-08 16:14:53"
// 3. "2026/04/08" -> "2026/04/08"；这里保留原值，由上层决定是否报错
func NormalizeReportTime(reportTime string) string {
	reportTime = strings.TrimSpace(reportTime)
	if reportTime == "" {
		return reportTime
	}

	normalized, err := NormalizeDateTimeString(reportTime)
	if err == nil {
		return normalized
	}

	return reportTime
}

// NormalizeClientTime 负责把客户端上报时间整理成系统内部统一格式，同时返回解析后的 time.Time。
//
// 示例：
// 1. "2026-04-08 16:14:53" -> "2026-04-08 16:14:53", parsedTime
// 2. "2026-04-08" -> "2026-04-08 00:00:00", parsedTime
func NormalizeClientTime(clientTime string) (string, time.Time, error) {
	clientTime = strings.TrimSpace(clientTime)
	for _, layout := range []string{util.TimeFormat, util.TimeFormatDay2} {
		if parsed, err := time.ParseInLocation(layout, clientTime, time.Local); err == nil {
			return parsed.Format(util.TimeFormat), parsed, nil
		}
	}

	return "", time.Time{}, fmt.Errorf("unsupported client time format %q", clientTime)
}

// ReportTimeHasClock 返回原始 report_time 是否自带时分秒。
//
// 这个判断必须基于“原始输入值”做，而不能基于归一化后的字符串反推。
// 否则像 "2026-04-08" 这种只有日期的值，在被归一化成
// "2026-04-08 00:00:00" 之后，就会错误地看起来像“原始就带了时分秒”。
//
// 示例：
// 1. "2026-04-08 16:14:53" -> true
// 2. "2026-04-08" -> false
// 3. "bad-time" -> false
func ReportTimeHasClock(reportTime string) bool {
	reportTime = strings.TrimSpace(reportTime)
	if _, err := time.ParseInLocation(util.TimeFormat, reportTime, time.Local); err == nil {
		return true
	}
	return false
}

// NormalizeReportTimeForValidation 用于“分钟级时间偏差校验”场景。
//
// 返回值说明：
// 1. 第一个返回值是校验时应该使用的 report_time。
// 2. 第二个返回值表示原始 report_time 是否自带时分秒。
//
// 示例：
// 1. report_time="2026-04-08" -> "2026-04-08 00:00:00", false
// 2. report_time="bad", fallback="2026-04-08 10:11:12" -> "2026-04-08 10:11:12", false
func NormalizeReportTimeForValidation(reportTime, fallback string) (string, bool) {
	reportTime = strings.TrimSpace(reportTime)
	for _, candidate := range []struct {
		layout   string
		hasClock bool
	}{
		{layout: util.TimeFormat, hasClock: true},
		{layout: util.TimeFormatDay2, hasClock: false},
	} {
		if parsed, err := time.ParseInLocation(candidate.layout, reportTime, time.Local); err == nil {
			return parsed.Format(util.TimeFormat), candidate.hasClock
		}
	}

	if fallback != "" {
		return fallback, false
	}

	return reportTime, false
}
