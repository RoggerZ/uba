# 接口与数据格式

本文集中说明 `analytics-core` 与 `analytics-service` 对外交互的数据格式。读源码时可以把它当作“字段地图”。

## 1. `/collect` 写入接口

位置：

- 路由分发：`src/analytics-service/internal/collectapi/handler.go`
- body 结构：`collectPayload`
- core 请求结构：`src/analytics-core/collect/request.go` 的 `collect.Request`
- core 输出结构：`src/analytics-core/contracts/event.go` 的 `contracts.EventEnvelope`

请求示例：

```json
{
  "write_key": "wk_live",
  "id": "evt_1",
  "tenant_id": "tenant_client",
  "project_id": "project_client",
  "source_id": "source_client",
  "source_type": "mobile",
  "event_name": "pageview",
  "distinct_id": "visitor_1",
  "session_id": "ses_client_optional",
  "visit_id": "vis_client_optional",
  "event_time": "2026-05-03T10:00:02Z",
  "properties": {
    "page.path": "/docs",
    "plan": "pro",
    "paid": true
  },
  "user_properties": {
    "email_domain": "example.com"
  },
  "source": "browser-sdk"
}
```

注意：客户端传的 `tenant_id/project_id/source_id/source_type` 不可信。`handleCollect` 会在解析 write key 后，用 `SourceConfig` 覆盖这些字段：

```go
request := payload.Request
request.TenantID = source.TenantID
request.ProjectID = source.ProjectID
request.SourceID = source.SourceID
request.SourceType = source.SourceType
request.Client = h.clientInfo(ctx)
```

### write key 来源与优先级

位置：`src/analytics-service/internal/collectapi/handler.go` 的 `writeKey`

按优先级读取：

1. Header：`X-SimpleTrack-Write-Key: wk_header`
2. Header：`Authorization: Bearer wk_bearer`
3. Query：`/collect?write_key=wk_query`
4. Body：`{"write_key":"wk_body"}`

这不是“只从三个地方取”。当前代码支持四个来源，其中 body 是最后兜底。这样浏览器 SDK、服务端 SDK、调试请求可以用不同承载方式，但最终都进入同一个 `Resolver.ResolveSource` 边界。

### 成功响应

位置：`AcceptedResponse`

```json
{
  "id": "evt_1",
  "received_at": "2026-05-03T10:00:00Z"
}
```

如果事件有效但被 bot/internal traffic 过滤：

```json
{
  "id": "evt_1",
  "received_at": "2026-05-03T10:00:00Z",
  "filtered": true
}
```

### 错误响应

位置：`ErrorResponse`

```json
{
  "error": "event_name: contains unsupported characters"
}
```

典型状态码：

- `400`：JSON 无效、字段校验失败、查询参数失败。
- `401`：write key 不存在，或内部 query token 无效。
- `403`：source disabled，或 Origin 不在允许列表。
- `404`：未知路由，或 query 功能未启用时访问 query 路由。
- `500`：resolver、stage、queue、storage 等服务端依赖错误。

## 2. collect.Request 与 EventEnvelope

### collect.Request

位置：`src/analytics-core/collect/request.go`

关键字段：

```go
type Request struct {
    ID         string         `json:"id"`
    TenantID   string         `json:"tenant_id"`
    ProjectID  string         `json:"project_id"`
    SourceID   string         `json:"source_id"`
    SourceType string         `json:"source_type"`
    EventName  string         `json:"event_name"`
    DistinctID string         `json:"distinct_id"`
    SessionID  string         `json:"session_id,omitempty"`
    VisitID    string         `json:"visit_id,omitempty"`
    EventTime  time.Time      `json:"event_time,omitempty"`
    Properties map[string]any `json:"properties,omitempty"`
    UserProps  map[string]any `json:"user_properties,omitempty"`
    Source     string         `json:"source,omitempty"`
    Client     ClientInfo     `json:"-"`
}
```

`Client` 不来自 JSON body，它由 `analytics-service` 从 HTTP 请求中填充：

- User-Agent
- 客户端 IP
- Referer

这些值是 transient metadata，只供 filter/enrichment/session/visit stage 使用。原始 IP 不会进入 `EventEnvelope`。

### EventEnvelope

位置：`src/analytics-core/contracts/event.go`

`EventEnvelope` 是 core 内部最重要的数据格式，贯穿 collect、EventBus、ingestion、storage：

```go
type EventEnvelope struct {
    ID         string         `json:"id"`
    TenantID   string         `json:"tenant_id"`
    ProjectID  string         `json:"project_id"`
    SourceID   string         `json:"source_id"`
    SourceType string         `json:"source_type"`
    EventName  string         `json:"event_name"`
    DistinctID string         `json:"distinct_id"`
    SessionID  string         `json:"session_id,omitempty"`
    VisitID    string         `json:"visit_id,omitempty"`
    EventTime  time.Time      `json:"event_time"`
    ReceivedAt time.Time      `json:"received_at"`
    Properties map[string]any `json:"properties,omitempty"`
    UserProps  map[string]any `json:"user_properties,omitempty"`
    Source     string         `json:"source,omitempty"`
}
```

两个时间字段的区别：

- `event_time`：事件源声称事件发生的时间。例如浏览器点击发生在 `10:00:02`。
- `received_at`：采集服务接受请求的时间。例如请求经过网络到服务端时是 `10:00:04`。

如果客户端没传 `event_time`，`Normalize` 会把它设为 `received_at`。如果客户端传了过远的未来时间，`Normalize` 会拒绝，避免 Realtime/Events 被未来事件污染。

## 3. 字段校验规则

位置：`src/analytics-core/collect/request.go`

四个正则各自作用如下：

| 正则变量 | 用在哪些字段 | 允许示例 | 拒绝示例 | 作用 |
| --- | --- | --- | --- | --- |
| `identifierPattern` | `id`, `tenant_id`, `project_id`, `source_id`, `distinct_id`, `session_id`, `visit_id` | `evt_1`, `tenant-1`, `user:42`, `vis_abc` | `_evt`, `evt 1`, 空字符串 | 保证业务边界 key 可安全用于幂等、路由、队列和存储 |
| `eventNamePattern` | `event_name` | `pageview`, `checkout.completed`, `button:click` | `/pageview`, `checkout completed` | 让事件名适合指标、筛选和查询 |
| `propertyKeyPattern` | `properties` 和 `user_properties` 的每个 key | `page.path`, `plan:tier` | `page path`, `$plan`, `items[]` | 防止开放属性 key 进入不可控命名空间 |
| `sourceTypePattern` | `source_type` | `web`, `server`, `mobile_app` | `Web`, `_web`, `web app` | 保持来源类别小写、稳定、可筛选 |

属性值类型限制：

- 允许：`nil`、bool、整数、浮点数、`json.Number`、string。
- 拒绝：对象、数组、NaN、Infinity、过长字符串。
- 属性数量最多 `64`，单个 key 最多 `128` 字符，单个字符串值最多 `2048` 字符。

这样设计的原因：属性是开放输入，如果不在 collect 边界做约束，后续 Redis、ClickHouse、属性索引、查询过滤都要反复防御。

## 4. SourceConfig 数据格式

位置：`src/analytics-service/internal/controlplane/resolver.go`

`SourceConfig` 是 SaaS control-plane 给 analytics runtime 的“可信配置视图”：

```go
type SourceConfig struct {
    WriteKey                 string                  `json:"write_key"`
    Enabled                  bool                    `json:"enabled"`
    TenantID                 string                  `json:"tenant_id"`
    ProjectID                string                  `json:"project_id"`
    SourceID                 string                  `json:"source_id"`
    SourceType               string                  `json:"source_type"`
    AllowedOrigins           []string                `json:"allowed_origins"`
    AllowedPropertyFilters   []AllowedPropertyFilter `json:"allowed_property_filters"`
    BotUserAgents            []string                `json:"bot_user_agents"`
    InternalCIDRs            []string                `json:"internal_cidrs"`
    InternalIPs              []string                `json:"internal_ips"`
    SessionSalt              string                  `json:"session_salt"`
    VisitSalt                string                  `json:"visit_salt"`
    VisitWindow              time.Duration           `json:"-"`
    VisitWindowSeconds       int                     `json:"visit_window_seconds"`
    ClientHashSalt           string                  `json:"client_hash_salt"`
    IncludeClientFingerprint bool                    `json:"include_client_fingerprint"`
}
```

用途：

- `WriteKey`：公共采集 key，用于找到这份配置。
- `TenantID/ProjectID/SourceID/SourceType`：覆盖客户端字段，成为 analytics-core 的可信边界。
- `AllowedOrigins`：限制浏览器来源。
- `AllowedPropertyFilters`：限制内部查询 API 可以按哪些属性过滤。
- `BotUserAgents/InternalCIDRs/InternalIPs`：运行时过滤规则。
- `SessionSalt/VisitSalt/ClientHashSalt`：服务端私有盐，用于派生 session、canonical visit 和 IP hash。
- `VisitWindow/VisitWindowSeconds`：服务端 visit 时间桶配置。JSON 中用秒数传输，Go runtime 内部用 `time.Duration`。

HTTP resolver 与 control-plane 的交互：

请求：

```json
{
  "write_key": "wk_live"
}
```

响应：

```json
{
  "write_key": "wk_live",
  "enabled": true,
  "tenant_id": "tenant_control",
  "project_id": "project_control",
  "source_id": "source_web",
  "source_type": "web",
  "allowed_origins": ["https://app.example.com"],
  "allowed_property_filters": [
    {"scope": "event", "name": "plan", "value_types": ["string"]}
  ],
  "bot_user_agents": ["bot", "crawler"],
  "internal_cidrs": ["10.0.0.0/8"],
  "internal_ips": ["127.0.0.1"],
  "session_salt": "server-only-session-salt",
  "visit_salt": "server-only-visit-salt",
  "visit_window_seconds": 1800,
  "client_hash_salt": "server-only-client-salt",
  "include_client_fingerprint": true
}
```

## 5. Redis Stream 消息格式

位置：`src/analytics-core/eventbus/redisstream/redisstream.go`

普通事件消息：

```text
Stream: analytics.events
Field: envelope
Value: JSON(EventEnvelope)
```

示例：

```json
{
  "envelope": "{\"id\":\"evt_1\",\"tenant_id\":\"tenant_control\",\"project_id\":\"project_control\",\"source_id\":\"source_web\",\"source_type\":\"web\",\"event_name\":\"pageview\",\"distinct_id\":\"visitor_1\",\"session_id\":\"ses_1\",\"visit_id\":\"vis_1\",\"event_time\":\"2026-05-03T10:00:02Z\",\"received_at\":\"2026-05-03T10:00:04Z\",\"properties\":{\"page.path\":\"/docs\"}}"
}
```

死信消息字段：

```text
envelope              原始 EventEnvelope JSON
attempt               第几次处理失败
consumer              具体 consumer 名称
consumer_group        consumer group 名称
error                 失败原因
failed_at             写入死信队列时间
original_message_id   Redis Stream 原消息 ID
```

## 6. MySQL checkpoint 表

位置：

- `src/analytics-core/storage/mysql/ingestion_status_guard.go`
- `src/analytics-core/storage/mysql/property_indexing_status_guard.go`

事件写入 checkpoint：`ingestion_status`

| 字段 | 类型/含义 |
| --- | --- |
| `tenant_id` | 主键之一，租户边界 |
| `project_id` | 主键之一，项目边界 |
| `source_id` | 主键之一，数据源边界 |
| `event_id` | 主键之一，事件幂等 ID |
| `status` | `processing` / `inserted` / `failed` |
| `attempt` | claim 次数 |
| `last_error` | 最近一次写入失败原因 |
| `received_at` | collect 接收时间 |
| `created_at` / `updated_at` | GORM 维护时间 |

属性索引 checkpoint：`property_indexing_status`

字段类似，但表示“这个事件对应的属性行是否已经写完”。它和主事件 checkpoint 分开，是因为主事件行可能已经写入成功，而属性表写入失败，需要后续重试修复。

## 7. ClickHouse 表格式

位置：`src/analytics-core/storage/clickhouse/schema.go`

事件主表列：

| 列 | 类型 | 来源 |
| --- | --- | --- |
| `event_id` | String | `EventEnvelope.ID` |
| `tenant_id` | String | `EventEnvelope.TenantID` |
| `project_id` | String | `EventEnvelope.ProjectID` |
| `source_id` | String | `EventEnvelope.SourceID` |
| `source_type` | String | `EventEnvelope.SourceType` |
| `event_name` | String | `EventEnvelope.EventName` |
| `distinct_id` | String | `EventEnvelope.DistinctID` |
| `session_id` | String | `EventEnvelope.SessionID` |
| `visit_id` | String | `EventEnvelope.VisitID` |
| `event_time` | DateTime64(3, UTC) | `EventEnvelope.EventTime.UTC()` |
| `received_at` | DateTime64(3, UTC) | `EventEnvelope.ReceivedAt.UTC()` |
| `properties` | String | `EventEnvelope.Properties` JSON 字符串 |
| `user_properties` | String | `EventEnvelope.UserProps` JSON 字符串 |
| `source` | String | `EventEnvelope.Source` |

属性表列：

| 列 | 类型 | 来源 |
| --- | --- | --- |
| `event_id` 到 `source` | 同事件主表，包含 `visit_id` | 从 envelope 复制 |
| `property_scope` | String | `event` 或 `user` |
| `property_name` | String | 属性 key |
| `property_type` | String | `null` / `string` / `number` / `bool` |
| `string_value` | String | string 属性值 |
| `number_value` | Float64 | number 属性值 |
| `bool_value` | Bool | bool 属性值 |

物理表名不是直接拼接 tenant/project/source 原文，而是：

```text
{prefix}_{sha256(tenant_id)前12位}_{sha256(project_id)前12位}_{sha256(source_id)前12位}
```

示意：

```text
events_a1b2c3d4e5f6_112233445566_abcdef012345
events_a1b2c3d4e5f6_112233445566_abcdef012345_properties
```

这样避免把租户、项目、来源 ID 原文泄露到 ClickHouse identifier。

## 8. 内部查询接口

位置：`src/analytics-service/internal/collectapi/query.go`

### `/v1/realtime`

请求：

```http
GET /v1/realtime?write_key=wk_live&since=2026-05-03T09:30:00Z&limit=50
Authorization: Bearer query-token
```

响应：

```json
{
  "source": {
    "tenant_id": "tenant_control",
    "project_id": "project_control",
    "source_id": "source_web",
    "source_type": "web"
  },
  "items": [
    {
      "id": "evt_1",
      "tenant_id": "tenant_control",
      "project_id": "project_control",
      "source_id": "source_web",
      "source_type": "web",
      "event_name": "pageview",
      "distinct_id": "visitor_1",
      "session_id": "ses_x",
      "visit_id": "vis_x",
      "event_time": "2026-05-03T10:00:02Z",
      "received_at": "2026-05-03T10:00:04Z",
      "properties": {"page.path": "/docs"}
    }
  ],
  "since": "2026-05-03T09:30:00Z",
  "limit": 50
}
```

### `/v1/events`

`/v1/events` 是默认内部 readback 路径。当前代码不再把它写死在 query 包里，而是由配置层提供默认值，并允许通过环境变量覆盖：

证据：

- `仓库: analytics-service, commit: 09656b6, file: internal/config/config.go:18-20`
- `仓库: analytics-service, commit: 09656b6, file: internal/config/config.go:97-99`
- `仓库: analytics-service, commit: 09656b6, file: internal/collectapi/handler.go:155-160`

```go
defaultEventsPath   = "/v1/events"
defaultRealtimePath = "/v1/realtime"

EventsPath:   envString("ANALYTICS_SERVICE_EVENTS_PATH", defaultEventsPath)
RealtimePath: envString("ANALYTICS_SERVICE_REALTIME_PATH", defaultRealtimePath)

app.Get(h.opts.RealtimePath, h.handleRealtime)
app.Get(h.opts.EventsPath, h.handleEvents)
```

所以 `/v1/events` 是 analytics-service 内部 Events readback API 的默认路径；部署时可以保留默认值，也可以改成更内部化的路径，例如 `/internal/events`。

这里的 `readback` 意思是“读回”：服务端从查询存储读取已经由 `/collect` 接收并入库的事件，再返回给 SaaS 页面展示。它不是事件上报，也不是 event replay / 重放历史事件。

请求：

```http
GET /v1/events?write_key=wk_live&from=2026-05-03T00:00:00Z&to=2026-05-04T00:00:00Z&event_name=pageview&visit_id=vis_x&limit=100&offset=0&sort_field=event_time&sort_direction=desc
Authorization: Bearer query-token
```

属性过滤参数是可重复 query 参数，每个值是一段 JSON；`analytics-service` 会先用 runtime source 的 allowlist 校验，再交给 `analytics-core` query builder：

```http
property_filter={"scope":"event","name":"plan","type":"string","op":"eq","value":"pro"}
property_filter={"scope":"user","name":"role","type":"string","op":"eq","value":"admin"}
```

对应 core query 结构：

```go
storage.EventPropertyFilter{
    Scope:       storage.PropertyScopeEvent,
    Name:        "plan",
    ValueType:   storage.PropertyValueString,
    Operator:    storage.EventFilterEquals,
    StringValue: "pro",
}
```

过滤被两层限制：

1. service 层：`SourceConfig.AllowsPropertyFilter(scope, name, valueType)`。
2. core 层：`EventQueryBuilder` 的 `AllowedPropertySelectors`。

这两层都是为了防止 query API 变成任意 ClickHouse 扫描入口。
