package report

import (
	"testing"

	"github.com/1340691923/xwl_bi/model"
	model2 "github.com/1340691923/xwl_bi/platform-basic-libs/sinker/model"
)

func TestPayloadBuilderRegistryBuild(t *testing.T) {
	registry := DefaultPayloadBuilderRegistry()

	userPayload, err := registry.Build(model2.ReportUserProperties, PayloadBuildInput{
		APPID:              "1001",
		TableID:            "51",
		Debug:              "1",
		ReportTime:         "2026-04-10 12:00:00",
		ReportTimeHasClock: true,
		EventName:          "ignored",
		IP:                 "1.1.1.1",
		Body:               []byte(`{"name":"demo"}`),
	})
	if err != nil {
		t.Fatalf("Build user payload returned error: %v", err)
	}
	if userPayload.ReportType != model.UserReportType {
		t.Fatalf("user payload ReportType = %d, want %d", userPayload.ReportType, model.UserReportType)
	}
	if userPayload.EventName != "用户属性" {
		t.Fatalf("user payload EventName = %q, want %q", userPayload.EventName, "用户属性")
	}
	if userPayload.ReportTimeHasClock == nil || !*userPayload.ReportTimeHasClock {
		t.Fatalf("user payload should keep report time clock semantics, got %+v", userPayload.ReportTimeHasClock)
	}

	eventPayload, err := registry.Build(model2.ReportEventProperties, PayloadBuildInput{
		APPID:              "1002",
		TableID:            "52",
		Debug:              "2",
		ReportTime:         "2026-04-10 13:00:00",
		ReportTimeHasClock: false,
		EventName:          "pay_success",
		IP:                 "2.2.2.2",
		Body:               []byte(`{"amount":18}`),
	})
	if err != nil {
		t.Fatalf("Build event payload returned error: %v", err)
	}
	if eventPayload.ReportType != model.EventReportType {
		t.Fatalf("event payload ReportType = %d, want %d", eventPayload.ReportType, model.EventReportType)
	}
	if eventPayload.EventName != "pay_success" {
		t.Fatalf("event payload EventName = %q, want %q", eventPayload.EventName, "pay_success")
	}
	if eventPayload.ReportTimeHasClock == nil || *eventPayload.ReportTimeHasClock {
		t.Fatalf("event payload should keep date-only report time semantics, got %+v", eventPayload.ReportTimeHasClock)
	}
}

func TestPayloadBuilderRegistryBuildInvalidType(t *testing.T) {
	_, err := DefaultPayloadBuilderRegistry().Build("bad-type", PayloadBuildInput{})
	if err == nil {
		t.Fatal("expected invalid typ to return error")
	}
	if err.Error() != ERROR_TABLE[ReportTypeErr] {
		t.Fatalf("error = %q, want %q", err.Error(), ERROR_TABLE[ReportTypeErr])
	}
}

func TestPayloadBuilderRegistryBuildDoesNotLeakState(t *testing.T) {
	registry := DefaultPayloadBuilderRegistry()

	first, err := registry.Build(model2.ReportEventProperties, PayloadBuildInput{
		APPID:      "1001",
		TableID:    "51",
		Debug:      "1",
		ReportTime: "2026-04-10 12:00:00",
		EventName:  "first_event",
		IP:         "1.1.1.1",
		Body:       []byte(`{"step":1}`),
	})
	if err != nil {
		t.Fatalf("first Build returned error: %v", err)
	}

	second, err := registry.Build(model2.ReportUserProperties, PayloadBuildInput{
		APPID:      "1002",
		TableID:    "52",
		Debug:      "2",
		ReportTime: "2026-04-10 13:00:00",
		EventName:  "should_be_ignored",
		IP:         "2.2.2.2",
		Body:       []byte(`{"step":2}`),
	})
	if err != nil {
		t.Fatalf("second Build returned error: %v", err)
	}

	if first.EventName != "first_event" {
		t.Fatalf("first payload EventName changed to %q", first.EventName)
	}
	if second.EventName != "用户属性" {
		t.Fatalf("second payload EventName = %q, want %q", second.EventName, "用户属性")
	}
}
