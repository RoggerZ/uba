package consumer_data

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/1340691923/xwl_bi/engine/db"
	"github.com/1340691923/xwl_bi/engine/logs"
	"github.com/1340691923/xwl_bi/model"
	model2 "github.com/1340691923/xwl_bi/platform-basic-libs/sinker/model"
	parser "github.com/1340691923/xwl_bi/platform-basic-libs/sinker/parse"
	"github.com/1340691923/xwl_bi/platform-basic-libs/util"
	clickhouse "github.com/ClickHouse/clickhouse-go/v2"
	"go.uber.org/zap"
)

var tableColumnMap sync.Map

const reportData2CKMaxFlushRowsPerChunk = 10000
const reportData2CKNativeBatchTimeout = 30 * time.Second

type reportData2CKNativeBatch interface {
	Abort() error
	Append(v ...any) error
	Send() error
}

type reportData2CKPrepareBatchFunc func(ctx context.Context, query string) (reportData2CKNativeBatch, error)

// ReportData2CK 负责把已经通过前置校验且完成补列的明细写入 ClickHouse。
type ReportData2CK struct {
	*batchCore[FastjsonMetricData]
	flushTableRowsFn func(tableName string, dims []*model2.ColumnWithType, rows [][]interface{}) error
	prepareBatchFn   reportData2CKPrepareBatchFunc
	maxFlushRows     int
}

type FastjsonMetricData struct {
	FastjsonMetric *parser.FastjsonMetric
	TableName      string
	AfterFlush     func()
}

type reportData2CKBatchEntry struct {
	index int
	item  FastjsonMetricData
	row   []interface{}
}

func LoadTableColumns(tableName string) ([]*model2.ColumnWithType, bool) {
	dimsValue, ok := tableColumnMap.Load(tableName)
	if !ok {
		return nil, false
	}

	dims, ok := dimsValue.([]*model2.ColumnWithType)
	if !ok {
		return nil, false
	}
	return dims, true
}

func StoreTableColumns(tableName string, dims []*model2.ColumnWithType) {
	tableColumnMap.Store(tableName, dims)
}

func ResetTableColumnsForTest() {
	tableColumnMap = sync.Map{}
}

func NewReportData2CK(config model.BatchConfig) *ReportData2CK {
	logs.Logger.Info("NewReportData2CK", zap.Int("batchSize", config.BufferSize), zap.Int("flushInterval", config.FlushInterval))

	store := &ReportData2CK{}
	store.batchCore = newBatchCore[FastjsonMetricData](config, "ReportData2CK RegularFlushing", store.flushBatch)
	store.flushTableRowsFn = store.flushTableRows
	store.prepareBatchFn = store.prepareBatch
	store.maxFlushRows = reportData2CKMaxFlushRowsPerChunk
	if config.FlushInterval > 0 {
		store.RegularFlushing()
	}

	return store
}

func (c *ReportData2CK) SetAsyncExecutor(executor AsyncExecutor) {
	c.batchCore.setAsyncExecutor(executor)
}

func (c *ReportData2CK) prepareBatch(ctx context.Context, query string) (reportData2CKNativeBatch, error) {
	if db.ClickHouseNative == nil {
		return nil, fmt.Errorf("reportdata2ck: clickhouse native connection is not initialized")
	}
	return db.ClickHouseNative.PrepareBatch(ctx, query)
}

func (c *ReportData2CK) restoreUnflushedTables(batch []FastjsonMetricData, flushedTables map[string]struct{}) {
	if len(batch) == 0 {
		return
	}

	pending := make([]FastjsonMetricData, 0, len(batch))
	for _, data := range batch {
		if _, ok := flushedTables[data.TableName]; ok {
			continue
		}
		pending = append(pending, data)
	}
	c.restoreBuffer(pending)
}

// flushBatch 先按表分组，再把每张表拆成多个小块依次刷库。
//
// 这样做的目的有两个：
// 1. 避免单次 native batch 体量失控。
// 2. 分片成功后立刻触发对应消息的 AfterFlush，缩短 gate 等待时间。
func (c *ReportData2CK) flushBatch(batch []FastjsonMetricData) (remaining []FastjsonMetricData, err error) {
	startNow := time.Now()
	flushLen := len(batch)

	tableEntries := make(map[string][]reportData2CKBatchEntry, flushLen)
	for idx, obj := range batch {
		dims, ok := LoadTableColumns(obj.TableName)
		if !ok {
			return batch, fmt.Errorf("reportdata2ck: missing table schema cache for %s", obj.TableName)
		}

		rowArr := make([]interface{}, 0, len(dims))
		for _, dim := range dims {
			val := parser.GetValueByType(obj.FastjsonMetric, dim)
			if dim.Name == "xwl_part_date" {
				if t, ok := val.(time.Time); ok && t.Equal(parser.Epoch) {
					val = time.Now()
				}
			}
			rowArr = append(rowArr, val)
		}

		tableEntries[obj.TableName] = append(tableEntries[obj.TableName], reportData2CKBatchEntry{
			index: idx,
			item:  obj,
			row:   rowArr,
		})
	}

	tableNames := make([]string, 0, len(tableEntries))
	for tableName := range tableEntries {
		tableNames = append(tableNames, tableName)
	}
	sort.Strings(tableNames)

	flushedIndexes := make(map[int]struct{}, flushLen)
	for _, tableName := range tableNames {
		dims, _ := LoadTableColumns(tableName)

		for _, chunk := range c.splitEntries(tableEntries[tableName]) {
			if err := c.flushTableRowsFn(tableName, dims, collectReportData2CKRows(chunk)); err != nil {
				return c.unflushedBatchByIndex(batch, flushedIndexes), err
			}

			c.notifyFlushedEntries(chunk)
			for _, entry := range chunk {
				flushedIndexes[entry.index] = struct{}{}
			}
		}
	}

	logs.Logger.Info("CK入库成功，", zap.String("所花时间", time.Since(startNow).String()), zap.Int("数据长度为", flushLen))
	return nil, nil
}

func (c *ReportData2CK) splitEntries(entries []reportData2CKBatchEntry) [][]reportData2CKBatchEntry {
	if len(entries) == 0 {
		return nil
	}

	limit := c.maxFlushRows
	if limit <= 0 {
		limit = reportData2CKMaxFlushRowsPerChunk
	}

	chunks := make([][]reportData2CKBatchEntry, 0, (len(entries)+limit-1)/limit)
	for start := 0; start < len(entries); start += limit {
		end := start + limit
		if end > len(entries) {
			end = len(entries)
		}
		chunks = append(chunks, entries[start:end])
	}
	return chunks
}

func collectReportData2CKRows(entries []reportData2CKBatchEntry) [][]interface{} {
	rows := make([][]interface{}, 0, len(entries))
	for _, entry := range entries {
		rows = append(rows, entry.row)
	}
	return rows
}

func (c *ReportData2CK) notifyFlushedEntries(entries []reportData2CKBatchEntry) {
	for _, entry := range entries {
		if entry.item.AfterFlush != nil {
			entry.item.AfterFlush()
		}
	}
}

func (c *ReportData2CK) unflushedBatchByIndex(batch []FastjsonMetricData, flushedIndexes map[int]struct{}) []FastjsonMetricData {
	if len(flushedIndexes) == 0 {
		return batch
	}

	remaining := make([]FastjsonMetricData, 0, len(batch)-len(flushedIndexes))
	for idx, item := range batch {
		if _, ok := flushedIndexes[idx]; ok {
			continue
		}
		remaining = append(remaining, item)
	}
	return remaining
}

func (c *ReportData2CK) flushTableRows(tableName string, dims []*model2.ColumnWithType, rows [][]interface{}) error {
	var (
		err              error
		startNow         = time.Now()
		insertSQL        = buildNativeInsertStatement(tableName, dims)
		batchCtx, cancel = context.WithTimeout(context.Background(), reportData2CKNativeBatchTimeout)
	)
	defer cancel()

	batch, err := c.prepareBatchFn(batchCtx, insertSQL)
	if err != nil {
		classification := util.RecordPersistenceError("clickhouse_prepare_failed", err)
		logs.Logger.Error(
			"reportdata2ck: prepare native batch failed",
			zap.String("tableName", tableName),
			zap.String("error_class", classification.ErrorClass),
			zap.Bool("count_toward_circuit_break", classification.CountTowardCircuitBreak),
			zap.Error(err),
		)
		c.logNativeBatchSummary(tableName, len(rows), time.Since(startNow), err)
		return err
	}

	for _, row := range rows {
		if err := batch.Append(row...); err != nil {
			_ = batch.Abort()
			classification := util.RecordPersistenceError("clickhouse_append_failed", err)
			logs.Logger.Error(
				"reportdata2ck: append native batch failed",
				zap.String("tableName", tableName),
				zap.String("error_class", classification.ErrorClass),
				zap.Bool("count_toward_circuit_break", classification.CountTowardCircuitBreak),
				zap.Error(err),
			)
			c.logNativeBatchSummary(tableName, len(rows), time.Since(startNow), err)
			return err
		}
	}

	if err := batch.Send(); err != nil {
		_ = batch.Abort()
		classification := util.RecordPersistenceError("clickhouse_send_failed", err)
		if exception, ok := err.(*clickhouse.Exception); ok {
			logs.Logger.Error(
				"reportdata2ck: send native batch failed",
				zap.String("tableName", tableName),
				zap.String("error_class", classification.ErrorClass),
				zap.Bool("count_toward_circuit_break", classification.CountTowardCircuitBreak),
				zap.Int32("code", exception.Code),
				zap.String("message", exception.Message),
			)
		} else {
			logs.Logger.Error(
				"reportdata2ck: send native batch failed",
				zap.String("tableName", tableName),
				zap.String("error_class", classification.ErrorClass),
				zap.Bool("count_toward_circuit_break", classification.CountTowardCircuitBreak),
				zap.Error(err),
			)
		}
		c.logNativeBatchSummary(tableName, len(rows), time.Since(startNow), err)
		return err
	}

	c.logNativeBatchSummary(tableName, len(rows), time.Since(startNow), nil)
	return nil
}

func buildNativeInsertStatement(tableName string, dims []*model2.ColumnWithType) string {
	quotedDims := make([]string, len(dims))
	for i, dim := range dims {
		quotedDims[i] = "`" + dim.Name + "`"
	}

	var builder strings.Builder
	builder.WriteString("INSERT INTO ")
	builder.WriteString(tableName)
	builder.WriteString(" (")
	builder.WriteString(strings.Join(quotedDims, ","))
	builder.WriteString(")")
	return builder.String()
}

func (c *ReportData2CK) logNativeBatchSummary(tableName string, rowCount int, cost time.Duration, err error) {
	if err == nil && rowCount < 1000 && cost < 500*time.Millisecond {
		return
	}

	logs.Logger.Info(
		"report data2ck native batch summary",
		zap.String("table_name", tableName),
		zap.Int("row_count", rowCount),
		zap.Duration("cost", cost),
		zap.Error(err),
	)
}
