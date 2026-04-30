package runner

import (
	"github.com/1340691923/xwl_bi/model"
	"github.com/1340691923/xwl_bi/platform-basic-libs/service/consumer_data"
	parser "github.com/1340691923/xwl_bi/platform-basic-libs/sinker/parse"
)

// realTimeSink 抽象“实时表批次写入端”。
// handler 只依赖 Add，不直接依赖具体批量器实现。
type realTimeSink interface {
	Add(*consumer_data.RealTimeWarehousingData) error
}

// acceptanceStatusSink 抽象“验收状态写入端”。
//
// 它存在的意义是把 handler 和具体批量器解耦：
// 1. handler 只关心“我要记录一条成功/失败状态”。
// 2. 至于状态是先进内存批次、还是将来换成别的实现，不由 handler 关心。
type acceptanceStatusSink interface {
	Add(*consumer_data.ReportAcceptStatusData) error
}

// metricSink 抽象“CK 明细批次写入端”。
//
// 示例：
//  1. 当前实现是 ReportData2CK。
//  2. 如果将来要改成另一种批量器或异步队列，只要实现 Add 即可，
//     report handler 本身不需要改主流程。
type metricSink interface {
	Add(consumer_data.FastjsonMetricData) error
}

// schemaSynchronizer 抽象“动态补列与列缓存刷新”能力。
//
// 它把明细 handler 和具体的补列实现隔离开，
// 避免主 handler 直接依赖一个大而重的 action 文件。
type schemaSynchronizer interface {
	EnsureColumns(model.KafkaData, string, *parser.FastjsonMetric, func(consumer_data.ReportAcceptStatusData)) error
}

// metaEventRecorder 抽象“事件元数据记录”能力。
type metaEventRecorder interface {
	Record(model.KafkaData) error
}

// geoPayloadEnricher 抽象“根据 IP 补充地理字段”能力。
//
// 输入：
// 1. 原始 ReqData JSON
// 2. 上报 IP
//
// 输出：
// 1. 补充过 xwl_ip / xwl_city / xwl_country 等字段后的 JSON
// 2. 查询过程中的错误
type geoPayloadEnricher interface {
	Enrich([]byte, string) ([]byte, error)
}

// schemaSynchronizerFunc 让简单函数也能直接适配 schemaSynchronizer 接口。
type schemaSynchronizerFunc func(model.KafkaData, string, *parser.FastjsonMetric, func(consumer_data.ReportAcceptStatusData)) error

func (f schemaSynchronizerFunc) EnsureColumns(kafkaData model.KafkaData, tableName string, metric *parser.FastjsonMetric, failFunc func(consumer_data.ReportAcceptStatusData)) error {
	return f(kafkaData, tableName, metric, failFunc)
}

type actionSchemaSynchronizer struct {
	addColumns func(model.KafkaData, func(consumer_data.ReportAcceptStatusData), string, *parser.FastjsonMetric) error
}

// newActionSchemaSynchronizer 把 action.AddTableColumn 这种现有函数包装成接口实现。
//
// 这样做的目的是避免为了重构而大改 action 包对外签名，
// 同时又能让 handler 依赖抽象而不是依赖具体函数。
func newActionSchemaSynchronizer(addColumns func(model.KafkaData, func(consumer_data.ReportAcceptStatusData), string, *parser.FastjsonMetric) error) *actionSchemaSynchronizer {
	return &actionSchemaSynchronizer{addColumns: addColumns}
}

// EnsureColumns 只是一个薄适配层，不新增业务语义。
func (s *actionSchemaSynchronizer) EnsureColumns(kafkaData model.KafkaData, tableName string, metric *parser.FastjsonMetric, failFunc func(consumer_data.ReportAcceptStatusData)) error {
	return s.addColumns(kafkaData, failFunc, tableName, metric)
}

// metaEventRecorderFunc 让简单函数直接适配 metaEventRecorder 接口。
type metaEventRecorderFunc func(model.KafkaData) error

func (f metaEventRecorderFunc) Record(kafkaData model.KafkaData) error {
	return f(kafkaData)
}
