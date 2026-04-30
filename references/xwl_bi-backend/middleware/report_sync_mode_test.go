package middleware

import (
	"encoding/json"
	"net/url"
	"testing"

	"github.com/1340691923/xwl_bi/model"
	"github.com/tidwall/gjson"
	"github.com/valyala/fasthttp"
)

func TestReportSyncModeDefaultSignedMergesHeadersAndVerifiesSignature(t *testing.T) {
	var (
		called   bool
		gotBody  []byte
		reporter = ReportSyncMode(func(ctx *fasthttp.RequestCtx) {
			called = true
			gotBody = append([]byte(nil), ctx.PostBody()...)
			ctx.SetBodyString("ok")
		})
	)

	oldPrefix := model.GlobConfig.Report.SignaturePathPrefix
	model.GlobConfig.Report.SignaturePathPrefix = "/api/v3/reporter"
	defer func() {
		model.GlobConfig.Report.SignaturePathPrefix = oldPrefix
	}()

	ctx := newReportSyncTestContext(`/ingress/reportEvent/616077185416953866/fe6ba3206fcf04fe71bedf1ea112cd81/AppLaunch/0`, `{}`)
	applySampleSignedHeaders(ctx)
	setLiveSignature(t, ctx, `/api/v3/reporter/ingress/reportEvent/616077185416953866/fe6ba3206fcf04fe71bedf1ea112cd81/AppLaunch/0`)

	reporter(ctx)

	if !called {
		t.Fatal("signed request should reach next handler")
	}
	assertJSONField(t, gotBody, "xwl_distinct_id", "10088711")
	assertJSONField(t, gotBody, "user_id", "10088711")
	assertJSONField(t, gotBody, "device_id", "235520ce1ba2158b8a151f89779494445")
	assertJSONField(t, gotBody, "referrer", "GooglePlay")
	assertJSONField(t, gotBody, "app_version", "20304")
	assertJSONField(t, gotBody, "platform", "android")
	assertJSONField(t, gotBody, "lang", "zh")
	assertJSONField(t, gotBody, "brand", "samsung")
	assertJSONField(t, gotBody, "model", "SM-S931U")
	assertJSONField(t, gotBody, "osversion", "35")
	assertJSONField(t, gotBody, "source", "google-play")

	reportTime, ok := parseHeaderReportTime("1776750213073556", "Asia/Shanghai")
	if !ok {
		t.Fatal("sample timestamp should be parsed")
	}
	assertJSONField(t, gotBody, "xwl_client_time", reportTime)
	assertJSONField(t, gotBody, "xwl_part_date", reportTime)
}

func TestReportSyncModeSkipSignedHeaderBypassesSignatureButStillMergesHeaders(t *testing.T) {
	var (
		called   bool
		gotBody  []byte
		reporter = ReportSyncMode(func(ctx *fasthttp.RequestCtx) {
			called = true
			gotBody = append([]byte(nil), ctx.PostBody()...)
			ctx.SetBodyString("ok")
		})
	)

	oldValue := model.GlobConfig.Report.SkipSigned
	model.GlobConfig.Report.SkipSigned = "report-tools-bypass"
	defer func() {
		model.GlobConfig.Report.SkipSigned = oldValue
	}()

	ctx := newReportSyncTestContext(`/ingress/reportEvent/1001/demo/pay_success/1`, `{"hello":"world"}`)
	applySampleSignedHeaders(ctx)
	ctx.Request.Header.Set("X-SkipSigned", "report-tools-bypass")

	reporter(ctx)

	if !called {
		t.Fatal("skip-signed request should reach next handler")
	}
	assertJSONField(t, gotBody, "hello", "world")
	assertJSONField(t, gotBody, "xwl_distinct_id", "10088711")
}

func TestReportSyncModeFallsBackToInternalPathWhenProxyPrefixIsNotConfigured(t *testing.T) {
	var (
		called   bool
		reporter = ReportSyncMode(func(ctx *fasthttp.RequestCtx) {
			called = true
			ctx.SetBodyString("ok")
		})
	)

	oldPrefix := model.GlobConfig.Report.SignaturePathPrefix
	model.GlobConfig.Report.SignaturePathPrefix = ""
	defer func() {
		model.GlobConfig.Report.SignaturePathPrefix = oldPrefix
	}()

	ctx := newReportSyncTestContext(`/ingress/reportEvent/616077185416953866/fe6ba3206fcf04fe71bedf1ea112cd81/AppLaunch/0`, `{}`)
	applySampleSignedHeaders(ctx)
	setLiveSignature(t, ctx, `/ingress/reportEvent/616077185416953866/fe6ba3206fcf04fe71bedf1ea112cd81/AppLaunch/0`)

	reporter(ctx)

	if !called {
		t.Fatal("internal-path signature should still reach next handler")
	}
}

func newReportSyncTestContext(path, body string) *fasthttp.RequestCtx {
	var req fasthttp.Request
	req.Header.SetMethod(fasthttp.MethodPost)
	req.Header.SetContentType("application/json")
	req.SetBodyString(body)

	var ctx fasthttp.RequestCtx
	ctx.Init(&req, nil, nil)
	ctx.Request.SetRequestURI(path)
	return &ctx
}

func applySampleSignedHeaders(ctx *fasthttp.RequestCtx) {
	headers := map[string]string{
		"X-Lang":      "zh",
		"X-Pkg":       "com.freshflick.app",
		"X-PkgType":   "mobile",
		"X-Ts":        "1776750213073556",
		"X-Tz":        "Asia/Shanghai",
		"X-Dev":       "235520ce1ba2158b8a151f89779494445",
		"X-Name":      "10088711",
		"X-Version":   "20304",
		"X-Referrer":  "GooglePlay",
		"X-Brand":     "samsung",
		"X-model":     "SM-S931U",
		"X-OsVersion": "35",
		"X-EXT":       "2",
		"X-Platform":  "android",
		"X-Source":    "google-play",
	}
	for key, value := range headers {
		ctx.Request.Header.Set(key, value)
	}
}

func setLiveSignature(t *testing.T, ctx *fasthttp.RequestCtx, path string) {
	t.Helper()

	headers := collectReportSignatureHeaders(ctx)
	headerText := buildLiveSignatureHeaderText(headers)
	signature, err := buildShortSignature(path, fasthttp.MethodPost, buildURLValuesFromQueryArgs(ctx.QueryArgs()), headerText)
	if err != nil {
		t.Fatalf("failed to build signature: %v", err)
	}
	ctx.Request.Header.Set("X-Sign", signature)
}

func buildShortSignature(path, method string, queryValues url.Values, headerText string) (string, error) {
	queryText, err := url.QueryUnescape(queryValues.Encode())
	if err != nil {
		return "", err
	}
	return shortHmacSumByTag([]string{path + "|" + method + "|" + queryText + "|" + headerText}), nil
}

func assertJSONField(t *testing.T, body []byte, key, want string) {
	t.Helper()
	if got := gjson.GetBytes(body, key).String(); got != want {
		t.Fatalf("%s = %q, want %q", key, got, want)
	}
}

func assertResponseCode(t *testing.T, body []byte, want int) {
	t.Helper()
	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		t.Fatalf("failed to decode response body %s: %v", string(body), err)
	}
	if got := int(response["code"].(float64)); got != want {
		t.Fatalf("code = %d, want %d", got, want)
	}
}
