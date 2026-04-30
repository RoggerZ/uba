package runner

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/1340691923/xwl_bi/model"
	"github.com/1340691923/xwl_bi/platform-basic-libs/util"
)

func TestNewSinkerAdminServerRequiresTokenForRemoteBinding(t *testing.T) {
	_, err := newSinkerAdminServer(model.SinkerConfig{
		AdminHttpHost: "0.0.0.0",
		AdminToken:    "",
	})
	if err == nil {
		t.Fatal("expected remote admin server without token to fail")
	}
}

func TestHandleEnableDiagnosticRequiresTokenForRemoteRequest(t *testing.T) {
	util.ResetSinkerDiagnosticSessionForTest()
	server := &sinkerAdminServer{
		config: model.SinkerConfig{
			AdminHttpHost: "0.0.0.0",
			AdminToken:    "secret",
		}.Normalize(),
	}

	request := httptest.NewRequest(http.MethodPost, "/admin/diagnostic/enable", bytes.NewReader([]byte(`{}`)))
	request.RemoteAddr = "10.0.0.2:12345"
	recorder := httptest.NewRecorder()

	server.handleEnableDiagnostic(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", recorder.Code)
	}
}

func TestHandleEnableDiagnosticAllowsLoopbackWithoutToken(t *testing.T) {
	util.ResetSinkerDiagnosticSessionForTest()
	server := &sinkerAdminServer{
		config: model.SinkerConfig{
			AdminHttpHost: "127.0.0.1",
		}.Normalize(),
	}

	request := httptest.NewRequest(http.MethodPost, "/admin/diagnostic/enable", bytes.NewReader([]byte(`{"durationSeconds":5,"traceOffset":123,"reportHandlerStageTimingThreshold":"5s"}`)))
	request.RemoteAddr = "127.0.0.1:12345"
	recorder := httptest.NewRecorder()

	server.handleEnableDiagnostic(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}

	session := util.CurrentSinkerDiagnosticSession()
	if !session.Enabled || !session.TraceOffsetEnabled || session.TraceOffset != 123 {
		t.Fatalf("unexpected diagnostic session: %+v", session)
	}
	if session.ReportHandlerStageTimingThreshold != 5*time.Second {
		t.Fatalf("unexpected report handler threshold: %+v", session)
	}
}

func TestHandleEnableDiagnosticRejectsInvalidReportHandlerThreshold(t *testing.T) {
	util.ResetSinkerDiagnosticSessionForTest()
	server := &sinkerAdminServer{
		config: model.SinkerConfig{
			AdminHttpHost: "127.0.0.1",
		}.Normalize(),
	}

	request := httptest.NewRequest(http.MethodPost, "/admin/diagnostic/enable", bytes.NewReader([]byte(`{"reportHandlerStageTimingThreshold":"not-a-duration"}`)))
	request.RemoteAddr = "127.0.0.1:12345"
	recorder := httptest.NewRecorder()

	server.handleEnableDiagnostic(recorder, request)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", recorder.Code)
	}
}

func TestBuildDiagnosticStatusResponse(t *testing.T) {
	now := time.Date(2026, 4, 15, 12, 0, 0, 0, time.Local)
	expiresAt := now.Add(3 * time.Minute)
	response := buildDiagnosticStatusResponse(util.SinkerDiagnosticSession{
		Enabled:                           true,
		TraceOffsetEnabled:                true,
		TraceOffset:                       321,
		ReportHandlerStageTimingThreshold: 5 * time.Second,
		RateLogInterval:                   5 * time.Second,
		ExpiresAt:                         expiresAt,
		Source:                            "test",
	}, now)

	if !response.Enabled || !response.TraceOffsetEnabled || response.TraceOffset != 321 {
		t.Fatalf("unexpected response: %+v", response)
	}
	if response.RemainingSeconds != int((3 * time.Minute).Seconds()) {
		t.Fatalf("remainingSeconds = %d, want %d", response.RemainingSeconds, int((3 * time.Minute).Seconds()))
	}
	if response.ReportHandlerStageTimingThreshold != "5s" {
		t.Fatalf("reportHandlerStageTimingThreshold = %q, want 5s", response.ReportHandlerStageTimingThreshold)
	}
	if response.DefaultReportHandlerStageTimingThreshold != reportHandlerStageTimingSlowThresholdDefault.String() {
		t.Fatalf("defaultReportHandlerStageTimingThreshold = %q, want %q", response.DefaultReportHandlerStageTimingThreshold, reportHandlerStageTimingSlowThresholdDefault.String())
	}
}

func TestHandleDiagnosticStatusReturnsJSON(t *testing.T) {
	util.ResetSinkerDiagnosticSessionForTest()
	reportHandlerThreshold := 1500 * time.Millisecond
	util.EnableSinkerDiagnosticSession(2*time.Minute, nil, &reportHandlerThreshold, nil, "test", time.Now())
	server := &sinkerAdminServer{
		config: model.SinkerConfig{
			AdminHttpHost: "127.0.0.1",
		}.Normalize(),
	}

	request := httptest.NewRequest(http.MethodGet, "/admin/diagnostic/status", nil)
	request.RemoteAddr = "127.0.0.1:12345"
	recorder := httptest.NewRecorder()

	server.handleDiagnosticStatus(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}

	var payload diagnosticStatusResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("status body is not valid json: %v", err)
	}
	if !payload.Enabled {
		t.Fatalf("unexpected payload: %+v", payload)
	}
	if payload.ReportHandlerStageTimingThreshold != "1.5s" {
		t.Fatalf("unexpected payload threshold: %+v", payload)
	}
	if payload.DefaultReportHandlerStageTimingThreshold != reportHandlerStageTimingSlowThresholdDefault.String() {
		t.Fatalf("unexpected payload default threshold: %+v", payload)
	}
}
