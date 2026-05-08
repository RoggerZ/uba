# guardrails 是什么意思？

## Q：文档里说的 guardrails 是什么意思？

A：`guardrails` 直译是“护栏”。在技术方案里，它指一组防止系统跑偏的约束规则。

放到 SimpleTrack 的 ClickHouse 读侧里，`guardrails` 不是一个具体功能页面，而是类似这些保护边界：

- query limit 上限：避免一次查询返回过多行，把 ClickHouse 或页面打爆。
- filter 数量上限：避免用户组合太多过滤条件，导致 SQL 复杂度失控。
- property allowlist：只有明确允许的事件属性 / 用户属性才能作为查询过滤条件。
- sort allowlist：只能按白名单字段排序，不能把前端字符串直接拼进 SQL。
- 统一 query builder：所有 Events / Realtime 查询都必须走 `EventQueryBuilder`，不能业务层临时手写 SQL。

## Q：为什么不直接翻译成“防护栏”？

A：可以翻译，但在工程文档里建议写成“读侧约束 / 查询护栏 / guardrails”。中文读者能理解含义，英文词也方便和代码里的 `readSidePolicy`、query limit、allowlist 对上。

## Q：SimpleTrack 当前的 guardrails 对应什么代码？

A：当前对应 `analytics-core/storage/clickhouse` 内部的 `readSidePolicy` 和 `EventQueryBuilder`。

它们负责把查询限制留在 `analytics-core` 的 ClickHouse adapter 内部：上层的 `simpletrack-anaysitics-service` 只传入经过校验的 query 参数，`simpletrack-saas` 只通过 readback API 读数据，不能直接拼 ClickHouse SQL。
