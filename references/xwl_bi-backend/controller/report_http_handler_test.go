package controller

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/1340691923/xwl_bi/engine/logs"
	"github.com/1340691923/xwl_bi/model"
	"github.com/1340691923/xwl_bi/platform-basic-libs/service/report"
	sinkerModel "github.com/1340691923/xwl_bi/platform-basic-libs/sinker/model"
	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
)

func TestReportHTTPHandlerOutputs(t *testing.T) {
	logs.Logger = zap.NewNop()

	handler := NewReportHandler(ReportHandlerDependencies{
		Now: func() time.Time {
			return time.Date(2026, 4, 10, 12, 0, 0, 0, time.Local)
		},
		ClientIP: func(ctx *fasthttp.RequestCtx) string {
			return "10.10.10.10"
		},
		ResolveTableID: func(appid, appkey string) (string, error) {
			return "51", nil
		},
		BuildPayload: report.DefaultPayloadBuilderRegistry().Build,
		IsDebugDevice: func(debug, distinctID, tableID string) bool {
			return debug == report.DebugNotToDB
		},
		SendDebugData: func(data map[string]interface{}) error {
			return nil
		},
		SendReportData: func(data model.KafkaData) error {
			return nil
		},
		LoadDims: func(tableName string) ([]*sinkerModel.ColumnWithType, error) {
			return []*sinkerModel.ColumnWithType{}, nil
		},
	})

	t.Run("缺少 distinct_id 返回错误结构", func(t *testing.T) {
		ctx := newHTTPHandlerContext(`{"xwl_part_date":"2026-04-08"}`, "pay_success", "1")
		handler(ctx)

		got := decodeBodyMap(t, ctx.Response.Body())
		if got["code"].(float64) != 500 {
			t.Fatalf("code = %v, want %v", got["code"], 500)
		}
		if got["msg"] != "xwl_distinct_id 不能为空" {
			t.Fatalf("msg = %v, want %q", got["msg"], "xwl_distinct_id 不能为空")
		}
	})

	t.Run("debug=2 校验通过返回成功但不入库", func(t *testing.T) {
		ctx := newHTTPHandlerContext(`{"xwl_distinct_id":"abc","xwl_update_time":"2026-04-10 12:00:00"}`, "pay_success", "2")
		handler(ctx)

		got := decodeBodyMap(t, ctx.Response.Body())
		if got["code"].(float64) != 0 {
			t.Fatalf("code = %v, want %v", got["code"], 0)
		}
		if got["msg"] != "上报成功（数据不入库）" {
			t.Fatalf("msg = %v, want %q", got["msg"], "上报成功（数据不入库）")
		}
	})

	t.Run("普通成功返回标准成功文案", func(t *testing.T) {
		ctx := newHTTPHandlerContext(`{"xwl_distinct_id":"abc"}`, "pay_success", "0")
		handler(ctx)

		got := decodeBodyMap(t, ctx.Response.Body())
		if got["code"].(float64) != 0 {
			t.Fatalf("code = %v, want %v", got["code"], 0)
		}
		if got["msg"] != "上报成功" {
			t.Fatalf("msg = %v, want %q", got["msg"], "上报成功")
		}
	})
}

func newHTTPHandlerContext(body, eventName, debug string) *fasthttp.RequestCtx {
	ctx := newReportRequestContext(body)
	ctx.SetUserValue("typ", sinkerModel.ReportEventProperties)
	ctx.SetUserValue("appid", "1001")
	ctx.SetUserValue("appkey", "demo")
	ctx.SetUserValue("debug", debug)
	ctx.SetUserValue("eventName", eventName)
	return ctx
}

func decodeBodyMap(t *testing.T, body []byte) map[string]interface{} {
	t.Helper()

	var got map[string]interface{}
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("failed to decode response body %s: %v", string(body), err)
	}
	return got
}
