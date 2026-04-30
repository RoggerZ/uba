package controller

import (
	"github.com/1340691923/xwl_bi/model"
	"github.com/1340691923/xwl_bi/platform-basic-libs/service/report"
)

type resolveTableIDFunc func(appid, appkey string) (string, error)
type buildReportPayloadFunc func(typ string, input report.PayloadBuildInput) (model.KafkaData, error)
type sendReportDataFunc func(data model.KafkaData) error

type reportIngressHandler struct {
	resolveTableID resolveTableIDFunc
	buildPayload   buildReportPayloadFunc
	inspectDebug   func(request DecodedReportRequest, kafkaData model.KafkaData) (debugInspectionDecision, error)
	sendReportData sendReportDataFunc
}

// newReportIngressHandler 创建正式上报主流程处理器。
//
// 这个处理器是 report 链路的“业务编排中枢”：
// 1. 它本身不关心 HTTP，也不直接关心 Redis/Kafka 的具体实现。
// 2. 它只负责把“解析 tableId -> 构造 payload -> debug 检查 -> 正式发送”这四个动作串起来。
func newReportIngressHandler(
	resolveTableID resolveTableIDFunc,
	buildPayload buildReportPayloadFunc,
	inspectDebug func(request DecodedReportRequest, kafkaData model.KafkaData) (debugInspectionDecision, error),
	sendReportData sendReportDataFunc,
) *reportIngressHandler {
	return &reportIngressHandler{
		resolveTableID: resolveTableID,
		buildPayload:   buildPayload,
		inspectDebug:   inspectDebug,
		sendReportData: sendReportData,
	}
}

// Handle 负责完成一次正式上报请求的主流程编排。
//
// 执行顺序：
// 1. 先根据 appid/appkey 解析 tableId。
// 2. 再按 typ 构造 KafkaData。
// 3. 然后调用 debug 检查处理器决定是否提前结束。
// 4. 只有未被 debug 分支拦截时，才进入正式 Kafka。
//
// 示例：
// 1. 非 debug 设备 -> Resolve -> Build -> SendReportData -> 返回“上报成功”
// 2. debug=2 且校验通过 -> Resolve -> Build -> inspectDebug 返回 Stop=true -> 不发送正式 Kafka
func (h *reportIngressHandler) Handle(request DecodedReportRequest) (reportIngressResult, error) {
	// 第一步解析 tableId。
	//
	// 只有 tableId 成功解析，后续才知道：
	// 1. 最终写哪个业务表
	// 2. debug 设备集合该查哪一个 Redis key
	tableID, err := h.resolveTableID(request.APPID, request.AppKey)
	if err != nil {
		return reportIngressResult{}, err
	}

	// 第二步构造 KafkaData。
	//
	// 这一步把 HTTP 请求上下文转换为统一消息结构，
	// 后续 debug 检查和正式发送都只认 KafkaData，不再认 HTTP 参数。
	kafkaData, err := h.buildPayload(request.Typ, report.PayloadBuildInput{
		APPID:              request.APPID,
		TableID:            tableID,
		Debug:              request.Debug,
		ReportTime:         request.ReportTime,
		ReportTimeHasClock: request.ReportTimeHasClock,
		EventName:          request.EventName,
		IP:                 request.ClientIP,
		Body:               request.Body,
	})
	if err != nil {
		return reportIngressResult{}, err
	}

	// 第三步进入 debug 检查。
	//
	// 这里的返回值不是简单 bool，而是 decision：
	// 1. Stop=false 表示继续正式 Kafka。
	// 2. Stop=true 且 Message 有值，表示正常提前结束，例如 debug=2。
	// 3. error != nil 表示业务失败，直接向客户端返回错误。
	decision, err := h.inspectDebug(request, kafkaData)
	if err != nil {
		return reportIngressResult{}, err
	}
	if decision.Stop {
		return reportIngressResult{Message: decision.Message}, nil
	}

	// 第四步才是真正的正式 Kafka 发送。
	//
	// 到这里说明：
	// 1. tableId 已经解析成功。
	// 2. payload 已经构造完成。
	// 3. debug 分支没有要求提前结束。
	if err := h.sendReportData(kafkaData); err != nil {
		return reportIngressResult{}, err
	}

	return reportIngressResult{Message: "上报成功"}, nil
}
