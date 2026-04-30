package consumer_data

import "time"

const (
	sidecarRetentionMonths = 3

	TableNameAcceptanceStatus    = "xwl_acceptance_status"
	TableNameRealTimeWarehousing = "xwl_real_time_warehousing"
)

// SidecarRetentionMonths 返回实时旁路表统一使用的保留月数。
//
// 当前这里统一约束两张表：
// 1. xwl_real_time_warehousing
// 2. xwl_acceptance_status
//
// 它们都属于“排障 / 实时旁路”表，保留窗口必须和建表 TTL 保持一致，
// 避免出现“代码按 3 个月拦截，表却按别的周期清理”的漂移。
func SidecarRetentionMonths() int {
	return sidecarRetentionMonths
}

// SidecarRetentionCutoff 按统一保留月数计算历史回放拦截下界。
func SidecarRetentionCutoff(now time.Time) time.Time {
	return now.AddDate(0, -sidecarRetentionMonths, 0)
}
