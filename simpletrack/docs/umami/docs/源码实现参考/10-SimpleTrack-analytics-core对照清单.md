# 10-SimpleTrack 与 analytics-core 对照清单

## 结论

Umami 是实现参考，不是架构蓝本。SimpleTrack 要借鉴它的首价值体验和字段口径；`analytics-core` 要借鉴它的数据点、属性模型、Realtime/Events 查询口径，同时坚持自己的队列、幂等、storage adapter 和 query builder 边界。

## P1 能力对照

| SimpleTrack P1 能力 | Umami 参考 | analytics-core 落点 |
| --- | --- | --- |
| 安装 snippet | `src/tracker/index.js` | tracker SDK / docs quickstart |
| pageview | tracker auto track + `/api/send` event branch | `collect.Request` -> `EventEnvelope` |
| 自定义事件 | `track(name, data)` + `EVENT_TYPE.customEvent` | `EventName` + `Properties` |
| identify | `identify(id, data)` + `saveSessionData` | `DistinctID` + `UserProps` |
| performance | tracker performance branch + `EVENT_TYPE.performance` | 可选 event category / metrics properties |
| Realtime | `/api/realtime/[websiteId]` + `getRealtimeData` | `BuildRealtimeQuery` + `EventReader` |
| Events | `/api/websites/[websiteId]/events` + `getWebsiteEvents` | `BuildEventsQuery` + pagination |
| 事件属性 | `event_data` typed values | 属性表或等价属性索引 |
| 写入可靠性 | Umami 直接写入 / Kafka optional | EventBus + ingestion + EventWriteGuard |
| 动态表路由 | Umami 以 website_id 查询 | `TableRouter` 按 tenant/project/source |

## 可借鉴

| 方向 | 建议 |
| --- | --- |
| 首价值 | 先让用户在 Realtime 和 Events 看到数据 |
| tracker | SDK 保持轻量，自动 pageview + 自定义事件 + identify |
| 字段口径 | URL、referrer、UTM、click IDs、browser/os/device/geo 都值得保留 |
| 属性模型 | 动态属性展开成 typed rows，避免只存不可查询 JSON |
| 读侧 | Realtime 固定短窗口，Events 支持分页、搜索、属性标记 |
| 过滤 | 字段白名单 + operator enum 是后续分析能力基础 |

## 不可照搬

| 风险 | 原因 | analytics-core 应对 |
| --- | --- | --- |
| 单体 route 过重 | Umami `/api/send` 聚合了太多职责 | 拆为 HTTP adapter、collect handler、EventBus、ingestion、storage |
| 环境变量分发后端 | `runQuery` 适合应用单体 | 用接口注入 `EventWriter` / `EventReader` |
| 写入副作用混杂 | event 写入后可能触发 eventData/revenue | 明确写入事务语义和幂等状态 |
| Kafka 错误不透明 | 调用侧需要可靠感知失败 | EventBus handler 返回错误，由 backend retry/dead-letter |
| SQL 字符串扩散 | 动态 filter 容易穿透 | 结构化 query plan + 字段白名单 |
| P1 范围膨胀 | Umami 主干包含大量高级能力 | SimpleTrack P1 只做数据管道和公开入口 |

## P1 验收映射

| 验收项 | Umami 证据 | SimpleTrack / analytics-core 验收 |
| --- | --- | --- |
| 单条 pageview 进入系统 | tracker auto track + `/api/send` | collect 返回成功，EventBus 有消息，ClickHouse 有明细 |
| 自定义事件带属性 | `track(name, data)` + `event_data` | Events 行可查，属性可展开 |
| identify 写用户属性 | `identify(id, data)` + `session_data` | UserProps 可进入属性模型 |
| Realtime 快速出数 | `REALTIME_RANGE` + `getRealtimeData` | 最近窗口事件可查询 |
| Events 可排障 | `getWebsiteEvents` + `hasData` | 分页、搜索、属性存在标记 |
| 幂等消费 | Umami 参考不足 | `EventWriteGuard` 重复事件不重复入库 |
| 队列失败恢复 | Umami 参考不足 | Redis Stream pending 优先、MaxAttempts、dead-letter |

## 给 SimpleTrack 的启发

- 产品叙事聚焦“接入后马上看见数据”，不要把 P1 讲成完整 BI 平台。
- 官网/docs/quickstart 要提供可复制的 tracker 片段、custom event 示例、identify 示例和 Realtime/Events 验收步骤。
- UI 信息架构上，Realtime 是“活性检测”，Events 是“原始证据”，Goal 是“最小业务目标”，高级分析放 P2/P3。

## 给 analytics-core 的启发

- `EventEnvelope` 字段要能覆盖 Umami `website_event` 的核心语义，但命名保持业务无关。
- P1 写入链路必须优先完成可靠性：EventBus、pending retry、dead-letter、EventWriteGuard、BatchWriter。
- P1 读侧先完成 Realtime 和 Events query plan，不急于实现 Funnels/Retention。
- 字段白名单、operator、分页上限和动态表路由要成为 query builder 的强约束，而不是页面层约定。

