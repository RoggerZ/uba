package consumer_data

import (
	"fmt"
	"time"

	"github.com/1340691923/xwl_bi/engine/db"
	"github.com/1340691923/xwl_bi/engine/logs"
	"github.com/1340691923/xwl_bi/model"
	"github.com/1340691923/xwl_bi/platform-basic-libs/util"
	"go.uber.org/zap"
)

type ReportAcceptStatusData struct {
	IngestTime     string
	PartDate       string
	ReportType     string
	DataName       string
	ErrorReason    string
	ErrorHandling  string
	ReportData     string
	XwlKafkaOffset int64
	TableId        int
	Status         int
	AfterFlush     func()
}

// ReportAcceptStatus 负责把 sinker 的验收结果批量写入 xwl_acceptance_status。
//
// xwl_acceptance_status 不是最终业务明细表，而是一张“验收状态账本”：
// 1. 记录某条消息是处理成功还是失败。
// 2. 记录失败原因、处理方式、原始上报数据和 Kafka offset。
// 3. 供后台排查“为什么这条消息没有进入最终业务表”。
//
// 示例：
// 1. xwl_distinct_id 缺失 -> status=FailStatus，error_reason="xwl_distinct_id 不能为空"
// 2. 通过前置校验并成功进入明细批次 -> status=SuccessStatus
type ReportAcceptStatus struct {
	*batchCore[*ReportAcceptStatusData]
	planner       *datePartitionBucketPlanner[*ReportAcceptStatusData]
	flushBucketFn func(bucket datePartitionBucket[*ReportAcceptStatusData], bucketIndex, bucketCount int) (datePartitionBucketTiming, error)
	partsGuard    *PartsPressureGuard
}

const (
	FailStatus    = 0
	SuccessStatus = 1
)

func NewReportAcceptStatus(config model.BatchConfig) *ReportAcceptStatus {
	return NewReportAcceptStatusWithPartitionLimit(config, 0)
}

// NewReportAcceptStatusWithPartitionLimit 创建带分桶规划器的验收状态批量器。
//
// 参数说明：
//  1. runtimePartitionLimit 来自 sinker runtime 启动时对
//     `system.settings.max_partitions_per_insert_block` 的一次性查询。
//  2. 如果该值大于 0，则按 `floor(limit * 0.5)` 计算实际目标上限。
//  3. 如果该值无效或查询失败，则回退到 30。
//
// 这里特意把“查询 setting”和“按比例得出实际目标上限”拆开，是为了把两层含义说清楚：
// 1. ClickHouse 的环境上限是多少。
// 2. 当前程序实际打算把每个 bucket 控制在多少分区内。
func NewReportAcceptStatusWithPartitionLimit(config model.BatchConfig, runtimePartitionLimit int) *ReportAcceptStatus {
	logs.Logger.Info("NewReportAcceptStatus", zap.Int("batchSize", config.BufferSize), zap.Int("flushInterval", config.FlushInterval))
	store := &ReportAcceptStatus{
		planner: newDatePartitionBucketPlannerWithLayout(runtimePartitionLimit, func(data *ReportAcceptStatusData) string {
			return data.IngestTime
		}, datePartitionKeyLayoutMonth),
	}
	store.batchCore = newBatchCore[*ReportAcceptStatusData](config, "ReportAcceptStatus RegularFlushing", store.flushBatch)
	store.enableTraceDebug()
	store.flushBucketFn = store.flushBucket

	logs.Logger.Info("ReportAcceptStatus partition planner", store.planner.logFields()...)

	if config.FlushInterval > 0 {
		store.RegularFlushing()
	}

	return store
}

func (s *ReportAcceptStatus) SetAsyncExecutor(executor AsyncExecutor) {
	s.batchCore.setAsyncExecutor(executor)
}

func (s *ReportAcceptStatus) SetPartsPressureGuard(guard *PartsPressureGuard) {
	s.partsGuard = guard
	if guard != nil {
		s.batchCore.setBeforeFlush(func(bufferLength int) error {
			return guard.BeforeFlush(bufferLength)
		})
	}
}

func (s *ReportAcceptStatus) BypassPartsPressureGuard(enabled bool) {
	if s.partsGuard == nil {
		return
	}
	s.partsGuard.SetBypass(enabled)
}

func (s *ReportAcceptStatus) flushBatch(batch []*ReportAcceptStatusData) (remaining []*ReportAcceptStatusData, err error) {
	startNow := time.Now()
	buckets, err := s.planner.BuildBuckets(batch)
	if err != nil {
		return batch, err
	}

	if len(buckets) == 0 {
		return nil, nil
	}

	type failedBucket struct {
		bucket datePartitionBucket[*ReportAcceptStatusData]
		timing datePartitionBucketTiming
		err    error
		index  int
	}

	failedBuckets := make([]failedBucket, 0)
	var firstErr error
	for idx, bucket := range buckets {
		timing, bucketErr := s.flushBucketFn(bucket, idx+1, len(buckets))
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
		s.logBucketTiming(bucket, idx+1, len(buckets), timing, 0, nil)
		s.notifyBucketFlushed(bucket)
	}

	totalCost := time.Since(startNow)
	if len(failedBuckets) > 0 {
		failedOnly := make([]datePartitionBucket[*ReportAcceptStatusData], 0, len(failedBuckets))
		for _, failed := range failedBuckets {
			failedOnly = append(failedOnly, failed.bucket)
		}
		remaining = flattenDatePartitionBuckets(failedOnly)
		for _, failed := range failedBuckets {
			s.logBucketTiming(failed.bucket, failed.index, len(buckets), failed.timing, len(remaining), failed.err)
		}

		logs.Logger.Warn(
			"入库数据状态部分失败",
			zap.String("所花时间", totalCost.String()),
			zap.Int("数据长度为", len(batch)),
			zap.Int("桶数量", len(buckets)),
			zap.Int("失败桶数量", len(failedBuckets)),
			zap.Int("剩余数据长度", len(remaining)),
			zap.Error(firstErr),
		)
		if s.partsGuard != nil {
			s.partsGuard.ObserveFlushError(firstErr, len(batch))
		}
		return remaining, fmt.Errorf("report accept status flush partial failure: %w", firstErr)
	}

	logs.Logger.Info(
		"入库数据状态成功",
		zap.String("所花时间", totalCost.String()),
		zap.Int("数据长度为", len(batch)),
		zap.Int("桶数量", len(buckets)),
	)
	return nil, nil
}

func (s *ReportAcceptStatus) notifyBucketFlushed(bucket datePartitionBucket[*ReportAcceptStatusData]) {
	for _, row := range bucket.rows {
		if row.item.AfterFlush != nil {
			row.item.AfterFlush()
		}
	}
}

// buildReportAcceptStatusInsertArgs 负责把验收状态行转换成 ClickHouse v2 标准驱动可稳定接收的参数类型。
func buildReportAcceptStatusInsertArgs(row datePartitionRow[*ReportAcceptStatusData]) ([]interface{}, error) {
	partDate, err := ParseDateTimeString(row.item.PartDate)
	if err != nil {
		return nil, fmt.Errorf("parse acceptance status part_date failed: %w", err)
	}

	return []interface{}{
		int32(row.item.Status),
		row.partitionTime,
		partDate,
		int64(row.item.TableId),
		row.item.ReportType,
		row.item.DataName,
		row.item.ErrorReason,
		row.item.ErrorHandling,
		row.item.ReportData,
		row.item.XwlKafkaOffset,
	}, nil
}

// flushBucket 负责把一个 bucket 独立写入 xwl_acceptance_status。
//
// 一个 bucket 可以覆盖多个 ClickHouse 分区，但分区数不会超过 planner 算出的安全上限。
// 这样做的目的不是让 bucket 越大越好，而是：
// 1. 避免把过多分区混进一个 INSERT block。
// 2. 降低触发 `Too many partitions for single INSERT block` 的风险。
// 3. 在安全范围内尽量减少事务次数。
//
// 注意：
// 1. 单个 bucket 失败时，调用方不会立刻中断整个批次。
// 2. 后续 bucket 仍然会继续尝试。
// 3. 最终只把失败 bucket 对应的数据恢复回 buffer，已成功 bucket 保留为已落库状态。
func (s *ReportAcceptStatus) flushBucket(bucket datePartitionBucket[*ReportAcceptStatusData], bucketIndex, bucketCount int) (timing datePartitionBucketTiming, err error) {
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
	stmt, err := tx.Prepare("INSERT INTO xwl_acceptance_status (status,ingest_time,part_date, table_id,report_type, data_name, error_reason, error_handling, report_data, xwl_kafka_offset) VALUES (?,?,?,?,?,?,?,?,?,?)")
	timing.prepareCost = time.Since(stageBegin)
	if err != nil {
		util.RecordPersistenceError("clickhouse_prepare_failed", err)
		return timing, err
	}
	defer stmt.Close()

	timing.normalizeCost = bucket.normalizeCost
	for _, row := range bucket.rows {
		stageBegin = time.Now()
		args, err := buildReportAcceptStatusInsertArgs(row)
		if err != nil {
			timing.execCost += time.Since(stageBegin)
			timing.totalCost = time.Since(startNow)
			logs.Logger.Error("验收状态参数转换失败", zap.Error(err))
			return timing, err
		}
		if _, err := stmt.Exec(args...); err != nil {
			timing.execCost += time.Since(stageBegin)
			timing.totalCost = time.Since(startNow)
			util.RecordPersistenceError("clickhouse_sql_driver_error", err)
			logs.Logger.Error("入库数据状态出现错误", zap.Error(err))
			return timing, err
		}
		timing.execCost += time.Since(stageBegin)
	}

	stageBegin = time.Now()
	if err := tx.Commit(); err != nil {
		timing.commitCost = time.Since(stageBegin)
		timing.totalCost = time.Since(startNow)
		util.RecordPersistenceError("clickhouse_sql_driver_error", err)
		logs.Logger.Error("入库数据状态出现错误", zap.Error(err))
		return timing, err
	}
	timing.commitCost = time.Since(stageBegin)
	timing.totalCost = time.Since(startNow)
	return timing, nil
}

func (s *ReportAcceptStatus) logBucketTiming(bucket datePartitionBucket[*ReportAcceptStatusData], bucketIndex, bucketCount int, timing datePartitionBucketTiming, remainingSize int, err error) {
	// 这里只在 bucket 较慢或失败时打调试日志，目的是让下一轮排障能直接看到：
	// 1. 当前 bucket 覆盖了多少分区。
	// 2. 这些分区键是什么。
	// 3. 当前 bucket 是慢，还是直接失败。
	// 4. 失败后还有多少数据会被恢复回 buffer。
	//
	// 风险说明：
	// 如果 bucket 覆盖分区过多，就可能再次触发：
	// `Too many partitions for single INSERT block`
	// 因此这里把 `bucket_partition_count` 和 `bucket_partition_keys` 一并打出来。
	if err == nil && timing.totalCost < 100*time.Millisecond {
		return
	}

	bucketKey := ""
	if len(bucket.partitionKeys) == 1 {
		bucketKey = bucket.partitionKeys[0]
	}

	logs.Logger.Debug(
		"report accept status flush timing",
		append(
			s.planner.logFields(),
			zap.String("table_name", TableNameAcceptanceStatus),
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
