# 问：Prisma schema 是什么？

## 答

Prisma schema 是 Prisma ORM 的数据库模型定义文件。它用一份文本文件描述数据库里有哪些表、字段、字段类型、索引和关系，然后 Prisma 根据这份 schema 生成数据库访问客户端和迁移信息。

在 Umami 里，它位于：

```text
references/umami/prisma/schema.prisma
```

## 通俗类比

可以把 Prisma schema 理解成“数据库表结构说明书”。

例如 Umami 的 `WebsiteEvent` model 就是在说明：

- 每条事件有 `event_id`、`website_id`、`session_id`、`visit_id`。
- 每条事件可以有 URL、referrer、UTM、event name、performance 指标。
- 这些字段最终映射到数据库表 `website_event`。

## 它和 ClickHouse schema 的区别

| 项目 | Prisma schema | ClickHouse schema |
| --- | --- | --- |
| 主要面向 | PostgreSQL / Prisma ORM | ClickHouse |
| 文件 | `prisma/schema.prisma` | `db/clickhouse/schema.sql` |
| 表达方式 | Prisma model | SQL DDL |
| 用途 | 生成 ORM client、关系型表模型 | 建分析明细表、属性表、聚合表 |

## 给 SimpleTrack 的启发

SimpleTrack 产品层如果使用 Supastarter / Next.js，可以继续用 Prisma 管控制面数据，比如 workspace、site、subscription、user。事件明细不要简单塞进控制面 Prisma 表里。

## 给 analytics-core 的启发

`analytics-core` 是 Go 数据面核心，不需要使用 Prisma schema。但可以参考 Umami 的 Prisma model 理解事件对象边界，然后用 Go struct、ClickHouse DDL、GORM/MySQL 状态表表达自己的模型。

