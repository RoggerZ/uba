package main

import (
	"encoding/json"
	"testing"

	"github.com/1340691923/xwl_bi/model"
	"github.com/valyala/fasthttp"
)

func TestNewReportRouterRegistersSingleSyncRoute(t *testing.T) {
	router := newReportRouter(func(ctx *fasthttp.RequestCtx) {
		ctx.SetBodyString(`{"ok":true}`)
	})

	t.Run("默认 6 段路由按 signed 处理", func(t *testing.T) {
		ctx := newRuntimeRouteTestContext(`/ingress/reportEvent/1001/demo/pay_success/1`, `{}`)
		router.Handler(ctx)
		assertRuntimeResponseCode(t, ctx.Response.Body(), 400)
	})

	t.Run("X-SkipSigned 命中配置时直传放行", func(t *testing.T) {
		oldValue := model.GlobConfig.Report.SkipSigned
		model.GlobConfig.Report.SkipSigned = "report-tools-bypass"
		defer func() {
			model.GlobConfig.Report.SkipSigned = oldValue
		}()

		ctx := newRuntimeRouteTestContext(`/ingress/reportEvent/1001/demo/pay_success/1`, `{}`)
		ctx.Request.Header.Set("X-SkipSigned", "report-tools-bypass")
		router.Handler(ctx)
		if string(ctx.Response.Body()) != `{"ok":true}` {
			t.Fatalf("response = %s, want %s", string(ctx.Response.Body()), `{"ok":true}`)
		}
	})
}

func newRuntimeRouteTestContext(path, body string) *fasthttp.RequestCtx {
	var req fasthttp.Request
	req.Header.SetMethod(fasthttp.MethodPost)
	req.Header.SetContentType("application/json")
	req.SetBodyString(body)

	var ctx fasthttp.RequestCtx
	ctx.Init(&req, nil, nil)
	ctx.Request.SetRequestURI(path)
	return &ctx
}

func assertRuntimeResponseCode(t *testing.T, body []byte, want int) {
	t.Helper()
	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		t.Fatalf("failed to decode response body %s: %v", string(body), err)
	}
	if got := int(response["code"].(float64)); got != want {
		t.Fatalf("code = %d, want %d", got, want)
	}
}
