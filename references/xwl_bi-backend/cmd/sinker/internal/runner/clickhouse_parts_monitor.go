package runner

import (
	"sync"
	"time"

	"github.com/1340691923/xwl_bi/engine/db"
	"github.com/1340691923/xwl_bi/engine/logs"
	"github.com/1340691923/xwl_bi/model"
	"github.com/1340691923/xwl_bi/platform-basic-libs/service/consumer_data"
	"go.uber.org/zap"
)

const (
	// clickHousePartsMonitorSampleInterval 控制 system.parts 采样频率。
	//
	// 这里故意不做太高频：
	// 1. system.parts 是系统表，不适合在秒级热轮询。
	// 2. 当前目标是给旁路表提供“够用的压力快照”，而不是精细秒级监控。
	clickHousePartsMonitorSampleInterval = time.Minute
	// clickHousePartsMonitorIdleTimeout 控制“多久没有旁路写入就自动停掉采样循环”。
	//
	// 这个值需要比稀疏流量的自然抖动更长，否则会出现：
	// 1. 刚停掉
	// 2. 过几十秒又来一条写入
	// 3. 又重新拉起 goroutine
	//
	// 当前先固定为 5 分钟，优先稳住启停频率。
	clickHousePartsMonitorIdleTimeout = 5 * time.Minute
)

// clickHouseInsertPressureSettings 汇总当前进程关心的 ClickHouse 插入压力 setting。
//
// 这里把它们单独收口，而不是散落在多个查询函数里，原因是：
// 1. 采样日志要同时打印这些值。
// 2. guard 判定阈值时需要复用同一份快照。
// 3. 后续如果要继续扩展 `parts_to_delay_insert` 等 setting，入口会更集中。
type clickHouseInsertPressureSettings struct {
	MaxPartitionsPerInsertBlock int
	MaxPartsInTotal             int64
	PartsToThrowInsert          int64
}

// clickHousePartsMonitor 负责按需采样 ClickHouse parts 压力，并把结果分发给对应表的 guard。
//
// 它和最初“启动即常驻”的版本不同：
// 1. 只有实时表或状态表发生写入尝试时才会被唤醒。
// 2. 连续空闲超过 idleTimeout 后会自行退出。
// 3. 后续又有写入时，再由 NotifyWrite 重新拉起。
//
// 这套设计是为了兼顾两类流量：
// 1. 高峰连续流量：monitor 常驻，持续提供压力快照。
// 2. 稀疏/阵发流量：monitor 不会长期空转，也不会因短暂空闲频繁误停。
type clickHousePartsMonitor struct {
	settings       clickHouseInsertPressureSettings
	guards         map[string]*consumer_data.PartsPressureGuard
	stop           <-chan struct{}
	sampleInterval time.Duration
	idleTimeout    time.Duration
	sampleFn       func()

	mutex       sync.Mutex
	running     bool
	lastWriteAt time.Time
}

type clickHouseTopPartitionRow struct {
	PartitionID string `db:"partition_id"`
	Parts       int64  `db:"parts"`
}

// newClickHousePartsMonitor 创建生产环境默认配置的 parts 采样器。
func newClickHousePartsMonitor(settings clickHouseInsertPressureSettings, guards map[string]*consumer_data.PartsPressureGuard, stop <-chan struct{}) *clickHousePartsMonitor {
	return newClickHousePartsMonitorWithOptions(settings, guards, stop, clickHousePartsMonitorSampleInterval, clickHousePartsMonitorIdleTimeout, nil)
}

// newClickHousePartsMonitorWithOptions 允许测试注入更短周期和自定义 sampleFn。
//
// 示例：
// 1. 生产环境：sampleFn=nil，内部真实查询 system.parts。
// 2. 单元测试：sampleFn=mock，避免测试依赖真实 ClickHouse。
func newClickHousePartsMonitorWithOptions(
	settings clickHouseInsertPressureSettings,
	guards map[string]*consumer_data.PartsPressureGuard,
	stop <-chan struct{},
	sampleInterval time.Duration,
	idleTimeout time.Duration,
	sampleFn func(),
) *clickHousePartsMonitor {
	return &clickHousePartsMonitor{
		settings:       settings,
		guards:         guards,
		stop:           stop,
		sampleInterval: sampleInterval,
		idleTimeout:    idleTimeout,
		sampleFn:       sampleFn,
	}
}

// SampleOnce 执行一次完整的 parts 采样，并把结果推给每张目标表的 guard。
//
// 处理步骤：
// 1. 遍历当前接管的表名。
// 2. 查询该表 active parts 和 top partitions。
// 3. 更新对应 guard 的最近一次快照。
// 4. 记录调试日志，便于排查“为什么 guard 进入了冷却期”。
func (m *clickHousePartsMonitor) SampleOnce() {
	if m == nil {
		return
	}

	for tableName, guard := range m.guards {
		snapshot, err := queryTablePartsSnapshot(tableName, m.settings)
		if err != nil {
			logs.Logger.Warn("query clickhouse parts snapshot failed", zap.String("table", tableName), zap.Error(err))
			continue
		}

		guard.UpdateSnapshot(snapshot)
		logs.Logger.Debug(
			"clickhouse parts snapshot",
			zap.String("table", snapshot.Table),
			zap.Int64("active_parts", snapshot.ActiveParts),
			zap.Int64("max_parts_in_total", snapshot.MaxPartsInTotal),
			zap.Int64("parts_to_throw_insert", snapshot.PartsToThrowInsert),
			zap.Any("top_partitions", snapshot.TopPartitions),
		)
	}
}

// NotifyWrite 在旁路表发生一次“写入尝试”时调用，用来驱动采样器按需启动。
//
// 这里的语义不是“真正写成功”，而是“当前这张表又开始有写入活动了”。
// 这样做的原因是：
// 1. 我们关心的是“表最近是否仍处于活跃写入阶段”。
// 2. 即使写入因为冷却保护被延后，这也说明后续仍可能继续写入，采样器不应过早退出。
// 3. 把触发点放在写入入口，而不是放在成功提交之后，可以更早恢复 parts 观测。
//
// 并发约束：
// 1. 每次调用都会刷新 lastWriteAt。
// 2. 如果采样器已经在跑，只更新时间，不重复起 goroutine。
// 3. 只有从“未运行”切到“运行中”时，才会真正创建新的后台循环。
//
// 示例：
// 1. 12:00:00 首次写入 -> 启动 monitor
// 2. 12:01:10 再来一条稀疏写入 -> 只刷新 lastWriteAt
// 3. 12:06:20 仍无新写入 -> runLoop 在下一次 tick 判断空闲后停止
func (m *clickHousePartsMonitor) NotifyWrite(tableName string) {
	if m == nil {
		return
	}

	now := time.Now()
	m.mutex.Lock()
	m.lastWriteAt = now
	if m.running {
		m.mutex.Unlock()
		return
	}
	m.running = true
	m.mutex.Unlock()

	logs.Logger.Debug(
		"clickhouse parts monitor started",
		zap.String("table", tableName),
		zap.Duration("sample_interval", m.sampleInterval),
		zap.Duration("idle_timeout", m.idleTimeout),
	)

	go m.runLoop()
}

// runLoop 是单个 monitor 实例的后台采样循环。
//
// 执行顺序：
// 1. 启动后先立刻采样一次，避免第一轮写入后长时间没有快照。
// 2. 之后按固定 sampleInterval 周期运行。
// 3. 每次 tick 先判断是否已经连续空闲足够久。
// 4. 若空闲超时则退出；否则继续采样。
// 5. 收到 stop 信号时，无条件退出。
func (m *clickHousePartsMonitor) runLoop() {
	if m == nil {
		return
	}

	ticker := time.NewTicker(m.sampleInterval)
	defer ticker.Stop()

	m.sample()

	for {
		select {
		case <-ticker.C:
			if m.shouldStopForIdle() {
				return
			}
			m.sample()
		case <-m.stop:
			m.markStopped()
			return
		}
	}
}

// sample 统一封装“执行一次采样”的入口。
//
// 这里单独抽一层的目的，是让生产代码和测试代码走同一套状态机：
// 1. 生产环境默认转到 SampleOnce，真实访问 ClickHouse。
// 2. 测试环境可注入 sampleFn，只验证启停与节流逻辑。
func (m *clickHousePartsMonitor) sample() {
	if m == nil {
		return
	}

	if m.sampleFn != nil {
		m.sampleFn()
		return
	}

	m.SampleOnce()
}

// shouldStopForIdle 判断当前 monitor 是否已经空闲足够久，可以安全退出。
//
// 这里不是“一次没写入就停”，而是“连续 idleTimeout 都没有新的写入活动才停”。
// 这样可以吸收稀疏流量抖动，避免出现频繁启停。
//
// 返回值含义：
// 1. true：当前循环应退出。
// 2. false：仍然保持运行，继续下一轮采样。
func (m *clickHousePartsMonitor) shouldStopForIdle() bool {
	if m == nil {
		return true
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	if time.Since(m.lastWriteAt) < m.idleTimeout {
		return false
	}

	m.running = false
	logs.Logger.Debug(
		"clickhouse parts monitor stopped for idle",
		zap.Duration("idle_timeout", m.idleTimeout),
		zap.Time("last_write_at", m.lastWriteAt),
	)
	return true
}

// markStopped 把 monitor 状态显式切回“未运行”。
//
// 它只负责状态收尾，不做额外日志和业务判断，
// 目的是让 stop 分支和 idle 自停分支都能复用同一处收尾逻辑。
func (m *clickHousePartsMonitor) markStopped() {
	if m == nil {
		return
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.running = false
}

// queryClickHouseInsertPressureSettings 启动时查询一次和插入压力相关的 ClickHouse setting。
//
// 这些值都属于“环境级配置”，不适合在 flush 热路径反复查询。
func queryClickHouseInsertPressureSettings() clickHouseInsertPressureSettings {
	settings := clickHouseInsertPressureSettings{
		MaxPartitionsPerInsertBlock: queryIntSetting(
			"system.settings",
			"max_partitions_per_insert_block",
			"query max_partitions_per_insert_block",
		),
		MaxPartsInTotal: queryInt64Setting(
			"system.merge_tree_settings",
			"max_parts_in_total",
			"query max_parts_in_total",
		),
		PartsToThrowInsert: queryInt64Setting(
			"system.merge_tree_settings",
			"parts_to_throw_insert",
			"query parts_to_throw_insert",
		),
	}

	logs.Logger.Info(
		"clickhouse insert pressure settings loaded",
		zap.Int("max_partitions_per_insert_block", settings.MaxPartitionsPerInsertBlock),
		zap.Int64("max_parts_in_total", settings.MaxPartsInTotal),
		zap.Int64("parts_to_throw_insert", settings.PartsToThrowInsert),
	)
	return settings
}

func queryIntSetting(tableName, settingName, logPrefix string) int {
	return int(queryInt64Setting(tableName, settingName, logPrefix))
}

// queryInt64Setting 从指定系统表读取单个 Int64 类型 setting。
//
// 查询失败或返回非法值时统一回退为 0，由上层决定如何使用“未知值”。
func queryInt64Setting(tableName, settingName, logPrefix string) int64 {
	var value int64
	query := "SELECT toInt64(value) FROM " + tableName + " WHERE name = ?"
	if err := db.ClickHouseSqlx.Get(&value, query, settingName); err != nil {
		logs.Logger.Warn(logPrefix+" failed", zap.String("setting", settingName), zap.Error(err))
		return 0
	}

	if value <= 0 {
		logs.Logger.Warn(logPrefix+" returned invalid value", zap.String("setting", settingName), zap.Int64("value", value))
		return 0
	}

	return value
}

// queryTablePartsSnapshot 查询某张目标表当前的 parts 压力快照。
//
// 输出包含两类信息：
// 1. ActiveParts：整体活跃 part 数量，用于 guard 阈值判断。
// 2. TopPartitions：part 最多的前几个分区，用于日志诊断。
//
// 示例：
// 1. 如果 active_parts=80000，max_parts_in_total=100000，guard 可能进入冷却。
// 2. 如果 top_partitions 显示某几个月份异常集中，就能直接判断是哪个时间窗口制造了 parts 压力。
func queryTablePartsSnapshot(tableName string, settings clickHouseInsertPressureSettings) (consumer_data.PartsPressureSnapshot, error) {
	var activeParts int64
	if err := db.ClickHouseSqlx.Get(
		&activeParts,
		"SELECT count() FROM system.parts WHERE active = 1 AND database = ? AND table = ?",
		model.GlobConfig.Comm.ClickHouse.DbName,
		tableName,
	); err != nil {
		return consumer_data.PartsPressureSnapshot{}, err
	}

	topPartitions := make([]clickHouseTopPartitionRow, 0)
	if err := db.ClickHouseSqlx.Select(
		&topPartitions,
		`SELECT partition_id, count() AS parts
FROM system.parts
WHERE active = 1 AND database = ? AND table = ?
GROUP BY partition_id
ORDER BY parts DESC
LIMIT 5`,
		model.GlobConfig.Comm.ClickHouse.DbName,
		tableName,
	); err != nil {
		return consumer_data.PartsPressureSnapshot{}, err
	}

	partitions := make([]consumer_data.PartsPressureTopPartition, 0, len(topPartitions))
	for _, row := range topPartitions {
		partitions = append(partitions, consumer_data.PartsPressureTopPartition{
			PartitionID: row.PartitionID,
			Parts:       row.Parts,
		})
	}

	return consumer_data.PartsPressureSnapshot{
		Table:              tableName,
		ActiveParts:        activeParts,
		MaxPartsInTotal:    settings.MaxPartsInTotal,
		PartsToThrowInsert: settings.PartsToThrowInsert,
		TopPartitions:      partitions,
		SampledAt:          time.Now(),
	}, nil
}
