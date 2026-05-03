# Umami P1 数据管道源码实现参考

> 状态：已确定  
> 来源：`references/umami/` 只读源码快照，commit `c78ff36db0c82e13c86e5073020472c6546313a3`  
> 用途：为 SimpleTrack 和 `analytics-core` 的 P1 数据管道实现提供技术对照，不搬运第三方源码

## 阅读目标

这组文档只深挖 Umami 中对 P1 最有价值的链路：

- 浏览器 `tracker` 如何采集 pageview、自定义事件、identify 和 performance。
- `POST /api/send` 如何校验 payload、识别 session/visit、解析 URL/UTM/click id。
- `website_event`、`event_data`、`session_data` 如何承接明细事件、事件属性和用户属性。
- PostgreSQL 与 ClickHouse 两套存储模型如何保持相近字段口径。
- Realtime 与 Events 页面如何通过统一过滤器和查询层读取数据。

不在本轮深挖 Boards、Links、Pixels、Teams、Replays、Revenue、Attribution 等后续能力；这些只在必要处作为边界说明出现。

## 章节索引

| 章节 | 主题 | 适合回答的问题 |
| --- | --- | --- |
| [01-整体架构与模块交互](./01-整体架构与模块交互.md) | 运行时模块、周边服务、整体链路 | Umami 的 P1 数据管道由哪些模块组成？ |
| [02-代码结构与目录地图](./02-代码结构与目录地图.md) | 源码目录和文件职责 | 读源码应该从哪些目录和文件开始？ |
| [03-Tracker采集SDK数据流](./03-Tracker采集SDK数据流.md) | 浏览器 SDK 数据点和发送动作 | tracker 如何生成 payload？ |
| [04-Collect-API与会话识别](./04-Collect-API与会话识别.md) | `/api/send`、session、visit、cache token | 服务端如何把请求变成事件？ |
| [05-事件写入与属性展开](./05-事件写入与属性展开.md) | saveEvent、saveEventData、flattenJSON | 事件和属性如何写入不同存储？ |
| [06-数据模型-Postgres与ClickHouse](./06-数据模型-Postgres与ClickHouse.md) | Prisma schema 与 ClickHouse schema | 明细表、属性表、聚合表怎么建模？ |
| [07-Realtime与Events读侧查询](./07-Realtime与Events读侧查询.md) | Realtime、Events、聚合结果 | 接入验收和 Raw Events 如何读数据？ |
| [08-过滤参数与查询构建](./08-过滤参数与查询构建.md) | query filters、字段白名单、动态 SQL | 如何限制查询输入并复用条件？ |
| [09-时序图与数据流图集](./09-时序图与数据流图集.md) | Mermaid 时序和数据流图 | 采集、写入、读取链路怎么串起来？ |
| [10-SimpleTrack-analytics-core对照清单](./10-SimpleTrack-analytics-core对照清单.md) | 可借鉴项、不可照搬项、P1 落地清单 | 具体给 SimpleTrack 和 `analytics-core` 哪些启发？ |

## Q&A

概念解释统一维护在 [Q&A](./Q&A/README.md)。当前已覆盖 identify、PostgreSQL/ClickHouse 存储选择、Prisma schema、字段白名单、Core Web Vitals、过滤参数、Realtime 短窗口、Events 分页、自动采集 SDK、bot/IP 过滤、Zod、storage dispatch、session 隐私机制和 JSON 属性风险。

## 总体结论

Umami 的强参考价值在于“最短可用数据管道”：浏览器 SDK 足够轻，服务端入口能把 pageview、自定义事件、identify、performance 统一进同一个事件模型，读侧又能用 Realtime 和 Events 很快验证数据是否活了。

它不适合被 `analytics-core` 原样照搬。Umami 是 Next.js 单体应用，采集入口、会话识别、过滤、写入分发和存储选择集中在应用层；`analytics-core` 已经建立 `collect.Handler`、EventBus、ingestion worker、`EventWriter`、`EventReader` 和 `EventQueryBuilder` 边界，应把 Umami 的字段口径和查询思路吸收进这些边界，而不是把 route 逻辑迁进去。

## 给 SimpleTrack 的启发

- P1 产品入口要围绕“数据管道活了”：安装 snippet、发送 pageview/custom event、Realtime 出数、Events 可查。
- Marketing site 和 docs/quickstart 应把 `data-website-id`、自动 pageview、自定义事件、identify 的最短链路讲清楚。
- 产品页面优先提供 Realtime 和 Raw Events / Events，先解决接入验收与排障，不把高级报表提前做成 P1 阻塞项。

## 给 analytics-core 的启发

- `EventEnvelope` 应覆盖 Umami 的核心字段口径：tenant/project/source、event id、event name、distinct id、session id、event time、URL、referrer、UTM、属性、性能指标。
- 写入链路继续保持 `collect -> EventBus -> ingestion -> EventWriter`，不要学习 Umami 在 API route 中直接写库的边界。
- 查询链路继续保持 `EventQueryBuilder` 和 `EventReader`，但需要吸收 Umami 的字段白名单、过滤参数、Realtime 短窗口和 Events 分页模型。

## 证据边界

| 项目 | 内容 |
| --- | --- |
| 源码快照 | `references/umami/` |
| 快照说明 | `references/umami/SIMPLETRACK_REFERENCE.md` |
| Tracker | `references/umami/src/tracker/index.js` |
| Collect API | `references/umami/src/app/api/send/route.ts` |
| 写入查询层 | `references/umami/src/queries/sql/` |
| 常量与过滤 | `references/umami/src/lib/constants.ts`、`references/umami/src/lib/request.ts`、`references/umami/src/lib/prisma.ts`、`references/umami/src/lib/clickhouse.ts` |
| 数据模型 | `references/umami/prisma/schema.prisma`、`references/umami/db/clickhouse/schema.sql` |
| SimpleTrack 对照 | `src/analytics-core/README.md`、`src/analytics-core/pkg/contracts/event.go` |
