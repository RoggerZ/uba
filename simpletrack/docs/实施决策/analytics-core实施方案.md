# analytics-core 实施方案

> 状态：已确定 P1 执行，模块设计持续评审  
> 最近更新：2026-04-30  
> 来源：基于 xwl_bi 本地代码的 analyze + code-review 梳理，并结合 Umami、Litlyx 两个参考产品的调研资产。

## 结论

P1 新建独立业务无关仓库 `analytics-core`，从 xwl_bi 抽取分析数据面核心。它不是 xwl_bi 整仓改名，也不是 SimpleTrack 私有业务层。

`analytics-core` 只负责采集、事件、元数据、实时写入、查询聚合和分析模型，不负责定价、团队、订阅、账单、onboarding、产品官网和企业控制台。

后续关系是：`analytics-core` 逐步反向支撑 xwl_bi，作为 SimpleTrack 的分析数据面核心，也预留给 AppTrack 或其他行为分析产品复用。因此命名必须保持通用，不带具体业务含义。

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
- **统一查询构建**：从 `sqlx` 迁移到 GORM 最新稳定版本，统一使用 GORM 的 SQL Builder / Raw / Clauses / Scopes 能力承接查询构建；ClickHouse 高吞吐事件写入不走 ORM 热路径，优先使用原生 batch writer。

## 非目标

- 不复用 xwl_bi 旧 Vue2 后台界面。
- 不迁移 xwl_bi 旧菜单、旧权限、旧业务后台叙事。
- 不在 `analytics-core` 内实现登录、组织、订阅、账单、Admin、邮件。
- 不在 P1 产品层开放全量漏斗、留存、路径、归因页面。
- 不把 Kafka 作为 P1 必选运行依赖。

## xwl_bi 代码证据

| 发现 | 证据 | 判断 |
| --- | --- | --- |
| xwl_bi 已经是分析型 Go 服务 | `C:/Users/admin/Documents/src/xwl_bi/go.mod:1` module 为 `github.com/1340691923/xwl_bi`，依赖 ClickHouse、Sarama Kafka、MySQL、Redis、Fiber | 技术栈适合作为抽取来源，但命名和模块边界需要重构 |
| 启动层强绑定 Kafka、Redis、MySQL、ClickHouse | `cmd/report_server/main.go:58` 到 `63` 初始化 Kafka sync/async producer、Redis、MySQL、ClickHouse | 需要把全局初始化改成依赖注入或 adapter 装配 |
| 采集入口已经有较清晰 orchestration | `controller/report_ingress_handler.go:47` 到 `49` 描述 Resolve -> Build -> SendReportData 流程 | 可作为 `collect` 接口和 handler 的抽取参考 |
| 请求解码仍绑定旧字段 | `controller/report_request_decoder.go:70` 读取 `xwl_distinct_id`、`xwl_ip`、`xwl_part_date` | 需要定义新事件协议；旧字段只作为抽取参考，不做 legacy 兼容 |
| Producer 与 Kafka 耦合 | `platform-basic-libs/service/report/producer.go:15` 定义 `KafkaDataProducer`，`engine/db/kafka.go:9` 到 `12` 使用全局 Sarama producer/client | 需要抽出 `EventBus`，Kafka 只做 `KafkaBus` adapter |
| 已有多类分析命令 | `platform-basic-libs/service/analysis/interface.go:16` 到 `26` 包含 Funnel、Retention、Trace、Event、UserAttr、UserList、LTV、Attribution 等 | 适合映射到 `analysis/*` 模块，但查询层要重构 |
| 分析查询强依赖 xwl 表名和字段 | `analysis/event.go:178` 拼接 `xwl_event`，`analysis/funnel.go:118` 使用 ClickHouse `windowFunnel`，`analysis/retention.go:264` 到 `279` 使用 `xwl_distinct_id` 和 `xwl_part_date` | 查询思路可保留，SQL 需要参数化、统一表模型和命名 |
| sinker 已有完整 ETL 链路 | `cmd/sinker/internal/runner/report_handler.go:139` 描述 context extraction、geo enrich、metric parse、ensure columns、metadata、status、metric batch | 可作为 ingestion pipeline 参考，但要拆小接口 |
| realtime 链路轻量 | `cmd/sinker/internal/runner/realtime_handler.go:12` 说明不做动态补列和复杂校验，只快速入批并 ack | 可作为 P1 Realtime 写入路径参考 |
| ClickHouse 初始化仍有旧表语义 | `cmd/init_app/ck/init.go:30`、`70` 创建 acceptance status 和 realtime warehousing 表，包含 `xwl_kafka_offset` | P1 可保留 status/realtime 思路，但表名字段要重命名 |

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

- `EventQueryBuilder`：负责生成 ClickHouse 查询 SQL，底层使用 GORM Raw / SQL Builder / Scopes。
- `EventWriter`：负责 ClickHouse 事件明细写入，P1 默认使用 `clickhouse-go/v2 PrepareBatch`。
- `EventWriter` 后续必须保留压测口径，对比 GORM `CreateInBatches`、`clickhouse-go/v2 PrepareBatch` 和必要时的 `ch-go`。
- 不在 handler 或 analysis 模块里直接散落原生 SQL；即使用原生 batch writer，也必须藏在 ClickHouse storage adapter 内。

## 目标模块边界

建议目录草案：

```text
analytics-core/
  cmd/
    collect-api/
    worker/
    migrate/
  internal/
    collect/
      decoder/
      validator/
      handler/
    eventbus/
      eventbus.go
      direct/
      redisstream/
      kafka/
    ingestion/
      pipeline/
      realtime/
      batch/
    storage/
      clickhouse/
        query/
        writer/
      mysql/
      redis/
    metadata/
      events/
      properties/
      users/
    analysis/
      events/
      realtime/
      goals/
      funnels/
      retention/
      paths/
      ltv/
      attribution/
      segments/
      sessions/
      properties/
    schema/
      event/
      identity/
      tenant/
      project/
      source/
    platform/
      clock/
      logger/
      config/
  pkg/
    contracts/
    client/
```

目录原则：

- `internal/collect` 只处理请求解码、字段校验和事件标准化。
- `internal/eventbus` 屏蔽 Direct、Redis Stream、Kafka 差异。
- `internal/ingestion` 处理消费、补充字段、入批、ack 和失败状态。
- `internal/storage` 只封装外部依赖，不放业务分析逻辑；ClickHouse query 和 writer 分开，避免查询 builder 与高吞吐写入互相污染。
- `internal/analysis` 只暴露业务无关分析能力。
- `pkg/contracts` 放 xwl_bi、SimpleTrack、AppTrack 或其他上层产品可依赖的稳定契约。

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

验收：

- 非法事件能返回明确错误。
- 新协议字段不带 `xwl_`、`simpletrack`、`apptrack` 等业务前缀。

### Step 3：EventBus 与 Redis Stream

交付：

- 定义 `EventBus` 接口。
- 实现 `DirectBus` 和 `RedisStreamBus`。
- 保留 `KafkaBus` adapter 目录和接口，后续从 Sarama producer/consumer 迁入。
- 本地 Redis 使用 `redis/redis-stack:latest` 容器镜像。

验收：

- collect 发送事件后能进入 Redis Stream。
- consumer group 能消费并 ack。
- Redis Stream 与 KafkaBus 使用同一事件协议。

### Step 4：ClickHouse 写入和 Realtime

交付：

- 建立按 `tenant_id / project_id / source_id` 路由的事件物理表、实时事件表、ingestion status 表。
- worker 消费队列并通过 `EventWriter` 写入 ClickHouse。
- `EventWriter` 默认使用 `clickhouse-go/v2 PrepareBatch` 原生批量写入，GORM batch insert 只作为压测对照或低频管理写入选项。
- 写入前基于 `(tenant_id, project_id, source_id, event_id)` 做幂等判断，避免重复事件在数据库存两份。
- 提供 Realtime 和 Raw Events 查询。

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

## 与上层产品的集成边界

SimpleTrack / AppTrack / xwl_bi 产品层负责：

- Workspace、Site、App、Member、Role 等具体业务对象。
- 登录、组织、订阅、账单、邮件、Admin。
- 产品官网、定价页、docs/quickstart、onboarding。
- 企业分析控制台页面。

`analytics-core` 负责：

- Tenant / Project / Source 级事件接收和隔离。
- 事件协议、队列、写入、元数据、查询。
- Realtime、Events、Goal、后续 Funnels/Retention/Paths 等分析接口。

集成方式优先级：

1. P1 先按独立服务 + HTTP/gRPC/SDK 契约设计。
2. SimpleTrack 通过 source id / write key 调用 collect API。
3. 管理面由 Supastarter 承接，分析查询由 `analytics-core` 提供。

## 验收清单

| 验收项 | 标准 |
| --- | --- |
| 仓库边界 | `analytics-core` 独立存在，不带 SimpleTrack 或 xwl 命名 |
| P1 运行依赖 | Redis Stream + MySQL + ClickHouse 可跑通；Kafka 非必选 |
| Kafka 保留 | `KafkaBus` 接口和 adapter 边界存在，不删除高吞吐路线 |
| 事件协议 | 标准字段清楚，不提供 xwl_bi legacy 字段兼容 |
| 表策略 | P1 采用按 project/source 物理分表的方案 B，但上层仍只面对统一 `events` 逻辑模型 |
| ClickHouse 写入 | 事件明细高吞吐写入默认使用原生 batch writer，并建立与 GORM `CreateInBatches` 的压测对照 |
| 幂等入库 | 重复消费同一 `event_id` 不会在数据库产生两份事件明细 |
| Realtime | 最近事件能快速出现 |
| Events / Raw Events | 明细事件可查，能用于接入排障 |
| 元数据 | 事件名、事件属性、用户属性能被捕获 |
| Goal | 能定义关键事件并返回基础结果 |
| 业务无关 | 不出现订阅、账单、套餐、团队、Admin UI 逻辑 |
| 代码质量 | 查询、队列、存储、分析模块边界清楚，有最小单元测试 |
| 压测基线 | 需要建立 `analytics-core` 独立压测基线，覆盖 collect、Redis Stream、KafkaBus、ClickHouse 写入和典型查询 |

## 后续待评审

- 方案 B 下物理表名 hash 规则、DDL 迁移策略和跨 source 查询 fan-out / merge 细节。
- GORM `CreateInBatches` 与 `clickhouse-go/v2 PrepareBatch` 在 `analytics-core` 事件模型上的压测差异。
- Redis Stream 与 KafkaBus 的 ack、checkpoint、幂等、死信队列具体实现。
- KafkaBus 迁移时如何复用 xwl_bi 现有 consumer offset 和 acceptance status 思路。
- Funnel / Retention 查询如何落到统一 GORM query builder。
- `analytics-core` 压测基线指标、数据量级和验收阈值。
