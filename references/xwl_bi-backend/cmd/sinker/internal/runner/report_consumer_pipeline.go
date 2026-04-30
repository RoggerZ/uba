package runner

import (
	"runtime/debug"
	"sort"

	"github.com/1340691923/xwl_bi/engine/logs"
	"github.com/1340691923/xwl_bi/platform-basic-libs/util"
	"go.uber.org/zap"
)

type reportMessageProcessor interface {
	ProcessDecoded(decoded DecodedMessage, gate *reportCompletionGate) reportProcessResult
}

// reportConsumerPipeline 把“worker pool + 分区内顺序提交”两件事串起来。
//
// 这里不再自行做 Decode，公共解码统一交给 messageDecoder.Wrap。
// 换句话说，这一层只消费已经完成公共解码的 DecodedMessage。
//
// 它的职责非常明确：
// 1. Register 当前 offset 的顺序提交状态。
// 2. 把真正的重型 ETL 投递给 worker pool。
// 3. worker 执行结束后，只要任务跑完，就顺序 Complete。
//
// 这样做的目的是：
// 1. 保持公共解码逻辑真正只有一份。
// 2. 避免 claim goroutine 被明细 ETL 的重型逻辑拖慢。
// 3. 保证 offset 提交仍然按 partition 内原顺序推进。
type reportConsumerPipeline struct {
	processor reportMessageProcessor
	pool      *util.DynamicWorkerPool
	commits   *orderedCommitManager
	gates     *reportCompletionGateTracker
}

type reportConsumerPipelineSnapshot struct {
	CommitterCount       int                         `json:"committerCount"`
	PendingCount         int                         `json:"pendingCount"`
	DoneCount            int                         `json:"doneCount"`
	LargestPendingGap    int64                       `json:"largestPendingGap"`
	OldestPendingOffset  int64                       `json:"oldestPendingOffset"`
	NewestCompletedOffset int64                      `json:"newestCompletedOffset"`
	Gate                 reportCompletionGateSnapshot `json:"gate"`
}

func newReportConsumerPipeline(processor reportMessageProcessor, pool *util.DynamicWorkerPool) *reportConsumerPipeline {
	return &reportConsumerPipeline{
		processor: processor,
		pool:      pool,
		commits:   newOrderedCommitManager(),
		gates:     &reportCompletionGateTracker{},
	}
}

// HandleDecoded 接收公共解码后的消息，再投递给 worker pool。
func (p *reportConsumerPipeline) HandleDecoded(decoded DecodedMessage) {
	committer := p.commits.Get(decoded.Input.Topic, decoded.Input.Partition)
	committer.Register(decoded.Input.Offset, decoded.Input.GenerationID, decoded.MarkFn)
	logTraceStage(decoded, "committer_register", "start")
	task := p.callback(decoded, committer)

	if isReportConsumerDirectExec() || p.pool == nil {
		p.executeDirect(decoded, task)
		return
	}

	p.executeWithPool(decoded, task, committer)
}

func (p *reportConsumerPipeline) executeDirect(decoded DecodedMessage, task func()) {
	logTraceStage(decoded, "direct_exec", "begin")
	task()
	logTraceStage(decoded, "direct_exec", "end")
}

func (p *reportConsumerPipeline) executeWithPool(decoded DecodedMessage, task func(), committer *partitionOrderedCommitter) {
	logTraceStage(decoded, "pool_submit", "begin")
	if err := p.pool.Submit(func() {
		logTraceStage(decoded, "pool_task", "begin")
		task()
		logTraceStage(decoded, "pool_task", "end")
	}); err != nil {
		logs.Logger.Error("submit report worker task failed",
			zap.Error(err),
			zap.String("topic", decoded.Input.Topic),
			zap.Int("partition", decoded.Input.Partition),
			zap.Int64("offset", decoded.Input.Offset),
			zap.String("event_name", decoded.KafkaData.EventName),
			zap.String("table_id", decoded.KafkaData.TableId),
			zap.Int("report_type", decoded.KafkaData.ReportType),
			zap.String("report_time", decoded.KafkaData.ReportTime),
		)
		logTraceStage(decoded, "pool_submit", "error",
			zap.Error(err),
		)
		committer.Complete(decoded.Input.Offset)
		return
	}
	logTraceStage(decoded, "pool_submit", "success")
}

func (p *reportConsumerPipeline) callback(decoded DecodedMessage, committer *partitionOrderedCommitter) func() {
	return func() {
		gate := newReportCompletionGate(decoded.Input.Offset, committer, p.gates)
		logTraceStage(decoded, "pipeline_callback", "begin")
		defer func() {
			if r := recover(); r != nil {
				logs.Logger.Error("report worker panic",
					zap.Any("panic", r),
					zap.String("topic", decoded.Input.Topic),
					zap.Int("partition", decoded.Input.Partition),
					zap.Int64("offset", decoded.Input.Offset),
					zap.String("event_name", decoded.KafkaData.EventName),
					zap.String("table_id", decoded.KafkaData.TableId),
					zap.Int("report_type", decoded.KafkaData.ReportType),
					zap.String("report_time", decoded.KafkaData.ReportTime),
					zap.ByteString("stack", debug.Stack()))
			}

			// 当前版本的顺序提交策略按“任务完成”提交，而不是按“任务成功”提交。
			// 这样 offset=10 处理失败时，不会永远阻塞 offset=11 及之后的提交。
			logTraceStage(decoded, "pipeline_complete", "begin")
			gate.NoAsyncTaskCompleteNow()
			logTraceStage(decoded, "pipeline_complete", "end")
		}()

		result := p.processor.ProcessDecoded(decoded, gate)
		logTraceStage(decoded, "pipeline_callback", "after_process_decoded",
			zap.Bool("handled", result.handled),
		)
		if !result.handled {
			logs.Logger.Error("report worker handled=false",
				zap.String("topic", decoded.Input.Topic),
				zap.Int("partition", decoded.Input.Partition),
				zap.Int64("offset", decoded.Input.Offset),
				zap.String("event_name", decoded.KafkaData.EventName),
				zap.String("table_id", decoded.KafkaData.TableId),
				zap.Int("report_type", decoded.KafkaData.ReportType))
		}
	}
}

func (p *reportConsumerPipeline) Close() error {
	return p.pool.Close()
}

func (p *reportConsumerPipeline) OnSessionCleanup(generationID int32) {
	p.logSnapshot("report consumer session cleanup snapshot", generationID)
}

func (p *reportConsumerPipeline) LogFinalSnapshot(label string) {
	p.logSnapshot(label, 0)
}

func (p *reportConsumerPipeline) Snapshot() reportConsumerPipelineSnapshot {
	snapshots := p.commits.Snapshots()
	snapshot := reportConsumerPipelineSnapshot{
		CommitterCount: len(snapshots),
		Gate:           p.gates.Snapshot(),
	}
	for _, item := range snapshots {
		snapshot.PendingCount += item.pendingCount
		snapshot.DoneCount += item.doneCount
		if item.largestPendingGap > snapshot.LargestPendingGap {
			snapshot.LargestPendingGap = item.largestPendingGap
		}
		if snapshot.OldestPendingOffset == 0 || (item.oldestPendingOffset > 0 && item.oldestPendingOffset < snapshot.OldestPendingOffset) {
			snapshot.OldestPendingOffset = item.oldestPendingOffset
		}
		if item.nextOffset > snapshot.NewestCompletedOffset {
			snapshot.NewestCompletedOffset = item.nextOffset
		}
	}
	return snapshot
}

func (p *reportConsumerPipeline) logSnapshot(label string, generationID int32) {
	if p == nil || p.commits == nil || !util.IsSinkerDiagnosticLogEnabled() {
		return
	}

	snapshots := p.commits.Snapshots()
	sort.Slice(snapshots, func(i, j int) bool {
		return snapshots[i].key < snapshots[j].key
	})

	totalPending := 0
	totalDone := 0
	for _, snapshot := range snapshots {
		totalPending += snapshot.pendingCount
		totalDone += snapshot.doneCount
		if snapshot.pendingCount == 0 {
			continue
		}

		logs.Logger.Info(
			"ordered commit pending offsets on session cleanup",
			zap.Int32("generation_id", generationID),
			zap.String("committer", snapshot.key),
			zap.Bool("initialized", snapshot.initialized),
			zap.Int64("next_offset", snapshot.nextOffset),
			zap.Int("pending_count", snapshot.pendingCount),
			zap.Int("done_count", snapshot.doneCount),
		)
	}

	logs.Logger.Info(
		label,
		zap.Int32("generation_id", generationID),
		zap.Int("committer_count", len(snapshots)),
		zap.Int("pending_count", totalPending),
		zap.Int("done_count", totalDone),
		zap.Int64("gate_in_flight_messages", p.gates.Snapshot().InFlightMessages),
		zap.Int64("gate_waiting_tasks", p.gates.Snapshot().WaitingTasks),
		zap.Int64("gate_completed_messages", p.gates.Snapshot().CompletedMessages),
	)
}
