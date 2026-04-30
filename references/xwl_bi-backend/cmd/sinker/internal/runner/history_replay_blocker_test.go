package runner

import (
	"testing"
	"time"

	"github.com/1340691923/xwl_bi/platform-basic-libs/service/consumer_data"
)

func TestHistoryReplayBlockerSkipsOlderThanCutoff(t *testing.T) {
	cutoff := time.Date(2026, 1, 13, 0, 0, 0, 0, time.Local)
	blocker := newHistoryReplayBlocker(cutoff)

	if !blocker.ShouldSkip(consumer_data.TableNameRealTimeWarehousing, "2026-01-12 23:59:59") {
		t.Fatal("expected older business time to be skipped")
	}
	if blocker.ShouldSkip(consumer_data.TableNameRealTimeWarehousing, "2026-01-13 00:00:00") {
		t.Fatal("cutoff boundary should not be skipped")
	}
}

func TestHistoryReplayBlockerAggregatesByTable(t *testing.T) {
	cutoff := time.Date(2026, 1, 13, 0, 0, 0, 0, time.Local)
	blocker := newHistoryReplayBlocker(cutoff)

	blocker.ShouldSkip(consumer_data.TableNameAcceptanceStatus, "2025-12-01 10:00:00")
	blocker.ShouldSkip(consumer_data.TableNameAcceptanceStatus, "2025-11-01 10:00:00")

	stat := blocker.stats[consumer_data.TableNameAcceptanceStatus]
	if stat == nil {
		t.Fatal("expected aggregated stat for acceptance status")
	}
	if stat.count != 2 {
		t.Fatalf("stat.count = %d, want 2", stat.count)
	}
}
