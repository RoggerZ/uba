package main

import (
	"strings"
	"testing"
)

func TestFormatProtectCLIOutputIncludesReadableFields(t *testing.T) {
	raw := `{
		"enabled": true,
		"observeOnly": false,
		"state": "hard_paused",
		"softHoldUntil": "0001-01-01T00:00:00Z",
		"hardHoldUntil": "2026-04-17T11:14:42.1775252+08:00",
		"lastTransitionAt": "2026-04-17T11:13:07.1775252+08:00",
		"lastTransitionReason": "hard thresholds reached",
		"currentSoftSignals": ["ordered_commit_pending", "gate_in_flight"],
		"currentHardSignals": ["persistence_errors"],
		"currentRecoveryBlockers": ["hard_hold", "persistence_errors"],
		"reportRate": {
			"group": "reportData2CKGroup2",
			"topic": "test005",
			"currentOffset": 62159519,
			"logEndOffset": 63649577,
			"lag": 1490058,
			"deltaLag": -7170,
			"speedPerSecond": -1434.02,
			"sampledAt": "2026-04-17T11:14:32.1726459+08:00",
			"samplingMode": "normal"
		},
		"realTimeRate": {
			"group": "realTimeDataGroup2",
			"topic": "test005",
			"currentOffset": 63536210,
			"logEndOffset": 63649577,
			"lag": 113367,
			"deltaLag": -7169,
			"speedPerSecond": -1433.82,
			"sampledAt": "2026-04-17T11:14:32.1726459+08:00",
			"samplingMode": "normal"
		},
		"reportPipeline": {
			"committerCount": 1,
			"pendingCount": 0,
			"doneCount": 0,
			"largestPendingGap": 0,
			"oldestPendingOffset": 0,
			"newestCompletedOffset": 62159519,
			"gate": {
				"inFlightMessages": 0,
				"waitingTasks": 0,
				"completedMessages": 581623
			}
		},
		"reportConsumerPool": {
			"Queued": 0,
			"QueueCapacity": 4096,
			"QueueUsageRatio": 0,
			"Running": 0,
			"Capacity": 20,
			"Idle": 20,
			"BusyRatio": 0,
			"MinWorkers": 20,
			"MaxWorkers": 64,
			"SubmittedTotal": 581623,
			"CompletedTotal": 581623,
			"RejectedTotal": 0,
			"Closed": false
		},
		"reportPersistPool": {
			"Queued": 0,
			"QueueCapacity": 1024,
			"QueueUsageRatio": 0,
			"Running": 1,
			"Capacity": 10,
			"Idle": 9,
			"BusyRatio": 0.1,
			"MinWorkers": 10,
			"MaxWorkers": 16,
			"SubmittedTotal": 418,
			"CompletedTotal": 418,
			"RejectedTotal": 0,
			"Closed": false
		},
		"persistenceErrors": {
			"countLastMinute": 12,
			"lastClass": "clickhouse_sql_driver_error",
			"lastError": "read: EOF",
			"lastOccurredAt": "2026-04-17T11:13:14.1344961+08:00"
		},
		"mysqlHealth": {
			"Name": "mysql",
			"DriverName": "mysql",
			"Enabled": true,
			"Status": "healthy",
			"ConsecutiveFailures": 0,
			"LastError": "",
			"LastErrorAt": "0001-01-01T00:00:00Z",
			"LastRecoveredAt": "2026-04-17T11:12:22.0894629+08:00",
			"LastUpdatedAt": "2026-04-17T11:14:32.0897944+08:00"
		},
		"goroutines": 73,
		"heapAllocBytes": 53140176
	}`

	output := formatProtectCLIOutput("status", []byte(raw))
	for _, want := range []string{
		"保护开关: 已开启",
		"当前状态: hard_paused",
		"最近切换: 2026-04-17 11:13:07 +08:00 reason=hard thresholds reached",
		"当前 soft signals: ordered_commit_pending,gate_in_flight",
		"当前 hard signals: persistence_errors",
		"当前恢复阻塞: hard_hold,persistence_errors",
		"ReportData2CK: group=reportData2CKGroup2",
		"RealTime: group=realTimeDataGroup2",
		"持久化错误(近1m): 12",
		"last_class=clickhouse_sql_driver_error",
		"持久化最后错误: read: EOF",
		"MySQL健康: status=healthy",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("formatted output missing %q, got:\n%s", want, output)
		}
	}
}

func TestFormatProtectTimeHandlesZeroAndRFC3339(t *testing.T) {
	if got := formatProtectTime(""); got != "未设置" {
		t.Fatalf("formatProtectTime(empty) = %q, want 未设置", got)
	}
	if got := formatProtectTime("0001-01-01T00:00:00Z"); got != "未设置" {
		t.Fatalf("formatProtectTime(zero) = %q, want 未设置", got)
	}
	if got := formatProtectTime("2026-04-17T11:14:42.1775252+08:00"); !strings.Contains(got, "2026-04-17 11:14:42 +08:00") {
		t.Fatalf("formatProtectTime(parsed) = %q", got)
	}
}
