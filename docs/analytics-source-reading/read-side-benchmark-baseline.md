# SimpleTrack 读侧 ClickHouse Benchmark 基线

> 记录日期：2026-05-08
> 仓库：`src/analytics-core`
> commit：`979a29fcfd6a09ee4433a5e8f42d97ddb247dbc1`
> 目标：为 P1.5 ClickHouse 读侧优化提供真实 ClickHouse 基线，后续是否引入 projection、materialized view 或小时聚合表必须先和这份基线对比。

## 本次命令

```powershell
$env:ANALYTICS_CORE_CLICKHOUSE_BENCH='1'
go test ./internal/e2e -run '^$' -bench 'BenchmarkEventReaderClickHouseExecution' -benchmem -count=3
```

依赖状态：

- `analytics-core-clickhouse`：`clickhouse/clickhouse-server:25.3`，native TCP `127.0.0.1:29000`。
- `analytics-core-redis`：`redis/redis-stack:latest`，本次 reader benchmark 不使用 Redis。
- MySQL 不参与本次 reader benchmark。

代码证据：

- benchmark 入口：`仓库: analytics-core, commit: 979a29f, file: internal/e2e/clickhouse_reader_benchmark_test.go:25-31`。
- benchmark 只连接 ClickHouse，不混入 Redis / MySQL：`仓库: analytics-core, commit: 979a29f, file: internal/e2e/clickhouse_reader_benchmark_test.go:39-43`。
- benchmark 会先 seed deterministic events / properties：`仓库: analytics-core, commit: 979a29f, file: internal/e2e/clickhouse_reader_benchmark_test.go:58-63`。
- benchmark 场景覆盖 low realtime、medium scalar events、high property events：`仓库: analytics-core, commit: 979a29f, file: internal/e2e/clickhouse_reader_benchmark_test.go:78-147`。
- 计时区只测 `EventReader` 执行：`仓库: analytics-core, commit: 979a29f, file: internal/e2e/clickhouse_reader_benchmark_test.go:149-170`。

## 本次结果

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

## 当前判断

本次基线只证明三类读形状在本地 ClickHouse 上可以稳定执行，不证明现在必须引入 projection、materialized view 或小时聚合表。

当前应继续保持：

- Realtime / Events 默认走 direct fact table。
- `EventQueryBuilder` / `EventReader` 仍是唯一读侧入口。
- `query_evidence` 继续作为后续优化评审输入。

只有当同一 query shape 在更大数据量或连续 benchmark 中稳定超过基线，并且无法通过属性治理、过滤白名单、limit cap、时间窗约束解决时，才进入 projection / MV / 小时聚合表评审。
