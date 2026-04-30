package runner

import (
	"github.com/1340691923/xwl_bi/engine/logs"
	"go.uber.org/zap"
)

// consumeKafkaConsumerErrors 负责持续消费某个 Kafka consumer group 的异步错误通道。
//
// 这类错误和业务 handler 返回错误不是一回事：
// 1. handler 错误代表某条消息的业务处理失败。
// 2. ConsumerGroup.Errors() 代表 Sarama 消费层面的异步错误，例如底层分区消费异常、broker 通道问题等。
//
// 为什么要单独起后台协程消费它：
// 1. Sarama 已经提供了错误通道，如果不读，就会丢失这部分排障信息。
// 2. 这些错误不一定会直接让 Consume(...) 返回，因此只看 Run 循环日志并不完整。
//
// 示例：
// 1. `report-consumer` 收到某次 rebalance 相关消费错误
// 2. 日志会带上 consumer_name 和 topic，方便区分是哪条链路出的错
func consumeKafkaConsumerErrors(consumerName, topic string, errCh <-chan error) {
	if errCh == nil {
		return
	}

	for err := range errCh {
		if err == nil {
			continue
		}
		logs.Logger.Error(
			"kafka consumer async error",
			zap.String("consumer_name", consumerName),
			zap.String("topic", topic),
			zap.Error(err),
		)
	}
}
