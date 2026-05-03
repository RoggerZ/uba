# 问：runQuery 和 storage dispatch 是什么？

## 答

`runQuery` 是 Umami 用来在不同数据库实现之间做选择的函数。storage dispatch 可以理解成“同一个业务动作，根据当前部署配置，派发到不同存储后端执行”。

## 通俗类比

同样是“保存事件”：

- 小部署可以存 PostgreSQL。
- 大部署可以存 ClickHouse。
- 有 Kafka 时，可以先发 Kafka，再由后续链路写 ClickHouse。

`runQuery` 就像一个路由器，看环境变量决定走哪条路。

## Umami 源码中的位置

| 位置 | 作用 |
| --- | --- |
| `references/umami/src/lib/db.ts` | 定义 `runQuery`，根据 `CLICKHOUSE_URL` 和数据库类型选择实现 |
| `references/umami/src/queries/sql/events/saveEvent.ts` | 给 `runQuery` 传入 Prisma 和 ClickHouse 两套写入函数 |
| `references/umami/src/queries/sql/events/getWebsiteEvents.ts` | 给 `runQuery` 传入 PostgreSQL 和 ClickHouse 两套查询函数 |

## 为什么 SimpleTrack 不照搬

Umami 是应用单体，这种方式简单直接。但 `analytics-core` 是数据面核心，需要更清晰的边界：

- collect 不应该知道具体数据库。
- ingestion worker 应该通过 `EventWriter` 写入。
- 查询 API 应该通过 `EventQueryBuilder` 和 `EventReader` 执行。
- 存储实现通过依赖注入装配，而不是散落在业务函数里判断环境变量。

## 给 SimpleTrack 的启发

SimpleTrack 可以在产品层理解“不同部署可用不同存储能力”，但不要把这个复杂度暴露给 P1 用户。用户只需要知道数据能进入 Realtime 和 Events。

## 给 analytics-core 的启发

`analytics-core` 应把 storage dispatch 变成 adapter 装配问题：选择哪个 `EventWriter`、哪个 `EventReader`、哪个 EventBus。这样后续压测、替换 ClickHouse writer 或引入 Kafka 都更可控。

