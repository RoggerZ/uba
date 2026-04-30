package util

import (
	"testing"
	"time"
)

func TestIsTruthyEnvValue(t *testing.T) {
	testCases := []struct {
		name     string
		rawValue string
		want     bool
	}{
		{name: "one", rawValue: "1", want: true},
		{name: "true", rawValue: "true", want: true},
		{name: "yes", rawValue: "yes", want: true},
		{name: "on", rawValue: "on", want: true},
		{name: "trimmed uppercase", rawValue: "  TRUE  ", want: true},
		{name: "zero", rawValue: "0", want: false},
		{name: "false", rawValue: "false", want: false},
		{name: "empty", rawValue: "", want: false},
	}

	for _, testCase := range testCases {
		if got := IsTruthyEnvValue(testCase.rawValue); got != testCase.want {
			t.Fatalf("%s: IsTruthyEnvValue(%q) = %v, want %v", testCase.name, testCase.rawValue, got, testCase.want)
		}
	}
}

func TestCurrentSinkerDiagnosticSessionFromEnv(t *testing.T) {
	t.Setenv(SinkerDiagnosticLogEnv, "1")
	t.Setenv(SinkerTraceOffsetEnv, "123")
	ResetSinkerDiagnosticSessionForTest()

	session := CurrentSinkerDiagnosticSession()
	if !session.Enabled {
		t.Fatal("expected diagnostic session to be enabled from env")
	}
	if !session.TraceOffsetEnabled || session.TraceOffset != 123 {
		t.Fatalf("unexpected trace offset state: %+v", session)
	}
}

func TestEnableAndDisableSinkerDiagnosticSession(t *testing.T) {
	ResetSinkerDiagnosticSessionForTest()

	now := time.Date(2026, 4, 15, 10, 0, 0, 0, time.Local)
	traceOffset := int64(321)
	reportHandlerThreshold := 5 * time.Second
	session := EnableSinkerDiagnosticSession(3*time.Minute, &traceOffset, &reportHandlerThreshold, nil, "test", now)
	if !session.Enabled || !session.TraceOffsetEnabled || session.TraceOffset != traceOffset {
		t.Fatalf("unexpected enabled session: %+v", session)
	}
	if session.ReportHandlerStageTimingThreshold != reportHandlerThreshold {
		t.Fatalf("report handler threshold = %v, want %v", session.ReportHandlerStageTimingThreshold, reportHandlerThreshold)
	}
	if session.ExpiresAt != now.Add(3*time.Minute) {
		t.Fatalf("expiresAt = %v, want %v", session.ExpiresAt, now.Add(3*time.Minute))
	}

	session = DisableSinkerDiagnosticSession("test_disable", now.Add(time.Minute))
	if session.Enabled || session.TraceOffsetEnabled {
		t.Fatalf("unexpected disabled session: %+v", session)
	}
	if session.ReportHandlerStageTimingThreshold != 0 {
		t.Fatalf("report handler threshold should be cleared after disable, got %+v", session)
	}
}

func TestSinkerDiagnosticSessionExpires(t *testing.T) {
	ResetSinkerDiagnosticSessionForTest()

	now := time.Date(2026, 4, 15, 10, 0, 0, 0, time.Local)
	reportHandlerThreshold := 3 * time.Second
	EnableSinkerDiagnosticSession(time.Second, nil, &reportHandlerThreshold, nil, "test", now)

	sinkerDiagnosticSessionStore.mutex.Lock()
	expireSinkerDiagnosticSessionLocked(now.Add(2 * time.Second))
	session := sinkerDiagnosticSessionStore.state
	sinkerDiagnosticSessionStore.mutex.Unlock()

	if session.Enabled {
		t.Fatalf("expected expired session to be disabled, got %+v", session)
	}
	if session.Source != "expired" {
		t.Fatalf("source = %q, want expired", session.Source)
	}
	if session.ReportHandlerStageTimingThreshold != 0 {
		t.Fatalf("report handler threshold should be cleared after expire, got %+v", session)
	}
}
