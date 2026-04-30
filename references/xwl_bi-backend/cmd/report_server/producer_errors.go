package main

import (
	"github.com/1340691923/xwl_bi/engine/logs"
	"github.com/IBM/sarama"
	"go.uber.org/zap"
)

// consumeAsyncProducerErrors 负责持续消费 Kafka async producer 的错误通道。
//
// 这样做的原因是：
// 1. sarama async producer 的错误如果无人消费，排障信息会丢失。
// 2. 独立成函数后，既方便在 runtime 中复用，也方便单元测试验证退出语义。
func consumeAsyncProducerErrors(errCh <-chan *sarama.ProducerError) {
	for err := range errCh {
		if err == nil {
			continue
		}
		logs.Logger.Error("db.KafkaASyncProducer.Errors", zap.Error(err))
	}
}
