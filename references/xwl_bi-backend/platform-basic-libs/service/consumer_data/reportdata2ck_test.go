package consumer_data

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/1340691923/xwl_bi/model"
	model2 "github.com/1340691923/xwl_bi/platform-basic-libs/sinker/model"
	"github.com/1340691923/xwl_bi/platform-basic-libs/util"
)

type fakeReportData2CKNativeBatch struct {
	appendErrAt int
	appendCalls int
	appended    [][]any
	sendCalls   int
	sendErr     error
	abortCalls  int
	abortErr    error
	sent        bool
	isSent      bool
	isSentCalls int
	flushCalls  int
}

func (f *fakeReportData2CKNativeBatch) Abort() error {
	f.abortCalls++
	return f.abortErr
}

func (f *fakeReportData2CKNativeBatch) Append(values ...any) error {
	f.appendCalls++
	if f.appendErrAt > 0 && f.appendCalls == f.appendErrAt {
		return errors.New("append failed")
	}

	copied := append([]any(nil), values...)
	f.appended = append(f.appended, copied)
	return nil
}

func (f *fakeReportData2CKNativeBatch) Send() error {
	f.sendCalls++
	f.sent = f.sendErr == nil
	return f.sendErr
}

func TestReportData2CKAddBelowBatchSizeOnlyBuffers(t *testing.T) {
	store := NewReportData2CK(model.BatchConfig{BufferSize: 3, FlushInterval: 0})

	if err := store.Add(FastjsonMetricData{TableName: "xwl_event1"}); err != nil {
		t.Fatalf("Add returned error: %v", err)
	}
	if err := store.Add(FastjsonMetricData{TableName: "xwl_event1"}); err != nil {
		t.Fatalf("Add returned error: %v", err)
	}

	if got := store.getBufferLength(); got != 2 {
		t.Fatalf("buffer len = %d, want 2", got)
	}
}

func TestReportData2CKSwapAndRestoreBuffer(t *testing.T) {
	store := NewReportData2CK(model.BatchConfig{BufferSize: 8, FlushInterval: 0})
	store.buffer = []FastjsonMetricData{
		{TableName: "A"},
		{TableName: "B"},
	}

	batch := store.swapBuffer()
	if len(batch) != 2 {
		t.Fatalf("swapBuffer len = %d, want 2", len(batch))
	}
	if got := store.getBufferLength(); got != 0 {
		t.Fatalf("buffer len after swap = %d, want 0", got)
	}

	store.buffer = append(store.buffer, FastjsonMetricData{TableName: "C"})
	store.restoreBuffer(batch)

	if got := store.getBufferLength(); got != 3 {
		t.Fatalf("buffer len after restore = %d, want 3", got)
	}
	if store.buffer[0].TableName != "A" || store.buffer[1].TableName != "B" || store.buffer[2].TableName != "C" {
		t.Fatalf("unexpected buffer order after restore: %+v", store.buffer)
	}
}

func TestReportData2CKRestoreUnflushedTables(t *testing.T) {
	store := NewReportData2CK(model.BatchConfig{BufferSize: 8, FlushInterval: 0})
	batch := []FastjsonMetricData{
		{TableName: "xwl_event1"},
		{TableName: "xwl_event2"},
		{TableName: "xwl_event3"},
		{TableName: "xwl_event2"},
	}
	flushedTables := map[string]struct{}{
		"xwl_event1": {},
	}

	store.restoreUnflushedTables(batch, flushedTables)

	if got := store.getBufferLength(); got != 3 {
		t.Fatalf("buffer len after restoreUnflushedTables = %d, want 3", got)
	}
	if store.buffer[0].TableName != "xwl_event2" || store.buffer[1].TableName != "xwl_event3" || store.buffer[2].TableName != "xwl_event2" {
		t.Fatalf("unexpected buffer content after restoreUnflushedTables: %+v", store.buffer)
	}
}

func TestBuildNativeInsertStatement(t *testing.T) {
	dims := []*model2.ColumnWithType{
		{Name: "xwl_part_date"},
		{Name: "xwl_distinct_id"},
	}

	got := buildNativeInsertStatement("xwl_event52", dims)
	want := "INSERT INTO xwl_event52 (`xwl_part_date`,`xwl_distinct_id`)"
	if got != want {
		t.Fatalf("buildNativeInsertStatement = %q, want %q", got, want)
	}
}

func TestReportData2CKSplitEntries(t *testing.T) {
	store := NewReportData2CK(model.BatchConfig{BufferSize: 8, FlushInterval: 0})
	store.maxFlushRows = 2

	chunks := store.splitEntries([]reportData2CKBatchEntry{
		{index: 0},
		{index: 1},
		{index: 2},
		{index: 3},
		{index: 4},
	})
	if got := len(chunks); got != 3 {
		t.Fatalf("chunk count = %d, want 3", got)
	}
	if got := len(chunks[0]); got != 2 {
		t.Fatalf("chunk[0] len = %d, want 2", got)
	}
	if got := len(chunks[1]); got != 2 {
		t.Fatalf("chunk[1] len = %d, want 2", got)
	}
	if got := len(chunks[2]); got != 1 {
		t.Fatalf("chunk[2] len = %d, want 1", got)
	}
}

func TestReportData2CKFlushTableRowsUsesNativeBatch(t *testing.T) {
	store := NewReportData2CK(model.BatchConfig{BufferSize: 8, FlushInterval: 0})
	fakeBatch := &fakeReportData2CKNativeBatch{}

	var (
		deadline time.Time
		ok       bool
		query    string
	)
	store.prepareBatchFn = func(ctx context.Context, insertSQL string) (reportData2CKNativeBatch, error) {
		deadline, ok = ctx.Deadline()
		query = insertSQL
		return fakeBatch, nil
	}

	rows := [][]interface{}{
		{1, "a"},
		{2, "b"},
	}
	dims := []*model2.ColumnWithType{
		{Name: "xwl_part_date"},
		{Name: "xwl_distinct_id"},
	}

	if err := store.flushTableRows("xwl_event52", dims, rows); err != nil {
		t.Fatalf("flushTableRows returned error: %v", err)
	}

	wantQuery := "INSERT INTO xwl_event52 (`xwl_part_date`,`xwl_distinct_id`)"
	if query != wantQuery {
		t.Fatalf("prepare batch query = %q, want %q", query, wantQuery)
	}
	if !ok {
		t.Fatal("prepare batch context should carry deadline")
	}
	if remaining := time.Until(deadline); remaining <= 0 || remaining > reportData2CKNativeBatchTimeout {
		t.Fatalf("prepare batch context deadline remaining = %s, want within (0,%s]", remaining, reportData2CKNativeBatchTimeout)
	}
	if fakeBatch.appendCalls != 2 {
		t.Fatalf("append calls = %d, want 2", fakeBatch.appendCalls)
	}
	if fakeBatch.sendCalls != 1 {
		t.Fatalf("send calls = %d, want 1", fakeBatch.sendCalls)
	}
	if fakeBatch.abortCalls != 0 {
		t.Fatalf("abort calls = %d, want 0", fakeBatch.abortCalls)
	}
	wantRows := [][]any{
		{1, "a"},
		{2, "b"},
	}
	if !reflect.DeepEqual(fakeBatch.appended, wantRows) {
		t.Fatalf("appended rows = %#v, want %#v", fakeBatch.appended, wantRows)
	}
}

func TestReportData2CKFlushTableRowsPrepareBatchFailure(t *testing.T) {
	store := NewReportData2CK(model.BatchConfig{BufferSize: 8, FlushInterval: 0})
	store.prepareBatchFn = func(ctx context.Context, insertSQL string) (reportData2CKNativeBatch, error) {
		return nil, errors.New("prepare batch failed")
	}

	err := store.flushTableRows("xwl_event52", []*model2.ColumnWithType{{Name: "a"}}, [][]interface{}{{1}})
	if err == nil || err.Error() != "prepare batch failed" {
		t.Fatalf("flushTableRows error = %v, want prepare batch failed", err)
	}
}

func TestReportData2CKFlushTableRowsAppendFailure(t *testing.T) {
	store := NewReportData2CK(model.BatchConfig{BufferSize: 8, FlushInterval: 0})
	fakeBatch := &fakeReportData2CKNativeBatch{appendErrAt: 2}
	store.prepareBatchFn = func(ctx context.Context, insertSQL string) (reportData2CKNativeBatch, error) {
		return fakeBatch, nil
	}

	err := store.flushTableRows("xwl_event52", []*model2.ColumnWithType{{Name: "a"}}, [][]interface{}{{1}, {2}})
	if err == nil || err.Error() != "append failed" {
		t.Fatalf("flushTableRows error = %v, want append failed", err)
	}
	if fakeBatch.sendCalls != 0 {
		t.Fatalf("send calls = %d, want 0", fakeBatch.sendCalls)
	}
	if fakeBatch.abortCalls != 1 {
		t.Fatalf("abort calls = %d, want 1", fakeBatch.abortCalls)
	}
}

func TestReportData2CKFlushTableRowsSendFailure(t *testing.T) {
	store := NewReportData2CK(model.BatchConfig{BufferSize: 8, FlushInterval: 0})
	fakeBatch := &fakeReportData2CKNativeBatch{sendErr: errors.New("send failed")}
	store.prepareBatchFn = func(ctx context.Context, insertSQL string) (reportData2CKNativeBatch, error) {
		return fakeBatch, nil
	}

	err := store.flushTableRows("xwl_event52", []*model2.ColumnWithType{{Name: "a"}}, [][]interface{}{{1}})
	if err == nil || err.Error() != "send failed" {
		t.Fatalf("flushTableRows error = %v, want send failed", err)
	}
	if fakeBatch.sendCalls != 1 {
		t.Fatalf("send calls = %d, want 1", fakeBatch.sendCalls)
	}
	if fakeBatch.abortCalls != 1 {
		t.Fatalf("abort calls = %d, want 1", fakeBatch.abortCalls)
	}
}

func TestReportData2CKFlushBatchNotifiesAllEntriesOnFullSuccess(t *testing.T) {
	resetReportData2CKTableColumnMap()
	store := NewReportData2CK(model.BatchConfig{BufferSize: 8, FlushInterval: 0})

	var order []string
	store.flushTableRowsFn = func(tableName string, dims []*model2.ColumnWithType, rows [][]interface{}) error {
		order = append(order, tableName)
		return nil
	}

	StoreTableColumns("b_table", []*model2.ColumnWithType{})
	StoreTableColumns("a_table", []*model2.ColumnWithType{})

	counts := make([]int, 4)
	batch := []FastjsonMetricData{
		{TableName: "b_table", AfterFlush: func() { counts[0]++ }},
		{TableName: "a_table", AfterFlush: func() { counts[1]++ }},
		{TableName: "b_table", AfterFlush: func() { counts[2]++ }},
		{TableName: "a_table", AfterFlush: func() { counts[3]++ }},
	}

	remaining, err := store.flushBatch(batch)
	if err != nil {
		t.Fatalf("flushBatch returned error: %v", err)
	}
	if len(remaining) != 0 {
		t.Fatalf("remaining len = %d, want 0", len(remaining))
	}
	if !reflect.DeepEqual(order, []string{"a_table", "b_table"}) {
		t.Fatalf("flush order = %#v, want %#v", order, []string{"a_table", "b_table"})
	}
	for idx, count := range counts {
		if count != 1 {
			t.Fatalf("AfterFlush count[%d] = %d, want 1", idx, count)
		}
	}
}

func TestReportData2CKFlushBatchFailureRestoresOnlyUnflushedTables(t *testing.T) {
	resetReportData2CKTableColumnMap()
	store := NewReportData2CK(model.BatchConfig{BufferSize: 8, FlushInterval: 0})

	var order []string
	store.flushTableRowsFn = func(tableName string, dims []*model2.ColumnWithType, rows [][]interface{}) error {
		order = append(order, tableName)
		if tableName == "b_table" {
			return fmt.Errorf("flush b_table failed")
		}
		return nil
	}

	StoreTableColumns("a_table", []*model2.ColumnWithType{})
	StoreTableColumns("b_table", []*model2.ColumnWithType{})
	StoreTableColumns("c_table", []*model2.ColumnWithType{})

	counts := make([]int, 5)
	batch := []FastjsonMetricData{
		{TableName: "c_table", AfterFlush: func() { counts[0]++ }},
		{TableName: "a_table", AfterFlush: func() { counts[1]++ }},
		{TableName: "b_table", AfterFlush: func() { counts[2]++ }},
		{TableName: "a_table", AfterFlush: func() { counts[3]++ }},
		{TableName: "c_table", AfterFlush: func() { counts[4]++ }},
	}

	remaining, err := store.flushBatch(batch)
	if err == nil {
		t.Fatal("flushBatch should return error on table failure")
	}
	if !reflect.DeepEqual(order, []string{"a_table", "b_table"}) {
		t.Fatalf("flush order = %#v, want %#v", order, []string{"a_table", "b_table"})
	}

	wantRemainingTables := []string{"c_table", "b_table", "c_table"}
	gotRemainingTables := make([]string, 0, len(remaining))
	for _, item := range remaining {
		gotRemainingTables = append(gotRemainingTables, item.TableName)
	}
	if !reflect.DeepEqual(gotRemainingTables, wantRemainingTables) {
		t.Fatalf("remaining tables = %#v, want %#v", gotRemainingTables, wantRemainingTables)
	}

	wantCounts := []int{0, 1, 0, 1, 0}
	if !reflect.DeepEqual(counts, wantCounts) {
		t.Fatalf("AfterFlush counts = %#v, want %#v", counts, wantCounts)
	}
}

func TestReportData2CKFlushBatchChunkFailureReturnsOnlyUnflushedChunk(t *testing.T) {
	resetReportData2CKTableColumnMap()
	store := NewReportData2CK(model.BatchConfig{BufferSize: 8, FlushInterval: 0})
	store.maxFlushRows = 2

	callCount := 0
	store.flushTableRowsFn = func(tableName string, dims []*model2.ColumnWithType, rows [][]interface{}) error {
		callCount++
		if callCount == 2 {
			return fmt.Errorf("flush chunk failed")
		}
		return nil
	}

	StoreTableColumns("a_table", []*model2.ColumnWithType{})

	counts := make([]int, 3)
	batch := []FastjsonMetricData{
		{TableName: "a_table", AfterFlush: func() { counts[0]++ }},
		{TableName: "a_table", AfterFlush: func() { counts[1]++ }},
		{TableName: "a_table", AfterFlush: func() { counts[2]++ }},
	}

	remaining, err := store.flushBatch(batch)
	if err == nil {
		t.Fatal("flushBatch should return error on chunk failure")
	}
	if callCount != 2 {
		t.Fatalf("flushTableRowsFn callCount = %d, want 2", callCount)
	}
	if got := len(remaining); got != 1 {
		t.Fatalf("remaining len = %d, want 1", got)
	}
	if remaining[0].TableName != "a_table" {
		t.Fatalf("unexpected remaining item: %+v", remaining[0])
	}

	wantCounts := []int{1, 1, 0}
	if !reflect.DeepEqual(counts, wantCounts) {
		t.Fatalf("AfterFlush counts = %#v, want %#v", counts, wantCounts)
	}
}

func resetReportData2CKTableColumnMap() {
	ResetTableColumnsForTest()
	util.ResetPersistenceErrorTrackerForTest()
}
