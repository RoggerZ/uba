# 问：字段白名单、过滤参数、Realtime 短窗口和 Events 分页模型怎么理解？

## 答

可以把这四个概念理解成“读侧查询的四道门”。

## 1. 字段白名单

字段白名单决定“哪些字段能查”。

例子：

```text
允许 path -> url_path
允许 event -> event_name
允许 utmSource -> utm_source
```

如果用户筛选“事件名等于 checkout_completed”，后端实际查的是 `event_name = 'checkout_completed'`。

## 2. 过滤参数

过滤参数决定“这次查询想看哪一部分数据”。

例子：

```text
?startAt=...&endAt=...&event=checkout_completed&path=/pricing
```

意思是：在某个时间范围内，只看事件名是 `checkout_completed`，并且路径是 `/pricing` 的数据。

Umami 会先从 query params 中挑出允许的字段，再转换成 PostgreSQL 或 ClickHouse SQL 条件。

## 3. Realtime 短窗口

Realtime 短窗口决定“只看最近一小段时间”。

Umami 的 Realtime API 不让前端随便查一年数据，而是在后端强制使用最近 `REALTIME_RANGE` 分钟。这样 Realtime 页面能快速刷新，不会变成重查询。

例子：

```text
当前时间是 10:30
REALTIME_RANGE = 30
Realtime 查询窗口 = 10:00 到 10:30
```

## 4. Events 分页模型

Events 分页决定“一次只拿一页事件”。

如果事件很多，不能一次返回十万条。分页模型会使用：

- `page`：第几页。
- `pageSize`：每页多少条。
- `orderBy` / `sortDescending`：排序方式。

例子：

```text
?page=2&pageSize=20
```

意思是取第二页，每页 20 条。

## 给 SimpleTrack 的启发

SimpleTrack 的 Events 页面 P1 就应该有时间范围、少量过滤字段和分页。Realtime 页面则应该固定最近窗口，不要让它承担历史分析任务。

## 给 analytics-core 的启发

`EventQueryBuilder` 应把这四件事变成强契约：字段白名单、filter plan、Realtime 固定窗口、Events 分页上限。这样后续 Funnels、Retention、Segments 才能共享同一套查询安全边界。

