# SimpleTrack 读侧 ClickHouse Benchmark 基线

> 记录日期：2026-05-08
> 仓库：`src/analytics-core`
> 初始基线 commit：`1e65684eff8a90d5eb210052e4566d03b7d1c984`
> 最近复测 commit：`5bac8d8e64234f711fe567e5df865d894e1a2409`
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

- benchmark 入口：`仓库: analytics-core, commit: 5bac8d8, file: internal/e2e/clickhouse_reader_benchmark_test.go:25-32`。
- benchmark 只连接 ClickHouse，不混入 Redis / MySQL：`仓库: analytics-core, commit: 5bac8d8, file: internal/e2e/clickhouse_reader_benchmark_test.go:40-44`。
- benchmark 会先 seed deterministic events / properties：`仓库: analytics-core, commit: 5bac8d8, file: internal/e2e/clickhouse_reader_benchmark_test.go:59-64`。
- benchmark 场景覆盖 recent-window Realtime、wide-since Realtime、medium scalar events、high property events：`仓库: analytics-core, commit: 5bac8d8, file: internal/e2e/clickhouse_reader_benchmark_test.go:78-174`。
- Realtime 场景会在计时前记录并校验 `since` 和 eligible row count，防止把 wide-since 压力查询误当成短窗口 Realtime：`仓库: analytics-core, commit: 5bac8d8, file: internal/e2e/clickhouse_reader_benchmark_test.go:174-196` 和 `internal/e2e/clickhouse_reader_benchmark_test.go:681-718`。
- 计时区只测 `EventReader` 执行：`仓库: analytics-core, commit: 5bac8d8, file: internal/e2e/clickhouse_reader_benchmark_test.go:174-196`。
- explain 测试与 benchmark 复用同一套路由表和数据夹具，并记录 Realtime eligible row evidence：`仓库: analytics-core, commit: 5bac8d8, file: internal/e2e/clickhouse_reader_benchmark_test.go:198-343`。
- explain 直接复用 sealed query plan SQL 和 bound args：`仓库: analytics-core, commit: 5bac8d8, file: internal/e2e/clickhouse_reader_benchmark_test.go:633-660`。

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

## 500k 行 Realtime 形状纠偏与压力观察

500k 行复测前发现一个基线口径问题：旧 `low_realtime` 场景的 `Since=baseTime-1m` 会随着 fixture 行数增大而变成“宽时间窗扫描”，不再代表产品里的短窗口 Realtime。

因此 `analytics-core` commit `5bac8d8` 将 Realtime 拆成两个场景：

- `low_realtime_recent_window`：`eligible_rows=300`，代表产品短窗口 Realtime。
- `low_realtime_wide_since`：`eligible_rows=500000`，代表宽时间窗压力查询。

500k explain 摘要：

| 场景 | eligible rows | Granules | 判断 |
| --- | --- | --- | --- |
| `low_realtime_recent_window` | `300` | `2/62` | 短窗口 Realtime 能利用时间下界缩小读取范围 |
| `low_realtime_wide_since` | `500000` | `62/62` | 宽时间窗会读完整 fixture，不能代表正常 Realtime |
| `medium_events_scalar` | 不适用 | `62/62` | 宽时间窗 Events 标量过滤仍是压力观察对象 |
| `high_events_property` | 不适用 | `62/62`，3 个 `event_id in 5000-element set` | 属性过滤仍是重点观察对象 |

500k benchmark 结果：

| 场景 | 3 次结果 | 判断 |
| --- | --- | --- |
| `low_realtime_recent_window` | `7.73ms/op`, `8.41ms/op`, `8.23ms/op` | 产品短窗口 Realtime 仍稳定，不触发物理结构优化 |
| `low_realtime_wide_since` | `45.74ms/op`, `33.88ms/op`, `33.74ms/op` | 宽时间窗压力明显，不能和短窗口 Realtime 混用 |
| `medium_events_scalar` | `37.75ms/op`, `38.36ms/op`, `42.77ms/op` | 宽时间窗 Events 标量过滤进入观察区 |
| `high_events_property` | `39.62ms/op`, `41.21ms/op`, `43.89ms/op` | 属性过滤继续是重点观察对象，但仍未单独证明必须新增物理结构 |

原始输出：

```text
BenchmarkEventReaderClickHouseExecution/low_realtime_recent_window-20  140   7734325 ns/op  162679 B/op  3225 allocs/op
BenchmarkEventReaderClickHouseExecution/low_realtime_recent_window-20  146   8408761 ns/op  163727 B/op  3225 allocs/op
BenchmarkEventReaderClickHouseExecution/low_realtime_recent_window-20  140   8234399 ns/op  162616 B/op  3225 allocs/op
BenchmarkEventReaderClickHouseExecution/low_realtime_wide_since-20      25  45743896 ns/op  167298 B/op  3227 allocs/op
BenchmarkEventReaderClickHouseExecution/low_realtime_wide_since-20      30  33884143 ns/op  167973 B/op  3227 allocs/op
BenchmarkEventReaderClickHouseExecution/low_realtime_wide_since-20      36  33743761 ns/op  167334 B/op  3227 allocs/op
BenchmarkEventReaderClickHouseExecution/medium_events_scalar-20         31  37750048 ns/op  170622 B/op  3347 allocs/op
BenchmarkEventReaderClickHouseExecution/medium_events_scalar-20         30  38359730 ns/op  170861 B/op  3347 allocs/op
BenchmarkEventReaderClickHouseExecution/medium_events_scalar-20         36  42773969 ns/op  170280 B/op  3347 allocs/op
BenchmarkEventReaderClickHouseExecution/high_events_property-20         31  39623755 ns/op  203806 B/op  3594 allocs/op
BenchmarkEventReaderClickHouseExecution/high_events_property-20         31  41206961 ns/op  203825 B/op  3594 allocs/op
BenchmarkEventReaderClickHouseExecution/high_events_property-20         32  43893362 ns/op  203683 B/op  3594 allocs/op
```

## 当前判断

本次基线只证明三类读形状在本地 ClickHouse 上可以稳定执行，不证明现在必须引入 projection、materialized view 或小时聚合表。

当前应继续保持：

- Realtime / Events 默认走 direct fact table。
- `EventQueryBuilder` / `EventReader` 仍是唯一读侧入口。
- `query_evidence` 继续作为后续优化评审输入。
- 产品短窗口 Realtime 暂不构成读侧物理结构优化压力。后续不要再用 wide-since benchmark 代表正常 Realtime。
- 下一条重点观察候选调整为“宽时间窗 Events 明细查询”和 `high_events_property`：500k 行下二者都进入约 37-44ms/op 区间，其中 high property explain 已出现 `CreatingSets`、3 个 `event_id in ... set` 和包含 `visit_id` 的主键条件；但还需要稳定 query pattern 和回归计划，才能决定是继续只做属性治理，还是进入 projection / MV / 小时聚合表评审。

只有当同一 query shape 在更大数据量或连续 benchmark 中稳定超过基线，并且 explain、属性治理、过滤白名单、limit cap、时间窗约束和回归计划都支持继续下沉时，才进入 projection / MV / 小时聚合表评审。
