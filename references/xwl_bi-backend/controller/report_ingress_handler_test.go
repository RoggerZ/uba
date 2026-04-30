package controller

import (
	"errors"
	"testing"

	"github.com/1340691923/xwl_bi/model"
	"github.com/1340691923/xwl_bi/platform-basic-libs/service/report"
)

func TestReportIngressHandlerHandle(t *testing.T) {
	request := DecodedReportRequest{
		Typ:        "reportEvent",
		APPID:      "1001",
		AppKey:     "demo",
		Debug:      "1",
		EventName:  "pay_success",
		ClientIP:   "1.1.1.1",
		ReportTime: "2026-04-10 12:00:00",
		Body:       []byte(`{"amount":18}`),
		DistinctID: "abc",
	}

	t.Run("正常上报进入正式 Kafka", func(t *testing.T) {
		reportSent := false
		handler := newReportIngressHandler(
			func(appid, appkey string) (string, error) {
				return "51", nil
			},
			report.DefaultPayloadBuilderRegistry().Build,
			func(request DecodedReportRequest, kafkaData model.KafkaData) (debugInspectionDecision, error) {
				return debugInspectionDecision{}, nil
			},
			func(data model.KafkaData) error {
				reportSent = true
				return nil
			},
		)

		result, err := handler.Handle(request)
		if err != nil {
			t.Fatalf("Handle returned error: %v", err)
		}
		if result.Message != "上报成功" {
			t.Fatalf("Message = %q, want %q", result.Message, "上报成功")
		}
		if !reportSent {
			t.Fatal("expected report data to be sent")
		}
	})

	t.Run("debug 决策为停止时不进入正式 Kafka", func(t *testing.T) {
		reportSent := false
		handler := newReportIngressHandler(
			func(appid, appkey string) (string, error) {
				return "51", nil
			},
			report.DefaultPayloadBuilderRegistry().Build,
			func(request DecodedReportRequest, kafkaData model.KafkaData) (debugInspectionDecision, error) {
				return debugInspectionDecision{
					Stop:    true,
					Message: "上报成功（数据不入库）",
				}, nil
			},
			func(data model.KafkaData) error {
				reportSent = true
				return nil
			},
		)

		result, err := handler.Handle(request)
		if err != nil {
			t.Fatalf("Handle returned error: %v", err)
		}
		if result.Message != "上报成功（数据不入库）" {
			t.Fatalf("Message = %q, want %q", result.Message, "上报成功（数据不入库）")
		}
		if reportSent {
			t.Fatal("expected report data not to be sent")
		}
	})

	t.Run("非法 typ 直接返回错误", func(t *testing.T) {
		handler := newReportIngressHandler(
			func(appid, appkey string) (string, error) {
				return "51", nil
			},
			report.DefaultPayloadBuilderRegistry().Build,
			func(request DecodedReportRequest, kafkaData model.KafkaData) (debugInspectionDecision, error) {
				return debugInspectionDecision{}, nil
			},
			func(data model.KafkaData) error {
				return nil
			},
		)

		invalidRequest := request
		invalidRequest.Typ = "bad-type"

		_, err := handler.Handle(invalidRequest)
		if err == nil || err.Error() != report.ERROR_TABLE[report.ReportTypeErr] {
			t.Fatalf("Handle error = %v, want %q", err, report.ERROR_TABLE[report.ReportTypeErr])
		}
	})

	t.Run("tableId 解析失败直接返回", func(t *testing.T) {
		handler := newReportIngressHandler(
			func(appid, appkey string) (string, error) {
				return "", errors.New("服务异常")
			},
			report.DefaultPayloadBuilderRegistry().Build,
			func(request DecodedReportRequest, kafkaData model.KafkaData) (debugInspectionDecision, error) {
				return debugInspectionDecision{}, nil
			},
			func(data model.KafkaData) error {
				return nil
			},
		)

		_, err := handler.Handle(request)
		if err == nil || err.Error() != "服务异常" {
			t.Fatalf("Handle error = %v, want %q", err, "服务异常")
		}
	})
}
