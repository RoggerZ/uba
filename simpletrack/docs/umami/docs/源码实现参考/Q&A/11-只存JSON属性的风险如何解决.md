# 问：只存 JSON 属性的风险如何解决？

## 答

只存 JSON 属性的问题是：写入很方便，但后续查询、过滤、索引和聚合会越来越难。

比如所有事件属性都塞进一列：

```json
{
  "plan": "pro",
  "amount": 99,
  "currency": "USD"
}
```

短期看很灵活，长期会遇到：

- 很难高效查 `plan = pro`。
- 数字、日期、布尔值类型容易混乱。
- 属性名没有治理，`plan`、`Plan`、`subscription_plan` 可能同时出现。
- Breakdown、Segments、Funnels 需要属性过滤时成本很高。

## Umami 的解决思路

Umami 把动态属性展开到 `event_data`：

- 属性名进入 `data_key`。
- 文本值进入 `string_value`。
- 数字值进入 `number_value`。
- 日期值进入 `date_value`。
- 类型进入 `data_type`。

这样既保留动态属性，又让后续查询有更稳定的结构。

## analytics-core 的解决办法

推荐采用分层策略：

1. 原始 properties 保留为调试和回放输入。
2. 常用属性同步展开成 typed rows 或 ClickHouse 可查询结构。
3. 建立属性字典和元数据，记录属性名、类型、首次出现时间、最近出现时间。
4. Query builder 只允许查询已登记或白名单内属性。
5. 对高频属性建立专门索引、projection 或物化视图。

## 给 SimpleTrack 的启发

产品层要提供事件属性字典，让用户知道哪些属性已经进入系统、是什么类型、在哪些事件上出现。否则 Events 能看到数据，但后续分析做不深。

## 给 analytics-core 的启发

P1 可以先保证属性入库和 Events 展示；P1.5/P2 应补属性字典、属性过滤、ClickHouse 优化和高频属性治理。不要把“JSON 能存”误认为“分析能力已经具备”。

