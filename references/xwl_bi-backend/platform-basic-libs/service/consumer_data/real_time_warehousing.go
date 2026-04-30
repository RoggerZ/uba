package consumer_data

import (
	"time"

	"github.com/1340691923/xwl_bi/engine/db"
	"github.com/1340691923/xwl_bi/engine/logs"
	"github.com/1340691923/xwl_bi/model"
	"github.com/1340691923/xwl_bi/platform-basic-libs/util"
	"go.uber.org/zap"
)

type RealTimeWarehousingData struct {
	Appid      int64
	EventName  string
	IngestTime string
	EventTime  string
	Data       []byte
}

// RealTimeWarehousing 负责把实时消息批量写入 xwl_real_time_warehousing。
// 它只关心“实时表写入”这一种职责，公共批量行为统一交给 batchCore。
type RealTimeWarehousing struct {
	*batchCore[*RealTimeWarehousingData]
	planner       *datePartitionBucketPlanner[*RealTimeWarehousingData]
	flushBucketFn func(bucket datePartitionBucket[*RealTimeWarehousingData], bucketIndex, bucketCount int) (datePartitionBucketTiming, error)
	partsGuard    *PartsPressureGuard
}

func NewRealTimeWarehousing(config model.BatchConfig) *RealTimeWarehousing {
	return NewRealTimeWarehousingWithPartitionLimit(config, 0)
}

// NewRealTimeWarehousingWithPartitionLimit 创建带日期分桶规划器的实时单表批量器。
//
// xwl_real_time_warehousing 现在按 ingest_time 的“月份”分区。
// 所以这里仍然保留共享 planner，但它关注的是：
// 1. 不能把过多不同 ingest_time 月份混成一个大 INSERT block。
// 2. 否则同样可能触发 `Too many partitions for single INSERT block`。
func NewRealTimeWarehousingWithPartitionLimit(config model.BatchConfig, runtimePartitionLimit int) *RealTimeWarehousing {
	logs.Logger.Info("NewRealTimeWarehousing", zap.Int("batchSize", config.BufferSize), zap.Int("flushInterval", config.FlushInterval))
	store := &RealTimeWarehousing{
		planner: newDatePartitionBucketPlannerWithLayout(runtimePartitionLimit, func(data *RealTimeWarehousingData) string {
			return data.IngestTime
		}, datePartitionKeyLayoutMonth),
	}
	store.batchCore = newBatchCore[*RealTimeWarehousingData](config, "RealTimeWarehousing RegularFlushing", store.flushBatch)
	store.enableTraceDebug()
	store.flushBucketFn = store.flushBucket

	logs.Logger.Info("RealTimeWarehousing partition planner", store.planner.logFields()...)

	if config.FlushInterval > 0 {
		store.RegularFlushing()
	}

	return store
}

func (w *RealTimeWarehousing) SetPartsPressureGuard(guard *PartsPressureGuard) {
	w.partsGuard = guard
	if guard != nil {
		w.batchCore.setBeforeFlush(func(bufferLength int) error {
			return guard.BeforeFlush(bufferLength)
		})
	}
}

func (w *RealTimeWarehousing) BypassPartsPressureGuard(enabled bool) {
	if w.partsGuard == nil {
		return
	}
	w.partsGuard.SetBypass(enabled)
}

func (w *RealTimeWarehousing) flushBatch(batch []*RealTimeWarehousingData) (remaining []*RealTimeWarehousingData, err error) {
	startNow := time.Now()
	buckets, err := w.planner.BuildBuckets(batch)
	if err != nil {
		return batch, err
	}
	if len(buckets) == 0 {
		return nil, nil
	}

	type failedBucket struct {
		bucket datePartitionBucket[*RealTimeWarehousingData]
		timing datePartitionBucketTiming
		err    error
		index  int
	}

	failedBuckets := make([]failedBucket, 0)
	var firstErr error
	for idx, bucket := range buckets {
		timing, bucketErr := w.flushBucketFn(bucket, idx+1, len(buckets))
		if bucketErr != nil {
			if firstErr == nil {
				firstErr = bucketErr
			}
			failedBuckets = append(failedBuckets, failedBucket{
				bucket: bucket,
				timing: timing,
				err:    bucketErr,
				index:  idx + 1,
			})
			continue
		}
		w.logBucketTiming(bucket, idx+1, len(buckets), timing, 0, nil)
	}

	totalCost := time.Since(startNow)
	if len(failedBuckets) > 0 {
		failedOnly := make([]datePartitionBucket[*RealTimeWarehousingData], 0, len(failedBuckets))
		for _, failed := range failedBuckets {
			failedOnly = append(failedOnly, failed.bucket)
		}
		remaining = flattenDatePartitionBuckets(failedOnly)
		for _, failed := range failedBuckets {
			w.logBucketTiming(failed.bucket, failed.index, len(buckets), failed.timing, len(remaining), failed.err)
		}

		logs.Logger.Warn(
			"入库实时数据部分失败",
			zap.String("所花时间", totalCost.String()),
			zap.Int("数据长度为", len(batch)),
			zap.Int("桶数量", len(buckets)),
			zap.Int("失败桶数量", len(failedBuckets)),
			zap.Int("剩余数据长度", len(remaining)),
			zap.Error(firstErr),
		)
		if w.partsGuard != nil {
			w.partsGuard.ObserveFlushError(firstErr, len(batch))
		}
		return remaining, firstErr
	}

	logs.Logger.Info("入库实时数据成功", zap.String("所花时间", totalCost.String()), zap.Int("数据长度为", len(batch)), zap.Int("桶数量", len(buckets)))
	return nil, nil
}

// buildRealTimeWarehousingInsertArgs 负责把实时表行转换成 ClickHouse v2 标准驱动可稳定接收的参数类型。
func buildRealTimeWarehousingInsertArgs(row datePartitionRow[*RealTimeWarehousingData]) ([]interface{}, error) {
	eventTime, err := ParseDateTimeString(row.item.EventTime)
	if err != nil {
		return nil, err
	}

	return []interface{}{
		int64(row.item.Appid),
		row.partitionTime,
		eventTime,
		row.item.EventName,
		util.Bytes2str(row.item.Data),
	}, nil
}

func (w *RealTimeWarehousing) flushBucket(bucket datePartitionBucket[*RealTimeWarehousingData], bucketIndex, bucketCount int) (timing datePartitionBucketTiming, err error) {
	startNow := time.Now()

	stageBegin := time.Now()
	tx, err := db.ClickHouseSqlx.Begin()
	timing.beginCost = time.Since(stageBegin)
	if err != nil {
		util.RecordPersistenceError("clickhouse_sql_driver_error", err)
		return timing, err
	}
	defer tx.Rollback()

	stageBegin = time.Now()
	stmt, err := tx.Prepare("INSERT INTO xwl_real_time_warehousing (table_id,ingest_time,event_time,event_name,report_data) VALUES (?,?,?,?,?)")
	timing.prepareCost = time.Since(stageBegin)
	if err != nil {
		util.RecordPersistenceError("clickhouse_prepare_failed", err)
		return timing, err
	}
	defer stmt.Close()

	timing.normalizeCost = bucket.normalizeCost
	for _, row := range bucket.rows {
		stageBegin = time.Now()
		args, err := buildRealTimeWarehousingInsertArgs(row)
		if err != nil {
			timing.execCost += time.Since(stageBegin)
			timing.totalCost = time.Since(startNow)
			logs.Logger.Error("实时数据参数转换失败", zap.Error(err))
			return timing, err
		}
		if _, err := stmt.Exec(args...); err != nil {
			timing.execCost += time.Since(stageBegin)
			timing.totalCost = time.Since(startNow)
			util.RecordPersistenceError("clickhouse_sql_driver_error", err)
			logs.Logger.Error("入库实时数据出现错误", zap.Error(err))
			return timing, err
		}
		timing.execCost += time.Since(stageBegin)
	}

	stageBegin = time.Now()
	if err := tx.Commit(); err != nil {
		timing.commitCost = time.Since(stageBegin)
		timing.totalCost = time.Since(startNow)
		util.RecordPersistenceError("clickhouse_sql_driver_error", err)
		logs.Logger.Error("入库实时数据出现错误", zap.Error(err))
		return timing, err
	}
	timing.commitCost = time.Since(stageBegin)
	timing.totalCost = time.Since(startNow)
	return timing, nil
}

func (w *RealTimeWarehousing) logBucketTiming(bucket datePartitionBucket[*RealTimeWarehousingData], bucketIndex, bucketCount int, timing datePartitionBucketTiming, remainingSize int, err error) {
	if err == nil && timing.totalCost < 100*time.Millisecond {
		return
	}

	bucketKey := ""
	if len(bucket.partitionKeys) == 1 {
		bucketKey = bucket.partitionKeys[0]
	}

	logs.Logger.Debug(
		"real time warehousing flush timing",
		append(
			w.planner.logFields(),
			zap.String("table_name", TableNameRealTimeWarehousing),
			zap.String("bucket_key", bucketKey),
			zap.Int("bucket_count", bucketCount),
			zap.Int("bucket_index", bucketIndex),
			zap.Int("bucket_partition_count", len(bucket.partitionKeys)),
			zap.Strings("bucket_partition_keys", bucket.partitionKeys),
			zap.Int("bucket_size", len(bucket.rows)),
			zap.Int("remaining_size", remainingSize),
			zap.Duration("total_cost", timing.totalCost),
			zap.Duration("begin_cost", timing.beginCost),
			zap.Duration("prepare_cost", timing.prepareCost),
			zap.Duration("normalize_cost", timing.normalizeCost),
			zap.Duration("exec_cost", timing.execCost),
			zap.Duration("commit_cost", timing.commitCost),
			zap.Error(err),
		)...,
	)
}
