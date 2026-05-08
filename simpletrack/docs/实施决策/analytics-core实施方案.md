# analytics-core 实施方案

> 状态：已确定 P1 执行，模块设计持续评审  
> 最近更新：2026-05-08
> 来源：基于 xwl_bi 本地代码的 analyze + code-review 梳理，并结合 Umami、Litlyx 两个参考产品的调研资产。`references/xwl_bi-backend/` 主要作为后端架构设计参考，不作为代码搬运来源。

## 结论

P1 新建独立业务无关仓库 `analytics-core`，从 xwl_bi 抽取分析数据面核心。它不是 xwl_bi 整仓改名，也不是 SimpleTrack 私有业务层，更不是一个长期独立运行的 SimpleTrack 业务服务。

`analytics-core` 只负责采集、事件、元数据、实时写入、查询聚合和分析模型，不负责定价、团队、订阅、账单、onboarding、产品官网和企业控制台。

后续关系是：`analytics-core` 逐步反向支撑 xwl_bi，作为 SimpleTrack 的分析数据面核心库，也预留给 AppTrack 或其他行为分析产品复用。因此命名必须保持通用，不带具体业务含义；外部 Go 服务通过根目录公共包引用它。

P1 先落地最小可用链路：

1. `collect` 接收事件。
2. `EventBus` 抽象队列。
3. `RedisStreamBus` 先行。
4. `KafkaBus` 保留。
5. ClickHouse 写入事件和实时表。
6. 提供 Realtime、Raw Events / Events、Goal 最小查询能力。
7. 预留漏斗、留存、路径、LTV、归因、分群、会话、事件属性、用户属性的模块边界。

P1 的 `visit_id` 已改为长期方案：在 collect 写入前确定、入库存储、查询直接读取。旧的 readback 临时派生只作为历史过渡，不再作为后续实现口径。

## 设计目标

- **业务无关**：不出现 SimpleTrack、AppTrack、xwl_bi 的套餐、团队、订阅、账单、客户私有化部署等控制面概念。
- **命名专业化**：剔除 `xwl_` 变量、函数、包、表和业务命名前缀，改成分析领域命名。
- **高性能优先**：保留 ClickHouse、Redis、Kafka 的高吞吐路线。
- **低运维起步**：P1 先用 Redis Stream 替代 Kafka，降低早期部署复杂度。
- **可替换边界**：MySQL、Redis、Kafka、ClickHouse 都通过接口或 adapter 隔离。
- **可被多产品引用**：xwl_bi、SimpleTrack、AppTrack 都通过通用 `tenant / project / source` 语义接入，不反向污染核心仓库。
- **Go library 优先**：公共能力放在根目录包，供 `simpletrack-anaysitics-service` 通过 Go module 引用；`internal` 只保留真正不该外部依赖的实现细节。
- **统一查询构建**：从 `sqlx` 迁移到 GORM 最新稳定版本，统一使用 GORM 的 SQL Builder / Raw / Clauses / Scopes 能力承接查询构建；ClickHouse 高吞吐事件写入不走 ORM 热路径，优先使用原生 batch writer。

## 非目标

- 不复用 xwl_bi 旧 Vue2 后台界面。
- 不迁移 xwl_bi 旧菜单、旧权限、旧业务后台叙事。
- 不在 `analytics-core` 内实现登录、组织、订阅、账单、Admin、邮件。
- 不在 `analytics-core` 内托管 Browser SDK 或承担 write key、domain allowlist、CORS、quota 的产品运行时配置生命周期。
- 不把 `analytics-core` 做成长期产品化 `cmd/server`；SimpleTrack 运行时服务由 `simpletrack-anaysitics-service` 承担。
- 不在 P1 产品层开放全量漏斗、留存、路径、归因页面。
- 不把 Kafka 作为 P1 必选运行依赖。

## xwl_bi 代码证据

| 发现 | 证据 | 判断 |
| --- | --- | --- |
| xwl_bi 已经是分析型 Go 服务 | `C:/Users/admin/Documents/src/xwl_bi/go.mod:1` module 为 `github.com/1340691923/xwl_bi`，依赖 ClickHouse、Sarama Kafka、MySQL、Redis、Fiber | 技术栈适合作为抽取来源，但命名和模块边界需要重构 |
| 本地参考快照已进入父仓 | `references/xwl_bi-backend/` 保留后端源码和关键文档，不包含旧 Vue2 前端、日志和二进制 | 主要参考模块边界、启动装配、消费链路、ClickHouse 写入/查询分层、元数据流转和分析服务拆分，不直接照搬旧业务代码 |
| 启动层强绑定 Kafka、Redis、MySQL、ClickHouse | `cmd/report_server/main.go:58` 到 `63` 初始化 Kafka sync/async producer、Redis、MySQL、ClickHouse | 需要把全局初始化改成依赖注入或 adapter 装配 |
| 采集入口已经有较清晰 orchestration | `controller/report_ingress_handler.go:47` 到 `49` 描述 Resolve -> Build -> SendReportData 流程 | 可作为 `collect` 接口和 handler 的抽取参考 |
| 请求解码仍绑定旧字段 | `controller/report_request_decoder.go:70` 读取 `xwl_distinct_id`、`xwl_ip`、`xwl_part_date` | 需要定义新事件协议；旧字段只作为抽取参考，不做 legacy 兼容 |
| Producer 与 Kafka 耦合 | `platform-basic-libs/service/report/producer.go:15` 定义 `KafkaDataProducer`，`engine/db/kafka.go:9` 到 `12` 使用全局 Sarama producer/client | 需要抽出 `EventBus`，Kafka 只做 `KafkaBus` adapter |
| 已有多类分析命令 | `platform-basic-libs/service/analysis/interface.go:16` 到 `26` 包含 Funnel、Retention、Trace、Event、UserAttr、UserList、LTV、Attribution 等 | 适合映射到 `analysis/*` 模块，但查询层要重构 |
| 分析查询强依赖 xwl 表名和字段 | `analysis/event.go:178` 拼接 `xwl_event`，`analysis/funnel.go:118` 使用 ClickHouse `windowFunnel`，`analysis/retention.go:264` 到 `279` 使用 `xwl_distinct_id` 和 `xwl_part_date` | 查询思路可保留，SQL 需要参数化、统一表模型和命名 |
| sinker 已有完整 ETL 链路 | `cmd/sinker/internal/runner/report_handler.go:139` 描述 context extraction、geo enrich、metric parse、ensure columns、metadata、status、metric batch | 可作为 ingestion pipeline 参考，但要拆小接口 |
| realtime 链路轻量 | `cmd/sinker/internal/runner/realtime_handler.go:12` 说明不做动态补列和复杂校验，只快速入批并 ack | 可作为 P1 Realtime 写入路径参考 |
| ClickHouse 初始化仍有旧表语义 | `cmd/init_app/ck/init.go:30`、`70` 创建 acceptance status 和 realtime warehousing 表，包含 `xwl_kafka_offset` | P1 可保留 status/realtime 思路，但表名字段要重命名 |
| HTTP 入口分裂且上报路由依赖偏旧 | `router/index.go` 使用 Fiber 做后台控制面；`cmd/report_server/runtime.go` 使用 fasthttp + `buaazp/fasthttprouter` 做上报入口；`go.mod` 中 `fasthttprouter` 为 v0.1.1 | 只参考 xwl_bi 的 collect 启动装配和中间件编排；`analytics-core` / `simpletrack-anaysitics-service` 当前采用 Fiber v3，不复用低活跃的 `fasthttprouter` |

## code-review 风险结论

| 风险 | 严重度 | 说明 | 处理方式 |
| --- | --- | --- | --- |
| 全局数据库和队列变量耦合 | 高 | xwl_bi 多处直接使用 `db.ClickHouseSqlx`、`db.KafkaSyncProducer`、`RedisPool` | `analytics-core` 必须采用 `Store`、`EventBus`、`MetadataStore` 等接口注入 |
| Kafka 与采集链路耦合 | 高 | 当前 producer/sinker 均围绕 Kafka topic 和 Sarama 实现 | P1 抽象 `EventBus`，Redis Stream 先行，KafkaBus 保留 |
| 旧字段和旧表名外溢 | 高 | `xwl_distinct_id`、`xwl_part_date`、`xwl_part_event`、`xwl_event{appid}` 贯穿采集和分析 | 不做 legacy mapper；新协议直接使用 `tenant_id`、`project_id`、`source_id`、`distinct_id`、`event_time`、`event_name` |
| 动态 SQL 与动态表名较多 | 中高 | Funnel、Retention、Event 查询拼接表名和条件 | 统一通过 GORM SQL Builder / Raw / Clauses / Scopes 封装 query builder，限制字段白名单，补查询测试 |
| ETL pipeline 复杂度高 | 中 | report handler 一次处理 context、geo、column、metadata、status、metric batch | 拆成 pipeline stage，P1 只保留必要 stage |
| Redis Stream 缺失 | 中 | xwl_bi 当前以 Kafka 为主，没有可直接复用的 Redis Stream bus | 新写 `RedisStreamBus`，接口与 KafkaBus 保持一致 |
| 旧 UI 和控制面容易污染核心 | 中 | xwl_bi 有旧后台、权限和业务模型 | 只抽取数据面，不迁移 UI、菜单和业务后台 |

## 技术依赖调整

| 依赖 | 决策 | 原因 |
| --- | --- | --- |
| MySQL ORM / SQL Builder | 使用 GORM v2 体系，替代 `sqlx` | GORM 生态、文档、插件和维护活跃度更好；官方文档已支持 Generics API 和 SQL Builder；`sqlx` 不再作为 `analytics-core` 的默认数据库访问层 |
| Squirrel | 不引入 | `github.com/Masterminds/squirrel` 更新节奏偏慢；GORM 的 Raw、Named Argument、Clauses、Scopes、Generics API 已能覆盖当前 query builder 需求 |
| ClickHouse 查询 | 统一走 GORM query builder / Raw / Scopes 出口 | Events、Realtime、Funnels、Retention、Segments 等查询必须共用字段白名单、时间范围、权限边界和属性过滤 |
| ClickHouse 事件写入 | 高吞吐写入优先用 `clickhouse-go/v2` 原生 batch writer | GORM 支持 `CreateInBatches`，但事件明细是写入热路径；xwl_bi 既有调优经验也指向原生批量插入更适合高频 ClickHouse 写入 |
| Redis | P1 使用 `redis/redis-stack:latest` 容器镜像 | Redis Stack 方便本地开发和后续扩展，P1 先用 Redis Stream 承接轻量事件队列 |
| Kafka | 保留 `KafkaBus` | Kafka 仍是高吞吐事件流路线，但不是 P1 必选运行依赖 |
| HTTP API | 使用 Fiber v3，不沿用 xwl_bi 的 `buaazp/fasthttprouter` 路由层 | Fiber v3 当前已作为 collect HTTP 适配层和运行时服务入口；`collect.Handler`、EventBus、ingestion、storage 仍保持框架无关 |

GORM 参考：

- GORM v2 的 import path 是 `gorm.io/gorm`。
- Go module release 号仍显示为 `v1.x.x`，当前可见版本为 `gorm.io/gorm v1.31.1`；这不等于老 GORM v1。
- 老 GORM v1 是 `github.com/jinzhu/gorm`，`analytics-core` 不使用这个旧包。
- GORM Generics API：`https://gorm.io/docs/the_generics_way.html`
- GORM SQL Builder：`https://gorm.io/docs/sql_builder.html`
- GORM Create / CreateInBatches：`https://gorm.io/docs/create.html`
- GORM ClickHouse driver：`https://github.com/go-gorm/clickhouse`
- ClickHouse Go driver batch insert：`https://clickhouse.com/docs/integrations/language-clients/go/clickhouse-api`

写入策略：

- `EventQueryBuilder`：负责生成 ClickHouse 查询 SQL，底层使用 GORM Raw / SQL Builder / Scopes；当前已落地 `storage.EventQueryBuilder` 契约和 ClickHouse/GORM dry-run query plan builder，先覆盖 Events 与 Realtime。
- `EventWriter`：负责 ClickHouse 事件明细写入，P1 默认使用 `clickhouse-go/v2 PrepareBatch`；当前已落地 native batch `BatchWriter`、`EventWriteGuard` 幂等接口边界和 GORM/MySQL `IngestionStatusGuard`。
- `EventWriter` 必须保留压测口径；当前已对比 GORM `CreateInBatches` 与 `clickhouse-go/v2 PrepareBatch`，结论是 GORM 可作为低频管理写入或对照路径，但事件热路径继续优先 native `PrepareBatch`。
- 不在 handler 或 analysis 模块里直接散落原生 SQL；即使用原生 batch writer，也必须藏在 ClickHouse storage adapter 内。

## 目标模块边界

建议目录草案：

```text
analytics-core/
  contracts/
  collect/
    httpapi/
  eventbus/
    direct/
    redisstream/
    kafka/
  ingestion/
  storage/
    clickhouse/
    mysql/
  internal/
    analysis/
    e2e/
    metadata/
```

目录原则：

- `collect` 只处理请求字段校验、事件标准化和可替换 pre-queue stage；`collect/httpapi` 是可选 Fiber 适配器，不处理 SimpleTrack 业务鉴权。
- `eventbus` 屏蔽 Direct、Redis Stream、Kafka 差异。
- `ingestion` 处理消费、入库和 ack/nack 语义，不拥有 HTTP 或 SaaS 配置。
- `storage` 只封装外部依赖，不放业务分析逻辑；ClickHouse query 和 writer 分开，避免查询 builder 与高吞吐写入互相污染。
- `internal/analysis` 只暴露业务无关分析能力。
- `contracts` 放 xwl_bi、SimpleTrack、AppTrack 或其他上层产品可依赖的稳定契约。

## EventBus 草案

```go
type EventBus interface {
    Publish(ctx context.Context, event EventEnvelope) error
    Subscribe(ctx context.Context, group ConsumerGroup, handler EventHandler) error
}

type EventEnvelope struct {
    ID          string
    TenantID    string
    ProjectID   string
    SourceID    string
    SourceType  string
    EventName   string
    DistinctID  string
    SessionID   string
    VisitID     string
    EventTime   time.Time
    ReceivedAt  time.Time
    Properties  map[string]any
    UserProps    map[string]any
    Source      string
}
```

实现策略：

- `DirectBus`：本地开发、单进程测试、最小 demo。
- `RedisStreamBus`：P1 默认实现，降低部署复杂度。
- `KafkaBus`：保留 xwl_bi 高吞吐路线，后续事件量上来后启用。

## 命名映射规则

| xwl_bi 旧命名 | analytics-core 新命名 | 说明 |
| --- | --- | --- |
| `appid` / `table_id` | `source_id`，必要时映射到 `project_id` | 核心库不关心是网站、App、后台服务还是旧 xwl_bi 应用；统一称为数据源 |
| `xwl_distinct_id` | `distinct_id` | 访客或用户的稳定标识 |
| 旧会话或访问派生字段 | `session_id` / `visit_id` | `session_id` 保留 SDK/服务端会话来源语义，`visit_id` 是分析口径里的 canonical visit key |
| `xwl_part_event` | `event_name` | 事件名称 |
| `xwl_part_date` | `event_time` | 客户端事件时间 |
| `xwl_server_time` | `received_at` | 服务端接收或消费时间 |
| `xwl_client_date` | `client_date` | 客户端日期派生字段 |
| `xwl_event{appid}` | `events_${tenant/project/source}` 物理表，统一逻辑名仍叫 `events` | 已确定采用方案 B；动态物理表名只允许出现在 ClickHouse adapter 内，对上层隐藏 |
| `xwl_real_time_warehousing` | `realtime_events` | 实时验收和最近事件 |
| `xwl_acceptance_status` | `ingestion_status` | 写入验收、失败原因、消费 checkpoint |
| `xwl_kafka_offset` | 不进入事件协议 | consumer offset / checkpoint 是内部消费进度，不放进业务事件模型 |

## ClickHouse 表策略

当前已确定：**直接采用方案 B，一步到位按 project/source 做物理表策略，但对上层仍暴露统一 `events` 逻辑模型。**

这不是回到 xwl_bi 那种把动态表名散落在业务代码里的做法。核心区别是：

- xwl_bi 主要按 `appid` 一维分表，业务代码容易感知 `xwl_event{appid}`。
- `analytics-core` 使用 `tenant_id / project_id / source_id` 三层业务无关语义。
- 物理表名、建表、迁移、路由和跨表查询只允许存在于 ClickHouse adapter / query builder 内。
- Events、Realtime、Funnels、Retention、Segments 等上层分析模块只面对统一逻辑模型。

### 方案 A：统一 `events` 表

做法：

- 所有产品、项目、数据源的事件都进入同一张逻辑事件表。
- 表内固定保留 `tenant_id`、`project_id`、`source_id`、`source_type`、`event_name`、`distinct_id`、`session_id`、`event_time`、`received_at` 等字段。
- 事件属性和用户属性使用 JSON / Map / 动态列策略承接。
- ClickHouse 通过 `PARTITION BY`、`ORDER BY`、索引、projection、materialized view 等方式优化查询。

优点：

- 最适合 `analytics-core` 的业务无关定位。
- Query Builder 统一，跨 source / project 查询更容易。
- 不会像 `xwl_event{appid}` 一样产生大量动态表名。
- 对 SimpleTrack、AppTrack、xwl_bi 反向接入都更干净。

风险：

- 表会快速变大，必须认真设计分区、排序键、TTL 和冷热策略。
- 不同 source 的事件属性差异较大时，需要较好的 metadata 和属性字典治理。
- 极高吞吐场景可能需要后续拆分或集群分片。

当前结论：

- 不作为 P1 默认方案。
- 只保留为本地 demo、测试环境或极低流量部署的可选简化模式。
- 如果使用方案 A，也必须复用同一套 `EventQueryBuilder` 和 `EventWriter` 接口，不能形成第二套查询逻辑。

### 方案 B：按 project/source 物理分表

做法：

- 为不同 `tenant_id / project_id / source_id` 或其稳定 hash 创建独立事件表，表名由 ClickHouse adapter 生成。
- 表内仍固定保留 `tenant_id`、`project_id`、`source_id`、`source_type`、`event_id`、`event_name`、`distinct_id`、`session_id`、`event_time`、`received_at` 等标准字段。
- `EventQueryBuilder` 根据查询范围解析到一个或多个物理表，再生成 ClickHouse SQL。
- `EventWriter` 根据事件三元组路由到目标物理表，并使用原生 batch writer 批量写入。

优点：

- 物理隔离更强。
- 单表规模较小。
- 某些大客户或超高吞吐 source 可单独调优。
- 与 xwl_bi 既有 `appid` 分表经验更接近，迁移和性能调优路径更可控。

风险：

- 动态表名会让 query builder、迁移、权限和测试复杂很多。
- 元数据、属性列、DDL 变更会分散到大量表。
- 多项目、多 source 聚合查询更麻烦。
- 容易把 xwl_bi 的历史结构问题带进 `analytics-core`。

已确定约束：

- P1 直接采用方案 B。
- 动态物理表名只能由 `TableRouter` 生成，不能在 handler、analysis 或上层产品代码里拼接。
- 所有查询仍统一走 `EventQueryBuilder`，不允许每个分析模块自己决定表名。
- 所有写入仍统一走 `EventWriter`，默认使用 `clickhouse-go/v2 PrepareBatch`。
- 需要为跨 source 查询设计 fan-out + merge 策略，但 P1 产品层先以单 source / 单 project 查询为主。
- 需要为 DDL 变更建立 migration plan，确保新增标准字段或索引时能批量应用到所有物理表。
- 表命名不得包含 SimpleTrack、xwl_bi、AppTrack 业务词，只使用通用 `tenant/project/source` 或其 hash。

建议表路由形态：

```text
logical table: events
physical table: events_{tenant_hash}_{project_hash}_{source_hash}
```

如果担心表名过长或暴露业务 ID，可以统一使用短 hash，并在 MySQL metadata 中维护映射。

## ClickHouse 读侧优化方案取舍

P1.5 的目标不是把所有 ClickHouse 手段一次性上完，而是先把读侧长期路线定清楚。四个候选方案不是互斥关系，而是分层能力：

- 属性治理和 query plan 约束是底座。
- projection 适合单表明细路径的局部加速。
- materialized view 适合稳定派生结果的持续写入。
- 小时聚合表适合趋势图、Dashboard 和高频固定口径指标。

### 方案 1：ClickHouse Projection

是什么：

- 在同一张明细表上，给 ClickHouse 增加额外投影，提前组织局部排序、过滤或聚合路径。
- 适合不改变主事实表语义、但希望缩短某些固定读路径的场景。

适合什么：

- `Realtime`。
- `Events` 的常用明细筛选。
- 以单表读为主、字段组合相对稳定的短查询。

优点：

- 对上层逻辑侵入最小。
- 不需要额外维护一套独立派生表的写入同步逻辑。
- 对明细读路径有机会获得较低改造成本的收益。

风险：

- 收益强依赖查询形状，查询稍微变化就可能失效。
- 不适合承载复杂指标语义，也不适合替代分析层的长期建模。
- 需要更谨慎地观察 ClickHouse 版本行为和执行计划。

结论：

- Projection 更像“明细读路径加速器”，不是分析模型本身。
- 适合后续只对少数热点查询选择性使用，不适合当作 P1.5 唯一方案。

### 方案 2：Materialized View

是什么：

- 通过 ClickHouse 物化视图把事实事件流持续投影到派生表。
- 适合把稳定查询口径提前固化成可重复使用的派生数据集。

适合什么：

- 稳定的 Breakdown / Compare。
- 某些固定维度的事件计数、活跃统计、Top pages、Top events。
- 需要持续写入但读取频繁的中间结果。

优点：

- 查询时不用每次重算，适合读多写少或读写比高的稳定口径。
- 能把复杂聚合从在线查询中移出。
- 适合为 Dashboard 和固定指标页面提供长期支撑。

风险：

- 维护成本高于纯 query plan 约束。
- 口径变化时需要同步调整派生表逻辑。
- 如果指标定义还在变，物化视图会把不稳定口径提前固化，后续返工成本高。

结论：

- MV 适合“口径稳定之后的长期承载层”，不适合一开始就替代所有查询。

### 方案 3：小时聚合表

是什么：

- 把事件按小时、项目、来源、事件名、属性维度等口径做预聚合。
- 通常是 Dashboard 和趋势图最直接的提速方案。

适合什么：

- 时间序列趋势图。
- 固定维度的事件计数、UV、活跃度、基础漏斗前置统计。
- 产品首页、概览页、看板类高频查询。

优点：

- 对趋势类查询收益稳定。
- 语义最接近产品展示层的固定口径。
- 能显著降低明细大扫表压力。

风险：

- 不适合明细列表。
- 维度一多，聚合表会快速膨胀。
- 如果没有明确的 query pattern，容易做出一堆用不上的聚合表。

结论：

- 小时聚合表是长期分析产品最稳的读侧能力之一，但必须建立在稳定指标口径之上。

### 方案 4：只做属性治理 + Query Plan 约束

是什么：

- 不新增 MV、projection 或聚合表，只先把属性字典、过滤白名单、排序、分页、字段选择和 query builder 约束收紧。
- 让读侧先保持“正确、可控、可回归”，再决定是否加速。

适合什么：

- 当前还在快速迭代阶段。
- 查询模式没有完全稳定。
- 还不想提前引入 ClickHouse 派生层的维护成本。

优点：

- 最适合作为长期方案的第一步。
- 改造成本低，最容易确保 `visit_id`、`property_filter`、排序和分页契约不退化。
- 便于把 SQL 复杂度牢牢锁在 `analytics-core/storage`。

风险：

- 单靠这层不会解决所有读侧性能问题。
- 数据量继续上来后，Dashboard 和 Breakdown 还是会需要真正的预聚合层。

结论：

- 这是必须先做的底座，但不是最终答案。

## 推荐长期方案

长期收益最高的顺序不是“现在就选一个重方案”，而是分层推进：

1. 先做属性治理和 query plan 约束，把字段白名单、属性过滤、排序、分页、表路由和执行边界收紧。
2. 保持当前事实表 + `EventQueryBuilder` + `EventReader` 作为唯一读侧入口，先让 Realtime / Events 稳定可回归。
3. 当发现少数明细路径反复成为热点时，再对这些路径选择性引入 projection。
4. 当指标口径稳定、且页面读多写少时，再把稳定指标下沉到 materialized view 或小时聚合表。
5. Funnels / Journeys / Retention 这类更复杂分析，单独按 query pattern 评估，不和 Realtime / Events 混成一锅。

这条路的核心判断是：

- 属性治理是长期底座。
- projection 是局部提速。
- materialized view 和小时聚合表是稳定指标的长期承载层。
- 不把它们当成互相替代的单选题，而是当成不同层的长期组合。

## ClickHouse 读侧长期规范

本节是后续 `analytics-core` 读侧实现的约束规范，不只是当前方案说明。凡是修改 Events、Realtime、Breakdown、Dashboard、Goal、Funnels、Journeys、Retention 或其他 ClickHouse 查询能力，都必须遵守。

`analytics-core` 仓库内的同步规范见 `src/analytics-core/docs/read-side-optimization-policy.md`。该文件把必须提交的 evidence、benchmark、适用边界和验证命令落到子仓，后续新增 projection、materialized view 或小时聚合表时必须同时满足本节和子仓策略文档。

### 入口规范

- `EventQueryBuilder` 是唯一 query plan 构建入口。
- `EventReader` 是唯一 query plan 执行入口。
- `readSidePolicy` 是 `analytics-core/storage/clickhouse` 内部的读侧 guardrail 容器，负责 query limit、filter cap 和 property allowlist，不向外暴露。
- `simpletrack-anaysitics-service` 只能把经过 runtime source 校验后的查询参数传入 `analytics-core`，不能拼接 ClickHouse SQL。
- `simpletrack-saas` 只能通过服务端 readback helper 调用内部查询 API，不能直接连接 ClickHouse，也不能感知动态物理表名。
- `TableRouter` 是唯一物理表路由入口，handler、service、产品页面和 analysis 模块都不能自己拼 `events_*` 表名。

### 分层推进规范

读侧优化必须按以下顺序推进：

1. **属性治理和 query plan 约束先行**：先稳定字段白名单、属性白名单、operator enum、排序字段、分页上限、时间窗、表路由和参数绑定。
2. **明细查询先保持事实表读取**：Realtime 和 Events 默认读取事实事件表，不提前强制走聚合表。
3. **projection 只用于热点明细路径**：只有当某类明细查询反复成为热点，并且查询形状稳定时，才允许为对应物理表增加 projection。
4. **materialized view 只用于稳定派生口径**：只有当指标定义稳定、读多写少、且查询已经不适合实时聚合时，才允许新增 MV。
5. **小时聚合表只用于趋势和固定指标**：Dashboard、Trend、Goal count、Top events、Top pages 等固定时间粒度查询可进入小时聚合表；Raw Events 明细列表不得依赖小时聚合表。
6. **复杂分析单独建模**：Funnels、Journeys、Retention、Attribution 不能复用 Realtime / Events 的临时优化结构硬凑，应按自己的 query pattern 单独设计。

### 禁止事项

- 禁止在 `analytics-service` handler、`simpletrack-saas` 页面或 API helper 中出现 ClickHouse SQL。
- 禁止绕过 `EventQueryBuilder` 直接在业务模块中写动态 SQL。
- 禁止绕过 `EventReader` 直接扫描 ClickHouse row 并返回产品 DTO。
- 禁止为了短期性能把 `visit_id`、`property_filter`、排序、分页或 allowlist 语义回退。
- 禁止在没有 query pattern、压测数据和回归测试的情况下提前新增 MV、projection 或小时聚合表。
- 禁止把 projection / MV / 小时聚合表作为彼此替代的单选题；它们属于不同层的优化手段。

### 引入门槛

新增 projection、materialized view 或小时聚合表前，必须先补齐这些材料：

- 目标查询属于哪个产品能力：Realtime、Events、Dashboard、Breakdown、Goal、Funnels、Journeys、Retention 等。
- 当前查询的 query plan、过滤字段、排序字段、时间范围、分页方式和预期数据量级。
- 为什么属性治理和 query plan 约束不足以解决问题。
- 为什么选择 projection、MV 或小时聚合表，而不是另外两个。
- 对方案 B 多物理表的影响：是否每个 `tenant/project/source` 物理表都需要同构结构，如何批量创建和校验。
- 对写入链路的影响：是否增加写入延迟、写放大、失败恢复或数据一致性风险。
- 回归测试计划：至少覆盖正常查询、非法字段、非法属性、分页、排序、`visit_id` 和跨 source 边界。

### 验收规范

读侧优化完成后必须证明：

- 原有 Realtime / Events 查询契约不退化。
- `property_filter` 仍然走 source-scoped allowlist 和参数绑定。
- `visit_id` 仍然读取存储字段，不回退到 readback 派生。
- `EventQueryPlan.QueryEvidence()` 必须能说明当前查询家族、读路径、优化策略、effective limit、offset、time lower/upper bound、time window、属性表参与情况、过滤数量、value-free property filter shape 和排序口径，不能只靠读 SQL 字符串判断，也不能把属性过滤值暴露到 evidence。
- ClickHouse 物理结构只存在于 `analytics-core/storage/clickhouse` adapter 内。
- 相关实施决策 README、分阶段计划和 `docs/analytics-source-reading/` 已同步。

### 当前真实 ClickHouse reader 基线

2026-05-08 已按 `src/analytics-core/docs/read-side-optimization-policy.md` 跑过一次真实 ClickHouse reader benchmark：

```powershell
$env:ANALYTICS_CORE_CLICKHOUSE_BENCH='1'
go test ./internal/e2e -run '^$' -bench 'BenchmarkEventReaderClickHouseExecution' -benchmem -count=3
```

结果摘要：

- `low_realtime`：约 8.6-10.4ms/op。
- `medium_events_scalar`：约 8.6-9.7ms/op。
- `high_events_property`：约 15.2-16.3ms/op。

随后用 `ANALYTICS_CORE_CLICKHOUSE_BENCH_ROWS=100000` 做 100k 行 pressure run：

- `low_realtime`：约 11.6-13.5ms/op。
- `medium_events_scalar`：约 10.7-11.6ms/op。
- `high_events_property`：约 31.3-34.2ms/op。

在 `analytics-core` commit `4393bbd` 补齐 value-free property filter shape evidence 后，同日再次做 100k 行复测：

- `low_realtime`：约 12.3-18.1ms/op，第一轮偏高但后两轮回到原观察区。
- `medium_events_scalar`：约 10.4-11.2ms/op。
- `high_events_property`：约 29.8-32.8ms/op。

这份基线和复测只作为后续对比依据，不触发立即新增 projection、materialized view 或小时聚合表。下一条重点观察候选仍是 typed property 过滤读路径；当前 query plan、value-free property filter shape 和 ClickHouse explain 已补齐，但是否进入物理结构评审，仍要继续看更大数据量、稳定 query pattern 和回归计划。完整记录见 `docs/analytics-source-reading/read-side-benchmark-baseline.md`。

补充结对审查后，`analytics-core` 已按 `$ai-slop-cleaner` 的小范围流程修复 query evidence 快照边界：先用 `go test ./storage ./storage/clickhouse` 锁住行为，再在 `NewEventQueryPlan` 和 `QueryEvidence()` 两个边界复制 `PropertyFilters` slice，最后补 `ExampleEventQueryPlan_QueryEvidence` 和 `TestEventQueryPlanQueryEvidenceClonesPropertyFilters`。这条规则的长期含义是：query evidence 可以被 service 暴露和操作员阅读，但调用方不能通过修改返回 slice 反向污染查询计划内保存的 canonical evidence。

同日已补第一轮 ClickHouse explain 证据：

```powershell
$env:ANALYTICS_CORE_CLICKHOUSE_BENCH='1'
$env:ANALYTICS_CORE_CLICKHOUSE_BENCH_ROWS='100000'
go test ./internal/e2e -run TestEventReaderClickHouseExplain -count=1 -v
```

当前观察到：

- `low_realtime` 和 `medium_events_scalar` 仍是 routed fact table 的 `ReadFromMergeTree` 主键路径。
- `high_events_property` 的 query evidence 只暴露属性过滤形状，不暴露实际属性值；explain 已出现 `CreatingSets`，并且主查询上出现 3 个 `event_id in ... set` 条件。
- 同一个高属性过滤场景的主键条件已经包含 `visit_id`，但在本次 100k 行数据量下仍是 `Granules: 13/13`。

这说明 typed property 过滤读路径值得继续观察，但仍然只是“是否新增物理结构”的证据补齐，不是立即上 projection、materialized view 或小时聚合表的触发器。下一步仍应优先保持 direct fact table，并继续做属性治理、query plan 约束、更大数据量 benchmark 和回归计划。

## P1 执行步骤

### Step 1：新建仓库骨架

交付：

- 新建 `analytics-core` 独立仓库。
- 建立 Go module、目录结构、lint/test 基线。
- 不引入旧 Vue2、旧后台、旧菜单。

验收：

- 仓库名、包名、函数名、变量名不带 `simpletrack` 或 `xwl`。
- 能运行最小单元测试。

### Step 2：事件协议和新字段标准化

交付：

- 定义标准 `EventEnvelope`。
- 定义 `tenant_id`、`project_id`、`source_id`、`source_type`、`event_name`、`event_time`、`distinct_id`、`session_id`、`visit_id`、`properties`、`user_properties` 等字段。
- 定义 event name、source id、timestamp、distinct id、properties 校验规则。
- 不提供 xwl_bi legacy 字段兼容。xwl_bi 后续迁移时应改为写入新协议。
- 当前已落地 `collect.Normalize`，负责把 collect 请求标准化为 `EventEnvelope`，并校验事件 ID、租户、项目、数据源、事件名、用户标识和时间戳。
- 当前已落地 `collect.Handler`，这里的 Handler 指 `collect/handler.go` 中的事件上报核心处理器，不是 HTTP 路由 handler；它负责调用 `Normalize` 并把标准化后的事件发布到 `EventBus`。HTTP collect API 只做协议适配，不重复实现校验和发布逻辑。
- 当前已落地 `VisitResolverStage`，负责在 SDK 未传 `visit_id` 时，使用 server-only visit salt、30 分钟默认窗口和最终 `session_id` 派生 canonical analytics visit key；如果 SDK 已传合法 `visit_id`，则保留原值。
- 当前已落地 Fiber `POST /collect` 入口，负责 JSON 解码、HTTP 状态码和响应格式；`collect.Handler` 继续保持框架无关。

验收：

- 非法事件能返回明确错误。
- 新协议字段不带 `xwl_`、`simpletrack`、`apptrack` 等业务前缀。

### Step 3：EventBus 与 Redis Stream

交付：

- 定义 `EventBus` 接口。
- 实现 `DirectBus` 和 `RedisStreamBus`。
- 保留 `KafkaBus` adapter 目录和接口，后续从 Sarama producer/consumer 迁入。
- 本地 Redis 使用 `redis/redis-stack:latest` 容器镜像。
- Redis Stream 消费采用 pending 优先：写入成功才 ack，失败保留 pending 并优先重试。
- Redis Stream 支持 `MaxAttempts` 与 `DeadLetterStream`，达到上限后写入死信队列并 ack 原消息。
- `ingestion` processor 将重复事件写入视为成功处理，避免 at-least-once delivery 造成重复入库。

验收：

- collect 发送事件后能进入 Redis Stream。
- consumer group 能消费并 ack。
- 消费失败后能重新读取 pending 消息，并按 Redis consumer group pending metadata 计算 attempt。
- 达到 `MaxAttempts` 后能进入死信队列，原消息 pending 数归零。
- 重复事件写入能被视为成功消费，不触发无限重试。
- Redis Stream 与 KafkaBus 使用同一事件协议。

### Step 4：ClickHouse 写入和 Realtime

交付：

- 建立按 `tenant_id / project_id / source_id` 路由的事件物理表、实时事件表、ingestion status 表。
- ClickHouse 事件表和 `_properties` 表都包含 `visit_id`，并在写入前由 collect 阶段确定，不在读回放阶段临时补算。
- 当前已落地 ClickHouse `TableRouter`，按 tenant / project / source 生成稳定 hash 物理表名，对上层仍暴露统一 `events` 逻辑模型。
- 当前已落地 storage `EventWriter` 接口和 ClickHouse native batch `BatchWriter`，真实写入统一复用该边界。
- 当前已落地 GORM/MySQL `IngestionStatusGuard` 和 `ingestion_status` 表，按 `(tenant_id, project_id, source_id, event_id)` 做 `processing / inserted / failed` 状态占用、重复跳过、失败回滚和重试再占用。
- worker 消费队列并通过 `EventWriter` 写入 ClickHouse。
- `EventWriter` 默认使用 `clickhouse-go/v2 PrepareBatch` 原生批量写入，GORM batch insert 只作为压测对照或低频管理写入选项。
- 当前已明确 `ingestion.Processor` 是 P1 worker 边界：Redis/Kafka adapter 负责 ack/nack，Processor 负责调用 `storage.EventWriter`，后续真实 worker 入口不得复制一套消费写入逻辑。
- worker 后续运行时必须把 `EventWriteGuard` claim、ClickHouse `EventWriter` append、guard commit/rollback 和 Redis/Kafka ack 串成同一条可恢复链路，避免重复事件在数据库存两份。
- 当前已提供 Realtime 和 Raw Events / Events 的 query plan builder 与 ClickHouse/GORM query reader；`EventQueryBuilder` 支持 `visit_id` 白名单过滤，`EventReader` 读取存储中的真实 `visit_id`，不再依赖 readback 派生。

验收：

- 单条 pageview 能进入 Realtime。
- 自定义事件能在 Events 表查到。
- 写入失败能记录状态。
- 重复消费同一 `event_id` 不会生成两条 ClickHouse 明细事件。

### Step 5：元数据与最小 Goal

交付：

- 捕获事件名、事件属性、用户属性、站点元数据。
- 提供 Goal 最小定义和查询接口。

验收：

- 首次出现的事件名能进入元数据。
- Goal 能基于事件名返回是否有数据和基础计数。

### Step 6：分析模块预留

交付：

- 为 funnels、retention、paths、ltv、attribution、segments、sessions 建立模块边界。
- 从 xwl_bi 映射可保留查询思路和需要重写部分。

验收：

- P1 不开放完整页面，但接口边界不会阻塞 P2。
- ClickHouse 查询不再散落在业务 handler 里，统一进入基于 GORM SQL Builder / Raw / Clauses / Scopes 的查询构建层。

## Umami 与 Litlyx 的参考边界

SimpleTrack 已经有两个参考产品资产，`analytics-core` 需要吸收它们各自最有价值的部分。

| 参考 | 用在什么地方 | 不照搬什么 |
| --- | --- | --- |
| Umami | 分析对象体系：Realtime、Events、Filters、Segments、Funnels、Journeys、Retention、Attribution 的边界；事件字典和 distinct id 语义；先 Realtime/Events 后 Funnels/Retention 的阶段顺序 | 不照搬 Umami 的完整产品形态，也不在 P1 一次性开放所有高级页面 |
| Litlyx | 新用户首价值：短接入链路、Raw Events 验收、Product 骨架、Show test data、docs 引导和商业化入口 | 不把示例数据当真实验收，不把 Product 单页骨架变成长期大杂烩 |

落实到 P1：

- Realtime 参考 Umami，作为“数据有没有进来”的最快验收页。
- Raw Events 参考 Litlyx，作为“事件到底有没有入库”的排障入口。
- Product / Overview 可以借 Litlyx 的空态、示例态、真实态思路。
- Funnels、Journeys、Retention、Segments 的模型边界参考 Umami，但产品页放到 P2 以后。

## Umami 源码启发的优化落点

Umami 源码深解已经把 P1 数据管道拆成 tracker、collect、session/visit、事件写入、属性展开、ClickHouse schema、Realtime/Events 查询和过滤构建。`analytics-core` 吸收这些内容时不改变既有 Go 边界，只把实现经验落到可测试的 stage、adapter 和 query plan。

| 优化项 | Umami 证据 | `analytics-core` 落点 | 当前处理 |
| --- | --- | --- | --- |
| 事件属性与用户属性 | `event_data`、`session_data`、typed value | collect 阶段已落地属性 key、数量、标量类型、字符串长度和有限数字入口约束；storage 层已提供 `EventPropertyRecord` / `FlattenEventProperties` typed row 逻辑展开；ClickHouse `PropertyBatchWriter` 写入同源路由 `_properties` 表；`PropertyIndexingEventWriter` 已把属性索引组合进 ingestion 热路径；MySQL `property_indexing_status` 单独处理属性 checkpoint，failed 可原子 reclaim，processing ambiguous 不自动重试；真实 e2e 已验证写入、读取和属性过滤；P1.5 已新增 `PropertyCatalog` 契约、MySQL `property_catalog` adapter、`PropertyCatalogingEventWriter`、source-scoped `PropertyCatalogReader` 和 `simpletrack-anaysitics-service` 内部 `/v1/properties`，先记录并读取 selector/type/first_seen/last_seen，不记录会被重试放大的计数字段；`simpletrack-saas` Events filter builder 已读取 source-scoped 属性目录作为字段建议 | P1-002A 已完成；属性字典治理和 SaaS filter builder 接入已完成，ambiguous 恢复和 ClickHouse 去重/聚合优化放 P1.5/P2 |
| client info enrich | collect 入口补 IP、UA、browser、os、device、geo | P1-002B 已落地 collect `Stage`：UA/referrer 可进入 bounded properties，IP 只允许盐化为 `client.ip_hash`；browser / OS / device 通过可替换 `UserAgentParser` 派生，geo 通过可替换 `GeoResolver` 派生为 `geo.country`、`geo.region`、`geo.city`；`simpletrack-anaysitics-service` 可用 `ANALYTICS_SERVICE_GEOIP_MMDB_FILE` 装配离线 MaxMind mmdb；HTTP 默认不信任 forwarded IP，可信代理需显式配置；浏览器 SDK 已支持 opt-in DNT，并自动补 allowlisted UTM/click id | 已补 enrich 边界，禁止放入 ClickHouse writer；后续只评审更专业 UA parser 替换和 geo 数据文件部署 |
| bot/IP/internal traffic 过滤 | collect 入口做 bot/IP 判断 | P1-002B 第一版已落地 `TrafficFilterStage`：按 bot UA token、internal CIDR/IP 在 EventBus publish 前返回 `FilteredError`，HTTP 返回 accepted filtered 响应，不写入分析明细；DNT active 时浏览器 SDK 不发送也不持久化 identity | 后续评审 allow/deny 配置来源、产品 UI、internal traffic 和审计/采样策略 |
| session/visit resolver | source + id 或 IP/UA/salt 派生 session，visit 使用短窗口 | P1-002C 已从 readback 临时派生升级为写入前 resolver：`SessionResolverStage` 先补 `session_id`，`VisitResolverStage` 再按 server-only visit salt、30 分钟默认窗口和最终 session 派生缺失的 canonical `visit_id`；`visit_id` 已进入 `EventEnvelope`、ClickHouse event 表和 `_properties` 表；IP/UA 只能作为 transient hash 输入；浏览器 DNT opt-in 避免持久本地 identity | P1 已定稿 `visit_id` 持久化；salt 轮换、cookie/no-cookie、Sessions 专页和 retention 产品化放 P1.5/P2 |
| 查询白名单与过滤 | `FILTER_COLUMNS`、operator mapping、分页 | `EventQueryBuilder` 字段白名单、排序白名单、过滤 operator enum、分页上限和 typed property filter allowlist；属性过滤使用 ClickHouse tuple `IN` 半连接查询属性表，避免 correlated `EXISTS` 外层 alias 兼容问题；`simpletrack-saas` Events 页面现在也把 `event_name`、`distinct_id`、`limit`、`offset`、`sort_field`、`sort_direction` 做服务端归一化后再请求内部读回放；`simpletrack-anaysitics-service` 已把 runtime source config 的 `allowed_property_filters` 映射为 URL 编码 JSON `property_filter` 的服务层白名单，并把启动 source surface 下发给 ClickHouse query builder 兜底 | P1-002D 已完成，P1-005D 已补属性过滤入口，后续复杂查询继续复用 allowlist + 真实 ClickHouse e2e |
| Realtime/Events 验收 | Realtime 短窗口、Events 分页明细 | `EventReader` 读取 ClickHouse query plan 结果；e2e 入口已增加 Redis/MySQL/ClickHouse 冷启动 readiness 重试，避免 compose 刚启动时 native handshake EOF 误伤验收；Events 产品页使用额外读取一条的 `hasMore` 模型，不做总数查询；内部读回放 token 可用 `ANALYTICS_SERVICE_QUERY_TOKENS_JSON` 做短窗口轮换，并可附带 `id`、`not_before`、`expires_at` 供运行时拒绝过期/未来 token 和记录审计日志 | P1-002E 已完成，P1-005D 已补页面分页交互、query token 轮换和生命周期校验，后续作为回归入口 |
| Web tracker SDK | auto pageview、custom event、identify、performance | P1 已落地 SimpleTrack 浏览器 SDK，但已从 `analytics-core` 迁出；当前由 `simpletrack-anaysitics-service` 的 `/tracker.js` 静态交付，并通过 `data-write-key` 进入运行时 collect 服务 | P1-004 已完成；React/Next/Node/mobile SDK、CDN 版本化和 performance metrics 后续评审 |
| ClickHouse 读侧优化 | materialized view、小时聚合表、projection、typed 属性 | ClickHouse adapter 的属性治理、query plan 约束、聚合表和 projection 分层策略 | P1.5-001，长期路径先做属性治理 + query plan 约束，再按查询稳定度引入 MV / 小时聚合表 / projection |
| Performance metrics | LCP、INP、CLS、FCP、TTFB | 可作为事件类型或属性组进入协议扩展 | P2-001，P1 只预留承接能力 |

实现顺序：

1. P1-002E 已完成：pageview、自定义事件属性和 user properties 已能从 collect 进入 ClickHouse 并被 Realtime/Events 查询；冷启动 e2e readiness 已复验稳定。
2. P1-002A 已完成：`PropertyBatchWriter` 已通过 `PropertyIndexingEventWriter` 组合进 ingestion worker，属性跨表幂等使用 `property_indexing_status` guard；processing ambiguous 不自动恢复，后续作为 P1.5/P2 运维和 ClickHouse 去重策略评审项。
3. P1-002C 正在按长期方案收口：`visit_id` 已进入 collect 契约、ClickHouse schema、event/property writer、reader 和 query builder；`simpletrack-anaysitics-service` 负责装配 visit resolver，`simpletrack-saas` runtime source 输出 server-only `visit_salt` / `visit_window_seconds`。
4. P1-004 已完成并纠偏：浏览器 SDK 最短链路和 docs/quickstart 已改为 write key 接入，SDK 由 `simpletrack-anaysitics-service` 托管，不再属于 `analytics-core`；后续继续评审 geo、SDK 发布策略和多语言 SDK。
5. P1-005D 正在推进：内部 `/v1/realtime`、`/v1/events` 已由 `simpletrack-anaysitics-service` 读回放，SaaS 页面只走 server-side helper；Events 已补白名单筛选、排序、属性过滤和 `hasMore` 分页，内部 query token 不进入浏览器，并已支持服务端短窗口轮换 allowlist、结构化生命周期和轮换命中/拒绝审计日志。
6. P1 数据闭环稳定后，P1.5-001 先做属性治理和 query plan 约束；当前已补 `readSidePolicy`、`EventQueryEvidence`、`PropertyCatalog` 基础契约、MySQL catalog adapter、`PropertyCatalogingEventWriter`、source-scoped `PropertyCatalogReader`、`simpletrack-anaysitics-service` ingestion 运行时装配和内部 `/v1/properties` 属性目录读回；`simpletrack-saas` 已在服务端读取 `/v1/properties` 并把属性目录接入 Events filter builder，浏览器只拿到元数据建议，不拿内部 query token；`simpletrack-anaysitics-service` 的 readback 响应已开始透出 `query_evidence` 与 `pressure`，其中 `pressure` 只是 low / medium / high 的 triage 桶；`analytics-core` 已新增 builder-only read-side shape benchmark、真实 ClickHouse EventReader benchmark、真实 ClickHouse BatchWriter benchmark、GORM `CreateInBatches` 对照 benchmark、Redis Stream publish / subscribe+ack benchmark、`collect.Handler` 热路径 benchmark 和 `docs/read-side-optimization-policy.md`，固定 low realtime、medium scalar events、high property events 三类查询的 plan 构建/实际执行基线、当前单事件 native/GORM 写入对照基线、队列 publish/consume/ack 基线，以及 collect normalize / identity / client enrich 基线；当前 `EventQueryEvidence` 已进一步补齐 effective limit、offset、time lower/upper bound、time window 和 value-free property filter shape，并由 `simpletrack-anaysitics-service` 的 readback API / OpenAPI 透出；projection、materialized view 和小时聚合表只在 query evidence、benchmark 基线、热点路径、稳定指标口径和回归计划明确后逐步引入。

## 与上层产品的集成边界

SimpleTrack / AppTrack / xwl_bi 产品层负责：

- Workspace、Site、App、Member、Role 等具体业务对象。
- 登录、组织、订阅、账单、邮件、Admin。
- 产品官网、定价页、docs/quickstart、onboarding。
- 企业分析控制台页面。
- write key、domain allowlist、internal traffic、quota 等配置的创建、修改和展示。

`simpletrack-anaysitics-service` 负责：

- `/collect`、`/tracker.js`、CORS preflight 和运行时服务健康检查。
- 读取 SimpleTrack 控制面的 runtime source config。
- 本地可用 memory resolver；生产接入雏形使用 HTTP resolver 通过 bearer token 按 write key 读取 `simpletrack-saas` 的内部 runtime-source API，并用短 TTL 缓存降低热路径依赖；控制面 URL 默认必须是 HTTPS，本地 loopback HTTP 只能显式 opt-in。
- 执行 write key、Origin/domain allowlist、server-only privacy salts、internal traffic、bot 过滤和后续 quota runtime enforcement。
- 不信任客户端传入的 tenant/project/source/source_type，统一由控制面配置覆盖后调用 `analytics-core`。
- 装配 session / visit resolver：`session_salt`、`visit_salt`、`visit_window_seconds` 由控制面 runtime source config 提供，公开 write key 不参与派生 salt。
- 显式开启 ingestion 时，装配 Redis Stream consumer、MySQL checkpoint guard、ClickHouse native writer 和 typed property indexing；默认启动时校验 ClickHouse event/property 表存在，HTTP resolver 返回的 source 也必须落在启动 schema surface 内；也可在本地/小部署通过显式开关创建当前 runtime config 内所有启用 source 的 routed tables 后再校验，仍不拥有配置生命周期。

`analytics-core` 负责：

- Tenant / Project / Source 级事件接收和隔离。
- 事件协议、队列、写入、元数据、查询。
- Realtime、Events、Goal、后续 Funnels/Retention/Paths 等分析接口。

集成方式优先级：

1. P1 先按 Go library + runtime service 契约设计。
2. SimpleTrack 浏览器 SDK 通过 write key 调用 `simpletrack-anaysitics-service` 的 collect API。
3. 管理面由 Supastarter 承接，运行时数据面由 `simpletrack-anaysitics-service` 承接，通用分析核心由 `analytics-core` 提供。

## 验收清单

| 验收项 | 标准 |
| --- | --- |
| 仓库边界 | `analytics-core` 独立存在，不带 SimpleTrack 或 xwl 命名 |
| Go library 边界 | `analytics-core` 公共能力位于根目录包，可被外部 Go 服务 import；不作为 SimpleTrack 业务服务运行 |
| P1 运行依赖 | Redis Stream + MySQL + ClickHouse 可跑通；Kafka 非必选 |
| 本地依赖 | 当前已提供 `src/analytics-core/docker-compose.yml`，包含 Redis Stack、MySQL 8.4、ClickHouse 25.3；默认使用高位端口避开本机已有数据库 |
| Kafka 保留 | `KafkaBus` 接口和 adapter 边界存在，不删除高吞吐路线 |
| 事件协议 | 标准字段清楚，不提供 xwl_bi legacy 字段兼容 |
| HTTP collect API | `collect/httpapi` 可作为 Fiber 协议适配器；SimpleTrack 产品运行时的 write key、domain/CORS、quota 由 `simpletrack-anaysitics-service` 执行 |
| Redis Stream 消费 | pending 优先重试，写入成功才 ack，超过 `MaxAttempts` 后进入死信队列 |
| ingestion worker | 当前已明确 `ingestion.Processor` 为 P1 worker 边界，`simpletrack-anaysitics-service` 可显式开启同进程 worker 并复用它 |
| 表策略 | P1 采用按 project/source 物理分表的方案 B，但上层仍只面对统一 `events` 逻辑模型 |
| 启动校验 | `simpletrack-anaysitics-service` 开启 ingestion 时必须确认启用 source 的 routed event/property 表存在；默认缺表 fail-closed，本地/小部署可显式开启 ClickHouse auto migrate 创建当前 runtime config 内所有启用 source 的 routed tables；生产级批量迁移和回滚后续评审 |
| ClickHouse 写入 | 事件明细高吞吐写入默认使用原生 batch writer；当前已落地 `clickhouse-go/v2 PrepareBatch` 的 `BatchWriter`，后续建立与 GORM `CreateInBatches` 的压测对照 |
| 幂等入库 | 重复消费同一 `event_id` 不会在数据库产生两份事件明细；当前已落地 `EventWriteGuard` 边界和 GORM/MySQL `IngestionStatusGuard` 真实状态守卫 |
| Realtime | 当前已落地 query plan builder 和 ClickHouse query reader，并通过 opt-in e2e 验证最近事件能被读出；e2e 已补依赖 readiness retry |
| Events / Raw Events | 当前已落地 query plan builder 和 ClickHouse query reader，并通过 opt-in e2e 验证明细事件可查且属性可确认；SaaS Events 页面已把事件名、distinct id、排序、分页参数限制在白名单内，并通过额外读取一条判断 `hasMore`；重复 query 参数和空页偏移已补回归硬化 |
| 事件属性 | P1-002A 已约束属性入口、提供 typed row 逻辑展开，并通过 `PropertyIndexingEventWriter` + ClickHouse `PropertyBatchWriter` + MySQL `property_indexing_status` 接入 ingestion 热路径；真实 ClickHouse e2e 已确认 property rows 可写入、可查询、可按 allowlisted property filter 精确过滤 |
| 用户属性 | identify 语义进入 `DistinctID` + `UserProps`，用户属性和事件属性分开处理，不混成一类 JSON；当前已共用 collect 属性入口约束、typed row 展开模型、属性热路径写入和属性过滤模型 |
| session/visit | `SessionResolverStage` 与 `VisitResolverStage` 已落地；`visit_id` 是写入前确定并持久化的 canonical analytics visit key，cookie/no-cookie、salt 轮换和 Sessions 专页放 P1.5/P2 |
| client enrich / bot 过滤 | P1-002B 需要以 stage 形式实现 IP/UA/geo/utm/click id 补齐和 bot/IP/internal traffic 过滤，不进入 writer；当前 browser / OS / device 和 geo enrich 边界已落地，后续继续收口 internal traffic 产品配置与过滤审计 |
| 查询安全 | P1-002D 已落地 Events 排序/过滤字段白名单、operator enum、filter 数量上限、typed property filter allowlist、非法输入测试和真实 ClickHouse e2e；后续复杂查询必须补真实执行验证 |
| 元数据 | 事件名、事件属性、用户属性能被捕获 |
| Goal | 能定义关键事件并返回基础结果 |
| 业务无关 | 不出现订阅、账单、套餐、团队、Admin UI 逻辑 |
| 代码质量 | 查询、队列、存储、分析模块边界清楚，有最小单元测试 |
| 压测基线 | 需要建立 `analytics-core` 独立压测基线，覆盖 collect、Redis Stream、KafkaBus、ClickHouse 写入和典型查询；当前已补 builder-only read-side shape benchmark、真实 ClickHouse EventReader benchmark、真实 ClickHouse BatchWriter benchmark、GORM `CreateInBatches` 对照 benchmark、Redis Stream publish / subscribe+ack benchmark 和 `collect.Handler` 热路径 benchmark；collect 本地基线约为 normalize+publish 3.2-6.1µs/op、identity resolver 11.0-12.0µs/op、identity + client enrichment 15.2-18.2µs/op；ClickHouse 单事件 native writer 约 5.2-6.3ms/op，GORM 单事件约 7.3-8.2ms/op；100 行 native bulk 约 6.3-8.2ms/op，GORM bulk 约 8.0-10.1ms/op；后续继续补 KafkaBus 和稳定聚合查询压测 |

## 后续待评审

- 方案 B 下物理表名 hash 规则、生产 DDL 迁移/回滚策略和跨 source 查询 fan-out / merge 细节；本地/小部署 auto migrate 只解决当前 runtime config 内所有启用 source 的 routed tables 创建，不替代生产 migration pipeline。
- GORM `CreateInBatches` 与 `clickhouse-go/v2 PrepareBatch` 的初步压测差异已完成；后续只在批量大小、网络环境或 ClickHouse 版本变化时复测。
- KafkaBus 的 ack、checkpoint、幂等、死信队列具体实现如何与 Redis Stream 保持一致。
- KafkaBus 迁移时如何复用 xwl_bi 现有 consumer offset 和 acceptance status 思路。
- Funnel / Retention 查询如何落到统一 GORM query builder。
- `analytics-core` 压测基线指标、数据量级和验收阈值。
- 事件属性存储选择：typed rows、ClickHouse Map/JSON、原始 JSON + 高频属性展开的混合模型，分别对应哪些查询能力和迁移成本。
- session/visit 隐私策略：`visit_id` 持久化方案已定，后续评审 salt 轮换、IP 保留策略、cookie/no-cookie、server identity、retention 默认值和 Sessions 产品页；DNT 浏览器侧 opt-in 已落地，后续只评审产品配置和 audit。
- client info enrich 与 bot/IP 过滤的执行位置、配置面和失败语义。
- ClickHouse 读侧优化的压测阈值、projection / materialized view / 小时聚合表的具体落地顺序，以及方案 B 多物理表如何批量应用这些结构。
- Web tracker SDK 与多语言 SDK 的阶段路线：P1 浏览器最短链路由 `simpletrack-anaysitics-service` 静态托管；React/Next/Node/mobile SDK 后续评审。
- Performance metrics 是事件属性、独立事件类型还是独立模型，是否进入 P2。
