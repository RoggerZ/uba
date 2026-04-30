18

## Q

1. `api/analysis/RetentionList` 和 `api/analysis/RetentionList` 两个接口请求耗时久，请分析整个调用链路，找出慢的地方和优化策略，现在不做代码修改。

## A

### 1. 结论

题面里写的是两个 `RetentionList`，但从代码看，后端实际只有一个接口 `/api/analysis/RetentionList`。

当前“感觉很慢”其实是两件事叠加：

- 同一个页面动作会触发同一个接口多次请求。
- 单次 `RetentionList` 请求本身就很重，主耗时在 ClickHouse 查询。

主瓶颈不在路由和控制器，而在 `platform-basic-libs/service/analysis/retention.go` 里拼出来的大 SQL：

- 按日期逐天 `UNION ALL`
- 每天都对 `xwl_event{appid}` 做一次 `GROUP BY xwl_distinct_id`
- 每个格子还额外计算一份 uid 数组

### 2. 调用链路

#### 2.1 前端

看板卡片路径：

1. `vue/src/views/dashboard/index.vue`
2. `refreshCard(card)`
3. `vue/src/views/dashboard/components/analysis/retention.vue`
4. `go()`
5. `vue/src/api/analysis.js`
6. `RetentionList(data)`

留存分析页路径：

1. `vue/src/views/behavior-analysis/retention.vue`
2. `go()`
3. `vue/src/api/analysis.js`
4. `RetentionList(data)`

#### 2.2 后端

1. `router/analysis.go`
2. `BehaviorAnalysisController.RetentionList`
3. `analysis.NewAnalysisByCommand(analysis.RetentionComand, ctx.Body())`
4. `analysis.GetAnalysisRes(i)`
5. `platform-basic-libs/service/analysis/retention.go`
6. `(*Retention).GetList()`
7. `(*Retention).GetExecSql()`
8. `(*Retention).getSqlByDate()`
9. `db.ClickHouseSqlx.Select(&res, sqls, args...)`

#### 2.3 中间件

`/api/analysis/*` 这组接口还会经过：

1. `middleware.FilterAppid`
2. `middleware.OperaterLog`
3. `fiber limiter`
4. `middleware.Timer`

其中：

- `FilterAppid` 会调 `myapp.GetAppidsByToken()`，每次请求都查一次 MySQL `app` 表。
- `OperaterLog` 会往 MySQL `gm_operater_log` 插一条记录。
- `Timer` 会在请求结束后记录 body 和耗时。

这几层有开销，但不是主瓶颈。原因是日志里同样能看到一些小范围 `RetentionList` 查询只需要几十毫秒。

### 3. 慢点分析

#### 3.1 前端会放大请求次数

`vue/src/views/dashboard/components/analysis/retention.vue` 里有多个直接触发 `go()` 的入口：

- `beforeMount()` 会调一次 `go()`
- `watch.filterDate` 且 `immediate: true`，命中时会再调一次
- `watch.data` 且 `deep: true`，数据变化时会再调一次
- 看板刷新按钮 `refreshCard()` 也会直接调组件的 `go()`

另外，不管是留存分析页还是看板卡片，只要配置了 `compareDate`，前端还会顺序再发一次 `RetentionList` 用于对比数据。

所以用户感知到的慢，经常不是“1 次请求慢”，而是“同一动作把同一个重查询打了 2 到 3 次，甚至更多次”。

这不是单次 SQL 的根因，但会明显放大总等待时间。

#### 3.2 `RetentionList` 的 SQL 模型天然重

核心代码在 `platform-basic-libs/service/analysis/retention.go`：

- `parseReqDate()` 会把请求区间拆成每天一个 cohort
- `GetExecSql()` 会对每一天都调用一次 `getSqlByDate()`
- 最后用 `strings.Join(sqlArr, "union all")` 拼成一个大 SQL

也就是说，如果日期区间是 `2025-12-24 ~ 2026-03-23`，总共 90 天，就会拼出 90 段子查询，再 `UNION ALL`。

每一段子查询都会：

- 扫描 `xwl_event{appid}`
- 按 `xwl_distinct_id` 聚合
- 调用 ClickHouse `retention(...)`
- 产出当天 cohort 的 value 和 ui

这会带来两个问题：

- 同一批事件数据会被反复扫描
- 同一批用户会被反复聚合

对于日志里常见的 `windowTime = 3` 请求，可以理解成：

- 90 个 cohort 查询
- 每个 cohort 都要看当前日及后续几天的数据
- 同一自然日的数据会被重复读多次

这就是单次请求慢的第一大原因。

补充一点，SQL 里虽然有 `limit 1000`，但它出现在外层汇总之后。每个 cohort 分段最终本来就只返回 1 行，所以这个 `limit` 对降低扫描量和聚合量几乎没有帮助。

#### 3.3 查询在总览阶段就把 uid 明细一起算了

`retention.go` 不仅算了人数：

```sql
array(sum(r[1]), sum(r[2]), ...) as value
```

还额外算了 uid 数组：

```sql
array(groupUniqArray(if(r[1]=1, xwl_distinct_id, null)), ...) as ui
```

这意味着每个 cohort 日、每个留存天数，都要额外聚合一份用户 ID 列表。

以 `windowTime = 3` 为例，每一行会生成 5 组 value 和 5 组 ui。

如果 cohort 日期有 90 天，那么后端要额外聚合：

- `90 x 5 = 450` 组 uid 数组

这部分很重，原因是：

- `groupUniqArray` 本身就是高成本聚合
- 聚合结果会占更多内存
- 最后还要序列化成 JSON 返回给前端

而前端页面平时首先看的只是人数和比例，uid 只在点格子钻取时才真正需要。也就是说，当前实现把“明细查询”的成本提前压到了每一次总览查询里。

这是单次请求慢的第二大原因。

#### 3.4 表结构与留存查询模式不匹配

`platform-basic-libs/service/app/app_service.go` 里，事件表 `xwl_event{appid}` 的建表方式是：

- `PARTITION BY (toYYYYMM(xwl_part_date))`
- `ORDER BY (toYYYYMM(xwl_part_date), xwl_part_event)`

而留存查询的核心模式是：

- 按时间范围过滤
- 按事件过滤
- 最终按 `xwl_distinct_id` 聚合

也就是说，当前表结构更偏向“按月份、按事件”查，不偏向“按用户”查。

这不会导致功能错误，但会让留存这种大量依赖 `xwl_distinct_id` 聚合的查询天然更吃亏。

#### 3.5 用户筛选和分群会进一步放大成本

如果请求带了 `whereFilterByUser`，代码会再拼一层用户表子查询：

```sql
xwl_distinct_id in (
  select xwl_distinct_id
  from (select xwl_distinct_id, argMax(...) ... from xwl_user{appid} group by xwl_distinct_id)
  where ...
)
```

如果请求带了 `userGroup`，还会先：

1. 查 MySQL `user_group`
2. 解压 `user_list`
3. 再拼出一大段 uid 过滤条件

当前日志里的慢查询样本大多没有开启这两个筛选，所以它们不是主因；但一旦业务上叠加用户属性或分群过滤，留存查询会更慢。

#### 3.6 同步日志也会增加额外开销

当前请求链里还有两类同步日志：

- `retention.go` 在执行前打印完整 SQL 和参数
- `middleware.Timer` 在请求结束后记录完整 body

由于 `RetentionList` 的 SQL 很长，尤其是大日期范围时，日志字符串本身就很大。它不是主瓶颈，但会额外增加字符串拼接、序列化和磁盘 I/O。

### 4. 日志证据

根据 `cmd/manager/logs/manager/info.log*` 中现有的 `RetentionList` 记录，可以得到几个比较明确的现象。

#### 4.1 常见耗时分布

本地日志样本里一共抓到 518 条 `RetentionList` 记录，粗略统计结果如下：

- 平均耗时约 `756ms`
- `500ms ~ 800ms` 的最多
- `>= 1s` 的有 128 条
- 最大值约 `18.13s`

这说明它不是偶发慢，而是稳定偏慢，并且存在明显长尾。

#### 4.2 同时段下，`RetentionList` 明显重于 `EventList`

在 `2026-03-23 18:21` 左右的日志里：

- `EventList` 有 1 年时间范围的请求，耗时约 `121ms`
- 同时段的 `RetentionList`，同一个 app 下耗时约 `2.0s ~ 2.5s`

这说明问题不在 Fiber、路由或 Controller，而在留存 SQL 的计算模型。

#### 4.3 同一类请求既出现过几十毫秒，也出现过多秒甚至十几秒

日志里还能看到：

- 单天、`windowTime = 1` 的 `RetentionList` 有过 `31ms`
- 同类请求也出现过 `5.79s` 甚至 `18.13s`

并且在极慢样本附近，其他接口也同时出现了明显变慢。

所以要把问题拆成两层来看：

- 稳态慢：查询模型本身重
- 长尾慢：环境资源争用放大了尾延迟

### 5. 瓶颈排序

按影响程度排序，当前瓶颈基本可以定为：

1. 逐天 cohort + `UNION ALL` 带来的重复扫描和重复聚合
2. 总览阶段就计算所有格子的 uid 数组
3. 前端重复触发同一个重查询
4. 用户属性筛选 / 分群带来的额外子查询和过滤
5. MySQL 鉴权、操作日志和大文本日志的附加成本

### 6. 优化策略

当前不改代码，这里只给优化方向，按收益从高到低排序。

#### 6.1 先把“总览统计”和“uid 明细”拆开

建议后续改成两段式：

1. 总览接口只返回人数和比例
2. 点击格子时，再单独请求该格子的 uid 明细

这能直接减少：

- `groupUniqArray`
- 大结果集序列化
- 大响应体传输

这是最值得优先做的改造。

#### 6.2 把逐天 `UNION ALL` 改成单次扫描思路

后续建议不要继续沿用“每天一段子查询再 union all”的方案。

更合理的方向是：

1. 先把原始事件按 `distinct_id + event_day + event_name` 做日级预聚合
2. 再基于这张日级用户事件表计算留存矩阵
3. 尽量让一条查询处理整段日期，而不是拆成几十段

这样可以减少：

- 重复扫描
- 重复 `GROUP BY xwl_distinct_id`
- 超长 SQL 文本

#### 6.3 为留存分析增加预聚合层

如果留存是核心分析能力，建议不要长期直接从原始事件表现算。

可以考虑的方向：

- 日级用户事件明细表
- 物化视图
- 汇总表
- bitmap / aggregate state 方案

留存分析天然更适合读“日级用户行为快照”，不适合每次直接扫原始流水。

#### 6.4 收敛前端重复触发

看板卡片层建议后续收敛为“同一轮渲染只触发一次查询”：

- 初始化阶段不要让 `beforeMount`、`watch.filterDate`、`watch.data` 重复触发
- 刷新按钮只刷新当前卡片
- `compareDate` 只在确实配置时再请求

这类优化不一定降低单次 SQL 耗时，但会明显改善用户体感。

#### 6.5 给重复参数请求加缓存

仓库里其实已经有分析缓存基础设施：

- `platform-basic-libs/service/analysis/cache.go`

但 `RetentionList` 目前没有接入。

看板场景非常适合短 TTL 缓存，因为同一块卡片会反复请求高度重复的参数。

#### 6.6 降低同步日志开销

后续建议把大日志降级成摘要日志，例如：

- 只记 SQL 摘要和参数 hash
- 保留耗时
- 不记录完整长 SQL

这不是第一优先级，但在高频看板场景下会有帮助。

### 7. 最终判断

这次 `RetentionList` 慢，不是单点问题，而是三层叠加：

1. 前端会把同一个接口放大成多次请求
2. 后端 `RetentionList` 本身采用了高成本 SQL 模型
3. 查询还在总览阶段把所有 uid 明细一起算出来了

如果后续允许改造，建议优先顺序是：

1. 先拆分“总览统计”和“uid 明细”
2. 再消除前端重复请求
3. 最后重构留存查询模型和预聚合层

这三个动作的收益最大，也最符合当前代码结构。
