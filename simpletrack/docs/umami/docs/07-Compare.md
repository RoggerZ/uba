# 07-Compare

## 这个能力解决什么问题
Compare 用来把当前时间段和过去时间段放在一起看。它解决的不是“现在数据有多少”，而是“这段时间比之前变好了还是变差了”。

## 官方原话
> “comparing websites stats and metrics across previous date ranges.”

官方页面还明确说明了比较口径：
> “Previous period” / “Previous year”

结果呈现上，官方还强调：
> “percentage change indicator”

## 中文解读
Compare 是一个“时间对时间”的对照视图。它会把同一网站在两个时间范围内的指标叠在一起，让你快速判断趋势是不是向着预期方向走。

它真正有价值的地方不是多画一条线，而是把“现在值”和“对照值”放到同一个判断框里。用户不需要自己口算“比上周多了多少”，系统直接给出变化方向和变化比例。

## 通俗例子
比如你想知道本周首页浏览量是不是比上周高，或者这次活动周的访问量有没有超过去年同期。Compare 就是干这个的。

## 它和相邻能力的区别
- 和 Breakdown 的区别：Compare 看“前后两个时间段的差异”，Breakdown 看“同一时间段里按维度拆分后的结构”。
- 和 Goals 的区别：Compare 关注趋势变化，Goals 关注某个动作是否达成转化。
- 和 Funnel 的区别：Funnel 关注步骤漏斗，Compare 不要求步骤顺序，只看整体指标变化。

## 落地动作
- 在分析前先统一口径：日期范围、对比范围、核心指标。
- 优先拿 Compare 做周环比、月环比、同比。
- 对比口径要固定成少数默认项，例如 `previous period`、`previous year`，避免每个报表都出现无法解释的自定义比较。
- 对流量强周期业务，要优先用同比或同星期结构的周期，避免把周末和工作日硬比。
- 如果发现异常波动，再下钻到 Breakdown、Funnels 或 Journeys。

## 对 SimpleTrack 的启发
- SimpleTrack 可以把 Compare 做成“默认趋势对照层”，放在每个报表的顶部。
- 对比维度最好支持“上一周期”和“去年同期”两种默认模式。
- 结果展示不要只给数值，还要给涨跌方向和百分比变化。
- 当对比出现异常时，下一步按钮最好直接引导到 Breakdown 或 Filters，而不是让用户自己猜该去哪页。

## 关联现有证据
- 本地证据：`simpletrack/docs/umami/snapshots/phase-07-traffic-and-behavior-insights/README.md` 记录 `P07-S04` 已完成 Compare 截图，当前周期指标和路径表有数据。
- 本地可借鉴证据：`simpletrack/docs/umami/snapshots/phase-06-reports-review/README.md` 和 `flow.md` 已验证 Reports 入口、日期范围切换与高阶报表进入路径。
- 官方文档证据：Compare 页面与 Guide 页面都明确描述了跨时间段对比。

## 官方链接
- [Compare](https://docs.umami.is/docs/compare)
- [Compare traffic across time periods](https://docs.umami.is/docs/guides/compare-traffic-periods)
- [Insights](https://docs.umami.is/docs/insights)

## 继续阅读

- [08-Breakdown](./08-Breakdown.md)
- [09-Goals](./09-Goals.md)
- [14-Filters](./14-Filters.md)
