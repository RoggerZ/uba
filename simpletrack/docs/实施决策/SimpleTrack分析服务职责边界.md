# SimpleTrack 分析服务职责边界

> 状态：已确定
> 最近更新：2026-05-03
> 作用：明确 `simpletrack-saas`、`simpletrack-analytics-service` 和 `analytics-core` 的职责边界，避免控制面与数据面重复建设。

## 结论

SimpleTrack 后端分成三层：

1. `simpletrack-saas` 是商业控制面和产品后台。
2. `simpletrack-analytics-service` 是分析数据面的运行时服务。
3. `analytics-core` 是业务无关 Go 第三方库。

`analytics-core` 不作为独立业务服务运行，也不托管 Browser SDK。`simpletrack-analytics-service` 通过 Go module 引用 `analytics-core` 的根目录公共包，并负责把 SimpleTrack 的运行时配置映射到通用分析核心。

## 架构关系

```mermaid
flowchart LR
  SDK["Browser SDK"] --> AS["simpletrack-analytics-service<br/>运行时采集与查询服务"]
  SaaS["simpletrack-saas<br/>商业控制面和后台页面"] --> Config["Control-plane DB/API<br/>source/write key/domain/quota"]
  AS --> Config
  AS --> Core["analytics-core<br/>Go library"]
  Core --> Redis["Redis Stream / Kafka"]
  Core --> CH["ClickHouse"]
  Core --> MySQL["MySQL guard / metadata"]
```

## `simpletrack-saas` 负责什么

`simpletrack-saas` 继续承接 Supastarter 已提供的商业控制面能力：

- 用户登录、组织、成员和权限页面。
- 套餐、订阅、支付入口和 subscription gate。
- Website / Source 创建、展示和管理。
- write key 生成、轮换、禁用和展示。
- domain allowlist、internal traffic 规则、quota / plan limit 等配置的管理页面。
- Realtime、Events、Websites、Goal 等产品后台页面。

它不接收高频事件，不直接写 ClickHouse 明细事件，也不复制 ingestion worker。

## `simpletrack-analytics-service` 负责什么

`simpletrack-analytics-service` 是运行时数据面服务，负责执行已经由控制面产生的配置：

- `GET /healthz`：进程健康检查。
- `GET /tracker.js`：托管 P1 Browser SDK 静态资产。
- `OPTIONS /collect`：浏览器 CORS preflight。
- `POST /collect`：事件上报入口。
- 根据 `write_key` 读取 runtime source config。
- 执行 source enabled、Origin/domain allowlist、CORS、internal traffic、bot 过滤。
- 不信任客户端传来的 `tenant_id`、`project_id`、`source_id`、`source_type`，统一由控制面配置覆盖。
- 把 SimpleTrack 的 workspace/site/source 映射为 `analytics-core` 的 `tenant_id/project_id/source_id`。
- 调用 `analytics-core` 的 collect pipeline、EventBus、ingestion、storage、query 能力。

它不创建站点，不邀请成员，不处理购买套餐，不提供 Admin 页面，不拥有配置生命周期。

## `analytics-core` 负责什么

`analytics-core` 是 Go 第三方库，外部服务通过根目录公共包引用：

```go
import "github.com/simpletrack/analytics-core/collect"
import "github.com/simpletrack/analytics-core/eventbus/redisstream"
import "github.com/simpletrack/analytics-core/storage"
```

它负责：

- collect 请求标准化和字段校验。
- session resolver、client enrichment、bot/internal traffic filter 等可测试 stage。
- EventBus 抽象与 Redis Stream / Kafka adapter。
- ingestion processor。
- ClickHouse writer、table router、query builder、event reader。
- 事件属性和用户属性的 typed row 展开、写入和过滤边界。

它不负责：

- SimpleTrack 用户、组织、套餐、订阅、账单、权限。
- write key 生命周期和配置管理页面。
- Browser SDK 托管。
- 产品后台页面。
- 独立 `cmd/server` 业务服务。

## 配置权责

配置的 CRUD 在 `simpletrack-saas`，配置的 runtime enforcement 在 `simpletrack-analytics-service`。

| 配置项 | 创建/修改方 | 运行时执行方 |
| --- | --- | --- |
| Source / Website | `simpletrack-saas` | `simpletrack-analytics-service` |
| write key | `simpletrack-saas` | `simpletrack-analytics-service` |
| domain allowlist | `simpletrack-saas` | `simpletrack-analytics-service` |
| internal traffic rules | `simpletrack-saas` | `simpletrack-analytics-service` |
| quota / plan limit | `simpletrack-saas` | `simpletrack-analytics-service` |
| collect validation / queue / storage | `analytics-core` library | `simpletrack-analytics-service` 调用 |

## 当前落地状态

- `src/analytics-core` 已调整为根目录公共 Go 包形态，Browser SDK 已从 core 移出。
- `src/analytics-service` 已创建本地 Go 仓库，当前提供 `/healthz`、`/tracker.js`、`OPTIONS /collect`、`POST /collect`。
- 初版 `MemoryResolver` 只用于本地开发和测试；生产后续替换为读取 SimpleTrack 控制面数据库、API 或缓存的 resolver。
- Events / Realtime 查询 HTTP API 暂不在本轮实现，进入后续 `P1-005D`。
