package consumer_data

import (
	"errors"
	"testing"
	"time"

	"github.com/1340691923/xwl_bi/model"
)

func TestBuildRealTimeWarehousingInsertArgs(t *testing.T) {
	row := datePartitionRow[*RealTimeWarehousingData]{
		item: &RealTimeWarehousingData{
			Appid:      99,
			EventName:  "pay",
			EventTime:  "2026-04-08",
			Data:       []byte(`{"amount":1}`),
			IngestTime: "2026-04-08 10:11:12",
		},
		partitionTime: time.Date(2026, 4, 8, 10, 11, 12, 0, time.Local),
	}

	args, err := buildRealTimeWarehousingInsertArgs(row)
	if err != nil {
		t.Fatalf("buildRealTimeWarehousingInsertArgs returned error: %v", err)
	}
	if got := args[0].(int64); got != 99 {
		t.Fatalf("appid = %d, want 99", got)
	}
	if got := args[1].(time.Time).Format("2006-01-02 15:04:05"); got != "2026-04-08 10:11:12" {
		t.Fatalf("ingest_time = %s, want 2026-04-08 10:11:12", got)
	}
	if got := args[2].(time.Time).Format("2006-01-02 15:04:05"); got != "2026-04-08 00:00:00" {
		t.Fatalf("event_time = %s, want 2026-04-08 00:00:00", got)
	}
	if got := args[3].(string); got != "pay" {
		t.Fatalf("event_name = %s, want pay", got)
	}
	if got := args[4].(string); got != `{"amount":1}` {
		t.Fatalf("report_data = %s, want %s", got, `{"amount":1}`)
	}
}

func TestRealTimeWarehousingSwapAndRestoreBuffer(t *testing.T) {
	store := NewRealTimeWarehousing(model.BatchConfig{BufferSize: 8, FlushInterval: 0})
	store.buffer = []*RealTimeWarehousingData{
		{Appid: 1, EventName: "A"},
		{Appid: 2, EventName: "B"},
	}

	batch := store.swapBuffer()
	if len(batch) != 2 {
		t.Fatalf("swapBuffer len = %d, want 2", len(batch))
	}
	if got := store.getBufferLength(); got != 0 {
		t.Fatalf("buffer len after swap = %d, want 0", got)
	}

	store.buffer = append(store.buffer, &RealTimeWarehousingData{Appid: 3, EventName: "C"})
	store.restoreBuffer(batch)

	if got := store.getBufferLength(); got != 3 {
		t.Fatalf("buffer len after restore = %d, want 3", got)
	}
	if store.buffer[0].EventName != "A" || store.buffer[1].EventName != "B" || store.buffer[2].EventName != "C" {
		t.Fatalf("unexpected buffer order after restore: %+v", store.buffer)
	}
}

func TestNewRealTimeWarehousingWithPartitionLimit(t *testing.T) {
	store := NewRealTimeWarehousingWithPartitionLimit(model.BatchConfig{BufferSize: 8, FlushInterval: 0}, 100)
	if store.planner.runtimePartitionLimit != 100 {
		t.Fatalf("runtimePartitionLimit = %d, want 100", store.planner.runtimePartitionLimit)
	}
	if store.planner.effectivePartitionTarget != 50 {
		t.Fatalf("effectivePartitionTarget = %d, want 50", store.planner.effectivePartitionTarget)
	}
}

func TestRealTimeWarehousingBuildBuckets(t *testing.T) {
	planner := newDatePartitionBucketPlannerWithLayout(10, func(data *RealTimeWarehousingData) string {
		return data.IngestTime
	}, datePartitionKeyLayoutMonth)
	planner.effectivePartitionTarget = 2

	buckets, err := planner.BuildBuckets([]*RealTimeWarehousingData{
		{IngestTime: "2026-04-08 10:00:00", EventName: "A"},
		{IngestTime: "2026-04-08 11:00:00", EventName: "B"},
		{IngestTime: "2026-05-09 09:00:00", EventName: "C"},
		{IngestTime: "2026-06-10 12:00:00", EventName: "D"},
	})
	if err != nil {
		t.Fatalf("BuildBuckets returned error: %v", err)
	}
	if len(buckets) != 2 {
		t.Fatalf("bucket count = %d, want 2", len(buckets))
	}
	if len(buckets[0].partitionKeys) != 2 {
		t.Fatalf("bucket[0] partition count = %d, want 2", len(buckets[0].partitionKeys))
	}
	if len(buckets[1].partitionKeys) != 1 {
		t.Fatalf("bucket[1] partition count = %d, want 1", len(buckets[1].partitionKeys))
	}
	if buckets[0].partitionKeys[0] != "202604" || buckets[0].partitionKeys[1] != "202605" {
		t.Fatalf("unexpected bucket[0] partition keys: %+v", buckets[0].partitionKeys)
	}
}

func TestRealTimeWarehousingFlushBatchReturnsFailedBucketsOnly(t *testing.T) {
	store := NewRealTimeWarehousingWithPartitionLimit(model.BatchConfig{BufferSize: 8, FlushInterval: 0}, 10)
	store.planner.effectivePartitionTarget = 1

	callCount := 0
	store.flushBucketFn = func(bucket datePartitionBucket[*RealTimeWarehousingData], bucketIndex, bucketCount int) (datePartitionBucketTiming, error) {
		callCount++
		if bucketIndex == 2 {
			return datePartitionBucketTiming{totalCost: 200 * time.Millisecond}, errors.New("bucket failed")
		}
		return datePartitionBucketTiming{totalCost: 50 * time.Millisecond}, nil
	}

	batch := []*RealTimeWarehousingData{
		{IngestTime: "2026-04-08 10:00:00", EventName: "A"},
		{IngestTime: "2026-05-09 10:00:00", EventName: "B"},
		{IngestTime: "2026-06-10 10:00:00", EventName: "C"},
	}

	remaining, err := store.flushBatch(batch)
	if err == nil {
		t.Fatal("expected flushBatch to return error")
	}
	if callCount != 3 {
		t.Fatalf("flushBucketFn callCount = %d, want 3", callCount)
	}
	if len(remaining) != 1 {
		t.Fatalf("remaining len = %d, want 1", len(remaining))
	}
	if remaining[0].EventName != "B" {
		t.Fatalf("unexpected remaining items: %+v", remaining)
	}
}
