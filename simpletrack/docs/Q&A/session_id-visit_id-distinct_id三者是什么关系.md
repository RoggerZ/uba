# session_id、visit_id、distinct_id 三者是什么关系

## Q：这三个字段分别回答什么问题？

A：三者都和“用户/访问识别”有关，但语义不同。

| 字段 | 回答的问题 | 生命周期 | 是否进入长期模型 |
| --- | --- | --- | --- |
| `distinct_id` | “是谁” | 尽量稳定，可能跨多次访问、多次会话存在 | 是 |
| `session_id` | “SDK 或服务端这次会话来源是什么” | 一次浏览器会话、App 会话或服务端会话 | 是 |
| `visit_id` | “分析口径里的一次访问是什么” | 默认 30 分钟访问窗口，可由 SDK 显式传入或服务端派生 | 是 |

一句话：`distinct_id` 是人或设备的长期身份，`session_id` 是采集来源保留的会话语义，`visit_id` 是分析系统用于 Realtime、Events、Sessions、Funnels、Journeys、Retention 的标准访问键。

## Q：为什么不能只用 `session_id`？

A：因为 `session_id` 更偏 SDK/运行时来源语义，不一定等于分析口径里的“一次访问”。

例如：

- 浏览器 SDK 可以有自己的 session 机制。
- 服务端 SDK 可能没有浏览器会话。
- App SDK 的 session 生命周期和 Web 的 session 生命周期不一样。
- 后续如果做 no-cookie、跨设备身份合并、服务端补齐，`session_id` 的来源会更复杂。

所以 SimpleTrack 保留 `session_id`，但不把它强行当成所有分析的 visit key。长期方案是把 `visit_id` 单独作为 canonical analytics visit key。

## Q：`visit_id` 当前怎么定？

A：P1 已定稿：`visit_id` 必须在写入前确定，并持久化进入事件表和属性表。

规则是：

1. SDK 或服务端请求如果传入合法 `visit_id`，优先保留。
2. 如果请求没有传 `visit_id`，`simpletrack-anaysitics-service` 在 collect 阶段装配 `analytics-core` 的 `VisitResolverStage`。
3. `VisitResolverStage` 基于 `tenant_id / project_id / source_id / distinct_id / session_id / event_time bucket / server-only visit salt` 派生稳定 `visit_id`。
4. ClickHouse 事件表和 `_properties` 表都存储 `visit_id`。
5. Realtime / Events 查询直接读取存储字段，不再依赖 readback 临时派生。

这里的 readback 指“从存储读取数据给 Realtime / Events 页面展示”的读回放链路。之前的临时方案是在读回放时补一个 provisional `visit_id`，现在这个方案已经被长期方案替代。

## Q：Umami 是怎么处理的？

A：Umami 的 ClickHouse schema 中有显式 `visit_id` 字段，并在事件表排序、聚合和 visit/session 相关查询中使用它。

这给 SimpleTrack 的启发是：`visit_id` 不应该只是页面展示时临时算出来的字段，而应该成为事件明细里的正式列。

SimpleTrack 不照搬 Umami 的完整产品结构，但采纳这个关键方向：访问分析所依赖的 visit key 要进入写入链路和存储模型。

## Q：Litlyx 是怎么对应的？

A：Litlyx 产品表达更偏 `workspace / product / raw events`，适合参考短接入链路、Raw Events 验收、空态/示例态/真实态教育方式。

目前不强行推断 Litlyx 内部 schema。对于 SimpleTrack 来说，只做产品概念对齐：

| Litlyx 产品概念 | SimpleTrack / analytics-core 对齐 |
| --- | --- |
| workspace | `tenant_id` |
| product | `project_id` |
| raw events | Events / Realtime 读侧 |
| 真实产生事件的端 | `source_id` |

`visit_id` 是 SimpleTrack 自己的数据面长期字段，不需要依赖 Litlyx 是否显式暴露类似字段。

## Q：SimpleTrack P1 具体做什么？

A：P1 做长期收益的实现，而不是临时展示方案。

- collect 请求契约支持 `session_id`、`visit_id`、`distinct_id`。
- `analytics-core` 的 `EventEnvelope`、`collect.Request`、`storage.EventRecord`、`EventPropertyRecord` 都包含 `visit_id`。
- `simpletrack-anaysitics-service` runtime source config 提供 server-only `visit_salt` 和 `visit_window_seconds`。
- ClickHouse event table 和 `_properties` table 增加 `visit_id`。
- `BatchWriter` / `PropertyBatchWriter` 写入 `visit_id`。
- `EventReader` 读取存储字段。
- `EventQueryBuilder` 支持按 `visit_id` 白名单过滤。
- `simpletrack-saas` 的 Realtime / Events 页面展示真实存储的 `visit_id`。

## Q：P1.5 / P2 延后什么？

A：这些先不阻塞 P1：

- 用户自定义 session timeout UI。
- 复杂 no-cookie / cookie 混合策略。
- 跨设备身份合并。
- 登录用户身份和匿名身份 merge。
- visit salt rotation 和历史 backfill。
- Sessions 独立页面。
- Funnels / Journeys / Retention 产品化页面。
- 基于 visit 的复杂 cohort 和 attribution UI。

P1 先保证事件入库时有稳定 `visit_id`，后续高级分析才能站得住。

## Q：geo 怎么处理？

A：方向是参考 Umami，但结合 xwl_bi 的离线文件优势。

- 优先识别可信平台 headers，例如部署平台或反向代理提供的国家/地区头。
- 缺失可信 headers 时，使用离线 MaxMind / GeoLite2 `.mmdb` 或 xwl_bi 已有离线 IP 文件查询。
- 原始 IP 不落库。
- 只落分析需要的派生字段，例如 `geo.country`、`geo.region`、`geo.city`。

这样既能保留地域分析能力，也避免把原始 IP 变成长期存储风险。
