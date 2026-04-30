// Package model 应用启动引擎层
package model

import (
	"errors"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"os"
	"strings"
)

var GlobConfig Config

var CmdName string

// Config 全局配置结构体
type Config struct {
	Manager ManagerConfig `json:"manager"`
	Report  ReportConfig  `json:"report"`
	Sinker  SinkerConfig  `json:"sinker"`
	Comm    struct {
		Log        LogConfig        `json:"log"`
		Mysql      MysqlConfig      `json:"mysql"`
		ClickHouse ClickHouseConfig `json:"clickhouse"`
		Kafka      KafkaCfg         `json:"kafka"`
		Redis      RedisConfig      `json:"redis"`
	} `json:"comm"`
}

func (c Config) Validate() error {
	return c.Sinker.Validate()
}

type ManagerConfig struct {
	Port              uint16 `json:"port"`              //铸龙分析系统http启动端口
	CkQueryLimit      int    `json:"ckQueryLimit"`      //clickhouse 查询限流器阈值
	CkQueryExpiration int    `json:"ckQueryExpiration"` //clickhouse 查询限流器阈值
	JwtSecret         string `json:"jwtSecret"`
	DeBug             bool   `json:"deBug"`
}

type SinkerConfig struct {
	ReportAcceptStatus          BatchConfig                 `json:"reportAcceptStatus"`
	ReportData2CK               BatchConfig                 `json:"reportData2CK"`
	RealTimeWarehousing         BatchConfig                 `json:"realTimeWarehousing"`
	ReportConsumerPool          DynamicWorkerPoolConfigJSON `json:"reportConsumerPool"`
	ReportPersistPool           DynamicWorkerPoolConfigJSON `json:"reportPersistPool"`
	Protection                  ProtectionConfig            `json:"protection"`
	ReportConsumerDirectExec    bool                        `json:"reportConsumerDirectExec"`
	AdminHttpHost               string                      `json:"adminHttpHost"`
	AdminHttpPort               uint16                      `json:"adminHttpPort"`
	AdminToken                  string                      `json:"adminToken"`
	DiagnosticDefaultTTLSeconds int                         `json:"diagnosticDefaultTTLSeconds"`
	DiagnosticMaxTTLSeconds     int                         `json:"diagnosticMaxTTLSeconds"`
	PprofHttpPort               uint16                      `json:"pprofHttpPort"`
}

func (c SinkerConfig) Validate() error {
	return c.validateProtectionPreset()
}

func (c SinkerConfig) Normalize() SinkerConfig {
	c.ReportAcceptStatus = c.ReportAcceptStatus.Normalize()
	c.ReportData2CK = c.ReportData2CK.Normalize()
	c.RealTimeWarehousing = c.RealTimeWarehousing.Normalize()
	c.ReportConsumerPool = c.ReportConsumerPool.Normalize()
	c.ReportPersistPool = c.ReportPersistPool.Normalize()
	c.Protection = c.Protection.Normalize()
	c.applyProtectionPreset()
	c.ReportConsumerPool = c.ReportConsumerPool.Normalize()
	c.ReportPersistPool = c.ReportPersistPool.Normalize()
	c.Protection = c.Protection.Normalize()
	if c.AdminHttpHost == "" {
		c.AdminHttpHost = "127.0.0.1"
	}
	if c.AdminHttpPort == 0 {
		c.AdminHttpPort = 8094
	}
	if c.DiagnosticDefaultTTLSeconds <= 0 {
		c.DiagnosticDefaultTTLSeconds = 180
	}
	if c.DiagnosticMaxTTLSeconds <= 0 {
		c.DiagnosticMaxTTLSeconds = 3600
	}
	return c
}

// applyProtectionPreset 把测试/演练场景预设展开成具体的保护配置、worker pool 和 mock 覆盖值。
//
// 设计目的：
// 1. 让 sinker 在保留单一主配置入口时，仍能快速切换常见保护验证场景。
// 2. 把“场景选择”与“具体阈值/池大小细节”分开，降低重复配置文件数量。
// 3. 预设只覆盖测试场景所需字段，生产默认配置不受影响。
//
// 当前约定：
// - 空 preset：不做任何额外覆盖
// - consumer_low：压低 consumer/persist pool 并保持默认阈值
// - protect_mock：在 consumer_low 基础上打开 mock 注入
// - consumer_low_soft_backlog：在 consumer_low 基础上下调 soft backlog 阈值，不改 hard 阈值
//
// 使用示例：
//
//	{
//	  "sinker": {
//	    "protection": {
//	      "mock": {
//	        "preset": "consumer_low_soft_backlog",
//	        "enabled": false
//	      }
//	    }
//	  }
//	}
//
// 说明：
// 1. preset 只负责“展开场景”，不是新的运行时状态机。
// 2. preset 发生在 Normalize 阶段，所以后续 runtime 看到的仍然是扁平化后的普通配置。
// 3. 如果调用方同时手写了与 preset 冲突的显式字段，启动前 Validate 会直接报错，而不是静默覆盖。
func (c *SinkerConfig) applyProtectionPreset() {
	preset := strings.TrimSpace(strings.ToLower(c.Protection.Mock.Preset))
	switch preset {
	case "":
		return
	case protectionMockPresetConsumerLow:
		applyConsumerLowPoolPreset(c)
	case protectionMockPresetProtectMock:
		applyConsumerLowPoolPreset(c)
		applyProtectMockPreset(&c.Protection)
	case protectionMockPresetConsumerLowSoftBacklog:
		applyConsumerLowPoolPreset(c)
		applyConsumerLowSoftBacklogPreset(&c.Protection)
	}
}

func (c SinkerConfig) validateProtectionPreset() error {
	preset := strings.TrimSpace(strings.ToLower(c.Protection.Mock.Preset))
	switch preset {
	case "":
		return nil
	case protectionMockPresetConsumerLow:
		if err := validateConsumerLowPoolPreset(c.ReportConsumerPool, "reportConsumerPool"); err != nil {
			return err
		}
		if err := validateConsumerLowPoolPreset(c.ReportPersistPool, "reportPersistPool"); err != nil {
			return err
		}
		return nil
	case protectionMockPresetProtectMock:
		if err := validateConsumerLowPoolPreset(c.ReportConsumerPool, "reportConsumerPool"); err != nil {
			return err
		}
		if err := validateConsumerLowPoolPreset(c.ReportPersistPool, "reportPersistPool"); err != nil {
			return err
		}
		return validateProtectMockPreset(c.Protection.Mock)
	case protectionMockPresetConsumerLowSoftBacklog:
		if err := validateConsumerLowPoolPreset(c.ReportConsumerPool, "reportConsumerPool"); err != nil {
			return err
		}
		if err := validateConsumerLowPoolPreset(c.ReportPersistPool, "reportPersistPool"); err != nil {
			return err
		}
		return validateConsumerLowSoftBacklogPreset(c.Protection)
	default:
		return fmt.Errorf("unknown protection.mock.preset: %s", c.Protection.Mock.Preset)
	}
}

type RedisConfig struct {
	Addr      string `json:"addr"`
	Passwd    string `json:"passwd"`
	Db        int    `json:"db"`
	MaxIdle   int    `json:"maxIdle"`
	MaxActive int    `json:"maxActive"`
}

type ClickHouseConfig struct {
	Username             string `json:"username"`
	Pwd                  string `json:"pwd"`
	IP                   string `json:"ip"`
	Port                 string `json:"port"`
	DbName               string `json:"dbName"`
	MaxOpenConns         int    `json:"maxOpenConns"`
	MaxIdleConns         int    `json:"maxIdleConns"`
	MacrosShardKeyName   string `json:"macrosShardKeyName"`
	MacrosReplicaKeyName string `json:"macrosReplicaKeyName"`
	ClusterName          string `json:"clusterName"`
	MaxQuerySize         int    `json:"maxQuerySize"`
}

func (this ClickHouseConfig) GetMaxQuerySize() int {
	if this.MaxQuerySize <= 0 {
		return 1048576
	}
	return this.MaxQuerySize
}

type MysqlConfig struct {
	Username     string         `json:"username"`
	Pwd          string         `json:"pwd"`
	IP           string         `json:"ip"`
	Port         string         `json:"port"`
	DbName       string         `json:"dbName"`
	MaxOpenConns int            `json:"maxOpenConns"`
	MaxIdleConns int            `json:"maxIdleConns"`
	HealthCheck  DBHealthConfig `json:"healthCheck"`
}

func (c MysqlConfig) Normalize() MysqlConfig {
	c.HealthCheck = c.HealthCheck.Normalize()
	return c
}

type ReportConfig struct {
	ReportPort          uint16   `json:"reportPort"`          // 上报程序启动端口
	ReadTimeout         int      `json:"readTimeout"`         // 读取超时，单位秒
	WriteTimeout        int      `json:"writeTimeout"`        // 写入超时，单位秒
	MaxConnsPerIP       int      `json:"maxConnsPerIP"`       // 单 IP 最大连接数
	MaxRequestsPerConn  int      `json:"maxRequestsPerConn"`  // 单连接最大请求数
	IdleTimeout         int      `json:"idleTimeout"`         // 空闲超时，单位秒
	SkipSigned          string   `json:"skipSigned"`          // tools 直传时用于跳过签名的请求头取值
	SignaturePathPrefix string   `json:"signaturePathPrefix"` // 验签时参与签名底串的代理前缀，例如 /api/v3/reporter
	UserAgentBanList    []string `json:"userAgentBanList"`    // 禁止访问的 User-Agent 关键字
}

type LogConfig struct {
	StorageDays int    `json:"storageDays"` //日志保留天数
	LogDir      string `json:"logDir"`      //日志保留文件夹地址
	Level       string `json:"level"`       //日志级别，例如 debug / info / warn / error
}

func (c *Config) GetCkQueryLimit() int {
	if c.Manager.CkQueryLimit == 0 {
		return 30
	}
	return c.Manager.CkQueryLimit
}

func (c *Config) GetCkQueryExpiration() int {
	if c.Manager.CkQueryExpiration == 0 {
		return 2
	}
	return c.Manager.CkQueryExpiration
}

func (c *Config) GetKafkaCfgProducerType() string {
	if c.Comm.Kafka.ProducerType == "" {
		return "sync"
	}
	return c.Comm.Kafka.ProducerType
}

type KafkaCfg struct {
	NumPartitions      int32    `json:"numPartitions"`
	Addresses          []string `json:"addresses"`
	Username           string   `json:"username"`
	Password           string   `json:"password"`
	ReportTopicName    string   `json:"reportTopicName"`
	ConsumerGroupName  string   `json:"consumerGroupName"`
	RealTimeDataGroup  string   `json:"realTimeDataGroup"`
	ReportData2CKGroup string   `json:"reportData2CKGroup"`
	DebugDataTopicName string   `json:"debugDataTopicName"`
	DebugDataGroup     string   `json:"debugDataGroup"`
	ProducerType       string   `json:"producer_type"`
}

type BatchConfig struct {
	BufferSize    int `json:"bufferSize"`
	FlushInterval int `json:"flushInterval"`
}

type DynamicWorkerPoolConfigJSON struct {
	MinWorkers   int `json:"minWorkers"`
	MaxWorkers   int `json:"maxWorkers"`
	QueueSize    int `json:"queueSize"`
	TuneInterval int `json:"tuneIntervalSeconds"`
	DrainTimeout int `json:"drainTimeoutSeconds"`
}

func (c DynamicWorkerPoolConfigJSON) Normalize() DynamicWorkerPoolConfigJSON {
	if c.MinWorkers < 0 {
		c.MinWorkers = 0
	}
	if c.MaxWorkers < 0 {
		c.MaxWorkers = 0
	}
	if c.QueueSize < 0 {
		c.QueueSize = 0
	}
	if c.TuneInterval < 0 {
		c.TuneInterval = 0
	}
	if c.DrainTimeout < 0 {
		c.DrainTimeout = 0
	}
	return c
}

type ProtectionThresholds struct {
	OrderedCommitPendingCount int64  `json:"orderedCommitPendingCount"`
	GateInFlightMessages      int64  `json:"gateInFlightMessages"`
	GateWaitingTasks          int64  `json:"gateWaitingTasks"`
	WorkerQueueUsagePermille  int    `json:"workerQueueUsagePermille"`
	WorkerBusyPermille        int    `json:"workerBusyPermille"`
	HeapAllocBytes            uint64 `json:"heapAllocBytes"`
	Goroutines                int    `json:"goroutines"`
	SlowReportHandlerSamples  int    `json:"slowReportHandlerSamples"`
	PersistenceErrors         int    `json:"persistenceErrors"`
	DBConsecutiveFailures     int    `json:"dbConsecutiveFailures"`
}

// ProtectionConfig 描述 sinker 运行保护的基础开关、采样节奏和状态切换阈值。
// 它覆盖三类配置：
// 1. 采样与日志频率：决定保护状态机多久刷新一次，以及平时/诊断模式多久打一次速率日志。
// 2. 保护动作：决定 softLimited / hardPaused 的目标速率、驻留时间和恢复窗口。
// 3. 触发阈值：决定 ordered_commit、gate、worker pool、持久化错误、DB 健康等信号何时进入保护态。
// 示例：
//   - SampleIntervalSeconds=5, NormalRateLogIntervalSeconds=180, DefaultDiagnosticRateLogIntervalSeconds=5
//     表示内部每 5 秒采样一次，平时每 3 分钟打一次速率日志，诊断窗口默认每 5 秒打一条。
//   - SoftThresholds.OrderedCommitPendingCount=20000
//     表示 ordered_commit.pending_count 连续命中该阈值时，可作为进入 softLimited 的判定信号之一。
type ProtectionConfig struct {
	// Enabled 控制保护逻辑默认是否启用；为 nil 时 Normalize 会回填为 true。
	Enabled *bool `json:"enabled"`
	// ObserveOnly 为 true 时只记录状态和日志，不真正执行限流、暂停等动作。
	ObserveOnly bool `json:"observeOnly"`
	// SampleIntervalSeconds 是保护状态机、消费速率采样器和 runtime status 的基础采样窗口，单位秒。
	SampleIntervalSeconds int `json:"sampleIntervalSeconds"`
	// NormalRateLogIntervalSeconds 是平时模式的速率日志间隔，单位秒，无 TTL，常驻生效。
	NormalRateLogIntervalSeconds int `json:"normalRateLogIntervalSeconds"`
	// DefaultDiagnosticRateLogIntervalSeconds 是诊断窗口开启后的默认速率日志间隔，单位秒。
	// 如果 diagnostic enable 显式传入 --rate-log-interval，则由会话值覆盖。
	DefaultDiagnosticRateLogIntervalSeconds int `json:"defaultDiagnosticRateLogIntervalSeconds"`
	// RecoveryHealthyWindows 指定恢复时需要连续命中的健康窗口数量。
	RecoveryHealthyWindows int `json:"recoveryHealthyWindows"`
	// HardEscalationWindows 指定 hardPaused 连续命中多少个采样窗口后，再扩大到更多 consumer group。
	HardEscalationWindows int `json:"hardEscalationWindows"`
	// SoftTargetRatePerSecond 是 softLimited 状态下 intake limiter 目标速率。
	SoftTargetRatePerSecond int `json:"softTargetRatePerSecond"`
	// SoftMinHoldSeconds 是 softLimited 最短驻留时间，单位秒。
	SoftMinHoldSeconds int `json:"softMinHoldSeconds"`
	// HardMinHoldSeconds 是 hardPaused 最短驻留时间，单位秒。
	HardMinHoldSeconds int `json:"hardMinHoldSeconds"`
	// MaxAdaptiveHoldSeconds 是自适应驻留时间的上限，避免保护态抖动时无限放大。
	MaxAdaptiveHoldSeconds int `json:"maxAdaptiveHoldSeconds"`
	// SoftThresholds 定义进入 softLimited 时参与判定的默认阈值集合。
	SoftThresholds ProtectionThresholds `json:"softThresholds"`
	// HardThresholds 定义进入 hardPaused 时参与判定的默认阈值集合。
	HardThresholds ProtectionThresholds `json:"hardThresholds"`
	// Mock 是运行时 mock 注入配置，用于在测试环境稳定演练 softLimited / hardPaused。
	Mock ProtectionMockConfig `json:"mock"`
}

const (
	protectionMockPresetConsumerLow            = "consumer_low"
	protectionMockPresetProtectMock            = "protect_mock"
	protectionMockPresetConsumerLowSoftBacklog = "consumer_low_soft_backlog"
)

// ProtectionMockConfig 是 sinker 进程内的运行态 mock 注入入口。
// 这些字段不是单元测试 fake，而是给真实进程在测试环境里“伪造压力信号”用的。
// 示例：
//   - Enabled=true, PipelinePending=120000, GateWaitingTasks=220000
//     可直接把保护状态机推向 hardPaused。
//   - Enabled=true, MySQLStatus="degraded", MySQLConsecutiveFailures=5
//     可模拟 DB 异常恢复场景，而不必真的停 MySQL。
type ProtectionMockConfig struct {
	// Preset 用于选择内置测试/演练场景。
	// 它可以同时展开 mock 值、worker pool 边界和 soft backlog 阈值。
	// 示例：
	// - "consumer_low"：压低消费能力但不启用 mock
	// - "protect_mock"：压低消费能力并启用 mock 注入
	// - "consumer_low_soft_backlog"：压低消费能力并下调 soft backlog 阈值
	//
	// 推荐写法：
	// {
	//   "sinker": {
	//     "protection": {
	//       "mock": {
	//         "preset": "protect_mock",
	//         "enabled": true
	//       }
	//     }
	//   }
	// }
	//
	// 边界说明：
	// 1. preset 只建议用于测试/演练配置，不建议直接带进生产主配置。
	// 2. preset 不是“弱提示”，而是强约束：
	//    - 非法值会在启动时失败
	//    - 与显式配置冲突时会在启动时失败
	// 3. protect_mock 场景要求 mock.enabled=true；
	//    consumer_low_soft_backlog 场景要求 mock.enabled=false。
	Preset                           string   `json:"preset,omitempty"`
	Enabled                          bool     `json:"enabled"`
	ReportLag                        *int64   `json:"reportLag"`
	ReportSpeedPerSecond             *float64 `json:"reportSpeedPerSecond"`
	RealTimeLag                      *int64   `json:"realTimeLag"`
	RealTimeSpeedPerSecond           *float64 `json:"realTimeSpeedPerSecond"`
	PipelinePending                  *int     `json:"pipelinePending"`
	GateInFlight                     *int64   `json:"gateInFlight"`
	GateWaitingTasks                 *int64   `json:"gateWaitingTasks"`
	ReportConsumerQueueUsagePermille *int     `json:"reportConsumerQueueUsagePermille"`
	ReportConsumerBusyPermille       *int     `json:"reportConsumerBusyPermille"`
	ReportPersistQueueUsagePermille  *int     `json:"reportPersistQueueUsagePermille"`
	ReportPersistBusyPermille        *int     `json:"reportPersistBusyPermille"`
	HeapAllocBytes                   *uint64  `json:"heapAllocBytes"`
	Goroutines                       *int     `json:"goroutines"`
	PersistenceErrorCount            *int     `json:"persistenceErrorCount"`
	MySQLStatus                      *string  `json:"mysqlStatus"`
	MySQLConsecutiveFailures         *int     `json:"mysqlConsecutiveFailures"`
}

func (c ProtectionConfig) Normalize() ProtectionConfig {
	if c.Enabled == nil {
		c.Enabled = boolPtr(true)
	}
	if c.SampleIntervalSeconds <= 0 {
		c.SampleIntervalSeconds = 5
	}
	if c.NormalRateLogIntervalSeconds <= 0 {
		c.NormalRateLogIntervalSeconds = 180
	}
	if c.DefaultDiagnosticRateLogIntervalSeconds <= 0 {
		c.DefaultDiagnosticRateLogIntervalSeconds = 5
	}
	if c.RecoveryHealthyWindows <= 0 {
		c.RecoveryHealthyWindows = 3
	}
	if c.HardEscalationWindows <= 0 {
		c.HardEscalationWindows = 3
	}
	if c.SoftTargetRatePerSecond <= 0 {
		c.SoftTargetRatePerSecond = 1000
	}
	if c.SoftMinHoldSeconds <= 0 {
		c.SoftMinHoldSeconds = 60
	}
	if c.HardMinHoldSeconds <= 0 {
		c.HardMinHoldSeconds = 120
	}
	if c.MaxAdaptiveHoldSeconds <= 0 {
		c.MaxAdaptiveHoldSeconds = 600
	}
	if c.SoftThresholds.OrderedCommitPendingCount <= 0 {
		c.SoftThresholds.OrderedCommitPendingCount = 20000
	}
	if c.SoftThresholds.GateInFlightMessages <= 0 {
		c.SoftThresholds.GateInFlightMessages = 20000
	}
	if c.SoftThresholds.GateWaitingTasks <= 0 {
		c.SoftThresholds.GateWaitingTasks = 40000
	}
	if c.SoftThresholds.WorkerQueueUsagePermille <= 0 {
		c.SoftThresholds.WorkerQueueUsagePermille = 500
	}
	if c.SoftThresholds.WorkerBusyPermille <= 0 {
		c.SoftThresholds.WorkerBusyPermille = 900
	}
	if c.SoftThresholds.HeapAllocBytes <= 0 {
		c.SoftThresholds.HeapAllocBytes = 1 << 30
	}
	if c.SoftThresholds.Goroutines <= 0 {
		c.SoftThresholds.Goroutines = 2000
	}
	if c.SoftThresholds.SlowReportHandlerSamples <= 0 {
		c.SoftThresholds.SlowReportHandlerSamples = 50
	}
	if c.SoftThresholds.PersistenceErrors <= 0 {
		c.SoftThresholds.PersistenceErrors = 3
	}
	if c.HardThresholds.OrderedCommitPendingCount <= 0 {
		c.HardThresholds.OrderedCommitPendingCount = 100000
	}
	if c.HardThresholds.GateInFlightMessages <= 0 {
		c.HardThresholds.GateInFlightMessages = 100000
	}
	if c.HardThresholds.GateWaitingTasks <= 0 {
		c.HardThresholds.GateWaitingTasks = 200000
	}
	if c.HardThresholds.WorkerQueueUsagePermille <= 0 {
		c.HardThresholds.WorkerQueueUsagePermille = 800
	}
	if c.HardThresholds.WorkerBusyPermille <= 0 {
		c.HardThresholds.WorkerBusyPermille = 900
	}
	if c.HardThresholds.HeapAllocBytes <= 0 {
		c.HardThresholds.HeapAllocBytes = 2 << 30
	}
	if c.HardThresholds.Goroutines <= 0 {
		c.HardThresholds.Goroutines = 4000
	}
	if c.HardThresholds.PersistenceErrors <= 0 {
		c.HardThresholds.PersistenceErrors = 10
	}
	if c.HardThresholds.DBConsecutiveFailures <= 0 {
		c.HardThresholds.DBConsecutiveFailures = 3
	}
	return c
}

type DBHealthConfig struct {
	Enabled                *bool `json:"enabled"`
	PingIntervalSeconds    int   `json:"pingIntervalSeconds"`
	FailuresBeforeDegraded int   `json:"failuresBeforeDegraded"`
}

func (c DBHealthConfig) Normalize() DBHealthConfig {
	if c.Enabled == nil {
		c.Enabled = boolPtr(true)
	}
	if c.PingIntervalSeconds <= 0 {
		c.PingIntervalSeconds = 10
	}
	if c.FailuresBeforeDegraded <= 0 {
		c.FailuresBeforeDegraded = 3
	}
	return c
}

func boolPtr(v bool) *bool {
	return &v
}

func intPtr(v int) *int {
	return &v
}

func int64Ptr(v int64) *int64 {
	return &v
}

func float64Ptr(v float64) *float64 {
	return &v
}

func uint64Ptr(v uint64) *uint64 {
	return &v
}

func stringPtr(v string) *string {
	return &v
}

func validateConsumerLowPoolPreset(pool DynamicWorkerPoolConfigJSON, field string) error {
	if pool == (DynamicWorkerPoolConfigJSON{}) {
		return nil
	}
	if pool.MinWorkers == 1 &&
		pool.MaxWorkers == 1 &&
		pool.QueueSize == 64 &&
		pool.TuneInterval == 2 &&
		pool.DrainTimeout == 30 {
		return nil
	}
	return fmt.Errorf("protection.mock.preset requires %s to stay empty or match consumer_low preset", field)
}

func validateProtectMockPreset(mock ProtectionMockConfig) error {
	if !mock.Enabled {
		return fmt.Errorf("protection.mock.preset=protect_mock requires mock.enabled=true")
	}
	if !valueMatchesInt64(mock.ReportLag, 2500000) {
		return fmt.Errorf("protection.mock.preset=protect_mock conflicts with mock.reportLag")
	}
	if !valueMatchesFloat64(mock.ReportSpeedPerSecond, 0) {
		return fmt.Errorf("protection.mock.preset=protect_mock conflicts with mock.reportSpeedPerSecond")
	}
	if !valueMatchesInt64(mock.RealTimeLag, 500000) {
		return fmt.Errorf("protection.mock.preset=protect_mock conflicts with mock.realTimeLag")
	}
	if !valueMatchesFloat64(mock.RealTimeSpeedPerSecond, 0) {
		return fmt.Errorf("protection.mock.preset=protect_mock conflicts with mock.realTimeSpeedPerSecond")
	}
	if !valueMatchesInt(mock.PipelinePending, 120000) {
		return fmt.Errorf("protection.mock.preset=protect_mock conflicts with mock.pipelinePending")
	}
	if !valueMatchesInt64(mock.GateInFlight, 110000) {
		return fmt.Errorf("protection.mock.preset=protect_mock conflicts with mock.gateInFlight")
	}
	if !valueMatchesInt64(mock.GateWaitingTasks, 220000) {
		return fmt.Errorf("protection.mock.preset=protect_mock conflicts with mock.gateWaitingTasks")
	}
	if !valueMatchesInt(mock.ReportConsumerQueueUsagePermille, 850) {
		return fmt.Errorf("protection.mock.preset=protect_mock conflicts with mock.reportConsumerQueueUsagePermille")
	}
	if !valueMatchesInt(mock.ReportConsumerBusyPermille, 950) {
		return fmt.Errorf("protection.mock.preset=protect_mock conflicts with mock.reportConsumerBusyPermille")
	}
	if !valueMatchesInt(mock.ReportPersistQueueUsagePermille, 700) {
		return fmt.Errorf("protection.mock.preset=protect_mock conflicts with mock.reportPersistQueueUsagePermille")
	}
	if !valueMatchesInt(mock.ReportPersistBusyPermille, 900) {
		return fmt.Errorf("protection.mock.preset=protect_mock conflicts with mock.reportPersistBusyPermille")
	}
	if !valueMatchesUint64(mock.HeapAllocBytes, 2147483648) {
		return fmt.Errorf("protection.mock.preset=protect_mock conflicts with mock.heapAllocBytes")
	}
	if !valueMatchesInt(mock.Goroutines, 4500) {
		return fmt.Errorf("protection.mock.preset=protect_mock conflicts with mock.goroutines")
	}
	if !valueMatchesInt(mock.PersistenceErrorCount, 12) {
		return fmt.Errorf("protection.mock.preset=protect_mock conflicts with mock.persistenceErrorCount")
	}
	if !valueMatchesString(mock.MySQLStatus, "degraded") {
		return fmt.Errorf("protection.mock.preset=protect_mock conflicts with mock.mysqlStatus")
	}
	if !valueMatchesInt(mock.MySQLConsecutiveFailures, 5) {
		return fmt.Errorf("protection.mock.preset=protect_mock conflicts with mock.mysqlConsecutiveFailures")
	}
	return nil
}

func validateConsumerLowSoftBacklogPreset(protection ProtectionConfig) error {
	if protection.Mock.Enabled {
		return fmt.Errorf("protection.mock.preset=consumer_low_soft_backlog requires mock.enabled=false")
	}

	thresholds := protection.SoftThresholds
	if thresholds.OrderedCommitPendingCount != 0 && thresholds.OrderedCommitPendingCount != 20000 && thresholds.OrderedCommitPendingCount != 500 {
		return fmt.Errorf("protection.mock.preset=consumer_low_soft_backlog conflicts with softThresholds.orderedCommitPendingCount")
	}
	if thresholds.GateInFlightMessages != 0 && thresholds.GateInFlightMessages != 20000 && thresholds.GateInFlightMessages != 500 {
		return fmt.Errorf("protection.mock.preset=consumer_low_soft_backlog conflicts with softThresholds.gateInFlightMessages")
	}
	if thresholds.GateWaitingTasks != 0 && thresholds.GateWaitingTasks != 40000 && thresholds.GateWaitingTasks != 1000 {
		return fmt.Errorf("protection.mock.preset=consumer_low_soft_backlog conflicts with softThresholds.gateWaitingTasks")
	}
	return nil
}

func valueMatchesInt(actual *int, expected int) bool {
	return actual == nil || *actual == expected
}

func valueMatchesInt64(actual *int64, expected int64) bool {
	return actual == nil || *actual == expected
}

func valueMatchesFloat64(actual *float64, expected float64) bool {
	return actual == nil || *actual == expected
}

func valueMatchesUint64(actual *uint64, expected uint64) bool {
	return actual == nil || *actual == expected
}

func valueMatchesString(actual *string, expected string) bool {
	return actual == nil || *actual == expected
}

func (c BatchConfig) Normalize() BatchConfig {
	if c.BufferSize <= 0 {
		c.BufferSize = 1000
	}
	if c.FlushInterval <= 0 {
		c.FlushInterval = 2
	}
	return c
}

func applyConsumerLowPoolPreset(c *SinkerConfig) {
	c.ReportConsumerPool.MinWorkers = 1
	c.ReportConsumerPool.MaxWorkers = 1
	c.ReportConsumerPool.QueueSize = 64
	c.ReportConsumerPool.TuneInterval = 2
	c.ReportConsumerPool.DrainTimeout = 30

	c.ReportPersistPool.MinWorkers = 1
	c.ReportPersistPool.MaxWorkers = 1
	c.ReportPersistPool.QueueSize = 64
	c.ReportPersistPool.TuneInterval = 2
	c.ReportPersistPool.DrainTimeout = 30
}

// applyProtectMockPreset 展开“低消费 + mock 注入”场景。
//
// 这个场景用于验证 hard/soft breaker 的状态切换和日志链路，
// 不用于证明真实外部压力已经存在。
func applyProtectMockPreset(protection *ProtectionConfig) {
	protection.Mock.ReportLag = int64Ptr(2500000)
	protection.Mock.ReportSpeedPerSecond = float64Ptr(0)
	protection.Mock.RealTimeLag = int64Ptr(500000)
	protection.Mock.RealTimeSpeedPerSecond = float64Ptr(0)
	protection.Mock.PipelinePending = intPtr(120000)
	protection.Mock.GateInFlight = int64Ptr(110000)
	protection.Mock.GateWaitingTasks = int64Ptr(220000)
	protection.Mock.ReportConsumerQueueUsagePermille = intPtr(850)
	protection.Mock.ReportConsumerBusyPermille = intPtr(950)
	protection.Mock.ReportPersistQueueUsagePermille = intPtr(700)
	protection.Mock.ReportPersistBusyPermille = intPtr(900)
	protection.Mock.HeapAllocBytes = uint64Ptr(2147483648)
	protection.Mock.Goroutines = intPtr(4500)
	protection.Mock.PersistenceErrorCount = intPtr(12)
	protection.Mock.MySQLStatus = stringPtr("degraded")
	protection.Mock.MySQLConsecutiveFailures = intPtr(5)
}

// applyConsumerLowSoftBacklogPreset 展开“低消费 + 更早命中 soft backlog 阈值”的场景。
//
// 这个场景用于验证“纯 backlog 足以驱动 soft_limited 的机制”，
// 但它不代表默认生产阈值也会自然触发。
func applyConsumerLowSoftBacklogPreset(protection *ProtectionConfig) {
	protection.SoftThresholds.OrderedCommitPendingCount = 500
	protection.SoftThresholds.GateInFlightMessages = 500
	protection.SoftThresholds.GateWaitingTasks = 1000
}

func (c BatchConfig) GetBufferSize() int {
	return c.Normalize().BufferSize
}

func (c BatchConfig) GetFlushInterval() int {
	return c.Normalize().FlushInterval
}

// DownloadConfigFile 下载配置文件
func DownloadConfigFile(fname string) (err error) {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	var config Config
	config.Sinker = config.Sinker.Normalize()
	config.Comm.Mysql = config.Comm.Mysql.Normalize()
	filePtr, err := os.Create(fname)
	if err != nil {
		return errors.New(fmt.Sprintf("创建配置文件异常:%s", err.Error()))
	}
	defer filePtr.Close()
	// 带JSON缩进格式写文件
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return errors.New(fmt.Sprintf("创建配置文件异常:%s", err.Error()))
	}
	_, err = filePtr.Write(data)
	return
}
