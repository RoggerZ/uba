# 问：Umami 的事件存储是 PostgreSQL 和 ClickHouse 二选一吗？

## 答

从源码看，Umami 支持两类存储路径：

- 没有配置 `CLICKHOUSE_URL` 时，主要走 PostgreSQL / Prisma。
- 配置了 `CLICKHOUSE_URL` 时，事件查询和写入走 ClickHouse 路径，必要时还可以通过 Kafka 把消息送入 ClickHouse。

所以对“上报事件数据”来说，可以理解为 Umami 在部署层面支持 PostgreSQL 和 ClickHouse 两种后端选择；它不是每条事件同时默认写两份。具体走哪条路径，由 `references/umami/src/lib/db.ts` 里的 `runQuery` 根据环境变量决定。

## 通俗类比

PostgreSQL 像一个通用业务数据库，适合自托管、低门槛、小中规模部署。

ClickHouse 像专门为海量分析查询准备的列式仓库，适合高吞吐写入和大范围聚合查询。

## Umami 源码中的位置

| 位置 | 作用 |
| --- | --- |
| `references/umami/src/lib/db.ts` | `runQuery` 判断是否有 `CLICKHOUSE_URL` |
| `references/umami/src/queries/sql/events/saveEvent.ts` | 同一个 `saveEvent` 同时提供 Prisma 和 ClickHouse 实现 |
| `references/umami/prisma/schema.prisma` | PostgreSQL 模型 |
| `references/umami/db/clickhouse/schema.sql` | ClickHouse 表模型 |

## 给 SimpleTrack 的启发

SimpleTrack 面向生产分析产品时，不应把 PostgreSQL 当作高频事件明细的最终主路径。控制面可以用 PostgreSQL，事件分析热路径应优先围绕 ClickHouse。

## 给 analytics-core 的启发

`analytics-core` 不应照搬 Umami 的环境变量二选一分发。更合适的做法是用接口注入：`EventWriter` 负责写入，`EventReader` 负责查询，ClickHouse adapter 是 P1 主路径，MySQL/GORM 用于幂等状态和控制类元数据。

