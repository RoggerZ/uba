package runner

import (
	"testing"
	"time"

	"github.com/1340691923/xwl_bi/model"
	jsoniter "github.com/json-iterator/go"
)

func TestMessageDecoderDecode(t *testing.T) {
	decoder := newMessageDecoder(jsoniter.ConfigCompatibleWithStandardLibrary)
	marked := false

	decoded, err := decoder.Decode(newReportInputMessage(model.KafkaData{
		TableId:    "51",
		EventName:  "AppLaunch",
		ReportTime: "2026-04-08",
		ReqData:    []byte(`{"xwl_client_time":"2026-04-08 16:14:53","xwl_distinct_id":"u-1"}`),
	}), func() {
		marked = true
	})
	if err != nil {
		t.Fatalf("Decode returned error: %v", err)
	}

	if decoded.KafkaData.ReportTime != "2026-04-08 00:00:00" {
		t.Fatalf("ReportTime = %q, want normalized datetime", decoded.KafkaData.ReportTime)
	}
	if decoded.KafkaData.Offset != 123 {
		t.Fatalf("Offset = %d, want 123", decoded.KafkaData.Offset)
	}
	if decoded.KafkaData.ConsumptionTime == "" {
		t.Fatalf("ConsumptionTime should be injected")
	}

	decoded.MarkFn()
	if !marked {
		t.Fatalf("MarkFn should remain usable after decode")
	}
}

func TestMessageDecoderWrapMarksOnDecodeFailure(t *testing.T) {
	decoder := newMessageDecoder(jsoniter.ConfigCompatibleWithStandardLibrary)
	handler := &spyDecodedHandler{}
	marked := false
	now := time.Now()

	decoder.Wrap(handler)(model.InputMessage{
		Value:     []byte(`not-json`),
		Timestamp: &now,
	}, func() {
		marked = true
	})

	if !marked {
		t.Fatalf("markFn should be called when decode fails")
	}
	if handler.called {
		t.Fatalf("handler should not be called when decode fails")
	}
}

type spyDecodedHandler struct {
	called bool
}

func (h *spyDecodedHandler) HandleDecoded(decoded DecodedMessage) {
	h.called = true
}
