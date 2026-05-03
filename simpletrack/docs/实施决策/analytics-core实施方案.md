# analytics-core 实施方案

> 状态：已确定 P1 执行，模块设计持续评审  
> 最近更新：2026-05-03
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
| HTTP 入口分裂且上报路由依赖偏旧 | `router/index.go` 使用 Fiber 做后台控制面；`cmd/report_server/runtime.go` 使用 fasthttp + `buaazp/fasthttprouter` 做上报入口；`go.mod` 中 `fasthttprouter` 为 v0.1.1 | 参考 xwl_bi 的 fasthttp collect 启动装配和中间件编排；`analytics-core` P1 采用 fasthttp，但不复用低活跃的 `fasthttprouter` |

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
| HTTP API | 使用活跃维护的 fasthttp，不沿用 xwl_bi 的 `buaazp/fasthttprouter` 路由层 | fasthttp 本体更新活跃，适合事件上报热路径；xwl_bi 的 `fasthttprouter` 依赖活跃度低。fasthttp 只放在 HTTP 适配层，避免对 `collect.Handler` 形成框架耦合 |

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
- `EventWriter` 后续必须保留压测口径，对比 GORM `CreateInBatches`、`clickhouse-go/v2 PrepareBatch` 和必要时的 `ch-go`。
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

- `collect` 只处理请求字段校验、事件标准化和可替换 pre-queue stage；`collect/httpapi` 是可选 fasthttp 适配器，不处理 SimpleTrack 业务鉴权。
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
- 定义 `tenant_id`、`project_id`、`source_id`、`source_type`、`event_name`、`event_time`、`distinct_id`、`session_id`、`properties`、`user_properties` 等字段。
- 定义 event name、source id、timestamp、distinct id、properties 校验规则。
- 不提供 xwl_bi legacy 字段兼容。xwl_bi 后续迁移时应改为写入新协议。
- 当前已落地 `collect.Normalize`，负责把 collect 请求标准化为 `EventEnvelope`，并校验事件 ID、租户、项目、数据源、事件名、用户标识和时间戳。
- 当前已落地 `collect.Handler`，这里的 Handler 指 `collect/handler.go` 中的事件上报核心处理器，不是 HTTP 路由 handler；它负责调用 `Normalize` 并把标准化后的事件发布到 `EventBus`。HTTP collect API 只做协议适配，不重复实现校验和发布逻辑。
- 当前已落地 fasthttp `POST /collect` 入口，负责 JSON 解码、HTTP 状态码和响应格式；`collect.Handler` 继续保持框架无关。

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
- 当前已落地 ClickHouse `TableRouter`，按 tenant / project / source 生成稳定 hash 物理表名，对上层仍暴露统一 `events` 逻辑模型。
- 当前已落地 storage `EventWriter` 接口和 ClickHouse native batch `BatchWriter`，真实写入统一复用该边界。
- 当前已落地 GORM/MySQL `IngestionStatusGuard` 和 `ingestion_status` 表，按 `(tenant_id, project_id, source_id, event_id)` 做 `processing / inserted / failed` 状态占用、重复跳过、失败回滚和重试再占用。
- worker 消费队列并通过 `EventWriter` 写入 ClickHouse。
- `EventWriter` 默认使用 `clickhouse-go/v2 PrepareBatch` 原生批量写入，GORM batch insert 只作为压测对照或低频管理写入选项。
- 当前已明确 `ingestion.Processor` 是 P1 worker 边界：Redis/Kafka adapter 负责 ack/nack，Processor 负责调用 `storage.EventWriter`，后续真实 worker 入口不得复制一套消费写入逻辑。
- worker 后续运行时必须把 `EventWriteGuard` claim、ClickHouse `EventWriter` append、guard commit/rollback 和 Redis/Kafka ack 串成同一条可恢复链路，避免重复事件在数据库存两份。
- 当前已提供 Realtime 和 Raw Events / Events 的 query plan builder 与 ClickHouse/GORM query reader，并已通过 opt-in e2e 验证 collect -> Redis Stream -> ingestion -> ClickHouse -> Realtime/Events reader。

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
| 事件属性与用户属性 | `event_data`、`session_data`、typed value | collect 阶段已落地属性 key、数量、标量类型、字符串长度和有限数字入口约束；storage 层已提供 `EventPropertyRecord` / `FlattenEventProperties` typed row 逻辑展开；ClickHouse `PropertyBatchWriter` 写入同源路由 `_properties` 表；`PropertyIndexingEventWriter` 已把属性索引组合进 ingestion 热路径；MySQL `property_indexing_status` 单独处理属性 checkpoint，failed 可原子 reclaim，processing ambiguous 不自动重试；真实 e2e 已验证写入、读取和属性过滤 | P1-002A 已完成；属性字典、ambiguous 恢复和 ClickHouse 去重/聚合优化放 P1.5/P2 |
| client info enrich | collect 入口补 IP、UA、browser、os、device、geo | P1-002B 第一版已落地 collect `Stage`：UA/referrer 可进入 bounded properties，IP 只允许盐化为 `client.ip_hash`；HTTP 默认不信任 forwarded IP，可信代理需显式 `WithTrustedProxyHeaders()`；浏览器 SDK 已支持 opt-in DNT，并自动补 allowlisted UTM/click id | 继续评审 geo、browser/os/device，禁止放入 ClickHouse writer |
| bot/IP/internal traffic 过滤 | collect 入口做 bot/IP 判断 | P1-002B 第一版已落地 `TrafficFilterStage`：按 bot UA token、internal CIDR/IP 在 EventBus publish 前返回 `FilteredError`，HTTP 返回 accepted filtered 响应，不写入分析明细；DNT active 时浏览器 SDK 不发送也不持久化 identity | 后续评审 allow/deny 配置来源、产品 UI、internal traffic 和审计/采样策略 |
| session/visit resolver | source + id 或 IP/UA/salt 派生 session，visit 使用短窗口 | P1-002C 第一版已落地可替换 `SessionResolverStage`，在缺失 `session_id` 时用 salt + 时间窗口 + tenant/project/source/distinct_id 派生匿名 session；IP/UA 只能作为 transient hash 输入；浏览器 DNT opt-in 避免持久本地 identity | `visit_id` 尚未进入事件契约，后续评审 schema、salt 轮换、cookie/no-cookie 和 retention |
| 查询白名单与过滤 | `FILTER_COLUMNS`、operator mapping、分页 | `EventQueryBuilder` 字段白名单、排序白名单、过滤 operator enum、分页上限和 typed property filter allowlist；属性过滤使用 ClickHouse tuple `IN` 半连接查询属性表，避免 correlated `EXISTS` 外层 alias 兼容问题；`simpletrack-saas` Events 页面现在也把 `event_name`、`distinct_id`、`limit`、`offset`、`sort_field`、`sort_direction` 做服务端归一化后再请求内部读回放 | P1-002D 已完成，P1-005D 继续把查询安全落到 SaaS readback 边界，后续复杂查询继续复用 allowlist + 真实 ClickHouse e2e |
| Realtime/Events 验收 | Realtime 短窗口、Events 分页明细 | `EventReader` 读取 ClickHouse query plan 结果；e2e 入口已增加 Redis/MySQL/ClickHouse 冷启动 readiness 重试，避免 compose 刚启动时 native handshake EOF 误伤验收；Events 产品页使用额外读取一条的 `hasMore` 模型，不做总数查询；内部读回放 token 可用 `ANALYTICS_SERVICE_QUERY_TOKENS_JSON` 做短窗口轮换 | P1-002E 已完成，P1-005D 已补页面分页交互和 query token 轮换 allowlist，后续作为回归入口 |
| Web tracker SDK | auto pageview、custom event、identify、performance | P1 已落地 SimpleTrack 浏览器 SDK，但已从 `analytics-core` 迁出；当前由 `simpletrack-anaysitics-service` 的 `/tracker.js` 静态交付，并通过 `data-write-key` 进入运行时 collect 服务 | P1-004 已完成；React/Next/Node/mobile SDK、CDN 版本化和 performance metrics 后续评审 |
| ClickHouse 读侧优化 | materialized view、小时聚合表、projection、typed 属性 | ClickHouse adapter 的聚合表、projection、高频属性索引和迁移策略 | P1.5-001，P1 闭环后压测评审 |
| Performance metrics | LCP、INP、CLS、FCP、TTFB | 可作为事件类型或属性组进入协议扩展 | P2-001，P1 只预留承接能力 |

实现顺序：

1. P1-002E 已完成：pageview、自定义事件属性和 user properties 已能从 collect 进入 ClickHouse 并被 Realtime/Events 查询；冷启动 e2e readiness 已复验稳定。
2. P1-002A 已完成：`PropertyBatchWriter` 已通过 `PropertyIndexingEventWriter` 组合进 ingestion worker，属性跨表幂等使用 `property_indexing_status` guard；processing ambiguous 不自动恢复，后续作为 P1.5/P2 运维和 ClickHouse 去重策略评审项。
3. P1-004 已完成并纠偏：浏览器 SDK 最短链路和 docs/quickstart 已改为 write key 接入，SDK 由 `simpletrack-anaysitics-service` 托管，不再属于 `analytics-core`；后续继续评审 visit/geo、SDK 发布策略和多语言 SDK。
4. P1-005D 正在推进：内部 `/v1/realtime`、`/v1/events` 已由 `simpletrack-anaysitics-service` 读回放，SaaS 页面只走 server-side helper；Events 已补白名单筛选、排序和 `hasMore` 分页，内部 query token 不进入浏览器，并已支持服务端短窗口轮换 allowlist。
5. P1 数据闭环稳定后，再做 P1.5-001 的 ClickHouse 读侧优化压测，不提前用 MV/projection 增加迁移复杂度。

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
| HTTP collect API | `collect/httpapi` 可作为 fasthttp 协议适配器；SimpleTrack 产品运行时的 write key、domain/CORS、quota 由 `simpletrack-anaysitics-service` 执行 |
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
| session/visit | P1-002C 需要可替换 resolver，支持匿名 hash、业务 id、cookie/no-cookie 和 salt 轮换策略 |
| client enrich / bot 过滤 | P1-002B 需要以 stage 形式实现 IP/UA/geo/utm/click id 补齐和 bot/IP/internal traffic 过滤，不进入 writer |
| 查询安全 | P1-002D 已落地 Events 排序/过滤字段白名单、operator enum、filter 数量上限、typed property filter allowlist、非法输入测试和真实 ClickHouse e2e；后续复杂查询必须补真实执行验证 |
| 元数据 | 事件名、事件属性、用户属性能被捕获 |
| Goal | 能定义关键事件并返回基础结果 |
| 业务无关 | 不出现订阅、账单、套餐、团队、Admin UI 逻辑 |
| 代码质量 | 查询、队列、存储、分析模块边界清楚，有最小单元测试 |
| 压测基线 | 需要建立 `analytics-core` 独立压测基线，覆盖 collect、Redis Stream、KafkaBus、ClickHouse 写入和典型查询 |

## 后续待评审

- 方案 B 下物理表名 hash 规则、生产 DDL 迁移/回滚策略和跨 source 查询 fan-out / merge 细节；本地/小部署 auto migrate 只解决当前 runtime config 内所有启用 source 的 routed tables 创建，不替代生产 migration pipeline。
- GORM `CreateInBatches` 与 `clickhouse-go/v2 PrepareBatch` 在 `analytics-core` 事件模型上的压测差异。
- KafkaBus 的 ack、checkpoint、幂等、死信队列具体实现如何与 Redis Stream 保持一致。
- KafkaBus 迁移时如何复用 xwl_bi 现有 consumer offset 和 acceptance status 思路。
- Funnel / Retention 查询如何落到统一 GORM query builder。
- `analytics-core` 压测基线指标、数据量级和验收阈值。
- 事件属性存储选择：typed rows、ClickHouse Map/JSON、原始 JSON + 高频属性展开的混合模型，分别对应哪些查询能力和迁移成本。
- session/visit 隐私策略：salt 轮换、IP 保留策略、cookie/no-cookie、server identity 和 retention 的默认值；DNT 浏览器侧 opt-in 已落地，后续只评审产品配置和 audit。
- client info enrich 与 bot/IP 过滤的执行位置、配置面和失败语义。
- ClickHouse 读侧优化何时引入 materialized view、projection、小时聚合表和高频属性索引，方案 B 多物理表如何批量迁移。
- Web tracker SDK 与多语言 SDK 的阶段路线：P1 浏览器最短链路由 `simpletrack-anaysitics-service` 静态托管；React/Next/Node/mobile SDK 后续评审。
- Performance metrics 是事件属性、独立事件类型还是独立模型，是否进入 P2。
