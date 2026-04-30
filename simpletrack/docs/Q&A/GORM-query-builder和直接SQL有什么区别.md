# GORM Query Builder 和直接 SQL 有什么区别

## Q：直接迁移 xwl_bi 的 SQL 和使用 query builder 有什么区别？

A：直接迁移 SQL 是把 xwl_bi 里已有的 ClickHouse 查询字符串搬过来，改字段名后继续执行。query builder 是把查询条件、字段白名单、时间范围、分组、过滤、排序这些逻辑统一封装，由代码生成最终 SQL。

简单说：

- 直接 SQL：快，但容易散。
- query builder：慢一点设计，但长期可维护。

## Q：为什么推荐统一 query builder？

A：因为 `analytics-core` 后续会有 Events、Funnels、Retention、Paths、Segments、Attribution 等很多查询。如果每个模块都自己拼 SQL，会有几个问题：

- 字段名和表名重复散落。
- 过滤条件不一致。
- SQL 注入或非法字段风险更高。
- 后续改统一事件表、分区表、属性字段时成本很高。
- 同一个 filter 在 Events 和 Funnels 里可能表现不一致。

统一 query builder 可以让所有查看和分析都走同一套字段白名单、时间范围、权限边界和属性过滤逻辑。

## Q：GORM 能覆盖 sqlx 和 Squirrel 的用途吗？

A：对当前 `analytics-core` 的目标来说，可以。

我们使用的是 GORM v2 体系，也就是 import path 为 `gorm.io/gorm` 的版本。虽然 Go module release 号显示为 `v1.x.x`，例如 `v1.31.1`，但这不是老的 `github.com/jinzhu/gorm`。

GORM v2 支持：

- Raw SQL。
- Named Argument。
- Clauses。
- Scopes。
- Generics API。
- SQL Builder。
- context 传递。

这些能力足够覆盖 sqlx 的查询扫描能力，以及 Squirrel 这类 SQL builder 的大部分使用场景。

## Q：ClickHouse 查询也用 GORM 吗？

A：建议统一由 GORM query builder 层负责构建和执行。即使某些复杂 ClickHouse 查询最终需要 Raw SQL，也应该从统一 builder 出口生成，而不是散落在 handler 里。

例如：

- Events 查询用统一 filter builder。
- Funnels 查询复用事件过滤和用户过滤 builder。
- Retention 查询复用 cohort 条件 builder。
- Segments 查询复用属性过滤 builder。

## Q：Funnel / Retention 要直接迁移 SQL 吗？

A：不建议直接完整迁移。推荐做法是：

1. 保留 xwl_bi 中 ClickHouse 查询思路，例如 `windowFunnel`。
2. 重写字段命名和表模型。
3. 把时间范围、事件条件、用户条件、属性过滤统一放进 query builder。
4. 为每类查询补 SQL 生成测试和结果校验。

这样既不丢掉 xwl_bi 的分析经验，也不会把旧结构原样带进 `analytics-core`。
