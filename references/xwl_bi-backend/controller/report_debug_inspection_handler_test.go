package controller

import (
	"errors"
	"testing"

	"github.com/1340691923/xwl_bi/engine/logs"
	"github.com/1340691923/xwl_bi/model"
	sinkerModel "github.com/1340691923/xwl_bi/platform-basic-libs/sinker/model"
	parser "github.com/1340691923/xwl_bi/platform-basic-libs/sinker/parse"
	"go.uber.org/zap"
)

func TestDebugInspectionHandler(t *testing.T) {
	logs.Logger = zap.NewNop()

	tests := []struct {
		name               string
		request            DecodedReportRequest
		kafkaData          model.KafkaData
		parseMetric        parseMetricFunc
		loadDims           loadDimsFunc
		wantStop           bool
		wantMessage        string
		wantErr            string
		wantDebugDataJudge string
		wantProducerCalled bool
	}{
		{
			name: "非 debug 设备直接跳过",
			request: DecodedReportRequest{
				Debug:      "1",
				DistinctID: "abc",
				Typ:        "reportEvent",
				APPID:      "1001",
			},
			kafkaData: model.KafkaData{
				TableId:    "51",
				EventName:  "pay_success",
				ReportType: model.EventReportType,
				ReqData:    []byte(`{"amount":18}`),
			},
			parseMetric: realParseMetric,
			loadDims: func(tableName string) ([]*sinkerModel.ColumnWithType, error) {
				return []*sinkerModel.ColumnWithType{}, nil
			},
			wantStop: false,
		},
		{
			name: "debug 校验失败会写 debug topic 并阻断正式上报",
			request: DecodedReportRequest{
				Debug:      "1",
				DistinctID: "abc",
				Typ:        "reportEvent",
				APPID:      "1001",
				EventName:  "pay_success",
				Body:       []byte(`{"amount":"bad","xwl_update_time":"2026-04-10 12:00:00"}`),
			},
			kafkaData: model.KafkaData{
				TableId:    "51",
				EventName:  "pay_success",
				ReportTime: "2026-04-10 12:00:00",
				ReportType: model.EventReportType,
				ReqData:    []byte(`{"amount":"bad","xwl_update_time":"2026-04-10 12:00:00"}`),
			},
			parseMetric: realParseMetric,
			loadDims: func(tableName string) ([]*sinkerModel.ColumnWithType, error) {
				return []*sinkerModel.ColumnWithType{
					{Name: "amount", Type: parser.Int},
				}, nil
			},
			wantStop:           true,
			wantErr:            "amount的类型错误，正确类型为数字类型，上报类型为字符串类型(\"bad\")",
			wantDebugDataJudge: "事件属性类型不合法",
			wantProducerCalled: true,
		},
		{
			name: "debug=2 校验通过但不进入正式链路",
			request: DecodedReportRequest{
				Debug:      "2",
				DistinctID: "abc",
				Typ:        "reportEvent",
				APPID:      "1001",
				EventName:  "pay_success",
				Body:       []byte(`{"amount":18,"xwl_update_time":"2026-04-10 12:00:00"}`),
			},
			kafkaData: model.KafkaData{
				TableId:    "51",
				EventName:  "pay_success",
				ReportTime: "2026-04-10 12:00:00",
				ReportType: model.EventReportType,
				ReqData:    []byte(`{"amount":18,"xwl_update_time":"2026-04-10 12:00:00"}`),
			},
			parseMetric: realParseMetric,
			loadDims: func(tableName string) ([]*sinkerModel.ColumnWithType, error) {
				return []*sinkerModel.ColumnWithType{
					{Name: "amount", Type: parser.Int},
				}, nil
			},
			wantStop:           true,
			wantMessage:        "上报成功（数据不入库）",
			wantDebugDataJudge: "数据检验通过",
			wantProducerCalled: true,
		},
		{
			name: "debug=1 校验通过后继续正式链路",
			request: DecodedReportRequest{
				Debug:      "1",
				DistinctID: "abc",
				Typ:        "reportEvent",
				APPID:      "1001",
				EventName:  "pay_success",
				Body:       []byte(`{"amount":18,"xwl_update_time":"2026-04-10 12:00:00"}`),
			},
			kafkaData: model.KafkaData{
				TableId:    "51",
				EventName:  "pay_success",
				ReportTime: "2026-04-10 12:00:00",
				ReportType: model.EventReportType,
				ReqData:    []byte(`{"amount":18,"xwl_update_time":"2026-04-10 12:00:00"}`),
			},
			parseMetric: realParseMetric,
			loadDims: func(tableName string) ([]*sinkerModel.ColumnWithType, error) {
				return []*sinkerModel.ColumnWithType{
					{Name: "amount", Type: parser.Int},
				}, nil
			},
			wantStop:           false,
			wantDebugDataJudge: "数据检验通过",
			wantProducerCalled: true,
		},
		{
			name: "解析失败返回服务异常",
			request: DecodedReportRequest{
				Debug:      "1",
				DistinctID: "abc",
				Typ:        "reportEvent",
				APPID:      "1001",
			},
			kafkaData: model.KafkaData{
				TableId:    "51",
				EventName:  "pay_success",
				ReportType: model.EventReportType,
				ReqData:    []byte(`bad`),
			},
			parseMetric: func(body []byte) (*parser.FastjsonMetric, error) {
				return nil, errors.New("parse failed")
			},
			loadDims: func(tableName string) ([]*sinkerModel.ColumnWithType, error) {
				return []*sinkerModel.ColumnWithType{}, nil
			},
			wantErr: "服务异常",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			called := false
			var sentData map[string]interface{}

			handler := newDebugInspectionHandler(
				tc.parseMetric,
				tc.loadDims,
				func(debug, distinctID, tableID string) bool {
					return tc.name != "非 debug 设备直接跳过"
				},
				func(data map[string]interface{}) error {
					called = true
					sentData = data
					return nil
				},
			)

			decision, err := handler.Handle(tc.request, tc.kafkaData)
			if tc.wantErr != "" {
				if err == nil || err.Error() != tc.wantErr {
					t.Fatalf("Handle error = %v, want %q", err, tc.wantErr)
				}
			} else if err != nil {
				t.Fatalf("Handle returned error: %v", err)
			}

			if decision.Stop != tc.wantStop {
				t.Fatalf("decision.Stop = %v, want %v", decision.Stop, tc.wantStop)
			}
			if decision.Message != tc.wantMessage {
				t.Fatalf("decision.Message = %q, want %q", decision.Message, tc.wantMessage)
			}
			if called != tc.wantProducerCalled {
				t.Fatalf("producer called = %v, want %v", called, tc.wantProducerCalled)
			}
			if tc.wantDebugDataJudge != "" {
				if sentData["data_judge"] != tc.wantDebugDataJudge {
					t.Fatalf("data_judge = %v, want %q", sentData["data_judge"], tc.wantDebugDataJudge)
				}
			}
		})
	}
}

func realParseMetric(body []byte) (*parser.FastjsonMetric, error) {
	p := &parser.FastjsonParser{}
	return p.Parse(body)
}
