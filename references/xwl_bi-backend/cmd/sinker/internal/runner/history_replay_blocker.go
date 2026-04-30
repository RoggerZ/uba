package runner

import (
	"sync"
	"time"

	"github.com/1340691923/xwl_bi/engine/logs"
	"github.com/1340691923/xwl_bi/platform-basic-libs/service/consumer_data"
	"go.uber.org/zap"
)

type historyReplayDropStat struct {
	count      int64
	oldestTime time.Time
	newestTime time.Time
}

// historyReplayBlocker 负责把“超出旁路表保留窗口”的历史回放数据拦截掉。
//
// 这里的窗口是进程启动时固定下来的：
// 1. 启动时计算 cutoff。
// 2. 后续整次运行期间都用这条 cutoff 判断。
// 3. 这样能保证行为稳定，避免运行中途因为时钟推进导致边界漂移。
type historyReplayBlocker struct {
	cutoff     time.Time
	statsMutex sync.Mutex
	stats      map[string]*historyReplayDropStat
}

func newHistoryReplayBlocker(cutoff time.Time) *historyReplayBlocker {
	logs.Logger.Info(
		"history replay blocker initialized",
		zap.Time("cutoff", cutoff),
		zap.Int("retention_months", consumer_data.SidecarRetentionMonths()),
	)
	return &historyReplayBlocker{
		cutoff: cutoff,
		stats:  make(map[string]*historyReplayDropStat),
	}
}

// ShouldSkip 判断一条消息是否已经落在旁路表保留窗口之外。
//
// 处理步骤：
// 1. 先按统一时间规则解析业务时间。
// 2. 解析失败时不拦截，交给原有链路自行处理。
// 3. 若业务时间早于 cutoff，则记入聚合统计并返回 true。
//
// 为什么这里只做聚合统计而不是逐条打日志：
// 1. 历史回放数据通常是成批出现的。
// 2. 逐条日志会快速淹没真正的异常日志。
// 3. 聚合日志已经足够回答“丢了多少、最老/最新时间是多少”。
func (b *historyReplayBlocker) ShouldSkip(tableName string, businessTime string) bool {
	if b == nil {
		return false
	}

	parsedTime, err := consumer_data.ParseDateTimeString(businessTime)
	if err != nil {
		return false
	}
	if !parsedTime.Before(b.cutoff) {
		return false
	}

	b.statsMutex.Lock()
	defer b.statsMutex.Unlock()

	stat, ok := b.stats[tableName]
	if !ok {
		stat = &historyReplayDropStat{
			oldestTime: parsedTime,
			newestTime: parsedTime,
		}
		b.stats[tableName] = stat
	}

	stat.count++
	if parsedTime.Before(stat.oldestTime) {
		stat.oldestTime = parsedTime
	}
	if parsedTime.After(stat.newestTime) {
		stat.newestTime = parsedTime
	}

	return true
}

// FlushLogs 把当前周期累计的历史回放丢弃统计一次性写出。
//
// 这里的日志是“边界证据”：
// 1. table 告诉我们是哪张旁路表被拦截。
// 2. skipped_count 告诉我们规模。
// 3. oldest / newest_business_time 告诉我们被拦截数据覆盖的时间区间。
func (b *historyReplayBlocker) FlushLogs() {
	if b == nil {
		return
	}

	b.statsMutex.Lock()
	defer b.statsMutex.Unlock()

	for tableName, stat := range b.stats {
		if stat.count == 0 {
			continue
		}

		logs.Logger.Info(
			"history replay data skipped for sidecar table",
			zap.String("table", tableName),
			zap.Int64("skipped_count", stat.count),
			zap.Time("cutoff", b.cutoff),
			zap.Time("oldest_business_time", stat.oldestTime),
			zap.Time("newest_business_time", stat.newestTime),
		)

		stat.count = 0
		stat.oldestTime = time.Time{}
		stat.newestTime = time.Time{}
	}
}

// startHistoryReplayBlockerLogLoop 定时冲刷聚合日志，避免统计一直留在内存里。
func startHistoryReplayBlockerLogLoop(blocker *historyReplayBlocker, stop <-chan struct{}) {
	if blocker == nil {
		return
	}

	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				blocker.FlushLogs()
			case <-stop:
				blocker.FlushLogs()
				return
			}
		}
	}()
}
