package controller

import (
	"errors"
	"strings"
	"time"

	"github.com/1340691923/xwl_bi/platform-basic-libs/service/consumer_data"
	"github.com/1340691923/xwl_bi/platform-basic-libs/util"
	"github.com/tidwall/gjson"
	"github.com/valyala/fasthttp"
)

type reportRequestDecoder struct {
	now      func() time.Time
	clientIP func(ctx *fasthttp.RequestCtx) string
}

// newReportRequestDecoder 创建请求公共解码器。
//
// 为什么这里把时间函数和取 IP 函数做成依赖注入：
// 1. 业务逻辑里会用“当前时间”补默认值。
// 2. 单元测试需要稳定控制 now 和真实 IP，避免测试依赖系统时间和网络环境。
func newReportRequestDecoder(now func() time.Time, clientIP func(ctx *fasthttp.RequestCtx) string) *reportRequestDecoder {
	if now == nil {
		now = time.Now
	}
	if clientIP == nil {
		clientIP = util.CtxClientIP
	}
	return &reportRequestDecoder{
		now:      now,
		clientIP: clientIP,
	}
}

// Decode 负责把 HTTP 请求转换成后续 handler 可直接消费的结构化输入。
//
// 处理步骤：
// 1. 读取路由参数和 body。
// 2. 校验 eventName/appid/xwl_distinct_id 是否存在。
// 3. 补齐缺省 IP。
// 4. 把 xwl_part_date 归一化为统一时间格式；若缺省则使用当前服务端时间。
//
// 这里故意只做“公共准备动作”，不做 tableId 解析、debug 校验、Kafka 投递。
// 原因是这些动作属于业务编排层，不应该和 HTTP 解码耦合在一起。
//
// 示例：
//  1. xwl_part_date="2026-04-08" -> ReportTime="2026-04-08 00:00:00"
//     且 ReportTimeHasClock=false，表示原始值只有日期，没有时分秒。
//  2. xwl_ip 为空 -> ClientIP 使用真实请求 IP
//  3. 缺少 xwl_distinct_id -> 直接返回错误，不进入后续 handler
func (d *reportRequestDecoder) Decode(ctx *fasthttp.RequestCtx) (DecodedReportRequest, error) {
	request := DecodedReportRequest{
		Typ:       strings.TrimSpace(userValueString(ctx, "typ")),
		APPID:     strings.TrimSpace(userValueString(ctx, "appid")),
		AppKey:    strings.TrimSpace(userValueString(ctx, "appkey")),
		Debug:     strings.TrimSpace(userValueString(ctx, "debug")),
		EventName: strings.TrimSpace(userValueString(ctx, "eventName")),
		Body:      ctx.PostBody(),
	}

	if request.EventName == "" {
		return DecodedReportRequest{}, errors.New("事件名 不能为空")
	}
	if request.APPID == "" {
		return DecodedReportRequest{}, errors.New("appid 不能为空")
	}

	gjsonArr := gjson.GetManyBytes(request.Body, "xwl_distinct_id", "xwl_ip", "xwl_part_date")
	request.DistinctID = strings.TrimSpace(gjsonArr[0].String())
	if request.DistinctID == "" {
		return DecodedReportRequest{}, errors.New("xwl_distinct_id 不能为空")
	}

	request.ClientIP = strings.TrimSpace(gjsonArr[1].String())
	if request.ClientIP == "" {
		request.ClientIP = d.clientIP(ctx)
	}

	reportTime := strings.TrimSpace(gjsonArr[2].String())
	if reportTime == "" {
		reportTime = d.now().Format(util.TimeFormat)
	}
	request.ReportTimeHasClock = consumer_data.ReportTimeHasClock(reportTime)
	request.ReportTime = consumer_data.NormalizeReportTime(reportTime)

	return request, nil
}

func userValueString(ctx *fasthttp.RequestCtx, key string) string {
	value := ctx.UserValue(key)
	if value == nil {
		return ""
	}
	str, _ := value.(string)
	return str
}
