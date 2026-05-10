# SimpleTrack 读侧 ClickHouse Benchmark 基线

> 记录日期：2026-05-08
> 仓库：`src/analytics-core`
> 初始基线 commit：`1e65684eff8a90d5eb210052e4566d03b7d1c984`
> 最近复测 commit：`a99147f4da07ccfd6643722a891cad65c1270b3e`
> 目标：为 P1.5 ClickHouse 读侧优化提供真实 ClickHouse 基线，后续是否引入 projection、materialized view 或小时聚合表必须先和这份基线对比。

## 本次命令

默认 10k 行基线：

```powershell
$env:ANALYTICS_CORE_CLICKHOUSE_BENCH='1'
go test ./internal/e2e -run '^$' -bench 'BenchmarkEventReaderClickHouseExecution' -benchmem -count=3
```

100k 行 pressure run：

```powershell
$env:ANALYTICS_CORE_CLICKHOUSE_BENCH='1'
$env:ANALYTICS_CORE_CLICKHOUSE_BENCH_ROWS='100000'
go test ./internal/e2e -run '^$' -bench 'BenchmarkEventReaderClickHouseExecution' -benchmem -count=3
```

500k 行 pressure run：

```powershell
$env:ANALYTICS_CORE_CLICKHOUSE_BENCH='1'
$env:ANALYTICS_CORE_CLICKHOUSE_BENCH_ROWS='500000'
go test ./internal/e2e -run '^$' -bench 'BenchmarkEventReaderClickHouseExecution' -benchmem -count=3
```

依赖状态：

- `analytics-core-clickhouse`：`clickhouse/clickhouse-server:25.3`，native TCP `127.0.0.1:29000`。
- `analytics-core-redis`：`redis/redis-stack:latest`，本次 reader benchmark 不使用 Redis。
- MySQL 不参与本次 reader benchmark。

代码证据：

- benchmark 入口：`仓库: analytics-core, commit: a99147f, file: internal/e2e/clickhouse_reader_benchmark_test.go:22-60`。
- benchmark 只连接 ClickHouse，不混入 Redis / MySQL：`仓库: analytics-core, commit: a99147f, file: internal/e2e/clickhouse_reader_benchmark_test.go:89-91`。
- benchmark 会先 seed deterministic events / properties：`仓库: analytics-core, commit: a99147f, file: internal/e2e/clickhouse_reader_benchmark_test.go:108-113`。
- benchmark 场景覆盖 recent-window Realtime、wide-since Realtime、recent-window Events、条件启用的 bounded 24h / 72h / 7d scalar Events、recent-window typed property Events、wide-window scalar Events 和 wide-window typed property Events：`仓库: analytics-core, commit: a99147f, file: internal/e2e/clickhouse_reader_benchmark_test.go:128-225`。
- bounded scalar suite 现在按窗口规格展开：只有 `rowCount > windowRows` 才会进入 benchmark / explain 套件，避免默认小 fixture 把多日 bounded 场景伪装成 wide-window；窗口规格和 helper 位于 `仓库: analytics-core, commit: a99147f, file: internal/e2e/clickhouse_reader_benchmark_test.go:26-47` 与 `internal/e2e/clickhouse_reader_benchmark_test.go:443-531`。
- Realtime 场景会在计时前记录并校验 `since` 和 eligible row count，防止把 wide-since 压力查询误当成短窗口 Realtime：`仓库: analytics-core, commit: a99147f, file: internal/e2e/clickhouse_reader_benchmark_test.go:228-239` 和 `internal/e2e/clickhouse_reader_benchmark_test.go:883-899`。
- Events 场景会在计时前记录并校验 `from/to` eligible row count，并额外断言真实 `EventQueryPlan` 的 `QueryEvidence()` 和 bound args 都包含时间上下界，防止 helper 计算正确但 ClickHouse SQL 实际丢掉时间条件：`仓库: analytics-core, commit: a99147f, file: internal/e2e/clickhouse_reader_benchmark_test.go:236-246` 和 `internal/e2e/clickhouse_reader_benchmark_test.go:907-931`。
- 计时区只测 `EventReader` 执行：`仓库: analytics-core, commit: a99147f, file: internal/e2e/clickhouse_reader_benchmark_test.go:254-262`。
- explain 测试与 benchmark 复用同一套路由表和数据夹具，并记录 Realtime / Events window evidence：`仓库: analytics-core, commit: a99147f, file: internal/e2e/clickhouse_reader_benchmark_test.go:267-404`。
- explain 直接复用 sealed query plan SQL 和 bound args：`仓库: analytics-core, commit: 93cff0f, file: internal/e2e/clickhouse_reader_benchmark_test.go:773-795`。

## 默认 10k 行结果

| 场景 | 3 次结果 | 读侧含义 |
| --- | --- | --- |
| `low_realtime` | `10.43ms/op`, `8.63ms/op`, `8.70ms/op` | Realtime 短窗口、事实表直读、无属性表参与 |
| `medium_events_scalar` | `8.57ms/op`, `9.71ms/op`, `9.30ms/op` | Events 明细列表，事件名 + distinct id + 时间窗等标量过滤 |
| `high_events_property` | `15.17ms/op`, `15.84ms/op`, `16.25ms/op` | Events 明细列表，标量过滤 + typed property 表参与 |

原始输出：

```text
BenchmarkEventReaderClickHouseExecution/low_realtime-20              100  10432903 ns/op  169318 B/op  3229 allocs/op
BenchmarkEventReaderClickHouseExecution/low_realtime-20              128   8632088 ns/op  166715 B/op  3227 allocs/op
BenchmarkEventReaderClickHouseExecution/low_realtime-20              138   8697486 ns/op  165757 B/op  3228 allocs/op
BenchmarkEventReaderClickHouseExecution/medium_events_scalar-20      134   8572051 ns/op  169786 B/op  3271 allocs/op
BenchmarkEventReaderClickHouseExecution/medium_events_scalar-20      133   9710064 ns/op  168258 B/op  3270 allocs/op
BenchmarkEventReaderClickHouseExecution/medium_events_scalar-20      122   9298148 ns/op  167847 B/op  3270 allocs/op
BenchmarkEventReaderClickHouseExecution/high_events_property-20       66  15174430 ns/op  202709 B/op  3592 allocs/op
BenchmarkEventReaderClickHouseExecution/high_events_property-20       79  15838513 ns/op  203508 B/op  3593 allocs/op
BenchmarkEventReaderClickHouseExecution/high_events_property-20       85  16254566 ns/op  206283 B/op  3594 allocs/op
```

## 100k 行 pressure run 结果

| 场景 | 3 次结果 | 与 10k 基线的关系 |
| --- | --- | --- |
| `low_realtime` | `13.47ms/op`, `13.02ms/op`, `11.63ms/op` | 有上升，但仍是 direct fact table 可接受观察区 |
| `medium_events_scalar` | `11.09ms/op`, `11.58ms/op`, `10.74ms/op` | 有上升，但没有达到必须新增物理结构的证据强度 |
| `high_events_property` | `34.18ms/op`, `31.27ms/op`, `32.37ms/op` | 相比 10k 基线约 2x，当时列入重点观察候选；后续 500k 复测把候选收窄为宽时间窗 Events 与 typed property 过滤 |

原始输出：

```text
BenchmarkEventReaderClickHouseExecution/low_realtime-20               76  13465708 ns/op  162748 B/op  3225 allocs/op
BenchmarkEventReaderClickHouseExecution/medium_events_scalar-20      100  11092991 ns/op  167096 B/op  3346 allocs/op
BenchmarkEventReaderClickHouseExecution/high_events_property-20       39  34184221 ns/op  200720 B/op  3590 allocs/op
BenchmarkEventReaderClickHouseExecution/low_realtime-20               93  13023144 ns/op  165853 B/op  3227 allocs/op
BenchmarkEventReaderClickHouseExecution/low_realtime-20              109  11627696 ns/op  163816 B/op  3226 allocs/op
BenchmarkEventReaderClickHouseExecution/medium_events_scalar-20      100  11577511 ns/op  167105 B/op  3346 allocs/op
BenchmarkEventReaderClickHouseExecution/medium_events_scalar-20      100  10739546 ns/op  167066 B/op  3345 allocs/op
BenchmarkEventReaderClickHouseExecution/high_events_property-20       34  31274535 ns/op  199882 B/op  3590 allocs/op
BenchmarkEventReaderClickHouseExecution/high_events_property-20       32  32368522 ns/op  200155 B/op  3590 allocs/op
```

## 100k 行 value-free property evidence 复测

复测时间：2026-05-08。

复测 commit：`4393bbd0bdccf41b76f97670e9a57b26f3ecbd2a`。

复测目的：

- 验证 `query_evidence.property_filters` 只记录 `scope/name/value_type/operator` 后，100k 行真实 ClickHouse 读侧表现是否仍稳定。
- 确认 typed property 过滤路径是否已经需要立即进入 projection、materialized view 或小时聚合表实施。

复测结果：

| 场景 | 3 次结果 | 判断 |
| --- | --- | --- |
| `low_realtime` | `18.08ms/op`, `12.34ms/op`, `12.82ms/op` | 第一轮偏高但后两轮回到原 100k 观察区，继续 direct fact table |
| `medium_events_scalar` | `11.20ms/op`, `10.89ms/op`, `10.44ms/op` | 与上一轮 100k 基线一致 |
| `high_events_property` | `32.84ms/op`, `31.36ms/op`, `29.84ms/op` | 与上一轮 100k 基线一致，仍是重点观察候选，但不触发立即新增物理结构 |

原始输出：

```text
BenchmarkEventReaderClickHouseExecution/low_realtime-20               63  18082724 ns/op  167414 B/op  3228 allocs/op
BenchmarkEventReaderClickHouseExecution/low_realtime-20               93  12344256 ns/op  164035 B/op  3226 allocs/op
BenchmarkEventReaderClickHouseExecution/low_realtime-20               86  12824671 ns/op  166041 B/op  3226 allocs/op
BenchmarkEventReaderClickHouseExecution/medium_events_scalar-20       99  11195644 ns/op  168365 B/op  3346 allocs/op
BenchmarkEventReaderClickHouseExecution/medium_events_scalar-20      100  10888815 ns/op  168761 B/op  3346 allocs/op
BenchmarkEventReaderClickHouseExecution/medium_events_scalar-20      100  10438323 ns/op  167515 B/op  3346 allocs/op
BenchmarkEventReaderClickHouseExecution/high_events_property-20       38  32839516 ns/op  201101 B/op  3591 allocs/op
BenchmarkEventReaderClickHouseExecution/high_events_property-20       42  31357683 ns/op  199677 B/op  3590 allocs/op
BenchmarkEventReaderClickHouseExecution/high_events_property-20       40  29840255 ns/op  200748 B/op  3591 allocs/op
```

## 100k 行 bounded 24h scalar Events 证据

运行时间：2026-05-10。
运行 commit：`93cff0f140024ca006307490eed4b1fefef2cfb1`。

运行目的：

- 复核并修正 `analytics-service` 里 bounded scalar Events 的 pressure triage：先用真实 ClickHouse benchmark / explain 判断 `24h` 是否真的值得进入 `high` 桶。
- 验证 `93cff0f` 引入的 bounded 24h scalar Events 场景在 `rowCount > 86400` 时是否真的形成独立读形状，而不是默认小夹具下的伪分支。

运行命令：

```powershell
$env:ANALYTICS_CORE_CLICKHOUSE_BENCH='1'
$env:ANALYTICS_CORE_CLICKHOUSE_BENCH_ROWS='100000'
go test ./internal/e2e -run 'TestEventReaderClickHouseExplain/medium_events_scalar_bounded_24h_window' -count=1 -v
go test ./internal/e2e -run '^$' -bench 'BenchmarkEventReaderClickHouseExecution/medium_events_scalar_bounded_24h_window' -benchmem -count=3
```

Explain 结果摘要：

- `from=2026-05-01T11:46:40Z`
- `to=2026-05-02T11:46:40Z`
- `eligible_rows=86400`
- `query_evidence.time_window_seconds=86400`
- `ReadFromMergeTree` 主键条件仍然是 `tenant_id / project_id / source_id / event_time`
- `Granules: 12/13`

原始 explain 关键输出：

```text
events window evidence: from=2026-05-01T11:46:40Z to=2026-05-02T11:46:40Z eligible_rows=86400 row_count=100000
query evidence: {Family:events ReadPath:fact_events Optimization:direct_fact_table EffectiveLimit:50 Offset:0 HasTimeLowerBound:true HasTimeUpperBound:true TimeWindowSeconds:86400 ScalarFilterCount:4 PropertyFilterCount:0 UsesPropertyTable:false PropertyFilters:[] SortField:event_time SortDirection:desc}
ReadFromMergeTree (analytics_core.events_...)
Granules: 12/13
```

Benchmark 结果：

| 场景 | 3 次结果 | 判断 |
| --- | --- | --- |
| `medium_events_scalar_bounded_24h_window` | `10.79ms/op`, `12.61ms/op`, `15.52ms/op` | 明显重于 recent-window scalar Events，但仍明显轻于 500k wide-window scalar Events 的 `40ms+` 观察区 |

原始输出：

```text
BenchmarkEventReaderClickHouseExecution/medium_events_scalar_bounded_24h_window-20          99  10785435 ns/op  168995 B/op  3347 allocs/op
BenchmarkEventReaderClickHouseExecution/medium_events_scalar_bounded_24h_window-20          80  12607200 ns/op  169781 B/op  3347 allocs/op
BenchmarkEventReaderClickHouseExecution/medium_events_scalar_bounded_24h_window-20         100  15515889 ns/op  167806 B/op  3346 allocs/op
```

当前判断：

- `24h` bounded scalar Events 在 100k 行夹具下已经是一个真实独立的读形状，不再依赖 wide-window 压力结果做代称。
- 这个形状比 recent-window scalar Events 更重，但当前证据更接近“中高观察区”，还没有单独证明必须新增 projection、materialized view 或小时聚合表。
- 因此仅凭 `24h` bounded scalar Events 还不足以进入 `pressure=high`。如果要继续保留服务层时间窗 triage 阈值，至少需要更新鲜的 benchmark / explain 证据来支持它。

## 500k 行 bounded 24h scalar Events 证据

运行时间：2026-05-10。
运行 commit：`93cff0f140024ca006307490eed4b1fefef2cfb1`。

运行目的：

- 确认 bounded 24h scalar Events 在更大夹具下是否会向 wide-window scalar 的压力区靠拢。
- 验证 `analytics-service` 是否还应该保留 `24h => pressure=high` 这条 service heuristic。

运行命令：

```powershell
$env:ANALYTICS_CORE_CLICKHOUSE_BENCH='1'
$env:ANALYTICS_CORE_CLICKHOUSE_BENCH_ROWS='500000'
go test ./internal/e2e -run 'TestEventReaderClickHouseExplain/medium_events_scalar_bounded_24h_window' -count=1 -v
go test ./internal/e2e -run '^$' -bench 'BenchmarkEventReaderClickHouseExecution/medium_events_scalar_bounded_24h_window' -benchmem -count=3
```

Explain 结果摘要：

- `from=2026-05-06T02:53:20Z`
- `to=2026-05-07T02:53:20Z`
- `eligible_rows=86400`
- `query_evidence.time_window_seconds=86400`
- `ReadFromMergeTree` 主键条件仍然是 `tenant_id / project_id / source_id / event_time`
- `Granules: 12/62`

原始 explain 关键输出：

```text
events window evidence: from=2026-05-06T02:53:20Z to=2026-05-07T02:53:20Z eligible_rows=86400 row_count=500000
query evidence: {Family:events ReadPath:fact_events Optimization:direct_fact_table EffectiveLimit:50 Offset:0 HasTimeLowerBound:true HasTimeUpperBound:true TimeWindowSeconds:86400 ScalarFilterCount:4 PropertyFilterCount:0 UsesPropertyTable:false PropertyFilters:[] SortField:event_time SortDirection:desc}
ReadFromMergeTree (analytics_core.events_...)
Granules: 12/62
```

Benchmark 结果：

| 场景 | 3 次结果 | 判断 |
| --- | --- | --- |
| `medium_events_scalar_bounded_24h_window` | `10.02ms/op`, `10.14ms/op`, `10.63ms/op` | 在 500k 行夹具下仍接近 direct fact-table 的中等观察区，明显低于 wide-window scalar Events 的 `40ms+` 压力区 |

原始输出：

```text
BenchmarkEventReaderClickHouseExecution/medium_events_scalar_bounded_24h_window-20         100  10017510 ns/op  166781 B/op  3346 allocs/op
BenchmarkEventReaderClickHouseExecution/medium_events_scalar_bounded_24h_window-20         100  10135990 ns/op  167053 B/op  3346 allocs/op
BenchmarkEventReaderClickHouseExecution/medium_events_scalar_bounded_24h_window-20         100  10626158 ns/op  167035 B/op  3346 allocs/op
```

当前判断：

- 500k 行夹具下的 bounded 24h scalar Events 并没有向 wide-window scalar 压力区靠拢。
- 它依然保持 `TimeWindowSeconds=86400` 且只读 `12/62` granules，说明 ClickHouse 主键时间约束仍然有效。
- 这组证据直接支持后续撤回 `analytics-service` 的 `24h => pressure=high` heuristic：仅凭 24h bounded scalar 窗口不足以被标记为高压力。

## 500k 行 bounded 72h scalar Events 证据

运行时间：2026-05-10。
运行 commit：`a99147f4da07ccfd6643722a891cad65c1270b3e`。

运行目的：

- 用比 24h 更宽、但仍明显小于 wide-window 的 bounded scalar 时间窗，补齐 service triage 讨论所缺的中间证据层。
- 判断 bounded scalar 从 24h 扩到 72h 后，是否已经足够接近 wide-window scalar 的压力区。

运行命令：

```powershell
$env:ANALYTICS_CORE_CLICKHOUSE_BENCH='1'
$env:ANALYTICS_CORE_CLICKHOUSE_BENCH_ROWS='500000'
go test ./internal/e2e -run 'TestEventReaderClickHouseExplain/medium_events_scalar_bounded_72h_window' -count=1 -v
go test ./internal/e2e -run '^$' -bench 'BenchmarkEventReaderClickHouseExecution/medium_events_scalar_bounded_72h_window' -benchmem -count=3
```

Explain 结果摘要：

- `eligible_rows=259200`
- `query_evidence.time_window_seconds=259200`
- `Granules: 33/62`

Benchmark 结果：

| 场景 | 3 次结果 | 判断 |
| --- | --- | --- |
| `medium_events_scalar_bounded_72h_window` | `11.49ms/op`, `10.33ms/op`, `10.56ms/op` | 相比 24h 读了更多 granules，但整体仍留在 direct fact-table 的中等观察区，没有逼近 wide-window scalar 的 `40ms+` 压力区 |

当前判断：

- `72h` bounded scalar 已经明显比 `24h` 更宽，但在 500k 行夹具下仍没有出现需要立即新增物理结构的压力信号。
- 这组中间证据说明“只要 bounded 时间窗超过 24h 就进入 high”仍然过于激进。

## 1,000,000 行 bounded 7d scalar Events 证据

运行时间：2026-05-10。
运行 commit：`a99147f4da07ccfd6643722a891cad65c1270b3e`。

运行目的：

- 观察 bounded scalar 在完整 7 天窗口和更大 fixture 下，是否会进入与 wide-window scalar 接近的压力区。
- 为后续是否需要重新讨论 bounded scalar service heuristic 提供上限样本。

运行命令：

```powershell
$env:ANALYTICS_CORE_CLICKHOUSE_BENCH='1'
$env:ANALYTICS_CORE_CLICKHOUSE_BENCH_ROWS='1000000'
go test ./internal/e2e -run 'TestEventReaderClickHouseExplain/medium_events_scalar_bounded_7d_window' -count=1 -v
go test ./internal/e2e -run '^$' -bench 'BenchmarkEventReaderClickHouseExecution/medium_events_scalar_bounded_7d_window' -benchmem -count=3
```

Explain 结果摘要：

- `eligible_rows=604800`
- `query_evidence.time_window_seconds=604800`
- `Granules: 75/123`

Benchmark 结果：

| 场景 | 3 次结果 | 判断 |
| --- | --- | --- |
| `medium_events_scalar_bounded_7d_window` | `52.11ms/op`, `48.43ms/op`, `46.11ms/op` | 这是当前 bounded scalar 证据里第一次稳定进入 `46-52ms/op` 压力区，已经接近或达到 wide-window scalar 的观察区 |

当前判断：

- `7d` bounded scalar 终于表现出明确压力，但它对应的是更大 fixture 和更长历史窗口，不应被简单降格成“24h+ 全都 high”。
- 当前更合理的结论是：bounded scalar 需要基于更完整的 row-volume + window 证据判断，而不是只靠一个时间窗阈值拍板。

## ClickHouse explain 证据

命令：

```powershell
$env:ANALYTICS_CORE_CLICKHOUSE_BENCH='1'
$env:ANALYTICS_CORE_CLICKHOUSE_BENCH_ROWS='100000'
go test ./internal/e2e -run TestEventReaderClickHouseExplain -count=1 -v
```

观察摘要：

- `low_realtime`：仍是 routed fact table 的 `ReadFromMergeTree` 主键路径，没有属性表参与。
- `medium_events_scalar`：仍是 routed fact table 主键路径，主要受 `tenant_id / project_id / source_id / event_time` 条件约束。
- `high_events_property`：`query_evidence.property_filters` 只暴露 `{Scope:event Name:button ValueType:string Operator:eq}`、`{Scope:event Name:plan ValueType:string Operator:eq}`、`{Scope:user Name:tier ValueType:string Operator:eq}` 这类无 value 的过滤形状；explain 出现 `CreatingSets`，并且主查询上出现 3 个 `event_id in 1000-element set` 条件；主键条件已经包含 `visit_id`，但本次 100k 行下仍是 `Granules: 13/13`。

这说明：

- typed property filter 路径确实已经是下一条重点观察候选；
- 但当前更像“证据补齐”，还不是“立即新增 projection / materialized view / 小时聚合表”的触发器；
- 现阶段仍应先保持 direct fact table，并继续优先做属性治理、query plan 约束和更大数据量观察。

## 500k 行 Realtime / Events 形状纠偏与压力观察

500k 行复测前发现两个基线口径问题：

- 旧 `low_realtime` 场景的 `Since=baseTime-1m` 会随着 fixture 行数增大而变成“宽时间窗扫描”，不再代表产品里的短窗口 Realtime。
- 旧 `medium_events_scalar` / `high_events_property` 默认都是宽时间窗 Events，不能区分正常产品窗口和压力窗口。

因此 `analytics-core` commit `5bac8d8` 将 Realtime 拆成两个场景：

- `low_realtime_recent_window`：`eligible_rows=300`，代表产品短窗口 Realtime。
- `low_realtime_wide_since`：`eligible_rows=500000`，代表宽时间窗压力查询。

随后 `analytics-core` commit `caf314d` 将 Events 也拆成 recent / wide：

- `medium_events_scalar_recent_window`：`eligible_rows=5000`，代表近期 Events 明细标量过滤。
- `medium_events_scalar_wide_window`：`eligible_rows=500000`，代表宽时间窗 Events 标量过滤压力。
- `high_events_property_recent_window`：`eligible_rows=5000`，代表近期 Events typed property 过滤。
- `high_events_property_wide_window`：`eligible_rows=500000`，代表宽时间窗 typed property 过滤压力。

2026-05-10 的 `analytics-core` commit `93cff0f` 又把这个 benchmark / explain 基线再收紧了一步：

- `medium_events_scalar_bounded_24h_window` 只有在 fixture 大于 86400 行时才会启用。
- 默认 10k 本地基线不会再出现一个实际上等于 wide-window 的“伪 24h 场景”。
- `benchmarkEndTime` 改成精确排他上界后，bounded 24h slice 的 `QueryEvidence.TimeWindowSeconds` 会稳定等于 `86400`，不再因为 helper 偏差变成 `86401`。

500k explain 摘要：

| 场景 | eligible rows | Granules | 判断 |
| --- | --- | --- | --- |
| `low_realtime_recent_window` | `300` | `2/62` | 短窗口 Realtime 能利用时间下界缩小读取范围 |
| `low_realtime_wide_since` | `500000` | `62/62` | 宽时间窗会读完整 fixture，不能代表正常 Realtime |
| `medium_events_scalar_recent_window` | `5000` | `2/62` | 近期 Events 标量过滤能受时间窗约束 |
| `medium_events_scalar_wide_window` | `500000` | `62/62` | 宽时间窗 Events 标量过滤仍是压力观察对象 |
| `high_events_property_recent_window` | `5000` | `2/62`，3 个 `event_id in 5000-element set` | 近期属性过滤仍会创建 property set，但主表时间窗可缩小读取范围 |
| `high_events_property_wide_window` | `500000` | `62/62`，3 个 `event_id in 5000-element set` | 宽时间窗 typed property 过滤是重点观察对象 |

500k benchmark 结果：

| 场景 | 3 次结果 | 判断 |
| --- | --- | --- |
| `low_realtime_recent_window` | `8.95ms/op`, `7.86ms/op`, `8.65ms/op` | 产品短窗口 Realtime 仍稳定，不触发物理结构优化 |
| `low_realtime_wide_since` | `39.89ms/op`, `37.40ms/op`, `54.68ms/op` | 宽时间窗压力明显，不能和短窗口 Realtime 混用 |
| `medium_events_scalar_recent_window` | `8.24ms/op`, `9.02ms/op`, `8.44ms/op` | 近期 Events 标量过滤仍稳定 |
| `medium_events_scalar_wide_window` | `40.32ms/op`, `40.55ms/op`, `42.12ms/op` | 宽时间窗 Events 标量过滤进入观察区 |
| `high_events_property_recent_window` | `21.88ms/op`, `23.40ms/op`, `21.37ms/op` | 近期 typed property 过滤高于 scalar，但仍不是立即新增物理结构的证据 |
| `high_events_property_wide_window` | `43.06ms/op`, `43.02ms/op`, `44.12ms/op` | 宽时间窗 typed property 过滤是重点观察对象，但仍未单独证明必须新增物理结构 |

原始输出：

```text
BenchmarkEventReaderClickHouseExecution/low_realtime_recent_window-20          132   8945142 ns/op  162678 B/op  3225 allocs/op
BenchmarkEventReaderClickHouseExecution/low_realtime_recent_window-20          146   7860903 ns/op  162665 B/op  3225 allocs/op
BenchmarkEventReaderClickHouseExecution/low_realtime_recent_window-20          136   8650005 ns/op  162535 B/op  3225 allocs/op
BenchmarkEventReaderClickHouseExecution/low_realtime_wide_since-20              31  39889477 ns/op  164268 B/op  3226 allocs/op
BenchmarkEventReaderClickHouseExecution/low_realtime_wide_since-20              32  37396725 ns/op  166531 B/op  3226 allocs/op
BenchmarkEventReaderClickHouseExecution/low_realtime_wide_since-20              31  54684839 ns/op  167985 B/op  3227 allocs/op
BenchmarkEventReaderClickHouseExecution/medium_events_scalar_recent_window-20  144   8237374 ns/op  165435 B/op  3345 allocs/op
BenchmarkEventReaderClickHouseExecution/medium_events_scalar_recent_window-20  121   9021621 ns/op  165520 B/op  3345 allocs/op
BenchmarkEventReaderClickHouseExecution/medium_events_scalar_recent_window-20  146   8435766 ns/op  165580 B/op  3345 allocs/op
BenchmarkEventReaderClickHouseExecution/medium_events_scalar_wide_window-20     34  40319094 ns/op  170310 B/op  3347 allocs/op
BenchmarkEventReaderClickHouseExecution/medium_events_scalar_wide_window-20     27  40552074 ns/op  171213 B/op  3347 allocs/op
BenchmarkEventReaderClickHouseExecution/medium_events_scalar_wide_window-20     30  42116117 ns/op  169531 B/op  3346 allocs/op
BenchmarkEventReaderClickHouseExecution/high_events_property_recent_window-20   60  21876347 ns/op  200152 B/op  3593 allocs/op
BenchmarkEventReaderClickHouseExecution/high_events_property_recent_window-20   45  23395796 ns/op  200830 B/op  3593 allocs/op
BenchmarkEventReaderClickHouseExecution/high_events_property_recent_window-20   60  21367012 ns/op  200267 B/op  3593 allocs/op
BenchmarkEventReaderClickHouseExecution/high_events_property_wide_window-20     30  43055423 ns/op  203896 B/op  3594 allocs/op
BenchmarkEventReaderClickHouseExecution/high_events_property_wide_window-20     28  43015307 ns/op  204218 B/op  3594 allocs/op
BenchmarkEventReaderClickHouseExecution/high_events_property_wide_window-20     27  44119341 ns/op  204432 B/op  3595 allocs/op
```

## 当前判断

本次基线只证明三类读形状在本地 ClickHouse 上可以稳定执行，不证明现在必须引入 projection、materialized view 或小时聚合表。

当前应继续保持：

- Realtime / Events 默认走 direct fact table。
- `EventQueryBuilder` / `EventReader` 仍是唯一读侧入口。
- `query_evidence` 继续作为后续优化评审输入。
- 产品短窗口 Realtime 暂不构成读侧物理结构优化压力。后续不要再用 wide-since benchmark 代表正常 Realtime。
- 近期 Events 明细查询暂不构成读侧物理结构优化压力。后续不要再用 wide-window Events 结果代表正常近期 Events。
- `analytics-core` commit `f84024a` 之后，typed property filter 不再允许无限宽历史窗口：查询必须显式带 `from/to`，并且 direct fact-table 路径默认只允许 7 天内窗口。
- 下一条重点观察候选因此收窄为“宽时间窗 scalar Events 明细查询”和“7 天内 typed property 过滤读路径”。500k 行下旧 `high_events_property_wide_window` 仍保留为压力证据：它进入约 43-44ms/op 区间，explain 已出现 `CreatingSets`、3 个 `event_id in ... set` 和包含 `visit_id` 的主键条件；但这个旧宽窗口结果不能再被当成默认可放开的产品能力。
- `analytics-service` commit `c08e1da` 已撤回“bounded Events `24h+` 且无 property join => pressure=high`”这条 service heuristic。原因是 `24h` 和 `72h` bounded scalar 证据都仍落在 direct fact-table 的中等观察区，而不是 wide-window scalar 的压力区。
- `analytics-core` commit `a99147f` 之后，bounded scalar 已经形成 `24h -> 72h -> 7d` 的证据梯度：前两档仍偏中等，`7d @ 1,000,000 rows` 才首次稳定进入 `46-52ms/op` 压力区。
- 这说明 bounded scalar Events 的 pressure triage 目前不适合只靠一个时间窗阈值推断；如果未来要重引入类似规则，至少要把时间窗和 row volume 一起纳入判断，而不是简单回到 `24h+` 特判。
- 如果后续要支持超过 7 天的 property 历史过滤，必须先回到实施决策评审，补新的 query evidence、benchmark、explain 和物理结构方案；不能只删除 query-builder guardrail。

只有当同一 query shape 在更大数据量或连续 benchmark 中稳定超过基线，并且 explain、属性治理、过滤白名单、limit cap、时间窗约束和回归计划都支持继续下沉时，才进入 projection / MV / 小时聚合表评审。
