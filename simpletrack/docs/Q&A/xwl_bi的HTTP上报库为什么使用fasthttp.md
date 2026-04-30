# xwl_bi 的 HTTP 上报库为什么使用 fasthttp

## Q：xwl_bi 现在使用的 HTTP 栈是什么？

A：xwl_bi 后端里实际有两套 HTTP 入口：

1. 后台控制面使用 `github.com/gofiber/fiber/v2`，主要在 `references/xwl_bi-backend/router/index.go` 和各类 controller 中。
2. 数据上报 / collect 入口使用 `github.com/valyala/fasthttp` + `github.com/buaazp/fasthttprouter`，主要在 `references/xwl_bi-backend/cmd/report_server/runtime.go`。

也就是说，xwl_bi 不是单一 HTTP 框架。它的后台 API 偏 Fiber，采集热路径偏 fasthttp。

## Q：这些依赖的活跃度怎么样？

A：按 2026-04-30 的 GitHub API 和 Go module 版本信息看：

| 依赖 | xwl_bi 使用版本 | 当前可见最新版本 / 活跃状态 | 判断 |
| --- | --- | --- | --- |
| `github.com/gofiber/fiber/v2` | v2.22.0 | v2 仍有更新，主线已有 v3；GitHub repo 仍活跃 | 框架本身活跃，但 xwl_bi 锁定版本偏旧 |
| `github.com/valyala/fasthttp` | v1.31.0 | 最新可见 v1.70.0，repo 仍活跃 | 底层库活跃，但更偏底层性能库 |
| `github.com/buaazp/fasthttprouter` | v0.1.1 | 最新 release 仍是 v0.1.1，release 发布时间为 2016-12-20 | 活跃度明显不足，不适合作为新核心仓库默认路由层 |
| `github.com/gin-gonic/gin` | xwl_bi 未使用 | 最新可见 v1.12.0，repo 仍活跃 | 成熟 Web 框架，但不是事件上报热路径的必选项 |

这里说 `fasthttprouter` 活跃度低，不是凭感觉，而是看两个指标：

- 最近 release：`buaazp/fasthttprouter` 最新 release 是 v0.1.1，发布时间是 2016-12-20。
- 最近 push：`buaazp/fasthttprouter` 最近 push 是 2019-04-25。

对比 `fasthttp` 本体：

- `fasthttp` 最新 release 可见到 v1.70.0，发布时间是 2026-04-07。
- `fasthttp` 最近 push 是 2026-04-29。

所以判断是：`fasthttp` 可以继续用于上报热路径；`buaazp/fasthttprouter` 不建议作为新核心仓库的路由层依赖。

参考链接：

- Gin: https://github.com/gin-gonic/gin
- Fiber: https://github.com/gofiber/fiber
- fasthttp: https://github.com/valyala/fasthttp
- fasthttprouter: https://github.com/buaazp/fasthttprouter

## Q：为什么可以继续使用 fasthttp？

A：因为 `fasthttp` 本身仍然活跃，而且它适合事件上报这种高频、低延迟、协议相对简单的热路径。

真正要避免的是直接绑定 xwl_bi 使用的 `buaazp/fasthttprouter`。这个路由库最新 release 仍停在 v0.1.1，发布时间是 2016-12-20，活跃度明显低于 `fasthttp` 本体。

所以结论是：**保留 fasthttp 作为事件上报 HTTP 库，但不沿用低活跃的 fasthttprouter 路由层。**

## Q：为什么不直接使用 Gin？

A：Gin 适合后台 API、管理接口、通用 Web 服务和更复杂的路由/中间件体系，但 P1 的事件上报热路径更简单：

- 明确入口是 `POST /collect`。
- 请求体是标准 JSON 事件。
- 处理逻辑是解码、校验、发布到 `EventBus`。
- 性能瓶颈会更快进入队列、ClickHouse 写入、幂等去重和查询聚合，而不是 Web 路由。

如果未来 `analytics-core` 增加大量管理 API、调试 API 或查询 API，也可以在非热路径上重新评估 Gin。但事件上报入口当前优先用 `fasthttp`。

## Q：为什么不继续用 Fiber？

A：Fiber 本身仍然活跃，而且 xwl_bi 后台控制面用 Fiber 是合理的。但 `analytics-core` 当前不是后台管理系统，而是业务无关的数据面核心服务。

Fiber 可以作为未来后台/管理类 API 的候选，但 P1 collect 热路径更贴近 xwl_bi 原来的上报服务思路，因此先用 `fasthttp`。

## Q：是不是完全不用标准库 HTTP？

A：运行时服务入口不使用标准库 `net/http` 直接写 handler/router，而是使用活跃的第三方 HTTP 库 `fasthttp`。

测试代码可以使用 Go 官方测试工具或 `fasthttp.RequestCtx` 模拟请求，这不影响生产入口选择。

## Q：P1 明确的路由是什么？

A：P1 明确的事件上报路由是：

```text
POST /collect
```

它接收标准 JSON 事件协议，字段包括：

- `id`
- `tenant_id`
- `project_id`
- `source_id`
- `source_type`
- `event_name`
- `distinct_id`
- `session_id`
- `event_time`
- `properties`
- `user_properties`
- `source`

健康检查和查询接口会后续补充，例如 `GET /healthz`、`GET /v1/realtime`、`GET /v1/events`。这些不属于 P1 当前正在落地的事件上报热路径。

## Q：最终决定是什么？

A：已决定：`analytics-core` 的 collect HTTP API 使用 `fasthttp`。

实现边界是：

- `fasthttp` 只在 HTTP 适配层负责 JSON 解码、状态码和响应格式。
- `collect.Handler` 继续保持框架无关，负责标准化、校验和发布到 `EventBus`。
- 不把 Fiber、Gin、fasthttp 的上下文对象传入分析核心逻辑。
- xwl_bi 的 fasthttp collect 链路只参考架构设计，例如请求解码、签名/中间件、producer 编排和启动装配思路。
- 不沿用 `buaazp/fasthttprouter`，避免把低活跃路由库带进新的核心仓库。

## Q：为什么说 fasthttp 只放在 HTTP 适配层，不进入 `collect.Handler`？

A：因为 `collect.Handler` 是数据面核心的入口，不应该认识任何 HTTP 框架对象。

这里的 `collect.Handler` 指的是 `analytics-core` 代码里的一个 Go 类型：

```go
type Handler struct {
    bus eventbus.EventBus
    now Clock
}
```

它的位置在 `src/analytics-core/internal/collect/handler.go`。它不是 fasthttp 的 handler，也不是 Web 路由函数，而是“事件上报核心处理器”。

它负责三件事：

1. 接收已经从 HTTP JSON 解码出来的 `collect.Request`。
2. 调用 `collect.Normalize` 做字段校验和事件标准化。
3. 把标准化后的 `EventEnvelope` 发布到 `EventBus`。

如果让 `collect.Handler` 接收 `*fasthttp.RequestCtx`，会带来几个问题：

- 采集核心会被 fasthttp 绑定，后续想增加 worker、SDK、gRPC、测试入口时都要伪造 HTTP 上下文。
- 单元测试会从“传入结构体，验证事件发布”变成“构造 HTTP 请求上下文”，测试更重。
- 未来如果 HTTP 层调整为别的框架，核心 handler 也要改，破坏 `analytics-core` 的业务无关边界。
- 队列、校验、幂等、ClickHouse 写入这些核心能力和 HTTP 框架无关，不应该让 HTTP 类型穿透进去。

正确分层是：

```text
fasthttp RequestCtx
  -> httpapi 解码 JSON / 写响应
  -> collect.Request
  -> collect.Handler
  -> collect.Normalize
  -> EventBus
```

这样 `fasthttp` 仍然服务于高性能上报入口，但它不会污染 collect、EventBus、ingestion、storage 等核心模块。

## Q：这里说的“污染”是什么意思？

A：这里的“污染”不是说代码脏，也不是说 fasthttp 不好，而是指 **外层框架细节穿透到核心业务模块，导致核心模块和某个框架强耦合**。

更准确的说法是“框架耦合”或“边界穿透”。

举例：

```go
func (h *Handler) Handle(ctx *fasthttp.RequestCtx) error
```

如果 `collect.Handler` 这样设计，它就必须认识 fasthttp。这样会带来连锁影响：

- 单元测试必须构造 `fasthttp.RequestCtx`，而不是直接传一个 `collect.Request`。
- 如果以后接入 gRPC、worker、SDK、本地批处理，就还要伪造 HTTP 上下文。
- 如果上报层以后从 fasthttp 换成别的入口，核心 `collect.Handler` 也要跟着改。
- EventBus、ingestion、storage 本来只关心事件，却会被迫知道 HTTP 框架存在。

更好的设计是：

```go
func (h *Handler) Handle(ctx context.Context, request collect.Request) (EventEnvelope, error)
```

这样 `collect.Handler` 只关心“事件上报请求”，不关心这个请求来自 fasthttp、gRPC、worker、SDK 还是测试。

所以我后续会尽量少用“污染”这种容易误解的词，改说“框架耦合”或“边界穿透”。
