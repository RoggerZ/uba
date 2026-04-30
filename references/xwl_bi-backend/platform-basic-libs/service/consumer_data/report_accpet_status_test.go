package consumer_data

import (
	"errors"
	"testing"
	"time"

	"github.com/1340691923/xwl_bi/model"
)

func TestBuildReportAcceptStatusInsertArgs(t *testing.T) {
	row := datePartitionRow[*ReportAcceptStatusData]{
		item: &ReportAcceptStatusData{
			PartDate:       "2026-04-08",
			ReportType:     "report",
			DataName:       "event_a",
			ErrorReason:    "bad field",
			ErrorHandling:  "drop",
			ReportData:     `{"a":1}`,
			XwlKafkaOffset: 123,
			TableId:        456,
			Status:         SuccessStatus,
		},
		partitionTime: time.Date(2026, 4, 8, 10, 11, 12, 0, time.Local),
	}

	args, err := buildReportAcceptStatusInsertArgs(row)
	if err != nil {
		t.Fatalf("buildReportAcceptStatusInsertArgs returned error: %v", err)
	}
	if got := args[0].(int32); got != int32(SuccessStatus) {
		t.Fatalf("status type/value = %#v, want %d", args[0], SuccessStatus)
	}
	if got := args[1].(time.Time).Format("2006-01-02 15:04:05"); got != "2026-04-08 10:11:12" {
		t.Fatalf("ingest_time = %s, want 2026-04-08 10:11:12", got)
	}
	if got := args[2].(time.Time).Format("2006-01-02 15:04:05"); got != "2026-04-08 00:00:00" {
		t.Fatalf("part_date = %s, want 2026-04-08 00:00:00", got)
	}
	if got := args[3].(int64); got != int64(456) {
		t.Fatalf("table_id = %d, want 456", got)
	}
}

func TestReportAcceptStatusSwapAndRestoreBuffer(t *testing.T) {
	store := NewReportAcceptStatus(model.BatchConfig{BufferSize: 8, FlushInterval: 0})
	store.buffer = []*ReportAcceptStatusData{
		{DataName: "A", Status: SuccessStatus},
		{DataName: "B", Status: FailStatus},
	}

	batch := store.swapBuffer()
	if len(batch) != 2 {
		t.Fatalf("swapBuffer len = %d, want 2", len(batch))
	}
	if got := store.getBufferLength(); got != 0 {
		t.Fatalf("buffer len after swap = %d, want 0", got)
	}

	store.buffer = append(store.buffer, &ReportAcceptStatusData{DataName: "C", Status: SuccessStatus})
	store.restoreBuffer(batch)

	if got := store.getBufferLength(); got != 3 {
		t.Fatalf("buffer len after restore = %d, want 3", got)
	}
	if store.buffer[0].DataName != "A" || store.buffer[1].DataName != "B" || store.buffer[2].DataName != "C" {
		t.Fatalf("unexpected buffer order after restore: %+v", store.buffer)
	}
}

func TestNewReportAcceptStatusBucketPlanner(t *testing.T) {
	t.Run("运行时查询成功时按 30% 计算目标上限", func(t *testing.T) {
		planner := newDatePartitionBucketPlanner(100, func(data *ReportAcceptStatusData) string {
			return data.PartDate
		})
		if planner.runtimePartitionLimit != 100 {
			t.Fatalf("runtimePartitionLimit = %d, want 100", planner.runtimePartitionLimit)
		}
		if planner.effectivePartitionTarget != 50 {
			t.Fatalf("effectivePartitionTarget = %d, want 50", planner.effectivePartitionTarget)
		}
		if planner.usedFallbackTarget {
			t.Fatal("usedFallbackTarget = true, want false")
		}
	})

	t.Run("运行时查询失败时回退到固定目标上限", func(t *testing.T) {
		planner := newDatePartitionBucketPlanner(0, func(data *ReportAcceptStatusData) string {
			return data.PartDate
		})
		if planner.effectivePartitionTarget != DatePartitionFallbackTarget() {
			t.Fatalf("effectivePartitionTarget = %d, want %d", planner.effectivePartitionTarget, DatePartitionFallbackTarget())
		}
		if !planner.usedFallbackTarget {
			t.Fatal("usedFallbackTarget = false, want true")
		}
	})
}

func TestReportAcceptStatusBuildBuckets(t *testing.T) {
	planner := newDatePartitionBucketPlannerWithLayout(10, func(data *ReportAcceptStatusData) string {
		return data.IngestTime
	}, datePartitionKeyLayoutMonth)
	planner.effectivePartitionTarget = 2

	buckets, err := planner.BuildBuckets([]*ReportAcceptStatusData{
		{IngestTime: "2026-04-08 10:00:00", DataName: "A"},
		{IngestTime: "2026-04-08 11:00:00", DataName: "B"},
		{IngestTime: "2026-05-09 09:00:00", DataName: "C"},
		{IngestTime: "2026-06-10 12:00:00", DataName: "D"},
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
	if buckets[1].partitionKeys[0] != "202606" {
		t.Fatalf("unexpected bucket[1] partition keys: %+v", buckets[1].partitionKeys)
	}
}

func TestReportAcceptStatusFlushBatchReturnsRemainingBucketsOnPartialFailure(t *testing.T) {
	store := NewReportAcceptStatusWithPartitionLimit(model.BatchConfig{BufferSize: 8, FlushInterval: 0}, 10)
	store.planner.effectivePartitionTarget = 1

	callCount := 0
	store.flushBucketFn = func(bucket datePartitionBucket[*ReportAcceptStatusData], bucketIndex, bucketCount int) (datePartitionBucketTiming, error) {
		callCount++
		if callCount == 2 {
			return datePartitionBucketTiming{totalCost: 200 * time.Millisecond}, errors.New("bucket failed")
		}
		return datePartitionBucketTiming{totalCost: 50 * time.Millisecond}, nil
	}

	batch := []*ReportAcceptStatusData{
		{IngestTime: "2026-04-08 10:00:00", DataName: "A"},
		{IngestTime: "2026-05-09 10:00:00", DataName: "B"},
		{IngestTime: "2026-06-10 10:00:00", DataName: "C"},
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
	if remaining[0].DataName != "B" {
		t.Fatalf("unexpected remaining order: %+v", remaining)
	}
}

func TestReportAcceptStatusFlushBatchReturnsAllFailedBuckets(t *testing.T) {
	store := NewReportAcceptStatusWithPartitionLimit(model.BatchConfig{BufferSize: 8, FlushInterval: 0}, 10)
	store.planner.effectivePartitionTarget = 1

	callCount := 0
	store.flushBucketFn = func(bucket datePartitionBucket[*ReportAcceptStatusData], bucketIndex, bucketCount int) (datePartitionBucketTiming, error) {
		callCount++
		if bucketIndex == 2 || bucketIndex == 4 {
			return datePartitionBucketTiming{totalCost: 200 * time.Millisecond}, errors.New("bucket failed")
		}
		return datePartitionBucketTiming{totalCost: 50 * time.Millisecond}, nil
	}

	batch := []*ReportAcceptStatusData{
		{IngestTime: "2026-04-08 10:00:00", DataName: "A"},
		{IngestTime: "2026-05-09 10:00:00", DataName: "B"},
		{IngestTime: "2026-06-10 10:00:00", DataName: "C"},
		{IngestTime: "2026-07-11 10:00:00", DataName: "D"},
	}

	remaining, err := store.flushBatch(batch)
	if err == nil {
		t.Fatal("expected flushBatch to return error")
	}
	if callCount != 4 {
		t.Fatalf("flushBucketFn callCount = %d, want 4", callCount)
	}
	if len(remaining) != 2 {
		t.Fatalf("remaining len = %d, want 2", len(remaining))
	}
	if remaining[0].DataName != "B" || remaining[1].DataName != "D" {
		t.Fatalf("unexpected remaining order: %+v", remaining)
	}
}
