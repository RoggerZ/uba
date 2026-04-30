package sinker

import (
	"context"
	"github.com/1340691923/xwl_bi/engine/logs"
	"github.com/1340691923/xwl_bi/model"
	"github.com/1340691923/xwl_bi/platform-basic-libs/util"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"github.com/IBM/sarama"
	"github.com/pkg/errors"

	"go.uber.org/zap"
)

type KafkaSarama struct {
	topic             string
	group             string
	cfg               model.KafkaCfg
	cg                sarama.ConsumerGroup
	sess              sarama.ConsumerGroupSession
	ctx               context.Context
	cancel            context.CancelFunc
	wgRun             sync.WaitGroup
	putFn             func(msg model.InputMessage, markFn func())
	cleanupFn         func(generationID int32)
	markStats         sync.Map
	lastMarkedOffsets sync.Map
	dispatchRegulator interface{ Wait() }
	paused            uint32
}

func NewKafkaSarama() *KafkaSarama {
	return &KafkaSarama{}
}

type MyConsumerGroupHandler struct {
	k *KafkaSarama
}

func (h MyConsumerGroupHandler) Setup(sess sarama.ConsumerGroupSession) error {
	h.k.sess = sess
	// Setup 表示当前 consumer 已经加入某一代 consumer group，
	// 从这一刻开始才真正拿到了分区归属。
	logs.Logger.Info("consumer group setup",
		zap.String("topic", h.k.topic),
		zap.String("group", h.k.group),
		zap.Int32("generation id", sess.GenerationID()))
	return nil
}

func (h MyConsumerGroupHandler) Cleanup(sess sarama.ConsumerGroupSession) error {
	begin := time.Now()
	h.k.cleanupFn(sess.GenerationID())
	h.k.clearMarkedOffsets(sess.Claims()[h.k.topic])
	// Cleanup 通常发生在 rebalance、关闭 consumer group、或 session 结束时。
	// 这里把耗时打出来，便于观察某次 rebalance 清理是否异常变慢。
	logs.Logger.Info("consumer group cleanup",
		zap.String("group", h.k.group),
		zap.Int32("generation id", sess.GenerationID()),
		zap.Duration("cost", time.Since(begin)))
	return nil
}

func (h MyConsumerGroupHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	generationID := sess.GenerationID()
	// Sarama 在同一个 claim 内是串行回调 putFn 的。
	// 这意味着：
	// 1. 当前消息如果在 putFn 里阻塞，后面的消息就不会继续进来。
	// 2. 所以业务侧必须避免在 putFn 里长时间持锁或做不可控慢操作。
	for msg := range claim.Messages() {
		// 这里显式包一层闭包，是为了在单条消息级别兜住 panic。
		// 否则一条异常消息可能直接把整个 claim 消费线程打崩。
		func() {
			defer func() {
				if r := recover(); r != nil {
					// 出现 panic 时，除了打印堆栈，也直接 mark 当前消息，
					// 避免同一条毒数据反复卡住整个分区。
					logs.Logger.Error("KafkaSarama putFn panic",
						zap.Any("panic", r),
						zap.String("group", h.k.group),
						zap.String("topic", msg.Topic),
						zap.Int32("partition", msg.Partition),
						zap.Int64("offset", msg.Offset),
						zap.ByteString("stack", debug.Stack()))
					sess.MarkMessage(msg, "")
				}
			}()

			if h.k.dispatchRegulator != nil {
				h.k.dispatchRegulator.Wait()
			}

			h.k.putFn(model.InputMessage{
				Topic:        msg.Topic,
				Partition:    int(msg.Partition),
				Key:          msg.Key,
				Value:        msg.Value,
				Offset:       msg.Offset,
				Timestamp:    &msg.Timestamp,
				GenerationID: generationID,
			}, func() {
				h.k.storeMarkedOffset(msg.Partition, msg.Offset+1)
				if util.IsSinkerDiagnosticLogEnabled() && h.k.shouldLogMarkMessage(generationID) {
					logs.Logger.Info(
						"consumer session mark message",
						zap.String("group", h.k.group),
						zap.Int32("generation_id", generationID),
						zap.String("topic", msg.Topic),
						zap.Int32("partition", msg.Partition),
						zap.Int64("offset", msg.Offset),
					)
				}
				sess.MarkMessage(msg, "")
			})
		}()
	}
	return nil
}

func (k *KafkaSarama) Init(cfg model.KafkaCfg, topicName, consumerGroup string, putFn func(msg model.InputMessage, markFn func()), cleanupFn func(generationID int32)) (err error) {
	k.cfg = cfg
	k.ctx, k.cancel = context.WithCancel(context.Background())
	k.putFn = putFn
	k.cleanupFn = cleanupFn
	k.topic = topicName
	k.group = consumerGroup
	sarCfg, err := GetSaramaConfig(cfg)
	if err != nil {
		return err
	}
	sarCfg.Consumer.Offsets.Initial = sarama.OffsetOldest

	cg, err := sarama.NewConsumerGroup(cfg.Addresses, consumerGroup, sarCfg)
	if err != nil {
		return err
	}

	k.cg = cg
	return nil
}

func (k *KafkaSarama) shouldLogMarkMessage(generationID int32) bool {
	actual, _ := k.markStats.LoadOrStore(generationID, new(int64))
	counter := actual.(*int64)
	count := atomic.AddInt64(counter, 1)
	return count <= 10 || count%1000 == 0
}

func GetSaramaConfig(kfkCfg model.KafkaCfg) (sarCfg *sarama.Config, err error) {
	sarCfg = sarama.NewConfig()
	sarCfg.Version = sarama.V2_0_0_0
	// 开启 Consumer.Return.Errors 的原因是：
	// 1. Sarama 会把部分内部消费错误通过 ConsumerGroup.Errors() 异步上报。
	// 2. 如果这里保持 false，runtime 就无法统一消费这些错误通道，排障信息会丢失。
	// 3. report_server 已经对 async producer 做了错误通道消费，sinker 这里补齐 consumer 错误消费后，链路观测会更完整。
	sarCfg.Consumer.Return.Errors = true
	// 如果配置了用户名和密码，就开启 SASL 认证。
	if kfkCfg.Username != "" && kfkCfg.Password != "" {
		sarCfg.Net.SASL.Enable = true
		sarCfg.Net.SASL.User = kfkCfg.Username
		sarCfg.Net.SASL.Password = kfkCfg.Password
	}
	sarCfg.ChannelBufferSize = 1024
	return
}

func (k *KafkaSarama) storeMarkedOffset(partition int32, offset int64) {
	actual, _ := k.lastMarkedOffsets.LoadOrStore(partition, new(int64))
	atomic.StoreInt64(actual.(*int64), offset)
}

func (k *KafkaSarama) clearMarkedOffsets(partitions []int32) {
	for _, partition := range partitions {
		k.lastMarkedOffsets.Delete(partition)
	}
}

func (k *KafkaSarama) CurrentMarkedOffset() int64 {
	var current int64
	for _, offset := range k.CurrentMarkedOffsets() {
		current += offset
	}
	return current
}

// CurrentMarkedOffsets 返回当前进程内按 partition 记录的最新 mark offset 快照。
//
// 这个快照只反映当前 sinker 进程自启动以来、已经走到 markFn 的消息推进值。
// 它可能领先于 Kafka coordinator 上已提交的 committed offset，也可能在冷启动时为空。
// 因此运行态观测场景应优先把它和 committed offsets 结合使用，而不是单独作为权威值。
func (k *KafkaSarama) CurrentMarkedOffsets() map[int32]int64 {
	offsets := make(map[int32]int64)
	k.lastMarkedOffsets.Range(func(key, value any) bool {
		offsets[key.(int32)] = atomic.LoadInt64(value.(*int64))
		return true
	})
	return offsets
}

func (k *KafkaSarama) SetConsumeRegulator(regulator interface{ Wait() }) {
	k.dispatchRegulator = regulator
}

func (k *KafkaSarama) PauseConsumption() {
	type pauser interface {
		PauseAll()
	}

	if group, ok := k.cg.(pauser); ok {
		group.PauseAll()
		atomic.StoreUint32(&k.paused, 1)
	}
}

func (k *KafkaSarama) ResumeConsumption() {
	type resumer interface {
		ResumeAll()
	}

	if group, ok := k.cg.(resumer); ok {
		group.ResumeAll()
		atomic.StoreUint32(&k.paused, 0)
	}
}

func (k *KafkaSarama) IsPaused() bool {
	return atomic.LoadUint32(&k.paused) == 1
}

func (k *KafkaSarama) Run() {
	k.wgRun.Add(1)
	defer k.wgRun.Done()
loopSarama:
	for {
		// Sarama 的 Consume 必须放在循环里不断重入。
		//
		// 原因是：
		// 1. rebalance 发生后，旧 session 会自然结束。
		// 2. 这时 Consume 往往返回 nil，而不是报错。
		// 3. 只有继续下一轮 Consume，当前 consumer 才能重新加入下一代 group。
		handler := MyConsumerGroupHandler{k}
		if k.ctx.Err() != nil {
			return
		}

		if err := k.cg.Consume(k.ctx, []string{k.topic}, handler); err != nil {
			if errors.Is(err, context.Canceled) {
				logs.Logger.Info("KafkaSarama.Run quit due to context has been canceled", zap.String("task", k.topic))
				break loopSarama
			} else if errors.Is(err, sarama.ErrClosedConsumerGroup) {
				logs.Logger.Info("KafkaSarama.Run quit due to consumer group has been closed", zap.String("task", k.topic))
				break loopSarama
			} else {
				logs.Logger.Error("sarama.ConsumerGroup.Consume failed", zap.String("task", k.topic), zap.Error(err))
				continue
			}
		}
	}
}

func (k *KafkaSarama) CommitMessages(msg *model.InputMessage) error {
	k.sess.MarkOffset(msg.Topic, int32(msg.Partition), msg.Offset+1, "")
	return nil
}

func (k *KafkaSarama) Stop() error {
	k.cancel()
	_ = k.cg.Close()
	k.wgRun.Wait()
	return nil
}

// Errors 暴露底层 ConsumerGroup 的异步错误通道，供 runtime 后台协程统一消费。
//
// 注意：
// 1. 这里返回的是 Sarama ConsumerGroup 自带的错误通道，不是 putFn 或业务 handler 的错误。
// 2. 业务 handler 的错误仍然由主流程自己记录日志和控制 offset 语义。
// 3. 只有开启 Consumer.Return.Errors=true 后，这个通道才会真正有数据流出。
func (k *KafkaSarama) Errors() <-chan error {
	if k.cg == nil {
		return nil
	}
	return k.cg.Errors()
}

func (k *KafkaSarama) Description() string {
	return "kafka consumer of topic " + k.topic
}
