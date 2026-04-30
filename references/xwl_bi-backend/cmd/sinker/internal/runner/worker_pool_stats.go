package runner

import (
	"strings"
	"time"

	"github.com/1340691923/xwl_bi/engine/logs"
	"github.com/1340691923/xwl_bi/platform-basic-libs/util"
	"go.uber.org/zap"
)

const (
	sinkerWorkerPoolNormalLogInterval = 5 * time.Minute
	sinkerWorkerPoolChangeInterval    = 30 * time.Second
)

// startSinkerWorkerPoolStatsLoops 启动 worker pool 的两条统计日志循环。
//
// 设计目标：
// 1. 常规日志固定每 5 分钟一条，形成稳定趋势线。
// 2. 变化日志固定每 30 秒采样，只在快照变化时额外打印，不影响常规日志频率。
// 3. 两条循环都挂在同一个 stop channel 上，进程退出时能一起停止。
func startSinkerWorkerPoolStatsLoops(pool *util.DynamicWorkerPool, stop <-chan struct{}) {
	if pool == nil || stop == nil || !util.IsSinkerDiagnosticLogEnabled() {
		return
	}

	go runSinkerWorkerPoolNormalStatsLoop(pool, stop)
	go runSinkerWorkerPoolChangeStatsLoop(pool, stop)
}

func runSinkerWorkerPoolNormalStatsLoop(pool *util.DynamicWorkerPool, stop <-chan struct{}) {
	ticker := time.NewTicker(sinkerWorkerPoolNormalLogInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			logSinkerWorkerPoolStats("normal", pool.Name(), pool.Stats(), nil)
		case <-stop:
			return
		}
	}
}

func runSinkerWorkerPoolChangeStatsLoop(pool *util.DynamicWorkerPool, stop <-chan struct{}) {
	ticker := time.NewTicker(sinkerWorkerPoolChangeInterval)
	defer ticker.Stop()

	prev := pool.Stats()
	for {
		select {
		case <-ticker.C:
			current := pool.Stats()
			changeFields := diffWorkerPoolStats(prev, current)
			if len(changeFields) > 0 {
				logSinkerWorkerPoolStats("change", pool.Name(), current, changeFields)
			}

			if shouldLogWorkerPoolPressure(current) {
				logs.Logger.Warn(
					"worker pool under pressure",
					workerPoolStatsFields("pressure", pool.Name(), current, changeFields)...,
				)
			}

			if current.RejectedTotal > prev.RejectedTotal {
				logs.Logger.Warn(
					"worker pool submit rejected",
					append(
						workerPoolStatsFields("change", pool.Name(), current, changeFields),
						zap.Int64("rejected_delta", current.RejectedTotal-prev.RejectedTotal),
					)...,
				)
			}

			prev = current
		case <-stop:
			return
		}
	}
}

// diffWorkerPoolStats 负责识别这次快照相对上次快照发生了哪些变化。
//
// 变化判定同时覆盖：
// 1. 池负载变化
// 2. 总 goroutine 数变化
//
// 示例：
// 1. Queued 从 0 变成 128 -> 返回包含 tasks_queued
// 2. GoroutinesTotal 从 55 变成 63 -> 返回包含 goroutines_total
func diffWorkerPoolStats(prev, current util.WorkerPoolStats) []string {
	fields := make([]string, 0, 8)

	if prev.GoroutinesTotal != current.GoroutinesTotal {
		fields = append(fields, "goroutines_total")
	}
	if prev.Running != current.Running {
		fields = append(fields, "workers_running")
	}
	if prev.Idle != current.Idle {
		fields = append(fields, "workers_idle")
	}
	if prev.Capacity != current.Capacity {
		fields = append(fields, "workers_capacity")
	}
	if prev.Queued != current.Queued {
		fields = append(fields, "tasks_queued")
	}
	if prev.SubmittedTotal != current.SubmittedTotal {
		fields = append(fields, "submitted_total")
	}
	if prev.CompletedTotal != current.CompletedTotal {
		fields = append(fields, "completed_total")
	}
	if prev.RejectedTotal != current.RejectedTotal {
		fields = append(fields, "rejected_total")
	}

	return fields
}

func shouldLogWorkerPoolPressure(stats util.WorkerPoolStats) bool {
	return stats.Queued > 0 && stats.BusyRatio >= 0.8
}

func logSinkerWorkerPoolStats(mode, poolName string, stats util.WorkerPoolStats, changeFields []string) {
	logs.Logger.Info("sinker worker pool stats", workerPoolStatsFields(mode, poolName, stats, changeFields)...)
}

func workerPoolStatsFields(mode, poolName string, stats util.WorkerPoolStats, changeFields []string) []zap.Field {
	return []zap.Field{
		zap.String("pool_name", poolName),
		zap.String("sampling_mode", mode),
		zap.Bool("change_detected", len(changeFields) > 0),
		zap.String("change_fields", strings.Join(changeFields, ",")),
		zap.Int("goroutines_total", stats.GoroutinesTotal),
		zap.Int("workers_capacity", stats.Capacity),
		zap.Int("workers_running", stats.Running),
		zap.Int("workers_idle", stats.Idle),
		zap.Int64("tasks_queued", stats.Queued),
		zap.Int("queue_capacity", stats.QueueCapacity),
		zap.Float64("queue_usage_ratio", stats.QueueUsageRatio),
		zap.Float64("busy_ratio", stats.BusyRatio),
		zap.Int64("submitted_total", stats.SubmittedTotal),
		zap.Int64("completed_total", stats.CompletedTotal),
		zap.Int64("rejected_total", stats.RejectedTotal),
		zap.Bool("closed", stats.Closed),
	}
}
