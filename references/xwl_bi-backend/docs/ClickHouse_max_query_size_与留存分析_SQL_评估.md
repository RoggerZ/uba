# ClickHouse `max_query_size` 与留存分析 SQL 评估

## 背景

报错原文示例：

```text
Max query size exceeded (can be increased with the `max_query_size` setting): Syntax error: failed at position 262143 ...
```

这类报错的真实含义不是 SQL 本身先写错了，而是 SQL 文本在解析前已经超过了 ClickHouse 允许的单条查询文本大小。

`failed at position 262143` 说明解析器在第 262143 个字符附近失败，结合 ClickHouse 典型行为，可以判断当时的 `max_query_size` 大概率接近 `262144` bytes，也就是 256 KB。

## 在本项目中的触发原因

当前项目的 ClickHouse 连接初始化位于：

- `application/init.go`
- `model/config.go`
- `config/config.json`
- `scripts/config/config.json`

当前触发长 SQL 的高风险点在留存分析：

- `platform-basic-libs/service/analysis/retention.go`

留存分析当前的 SQL 生成方式有两个会把查询文本迅速放大的特征：

1. 对单个起始日 `t`，会把 `windowTime` 内每一天都展开成一段 `xwl_part_event = '...' and toYYYYMMDD(xwl_part_date) = '...'`，并作为 `retention(...)` 的参数列表直接拼到 SQL 文本里。
2. 对查询日期范围内的每一个起始日，又会生成一整段完整查询，最后再通过 `union all` 连接起来。

因此，查询文本长度大致与下面两个量相乘：

- 日期范围天数
- 留存窗口天数

如果事件条件、用户过滤条件、全局过滤条件再比较复杂，SQL 会进一步膨胀，最终触发 `max_query_size`。

## 本次已落地的处理

### 1. 项目配置中显式增加 `maxQuerySize`

在 `model/config.go` 的 `ClickHouseConfig` 中增加：

```go
MaxQuerySize int `json:"maxQuerySize"`
```

并补充默认值逻辑：

```go
func (this ClickHouseConfig) GetMaxQuerySize() int {
    if this.MaxQuerySize <= 0 {
        return 1048576
    }
    return this.MaxQuerySize
}
```

默认值设置为 `1048576`，即 1 MB。

### 2. 将 `max_query_size` 注入 ClickHouse DSN

在 `application/init.go` 的 ClickHouse 连接串中追加：

```text
&max_query_size=1048576
```

本项目使用的是 `github.com/ClickHouse/clickhouse-go v1.5.1`，该驱动支持通过 DSN 传入查询级设置，`max_query_size` 在驱动的 `query_settings.go` 中属于受支持的 query setting。

### 3. 在运行配置文件中设置默认值

已在以下配置文件中增加：

- `config/config.json`
- `scripts/config/config.json`

配置项：

```json
"maxQuerySize": 1048576
```

## 为什么这只是止血，不是根治

提高 `max_query_size` 的作用是让当前超长查询先别在解析阶段直接失败，但它并没有消除问题根源。

根因仍然是留存分析 SQL 生成方式为“二维展开”：

- 横向按窗口天数展开 retention 条件
- 纵向按日期范围展开 `union all`

这会带来几个问题：

1. SQL 文本长度不稳定，输入范围一大就指数级变长。
2. 日志难看，排查困难。
3. 解析成本升高，优化器很难处理过长表达式。
4. 单纯继续增大 `max_query_size` 只能延后失败点，不能保证未来不再撞上。

## 留存分析 SQL 改造方案

下面只给方案，不直接改代码。

### 方案 A：保留 `retention(...)`，但取消“按日期范围 `union all`”

思路：

1. 先用一个子查询把原始事件限定在总时间范围内。
2. 计算每条事件对应的“起始 cohort 日期”。
3. 以 `cohort_date, xwl_distinct_id` 为粒度聚合。
4. 对每个用户在单次扫描中构造完整的留存向量。
5. 最后按 `cohort_date` 分组输出留存结果。

优点：

- 能消掉最外层 `union all`，SQL 长度会明显下降。
- 结果仍然按 cohort date 输出，和现有接口语义接近。
- 前后端返回结构大概率可以保持兼容。

难点：

- 需要重新定义“某条事件归属于哪个 cohort_date”。
- 如果现有 `retention(...)` 的语义高度依赖“单起始日单独计算”，改写时要先确认一致性。

评估：

- 可行性中高。
- 改动中等。
- 值得优先评估。

### 方案 B：放弃 `retention(...)` 的长参数列表，改为“相对天偏移聚合”

思路：

1. 先找出每个用户在首日事件上的 cohort_date。
2. 再把后续目标事件转换为 `dateDiff('day', cohort_date, xwl_part_date)` 形式的相对天数。
3. 只保留 `0 ~ windowTime` 范围内的偏移。
4. 使用 `groupArray`、`groupUniqArrayIf`、`sumIf` 等方式按 offset 汇总。
5. 最终在应用层或 SQL 层重组成现有 `value/ui` 数组。

优点：

- 不再需要把每一天写成一段 `xwl_part_event='...' and toYYYYMMDD(...)='...'`。
- SQL 长度基本和 `windowTime` 线性相关，且常量部分更短。
- 更适合扩展到更大日期范围。

难点：

- 这是逻辑层面的重写，需要重新校验现有结果是否完全一致。
- 需要仔细处理首日事件、次日事件、重复触发、多事件过滤条件的组合。

评估：

- 可行性高。
- 改造收益最高。
- 需要补测试，是我更推荐的长期方案。

### 方案 C：应用层拆批，单日多次查询后合并

思路：

1. 每个起始日单独发一条查询。
2. 后端拿到所有结果后在 Go 层拼接返回。

优点：

- 改造成本最低。
- 现有 SQL 逻辑几乎不动。

缺点：

- 数据库请求次数会随日期范围增长。
- 服务端总耗时和并发压力可能更差。
- 只是把“大 SQL 一次失败”变成“多小 SQL 多次执行”。

评估：

- 可行，但不推荐作为最终方案。
- 适合临时兜底，不适合作为长期实现。

## 推荐结论

推荐顺序：

1. 先通过 `maxQuerySize = 1048576` 做止血。
2. 再按方案 B 重构留存分析 SQL。
3. 如果方案 B 的结果一致性验证成本过高，则退一步采用方案 A。

不建议只继续增大 `max_query_size` 而不改 SQL 生成方式。

## 评审建议

在真正动 `retention.go` 之前，建议先确认下面三件事：

1. 现有留存结果是否严格依赖 ClickHouse `retention(...)` 的返回语义。
2. 前端是否要求 `value`、`ui` 的数组长度和索引含义完全保持不变。
3. 是否允许在后端增加一层结果重组，而不是完全依赖 SQL 一次性输出最终数组。

如果这三点都可以接受，方案 B 可以推进。
