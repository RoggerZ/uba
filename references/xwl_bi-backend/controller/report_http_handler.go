package controller

import (
	"strings"
	"time"

	"github.com/1340691923/xwl_bi/platform-basic-libs/response"
	"github.com/1340691923/xwl_bi/platform-basic-libs/service/report"
	sinkerModel "github.com/1340691923/xwl_bi/platform-basic-libs/sinker/model"
	parser "github.com/1340691923/xwl_bi/platform-basic-libs/sinker/parse"
	"github.com/1340691923/xwl_bi/platform-basic-libs/util"
	"github.com/valyala/fasthttp"
)

type ReportHandlerDependencies struct {
	Now            func() time.Time
	ClientIP       func(ctx *fasthttp.RequestCtx) string
	ResolveTableID resolveTableIDFunc
	BuildPayload   buildReportPayloadFunc
	IsDebugDevice  isDebugDeviceFunc
	SendDebugData  sendDebugDataFunc
	SendReportData sendReportDataFunc
	ParseMetric    parseMetricFunc
	LoadDims       loadDimsFunc
}

type reportHTTPHandler struct {
	decoder   *reportRequestDecoder
	ingress   *reportIngressHandler
	responder response.Response
}

// newReportHTTPHandler 创建最外层 HTTP 适配器。
//
// 注意它不是业务 handler，而是“HTTP 协议层 -> 内部 handler” 的桥接层。
func newReportHTTPHandler(decoder *reportRequestDecoder, ingress *reportIngressHandler) *reportHTTPHandler {
	return &reportHTTPHandler{
		decoder: decoder,
		ingress: ingress,
	}
}

// NewReportHandler 根据依赖装配出 report_server 对外暴露的 HTTP 处理器。
//
// 这里把装配逻辑集中在一个构造函数，是为了让 main.go 只负责传入依赖，
// 而不必感知内部到底拆成了几层 handler。
//
// 依赖分层约定：
// 1. `Now/ClientIP` 属于请求公共解码层。
// 2. `ResolveTableID/BuildPayload/SendReportData` 属于正式上报编排层。
// 3. `IsDebugDevice/ParseMetric/LoadDims/SendDebugData` 属于 debug 检查层。
//
// 示例：
// 1. 生产环境由 runtime 注入真实 Redis/Kafka/CK 依赖。
// 2. 单元测试可以只替换其中一个函数，例如把 SendReportData 换成 spy。
func NewReportHandler(deps ReportHandlerDependencies) fasthttp.RequestHandler {
	if deps.Now == nil {
		deps.Now = time.Now
	}
	if deps.ClientIP == nil {
		deps.ClientIP = util.CtxClientIP
	}
	if deps.BuildPayload == nil {
		deps.BuildPayload = report.DefaultPayloadBuilderRegistry().Build
	}
	if deps.ParseMetric == nil {
		deps.ParseMetric = func(body []byte) (*parser.FastjsonMetric, error) {
			return (&parser.FastjsonParser{}).Parse(body)
		}
	}
	if deps.LoadDims == nil {
		deps.LoadDims = func(tableName string) ([]*sinkerModel.ColumnWithType, error) {
			return nil, nil
		}
	}
	if deps.IsDebugDevice == nil {
		checker := report.NewDebugMembershipChecker(nil)
		deps.IsDebugDevice = checker.IsDebugDevice
	}
	if deps.ResolveTableID == nil {
		deps.ResolveTableID = report.DefaultTableIDResolver().Resolve
	}
	if deps.SendDebugData == nil {
		producer := report.NewDefaultDebugDataProducer()
		deps.SendDebugData = producer.Send
	}
	if deps.SendReportData == nil {
		producer := report.NewDefaultKafkaDataProducer()
		deps.SendReportData = producer.Send
	}

	decoder := newReportRequestDecoder(deps.Now, deps.ClientIP)
	debugHandler := newDebugInspectionHandler(
		deps.ParseMetric,
		deps.LoadDims,
		deps.IsDebugDevice,
		deps.SendDebugData,
	)
	ingress := newReportIngressHandler(
		deps.ResolveTableID,
		deps.BuildPayload,
		debugHandler.Handle,
		deps.SendReportData,
	)

	return newReportHTTPHandler(decoder, ingress).Handle
}

// Handle 只负责 HTTP 层适配，不再承载完整业务流程。
//
// 旧的 ReportAction 把校验、debug 检查、Kafka 投递全部堆在一个函数里。
// 现在这里固定只做三件事：
// 1. 处理 OPTIONS。
// 2. 把请求交给 decoder 和 ingress handler。
// 3. 把结果按既有 JSON 结构写回客户端。
//
// 关键约束：
// 1. 这里不能再塞进 tableId 解析、debug 校验、Kafka 投递等业务细节。
// 2. 否则入口层又会退回旧的“胖控制器”状态。
func (h *reportHTTPHandler) Handle(ctx *fasthttp.RequestCtx) {
	if strings.ToUpper(util.Bytes2str(ctx.Method())) == "OPTIONS" {
		return
	}

	decoded, err := h.decoder.Decode(ctx)
	if err != nil {
		_ = h.responder.FastError(ctx, err)
		return
	}

	result, err := h.ingress.Handle(decoded)
	if err != nil {
		_ = h.responder.FastError(ctx, err)
		return
	}

	_ = h.responder.Output(ctx, map[string]interface{}{
		"code": 0,
		"msg":  result.Message,
	})
}
