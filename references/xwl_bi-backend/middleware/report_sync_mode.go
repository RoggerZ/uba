package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/1340691923/xwl_bi/model"
	"github.com/1340691923/xwl_bi/platform-basic-libs/util"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"github.com/valyala/fasthttp"
)

var (
	// reportSignatureHeaders 直接按 shortAAA 的顺序拷贝。
	// 这里不能为了“看起来更整齐”随意调整顺序，因为签名底串里
	// Header 的顺序一旦变化，线上 App 算出来的 X-Sign 就会全部失效。
	reportSignatureHeaders = []string{
		"X-Lang",
		"X-Pkg",
		"X-PkgType",
		"X-Tkn",
		"X-Ts",
		"X-Tz",
		"X-Dev",
		"X-Name",
		"X-Version",
		"X-Referrer",
		"X-Brand",
		"X-model",
		"X-OsVersion",
	}
	// reportSignatureMethods 与 shortAAA 的 VerifyShort 保持一致，
	// 这样请求方法大小写和合法性判断不会和线上链路出现分叉。
	reportSignatureMethods = map[string]bool{
		http.MethodGet:     true,
		http.MethodPost:    true,
		http.MethodHead:    true,
		http.MethodPut:     true,
		http.MethodPatch:   true,
		http.MethodDelete:  true,
		http.MethodConnect: true,
		http.MethodOptions: true,
		http.MethodTrace:   true,
	}
	// shortAPIHashSalt / shortAPIHashKey 从 shortAAA 拷贝而来。
	// 它们共同决定 VerifyShort 最终的 HMAC 结果，不能做本地“简化版”替换。
	shortAPIHashSalt = []byte{
		0xA7, 0x00, 0xFB, 0x15, 0x6A, 0xFD, 0x0D, 0x50, 0x64, 0xBC, 0x3F, 0x0B, 0x8D, 0x4C, 0x8F, 0xF1,
		0xC6, 0xD2, 0x7E, 0xF7, 0xC3, 0xF8, 0x62, 0x49, 0x62, 0x2A, 0xA5, 0x95, 0xAD, 0xD2, 0x59, 0x77,
		0xF7, 0xAC, 0xE2, 0x0B, 0x2D, 0x49, 0xDC, 0xC7, 0x61, 0x71, 0x58, 0x4E, 0x85, 0xD7, 0xA7, 0x2B,
		0x28, 0x22, 0xB1, 0x83, 0x76, 0x40, 0xBA, 0x9B, 0xE3, 0xEE, 0x56, 0x6D, 0xA6, 0xF9, 0x05, 0x4C,
		0x5B, 0x02, 0x77, 0xD5, 0xCD, 0x8E, 0x68, 0x3D, 0xDA, 0x3D, 0xD3, 0x30, 0x6A, 0x33, 0xBB, 0xC4,
		0x39, 0xB7, 0x6A, 0xED, 0x6E, 0x40, 0x33, 0xE0, 0xF9, 0x73, 0xE0, 0xCB, 0x4E, 0x5E, 0xDB, 0x72,
		0xCF, 0x4F, 0x64, 0xC3, 0xF5, 0x5F, 0x67, 0xA0, 0x12, 0xB1, 0xCE, 0xFB, 0x8A, 0x64, 0x13, 0xDE,
		0x1D, 0xA0, 0x22, 0x86, 0xB4, 0x78, 0x9F, 0x2C, 0xDE, 0x64, 0x23, 0xB2, 0x6D, 0xD6, 0x0E, 0x5D,
		0xC3, 0x9B, 0x4C, 0x7E, 0x0F, 0x94, 0x36, 0x44, 0x82, 0x6A, 0x10, 0xEC, 0x94, 0x75, 0xAD, 0xF4,
		0x69, 0xEE, 0xF9, 0x34, 0x92, 0x12, 0xDF, 0xE2, 0x4B, 0x63, 0x8C, 0x7B, 0x43, 0xA2, 0x28, 0xD6,
		0x25, 0xBA, 0x1B, 0xC2, 0xE6, 0x88, 0x69, 0xC7, 0x89, 0x1E, 0xD9, 0x9E, 0xA4, 0x68, 0x9A, 0x1B,
		0x1E, 0x73, 0x94, 0xD6, 0xFC, 0x02, 0x95, 0xD0, 0x0A, 0x3A, 0x00, 0x2C, 0xEF, 0x7A, 0x92, 0xCB,
		0x30, 0x6A, 0x6D, 0x70, 0xCB, 0x68, 0x6F, 0x6E, 0x97, 0x68, 0x0C, 0x68, 0x90, 0x7D, 0x38, 0x20,
		0xCA, 0x23, 0xAA, 0xF4, 0x82, 0x49, 0x8D, 0xDB, 0x39, 0x43, 0x5E, 0xCA, 0x4D, 0x48, 0x74, 0x83,
		0x1D, 0xBA, 0x75, 0x0D, 0xF3, 0xF5, 0xBF, 0xF9, 0x38, 0x37, 0x97, 0x54, 0x02, 0x54, 0x92, 0x67,
		0x84, 0xB9, 0xFA, 0xE3, 0xD7, 0x79, 0x84, 0x36, 0x76, 0x39, 0x0C, 0xA0, 0x38, 0xB1, 0x3B, 0x13,
		0xB5, 0xD6, 0x2B, 0x58, 0x4D, 0xFB, 0xD5, 0x1C, 0xA4, 0xBB, 0x2F, 0xF4, 0xDE, 0x83, 0xB3, 0xB6,
		0xA5, 0x15, 0xBE, 0x29, 0xBE, 0xF2, 0x93, 0x2E, 0x4D, 0xA6, 0x3F, 0x63, 0x11, 0x68, 0xE4, 0x7C,
		0xCF, 0x3F, 0xCA, 0x2D, 0xED, 0xCC, 0x40, 0xCA, 0x17, 0x62, 0x84, 0x45, 0x3D, 0x8E, 0x8A, 0xF3,
		0xFD, 0x6A, 0x8A, 0xA4, 0x7B, 0xCC, 0x85, 0xBD, 0x4D, 0xEB, 0x35, 0xC2, 0x59, 0x2E, 0x55, 0x12,
		0x7D, 0xFC, 0x99, 0x34, 0x3D, 0x78, 0x1A, 0x45, 0x5C, 0xA6, 0x7B, 0x2B, 0x96, 0x24, 0xD8, 0xC2,
		0xB8, 0x4A, 0xD2, 0xCE, 0x35, 0x28, 0x56, 0xC3, 0x0F, 0xD4, 0x26, 0xCB, 0xE6, 0xCB, 0x0F, 0x66,
		0xB5, 0x0D, 0x14, 0xA7, 0x37, 0xD3, 0x34, 0x72, 0xBE, 0x6E, 0x6E, 0x3F, 0x2F, 0x96, 0xB3, 0xA4,
		0x9C, 0x98, 0x3D, 0x5C, 0xD7, 0xDB, 0x4D, 0x2D, 0x64, 0xA6, 0x27, 0x9B, 0x5F, 0x9E, 0xB4, 0x13,
		0xF6, 0xF7, 0x74, 0x0D, 0xD4, 0x97, 0x79, 0xDE, 0x9A, 0x9E, 0xB2, 0x68, 0x6A, 0x4C, 0x06, 0x97,
		0x1F, 0x69, 0xBE, 0x82, 0x45, 0x0A, 0x5F, 0x9C, 0x56, 0xFB, 0xF7, 0x6E,
	}
	shortAPIHashKey = []byte{
		0x8C, 0x4D, 0x8D, 0x43, 0xB6, 0xAA, 0x50, 0x88, 0x2B, 0x1D, 0x73, 0x3D, 0x65, 0xAF, 0x3E, 0x8D,
		0x60, 0x6B, 0xD1, 0x35, 0x5A, 0x1C, 0xD7, 0x2D, 0x95, 0xCF, 0xEE, 0x5D, 0x2B, 0x8A, 0x31, 0x19,
		0x9B, 0xF9, 0xBB, 0xB7, 0xC7, 0xA8, 0x72, 0x67, 0x13, 0x6E, 0x4A, 0xED, 0x38, 0xEC, 0x8B, 0x5A,
		0xB7, 0x9D, 0x00, 0xBF, 0x5C, 0xAA, 0xFF, 0x4A, 0x2A, 0x57, 0xC1, 0xB0, 0x52, 0x32, 0x17, 0x5E,
		0x6B, 0x14, 0xBC, 0x33, 0x3F, 0xF1, 0x6C, 0x14, 0xA0, 0x8D, 0xB3, 0xBD, 0x5F, 0xED, 0x09, 0xC7,
		0xF0, 0xCD, 0x4B, 0x45, 0xBC, 0xE7, 0x8E, 0x41, 0x71, 0x70, 0xBE, 0x30, 0x5E, 0x7C, 0xF0, 0x08,
		0xE9, 0xE7, 0x9E, 0xEF, 0xE3, 0x83, 0xD8, 0x39, 0x34, 0xC5, 0xE0, 0xF2, 0xF6, 0xB1, 0xBB, 0xBB,
		0x57, 0x27, 0x5A, 0x09, 0x41, 0x87, 0x21, 0xAC, 0x9B, 0x87, 0x6B, 0xDB, 0x76, 0x9B, 0x3D, 0x7C,
		0xC3, 0x72, 0xA2, 0x5A, 0x25, 0xBE, 0x4B, 0xE1, 0x77, 0xE0, 0x0D, 0xCB, 0xFD, 0x2D, 0xAD, 0xB0,
		0xAA, 0xE0, 0xEF, 0x7C, 0x04, 0x17, 0xCE, 0xC5, 0xE2, 0x15, 0x1A, 0x30, 0x44, 0x40, 0xD4, 0x37,
		0x2E, 0x44, 0x1C, 0xDD, 0xB0, 0x25, 0x03, 0xC7, 0xBE, 0x24, 0x99, 0x5E, 0x24, 0x66, 0x48, 0x09,
		0x5A, 0xBD, 0xFD, 0xA1, 0x0D, 0x84, 0xA3, 0x96, 0x97, 0x3F, 0x7D, 0xA2, 0xE3, 0x9A, 0xC3, 0x9F,
		0xF5, 0x2D, 0x2E, 0x6E, 0xA4, 0x4C, 0xED, 0xDF, 0xEB, 0x8A, 0xFC, 0x03, 0xEF, 0xA1, 0x56, 0xC4,
		0xAB, 0x2E, 0xC2, 0x9A, 0xC5, 0x99, 0xA2, 0x34, 0x0A, 0x4B, 0xF2, 0xDE, 0x2B, 0xED, 0x75, 0x56,
		0xD1, 0xEB, 0xA4, 0x8A, 0x15, 0xF5, 0x3C, 0x70, 0x82, 0x40, 0x0A, 0x72, 0x2B, 0x2C, 0x0C, 0x92,
		0x25, 0x5F, 0x53, 0x0C, 0x83, 0xBB, 0x3C, 0x29, 0x3F, 0x9C, 0xE3, 0x48, 0x38, 0xE4, 0x25, 0xD2,
		0x8E, 0xA2, 0x69, 0x85, 0xC6, 0x62, 0x4D, 0x47, 0x01, 0x8E, 0x34, 0xE6, 0x7A, 0x38, 0x03, 0x46,
		0x17, 0x57, 0xCA, 0xCB, 0x5F, 0xEA, 0x88, 0xE7, 0x82, 0x25, 0x3E, 0x2D, 0x63, 0x94, 0xAD, 0x24,
		0x54, 0xE5, 0x23, 0xAC, 0xD0, 0x3F, 0x72, 0x22, 0x0D, 0x88, 0x63, 0x7E, 0x31, 0xD8, 0x5C, 0xDF,
		0x62, 0xE0, 0x21, 0xB5, 0x18, 0xE7, 0xDB, 0xE8, 0xC0, 0x43, 0xFC, 0x8D, 0x63, 0x7A, 0x45, 0x26,
		0xE1, 0xF4, 0x01, 0x63, 0xE6, 0x09, 0x22, 0xF7, 0x3C, 0xB4, 0xF8, 0x7D, 0x1B, 0x5E, 0x5D, 0xE4,
		0xA2, 0xF5, 0xBD, 0x1E, 0x3F, 0x7A, 0x17, 0x30, 0xD6, 0xF1, 0x6D, 0x09, 0x38, 0x63, 0x04, 0x0C,
		0xC2, 0xAC, 0x1C, 0x55, 0xE5, 0xEB, 0x50, 0xBB, 0x5B, 0x00, 0x91, 0xD9, 0xF5, 0xA7, 0xD6, 0x68,
		0xC1, 0x43, 0xC4, 0x53, 0xEC, 0xFB,
	}
)

type reportSyncModeError struct {
	code int
	msg  string
}

func (e *reportSyncModeError) Error() string {
	return e.msg
}

// ReportSyncMode 是 report_server 在 HTTP 入口层新增的兼容中间件。
//
// 处理顺序固定为：
// 1. 先把 Header 里的公共参数补进 body。
// 2. 再根据 X-SkipSigned 判断是否跳过签名。
// 3. 如未跳过，再执行签名校验。
//
// 这样安排的原因是：
// 1. tools 虽然可以跳过签名，但如果它们额外带了 Header，也仍然可以复用同一套“只补缺”的回填逻辑。
// 2. reportHandler / decoder / ingress 后面都继续只认 body，不需要知道前面是 App 还是 tools。
func ReportSyncMode(handle fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		if strings.EqualFold(util.Bytes2str(ctx.Method()), fasthttp.MethodOptions) {
			handle(ctx)
			return
		}

		if err := prepareReportRequest(ctx); err != nil {
			writeReportSyncModeError(ctx, err)
			return
		}
		handle(ctx)
	}
}

func prepareReportRequest(ctx *fasthttp.RequestCtx) *reportSyncModeError {
	// 先补齐 body，再决定是否验签。
	// 这样即便 tools 走跳签名分支，也不会失去 Header -> body 的兼容能力。
	body, err := mergeSignedReportBody(ctx)
	if err != nil {
		return err
	}
	ctx.Request.SetBodyRaw(body)

	if shouldSkipReportSignature(ctx) {
		return nil
	}

	if err := verifySignedReportRequest(ctx); err != nil {
		return err
	}
	return nil
}

// shouldSkipReportSignature 用配置值控制“谁可以跳过签名”。
//
// 这里故意不做：
// 1. 只要带 X-SkipSigned 就跳过
// 2. 前缀匹配
// 3. 大小写不敏感比较
//
// 原因是这个头本质上是一个共享秘密，必须要求“值完全相等”才放行。
func shouldSkipReportSignature(ctx *fasthttp.RequestCtx) bool {
	configuredValue := strings.TrimSpace(model.GlobConfig.Report.SkipSigned)
	if configuredValue == "" {
		return false
	}

	headerValue := readHeaderValue(ctx, "X-SkipSigned")
	if headerValue == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(headerValue), []byte(configuredValue)) == 1
}

// mergeSignedReportBody 负责把 App 可能放在 Header 里的公共参数回填到 body。
//
// 关键约束：
// 1. 只补缺，不覆盖 body 中已有值，避免改变现有直传语义。
// 2. 最终仍然要求请求体是 JSON 对象，因为 decoder / payload builder 都是按对象结构继续往下走。
// 3. xwl_ip 不在这里写入，继续沿用后面 decoder 的真实 IP 提取逻辑。
//
// 示例：
//   - body={}
//   - header 中有 X-Name=10088711、X-Platform=android
//   - 回填后 body 至少会包含：
//     {"xwl_distinct_id":"10088711","user_id":"10088711","platform":"android"}
func mergeSignedReportBody(ctx *fasthttp.RequestCtx) ([]byte, *reportSyncModeError) {
	body := bytes.TrimSpace(ctx.PostBody())
	if len(body) == 0 {
		body = []byte("{}")
	}
	if !json.Valid(body) {
		return nil, &reportSyncModeError{code: 400, msg: "请求体必须是合法 JSON"}
	}
	if bytes.TrimSpace(body)[0] != '{' {
		return nil, &reportSyncModeError{code: 400, msg: "请求体必须是 JSON 对象"}
	}

	var err error

	body, err = setJSONFieldIfMissing(body, "xwl_distinct_id", readHeaderValue(ctx, "X-Name"))
	if err != nil {
		return nil, newReportJSONWriteError(err)
	}
	body, err = setJSONFieldIfMissing(body, "user_id", readHeaderValue(ctx, "X-Name"))
	if err != nil {
		return nil, newReportJSONWriteError(err)
	}
	body, err = setJSONFieldIfMissing(body, "device_id", readHeaderValue(ctx, "X-Dev"))
	if err != nil {
		return nil, newReportJSONWriteError(err)
	}
	body, err = setJSONFieldIfMissing(body, "referrer", readHeaderValue(ctx, "X-Referrer"))
	if err != nil {
		return nil, newReportJSONWriteError(err)
	}
	body, err = setJSONFieldIfMissing(body, "app_version", readHeaderValue(ctx, "X-Version"))
	if err != nil {
		return nil, newReportJSONWriteError(err)
	}
	body, err = setJSONFieldIfMissing(body, "platform", readHeaderValue(ctx, "X-Platform"))
	if err != nil {
		return nil, newReportJSONWriteError(err)
	}
	body, err = setJSONFieldIfMissing(body, "lang", readHeaderValue(ctx, "X-Lang"))
	if err != nil {
		return nil, newReportJSONWriteError(err)
	}
	body, err = setJSONFieldIfMissing(body, "brand", readHeaderValue(ctx, "X-Brand"))
	if err != nil {
		return nil, newReportJSONWriteError(err)
	}
	body, err = setJSONFieldIfMissing(body, "model", firstNonEmpty(
		readHeaderValue(ctx, "X-Model"),
		readHeaderValue(ctx, "X-model"),
	))
	if err != nil {
		return nil, newReportJSONWriteError(err)
	}
	body, err = setJSONFieldIfMissing(body, "osversion", firstNonEmpty(
		readHeaderValue(ctx, "X-Osversion"),
		readHeaderValue(ctx, "X-OsVersion"),
	))
	if err != nil {
		return nil, newReportJSONWriteError(err)
	}
	body, err = setJSONFieldIfMissing(body, "source", readHeaderValue(ctx, "X-Source"))
	if err != nil {
		return nil, newReportJSONWriteError(err)
	}

	// X-Ts / X-Tz 只在 body 缺少时间字段时兜底写入。
	// 这样可以兼容：
	// 1. 线上 App 只发 Header 时间
	// 2. 本地工具直接在 body 里传 xwl_client_time / xwl_part_date
	reportTime, ok := parseHeaderReportTime(readHeaderValue(ctx, "X-Ts"), readHeaderValue(ctx, "X-Tz"))
	if ok {
		body, err = setJSONFieldIfMissing(body, "xwl_client_time", reportTime)
		if err != nil {
			return nil, newReportJSONWriteError(err)
		}
		body, err = setJSONFieldIfMissing(body, "xwl_part_date", reportTime)
		if err != nil {
			return nil, newReportJSONWriteError(err)
		}
	}

	return body, nil
}

// verifySignedReportRequest 负责执行 shortAAA 对齐版的 VerifyShort。
//
// 这里会尝试两类底串：
// 1. Header 全量参与的 strict 模式
// 2. 只拼接实际存在 Header 的 live 模式
//
// 同时也会尝试两类路径：
// 1. 带代理前缀的外部路径，例如 /api/v3/reporter/ingress/...
// 2. report_service 实际收到的内部路径 /ingress/...
//
// 这样做的原因是：
// 1. 线上签名底串通常包含 Nginx 暴露给客户端的前缀。
// 2. 本地直连 8091 调试时，签名底串可能只包含内部路径。
func verifySignedReportRequest(ctx *fasthttp.RequestCtx) *reportSyncModeError {
	sign := readHeaderValue(ctx, "X-Sign")
	if sign == "" {
		return &reportSyncModeError{code: 400, msg: "X-Sign 不能为空"}
	}
	timestamp := readHeaderValue(ctx, "X-Ts")
	if timestamp == "" {
		return &reportSyncModeError{code: 400, msg: "X-Ts 不能为空"}
	}
	timeZone := readHeaderValue(ctx, "X-Tz")
	if timeZone == "" {
		return &reportSyncModeError{code: 400, msg: "X-Tz 不能为空"}
	}

	headers := collectReportSignatureHeaders(ctx)
	queryValues := buildURLValuesFromQueryArgs(ctx.QueryArgs())
	method := util.Bytes2str(ctx.Method())
	paths := buildSignaturePaths(util.Bytes2str(ctx.Path()))

	strictHeaderText := buildStrictSignatureHeaderText(headers)
	liveHeaderText := buildLiveSignatureHeaderText(headers)
	for _, path := range paths {
		ok, err := verifyShortSignature(sign, timestamp, path, method, queryValues, strictHeaderText)
		if err != nil {
			return &reportSyncModeError{code: 400, msg: err.Error()}
		}
		if ok {
			return nil
		}

		ok, err = verifyShortSignature(sign, timestamp, path, method, queryValues, liveHeaderText)
		if err != nil {
			return &reportSyncModeError{code: 400, msg: err.Error()}
		}
		if ok {
			return nil
		}
	}

	return &reportSyncModeError{code: 402, msg: "签名校验失败"}
}

// buildStrictSignatureHeaderText 保留 shortAAA 的“空值也占位”规则。
func buildStrictSignatureHeaderText(headers map[string]string) string {
	parts := make([]string, 0, len(reportSignatureHeaders))
	for _, key := range reportSignatureHeaders {
		parts = append(parts, key+"="+headers[key])
	}
	return strings.Join(parts, "&")
}

// buildLiveSignatureHeaderText 对齐线上实样本的“只拼非空 Header”规则。
func buildLiveSignatureHeaderText(headers map[string]string) string {
	parts := make([]string, 0, len(reportSignatureHeaders))
	for _, key := range reportSignatureHeaders {
		value := headers[key]
		if value == "" {
			continue
		}
		parts = append(parts, key+"="+value)
	}
	return strings.Join(parts, "&")
}

// collectReportSignatureHeaders 把签名需要的 Header 收拢成稳定的 canonical map。
//
// 注意：
// 1. 这里只做签名别名归一化，不等价于 body 字段映射。
// 2. X-PkgType / X-PackageType 只在签名里使用，不会写入 body。
func collectReportSignatureHeaders(ctx *fasthttp.RequestCtx) map[string]string {
	return map[string]string{
		"X-Lang":      readHeaderValue(ctx, "X-Lang"),
		"X-Pkg":       readHeaderValue(ctx, "X-Pkg"),
		"X-PkgType":   firstNonEmpty(readHeaderValue(ctx, "X-PkgType"), readHeaderValue(ctx, "X-PackageType")),
		"X-Ts":        readHeaderValue(ctx, "X-Ts"),
		"X-Tz":        readHeaderValue(ctx, "X-Tz"),
		"X-Dev":       readHeaderValue(ctx, "X-Dev"),
		"X-Name":      readHeaderValue(ctx, "X-Name"),
		"X-Version":   readHeaderValue(ctx, "X-Version"),
		"X-Referrer":  readHeaderValue(ctx, "X-Referrer"),
		"X-Brand":     readHeaderValue(ctx, "X-Brand"),
		"X-model":     firstNonEmpty(readHeaderValue(ctx, "X-model"), readHeaderValue(ctx, "X-Model")),
		"X-OsVersion": firstNonEmpty(readHeaderValue(ctx, "X-OsVersion"), readHeaderValue(ctx, "X-Osversion")),
	}
}

// buildSignaturePaths 根据内部路径重建可参与验签的候选路径。
//
// 顺序固定为：
// 1. 配置的代理前缀 + 内部路径
// 2. 内部路径本身
//
// 这样可以同时兼容：
// - 客户端按外部路径 /api/v3/reporter/ingress/... 计算签名
// - 本地直连 8091 时按 /ingress/... 计算签名
func buildSignaturePaths(internalPath string) []string {
	paths := make([]string, 0, 2)
	addPath := func(path string) {
		path = strings.TrimSpace(path)
		if path == "" {
			return
		}
		for _, existing := range paths {
			if existing == path {
				return
			}
		}
		paths = append(paths, path)
	}

	prefix := strings.TrimSpace(model.GlobConfig.Report.SignaturePathPrefix)
	if prefix != "" {
		addPath(joinSignaturePathPrefix(prefix, internalPath))
	}
	addPath(internalPath)
	return paths
}

// joinSignaturePathPrefix 负责把代理前缀和程序内路径拼成一个稳定的验签路径。
//
// 示例：
// - prefix=/api/v3/reporter
// - internal=/ingress/reportEvent/...
// - output=/api/v3/reporter/ingress/reportEvent/...
func joinSignaturePathPrefix(prefix, internalPath string) string {
	prefix = strings.TrimRight(strings.TrimSpace(prefix), "/")
	if prefix == "" {
		return internalPath
	}
	if !strings.HasPrefix(internalPath, "/") {
		internalPath = "/" + internalPath
	}
	return prefix + internalPath
}

// verifyShortSignature 是 shortAAA VerifyShort 的本地等价实现。
// 它只负责“把参与签名的字段按固定格式拼串并比对摘要”，
// 不负责决定 Header 是否缺失、哪些路径候选需要尝试。
func verifyShortSignature(auth, date, path, method string, queryValues url.Values, headerText string) (bool, error) {
	if date == "" {
		return false, errors.New("X-Ts 不能为空")
	}
	if path == "" {
		return false, errors.New("请求路径不能为空")
	}
	if method == "" {
		return false, errors.New("请求方法不能为空")
	}

	methodName := strings.ToUpper(method)
	if !reportSignatureMethods[methodName] {
		return false, errors.New("请求方法不合法")
	}

	queryText, err := url.QueryUnescape(queryValues.Encode())
	if err != nil {
		return false, err
	}

	buffer := strings.Join([]string{path, methodName, queryText, headerText}, "|")
	digest := shortHmacSumByTag([]string{buffer})
	return subtle.ConstantTimeCompare([]byte(auth), []byte(digest)) == 1, nil
}

// shortHmacSumByTag 保留 shortAAA 的哈希盐、哈希 key 和 URL-safe Base64 输出格式。
func shortHmacSumByTag(keys []string) string {
	mac := hmac.New(sha256.New, shortAPIHashKey)
	mac.Write(shortAPIHashSalt)
	for _, key := range keys {
		mac.Write([]byte(key))
	}
	return base64.URLEncoding.EncodeToString(mac.Sum(nil))
}

// buildURLValuesFromQueryArgs 把 fasthttp.Args 转成 url.Values，
// 目的是沿用 shortAAA VerifyShort 里原本的 QueryUnescape + Encode 排序语义。
func buildURLValuesFromQueryArgs(args *fasthttp.Args) url.Values {
	values := make(url.Values)
	args.VisitAll(func(key, value []byte) {
		values.Add(util.Bytes2str(key), util.Bytes2str(value))
	})
	return values
}

// parseHeaderReportTime 把 Header 里的 X-Ts / X-Tz 转成 body 里使用的标准时间字符串。
//
// 当前显式兼容三种时间戳：
// 1. 10 位秒
// 2. 13 位毫秒
// 3. 16 位微秒
//
// 示例：
// - "1776750213" -> 秒级时间戳
// - "1776750213073" -> 毫秒级时间戳
// - "1776750213073556" -> 微秒级时间戳
func parseHeaderReportTime(ts, timeZone string) (string, bool) {
	ts = strings.TrimSpace(ts)
	timeZone = strings.TrimSpace(timeZone)
	if ts == "" || timeZone == "" {
		return "", false
	}

	location, err := time.LoadLocation(timeZone)
	if err != nil {
		return "", false
	}

	value, err := strconv.ParseInt(ts, 10, 64)
	if err != nil {
		return "", false
	}

	var tm time.Time
	switch len(ts) {
	case 10:
		tm = time.Unix(value, 0)
	case 13:
		tm = time.Unix(0, value*int64(time.Millisecond))
	case 16:
		tm = time.Unix(0, value*int64(time.Microsecond))
	default:
		return "", false
	}
	return tm.In(location).Format(util.TimeFormat), true
}

// setJSONFieldIfMissing 统一实现“只补缺”的 JSON 写入规则。
// 这能保证 header->body 转换不会悄悄覆盖 tools 或本地调用者显式传入的 body 字段。
func setJSONFieldIfMissing(body []byte, field, value string) ([]byte, error) {
	if strings.TrimSpace(value) == "" {
		return body, nil
	}
	if strings.TrimSpace(gjson.GetBytes(body, field).String()) != "" {
		return body, nil
	}
	return sjson.SetBytes(body, field, value)
}

func readHeaderValue(ctx *fasthttp.RequestCtx, key string) string {
	return strings.TrimSpace(util.Bytes2str(ctx.Request.Header.Peek(key)))
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func writeReportSyncModeError(ctx *fasthttp.RequestCtx, err *reportSyncModeError) {
	ctx.SetStatusCode(fasthttp.StatusOK)
	_ = util.WriteJSON(ctx, map[string]interface{}{
		"code": err.code,
		"msg":  err.msg,
	})
}

func newReportJSONWriteError(err error) *reportSyncModeError {
	return &reportSyncModeError{code: 400, msg: "请求体补齐失败: " + err.Error()}
}
