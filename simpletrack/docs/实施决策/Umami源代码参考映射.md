# Umami 源代码参考映射

> 状态：已确定
> 最近更新：2026-05-03
> 依据：已将 Umami 官方 GitHub 仓库 `master` 分支 `c78ff36db0c82e13c86e5073020472c6546313a3` 克隆为 `references/umami/` 只读参考快照，并完成 `$code-review` 视角源码审阅。
> 影响范围：SimpleTrack P1 tracker、collect、Realtime、Events、`analytics-core` 存储/查询边界、后续 P2/P3 分析能力参考。

## 决策

`references/umami/` 作为 SimpleTrack 的 Umami 官方源码只读参考资产保留在父仓，和 `references/xwl_bi-backend/` 一样只用于学习实现结构，不作为活跃产品模块开发。

SimpleTrack 使用 Umami 的方式是“参考对象体系和实现策略”，不是“复制代码或迁移运行时架构”。

## 参考边界

| 可参考 | 不直接采用 |
| --- | --- |
| tracker 最短采集链路、DOM 事件属性、SPA 路由监听 | Umami tracker 源码逐行复制 |
| `Website / Session / WebsiteEvent / EventData / SessionData` 对象拆分 | Umami 的表名和单租户 website_id 口径直接作为 `analytics-core` 命名 |
| Realtime、Events、Reports 共用 filter/date/query helper | Next.js API route 作为上报热路径 |
| ClickHouse 明细表、属性表、小时聚合表、projection 思路 | 环境变量驱动的 `runQuery` 全局存储切换 |
| Link、Pixel、Boards、Replays、Revenue、Attribution 的后续产品参考 | 把这些能力放进 P1 范围 |

## P1 采用口径

- P1 继续只做“数据管道活了 + 公开产品入口”：tracker、collect、Realtime、Events、Website 设置、Goal 最小闭环。
- `analytics-core` 继续保持 Go library 和框架无关边界：`collect.Handler` 不接收 HTTP 框架对象，EventBus adapter 负责 ack/nack，worker 负责写入 `storage.EventWriter`。
- `simpletrack-anaysitics-service` 承接 SimpleTrack 运行时数据面：write key、Origin/CORS、internal traffic、bot 过滤、tracker 静态交付、调用 `analytics-core` 和本地/小部署 ClickHouse routed table auto migration。
- P1-005C 已落地 `simpletrack-saas` 内部 runtime-source API，并与 `simpletrack-anaysitics-service` 的 HTTP resolver 形成闭环；它只读配置，不创建或修改站点/write key/domain/quota；默认要求 HTTPS，本地 loopback HTTP 必须显式 opt-in，同进程 ingestion 会把动态返回 source 限定在启动 schema surface 内。
- Umami 的 `event_data` 属性表和 ClickHouse 聚合设计作为 P1-002 后续自定义属性写入、查询和 Realtime/Events 验收的重要对照。
- Umami 的 Teams、Boards、Share、Links、Pixels、Replays、Revenue、Attribution 只进入 P2/P3 候选能力池，不改变当前 P1 范围。

## 排入 SimpleTrack 计划和评审的优化项

| Umami 源码结构启发 | SimpleTrack 计划编号 | 需要评审的问题 | `analytics-core` 落点 |
| --- | --- | --- | --- |
| `event_data` / `session_data` typed 属性表 | P1-002A | 已吸收为 collect 属性入口约束、storage typed row 逻辑展开、ClickHouse property writer、`PropertyIndexingEventWriter` 热路径组合和 MySQL `property_indexing_status` guard；真实 ClickHouse e2e 已证明属性随 ingestion 写入、读取和 allowlisted property filter 精确过滤；属性字典和 ClickHouse 去重/聚合优化后置 | `EventPropertyRecord` / `FlattenEventProperties`、`EventPropertyWriter` / `PropertyBatchWriter`、`PropertyIndexingEventWriter`、`PropertyWriteGuard`、后续属性字典 |
| collect 入口补 IP、UA、browser、os、device、geo | P1-002B | 第一版已吸收为 collect pre-queue stage：UA/referrer 可派生为 bounded properties，IP 只允许盐化 hash；浏览器 SDK 已自动收集 allowlisted UTM/click id；geo、browser/os/device 继续评审 | `ClientEnrichmentStage`，不进入 ClickHouse writer |
| bot/IP/internal traffic 过滤 | P1-002B | 第一版已吸收为 bot UA 与 internal CIDR/IP pre-queue 过滤；浏览器 DNT opt-in 已吸收为不发送且不持久化 identity；产品 UI、allow/deny 配置、audit 继续评审 | `TrafficFilterStage` + `FilteredError`，不进入 ClickHouse writer；SDK DNT 在发送前拦截 |
| source + id 或 IP/UA/salt 派生 session/visit | P1-002C | 第一版已吸收为盐化窗口 session resolver，原始 IP/UA 不落库；浏览器 DNT opt-in 已避免本地持久 identity；`visit_id`、salt 轮换、cookie/no-cookie 和 retention 继续评审 | `SessionResolverStage`，后续再决定 `visit_id` 契约 |
| `FILTER_COLUMNS`、operator mapping、分页 | P1-002D、P1-005D | 已吸收为 Events 排序/过滤字段白名单、operator enum、filter 数量上限、typed property filter allowlist 和非法 filter 返回规范；现在 SaaS Events 页面也把 `event_name`、`distinct_id`、`sort_field`、`sort_direction`、`limit`、`offset` 做服务端归一化后再读回放；属性过滤已通过真实 ClickHouse e2e | `EventQueryBuilder` 强契约和测试 + SaaS readback 归一化 |
| Realtime 短窗口与 Events 分页模型 | P1-002E、P1-005D | Realtime 窗口、分页上限、最近事件响应格式；Events 现在用额外读取一条判断 `hasMore`，不做总数查询 | 已通过 `EventReader` 端到端验收，并为 Redis/MySQL/ClickHouse 冷启动增加 readiness retry，后续作为回归入口 |
| tracker auto pageview、custom event、identify | P1-004、P1-005 | P1 已吸收为无依赖浏览器 SDK 和 docs/quickstart，并补齐 opt-in DNT 和 allowlisted UTM/click id；SDK 已从 `analytics-core` 迁出，由 `simpletrack-anaysitics-service` 静态托管，并改为 `data-write-key` 接入；多框架 SDK、CDN 版本化和 performance metrics 后置 | Browser SDK -> `simpletrack-anaysitics-service` `/collect` -> `analytics-core/collect.Request` 协议 |
| ClickHouse 明细表与属性表 DDL | P1-005B | P1 已吸收为 `simpletrack-anaysitics-service` 启动 schema 边界：默认缺表 fail-closed，本地/小部署可显式 auto migrate 当前 runtime config 内所有启用 source 的 event / `_properties` routed tables；生产批量迁移和回滚继续评审 | `ANALYTICS_SERVICE_CLICKHOUSE_AUTO_MIGRATE` + `analytics-core/storage/clickhouse.TableRouter` |
| ClickHouse materialized view、projection、小时聚合 | P1.5-001 | P1 后如何基于压测引入读侧优化，并兼容方案 B 多物理表 | ClickHouse adapter 聚合表、projection、迁移策略 |
| performance metrics | P2-001 | LCP、INP、CLS、FCP、TTFB 是属性、事件类型还是独立模型 | 协议扩展和 P2 查询能力 |

## 证据入口

- 源码快照说明：`references/umami/SIMPLETRACK_REFERENCE.md`
- 源码审阅文档：`simpletrack/docs/umami/docs/21-源代码实现参考.md`
- 关键上游文件：
  - `references/umami/src/tracker/index.js`
  - `references/umami/src/app/api/send/route.ts`
  - `references/umami/prisma/schema.prisma`
  - `references/umami/db/clickhouse/schema.sql`
  - `references/umami/src/queries/sql/events/getWebsiteEvents.ts`
  - `references/umami/src/queries/sql/getRealtimeData.ts`

## 下一步

1. P1-002E 端到端验证已复用现有 `analytics-core` 契约完成，没有绕回 Next.js API route 写库模式；`analytics-core` 已进一步纠偏为 Go library，SimpleTrack 运行时入口进入 `simpletrack-anaysitics-service`。
2. P1-005D 已把 Events 的服务端筛选、排序和分页归一化落在 SaaS readback 边界；后续在 `analytics-core` 自定义属性、session/visit、client enrich 和 query builder 继续推进时，对照 Umami 的 `event_data`、`session_data`、`FILTER_COLUMNS`、operator mapping 和 ClickHouse schema。
3. R3-U1 的 P1 写入组合已完成；P1-002B/C 第一版也已落地到 collect pre-queue stage；P1-005B 已补本地/小部署 routed table auto migration。R3-U2 和 ClickHouse 属性治理、去重、聚合优化进入 P1.5/P2 评审，visit/geo/DNT 等剩余项继续按 R3-U3 到 R3-U5 跟进。
4. 后续刷新 `references/umami/` 时，替换整份快照并同步更新本文件、`SIMPLETRACK_REFERENCE.md` 和实施决策 README。
