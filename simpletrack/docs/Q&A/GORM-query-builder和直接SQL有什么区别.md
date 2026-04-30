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

## Q：ClickHouse 的批量写入也用 GORM 吗？

A：不把 GORM 作为 ClickHouse 高吞吐写入热路径的默认方案。

GORM 本身支持批量创建：`Create(&slice)` 和 `CreateInBatches(slice, size)` 会生成批量 `INSERT`。`gorm.io/driver/clickhouse` 也有 ClickHouse driver，示例里可以 `db.Create(&users)` 做批量插入。因此从“能不能 insert into / batch insert”的角度看，GORM 是能做的。

但 `analytics-core` 的事件入库是高频写入热路径，不只是“能写进去”，还要关注吞吐、内存、反射开销、复杂字段类型和 ClickHouse 原生批处理能力。ClickHouse 官方 Go 文档推荐大批量写入使用 native API 的 `PrepareBatch` / `Append` / `Send`，并建议在可能时使用列式写入和强类型，减少转换开销。

因此当前决策是：

- ClickHouse 查询：统一走 GORM query builder / Raw / Scopes 出口。
- MySQL 元数据和状态表：使用 GORM。
- ClickHouse 低频管理写入：可以用 GORM。
- ClickHouse 事件明细高吞吐写入：优先使用 `clickhouse-go/v2` 原生 batch writer，例如 `PrepareBatch`，必要时再评估 `ch-go`。

## Q：GORM batch insert 和原生 ClickHouse batch insert 差多少？

A：不能在没有压测的情况下给固定倍数，但方向可以先拍板。

差异主要在这里：

| 方案 | 优点 | 风险 |
| --- | --- | --- |
| GORM `CreateInBatches` | 代码统一，模型映射方便，适合低频写入和管理表 | ORM 层有模型解析、反射、hook、字段转换和 SQL 生成开销；复杂 ClickHouse 类型要额外验证 |
| `clickhouse-go/v2 PrepareBatch` | 更贴近 ClickHouse 原生批量写入路径，支持 batch append、struct append、列式写入 | 需要自己维护字段映射和 writer adapter |
| `ch-go` | 更偏极致写入性能，适合百万级每秒写入目标 | 使用复杂度更高，类型更严格，P1 不默认上 |

xwl_bi 之前已经在性能调优中把 ClickHouse insert 改成原生批量插入，这个经验要保留。`analytics-core` 的设计应该让 `EventWriter` 接口隐藏写入实现：先用 `clickhouse-go/v2 PrepareBatch` 做默认高吞吐 writer，再用压测对比 GORM batch insert。如果压测证明 GORM 足够且维护收益更大，可以调整；在此之前不把 GORM batch insert 放到事件入库主路径。

当前已采纳这个建议：实施 `analytics-core` 时，事件明细入库按 `EventWriter + clickhouse-go/v2 PrepareBatch` 方案实施，GORM 只负责查询构建、元数据表和低频管理写入。

## Q：`ch-go` 是什么？

A：`ch-go` 是 ClickHouse 官方组织下的低层 Go 客户端，不是随便一个第三方库。它的包名是 `github.com/ClickHouse/ch-go`。

可以这样理解：

- `clickhouse-go/v2`：高层 Go driver，支持 `database/sql`、连接池、Row-oriented batch、`PrepareBatch`，开发体验更友好。
- `ch-go`：低层 native client，更偏列式写入和数据块流式传输，CPU 和内存开销可以更低，但类型更严格，使用复杂度更高。

ClickHouse 官方 Go 文档也把两者放在一起说明：`clickhouse-go` 更适合普通查询和批量插入，`ch-go` 更适合追求极限性能的低层场景。并且 `clickhouse-go/v2` 版本内部也会利用 `ch-go` 做编码、解码和压缩能力。

所以在 `analytics-core` 里，P1 不直接上 `ch-go`。先用 `clickhouse-go/v2 PrepareBatch`，因为它已经足够接近原生批量写入，同时复杂度更低。只有压测证明 `PrepareBatch` 成为瓶颈，才评估切到 `ch-go`。

## Q：目前 xwl_bi 的批量插入用的是哪种？

A：当前 xwl_bi 的事件明细写入已经使用 `clickhouse-go/v2` 的 native batch，不是 GORM，也不是直接使用 `ch-go`。

代码证据：

- `C:/Users/admin/Documents/src/xwl_bi/go.mod` 直接依赖 `github.com/ClickHouse/clickhouse-go/v2 v2.10.0`。
- `C:/Users/admin/Documents/src/xwl_bi/engine/db/clickhouse_native.go` 定义 `ClickHouseNative driver.Conn`。
- `C:/Users/admin/Documents/src/xwl_bi/application/init.go` 使用 `clickhouse.Open(...)` 初始化 native 连接，并开启 LZ4 压缩。
- `C:/Users/admin/Documents/src/xwl_bi/platform-basic-libs/service/consumer_data/reportdata2ck.go` 中 `prepareBatch` 调用 `db.ClickHouseNative.PrepareBatch(ctx, query)`。
- 同一个文件的 `flushTableRows` 使用 `batch.Append(row...)` 逐行追加，最后 `batch.Send()` 提交。
- `go.sum` 里能看到 `github.com/ClickHouse/ch-go v0.52.1`，但它是 `clickhouse-go/v2` 带来的间接依赖，不是 xwl_bi 业务代码直接调用的写入库。

因此 `analytics-core` 应该继承 xwl_bi 当前这条经验：高吞吐事件明细写入默认用 `clickhouse-go/v2 PrepareBatch`，不要退回 ORM batch insert。

参考：

- GORM Create / CreateInBatches：`https://gorm.io/docs/create.html`
- GORM ClickHouse driver：`https://github.com/go-gorm/clickhouse`
- ClickHouse Go driver batch insert：`https://clickhouse.com/docs/integrations/language-clients/go/clickhouse-api`
- ClickHouse Go client overview：`https://clickhouse.com/docs/en/integrations/go`
- ch-go：`https://github.com/ClickHouse/ch-go`

## Q：Funnel / Retention 要直接迁移 SQL 吗？

A：不建议直接完整迁移。推荐做法是：

1. 保留 xwl_bi 中 ClickHouse 查询思路，例如 `windowFunnel`。
2. 重写字段命名和表模型。
3. 把时间范围、事件条件、用户条件、属性过滤统一放进 query builder。
4. 为每类查询补 SQL 生成测试和结果校验。

这样既不丢掉 xwl_bi 的分析经验，也不会把旧结构原样带进 `analytics-core`。
