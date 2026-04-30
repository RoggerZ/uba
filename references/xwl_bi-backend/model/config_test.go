package model

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestBatchConfigNormalize(t *testing.T) {
	got := (BatchConfig{}).Normalize()
	if got.BufferSize != 1000 {
		t.Fatalf("BufferSize = %d, want 1000", got.BufferSize)
	}
	if got.FlushInterval != 2 {
		t.Fatalf("FlushInterval = %d, want 2", got.FlushInterval)
	}
}

func TestSinkerConfigNormalize(t *testing.T) {
	got := (SinkerConfig{}).Normalize()
	if got.ReportAcceptStatus.BufferSize != 1000 || got.ReportData2CK.BufferSize != 1000 || got.RealTimeWarehousing.BufferSize != 1000 {
		t.Fatalf("unexpected batch defaults: %+v", got)
	}
	if got.ReportAcceptStatus.FlushInterval != 2 || got.ReportData2CK.FlushInterval != 2 || got.RealTimeWarehousing.FlushInterval != 2 {
		t.Fatalf("unexpected flush defaults: %+v", got)
	}
	if got.AdminHttpHost != "127.0.0.1" || got.AdminHttpPort != 8094 {
		t.Fatalf("unexpected admin defaults: %+v", got)
	}
	if got.DiagnosticDefaultTTLSeconds != 180 || got.DiagnosticMaxTTLSeconds != 3600 {
		t.Fatalf("unexpected diagnostic ttl defaults: %+v", got)
	}
}

func TestProtectionConfigUnmarshalKeepsEnabledDefaultWhenPartiallyConfigured(t *testing.T) {
	var cfg ProtectionConfig
	if err := json.Unmarshal([]byte(`{"mock":{"enabled":true}}`), &cfg); err != nil {
		t.Fatalf("Unmarshal returned error: %v", err)
	}
	got := cfg.Normalize()
	if got.Enabled == nil || !*got.Enabled {
		t.Fatal("ProtectionConfig.Enabled = false, want default true when omitted")
	}
	if !got.Mock.Enabled {
		t.Fatal("ProtectionConfig.Mock.Enabled = false, want true")
	}
}

func TestProtectionConfigUnmarshalAllowsExplicitDisable(t *testing.T) {
	var cfg ProtectionConfig
	if err := json.Unmarshal([]byte(`{"enabled":false,"sampleIntervalSeconds":5}`), &cfg); err != nil {
		t.Fatalf("Unmarshal returned error: %v", err)
	}
	got := cfg.Normalize()
	if got.Enabled == nil || *got.Enabled {
		t.Fatal("ProtectionConfig.Enabled = true, want explicit false")
	}
}

func TestProtectionConfigUnmarshalKeepsMockFields(t *testing.T) {
	var cfg ProtectionConfig
	if err := json.Unmarshal([]byte(`{
		"mock":{
			"enabled":true,
			"reportLag":2500000,
			"pipelinePending":120000,
			"gateInFlight":110000,
			"gateWaitingTasks":220000,
			"reportConsumerQueueUsagePermille":850,
			"reportPersistQueueUsagePermille":700,
			"persistenceErrorCount":12,
			"mysqlStatus":"degraded",
			"mysqlConsecutiveFailures":5
		}
	}`), &cfg); err != nil {
		t.Fatalf("Unmarshal returned error: %v", err)
	}

	got := cfg.Normalize()
	if !got.Mock.Enabled {
		t.Fatal("ProtectionConfig.Mock.Enabled = false, want true")
	}
	if got.Mock.ReportLag == nil || *got.Mock.ReportLag != 2500000 {
		t.Fatalf("ProtectionConfig.Mock.ReportLag = %+v, want 2500000", got.Mock.ReportLag)
	}
	if got.Mock.PipelinePending == nil || *got.Mock.PipelinePending != 120000 {
		t.Fatalf("ProtectionConfig.Mock.PipelinePending = %+v, want 120000", got.Mock.PipelinePending)
	}
	if got.Mock.GateInFlight == nil || *got.Mock.GateInFlight != 110000 {
		t.Fatalf("ProtectionConfig.Mock.GateInFlight = %+v, want 110000", got.Mock.GateInFlight)
	}
	if got.Mock.GateWaitingTasks == nil || *got.Mock.GateWaitingTasks != 220000 {
		t.Fatalf("ProtectionConfig.Mock.GateWaitingTasks = %+v, want 220000", got.Mock.GateWaitingTasks)
	}
	if got.Mock.ReportConsumerQueueUsagePermille == nil || *got.Mock.ReportConsumerQueueUsagePermille != 850 {
		t.Fatalf("ProtectionConfig.Mock.ReportConsumerQueueUsagePermille = %+v, want 850", got.Mock.ReportConsumerQueueUsagePermille)
	}
	if got.Mock.ReportPersistQueueUsagePermille == nil || *got.Mock.ReportPersistQueueUsagePermille != 700 {
		t.Fatalf("ProtectionConfig.Mock.ReportPersistQueueUsagePermille = %+v, want 700", got.Mock.ReportPersistQueueUsagePermille)
	}
	if got.Mock.PersistenceErrorCount == nil || *got.Mock.PersistenceErrorCount != 12 {
		t.Fatalf("ProtectionConfig.Mock.PersistenceErrorCount = %+v, want 12", got.Mock.PersistenceErrorCount)
	}
	if got.Mock.MySQLStatus == nil || *got.Mock.MySQLStatus != "degraded" {
		t.Fatalf("ProtectionConfig.Mock.MySQLStatus = %+v, want degraded", got.Mock.MySQLStatus)
	}
	if got.Mock.MySQLConsecutiveFailures == nil || *got.Mock.MySQLConsecutiveFailures != 5 {
		t.Fatalf("ProtectionConfig.Mock.MySQLConsecutiveFailures = %+v, want 5", got.Mock.MySQLConsecutiveFailures)
	}
}

func TestSinkerConfigUnmarshalKeepsWorkerPoolJSONFields(t *testing.T) {
	var cfg SinkerConfig
	if err := json.Unmarshal([]byte(`{
		"reportConsumerPool":{
			"minWorkers":1,
			"maxWorkers":2,
			"queueSize":64,
			"tuneIntervalSeconds":2,
			"drainTimeoutSeconds":30
		},
		"reportPersistPool":{
			"minWorkers":3,
			"maxWorkers":4,
			"queueSize":128,
			"tuneIntervalSeconds":5,
			"drainTimeoutSeconds":60
		}
	}`), &cfg); err != nil {
		t.Fatalf("Unmarshal returned error: %v", err)
	}

	got := cfg.Normalize()
	if got.ReportConsumerPool.MinWorkers != 1 || got.ReportConsumerPool.MaxWorkers != 2 {
		t.Fatalf("ReportConsumerPool bounds = %+v, want 1~2", got.ReportConsumerPool)
	}
	if got.ReportConsumerPool.QueueSize != 64 || got.ReportConsumerPool.TuneInterval != 2 || got.ReportConsumerPool.DrainTimeout != 30 {
		t.Fatalf("ReportConsumerPool timings = %+v, want queue=64 tune=2 drain=30", got.ReportConsumerPool)
	}
	if got.ReportPersistPool.MinWorkers != 3 || got.ReportPersistPool.MaxWorkers != 4 {
		t.Fatalf("ReportPersistPool bounds = %+v, want 3~4", got.ReportPersistPool)
	}
	if got.ReportPersistPool.QueueSize != 128 || got.ReportPersistPool.TuneInterval != 5 || got.ReportPersistPool.DrainTimeout != 60 {
		t.Fatalf("ReportPersistPool timings = %+v, want queue=128 tune=5 drain=60", got.ReportPersistPool)
	}
}

func TestSinkerConfigNormalizeAppliesConsumerLowPreset(t *testing.T) {
	cfg := SinkerConfig{
		Protection: ProtectionConfig{
			Mock: ProtectionMockConfig{
				Preset: "consumer_low",
			},
		},
	}

	got := cfg.Normalize()
	if got.ReportConsumerPool.MinWorkers != 1 || got.ReportConsumerPool.MaxWorkers != 1 || got.ReportConsumerPool.QueueSize != 64 {
		t.Fatalf("consumer_low reportConsumerPool = %+v", got.ReportConsumerPool)
	}
	if got.ReportPersistPool.MinWorkers != 1 || got.ReportPersistPool.MaxWorkers != 1 || got.ReportPersistPool.QueueSize != 64 {
		t.Fatalf("consumer_low reportPersistPool = %+v", got.ReportPersistPool)
	}
	if got.Protection.Mock.Enabled {
		t.Fatal("consumer_low should not enable mock")
	}
	if err := got.Validate(); err != nil {
		t.Fatalf("consumer_low Validate() returned error: %v", err)
	}
}

func TestSinkerConfigNormalizeAppliesProtectMockPreset(t *testing.T) {
	cfg := SinkerConfig{
		Protection: ProtectionConfig{
			Mock: ProtectionMockConfig{
				Preset:  "protect_mock",
				Enabled: true,
			},
		},
	}

	got := cfg.Normalize()
	if !got.Protection.Mock.Enabled {
		t.Fatal("protect_mock should enable mock")
	}
	if got.Protection.Mock.PipelinePending == nil || *got.Protection.Mock.PipelinePending != 120000 {
		t.Fatalf("protect_mock pipelinePending = %+v", got.Protection.Mock.PipelinePending)
	}
	if got.ReportConsumerPool.MinWorkers != 1 || got.ReportPersistPool.MinWorkers != 1 {
		t.Fatalf("protect_mock worker pools = %+v %+v", got.ReportConsumerPool, got.ReportPersistPool)
	}
	if err := got.Validate(); err != nil {
		t.Fatalf("protect_mock Validate() returned error: %v", err)
	}
}

func TestSinkerConfigNormalizeAppliesConsumerLowSoftBacklogPreset(t *testing.T) {
	cfg := SinkerConfig{
		Protection: ProtectionConfig{
			Mock: ProtectionMockConfig{
				Preset: "consumer_low_soft_backlog",
			},
		},
	}

	got := cfg.Normalize()
	if got.Protection.SoftThresholds.OrderedCommitPendingCount != 500 {
		t.Fatalf("consumer_low_soft_backlog orderedCommitPendingCount = %d", got.Protection.SoftThresholds.OrderedCommitPendingCount)
	}
	if got.Protection.SoftThresholds.GateInFlightMessages != 500 {
		t.Fatalf("consumer_low_soft_backlog gateInFlightMessages = %d", got.Protection.SoftThresholds.GateInFlightMessages)
	}
	if got.Protection.SoftThresholds.GateWaitingTasks != 1000 {
		t.Fatalf("consumer_low_soft_backlog gateWaitingTasks = %d", got.Protection.SoftThresholds.GateWaitingTasks)
	}
	if got.Protection.HardThresholds.OrderedCommitPendingCount != 100000 {
		t.Fatalf("consumer_low_soft_backlog hard orderedCommitPendingCount = %d", got.Protection.HardThresholds.OrderedCommitPendingCount)
	}
	if got.Protection.Mock.Enabled {
		t.Fatal("consumer_low_soft_backlog should not enable mock")
	}
	if err := got.Validate(); err != nil {
		t.Fatalf("consumer_low_soft_backlog Validate() returned error: %v", err)
	}
}

func TestSinkerConfigValidateRejectsUnknownPreset(t *testing.T) {
	cfg := SinkerConfig{
		Protection: ProtectionConfig{
			Mock: ProtectionMockConfig{
				Preset: "unknown",
			},
		},
	}

	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate() = nil, want unknown preset error")
	}
}

func TestSinkerConfigValidateRejectsConflictingConsumerLowPreset(t *testing.T) {
	cfg := SinkerConfig{
		ReportConsumerPool: DynamicWorkerPoolConfigJSON{
			MinWorkers: 2,
		},
		Protection: ProtectionConfig{
			Mock: ProtectionMockConfig{
				Preset: "consumer_low",
			},
		},
	}

	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate() = nil, want consumer_low conflict error")
	}
}

func TestSinkerConfigValidateRejectsConflictingSoftBacklogPreset(t *testing.T) {
	cfg := SinkerConfig{
		Protection: ProtectionConfig{
			SoftThresholds: ProtectionThresholds{
				OrderedCommitPendingCount: 900,
			},
			Mock: ProtectionMockConfig{
				Preset: "consumer_low_soft_backlog",
			},
		},
	}

	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate() = nil, want consumer_low_soft_backlog conflict error")
	}
}

func TestSinkerConfigValidateRejectsSoftBacklogPresetWhenMockEnabled(t *testing.T) {
	cfg := SinkerConfig{
		Protection: ProtectionConfig{
			Mock: ProtectionMockConfig{
				Preset:  "consumer_low_soft_backlog",
				Enabled: true,
			},
		},
	}

	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate() = nil, want consumer_low_soft_backlog mock enabled conflict error")
	}
}

func TestSinkerConfigValidateRejectsConflictingProtectMockPreset(t *testing.T) {
	cfg := SinkerConfig{
		Protection: ProtectionConfig{
			Mock: ProtectionMockConfig{
				Preset:    "protect_mock",
				Enabled:   true,
				ReportLag: int64Ptr(123),
			},
		},
	}

	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate() = nil, want protect_mock conflict error")
	}
}

func TestSinkerConfigValidateRejectsProtectMockPresetWhenMockDisabled(t *testing.T) {
	cfg := SinkerConfig{
		Protection: ProtectionConfig{
			Mock: ProtectionMockConfig{
				Preset:  "protect_mock",
				Enabled: false,
			},
		},
	}

	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate() = nil, want protect_mock enabled requirement error")
	}
}

func TestProtectionConfigTemplatesKeepThresholdBlocks(t *testing.T) {
	for _, relativePath := range []string{
		"../config/config.json",
		"../scripts/config/config.json",
	} {
		raw, err := os.ReadFile(filepath.Clean(relativePath))
		if err != nil {
			t.Fatalf("ReadFile(%q) returned error: %v", relativePath, err)
		}

		var payload map[string]any
		if err := json.Unmarshal(raw, &payload); err != nil {
			t.Fatalf("Unmarshal(%q) returned error: %v", relativePath, err)
		}

		sinkerPayload := mustNestedMap(t, payload, relativePath, "sinker")
		protectionPayload := mustNestedMap(t, sinkerPayload, relativePath, "protection")
		mustNestedMap(t, protectionPayload, relativePath, "softThresholds")
		mustNestedMap(t, protectionPayload, relativePath, "hardThresholds")
		mustNestedMap(t, protectionPayload, relativePath, "mock")
		mustNestedMap(t, sinkerPayload, relativePath, "reportPersistPool")

		commPayload := mustNestedMap(t, payload, relativePath, "comm")
		mysqlPayload := mustNestedMap(t, commPayload, relativePath, "mysql")
		mustNestedMap(t, mysqlPayload, relativePath, "healthCheck")
	}
}

func TestDBHealthConfigUnmarshalKeepsEnabledDefaultWhenPartiallyConfigured(t *testing.T) {
	var cfg DBHealthConfig
	if err := json.Unmarshal([]byte(`{"pingIntervalSeconds":15}`), &cfg); err != nil {
		t.Fatalf("Unmarshal returned error: %v", err)
	}
	got := cfg.Normalize()
	if got.Enabled == nil || !*got.Enabled {
		t.Fatal("DBHealthConfig.Enabled = false, want default true when omitted")
	}
}

func mustNestedMap(t *testing.T, payload map[string]any, filePath string, key string) map[string]any {
	t.Helper()

	raw, ok := payload[key]
	if !ok {
		t.Fatalf("%s missing key %q", filePath, key)
	}

	result, ok := raw.(map[string]any)
	if !ok {
		t.Fatalf("%s key %q is not an object: %#v", filePath, key, raw)
	}
	return result
}
