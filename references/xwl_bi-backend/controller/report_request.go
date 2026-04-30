package controller

// DecodedReportRequest 表示已经完成公共解码、公共字段补齐的一次上报请求。
//
// 设计原因：
// 1. 把 fasthttp 上下文里的“路由参数 + body + 补齐字段”集中成稳定对象。
// 2. 后续 handler 只处理业务差异，不再每层重复从 ctx 里取值。
// 3. 当请求经过 decoder 之后，后面的 ingress/debug handler 都只依赖这个结构，不再依赖 HTTP 细节。
//
// 示例：
// 1. xwl_ip 为空时，ClientIP 会被补成请求真实 IP
// 2. xwl_part_date 为 "2026-04-08" 时，ReportTime 会被整理成 "2026-04-08 00:00:00"
// 3. DistinctID 会从原始 body 中抽出，后续 debug 判定、日志、debug topic 记账都会复用它
type DecodedReportRequest struct {
	Typ                string
	APPID              string
	AppKey             string
	Debug              string
	EventName          string
	Body               []byte
	DistinctID         string
	ClientIP           string
	ReportTime         string
	ReportTimeHasClock bool
}

type reportIngressResult struct {
	Message string
}

// debugInspectionDecision 表示 debug 检查阶段对主流程的控制结果。
//
// 字段含义：
// 1. Stop=true 表示主流程到此结束，不再进入正式 Kafka。
// 2. Message 只有在“成功但不入库”这类正常提前结束场景下才有值。
//
// 示例：
// 1. debug=2 且校验通过 -> Stop=true, Message="上报成功（数据不入库）"
// 2. 普通非 debug 设备 -> Stop=false
type debugInspectionDecision struct {
	Stop    bool
	Message string
}
