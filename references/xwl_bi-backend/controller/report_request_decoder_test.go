package controller

import (
	"testing"
	"time"

	"github.com/1340691923/xwl_bi/platform-basic-libs/util"
	"github.com/valyala/fasthttp"
)

func TestReportRequestDecoderDecode(t *testing.T) {
	decoder := newReportRequestDecoder(
		func() time.Time {
			return time.Date(2026, 4, 10, 12, 30, 0, 0, time.Local)
		},
		func(ctx *fasthttp.RequestCtx) string {
			return "10.10.10.10"
		},
	)

	t.Run("xwl_part_date 会归一化", func(t *testing.T) {
		ctx := newReportRequestContext(`{"xwl_distinct_id":"abc","xwl_part_date":"2026-04-08"}`)
		ctx.SetUserValue("typ", "reportEvent")
		ctx.SetUserValue("appid", "1001")
		ctx.SetUserValue("appkey", "demo")
		ctx.SetUserValue("debug", "1")
		ctx.SetUserValue("eventName", "pay_success")

		decoded, err := decoder.Decode(ctx)
		if err != nil {
			t.Fatalf("Decode returned error: %v", err)
		}
		if decoded.ReportTime != "2026-04-08 00:00:00" {
			t.Fatalf("ReportTime = %q, want %q", decoded.ReportTime, "2026-04-08 00:00:00")
		}
		if decoded.ReportTimeHasClock {
			t.Fatal("date-only xwl_part_date should not be treated as having clock fields")
		}
	})

	t.Run("xwl_ip 缺省时回退到真实 IP", func(t *testing.T) {
		ctx := newReportRequestContext(`{"xwl_distinct_id":"abc"}`)
		ctx.SetUserValue("typ", "reportEvent")
		ctx.SetUserValue("appid", "1001")
		ctx.SetUserValue("appkey", "demo")
		ctx.SetUserValue("debug", "1")
		ctx.SetUserValue("eventName", "pay_success")

		decoded, err := decoder.Decode(ctx)
		if err != nil {
			t.Fatalf("Decode returned error: %v", err)
		}
		if decoded.ClientIP != "10.10.10.10" {
			t.Fatalf("ClientIP = %q, want %q", decoded.ClientIP, "10.10.10.10")
		}
		if decoded.ReportTime != "2026-04-10 12:30:00" {
			t.Fatalf("ReportTime = %q, want %q", decoded.ReportTime, "2026-04-10 12:30:00")
		}
		if !decoded.ReportTimeHasClock {
			t.Fatal("server fallback time should be treated as having clock fields")
		}
	})

	t.Run("缺少事件名返回错误", func(t *testing.T) {
		ctx := newReportRequestContext(`{"xwl_distinct_id":"abc"}`)
		ctx.SetUserValue("typ", "reportEvent")
		ctx.SetUserValue("appid", "1001")
		ctx.SetUserValue("appkey", "demo")
		ctx.SetUserValue("debug", "1")
		ctx.SetUserValue("eventName", "")

		_, err := decoder.Decode(ctx)
		if err == nil || err.Error() != "事件名 不能为空" {
			t.Fatalf("Decode error = %v, want %q", err, "事件名 不能为空")
		}
	})

	t.Run("缺少 appid 返回错误", func(t *testing.T) {
		ctx := newReportRequestContext(`{"xwl_distinct_id":"abc"}`)
		ctx.SetUserValue("typ", "reportEvent")
		ctx.SetUserValue("appid", "")
		ctx.SetUserValue("appkey", "demo")
		ctx.SetUserValue("debug", "1")
		ctx.SetUserValue("eventName", "pay_success")

		_, err := decoder.Decode(ctx)
		if err == nil || err.Error() != "appid 不能为空" {
			t.Fatalf("Decode error = %v, want %q", err, "appid 不能为空")
		}
	})

	t.Run("缺少 distinct_id 返回错误", func(t *testing.T) {
		ctx := newReportRequestContext(`{"xwl_part_date":"2026-04-08"}`)
		ctx.SetUserValue("typ", "reportEvent")
		ctx.SetUserValue("appid", "1001")
		ctx.SetUserValue("appkey", "demo")
		ctx.SetUserValue("debug", "1")
		ctx.SetUserValue("eventName", "pay_success")

		_, err := decoder.Decode(ctx)
		if err == nil || err.Error() != "xwl_distinct_id 不能为空" {
			t.Fatalf("Decode error = %v, want %q", err, "xwl_distinct_id 不能为空")
		}
	})
}

func newReportRequestContext(body string) *fasthttp.RequestCtx {
	var req fasthttp.Request
	req.Header.SetMethod(fasthttp.MethodPost)
	req.Header.SetContentType("application/json")
	req.SetBodyString(body)

	var ctx fasthttp.RequestCtx
	ctx.Init(&req, nil, nil)
	ctx.SetConnectionClose()
	ctx.Request.SetRequestURI("/ingress/reportEvent/1001/demo/pay_success/1")
	ctx.ConnRequestNum()
	ctx.SetUserValue("remoteIP", util.Bytes2str(ctx.RemoteIP()))
	return &ctx
}
