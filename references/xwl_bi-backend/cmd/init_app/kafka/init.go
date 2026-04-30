package kafka

import (
	"fmt"
	"github.com/1340691923/xwl_bi/model"
	"github.com/IBM/sarama"
	"log"
	"time"
)

//初始化kafka数据
func Init() {
	config := sarama.NewConfig()

	config.Version = sarama.V2_0_0_0
	// 单 broker 首次创建高分区数 topic 时会明显变慢。
	// 这里显式放宽 Admin/网络超时，避免 300 分区这类初始化请求在 broker 已经可用时仍被过早判定失败。
	config.Admin.Timeout = 2 * time.Minute
	config.Net.DialTimeout = 30 * time.Second
	config.Net.ReadTimeout = 2 * time.Minute
	config.Net.WriteTimeout = 2 * time.Minute
	if model.GlobConfig.Comm.Kafka.Username != "" {
		config.Net.SASL.Enable = true
		config.Net.SASL.User = model.GlobConfig.Comm.Kafka.Username
		config.Net.SASL.Password = model.GlobConfig.Comm.Kafka.Password
		config.Net.SASL.Handshake = true
	}

	config.Consumer.Group.Session.Timeout = 15 * time.Second
	config.Consumer.Group.Heartbeat.Interval = 5 * time.Second

	conn, err := sarama.NewClusterAdmin(model.GlobConfig.Comm.Kafka.Addresses, config)
	if err != nil {
		log.Println(fmt.Sprintf("kafka 链接初始化失败:%s", err.Error()))
		panic(err)
	}
	s, err := conn.ListTopics()
	for topic := range s {
		log.Println("您所拥有的TOPIC为：", topic)
	}

	if _, ok := s[model.GlobConfig.Comm.Kafka.ReportTopicName]; !ok {
		detail := sarama.TopicDetail{NumPartitions: model.GlobConfig.Comm.Kafka.NumPartitions, ReplicationFactor: 1}
		err = conn.CreateTopic(model.GlobConfig.Comm.Kafka.ReportTopicName, &detail, false)
		if err != nil {
			log.Println("创建TOPIC失败！", model.GlobConfig.Comm.Kafka.ReportTopicName)
			panic(err)
		}

		err = conn.Close()
		if err != nil {
			panic(err)
		}
		log.Println("初始化TOPIC完成！", model.GlobConfig.Comm.Kafka.ReportTopicName)
	} else {
		log.Println("您已拥有该TOPIC：", model.GlobConfig.Comm.Kafka.ReportTopicName)
	}
}
