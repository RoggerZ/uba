package report

import (
	"fmt"
	"time"

	"github.com/1340691923/xwl_bi/engine/db"
	"github.com/1340691923/xwl_bi/engine/logs"
	"github.com/1340691923/xwl_bi/model"
	"github.com/IBM/sarama"
	jsoniter "github.com/json-iterator/go"
	"go.uber.org/zap"
)

// KafkaDataProducer 负责把正式上报数据发送到 report topic。
type KafkaDataProducer struct {
	topic    string
	producer *jsonTopicProducer
}

func NewKafkaDataProducer(topic string, producer *jsonTopicProducer) *KafkaDataProducer {
	if producer == nil {
		producer = newJSONTopicProducer()
	}
	return &KafkaDataProducer{
		topic:    topic,
		producer: producer,
	}
}

func NewDefaultKafkaDataProducer() *KafkaDataProducer {
	return NewKafkaDataProducer(model.GlobConfig.Comm.Kafka.ReportTopicName, newJSONTopicProducer())
}

// Send 把正式 KafkaData 发送到 report topic。
func (p *KafkaDataProducer) Send(data model.KafkaData) error {
	// 这里在真正发送前后都补一层调试日志，目的是把“report_service 已经把什么数据准备发往哪个 topic”
	// 这件事记录下来，便于排查：
	// 1. 请求明明进入了 report_server，但 Kafka 里没有消息。
	// 2. 某条消息写到了错误 topic。
	// 3. 某条消息的 tableId / eventName / reportType 与预期不一致。
	//
	// 示例：
	// 1. 发送前日志里可以看到 topic=report-topic, appid=1001, table_id=51, event_name=pay_success。
	// 2. 发送失败时，会在发送后日志里带上同一组上下文字段和 error。
	logs.Logger.Debug("report kafka send begin", reportKafkaDataLogFields(p.topic, data)...)

	err := p.producer.Send(p.topic, data)
	if err != nil {
		logs.Logger.Debug("report kafka send failed", append(reportKafkaDataLogFields(p.topic, data), zap.Error(err))...)
		return err
	}

	logs.Logger.Debug("report kafka send success", reportKafkaDataLogFields(p.topic, data)...)
	return nil
}

// DebugDataProducer 负责把 debug 检查结果发送到 debug topic。
type DebugDataProducer struct {
	topic    string
	producer *jsonTopicProducer
}

func NewDebugDataProducer(topic string, producer *jsonTopicProducer) *DebugDataProducer {
	if producer == nil {
		producer = newJSONTopicProducer()
	}
	return &DebugDataProducer{
		topic:    topic,
		producer: producer,
	}
}

func NewDefaultDebugDataProducer() *DebugDataProducer {
	return NewDebugDataProducer(model.GlobConfig.Comm.Kafka.DebugDataTopicName, newJSONTopicProducer())
}

// Send 把 debug 检查账本写入 debug topic。
func (p *DebugDataProducer) Send(data map[string]interface{}) error {
	// debug topic 的日志重点不是原始 payload 全量展开，而是打出最关键的定位字段：
	// 1. data_name：哪一个事件。
	// 2. appid：这里保持历史语义，实际是 tableId。
	// 3. distinct_id：哪一个设备或用户。
	// 4. data_judge / error_reason：这次 debug 判定结果是什么。
	logs.Logger.Debug("report debug kafka send begin", reportDebugDataLogFields(p.topic, data)...)

	err := p.producer.Send(p.topic, data)
	if err != nil {
		logs.Logger.Debug("report debug kafka send failed", append(reportDebugDataLogFields(p.topic, data), zap.Error(err))...)
		return err
	}

	logs.Logger.Debug("report debug kafka send success", reportDebugDataLogFields(p.topic, data)...)
	return nil
}

// jsonTopicProducer 是 report 包内部复用的 JSON Kafka producer 小封装。
//
// 它没有独立放到新目录，原因是：
// 1. 当前只有 report 领域在使用它。
// 2. 把它留在 report 包内，依赖关系最直接。
// 3. 如果未来确实被其他模块复用，再抽公共层会更合适。
type jsonTopicProducer struct {
	json jsoniter.API
	now  func() time.Time
	send func(msg *sarama.ProducerMessage) error
}

func newJSONTopicProducer() *jsonTopicProducer {
	return &jsonTopicProducer{
		json: jsoniter.ConfigCompatibleWithStandardLibrary,
		now:  time.Now,
		send: sendProducerMessage,
	}
}

func newJSONTopicProducerWithSender(send func(msg *sarama.ProducerMessage) error) *jsonTopicProducer {
	producer := newJSONTopicProducer()
	if send != nil {
		producer.send = send
	}
	return producer
}

// Send 把 payload 序列化为 JSON 后发送到指定 topic。
func (p *jsonTopicProducer) Send(topic string, payload interface{}) error {
	sendData, err := p.json.Marshal(payload)
	if err != nil {
		return err
	}

	msg := &sarama.ProducerMessage{
		Topic:     topic,
		Value:     sarama.ByteEncoder(sendData),
		Timestamp: p.now(),
	}
	return p.send(msg)
}

func sendProducerMessage(msg *sarama.ProducerMessage) error {
	switch resolveProducerType() {
	case "async":
		db.KafkaASyncProducer.Input() <- msg
		return nil
	default:
		_, _, err := db.KafkaSyncProducer.SendMessage(msg)
		return err
	}
}

// reportKafkaDataLogFields 统一构造正式上报 Kafka 发送日志字段。
//
// 这里特意只打“能快速定位一条上报消息”的关键字段，而不是把整个 ReqData 原文打出来：
// 1. 避免日志过大。
// 2. 避免把业务 body 原样刷进日志增加排查噪声。
// 3. 真要看原始 body 时，仍然可以从上游请求日志或 debug topic 里查。
func reportKafkaDataLogFields(topic string, data model.KafkaData) []zap.Field {
	return []zap.Field{
		zap.String("topic", topic),
		zap.String("producer_type", resolveProducerType()),
		zap.String("appid", data.APPID),
		zap.String("table_id", data.TableId),
		zap.String("event_name", data.EventName),
		zap.Int("report_type", data.ReportType),
		zap.String("debug", data.Debug),
		zap.String("report_time", data.ReportTime),
		zap.Int("req_data_len", len(data.ReqData)),
	}
}

// reportDebugDataLogFields 统一构造 debug topic 发送日志字段。
//
// 示例：
// 1. data_judge=数据检验通过 时，可以快速看到这条 debug 请求已通过检查。
// 2. error_reason 不为空时，可以直接从日志看出失败原因，而不用再解开 Kafka 消息。
func reportDebugDataLogFields(topic string, data map[string]interface{}) []zap.Field {
	return []zap.Field{
		zap.String("topic", topic),
		zap.String("producer_type", resolveProducerType()),
		zap.String("data_name", stringifyLogValue(data["data_name"])),
		zap.String("appid", stringifyLogValue(data["appid"])),
		zap.String("distinct_id", stringifyLogValue(data["distinct_id"])),
		zap.String("report_time", stringifyLogValue(data["report_time"])),
		zap.String("data_judge", stringifyLogValue(data["data_judge"])),
		zap.String("error_reason", stringifyLogValue(data["error_reason"])),
	}
}

func stringifyLogValue(value interface{}) string {
	if value == nil {
		return ""
	}
	return fmt.Sprint(value)
}

// resolveProducerType 统一决定当前发送边界实际采用的 producer 类型。
//
// 这里单独抽成函数，是为了让：
// 1. 真实发送分支
// 2. 调试日志里的 producer_type 字段
// 使用完全一致的判定逻辑，避免日志写 async、代码却走 sync 这种偏差。
func resolveProducerType() string {
	switch model.GlobConfig.GetKafkaCfgProducerType() {
	case "async":
		return "async"
	default:
		return "sync"
	}
}
