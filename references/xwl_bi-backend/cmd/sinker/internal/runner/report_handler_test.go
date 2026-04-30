package runner

import (
	"errors"
	"testing"
	"time"

	"github.com/1340691923/xwl_bi/model"
	"github.com/1340691923/xwl_bi/platform-basic-libs/service/consumer_data"
	parser "github.com/1340691923/xwl_bi/platform-basic-libs/sinker/parse"
	"github.com/1340691923/xwl_bi/platform-basic-libs/util"
	jsoniter "github.com/json-iterator/go"
)

type fakeStatusSink struct {
	items []*consumer_data.ReportAcceptStatusData
	err   error
	order *[]string
}

func (f *fakeStatusSink) Add(data *consumer_data.ReportAcceptStatusData) error {
	cloned := *data
	f.items = append(f.items, &cloned)
	if f.order != nil {
		*f.order = append(*f.order, "status")
	}
	return f.err
}

type fakeMetricSink struct {
	items []consumer_data.FastjsonMetricData
	err   error
	order *[]string
}

func (f *fakeMetricSink) Add(data consumer_data.FastjsonMetricData) error {
	f.items = append(f.items, data)
	if f.order != nil {
		*f.order = append(*f.order, "metric")
	}
	return f.err
}

type fakeSchemaSynchronizer struct {
	err    error
	called int
	order  *[]string
}

func (f *fakeSchemaSynchronizer) EnsureColumns(model.KafkaData, string, *parser.FastjsonMetric, func(consumer_data.ReportAcceptStatusData)) error {
	f.called++
	if f.order != nil {
		*f.order = append(*f.order, "schema")
	}
	return f.err
}

type fakeMetaRecorder struct {
	err    error
	called int
	order  *[]string
}

func (f *fakeMetaRecorder) Record(model.KafkaData) error {
	f.called++
	if f.order != nil {
		*f.order = append(*f.order, "meta")
	}
	return f.err
}

type fakeGeoPayloadEnricher struct {
	err    error
	called int
}

func (f *fakeGeoPayloadEnricher) Enrich(reqData []byte, _ string) ([]byte, error) {
	f.called++
	return reqData, f.err
}

func TestReportMessageHandlerMissingDistinctIDReturnsHandledAndWritesFailStatus(t *testing.T) {
	statusSink := &fakeStatusSink{}
	metricSink := &fakeMetricSink{}
	schema := &fakeSchemaSynchronizer{}
	meta := &fakeMetaRecorder{}
	handler := newReportMessageHandler(
		&fakeGeoPayloadEnricher{},
		schema,
		meta,
		statusSink,
		metricSink,
		nil,
		nil,
	)

	result := handler.ProcessDecoded(mustDecode(t, model.KafkaData{
		TableId:    "51",
		ReportType: model.EventReportType,
		EventName:  "AppLaunch",
		ReportTime: "2026-04-08 16:14:53",
		ReqData:    []byte(`{"xwl_client_time":"2026-04-08 16:14:53"}`),
	}), nil)

	if !result.handled {
		t.Fatalf("missing distinct id should be treated as handled business failure")
	}
	if len(statusSink.items) != 1 {
		t.Fatalf("status count = %d, want 1", len(statusSink.items))
	}
	if statusSink.items[0].Status != consumer_data.FailStatus {
		t.Fatalf("status = %d, want fail", statusSink.items[0].Status)
	}
	if len(metricSink.items) != 0 {
		t.Fatalf("metric sink should stay empty, got %d", len(metricSink.items))
	}
	if schema.called != 0 || meta.called != 0 {
		t.Fatalf("schema/meta should not be called, got schema=%d meta=%d", schema.called, meta.called)
	}
}

func TestReportMessageHandlerTimeSkewReturnsHandledAndWritesFailStatus(t *testing.T) {
	statusSink := &fakeStatusSink{}
	reportTimeHasClock := true
	handler := newReportMessageHandler(
		&fakeGeoPayloadEnricher{},
		&fakeSchemaSynchronizer{},
		&fakeMetaRecorder{},
		statusSink,
		&fakeMetricSink{},
		nil,
		nil,
	)

	result := handler.ProcessDecoded(mustDecode(t, model.KafkaData{
		TableId:            "51",
		ReportType:         model.EventReportType,
		EventName:          "AppLaunch",
		ReportTime:         "2026-04-08 16:44:53",
		ReportTimeHasClock: &reportTimeHasClock,
		ReqData:            []byte(`{"xwl_client_time":"2026-04-08 16:14:53","xwl_distinct_id":"u-1"}`),
	}), nil)

	if !result.handled {
		t.Fatalf("time skew should be treated as handled business failure")
	}
	if len(statusSink.items) != 1 {
		t.Fatalf("status count = %d, want 1", len(statusSink.items))
	}
	if statusSink.items[0].Status != consumer_data.FailStatus {
		t.Fatalf("status = %d, want fail", statusSink.items[0].Status)
	}
}

func TestExtractClientContextKeepsDateOnlyReportTimeAsNoClockWhenProducerProvidedFlag(t *testing.T) {
	handler := newReportMessageHandler(
		&fakeGeoPayloadEnricher{},
		&fakeSchemaSynchronizer{},
		&fakeMetaRecorder{},
		&fakeStatusSink{},
		&fakeMetricSink{},
		nil,
		nil,
	)
	reportTimeHasClock := false
	kafkaData := model.KafkaData{
		ReportTime:         "2026-04-08 00:00:00",
		ReportTimeHasClock: &reportTimeHasClock,
		ReqData:            []byte(`{"xwl_client_time":"2026-04-08 16:14:53","xwl_distinct_id":"u-1"}`),
	}

	_, _, _, hasClock, err := handler.extractClientContext(&kafkaData)
	if err != nil {
		t.Fatalf("extractClientContext returned error: %v", err)
	}
	if hasClock {
		t.Fatal("date-only report_time should stay no-clock when producer provided explicit flag")
	}
	if kafkaData.ReportTime != "2026-04-08 00:00:00" {
		t.Fatalf("ReportTime = %q, want %q", kafkaData.ReportTime, "2026-04-08 00:00:00")
	}
}

func TestReportMessageHandlerCallsSchemaBeforeMetricAndReturnsHandledOnSuccess(t *testing.T) {
	order := []string{}
	statusSink := &fakeStatusSink{order: &order}
	metricSink := &fakeMetricSink{order: &order}
	schema := &fakeSchemaSynchronizer{order: &order}
	meta := &fakeMetaRecorder{order: &order}
	handler := newReportMessageHandler(
		&fakeGeoPayloadEnricher{},
		schema,
		meta,
		statusSink,
		metricSink,
		nil,
		nil,
	)

	result := handler.ProcessDecoded(mustDecode(t, model.KafkaData{
		TableId:    "51",
		ReportType: model.EventReportType,
		EventName:  "AppLaunch",
		ReportTime: "2026-04-08 16:14:53",
		ReqData:    []byte(`{"xwl_client_time":"2026-04-08 16:14:53","xwl_distinct_id":"u-1"}`),
	}), nil)

	if !result.handled {
		t.Fatalf("successful message should be handled")
	}
	if len(metricSink.items) != 1 {
		t.Fatalf("metric count = %d, want 1", len(metricSink.items))
	}
	if len(statusSink.items) != 1 || statusSink.items[0].Status != consumer_data.SuccessStatus {
		t.Fatalf("expected one success status, got %+v", statusSink.items)
	}

	schemaIndex := indexOf(order, "schema")
	metricIndex := indexOf(order, "metric")
	if schemaIndex == -1 || metricIndex == -1 {
		t.Fatalf("unexpected call order: %+v", order)
	}
	if schemaIndex > metricIndex {
		t.Fatalf("schema should run before metric add, order: %+v", order)
	}
}

func TestReportMessageHandlerReturnsRetryableFailureWhenMetricBatchRejectsMessage(t *testing.T) {
	handler := newReportMessageHandler(
		&fakeGeoPayloadEnricher{},
		&fakeSchemaSynchronizer{},
		&fakeMetaRecorder{},
		&fakeStatusSink{},
		&fakeMetricSink{err: errors.New("metric sink failed")},
		nil,
		nil,
	)

	result := handler.ProcessDecoded(mustDecode(t, model.KafkaData{
		TableId:    "51",
		ReportType: model.EventReportType,
		EventName:  "AppLaunch",
		ReportTime: "2026-04-08 16:14:53",
		ReqData:    []byte(`{"xwl_client_time":"2026-04-08 16:14:53","xwl_distinct_id":"u-1"}`),
	}), nil)

	if result.handled {
		t.Fatalf("metric sink failure should stay retryable")
	}
}

func TestCurrentReportHandlerStageTimingSlowThresholdUsesDefaultWhenDiagnosticDisabled(t *testing.T) {
	util.ResetSinkerDiagnosticSessionForTest()

	if got := currentReportHandlerStageTimingSlowThreshold(); got != reportHandlerStageTimingSlowThresholdDefault {
		t.Fatalf("currentReportHandlerStageTimingSlowThreshold = %v, want %v", got, reportHandlerStageTimingSlowThresholdDefault)
	}
}

func TestReportProcessTimingShouldLogUsesDiagnosticOverrideThreshold(t *testing.T) {
	util.ResetSinkerDiagnosticSessionForTest()

	timing := newReportProcessTiming()
	totalCost := 1500 * time.Millisecond
	if timing.shouldLog(totalCost, "success", true) {
		t.Fatalf("default threshold should not log totalCost=%v", totalCost)
	}

	override := time.Second
	util.EnableSinkerDiagnosticSession(2*time.Minute, nil, &override, nil, "test", time.Now())
	if !timing.shouldLog(totalCost, "success", true) {
		t.Fatalf("override threshold should log totalCost=%v", totalCost)
	}
}

func mustDecode(t *testing.T, kafkaData model.KafkaData) DecodedMessage {
	t.Helper()

	decoder := newMessageDecoder(jsoniter.ConfigCompatibleWithStandardLibrary)
	decoded, err := decoder.Decode(newReportInputMessage(kafkaData), func() {})
	if err != nil {
		t.Fatalf("Decode returned error: %v", err)
	}
	return decoded
}

func newReportInputMessage(kafkaData model.KafkaData) model.InputMessage {
	json := jsoniter.ConfigCompatibleWithStandardLibrary
	value, err := json.Marshal(kafkaData)
	if err != nil {
		panic(err)
	}
	now := time.Date(2026, 4, 8, 16, 14, 53, 0, time.Local)
	return model.InputMessage{
		Value:     value,
		Offset:    123,
		Timestamp: &now,
	}
}

func indexOf(items []string, target string) int {
	for idx, item := range items {
		if item == target {
			return idx
		}
	}
	return -1
}
