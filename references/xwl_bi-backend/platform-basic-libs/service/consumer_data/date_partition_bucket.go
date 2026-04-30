package consumer_data

import (
	"math"
	"sort"
	"time"

	"go.uber.org/zap"
)

const (
	datePartitionBucketTargetRatio = 0.5
	datePartitionFallbackTarget    = 30
	datePartitionKeyLayoutDay      = "20060102"
	datePartitionKeyLayoutMonth    = "200601"
)

// DatePartitionFallbackTarget 返回查询 ClickHouse setting 失败时使用的保守回退值。
//
// 这个函数主要给运行时初始化日志和测试复用：
// 1. 避免把 30 这种策略常量散落在多个包里重复写。
// 2. 明确这个值的含义是“目标桶分区上限回退值”，不是 ClickHouse 的真实环境上限。
func DatePartitionFallbackTarget() int {
	return datePartitionFallbackTarget
}

type datePartitionRow[T any] struct {
	item          T
	partitionTime time.Time
	partitionKey  string
	normalizeCost time.Duration
}

type datePartitionBucket[T any] struct {
	partitionKeys []string
	rows          []datePartitionRow[T]
	normalizeCost time.Duration
}

type datePartitionBucketTiming struct {
	totalCost     time.Duration
	beginCost     time.Duration
	prepareCost   time.Duration
	normalizeCost time.Duration
	execCost      time.Duration
	commitCost    time.Duration
}

// datePartitionBucketPlanner 负责把一批“按日期分区”的写入数据规划成多个安全 bucket。
//
// 这个通用 planner 存在的直接原因，就是规避：
// `Too many partitions for single INSERT block`
//
// 风险背景：
//  1. 某些单表批量器的表按 `toYYYYMMDD(业务时间)` 分区。
//  2. 如果把很多不同日期的数据混成一个 INSERT block，就会同时触达过多分区。
//  3. 一旦单次 INSERT 覆盖分区数超过 ClickHouse 的
//     `max_partitions_per_insert_block`，整批就会被直接拒绝。
//
// 所以 planner 的核心目标不是“让桶更大”，而是：
// 1. 在不超过风险边界的前提下，限制单次 INSERT 涉及的分区数。
// 2. 给后续 bucket 级事务插入提供稳定输入。
//
// 默认策略说明：
//  1. 程序初始化时会查询一次 `max_partitions_per_insert_block`。
//  2. 如果查到真实值，例如 100，本轮默认只取其中的 50% 作为目标上限。
//  3. 之所以不是直接顶着 100 去写，是因为官方/社区都强调“跨分区越少越稳”，
//     但并没有给出明确推荐比例，所以这里采用保守值 50%。
//  4. 如果查询失败，则直接回退到 30，优先保证稳定性。
type datePartitionBucketPlanner[T any] struct {
	runtimePartitionLimit    int
	effectivePartitionTarget int
	partitionTargetRatio     float64
	usedFallbackTarget       bool
	partitionKeyLayout       string
	extractTime              func(T) string
}

func newDatePartitionBucketPlanner[T any](runtimePartitionLimit int, extractTime func(T) string) *datePartitionBucketPlanner[T] {
	return newDatePartitionBucketPlannerWithLayout(runtimePartitionLimit, extractTime, datePartitionKeyLayoutDay)
}

func newDatePartitionBucketPlannerWithLayout[T any](runtimePartitionLimit int, extractTime func(T) string, partitionKeyLayout string) *datePartitionBucketPlanner[T] {
	planner := &datePartitionBucketPlanner[T]{
		runtimePartitionLimit:    runtimePartitionLimit,
		effectivePartitionTarget: datePartitionFallbackTarget,
		partitionTargetRatio:     datePartitionBucketTargetRatio,
		usedFallbackTarget:       runtimePartitionLimit <= 0,
		partitionKeyLayout:       partitionKeyLayout,
		extractTime:              extractTime,
	}

	if runtimePartitionLimit > 0 {
		planner.effectivePartitionTarget = maxIntLocal(1, int(math.Floor(float64(runtimePartitionLimit)*datePartitionBucketTargetRatio)))
	}

	return planner
}

func (p *datePartitionBucketPlanner[T]) logFields() []zap.Field {
	return []zap.Field{
		zap.Int("runtime_partition_limit", p.runtimePartitionLimit),
		zap.Int("effective_partition_target", p.effectivePartitionTarget),
		zap.Float64("partition_target_ratio", p.partitionTargetRatio),
		zap.Bool("used_fallback_target", p.usedFallbackTarget),
		zap.String("partition_key_layout", p.partitionKeyLayout),
	}
}

// BuildBuckets 把一批按日期分区的数据规划成多个 bucket。
//
// 处理步骤：
// 1. 调 extractTime 取出每条记录的日期时间字符串。
// 2. 解析成 time.Time。
// 3. 根据 planner 配置的分区粒度生成分区键，例如：
//   - 按天：`YYYYMMDD`
//   - 按月：`YYYYMM`
//
// 4. 先按分区键聚合成 partition group。
// 5. 再按分区键排序，保证 bucket 构造稳定可预测。
// 6. 最后根据 `effectivePartitionTarget` 组装 bucket。
//
// 示例：
// 1. 如果 batch 中有 20260408、20260409、20260410 三个分区
// 2. 且 `effectivePartitionTarget=2`
// 3. 则最终会得到两个 bucket：
//   - bucket1: [20260408, 20260409]
//   - bucket2: [20260410]
func (p *datePartitionBucketPlanner[T]) BuildBuckets(items []T) ([]datePartitionBucket[T], error) {
	if len(items) == 0 {
		return nil, nil
	}

	type partitionGroup struct {
		partitionKey  string
		rows          []datePartitionRow[T]
		normalizeCost time.Duration
	}

	groups := make(map[string]*partitionGroup, len(items))
	for _, item := range items {
		begin := time.Now()
		partitionTime, err := normalizeDateTimeForClickHouse(p.extractTime(item))
		normalizeCost := time.Since(begin)
		if err != nil {
			return nil, err
		}

		partitionKey := partitionTime.Format(p.partitionKeyLayout)
		group, ok := groups[partitionKey]
		if !ok {
			group = &partitionGroup{partitionKey: partitionKey}
			groups[partitionKey] = group
		}

		group.rows = append(group.rows, datePartitionRow[T]{
			item:          item,
			partitionTime: partitionTime,
			partitionKey:  partitionKey,
			normalizeCost: normalizeCost,
		})
		group.normalizeCost += normalizeCost
	}

	partitionKeys := make([]string, 0, len(groups))
	for key := range groups {
		partitionKeys = append(partitionKeys, key)
	}
	sort.Strings(partitionKeys)

	buckets := make([]datePartitionBucket[T], 0, (len(partitionKeys)/p.effectivePartitionTarget)+1)
	current := datePartitionBucket[T]{
		partitionKeys: make([]string, 0, p.effectivePartitionTarget),
	}

	for _, key := range partitionKeys {
		group := groups[key]
		if len(current.partitionKeys) >= p.effectivePartitionTarget {
			buckets = append(buckets, current)
			current = datePartitionBucket[T]{
				partitionKeys: make([]string, 0, p.effectivePartitionTarget),
			}
		}

		current.partitionKeys = append(current.partitionKeys, key)
		current.rows = append(current.rows, group.rows...)
		current.normalizeCost += group.normalizeCost
	}

	if len(current.partitionKeys) > 0 {
		buckets = append(buckets, current)
	}

	return buckets, nil
}

func flattenDatePartitionBuckets[T any](buckets []datePartitionBucket[T]) []T {
	if len(buckets) == 0 {
		return nil
	}

	total := 0
	for _, bucket := range buckets {
		total += len(bucket.rows)
	}

	remaining := make([]T, 0, total)
	for _, bucket := range buckets {
		for _, row := range bucket.rows {
			remaining = append(remaining, row.item)
		}
	}
	return remaining
}

func maxIntLocal(a, b int) int {
	if a > b {
		return a
	}
	return b
}
