# collectapi tracker 与 query readback 源码解读

> 分析范围：`simpletrack-anaysitics-service` 的 `/tracker.js`、`/v1/realtime`、`/v1/events` 三个入口。
>
> 重点入口：
> - `writeTracker`：`仓库: analytics-service, commit: 09656b6, file: internal/collectapi/handler.go:345-355`
> - `handleRealtime`：`仓库: analytics-service, commit: 09656b6, file: internal/collectapi/query.go:72-120`
> - `handleEvents`：`仓库: analytics-service, commit: 09656b6, file: internal/collectapi/query.go:122-192`

## 0. 版本基线与引用规则

本文源码证据按以下子仓版本读取：

| 子仓 | commit id | 工作区状态说明 |
| --- | --- | --- |
| `src/analytics-service` | `580c547a29546ff9f0dd7314e075a7deff3d8412` | `git status --short` 仅显示未跟踪 `.idea/`，本文引用的源码文件与 `HEAD` 一致 |
| `src/analytics-core` | `ee455ac25790719cbd42dd7a5bb41492965741d9` | `git status --short` 仅显示未跟踪 `.idea/`；本文新增 query evidence 说明引用 `HEAD`，历史 P1 段落保留原始 commit 证据 |
| `src/simpletrack-saas` | `bce33354ae27dcba80e2f1ce77ff7ac2c5ed8765` | 基于 `HEAD bce33354ae27dcba80e2f1ce77ff7ac2c5ed8765` + 工作区未提交改动；当前 `packages/database/prisma/zod/index.ts` 有未提交改动、`.omx/` 未跟踪，本文引用的 readback / runtime-source 文件与 `HEAD` 一致 |

后文引用具体代码时使用格式：`仓库: <repo>, commit: <sha>, file: <path>:<line>`；跨多行使用 `<path>:<start>-<end>`。

## 1. 整体架构图

这三个入口分成两类：

- `/tracker.js` 是公开静态 SDK 入口，负责把浏览器 tracker 脚本发给网页。
- `/v1/realtime` 和 `/v1/events` 是内部 readback API，给 SaaS 服务端页面读取最近事件和原始事件列表，不是浏览器直接拿 query token 调的公开 API。

这里的 `readback` 可以按字面理解成“读回”：把已经通过 `/collect` 接收、并由后续链路写入存储的分析事件，再从查询存储里读出来给产品页面看。它不是事件写入，也不是 event replay / 重放历史事件；它只做受信服务端的读查询。

```mermaid
flowchart LR
    Browser["浏览器页面"] --> TrackerRoute["GET /tracker.js"]
    TrackerRoute --> TrackerAsset["analytics-service/public/tracker.js"]
    TrackerAsset --> CollectRoute["POST /collect"]

    SaaSPage["simpletrack-saas<br/>Realtime / Events 页面"] --> ReadbackHelper["server-only readback helper"]
    ReadbackHelper --> QueryRoutes["analytics-service<br/>GET /v1/realtime<br/>GET /v1/events"]
    QueryRoutes --> QueryAuth["QueryCredential / Bearer query token"]
    QueryRoutes --> Resolver["controlplane.Resolver"]
    Resolver --> SourceConfig["SourceConfig<br/>tenant/project/source/origin/policy"]
    QueryRoutes --> QueryReader["storage.EventReader"]
    QueryReader --> CoreReader["analytics-core<br/>ClickHouse EventReader"]
    CoreReader --> QueryBuilder["EventQueryBuilder<br/>route + allowlist + limit cap"]
    QueryBuilder --> ClickHouse["ClickHouse event / property tables"]

    SaaSControl["simpletrack-saas<br/>runtime-source API"] -. "HTTPResolver 用 write_key 查询" .-> Resolver
```

### Fiber app 如何注册 route

`NewApp` 创建 Fiber app，安装 CORS middleware，然后调用 `registerRoutes`。证据：`仓库: analytics-service, commit: 09656b6, file: internal/collectapi/handler.go:64-84`

```go
app := fiber.New(...)
app.Use(cors.New(...))
h.registerRoutes(app)
```

`registerRoutes` 把 `TrackerPath`、`RealtimePath`、`EventsPath` 分别绑定到三个 handler。证据：`仓库: analytics-service, commit: 09656b6, file: internal/collectapi/handler.go:155-160`

```go
app.Get(h.opts.TrackerPath, h.writeTracker)
app.Get(h.opts.RealtimePath, h.handleRealtime)
app.Get(h.opts.EventsPath, h.handleEvents)
```

默认路径来自配置层：`/tracker.js`、`/v1/events`、`/v1/realtime`。证据：`仓库: analytics-service, commit: 09656b6, file: internal/config/config.go:18-24`

```go
defaultTrackerPath  = "/tracker.js"
defaultEventsPath   = "/v1/events"
defaultRealtimePath = "/v1/realtime"
defaultTrackerFile  = "public/tracker.js"
```

### 它们与 Handler、Resolver、EventReader、analytics-core 的关系

`runtime.New` 是装配层：读取 tracker 文件，创建 query reader，再把依赖放进 `collectapi.Options`。证据：`仓库: analytics-service, commit: 09656b6, file: internal/runtime/runtime.go:49-85`

```go
tracker, err := os.ReadFile(cfg.TrackerFile)
queryReader, queryClosers, err := newQueryReader(cfg)
app, err := collectapi.NewApp(collectapi.Options{
    TrackerScript: tracker,
    Resolver:      resolver,
    QueryReader:   queryReader,
})
```

`newQueryReader` 只在 query enabled 时打开 ClickHouse 读连接，创建 `TableRouter`、`EventQueryBuilder`、`clickhouse.EventReader`。证据：`仓库: analytics-service, commit: 09656b6, file: internal/runtime/runtime.go:156-187`

```go
if !cfg.QueryEnabled {
    return nil, nil, nil
}
router, err := clickhouse.NewTableRouter(cfg.ClickHouseTablePrefix)
builder, err := clickhouse.NewEventQueryBuilder(router, builderOptions...)
reader, err := clickhouse.NewEventReader(queryDB, builder)
```

`storage.EventReader` 是 analytics-core 对读接口的抽象，HTTP 层只调用 `ListEvents` / `ListRealtime`。证据：`仓库: analytics-core, commit: 7c296670842d, file: storage/event_query.go:142-146`

```go
type EventReader interface {
    ListEvents(context.Context, EventListQuery) ([]EventRecord, error)
    ListRealtime(context.Context, RealtimeQuery) ([]EventRecord, error)
}
```

### 与 simpletrack-saas readback helper 的交互

SaaS 页面本身调用 server-only helper，不直接拼 analytics-service 请求。`analytics-readback.ts` 顶部导入 `server-only`，再把 `process.env`、`fetch`、网站列表传给 core helper。证据：`仓库: simpletrack-saas, commit: bce33354ae27, file: apps/saas/modules/simpletrack/lib/analytics-readback.ts:1-37`、`仓库: simpletrack-saas, commit: bce33354ae27, file: apps/saas/modules/simpletrack/lib/analytics-readback.ts:40-99`

```ts
import "server-only";

return readRealtimeReadback(organizationId, {
  env: process.env,
  fetch,
  listWebsites: async () => resolvedWebsites,
  now: () => new Date(),
}, { websiteId });
```

Realtime helper 构造 `/v1/realtime?write_key=...&limit=20`，然后带 query token 调 analytics-service。证据：`仓库: simpletrack-saas, commit: bce33354ae27, file: apps/saas/modules/simpletrack/lib/analytics-readback-core.ts:129-160`

```ts
const requestUrl = buildServiceUrl(config.baseUrl, "/v1/realtime", {
  write_key: source.writeKey,
  limit: String(realtimeLimit),
});
const response = await fetchAnalyticsService(requestUrl, config.queryToken, dependencies.fetch);
```

Events helper 构造 `/v1/events`，传 `write_key`、时间窗口、分页、排序、可选 `event_name` / `distinct_id` / `visit_id` / `property_filter`。证据：`仓库: simpletrack-saas, commit: bce33354ae27, file: apps/saas/modules/simpletrack/lib/analytics-readback-core.ts:176-235`

```ts
const requestUrl = buildServiceUrl(config.baseUrl, "/v1/events", {
  write_key: source.writeKey,
  from: from.toISOString(),
  to: to.toISOString(),
  limit: String(normalizedOptions.pageSize + 1),
  offset: String(normalizedOptions.offset),
  sort_field: normalizedOptions.sortField,
  sort_direction: normalizedOptions.sortDirection,
});
requestUrl.searchParams.append("property_filter", JSON.stringify(normalizedOptions.propertyFilter));
```

真正发请求时，query token 来自服务端环境变量，并作为 `Authorization: Bearer ...` header 发送。证据：`仓库: simpletrack-saas, commit: bce33354ae27, file: apps/saas/modules/simpletrack/lib/analytics-readback-core.ts:279-326`

```ts
const queryToken =
  env.SIMPLETRACK_ANALYTICS_SERVICE_QUERY_TOKEN?.trim() ||
  env.ANALYTICS_SERVICE_QUERY_TOKEN?.trim();

headers: {
  Authorization: `Bearer ${queryToken}`,
},
```

通俗地说：浏览器页面只看到网站选择和筛选条件；query token 留在 Next.js 服务端 helper 里，浏览器不应该拿到它。

## 2. 整体代码结构图

```text
src/analytics-service/
├─ cmd/simpletrack-anaysitics-service/main.go
├─ api/openapi.yaml
├─ public/tracker.js
└─ internal/
   ├─ config/config.go
   ├─ runtime/runtime.go
   ├─ collectapi/
   │  ├─ handler.go
   │  ├─ query.go
   │  └─ query_auth.go
   └─ controlplane/
      ├─ resolver.go
      └─ http_resolver.go

src/analytics-core/
└─ storage/
   ├─ event_query.go
   └─ clickhouse/
      ├─ event_reader.go
      └─ query_builder.go

src/simpletrack-saas/
├─ packages/api/
│  ├─ index.ts
│  └─ modules/simpletrack/runtime-source.ts
└─ apps/saas/
   ├─ app/api/[[...rest]]/route.ts
   ├─ app/(authenticated)/(main)/(organizations)/[organizationSlug]/realtime/page.tsx
   ├─ app/(authenticated)/(main)/(organizations)/[organizationSlug]/events/page.tsx
   └─ modules/simpletrack/lib/
      ├─ analytics-readback.ts
      ├─ analytics-readback-core.ts
      └─ events-query.ts
```

这些文件的分工：

| 文件 | 角色 |
| --- | --- |
| `src/analytics-service/internal/collectapi/handler.go` | HTTP app、route 注册、tracker asset 返回、公共 `/collect` helper |
| `src/analytics-service/internal/collectapi/query.go` | `/v1/realtime`、`/v1/events`、query token 要求、参数解析、响应整形 |
| `src/analytics-service/internal/collectapi/query_auth.go` | query token 归一化、常量时间比较、not-before / expires-at 生命周期判断 |
| `src/analytics-service/internal/runtime/runtime.go` | 读取 tracker 文件，创建 ClickHouse `EventReader`，把依赖注入 Handler |
| `src/analytics-service/internal/config/config.go` | 从环境变量读取 path、tracker file、query enabled、query token、ClickHouse 地址 |
| `src/analytics-service/public/tracker.js` | 浏览器 SDK，读取 `data-write-key` 并 POST `/collect` |
| `src/analytics-service/api/openapi.yaml` | OpenAPI 文档，列出 tracker / realtime / events 的请求响应形状 |
| `src/analytics-core/storage/event_query.go` | storage-neutral 的 query / record / reader 接口定义 |
| `src/analytics-core/storage/clickhouse/event_reader.go` | 执行 ClickHouse query plan，把行模型转换成 `storage.EventRecord` |
| `src/analytics-core/storage/clickhouse/query_builder.go` | 表路由、字段白名单、排序白名单、分页上限、property filter SQL 计划 |
| `src/simpletrack-saas/apps/saas/modules/simpletrack/lib/analytics-readback-core.ts` | 服务端 readback helper，构造 analytics-service 请求并映射成页面行 |
| `src/simpletrack-saas/packages/api/modules/simpletrack/runtime-source.ts` | SaaS control-plane runtime source API，供 analytics-service HTTPResolver 查询 write key 对应 source |

## 3. 内部功能模块

### 3.1 writeTracker

#### route 是如何注册的

`newHandler` 默认把 `TrackerPath` 设为 `/tracker.js`。证据：`仓库: analytics-service, commit: 09656b6, file: internal/collectapi/handler.go:86-105`

```go
if opts.TrackerPath == "" {
    opts.TrackerPath = "/tracker.js"
}
```

`registerRoutes` 绑定 `GET /tracker.js -> writeTracker`。证据：`仓库: analytics-service, commit: 09656b6, file: internal/collectapi/handler.go:155-157`

```go
app.Get(h.opts.TrackerPath, h.writeTracker)
```

#### tracker script 从哪里加载

配置默认 tracker 文件是 `public/tracker.js`。证据：`仓库: analytics-service, commit: 09656b6, file: internal/config/config.go:18-24`

```go
defaultTrackerFile = "public/tracker.js"
```

启动装配时 `runtime.New` 用 `os.ReadFile(cfg.TrackerFile)` 读入字节数组，再放进 `collectapi.Options.TrackerScript`。证据：`仓库: analytics-service, commit: 09656b6, file: internal/runtime/runtime.go:49-55`、`仓库: analytics-service, commit: 09656b6, file: internal/runtime/runtime.go:68-85`

```go
tracker, err := os.ReadFile(cfg.TrackerFile)
...
TrackerScript: tracker,
```

tracker 文件自身会从 script tag 读取 `data-write-key`、`data-collect-url` 等配置。证据：`仓库: analytics-service, commit: 09656b6, file: public/tracker.js:33-46`

```js
var config = {
  writeKey: attr('write-key'),
  collectURL: attr('collect-url') || defaultCollectURL(script),
  autoTrack: attr('auto-track') !== 'false',
};
```

它发送事件时把 `write_key` 放进 JSON body，并 POST 到 collect URL。证据：`仓库: analytics-service, commit: 09656b6, file: public/tracker.js:263-304`

```js
return {
  id: randomToken('evt'),
  write_key: config.writeKey,
  event_name: eventName,
  distinct_id: state.distinctID,
};

window.fetch(config.collectURL, {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify(request),
});
```

#### 返回 header 和缓存策略

`writeTracker` 返回：

- `Content-Type: application/javascript; charset=utf-8`
- `Cache-Control: public, max-age=300`
- HTTP 200 body 为启动时读入的 tracker bytes

证据：`仓库: analytics-service, commit: 09656b6, file: internal/collectapi/handler.go:345-355`

```go
ctx.Set("Content-Type", "application/javascript; charset=utf-8")
ctx.Set("Cache-Control", "public, max-age=300")
return ctx.Status(fiber.StatusOK).Send(h.opts.TrackerScript)
```

这里的缓存策略很短：浏览器和 CDN 可以缓存 300 秒。好处是减少 tracker 请求；代价是脚本发布后最多可能有约 5 分钟旧版本残留。

#### tracker 缺失时如何返回错误

如果 `TrackerScript` 为空，返回 404 JSON。证据：`仓库: analytics-service, commit: 09656b6, file: internal/collectapi/handler.go:345-348`

```go
if len(h.opts.TrackerScript) == 0 {
    return h.writeJSON(ctx, fiber.StatusNotFound, ErrorResponse{Error: "tracker script is not configured"})
}
```

另外，正常 `runtime.New` 会在启动时读取 tracker 文件；文件不存在会直接返回错误，不进入服务监听。证据：`仓库: analytics-service, commit: 09656b6, file: internal/runtime/runtime.go:49-55`

### 3.2 handleRealtime

#### route 是如何注册的

默认 path 是 `/v1/realtime`，也可由 `ANALYTICS_SERVICE_REALTIME_PATH` 覆盖。证据：`仓库: analytics-service, commit: 09656b6, file: internal/config/config.go:19-20`、`仓库: analytics-service, commit: 09656b6, file: internal/config/config.go:98-99`

```go
defaultRealtimePath = "/v1/realtime"
RealtimePath: envString("ANALYTICS_SERVICE_REALTIME_PATH", defaultRealtimePath),
```

注册点：`仓库: analytics-service, commit: 09656b6, file: internal/collectapi/handler.go:155-160`

```go
app.Get(h.opts.RealtimePath, h.handleRealtime)
```

#### 如何校验 internal query token

`handleRealtime` 第一层检查 `QueryReader`，没有配置时直接 404。然后调用 `requireQueryToken(ctx, controlplane.ReadbackRouteRealtime)`。现在这个函数不只检查“token 是否存在、是否生效”，还会检查这个 token 是否允许访问 `realtime` 这类 readback route family。证据：`仓库: analytics-service, commit: 580c547, file: internal/collectapi/query.go:146-153`、`仓库: analytics-service, commit: 580c547, file: internal/collectapi/query.go:448-470`

```go
if h.opts.QueryReader == nil {
    return h.writeJSON(ctx, fiber.StatusNotFound, ErrorResponse{Error: "not found"})
}
decision, ok := h.requireQueryToken(ctx, controlplane.ReadbackRouteRealtime)
```

`requireQueryToken` 从 `Authorization` header 提取 Bearer token，调用 `authorizeQueryToken`。未知、过期、未生效都返回 `401 unauthorized`；如果 token 本身有效，但 scope 不允许当前 route，则返回 `403 forbidden`。而且这个 scope 判断发生在 source resolution 之前，所以同一个 token 不能再跨 readback family 乱用。证据：`仓库: analytics-service, commit: 580c547, file: internal/collectapi/query.go:448-470`

```go
decision := authorizeQueryToken(bearerToken(ctx.Get("Authorization")), h.opts.QueryCredentials, h.opts.Now())
if decision.State == queryTokenAuthUnknown || decision.State == queryTokenAuthExpired || decision.State == queryTokenAuthNotYetValid {
    _ = h.writeJSON(ctx, fiber.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
    return queryTokenAuthDecision{}, false
}
if !decision.Credential.AllowsReadbackRoute(route) {
    _ = h.writeJSON(ctx, fiber.StatusForbidden, ErrorResponse{Error: "forbidden"})
    return queryTokenAuthDecision{}, false
}
```

`authorizeQueryToken` 仍然用常量时间比较遍历所有凭证，并判断 `NotBefore` / `ExpiresAt`。但凭证结构本身现在已经多了 `Scopes []ReadbackRoute`。`AllowsReadbackRoute` 的逻辑是：如果 scope 为空，保持旧兼容语义，等于允许所有内部读回放路由；如果显式列出 scope，则只允许那几个 route family。证据：`仓库: analytics-service, commit: 580c547, file: internal/collectapi/query_auth.go:13-18`、`仓库: analytics-service, commit: 580c547, file: internal/collectapi/query_auth.go:104-116`

```go
if subtle.ConstantTimeCompare([]byte(value), []byte(credential.Token)) == 1 {
    matched = idx
}
if !credential.NotBefore.IsZero() && now.Before(credential.NotBefore) { ... }
if !credential.ExpiresAt.IsZero() && !now.Before(credential.ExpiresAt) { ... }
```

#### route scope 从哪里配置

主 token 现在可以通过 `ANALYTICS_SERVICE_QUERY_TOKEN_SCOPES` 配置 route scope，多个值用逗号分隔，例如 `realtime,events,properties`。轮换 token 的 JSON 结构也支持 `"scopes": ["events"]` 这样的字段。证据：`仓库: analytics-service, commit: 580c547, file: internal/config/config.go:285-303`、`仓库: analytics-service, commit: 580c547, file: internal/config/config.go:307-349`

这层配置的意思不是“你读哪个 source”，而是“这个 token 允许碰哪一类内部读 API”。真正的 tenant/project/source 仍然要等 write key 解析到 `SourceConfig` 之后才确定。

#### write key 从哪里来

读接口只接受两个位置：

1. `X-SimpleTrack-Write-Key` header
2. `write_key` query 参数

证据：`仓库: analytics-service, commit: 09656b6, file: internal/collectapi/query.go:279-284`

```go
if value := strings.TrimSpace(ctx.Get("X-SimpleTrack-Write-Key")); value != "" {
    return value
}
return strings.TrimSpace(ctx.Query("write_key"))
```

#### 如何通过 resolver 得到 SourceConfig

`resolveQuerySource` 会先取 write key，空值返回 400；再调用 `h.opts.Resolver.ResolveSource`。证据：`仓库: analytics-service, commit: 09656b6, file: internal/collectapi/query.go:230-242`

```go
writeKey := h.queryWriteKey(ctx)
if writeKey == "" {
    _ = h.writeJSON(ctx, fiber.StatusBadRequest, ErrorResponse{Error: "write_key is required"})
    return controlplane.SourceConfig{}, false
}
source, err := h.opts.Resolver.ResolveSource(ctx.Context(), writeKey)
```

`SourceConfig` 是控制面给 analytics-service 的运行时 source 档案。证据：`仓库: analytics-service, commit: 09656b6, file: internal/controlplane/resolver.go:21-43`

```go
type SourceConfig struct {
    WriteKey string `json:"write_key"`
    Enabled bool `json:"enabled"`
    TenantID string `json:"tenant_id"`
    ProjectID string `json:"project_id"`
    SourceID string `json:"source_id"`
    SourceType string `json:"source_type"`
}
```

#### 如何覆盖 tenant/project/source scope

读接口不是“覆盖请求体字段”，而是根本不接受客户端传 tenant/project/source。`handleRealtime` 用 resolver 返回的 `source.TenantID`、`source.ProjectID`、`source.SourceID` 构造 `storage.RealtimeQuery`。证据：`仓库: analytics-service, commit: 09656b6, file: internal/collectapi/query.go:101-109`

```go
records, err := h.opts.QueryReader.ListRealtime(ctx.Context(), storage.RealtimeQuery{
    TenantID:  source.TenantID,
    ProjectID: source.ProjectID,
    SourceID:  source.SourceID,
    Since:     since,
    Limit:     limit,
})
```

这就是读接口的租户隔离点：调用方只给 write key，真正查询哪个 tenant/project/source 由控制面 SourceConfig 决定。

#### 如何构造 Realtime 查询

`since` 默认是当前时间减 30 分钟，`limit` 默认 50 且必须大于等于 1。证据：`仓库: analytics-service, commit: 09656b6, file: internal/collectapi/query.go:17-22`、`仓库: analytics-service, commit: 09656b6, file: internal/collectapi/query.go:90-99`

```go
defaultRealtimeWindow   = 30 * time.Minute
defaultRealtimeQueryCap = 50

since, err := parseQueryTimeOrDefault(ctx, "since", h.opts.Now().Add(-defaultRealtimeWindow))
limit, err := parseQueryLimitOrDefault(ctx, "limit", defaultRealtimeQueryCap)
```

时间格式只接受 RFC3339 / RFC3339Nano。证据：`仓库: analytics-service, commit: 09656b6, file: internal/collectapi/query.go:335-345`

```go
parsed, err := time.Parse(time.RFC3339Nano, value)
parsed, err = time.Parse(time.RFC3339, value)
```

#### 如何调用 QueryReader

`handleRealtime` 调 `ListRealtime`，底层 ClickHouse reader 会先让 builder 生成 Realtime query plan，再执行。query plan 的证据现在会通过内部 readback 响应透出给服务端调试入口，但不会进入 public tracker.js 响应。证据：`仓库: analytics-core, commit: 979a29f, file: storage/clickhouse/event_reader.go:41-90`

```go
plan, err := r.builder.BuildRealtimeQuery(ctx, query)
return r.executePlan(ctx, plan)
```

Realtime 在 query builder 里复用 Events 查询，只是把 `Since` 映射成 `EventListQuery.From`，并使用较小默认 limit；长期实现上它仍然会被标记成独立的 query family，避免后续 Realtime 和 Events 读侧优化证据混在一起。证据：`仓库: analytics-core, commit: ee455ac, file: storage/clickhouse/query_builder.go:251-267`

```go
return b.buildEventsQuery(ctx, storage.EventListQuery{
    TenantID: query.TenantID,
    ProjectID: query.ProjectID,
    SourceID: query.SourceID,
    From: query.Since,
    Limit: b.normalizeLimit(query.Limit, defaultRealtimeLimit),
}, storage.EventQueryFamilyRealtime)
```

#### 返回数据格式

返回结构是 `source + items + since + limit + query_evidence`。`query_evidence` 在 evidence-aware reader 路径下返回；旧 `EventReader` fallback 会省略它。证据：`仓库: analytics-service, commit: 64b0bda, file: internal/collectapi/query.go:61-66, 165-171, 301-314`

```go
type queryRealtimeResponse struct {
    Source        querySourceResponse    `json:"source"`
    Items         []queryEventResponse   `json:"items"`
    Since         string                 `json:"since"`
    Limit         int                    `json:"limit"`
    QueryEvidence *queryEvidenceResponse `json:"query_evidence,omitempty"`
}
```

`EventRecord` 会被映射成 JSON 字段，包括 `visit_id`、`properties`、`user_properties`。证据：`仓库: analytics-service, commit: 09656b6, file: internal/collectapi/query.go:492-514`

```go
return queryEventResponse{
    ID: record.ID,
    EventName: record.EventName,
    VisitID: record.VisitID,
    Properties: queryJSON(record.Properties),
}
```

#### 错误如何映射成 HTTP status/body

| 失败情况 | 证据 | HTTP |
| --- | --- | --- |
| query reader 未配置 | `仓库: analytics-service, commit: 09656b6, file: internal/collectapi/query.go:72-77` | `404 {"error":"not found"}` |
| query token 缺失/未知/过期/未生效 | `仓库: analytics-service, commit: 09656b6, file: internal/collectapi/query.go:212-227` | `401 {"error":"unauthorized"}` |
| write key 缺失 | `仓库: analytics-service, commit: 09656b6, file: internal/collectapi/query.go:234-238` | `400 {"error":"write_key is required"}` |
| write key 无效 | `仓库: analytics-service, commit: 09656b6, file: internal/collectapi/handler.go:333-344` | `401 {"error":"invalid write key"}` |
| source disabled | `仓库: analytics-service, commit: 09656b6, file: internal/collectapi/handler.go:333-344` | `403 {"error":"source is disabled"}` |
| origin 不允许 | `仓库: analytics-service, commit: 09656b6, file: internal/collectapi/query.go:244-248` | `403 {"error":"origin is not allowed"}` |
| query 参数非法 | `仓库: analytics-service, commit: 09656b6, file: internal/collectapi/query.go:92-99` | `400 {"error":"..."}` |
| reader 返回 `ErrInvalidEventQuery` | `仓库: analytics-service, commit: 09656b6, file: internal/collectapi/query.go:286-293` | `400 {"error":"invalid event query: ..."}` |
| reader 其他错误 | `仓库: analytics-service, commit: 09656b6, file: internal/collectapi/query.go:286-293` | `500 {"error":"internal server error"}` |

### 3.3 handleEvents

#### route 是如何注册的

默认 path 是 `/v1/events`，也可由 `ANALYTICS_SERVICE_EVENTS_PATH` 覆盖。证据：`仓库: analytics-service, commit: 09656b6, file: internal/config/config.go:19-20`、`仓库: analytics-service, commit: 09656b6, file: internal/config/config.go:98-99`

注册点：`仓库: analytics-service, commit: 09656b6, file: internal/collectapi/handler.go:155-160`

```go
app.Get(h.opts.EventsPath, h.handleEvents)
```

#### 如何校验 internal query token

与 `handleRealtime` 相同：先要求 `QueryReader`，再要求 query token。证据：`仓库: analytics-service, commit: 09656b6, file: internal/collectapi/query.go:122-130`

```go
if h.opts.QueryReader == nil {
    return h.writeJSON(ctx, fiber.StatusNotFound, ErrorResponse{Error: "not found"})
}
decision, ok := h.requireQueryToken(ctx)
```

query token 来源、生命周期和 route scopes 都来自环境变量。`LoadFromEnv` 读取 `ANALYTICS_SERVICE_QUERY_TOKEN`，解析可选 JSON rotation list，也会解析主 token 的 `ANALYTICS_SERVICE_QUERY_TOKEN_SCOPES`。证据：`仓库: analytics-service, commit: 580c547, file: internal/config/config.go:229-349`

```go
QueryEnabled: envBool("ANALYTICS_SERVICE_QUERY_ENABLED", false),
QueryToken: envString("ANALYTICS_SERVICE_QUERY_TOKEN", ""),
queryCredentials, err := queryCredentialsFromEnv(config.QueryToken)
```

主 token 可配置 `ANALYTICS_SERVICE_QUERY_TOKEN_ID`、`ANALYTICS_SERVICE_QUERY_TOKEN_NOT_BEFORE`、`ANALYTICS_SERVICE_QUERY_TOKEN_EXPIRES_AT`。证据：`仓库: analytics-service, commit: 09656b6, file: internal/config/config.go:274-289`

```go
ID:        envString("ANALYTICS_SERVICE_QUERY_TOKEN_ID", ""),
Token:     primary,
NotBefore: notBefore,
ExpiresAt: expiresAt,
```

轮换 token 可是字符串，也可以是带 `id`、`token`、`not_before`、`expires_at` 的对象。证据：`仓库: analytics-service, commit: 09656b6, file: internal/config/config.go:291-327`

#### write key 从哪里来

与 Realtime 相同：`X-SimpleTrack-Write-Key` header 优先，否则 `write_key` query。证据：`仓库: analytics-service, commit: 09656b6, file: internal/collectapi/query.go:279-284`

#### 支持哪些 query 参数

以 `handleEvents` 当前源码为准：

| 参数 | 是否必填 | 解析位置 | 说明 |
| --- | --- | --- | --- |
| `write_key` | 是 | `仓库: analytics-service, commit: 09656b6, file: internal/collectapi/query.go:212-220` | 也可用 `X-SimpleTrack-Write-Key` header |
| `from` | 是 | `仓库: analytics-service, commit: 09656b6, file: internal/collectapi/query.go:140-145` | RFC3339 / RFC3339Nano，包含起点 |
| `to` | 是 | `仓库: analytics-service, commit: 09656b6, file: internal/collectapi/query.go:146-149` | RFC3339 / RFC3339Nano，不含终点 |
| `limit` | 否 | `仓库: analytics-service, commit: 09656b6, file: internal/collectapi/query.go:150-153` | 默认 100，HTTP 层要求整数且至少 1 |
| `offset` | 否 | `仓库: analytics-service, commit: 09656b6, file: internal/collectapi/query.go:154-157` | 默认 0，HTTP 层要求整数且至少 0 |
| `event_name` | 否 | `仓库: analytics-service, commit: 09656b6, file: internal/collectapi/query.go:165-179` | trim 后进入 `EventListQuery.EventName` |
| `distinct_id` | 否 | `仓库: analytics-service, commit: 09656b6, file: internal/collectapi/query.go:165-179` | trim 后进入 `EventListQuery.DistinctID` |
| `sort_field` | 否 | `仓库: analytics-service, commit: 09656b6, file: internal/collectapi/query.go:165-179` | 传给 analytics-core 白名单校验 |
| `sort_direction` | 否 | `仓库: analytics-service, commit: 09656b6, file: internal/collectapi/query.go:165-179` | 传给 analytics-core 白名单校验 |
| `property_filter` | 否，可重复 | `仓库: analytics-service, commit: 09656b6, file: internal/collectapi/query.go:347-420` | 每个值是 URL 解码后的 JSON |

SaaS 页面还支持 `window=30m/6h/24h/7d`，但这是 SaaS 页面状态，不是 analytics-service `/v1/events` 直接参数。SaaS helper 会把它转换成 `from` / `to`。证据：`仓库: simpletrack-saas, commit: bce33354ae27, file: apps/saas/modules/simpletrack/lib/events-query.ts:13-20`、`仓库: simpletrack-saas, commit: bce33354ae27, file: apps/saas/modules/simpletrack/lib/events-query.ts:72-83`、`仓库: simpletrack-saas, commit: bce33354ae27, file: apps/saas/modules/simpletrack/lib/analytics-readback-core.ts:205-219`

#### 如何做参数白名单和范围限制

HTTP 层做基础范围限制：

- `limit >= 1`
- `offset >= 0`
- `from` / `to` 必须是 RFC3339 / RFC3339Nano
- `property_filter` 最多 5 个

证据：`仓库: analytics-service, commit: 09656b6, file: internal/collectapi/query.go:312-345`、`仓库: analytics-service, commit: 09656b6, file: internal/collectapi/query.go:347-367`

```go
if parsed < min {
    return 0, errors.New(key + " must be greater than or equal to " + strconv.Itoa(min))
}
if len(rawFilters) > defaultPropertyFilterCap {
    return nil, invalidPropertyFilterError("too many property filters: %d > %d", ...)
}
```

analytics-core builder 再做数据库层保护：

- `offset` 不能为负
- `from` 必须早于 `to`
- filters 总数最多 20
- limit 缺失或非法时用默认值，超大时 capped 到 `maxLimit`，默认 1000
- sort field 只允许 `event_time`、`received_at`、`event_name`
- sort direction 只允许 `asc`、`desc`

证据：`仓库: analytics-core, commit: 7c296670842d, file: storage/clickhouse/query_builder.go:15-20`、`仓库: analytics-core, commit: 7c296670842d, file: storage/clickhouse/query_builder.go:144-206`、`仓库: analytics-core, commit: 7c296670842d, file: storage/clickhouse/query_builder.go:334-388`

```go
defaultMaxQueryLimit = 1000
defaultMaxFilters    = 20

if query.Offset < 0 { ... }
if !query.From.Before(query.To) { ... }
if len(query.Filters)+len(query.PropertyFilters) > defaultMaxFilters { ... }
```

#### 如何通过 SourceConfig 限定 property filter

`SourceConfig` 有 `AllowedPropertyFilters` 字段。证据：`仓库: analytics-service, commit: 09656b6, file: internal/controlplane/resolver.go:21-43`

```go
AllowedPropertyFilters []AllowedPropertyFilter `json:"allowed_property_filters"`
```

每个 `property_filter` 先被解析成 JSON：`scope`、`name`、`type`、`op`、`value`。证据：`仓库: analytics-service, commit: 09656b6, file: internal/collectapi/query.go:64-70`

```go
type propertyFilterPayload struct {
    Scope string `json:"scope"`
    Name string `json:"name"`
    Type string `json:"type"`
    Operator string `json:"op"`
    Value any `json:"value"`
}
```

然后必须命中 source allowlist，否则直接返回 `ErrInvalidEventQuery`。证据：`仓库: analytics-service, commit: 09656b6, file: internal/collectapi/query.go:406-420`

```go
if !source.AllowsPropertyFilter(scope, name, valueType) {
    return storage.EventPropertyFilter{}, invalidPropertyFilterError("property filter %d %s.%s %s is not allowlisted", ...)
}
```

`AllowsPropertyFilter` 会按 scope/name 匹配，并可按 value type 限制。证据：`仓库: analytics-service, commit: 09656b6, file: internal/controlplane/resolver.go:80-108`

```go
if filter.Scope != scope || filter.Name != name { continue }
if len(filter.ValueTypes) == 0 { return true }
for _, allowed := range filter.ValueTypes {
    if allowed == valueType { return true }
}
```

最后 HTTP 层还把 selector surface 传给 analytics-core builder，形成第二道 fail-closed guard。证据：`仓库: analytics-service, commit: 09656b6, file: internal/collectapi/query.go:165-179`、`仓库: analytics-core, commit: 7c296670842d, file: storage/clickhouse/query_builder.go:313-332`

当前重要限制：SaaS runtime-source API 返回类型里没有 `allowed_property_filters` 字段。证据：`仓库: simpletrack-saas, commit: bce33354ae27, file: packages/api/modules/simpletrack/runtime-source.ts:30-44`、`仓库: simpletrack-saas, commit: bce33354ae27, file: packages/api/modules/simpletrack/runtime-source.ts:172-188`

```ts
type RuntimeSourceResponse = {
  write_key: string;
  enabled: boolean;
  tenant_id: string;
  project_id: string;
  source_id: string;
  source_type: string;
  allowed_origins: string[];
};
```

因此，从当前源码可推断：如果 `analytics-service` 通过 HTTPResolver 调 SaaS runtime-source，`AllowedPropertyFilters` 会是空列表，带 `property_filter` 的 `/v1/events` 会在 HTTP 层被拒绝。若要让 SaaS UI 的 property filter 可用，控制面响应需要补出 `allowed_property_filters`，或者 memory source config 需要配置该字段。

#### 如何调用 QueryReader

`handleEvents` 把解析后的参数装进 `storage.EventListQuery`，调用 `ListEvents`。证据：`仓库: analytics-service, commit: 09656b6, file: internal/collectapi/query.go:163-181`

```go
records, err := h.opts.QueryReader.ListEvents(ctx.Context(), storage.EventListQuery{
    TenantID: source.TenantID,
    ProjectID: source.ProjectID,
    SourceID: source.SourceID,
    EventName: strings.TrimSpace(ctx.Query("event_name")),
    DistinctID: strings.TrimSpace(ctx.Query("distinct_id")),
    PropertyFilters: propertyFilters,
})
```

ClickHouse reader 不拼业务参数，而是让 builder 生成 plan 后执行。证据：`仓库: analytics-core, commit: 7c296670842d, file: storage/clickhouse/event_reader.go:36-49`

```go
plan, err := r.builder.BuildEventsQuery(ctx, query)
return r.executePlan(ctx, plan)
```

#### 返回数据格式

返回结构是 `source + items + limit + offset + from + to + query_evidence`。`query_evidence` 在 evidence-aware reader 路径下返回；旧 `EventReader` fallback 会省略它。证据：`仓库: analytics-service, commit: 64b0bda, file: internal/collectapi/query.go:51-59, 238-246, 316-329`

```go
type queryEventsResponse struct {
    Source        querySourceResponse    `json:"source"`
    Items         []queryEventResponse   `json:"items"`
    Limit         int                    `json:"limit"`
    Offset        int                    `json:"offset"`
    From          string                 `json:"from"`
    To            string                 `json:"to"`
    QueryEvidence *queryEvidenceResponse `json:"query_evidence,omitempty"`
}
```

OpenAPI 同样定义了 `EventsResponse` / `RealtimeResponse` 的 `query_evidence` 结构。证据：`仓库: analytics-service, commit: 64b0bda, file: api/openapi.yaml:397-538`

#### 错误如何映射成 HTTP status/body

Events 与 Realtime 共用大部分错误映射。额外需要注意：

- `from` / `to` 缺失或时间格式错误，直接在 HTTP 层返回 400。
- property filter JSON、operator、value type、allowlist 错误会包装成 `storage.ErrInvalidEventQuery`，最终 `writeQueryError` 返回 400。
- builder 发现 `from >= to`、sort 不在白名单、offset 为负、property selector 不在白名单，也会通过 `ErrInvalidEventQuery` 返回 400。

证据：`仓库: analytics-service, commit: 09656b6, file: internal/collectapi/query.go:140-160`、`仓库: analytics-service, commit: 09656b6, file: internal/collectapi/query.go:286-293`、`仓库: analytics-core, commit: 7c296670842d, file: storage/clickhouse/query_builder.go:144-206`

## 4. 与外部交互的逻辑、数据和数据格式

### 4.1 `/tracker.js` 请求/响应示例

请求：

```http
GET /tracker.js HTTP/1.1
Host: events.example.com
```

成功响应：

```http
HTTP/1.1 200 OK
Content-Type: application/javascript; charset=utf-8
Cache-Control: public, max-age=300

(function (window) { ... })(window);
```

失败响应：

```json
{
  "error": "tracker script is not configured"
}
```

对应代码证据：`仓库: analytics-service, commit: 09656b6, file: internal/collectapi/handler.go:345-355`

### 4.2 `/v1/realtime` 请求/响应示例

请求：

```http
GET /v1/realtime?write_key=wk_live_123&limit=20 HTTP/1.1
Host: events.example.com
Authorization: Bearer query_token_server_side
```

成功响应：

```json
{
  "source": {
    "tenant_id": "org_123",
    "project_id": "site_123",
    "source_id": "site_123",
    "source_type": "web"
  },
  "items": [
    {
      "id": "evt_001",
      "tenant_id": "org_123",
      "project_id": "site_123",
      "source_id": "site_123",
      "source_type": "web",
      "event_name": "pageview",
      "distinct_id": "dst_abc",
      "session_id": "ses_abc",
      "visit_id": "vis_abc",
      "event_time": "2026-05-07T02:00:00Z",
      "received_at": "2026-05-07T02:00:02Z",
      "properties": {
        "page.path": "/docs"
      }
    }
  ],
  "since": "2026-05-07T01:30:00Z",
  "limit": 20,
  "query_evidence": {
    "family": "realtime",
    "read_path": "fact_events",
    "optimization": "direct_fact_table",
    "effective_limit": 20,
    "offset": 0,
    "has_time_lower_bound": true,
    "has_time_upper_bound": false,
    "time_window_seconds": 0,
    "scalar_filter_count": 1,
    "property_filter_count": 0,
    "uses_property_table": false,
    "sort_field": "event_time",
    "sort_direction": "desc",
    "pressure": "low"
  }
}
```

`query_evidence` 来自 `analytics-core` 的 query plan，不从 SQL 字符串反推；Realtime 的时间窗通常只有 lower bound，所以 `time_window_seconds` 为 0。响应结构证据：`仓库: analytics-service, commit: da852cd, file: internal/collectapi/query.go:62-109, 780-802`；对应回归位于 `仓库: analytics-service, commit: da852cd, file: internal/collectapi/handler_test.go:406-416`

### 4.3 `/v1/events` 请求/响应示例

请求：

```http
GET /v1/events?write_key=wk_live_123&from=2026-05-07T00:00:00Z&to=2026-05-07T02:00:00Z&limit=50&offset=0&event_name=signup_started&sort_field=received_at&sort_direction=desc&property_filter=%7B%22scope%22%3A%22event%22%2C%22name%22%3A%22plan%22%2C%22type%22%3A%22string%22%2C%22op%22%3A%22eq%22%2C%22value%22%3A%22pro%22%7D HTTP/1.1
Host: events.example.com
Authorization: Bearer query_token_server_side
```

成功响应：

```json
{
  "source": {
    "tenant_id": "org_123",
    "project_id": "site_123",
    "source_id": "site_123",
    "source_type": "web"
  },
  "items": [],
  "limit": 50,
  "offset": 0,
  "from": "2026-05-07T00:00:00Z",
  "to": "2026-05-07T02:00:00Z",
  "query_evidence": {
    "family": "events",
    "read_path": "fact_events",
    "optimization": "direct_fact_table",
    "effective_limit": 50,
    "offset": 0,
    "has_time_lower_bound": true,
    "has_time_upper_bound": true,
    "time_window_seconds": 7200,
    "scalar_filter_count": 3,
    "property_filter_count": 1,
    "uses_property_table": true,
    "property_filters": [
      {
        "scope": "event",
        "name": "plan",
        "value_type": "string",
        "operator": "eq"
      }
    ],
    "sort_field": "received_at",
    "sort_direction": "desc",
    "pressure": "high"
  }
}
```

`query_evidence` 中的 `uses_property_table=true` 表示这次查询用到了 typed property 表；`property_filters` 只描述 scope/name/value_type/operator，不返回实际过滤值；`pressure=high` 只是读侧 triage 桶，不是 SLA 或扩缩容信号。响应结构证据：`仓库: analytics-service, commit: da852cd, file: internal/collectapi/query.go:94-109, 780-802`；对应回归位于 `仓库: analytics-service, commit: da852cd, file: internal/collectapi/handler_test.go:645-670`

### 4.4 Authorization Bearer token 示例

读接口使用的是 internal query token：

```http
Authorization: Bearer query_token_server_side
```

Go 侧提取 Bearer token 的函数是通用 helper。证据：`仓库: analytics-service, commit: 09656b6, file: internal/collectapi/handler.go:413-421`

```go
if len(value) < len("Bearer ") || !strings.EqualFold(value[:len("Bearer ")], "Bearer ") {
    return ""
}
return strings.TrimSpace(value[len("Bearer "):])
```

注意：`/collect` 也支持 `Authorization: Bearer <write key>`，但 `/v1/realtime` 和 `/v1/events` 的 `Authorization` 是 query token，不是 write key。读接口的 write key 用 `write_key` query 或 `X-SimpleTrack-Write-Key` header 传递。

### 4.5 query token 生命周期/轮换逻辑

生命周期字段定义在配置层和 collectapi 层。`580c547` 之后，这个结构又多了 route scopes：生命周期解决“这个 token 何时有效”，route scope 解决“这个 token 在有效期内可以碰哪类内部读路由”。证据：`仓库: analytics-service, commit: 580c547, file: internal/config/config.go:34-40`、`仓库: analytics-service, commit: 580c547, file: internal/collectapi/query_auth.go:13-18`

```go
type QueryCredential struct {
    ID        string
    Token     string
    NotBefore time.Time
    ExpiresAt time.Time
    Scopes    []controlplane.ReadbackRoute
}
```

轮换逻辑：

- 第一个匹配的结构化凭证或主 token 是 current。
- 后续 token 被标记为 rotation。
- 匹配到非第 0 个 token 时状态是 `allowed_grace`，会打审计日志。
- 未到 `NotBefore` 返回 401。
- 到达或超过 `ExpiresAt` 返回 401。

证据：`仓库: analytics-service, commit: 580c547, file: internal/collectapi/query_auth.go:33-67`、`仓库: analytics-service, commit: 580c547, file: internal/collectapi/query_auth.go:69-98`、`仓库: analytics-service, commit: 580c547, file: internal/collectapi/query.go:448-470`

### 4.6 与 ClickHouse readback 的数据格式边界

HTTP 层与 ClickHouse 不直接耦合。边界是 `storage.EventRecord`。证据：`仓库: analytics-core, commit: 7c296670842d, file: storage/event_query.go:114-130`

```go
type EventRecord struct {
    ID string
    TenantID string
    EventName string
    Properties string
    UserProperties string
}
```

ClickHouse adapter 扫描自己的 row model，再转成 storage-neutral record。证据：`仓库: analytics-core, commit: 7c296670842d, file: storage/clickhouse/event_reader.go:66-100`

```go
if err := r.db.WithContext(ctx).Raw(plan.SQL, plan.Args...).Scan(&rows).Error; err != nil { ... }
records = append(records, row.toRecord())
```

HTTP 层再把 `Properties` 字符串转成 `json.RawMessage`。证据：`仓库: analytics-service, commit: 09656b6, file: internal/collectapi/query.go:500-514`

```go
if json.Valid([]byte(trimmed)) {
    return json.RawMessage(trimmed)
}
quoted, err := json.Marshal(trimmed)
```

### 4.7 与 SaaS control-plane runtime source config 的交互边界

如果 source resolver 是 HTTP 模式，analytics-service 会 POST write key 到 SaaS runtime-source API。证据：`仓库: analytics-service, commit: 09656b6, file: internal/controlplane/http_resolver.go:194-238`

```go
body, err := json.Marshal(resolveSourceRequest{WriteKey: writeKey})
request.Header.Set("Authorization", "Bearer "+r.bearerToken)
```

SaaS 侧 path 是 `/api/internal/analytics/runtime-source`：Hono app 有 `.basePath("/api")`，并注册 `runtimeSourcePath`。证据：`仓库: simpletrack-saas, commit: bce33354ae27, file: packages/api/index.ts:14-34`、`仓库: simpletrack-saas, commit: bce33354ae27, file: packages/api/modules/simpletrack/runtime-source.ts:55-60`

SaaS runtime-source 校验自己的 service token，按 write key 查 website，返回 source config，支持 ETag / 304。证据：`仓库: simpletrack-saas, commit: bce33354ae27, file: packages/api/modules/simpletrack/runtime-source.ts:61-170`

## 5. 时序图、数据流转图、数据处理方法

### 5.1 `writeTracker` 时序图

```mermaid
sequenceDiagram
    participant Browser as Browser
    participant Fiber as Fiber route
    participant Handler as collectapi.Handler
    participant Asset as TrackerScript bytes

    Browser->>Fiber: GET /tracker.js
    Fiber->>Handler: writeTracker(ctx)
    Handler->>Handler: len(TrackerScript) == 0?
    alt tracker missing
        Handler-->>Browser: 404 {"error":"tracker script is not configured"}
    else tracker ready
        Handler->>Asset: read preloaded bytes
        Handler-->>Browser: 200 application/javascript + Cache-Control max-age=300
    end
```

### 5.2 `handleRealtime` 成功路径时序图

```mermaid
sequenceDiagram
    participant SaaS as simpletrack-saas server helper
    participant Fiber as analytics-service Fiber
    participant Handler as collectapi.Handler
    participant Resolver as controlplane.Resolver
    participant Reader as storage.EventReader
    participant Core as analytics-core ClickHouse reader

    SaaS->>Fiber: GET /v1/realtime?write_key=wk&limit=20 + Bearer query token
    Fiber->>Handler: handleRealtime(ctx)
    Handler->>Handler: requireQueryToken
    Handler->>Resolver: ResolveSource(write_key)
    Resolver-->>Handler: SourceConfig
    Handler->>Handler: parse since / limit
    Handler->>Reader: ListRealtime(RealtimeQuery from SourceConfig)
    Reader->>Core: BuildRealtimeQuery + execute plan
    Core-->>Reader: []EventRecord
    Reader-->>Handler: []EventRecord
    Handler-->>SaaS: 200 {source, items, since, limit}
```

### 5.3 `handleEvents` 成功路径时序图

```mermaid
sequenceDiagram
    participant SaaS as simpletrack-saas server helper
    participant Handler as collectapi.Handler
    participant Resolver as controlplane.Resolver
    participant Source as SourceConfig
    participant Reader as storage.EventReader
    participant Builder as EventQueryBuilder

    SaaS->>Handler: GET /v1/events?... + Bearer query token
    Handler->>Handler: requireQueryToken
    Handler->>Resolver: ResolveSource(write_key)
    Resolver-->>Handler: SourceConfig
    Handler->>Handler: parse from/to/limit/offset
    Handler->>Source: AllowsPropertyFilter for each property_filter
    Handler->>Reader: ListEvents(EventListQuery)
    Reader->>Builder: BuildEventsQuery
    Builder->>Builder: table route + allowlists + limit cap + evidence
    Builder-->>Reader: EventQueryPlan + QueryEvidence()
    Reader-->>Handler: EventQueryResult(records, evidence)
    Handler-->>SaaS: 200 {source, items, limit, offset, from, to, query_evidence}
```

### 5.4 Realtime 数据对象转换图

```mermaid
flowchart LR
    Http["HTTP query<br/>write_key, since?, limit?"] --> Token["query token decision"]
    Token --> Source["SourceConfig<br/>tenant/project/source"]
    Source --> RealtimeQuery["storage.RealtimeQuery<br/>TenantID ProjectID SourceID Since Limit"]
    RealtimeQuery --> Plan["BuildRealtimeQuery<br/>maps Since to EventListQuery.From<br/>QueryEvidence family=realtime"]
    Plan --> Records["[]storage.EventRecord"]
    Records --> Response["queryRealtimeResponse<br/>source items since limit query_evidence"]
```

### 5.5 Events 数据对象转换图

```mermaid
flowchart LR
    Http["HTTP query<br/>from/to/limit/offset/sort/filter"] --> Source["SourceConfig"]
    Source --> PF["[]storage.EventPropertyFilter<br/>source allowlist checked"]
    PF --> EventList["storage.EventListQuery"]
    EventList --> Builder["EventQueryBuilder<br/>route + allowlist + cap"]
    Builder --> Plan["EventQueryPlan<br/>SQL + Args + Limit + QueryEvidence()"]
    Plan --> Records["[]storage.EventRecord"]
    Records --> Response["queryEventsResponse<br/>source items limit offset from to query_evidence"]
```

### 5.6 失败控制流图

```mermaid
flowchart TD
    Start["/v1/realtime 或 /v1/events"] --> Reader{"QueryReader configured?"}
    Reader -- "no" --> NotConfigured["404 not found"]
    Reader -- "yes" --> Auth{"Bearer query token valid?"}
    Auth -- "missing / unknown / expired / not yet valid" --> Unauthorized["401 unauthorized"]
    Auth -- "valid" --> WriteKey{"write_key present?"}
    WriteKey -- "no" --> BadWriteKey["400 write_key is required"]
    WriteKey -- "yes" --> Resolve["Resolver.ResolveSource"]
    Resolve --> ResolveResult{"source result"}
    ResolveResult -- "ErrSourceNotFound" --> InvalidKey["401 invalid write key"]
    ResolveResult -- "ErrSourceDisabled" --> Disabled["403 source is disabled"]
    ResolveResult -- "other resolver error" --> Resolve500["500 internal server error"]
    ResolveResult -- "ok" --> Origin{"Origin allowed?"}
    Origin -- "no" --> Origin403["403 origin is not allowed"]
    Origin -- "yes" --> Parse["parse query params"]
    Parse -- "invalid time / limit / offset" --> InvalidQuery["400 error message"]
    Parse --> Property{"property_filter valid?"}
    Property -- "bad JSON / unsupported / not allowlisted" --> Property400["400 invalid event query"]
    Property -- "ok" --> ReaderCall["QueryReader call"]
    ReaderCall -- "ErrInvalidEventQuery" --> Reader400["400 invalid event query"]
    ReaderCall -- "other error" --> Reader500["500 internal server error"]
    ReaderCall -- "ok" --> OK["200 response"]
```

## 6. 数据流分析

### 6.1 数据点分析

#### DP-01 TrackerScript

| 项 | 内容 |
| --- | --- |
| 定义位置 | `仓库: analytics-service, commit: 09656b6, file: internal/collectapi/handler.go:23-42` |
| 类型 | `[]byte` |
| 用途 | 保存启动时读取的 tracker JS bytes，供 `/tracker.js` 直接返回 |

代码片段：

```go
TrackerScript []byte // TrackerScript is the JavaScript asset returned by TrackerPath
```

数据变化：磁盘文件 `public/tracker.js` 在启动阶段被读成 bytes，运行时不再重新读盘；`writeTracker` 只检查是否为空并发送。

#### DP-02 QueryReader

| 项 | 内容 |
| --- | --- |
| 定义位置 | `仓库: analytics-service, commit: 09656b6, file: internal/collectapi/handler.go:23-42`、`仓库: analytics-core, commit: 7c296670842d, file: storage/event_query.go:142-146` |
| 类型 | `storage.EventReader` interface |
| 用途 | 服务内部 readback 的 storage 边界 |

代码片段：

```go
QueryReader storage.EventReader
type EventReader interface {
    ListEvents(context.Context, EventListQuery) ([]EventRecord, error)
    ListRealtime(context.Context, RealtimeQuery) ([]EventRecord, error)
}
```

数据变化：配置关闭时为 nil，读接口返回 404；配置开启时由 ClickHouse reader 实现。

#### DP-03 QueryCredential / query token

| 项 | 内容 |
| --- | --- |
| 定义位置 | `仓库: analytics-service, commit: 09656b6, file: internal/collectapi/query_auth.go:10-16` |
| 类型 | `QueryCredential` |
| 用途 | 内部读接口认证，不用于 `/collect` 写入 |

代码片段：

```go
type QueryCredential struct {
    ID        string
    Token     string
    NotBefore time.Time
    ExpiresAt time.Time
}
```

数据变化：环境变量解析成 `Config.QueryCredentials`，runtime 转成 `collectapi.QueryCredential`，请求时与 Authorization Bearer 值比较。

#### DP-04 Authorization header

| 项 | 内容 |
| --- | --- |
| 定义位置 | `仓库: analytics-service, commit: 09656b6, file: internal/collectapi/query.go:212-227`、`仓库: analytics-service, commit: 09656b6, file: internal/collectapi/handler.go:413-421` |
| 类型 | HTTP header string |
| 用途 | `/v1/realtime`、`/v1/events` 中承载 query token |

代码片段：

```go
decision := authorizeQueryToken(bearerToken(ctx.Get("Authorization")), h.opts.QueryCredentials, h.opts.Now())
```

数据变化：`Authorization: Bearer xxx` 被剥离成 `xxx`；不符合 Bearer 格式时当作空 token。

#### DP-05 writeKey

| 项 | 内容 |
| --- | --- |
| 定义位置 | `仓库: analytics-service, commit: 09656b6, file: internal/collectapi/query.go:279-284` |
| 类型 | `string` |
| 用途 | 用于解析 SourceConfig，不是查询 scope 本身 |

代码片段：

```go
if value := strings.TrimSpace(ctx.Get("X-SimpleTrack-Write-Key")); value != "" {
    return value
}
return strings.TrimSpace(ctx.Query("write_key"))
```

数据变化：header/query 被归一成一个字符串，传给 resolver；空值直接 400。

#### DP-06 SourceConfig

| 项 | 内容 |
| --- | --- |
| 定义位置 | `仓库: analytics-service, commit: 09656b6, file: internal/controlplane/resolver.go:21-43` |
| 类型 | `controlplane.SourceConfig` |
| 用途 | 服务端可信 source 档案，决定 tenant/project/source、origin、property filter allowlist |

代码片段：

```go
type SourceConfig struct {
    WriteKey string
    Enabled bool
    TenantID string
    ProjectID string
    SourceID string
    SourceType string
    AllowedPropertyFilters []AllowedPropertyFilter
}
```

数据变化：从 memory config 或 SaaS control-plane 解析得到，随后覆盖读查询 scope。

#### DP-07 Realtime query 请求对象

| 项 | 内容 |
| --- | --- |
| 定义位置 | `仓库: analytics-core, commit: 7c296670842d, file: storage/event_query.go:96-103` |
| 类型 | `storage.RealtimeQuery` |
| 用途 | 表示一个 source 的最近事件查询 |

代码片段：

```go
type RealtimeQuery struct {
    TenantID string
    ProjectID string
    SourceID string
    Since time.Time
    Limit int
}
```

数据变化：由 HTTP query 的 `since` / `limit` 加 SourceConfig scope 组合而来。

#### DP-08 Events query 请求对象

| 项 | 内容 |
| --- | --- |
| 定义位置 | `仓库: analytics-core, commit: 7c296670842d, file: storage/event_query.go:78-94` |
| 类型 | `storage.EventListQuery` |
| 用途 | 表示一个 source 的分页原始事件查询 |

代码片段：

```go
type EventListQuery struct {
    TenantID string
    ProjectID string
    SourceID string
    EventName string
    DistinctID string
    From time.Time
    To time.Time
    Limit int
    Offset int
    SortField EventSortField
    SortDirection EventSortDirection
    PropertyFilters []EventPropertyFilter
}
```

数据变化：HTTP query 参数先被解析，再进入 builder 做二次白名单和范围保护。

#### DP-09 property_filter

| 项 | 内容 |
| --- | --- |
| 定义位置 | `仓库: analytics-service, commit: 09656b6, file: internal/collectapi/query.go:64-70`、`仓库: analytics-service, commit: 09656b6, file: internal/collectapi/query.go:347-420` |
| 类型 | repeatable query parameter，每个值是 JSON |
| 用途 | 让 Events 按 typed event/user property 过滤 |

代码片段：

```go
type propertyFilterPayload struct {
    Scope string `json:"scope"`
    Name string `json:"name"`
    Type string `json:"type"`
    Operator string `json:"op"`
    Value any `json:"value"`
}
```

数据变化：URL query 值被 Fiber 解码后 JSON unmarshal，再转换成 `storage.EventPropertyFilter`。

#### DP-10 allowed_property_filters

| 项 | 内容 |
| --- | --- |
| 定义位置 | `仓库: analytics-service, commit: 09656b6, file: internal/controlplane/resolver.go:45-50` |
| 类型 | `[]AllowedPropertyFilter` |
| 用途 | 限定哪些 property 可以进入查询条件 |

代码片段：

```go
type AllowedPropertyFilter struct {
    Scope string `json:"scope"`
    Name string `json:"name"`
    ValueTypes []string `json:"value_types"`
}
```

数据变化：source config 归一化时 scope/value types 变小写；查询时先由 `AllowsPropertyFilter` 判断，再变成 analytics-core selector。

#### DP-11 EventReader 返回值

| 项 | 内容 |
| --- | --- |
| 定义位置 | `仓库: analytics-core, commit: 7c296670842d, file: storage/event_query.go:114-130` |
| 类型 | `[]storage.EventRecord` |
| 用途 | storage-neutral 的事件行结果 |

代码片段：

```go
type EventRecord struct {
    ID string
    EventName string
    DistinctID string
    VisitID string
    Properties string
}
```

数据变化：ClickHouse row model 通过 `toRecord` 转成 EventRecord；HTTP 层再转成 response DTO。

#### DP-12 HTTP response body

| 项 | 内容 |
| --- | --- |
| 定义位置 | `仓库: analytics-service, commit: 64b0bda, file: internal/collectapi/query.go:51-106` |
| 类型 | `queryRealtimeResponse` / `queryEventsResponse` |
| 用途 | 内部 readback JSON 响应 |

代码片段：

```go
type queryEventsResponse struct {
    Source        querySourceResponse    `json:"source"`
    Items         []queryEventResponse   `json:"items"`
    Limit         int                    `json:"limit"`
    QueryEvidence *queryEvidenceResponse `json:"query_evidence,omitempty"`
}
```

数据变化：`EventRecord` 的时间统一格式化为 RFC3339Nano；properties 字符串尽量作为 JSON 返回；`EventQueryEvidence` 被转换成 `query_evidence`，供服务端判断 query shape、读路径和压力分档；如果存在属性过滤，`property_filters` 只包含 scope/name/value_type/operator，不包含实际 value。

#### DP-13 错误响应结构

| 项 | 内容 |
| --- | --- |
| 定义位置 | `仓库: analytics-service, commit: 09656b6, file: internal/collectapi/handler.go:56-59` |
| 类型 | `ErrorResponse` |
| 用途 | 所有稳定错误响应都用 `{"error":"..."}` |

代码片段：

```go
type ErrorResponse struct {
    Error string `json:"error"`
}
```

数据变化：底层错误被映射成稳定 public error；除 validation/query 错误外，内部细节写日志不写响应。

### 6.2 处理动作分析

| 动作 | 输入数据点 | 输出数据点 | 失败路径 | 数据变化说明 |
| --- | --- | --- | --- | --- |
| route registration | `Options.*Path` | Fiber route table | path 为空、不是 `/` 开头、冲突时启动失败 | `TrackerPath` / `RealtimePath` / `EventsPath` 绑定 handler。证据：`仓库: analytics-service, commit: 09656b6, file: internal/collectapi/handler.go:126-169` |
| tracker asset response | `TrackerScript` | JS bytes response | script 为空返回 404 | 启动时文件 bytes 变成 HTTP JS 响应。证据：`仓库: analytics-service, commit: 09656b6, file: internal/collectapi/handler.go:345-355` |
| query auth validation | `Authorization`、`QueryCredentials`、route family | `queryTokenAuthDecision` | 401 unauthorized / 403 forbidden | Bearer token 先做常量时间比较和生命周期判断，再按 token-local route scope 判断是否允许访问当前 readback family。证据：`仓库: analytics-service, commit: 580c547, file: internal/collectapi/query_auth.go:69-98`、`仓库: analytics-service, commit: 580c547, file: internal/collectapi/query_auth.go:104-116`、`仓库: analytics-service, commit: 580c547, file: internal/collectapi/query.go:448-470` |
| write key extraction | header/query | `writeKey` | 空值返回 400 | 读接口不接受 body fallback。证据：`仓库: analytics-service, commit: 09656b6, file: internal/collectapi/query.go:279-284` |
| source resolution | `writeKey` | `SourceConfig` | invalid key 401、disabled 403、resolver error 500 | write key 变成可信 scope。证据：`仓库: analytics-service, commit: 09656b6, file: internal/collectapi/query.go:230-250` |
| source enabled / origin / scope enforcement | `SourceConfig`、`Origin` | scoped query | origin 403 | tenant/project/source 从 SourceConfig 进入 query object。证据：`仓库: analytics-service, commit: 09656b6, file: internal/collectapi/query.go:244-248` |
| query parameter parsing | HTTP query | `since/from/to/limit/offset/sort` | 400 | 时间转 UTC，数字做 min 限制。证据：`仓库: analytics-service, commit: 09656b6, file: internal/collectapi/query.go:296-345` |
| property filter parsing and allowlist enforcement | repeatable `property_filter`、SourceConfig allowlist | `[]storage.EventPropertyFilter` | 400 invalid event query | JSON 标量转 typed property filter，先过 SourceConfig allowlist。证据：`仓库: analytics-service, commit: 09656b6, file: internal/collectapi/query.go:347-420` |
| query reader call | `RealtimeQuery` / `EventListQuery` | `EventQueryResult` 或 `[]EventRecord` | invalid query 400，其他 500 | HTTP 层优先调用 evidence-aware reader，不拼 SQL；旧 reader fallback 只返回记录。证据：`仓库: analytics-service, commit: 64b0bda, file: internal/collectapi/query.go:301-329` |
| response serialization | `EventQueryResult` / `[]EventRecord` | JSON body | JSON 写入失败交给 Fiber error handler | 时间格式化，properties 尽量保持 JSON；evidence-aware 路径额外序列化 `query_evidence`，属性过滤证据只返回 value-free shape。证据：`仓库: analytics-service, commit: 64b0bda, file: internal/collectapi/query.go:51-106, 659-710` |
| error mapping | Go error | HTTP status + `ErrorResponse` | 无 | `ErrInvalidEventQuery` 映射 400，其余 reader error 映射 500。证据：`仓库: analytics-service, commit: 09656b6, file: internal/collectapi/query.go:286-293` |

### 6.3 数据流图示

#### writeTracker 数据流

```mermaid
flowchart LR
    File["public/tracker.js"] --> Runtime["runtime.New os.ReadFile"]
    Runtime --> Options["Options.TrackerScript"]
    Options --> Handler["writeTracker"]
    Handler --> Headers["Content-Type + Cache-Control"]
    Handler --> Body["JS bytes"]
```

#### handleRealtime 数据流

```mermaid
flowchart LR
    Auth["Authorization Bearer"] --> Decision["queryTokenAuthDecision"]
    WriteKey["write_key / X-SimpleTrack-Write-Key"] --> Resolver["Resolver.ResolveSource"]
    Resolver --> Source["SourceConfig"]
    Source --> Query["storage.RealtimeQuery"]
    SinceLimit["since + limit"] --> Query
    Query --> Reader["QueryReader.ListRealtime"]
    Reader --> Records["[]EventRecord"]
    Records --> Response["queryRealtimeResponse"]
```

#### handleEvents 数据流

```mermaid
flowchart LR
    Params["from/to/limit/offset/event_name/distinct_id/visit_id/sort"] --> EventList["storage.EventListQuery"]
    PFParam["property_filter JSON"] --> PFParse["parsePropertyFilter"]
    Source["SourceConfig"] --> PFAllow["AllowsPropertyFilter"]
    PFParse --> PFAllow
    PFAllow --> EventList
    Source --> Scope["TenantID ProjectID SourceID"]
    Scope --> EventList
    EventList --> Reader["QueryReader.ListEvents"]
    Reader --> Response["queryEventsResponse"]
```

#### query token 验证流

```mermaid
flowchart TD
    Header["Authorization header"] --> Bearer["bearerToken"]
    Bearer --> Empty{"empty?"}
    Empty -- "yes" --> Unknown["unknown -> 401"]
    Empty -- "no" --> Compare["constant-time compare all credentials"]
    Compare --> Match{"matched?"}
    Match -- "no" --> Unknown
    Match -- "yes" --> NB{"now before NotBefore?"}
    NB -- "yes" --> NotYet["not_yet_valid -> 401"]
    NB -- "no" --> Exp{"now >= ExpiresAt?"}
    Exp -- "yes" --> Expired["expired -> 401"]
    Exp -- "no" --> Index{"matched index > 0?"}
    Index -- "yes" --> Grace["allowed_grace -> accept + audit"]
    Index -- "no" --> Allowed["allowed -> accept"]
```

#### property filter 验证流

```mermaid
flowchart TD
    Raw["repeatable property_filter"] --> Count{"count <= 5?"}
    Count -- "no" --> TooMany["400 too many property filters"]
    Count -- "yes" --> JSON["json.Unmarshal"]
    JSON --> Required{"scope/name/type present?"}
    Required -- "no" --> BadRequired["400 required"]
    Required -- "yes" --> Operator{"op eq or neq?"}
    Operator -- "no" --> BadOp["400 unsupported operator"]
    Operator -- "yes" --> SourceAllow{"SourceConfig allowlist hit?"}
    SourceAllow -- "no" --> NotAllowed["400 not allowlisted"]
    SourceAllow -- "yes" --> Typed["assign typed scalar slot"]
    Typed --> CoreGuard["EventQueryBuilder propertySelectorAllowed"]
    CoreGuard --> Query["ClickHouse query plan + QueryEvidence()"]
```

## 7. 关键概念通俗解释

### internal readback API 是什么

它是给 SimpleTrack SaaS 服务端页面读回数据的内部接口。`readback` 在这里就是“读回”：`/collect` 负责收事件，`/v1/realtime` 和 `/v1/events` 负责把已入库的事件从 ClickHouse / storage 查询回来，给产品页面展示。

它不等于 event replay。replay 通常表示把历史消息重新投递、重新消费或重新计算；readback 只是读取已经落库的数据。因为它能读用户分析数据，所以必须放在 internal query token 后面，并且只应该由 SaaS 服务端 helper 调用。

### query token 是什么，为什么不能下发到浏览器

query token 是“读数据的服务端钥匙”。SaaS helper 从服务端环境变量读取它，并放进 `Authorization: Bearer ...`。证据：`仓库: simpletrack-saas, commit: bce33354ae27, file: apps/saas/modules/simpletrack/lib/analytics-readback-core.ts:279-326`

如果把 query token 下发到浏览器，任何拿到 token 的人都可能绕过 SaaS 页面权限，直接调用 readback API。当前设计用 `server-only` helper 把它留在服务端。证据：`仓库: simpletrack-saas, commit: bce33354ae27, file: apps/saas/modules/simpletrack/lib/analytics-readback.ts:1-11`

### write key 和 query token 的区别

- write key 是 source 的公开上报 key，浏览器 tracker 可以携带它，用来告诉服务端“我要往哪个 source 写入或读取这个 source 的 readback”。
- query token 是内部读接口认证 token，只能由 SaaS 服务端或可信后端保存；现在它还可以带 route scopes，把 `realtime`、`events`、`properties`、`goals` 这些内部读家族拆开授权。

读接口需要两个值：query token 证明“你有资格读哪类内部路由”，write key 决定“你要读哪个 source”。route scope 先于 source resolution 执行，所以 token 就算是有效的，也不能拿着它去探测自己没被授权的 readback family。证据：`仓库: analytics-service, commit: 580c547, file: internal/collectapi/query.go:448-470`

### Realtime 和 Events 查询的区别

Realtime 是最近窗口，默认最近 30 分钟，默认 limit 50，适合页面上“刚刚有没有收到事件”。证据：`仓库: analytics-service, commit: 09656b6, file: internal/collectapi/query.go:17-22`、`仓库: analytics-service, commit: 09656b6, file: internal/collectapi/query.go:90-119`

Events 是明确时间范围内的分页列表，必须传 `from` 和 `to`，支持 `event_name`、`distinct_id`、`visit_id`、排序、分页和 property filter。当前 `analytics-core` commit `f84024a` 把 typed property filter 收口到 query-builder guardrail：带属性过滤的 Events 查询必须显式带 `from/to`，并且 direct fact-table 路径默认只允许 7 天内窗口；`analytics-service` 先在 `6f6f742` 把这条规则锁到 `/v1/events` handler 回归，最新 `da852cd` 又补上宽时间窗 bounded Events 的 service-side `pressure=high` heuristic 与 exact-boundary 回归。证据：`仓库: analytics-service, commit: 09656b6, file: internal/collectapi/query.go:140-191`；`仓库: analytics-service, commit: da852cd, file: internal/collectapi/handler_test.go:1001-1152, 1968-2011`

### SourceConfig 为什么仍然要参与读接口

因为读接口不能相信客户端传 tenant/project/source。它只接受 write key，然后用 resolver 查出 SourceConfig，再把 SourceConfig 的 tenant/project/source 放进 query object。这样读写两条链路都使用同一套 source 边界。证据：`仓库: analytics-service, commit: 09656b6, file: internal/collectapi/query.go:230-250`、`仓库: analytics-service, commit: 09656b6, file: internal/collectapi/query.go:101-109`、`仓库: analytics-service, commit: 09656b6, file: internal/collectapi/query.go:163-179`

### property filter 为什么要白名单

property filter 最终会影响 ClickHouse 查询。如果任意 property name 都能查，既可能造成高成本扫描，也可能让产品界面暴露尚未设计好的属性面。当前实现要求 source config 明确允许某个 `scope + name + value type`，并且 builder 再检查 selector；此外，typed property filter 还必须保持显式 `from/to`，并默认只允许 7 天内 direct fact-table 窗口。现在 service 层回归还会把 handler 映射出的 `EventListQuery` 送进真实 core planner，确保 HTTP 边界没有把这条规则意外放松；同时 bounded Events `24h+` 的 `pressure=high` 也明确标注为 service-side triage heuristic，而不是 core planner 事实。证据：`仓库: analytics-service, commit: 09656b6, file: internal/controlplane/resolver.go:80-108`、`仓库: analytics-core, commit: 7c296670842d, file: storage/clickhouse/query_builder.go:257-332`、`仓库: analytics-service, commit: da852cd, file: internal/collectapi/query.go:120-130, 826-840`、`仓库: analytics-service, commit: da852cd, file: internal/collectapi/handler_test.go:1001-1152, 1968-2011`

### QueryReader 是什么，为什么 HTTP 层不直接拼 SQL

`QueryReader` 是 storage-neutral 接口。HTTP 层只负责认证和参数解码；ClickHouse 物理表名、字段白名单、排序、limit cap、property 子查询都由 analytics-core 的 builder 管。证据：`仓库: analytics-core, commit: 7c296670842d, file: storage/event_query.go:142-146`、`仓库: analytics-core, commit: 7c296670842d, file: storage/clickhouse/event_reader.go:15-23`

这样做的好处是：HTTP handler 不需要知道 ClickHouse 表怎么命名，也不允许 query string 直接变成 SQL 标识符。

## 8. 总结

`writeTracker` 解决的是“如何稳定交付浏览器 SDK”的问题：启动时把 `public/tracker.js` 读入内存，HTTP 请求时返回 JS 和 5 分钟缓存头。`handleRealtime` 解决的是“如何给 SaaS 服务端快速验证最近是否收到事件”的问题：它用 query token 保护读接口，用 write key 解析 SourceConfig，再查询最近窗口。`handleEvents` 解决的是“如何给 SaaS 服务端查看原始事件列表”的问题：它在同样的认证和 source 边界下，支持时间窗口、分页、排序、事件名、访客和白名单 property filter。它们和 `/collect` 的关系是：`/collect` 把 tracker 或 SDK 发送的事件写入队列和 ClickHouse，`/v1/realtime` 与 `/v1/events` 再从 ClickHouse 把这些已接受事件读回来，用于 P1 的“数据管道活了 + 可验证”的产品闭环。
