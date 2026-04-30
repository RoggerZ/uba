package runner

import (
	"math"
	"strconv"
	"time"

	"github.com/1340691923/xwl_bi/engine/logs"
	"github.com/1340691923/xwl_bi/model"
	"github.com/1340691923/xwl_bi/platform-basic-libs/service/consumer_data"
	parser "github.com/1340691923/xwl_bi/platform-basic-libs/sinker/parse"
	"github.com/1340691923/xwl_bi/platform-basic-libs/util"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"go.uber.org/zap"
)

const reportHandlerStageTimingSlowThresholdDefault = 5 * time.Second

// currentReportHandlerStageTimingSlowThreshold 返回当前生效的 report handler 慢日志阈值。
//
// 规则固定如下：
// 1. 默认阈值是 2s，用来避免常态成功消息把日志迅速打爆。
// 2. 如果运行中的 diagnostic session 显式覆盖了阈值，就临时使用会话里的值。
// 3. diagnostic disable 或会话到期后，会自动回到默认 2s。
//
// 示例：
//  1. 默认状态下：
//     total_cost=1.5s, result=success, handled=true -> 不因为“慢消息”打印
//  2. 诊断会话里把阈值调成 500ms：
//     total_cost=1.5s, result=success, handled=true -> 会因为“慢消息”打印
func currentReportHandlerStageTimingSlowThreshold() time.Duration {
	return effectiveReportHandlerStageTimingSlowThreshold(util.CurrentSinkerDiagnosticSession())
}

// effectiveReportHandlerStageTimingSlowThreshold 把“诊断会话里的覆盖值”和“代码默认值”合成为最终阈值。
//
// 这里故意把 session 版本单独拆出来，方便：
// 1. `report_handler` 在运行时直接取当前阈值
// 2. admin status 在返回 JSON 时复用完全一致的阈值解释规则
func effectiveReportHandlerStageTimingSlowThreshold(session util.SinkerDiagnosticSession) time.Duration {
	if session.ReportHandlerStageTimingThreshold > 0 {
		return session.ReportHandlerStageTimingThreshold
	}
	return reportHandlerStageTimingSlowThresholdDefault
}

func decodedLogFields(decoded DecodedMessage) []zap.Field {
	return []zap.Field{
		zap.String("topic", decoded.Input.Topic),
		zap.Int("partition", decoded.Input.Partition),
		zap.Int64("offset", decoded.Input.Offset),
		zap.String("event_name", decoded.KafkaData.EventName),
		zap.String("table_id", decoded.KafkaData.TableId),
		zap.Int("report_type", decoded.KafkaData.ReportType),
		zap.String("report_time", decoded.KafkaData.ReportTime),
	}
}

type reportProcessTiming struct {
	startedAt        time.Time
	extractClientCtx time.Duration
	geoEnrich        time.Duration
	parseMetric      time.Duration
	ensureColumns    time.Duration
	metaRecord       time.Duration
	statusAdd        time.Duration
	statusAddCount   int
	metricAdd        time.Duration
}

func newReportProcessTiming() *reportProcessTiming {
	return &reportProcessTiming{startedAt: time.Now()}
}

func (t *reportProcessTiming) observeStatusAdd(cost time.Duration) {
	t.statusAdd += cost
	t.statusAddCount++
}

// fields 返回单条消息在各处理阶段上的耗时字段。
//
// 这组字段是专门给线上排障准备的：
// 1. 一条消息结束时只打一条 Debug 日志，避免每个阶段都各自打日志把顺序打乱。
// 2. 无论是成功、业务丢弃还是可重试失败，都能看到已经走过的阶段耗时。
// 3. 如果某个阶段没走到，对应耗时就是 0，便于从日志直观看出“卡在了哪之前”。
//
// 示例：
//  1. extract_client_context=1ms, ensure_columns=85ms, metric_add=0ms
//     说明慢点更可能在 schema 校验而不是最终入批。
//  2. status_add_count=2
//     说明这条消息除了主流程状态写入，还在 schema 回调里额外写过一次失败状态。
func (t *reportProcessTiming) fields(totalCost time.Duration, result string, handled bool) []zap.Field {
	return []zap.Field{
		zap.String("stage_result", result),
		zap.Bool("handled", handled),
		zap.Duration("total_cost", totalCost),
		zap.Duration("extract_client_context", t.extractClientCtx),
		zap.Duration("geo_enrich", t.geoEnrich),
		zap.Duration("parse_metric", t.parseMetric),
		zap.Duration("ensure_columns", t.ensureColumns),
		zap.Duration("meta_record", t.metaRecord),
		zap.Duration("status_add", t.statusAdd),
		zap.Int("status_add_count", t.statusAddCount),
		zap.Duration("metric_add", t.metricAdd),
	}
}

// shouldLog 决定这条阶段耗时日志是否值得输出。
//
// 这里刻意不再“每条消息都打”：
// 1. 正常成功且很快的消息数量通常最大，逐条打印会迅速放大日志量。
// 2. 真正需要排查时，优先看异常消息、可重试消息和明显慢消息。
//
// 当前规则：
// 1. 业务结果不是 success -> 打印
// 2. handled=false，说明仍保留重试语义 -> 打印
// 3. totalCost >= 当前生效阈值，说明单条链路已经进入慢路径 -> 打印
func (t *reportProcessTiming) shouldLog(totalCost time.Duration, result string, handled bool) bool {
	return result != "success" || !handled || totalCost >= currentReportHandlerStageTimingSlowThreshold()
}

// logIfNeeded 把“是否打印阶段耗时日志”的判断和真正的日志输出收口到 timing 自身。
//
// 这样做的原因是：
// 1. `ProcessDecoded` 只负责业务链路本身，不再关心日志筛选细节。
// 2. 阶段耗时日志的阈值、字段拼装、输出策略都集中在 timing 结构上，更容易统一演进。
// 3. 运行态 diagnostic 现在通过“覆盖慢日志阈值”来细化这条日志，而不是简单粗暴地把所有消息都打出来。
func (t *reportProcessTiming) logIfNeeded(decoded DecodedMessage, result string, handled bool) {
	totalCost := time.Since(t.startedAt)
	if !t.shouldLog(totalCost, result, handled) {
		return
	}

	logs.Logger.Debug("report handler stage timing", append(decodedLogFields(decoded), t.fields(totalCost, result, handled)...)...)
}

// reportMessageHandler 负责完整的明细 ETL 链路：
// 关键上下文提取 -> 校验 -> 地域补充 -> 分区字段补充 -> 解析 metric ->
// 动态补列 -> 写入验收状态 -> 写入 CK 明细批次。
type reportMessageHandler struct {
	geo            geoPayloadEnricher
	schema         schemaSynchronizer
	metaRecorder   metaEventRecorder
	statusSink     acceptanceStatusSink
	metricSink     metricSink
	historyBlocker *historyReplayBlocker
	notifyWrite    func(tableName string)
}

// newReportMessageHandler 创建明细 ETL 处理器。
//
// 这里把可变点都通过接口注入：
// 1. 地理补充
// 2. 动态补列
// 3. 元数据记录
// 4. 验收状态写入
// 5. 明细批次写入
func newReportMessageHandler(
	geo geoPayloadEnricher,
	schema schemaSynchronizer,
	metaRecorder metaEventRecorder,
	statusSink acceptanceStatusSink,
	metricSink metricSink,
	historyBlocker *historyReplayBlocker,
	notifyWrite func(tableName string),
) *reportMessageHandler {
	return &reportMessageHandler{
		geo:            geo,
		schema:         schema,
		metaRecorder:   metaRecorder,
		statusSink:     statusSink,
		metricSink:     metricSink,
		historyBlocker: historyBlocker,
		notifyWrite:    notifyWrite,
	}
}

// ProcessDecoded 是 sinker 最重的一条消息处理链路。
//
// 典型处理顺序如下：
// 1. 读取已公共解码好的 KafkaData。
// 2. 提取 distinct_id、client_time、report_time 等关键上下文。
// 3. 通过前置校验后，补充地理信息和分区字段。
// 4. 解析 ReqData，完成动态补列。
// 5. 先记验收状态，再入 CK 明细批次。
// 6. 只有在消息已经成功交给批量器后，才允许 markFn。
//
// 示例：
// 1. 缺少 xwl_distinct_id -> 写 fail status -> 返回 true，允许提交 offset
// 2. schema 补列失败 -> 返回 false，阻塞该分区后续连续提交
//
// 返回值说明：
// 1. true 表示这条消息已经被系统“消费完成”，允许推进 offset。
// 2. false 表示这条消息仍应保留重试语义，不允许推进 offset。
// ProcessDecoded 负责单条 report 消息的完整 ETL 处理，同时把这条消息的异步落库任务登记到 gate。
//
// 这里和旧版本最大的区别是：
// 1. 函数返回不再直接等价于“这条消息可以 Complete(offset)”
// 2. status_add、metric_add 会把自己的异步 flush 回调挂到 gate 上
// 3. 等这条消息自己关联的任务全部结束后，gate 才会推进一次 `committer.Complete(offset)`
//
// 示例一：成功消息
// 1. 先登记 success status 入批任务
// 2. 再登记 metric 入批任务
// 3. ProcessDecoded 会先返回 `handled=true`
// 4. 但 offset 不会立刻 Complete
// 5. 要等 status 和 metric 两个 AfterFlush 都回调完，gate 才会真正推进提交
//
// 示例二：丢弃但记状态的消息
// 1. 缺少 distinct_id，只会登记一个 fail status 任务
// 2. 这一个任务 flush 完成后，就可以直接 Complete(offset)
//
// 示例三：没有异步任务的路径
// 1. 某条路径提前返回，但没有真正登记任何异步持久化任务
// 2. callback defer 会调用 `gate.NoAsyncTaskCompleteNow()`
// 3. 这时仍然按当前语义直接完成
func (h *reportMessageHandler) ProcessDecoded(decoded DecodedMessage, gate *reportCompletionGate) (result reportProcessResult) {
	timing := newReportProcessTiming()
	outcome := "success"
	logTraceStage(decoded, "process_decoded", "begin")
	defer func() {
		logTraceStage(decoded, "process_decoded", "end",
			zap.String("outcome", outcome),
			zap.Bool("handled", result.handled),
			zap.Duration("total_cost", time.Since(timing.startedAt)),
		)
		timing.logIfNeeded(decoded, outcome, result.handled)
	}()

	kafkaData := decoded.KafkaData

	tableId, err := strconv.Atoi(kafkaData.TableId)
	if err != nil {
		outcome = "invalid_table_id"
		logs.Logger.Error("strconv.Atoi(kafkaData.TableId) err", append(decodedLogFields(decoded), zap.Error(err))...)
		return reportProcessResult{handled: true}
	}

	if kafkaData.EventName == "" {
		outcome = "empty_event_name"
		return reportProcessResult{handled: true}
	}

	// recordStatusAdd 是 gate 在 report_handler 里的第一条典型使用路径。
	//
	// 它会把“写 acceptance_status”登记成当前消息自己的一个异步任务。
	//
	// 示例：
	// 1. 某条成功消息先写 success status
	// 2. 这里先 `gate.AddTask()`
	// 3. 再把 `data.AfterFlush = gate.TaskDone`
	// 4. 等 status batch 真正 flush 完成后，这个回调才会让 gate 少一个任务
	recordStatusAdd := func(data *consumer_data.ReportAcceptStatusData, logLabel string) bool {
		logTraceStage(decoded, "status_add", "begin",
			zap.String("log_label", logLabel),
			zap.Int("status", data.Status),
			zap.String("part_date", data.PartDate),
		)
		start := time.Now()
		accepted := h.addStatus(data, logLabel, decoded, gate)
		cost := time.Since(start)
		timing.observeStatusAdd(cost)
		logTraceStage(decoded, "status_add", "end",
			zap.String("log_label", logLabel),
			zap.Duration("cost", cost),
		)
		return accepted
	}

	logTraceStage(decoded, "extract_client_context", "begin")
	stageBegin := time.Now()
	xwlDistinctId, xwlClientTime, clientTime, reportTimeHasClock, err := h.extractClientContext(&kafkaData)
	timing.extractClientCtx = time.Since(stageBegin)
	logTraceStage(decoded, "extract_client_context", "end",
		zap.Duration("cost", timing.extractClientCtx),
		zap.Error(err),
	)
	if err != nil {
		outcome = "invalid_client_time"
		logs.Logger.Error(
			"invalid xwl_client_time",
			append(decodedLogFields(decoded), zap.Error(err), zap.ByteString("reqData", kafkaData.ReqData))...,
		)
		return reportProcessResult{handled: true}
	}

	if xwlDistinctId == "" {
		outcome = "missing_distinct_id"
		recordStatusAdd(&consumer_data.ReportAcceptStatusData{
			IngestTime:     resolveIngestTime(decoded),
			PartDate:       kafkaData.ReportTime,
			TableId:        tableId,
			ReportType:     h.invalidReportType(kafkaData),
			DataName:       kafkaData.EventName,
			ErrorReason:    "xwl_distinct_id 不能为空",
			ErrorHandling:  "丢弃数据",
			ReportData:     util.Bytes2str(kafkaData.ReqData),
			XwlKafkaOffset: kafkaData.Offset,
			Status:         consumer_data.FailStatus,
		}, "reportAcceptStatus missing distinct id")
		logs.Logger.Error(
			"xwl_distinct_id 为空",
			append(decodedLogFields(decoded), zap.String("reqData", util.Bytes2str(kafkaData.ReqData)))...,
		)
		return reportProcessResult{handled: true}
	}

	if kafkaData.Ip != "" {
		logTraceStage(decoded, "geo_enrich", "begin",
			zap.String("ip", kafkaData.Ip),
		)
		stageBegin = time.Now()
		kafkaData.ReqData, err = h.geo.Enrich(kafkaData.ReqData, kafkaData.Ip)
		timing.geoEnrich = time.Since(stageBegin)
		logTraceStage(decoded, "geo_enrich", "end",
			zap.String("ip", kafkaData.Ip),
			zap.Duration("cost", timing.geoEnrich),
			zap.Error(err),
		)
		if err != nil {
			logs.Logger.Error("geo lookup err", append(decodedLogFields(decoded), zap.Error(err), zap.String("ip", kafkaData.Ip))...)
		}
	} else {
		logTraceStage(decoded, "geo_enrich", "skip",
			zap.String("reason", "empty_ip"),
		)
	}

	serverT := util.Str2Time(kafkaData.ReportTime, util.TimeFormat)
	if reportTimeHasClock && math.Abs(serverT.Sub(clientTime).Minutes()) > 10 {
		outcome = "client_server_time_drift"
		recordStatusAdd(&consumer_data.ReportAcceptStatusData{
			IngestTime:     resolveIngestTime(decoded),
			PartDate:       kafkaData.ReportTime,
			TableId:        tableId,
			ReportType:     kafkaData.GetReportTypeErr(),
			DataName:       kafkaData.EventName,
			ErrorReason:    "客户端上报时间误差大于十分钟",
			ErrorHandling:  "丢弃数据",
			ReportData:     util.Bytes2str(kafkaData.ReqData),
			XwlKafkaOffset: kafkaData.Offset,
			Status:         consumer_data.FailStatus,
		}, "reportAcceptStatus client/server time drift")
		logs.Logger.Error(
			"客户端上报时间误差大于十分钟", append(decodedLogFields(decoded),
				zap.String("client_time", xwlClientTime),
				zap.String("server_time", kafkaData.ReportTime),
			)...,
		)
		return reportProcessResult{handled: true}
	}

	xwlClientDate := time.Date(clientTime.Year(), clientTime.Month(), clientTime.Day(), 0, 0, 0, 0, clientTime.Location()).Format(util.TimeFormat)
	kafkaData.ReqData = h.enrichRequestPayload(kafkaData.ReqData, kafkaData.EventName, xwlClientTime, xwlClientDate, kafkaData.ConsumptionTime)

	pp := parser.FastjsonParser{}
	logTraceStage(decoded, "parse_metric", "begin")
	stageBegin = time.Now()
	metric, err := pp.Parse(kafkaData.ReqData)
	timing.parseMetric = time.Since(stageBegin)
	logTraceStage(decoded, "parse_metric", "end",
		zap.Duration("cost", timing.parseMetric),
		zap.Error(err),
	)
	if err != nil {
		outcome = "parse_metric_failed"
		logs.Logger.Error("ParseKafkaData err", append(decodedLogFields(decoded), zap.Error(err), zap.ByteString("reqData", kafkaData.ReqData))...)
		return reportProcessResult{handled: true}
	}

	tableName := kafkaData.GetTableName()
	logTraceStage(decoded, "ensure_columns", "begin",
		zap.String("target_table", tableName),
	)
	stageBegin = time.Now()
	if err := h.schema.EnsureColumns(
		kafkaData,
		tableName,
		metric,
		func(data consumer_data.ReportAcceptStatusData) {
			recordStatusAdd(&data, "reportAcceptStatus schema validation")
		},
	); err != nil {
		timing.ensureColumns = time.Since(stageBegin)
		logTraceStage(decoded, "ensure_columns", "end",
			zap.String("target_table", tableName),
			zap.Duration("cost", timing.ensureColumns),
			zap.Error(err),
		)
		outcome = "ensure_columns_failed"
		logs.Logger.Error(
			"addTableColumn err",
			append(decodedLogFields(decoded), zap.String("target_table", tableName), zap.Error(err))...,
		)
		return reportProcessResult{handled: false}
	}
	timing.ensureColumns = time.Since(stageBegin)
	logTraceStage(decoded, "ensure_columns", "end",
		zap.String("target_table", tableName),
		zap.Duration("cost", timing.ensureColumns),
	)

	logTraceStage(decoded, "meta_record", "begin")
	stageBegin = time.Now()
	if err := h.metaRecorder.Record(kafkaData); err != nil {
		timing.metaRecord = time.Since(stageBegin)
		logTraceStage(decoded, "meta_record", "end",
			zap.Duration("cost", timing.metaRecord),
			zap.Error(err),
		)
		outcome = "meta_record_failed_non_blocking"
		logs.Logger.Error("addMetaEvent err", append(decodedLogFields(decoded), zap.Error(err))...)
	} else {
		timing.metaRecord = time.Since(stageBegin)
		logTraceStage(decoded, "meta_record", "end",
			zap.Duration("cost", timing.metaRecord),
		)
	}

	recordStatusAdd(&consumer_data.ReportAcceptStatusData{
		IngestTime:     resolveIngestTime(decoded),
		PartDate:       kafkaData.ReportTime,
		TableId:        tableId,
		DataName:       kafkaData.EventName,
		XwlKafkaOffset: kafkaData.Offset,
		Status:         consumer_data.SuccessStatus,
	}, "reportAcceptStatus Add SuccessStatus err")

	logTraceStage(decoded, "metric_add", "begin",
		zap.String("target_table", tableName),
	)
	stageBegin = time.Now()
	metricData := consumer_data.FastjsonMetricData{
		TableName:      tableName,
		FastjsonMetric: metric,
	}
	if gate != nil {
		// 这里是 gate 的第二条关键使用路径。
		//
		// 成功消息通常至少会关联两个异步任务：
		// 1. success status 入批
		// 2. metric 明细入批
		//
		// 因此到 metric_add 这里还要再给这条消息追加一个任务名额。
		//
		// 示例：
		// 1. status_add 已经让 remainingTasks=1
		// 2. 这里再 `AddTask()` 后变成 remainingTasks=2
		// 3. 只有 status 和 metric 两个 flush 回调都跑完，gate 才会触发 Complete(offset)
		gate.AddTask()
		metricData.AfterFlush = gate.TaskDone
	}
	if err := h.metricSink.Add(metricData); err != nil {
		if gate != nil {
			gate.TaskDone()
		}
		timing.metricAdd = time.Since(stageBegin)
		logTraceStage(decoded, "metric_add", "end",
			zap.String("target_table", tableName),
			zap.Duration("cost", timing.metricAdd),
			zap.Error(err),
		)
		outcome = "metric_add_failed"
		logs.Logger.Error(
			"reportData2CK err",
			append(decodedLogFields(decoded), zap.String("target_table", tableName), zap.Error(err))...,
		)
		return reportProcessResult{handled: false}
	}
	timing.metricAdd = time.Since(stageBegin)
	logTraceStage(decoded, "metric_add", "end",
		zap.String("target_table", tableName),
		zap.Duration("cost", timing.metricAdd),
	)

	outcome = "success"
	return reportProcessResult{handled: true}
}

// extractClientContext 把主链路里最容易反复出错的时间/身份字段提取集中到一处。
//
// 输出值含义：
// 1. distinctID：访客标识
// 2. xwlClientTime：归一化后的客户端时间字符串
// 3. clientTime：对应的 time.Time
// 4. reportTimeHasClock：report_time 是否自带时分秒，用于决定是否做分钟级误差校验
func (h *reportMessageHandler) extractClientContext(kafkaData *model.KafkaData) (distinctID string, xwlClientTime string, clientTime time.Time, reportTimeHasClock bool, err error) {
	gjsonArr := gjson.GetManyBytes(kafkaData.ReqData, "xwl_distinct_id", "xwl_client_time", "xwl_client_date")
	distinctID = gjsonArr[0].String()

	xwlClientTime, clientTime, err = consumer_data.NormalizeClientTime(gjsonArr[1].String())
	if err != nil {
		return "", "", time.Time{}, false, err
	}

	// 优先使用 report_server 在生产 KafkaData 时保留下来的原始时钟语义。
	// 这样像 "2026-04-08" 这种只带日期的上报，在被归一化成
	// "2026-04-08 00:00:00" 后，仍然不会被误判成“原始就带时分秒”。
	if kafkaData.ReportTimeHasClock != nil {
		kafkaData.ReportTime = consumer_data.NormalizeReportTime(kafkaData.ReportTime)
		return distinctID, xwlClientTime, clientTime, *kafkaData.ReportTimeHasClock, nil
	}

	kafkaData.ReportTime, reportTimeHasClock = consumer_data.NormalizeReportTimeForValidation(kafkaData.ReportTime, xwlClientTime)
	return distinctID, xwlClientTime, clientTime, reportTimeHasClock, nil
}

// enrichRequestPayload 统一补齐写 CK 前需要的标准字段。
//
// 示例：
// 1. xwl_part_event 表示分区事件名
// 2. xwl_part_date 使用客户端事件时间
// 3. xwl_server_time 使用服务端实际消费时间
func (h *reportMessageHandler) enrichRequestPayload(reqData []byte, eventName, xwlClientTime, xwlClientDate, consumptionTime string) []byte {
	reqData, _ = sjson.SetBytes(reqData, "xwl_part_event", eventName)
	reqData, _ = sjson.SetBytes(reqData, "xwl_part_date", xwlClientTime)
	reqData, _ = sjson.SetBytes(reqData, "xwl_client_date", xwlClientDate)
	reqData, _ = sjson.SetBytes(reqData, "xwl_server_time", consumptionTime)
	return reqData
}

// invalidReportType 把 report_type 转成更稳定的错误类别描述，
// 避免主流程到处 scattered switch。
func (h *reportMessageHandler) invalidReportType(kafkaData model.KafkaData) string {
	switch kafkaData.ReportType {
	case model.UserReportType:
		return "用户属性类型不合法"
	case model.EventReportType:
		return "事件属性类型不合法"
	default:
		return kafkaData.GetReportTypeErr()
	}
}

// addStatus 统一封装验收状态写入和失败日志，避免主流程里重复样板错误处理。
// addStatus 统一封装 status 入批、gate 登记和失败日志。
//
// 它的职责不只是把 status 数据交给 batcher，
// 还要把这次 status 入批登记成“当前消息自己的一个异步任务”。
//
// 示例：
// 1. 某条消息进入 addStatus
// 2. 先 `gate.AddTask()`，remainingTasks+1
// 3. 再把 `data.AfterFlush = gate.TaskDone`
// 4. status batch flush 成功后，AfterFlush 被调用，remainingTasks-1
// 5. 如果这是这条消息唯一的异步任务，就会立即触发 `Complete(offset)`
func (h *reportMessageHandler) addStatus(data *consumer_data.ReportAcceptStatusData, logLabel string, decoded DecodedMessage, gate *reportCompletionGate) bool {
	if data.IngestTime == "" {
		data.IngestTime = resolveIngestTime(decoded)
	}
	if h.historyBlocker != nil && h.historyBlocker.ShouldSkip(consumer_data.TableNameAcceptanceStatus, data.PartDate) {
		return false
	}
	if gate != nil {
		gate.AddTask()
		data.AfterFlush = gate.TaskDone
	}
	if h.notifyWrite != nil {
		h.notifyWrite(consumer_data.TableNameAcceptanceStatus)
	}
	if err := h.statusSink.Add(data); err != nil {
		if gate != nil {
			gate.TaskDone()
		}
		if consumer_data.IsDeferredFlushError(err) {
			return false
		}
		logs.Logger.Error(logLabel, append(decodedLogFields(decoded), zap.Error(err))...)
		return false
	}
	return true
}
