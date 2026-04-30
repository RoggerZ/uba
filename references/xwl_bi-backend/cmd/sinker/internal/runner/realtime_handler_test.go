package runner

import (
	"errors"
	"testing"

	"github.com/1340691923/xwl_bi/model"
	"github.com/1340691923/xwl_bi/platform-basic-libs/service/consumer_data"
	jsoniter "github.com/json-iterator/go"
)

type fakeRealTimeSink struct {
	items []*consumer_data.RealTimeWarehousingData
	err   error
}

func (f *fakeRealTimeSink) Add(data *consumer_data.RealTimeWarehousingData) error {
	cloned := *data
	f.items = append(f.items, &cloned)
	return f.err
}

func TestRealTimeMessageHandlerMarksEvenWhenSinkFails(t *testing.T) {
	sink := &fakeRealTimeSink{err: errors.New("sink failed")}
	handler := newRealTimeMessageHandler(sink, nil, nil)
	decoder := newMessageDecoder(jsoniter.ConfigCompatibleWithStandardLibrary)

	marked := false
	decoded, err := decoder.Decode(newReportInputMessage(model.KafkaData{
		TableId:    "51",
		EventName:  "AppLaunch",
		ReportTime: "2026-04-08",
		ReqData:    []byte(`{"k":"v"}`),
	}), func() {
		marked = true
	})
	if err != nil {
		t.Fatalf("Decode returned error: %v", err)
	}

	handler.HandleDecoded(decoded)

	if !marked {
		t.Fatalf("markFn should always be called on realtime path")
	}
	if len(sink.items) != 1 {
		t.Fatalf("realtime sink count = %d, want 1", len(sink.items))
	}
	if sink.items[0].EventTime != "2026-04-08 00:00:00" {
		t.Fatalf("EventTime = %q, want normalized datetime", sink.items[0].EventTime)
	}
	if sink.items[0].IngestTime == "" {
		t.Fatal("IngestTime should be set")
	}
}
