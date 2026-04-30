package consumer_data

import (
	"errors"
	"testing"

	"github.com/1340691923/xwl_bi/platform-basic-libs/util"
)

func TestPartsPressureGuardEntersCooldownWhenActivePartsHigh(t *testing.T) {
	guard := NewPartsPressureGuard(TableNameRealTimeWarehousing)
	guard.UpdateSnapshot(PartsPressureSnapshot{
		Table:           TableNameRealTimeWarehousing,
		ActiveParts:     90,
		MaxPartsInTotal: 100,
	})

	err := guard.BeforeFlush(128)
	if !IsDeferredFlushError(err) {
		t.Fatalf("BeforeFlush err = %v, want deferred flush error", err)
	}
}

func TestPartsPressureGuardIgnoresLowPressure(t *testing.T) {
	guard := NewPartsPressureGuard(TableNameAcceptanceStatus)
	guard.UpdateSnapshot(PartsPressureSnapshot{
		Table:           TableNameAcceptanceStatus,
		ActiveParts:     20,
		MaxPartsInTotal: 100,
	})

	if err := guard.BeforeFlush(64); err != nil {
		t.Fatalf("BeforeFlush err = %v, want nil", err)
	}
}

func TestPartsPressureGuardDetectsTooManyPartsError(t *testing.T) {
	err := errors.New("code: 252, message: Too many parts (100000)")
	if !IsTooManyPartsError(err) {
		t.Fatal("expected Too many parts error to be detected")
	}
}

func TestPartsPressureGuardObserveFlushErrorRecordsPersistenceSignal(t *testing.T) {
	util.ResetPersistenceErrorTrackerForTest()

	guard := NewPartsPressureGuard(TableNameRealTimeWarehousing)
	guard.ObserveFlushError(errors.New("code: 252, message: Too many parts (100000)"), 256)

	snapshot := util.GetPersistenceErrorSnapshot()
	if snapshot.CountLastMinute != 1 {
		t.Fatalf("CountLastMinute = %d, want 1", snapshot.CountLastMinute)
	}
	if snapshot.LastClass != "too_many_parts" {
		t.Fatalf("LastClass = %q, want too_many_parts", snapshot.LastClass)
	}
}
