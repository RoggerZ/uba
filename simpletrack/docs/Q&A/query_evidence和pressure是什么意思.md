# query_evidence 和 pressure 是什么意思？

## Q：`query_evidence` 是什么？

A：`query_evidence` 是读侧查询的结构化证据，来自 `analytics-core` 的 `EventQueryEvidence`。它不是 SQL 文本，也不是执行计划全文，而是把当前这次查询的关键形状摘要出来，方便判断这条读请求走了什么路径。

通常会包含：

- 查询家族，例如 `events` 或 `realtime`
- 读路径，例如事实表读取
- 当前优化方式，例如直接查明细表
- 标量过滤数量
- 属性过滤数量
- 是否用了属性表
- 排序字段和排序方向

## Q：`pressure` 是什么？

A：`pressure` 是从 `query_evidence` 派生出来的粗粒度 triage 桶，只有 `low`、`medium`、`high` 三档。

它的作用是帮我们先判断这条查询“轻不轻”，用于后续优化和压测取舍，不是延迟 SLA，也不是自动扩缩容信号。

当前口径是：

- `low`：属性过滤为 0，且标量过滤不多
- `medium`：过滤开始增多，但还没到复杂查询
- `high`：过滤更多，或属性条件更重

## Q：为什么要单独看这两个字段？

A：`query_evidence` 负责解释“这条查询为什么长这样”，`pressure` 负责给出“这条查询大概重不重”的第一眼判断。

它们配合起来，方便后续决定要不要引入 projection、materialized view 或小时聚合表。

## Q：现在应该怎么理解它们？

A：当前阶段它们都属于读侧优化的辅助信息，不会改变业务返回结果。SimpleTrack 仍然保持 `EventQueryBuilder` / `EventReader` 作为唯一读侧入口，`query_evidence` 和 `pressure` 只是帮助我们做长期读侧优化决策。
