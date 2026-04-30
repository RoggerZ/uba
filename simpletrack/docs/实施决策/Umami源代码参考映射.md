# Umami 源代码参考映射

> 状态：已确定
> 最近更新：2026-05-01
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
- `analytics-core` 继续保持 Go 核心链路和框架无关边界：`collect.Handler` 不接收 HTTP 框架对象，EventBus adapter 负责 ack/nack，worker 负责写入 `storage.EventWriter`。
- Umami 的 `event_data` 属性表和 ClickHouse 聚合设计作为 P1-002 后续自定义属性写入、查询和 Realtime/Events 验收的重要对照。
- Umami 的 Teams、Boards、Share、Links、Pixels、Replays、Revenue、Attribution 只进入 P2/P3 候选能力池，不改变当前 P1 范围。

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

1. P1-002 端到端验证继续复用现有 `analytics-core` 契约，不绕回 Next.js API route 写库模式。
2. 在 `analytics-core` 自定义属性与 query builder 继续推进时，对照 Umami 的 `event_data`、`FILTER_COLUMNS`、operator mapping 和 ClickHouse schema。
3. 后续刷新 `references/umami/` 时，替换整份快照并同步更新本文件、`SIMPLETRACK_REFERENCE.md` 和实施决策 README。
