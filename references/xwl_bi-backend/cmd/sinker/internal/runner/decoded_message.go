package runner

import (
	"fmt"

	"github.com/1340691923/xwl_bi/engine/logs"
	"github.com/1340691923/xwl_bi/model"
	"github.com/1340691923/xwl_bi/platform-basic-libs/service/consumer_data"
	"github.com/1340691923/xwl_bi/platform-basic-libs/util"
	jsoniter "github.com/json-iterator/go"
	"go.uber.org/zap"
)

// DecodedMessage 表示已经完成公共解码和公共字段补齐的一条消息。
//
// 它存在的目的，是把“消息公共准备动作”和“链路特有业务动作”拆开：
// 1. 公共层统一负责 json.Unmarshal、report_time 归一化、Offset 注入、ConsumptionTime 注入。
// 2. 实时链路和明细链路只处理各自差异逻辑，不再各自重复解码。
//
// 字段含义：
// 1. Input 保留原始 Kafka 输入，便于后续需要时访问 topic/partition/key。
// 2. KafkaData 是已经标准化后的业务消息。
// 3. MarkFn 保留原始提交函数，供不同链路按自己的时机调用。
type DecodedMessage struct {
	Input     model.InputMessage
	KafkaData model.KafkaData
	MarkFn    func()
}

type decodedMessageHandler interface {
	HandleDecoded(decoded DecodedMessage)
}

// messageDecoder 负责所有 consumer 共享的消息解码过程。
type messageDecoder struct {
	json jsoniter.API
}

func newMessageDecoder(json jsoniter.API) *messageDecoder {
	return &messageDecoder{json: json}
}

// Decode 执行公共解码逻辑。
//
// 处理步骤：
// 1. 把 Kafka 的 Value 反序列化成 KafkaData。
// 2. 统一归一化 report_time。
// 3. 注入 offset。
// 4. 注入服务端消费时间。
//
// 示例：
// 1. 输入 report_time="2026-04-08"
// 2. 输出 KafkaData.ReportTime="2026-04-08 00:00:00"
func (d *messageDecoder) Decode(msg model.InputMessage, markFn func()) (DecodedMessage, error) {
	var kafkaData model.KafkaData
	if err := d.json.Unmarshal(msg.Value, &kafkaData); err != nil {
		return DecodedMessage{}, err
	}

	kafkaData.ReportTime = consumer_data.NormalizeReportTime(kafkaData.ReportTime)
	kafkaData.Offset = msg.Offset
	if msg.Timestamp != nil {
		kafkaData.ConsumptionTime = msg.Timestamp.Local().Format(util.TimeFormat)
	}

	return DecodedMessage{
		Input:     msg,
		KafkaData: kafkaData,
		MarkFn:    markFn,
	}, nil
}

// Wrap 把旧的 KafkaSarama 回调签名适配成“先公共解码、后进入具体 handler”的调用方式。
//
// 对于解码失败，统一处理策略是：
// 1. 记录错误日志。
// 2. 直接 mark 当前消息，避免同一条非法外层消息卡死消费。
func (d *messageDecoder) Wrap(handle decodedMessageHandler) func(model.InputMessage, func()) {
	return func(msg model.InputMessage, markFn func()) {
		decoded, err := d.Decode(msg, markFn)
		if err != nil {
			logs.Logger.Error("decode kafka message failed", zap.Error(fmt.Errorf("decode input message: %w", err)))
			markFn()
			return
		}
		handle.HandleDecoded(decoded)
	}
}
