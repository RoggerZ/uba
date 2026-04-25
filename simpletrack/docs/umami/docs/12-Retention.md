# 12-Retention

## 这个能力解决什么问题
Retention 用来衡量用户会不会回来。它解决的是“用户第一次来了之后，还会不会继续用”的问题。

## 官方原话
> “tracking how often users return.”

官方文档还把它定义为 cohort 分析：
> “cohort analysis chart”

图表读法上，官方说明：
> “Rows represent cohorts”

## 中文解读
Retention 是留存分析。它把某个月的首访用户按 cohort 组织起来，然后看他们在后续几天、几周是否再次访问。

读 Retention 图时，可以把它想成一张“回来了吗”的矩阵：每一行是一批起始用户，每一列是后续时间窗口，单元格里的数字表示这批人在那个时间点还剩多少活跃度。

## 通俗例子
比如 1 月新注册的一批用户，到了 2 月还能回来多少，就属于 Retention 要看的内容。

## 它和相邻能力的区别
- 和 Compare 的区别：Compare 看不同时间段的整体变化，Retention 看首访 cohort 的回访行为。
- 和 Goals 的区别：Goals 看是否完成目标，Retention 看是否持续回来。
- 和 Journey 的区别：Journey 看路径，Retention 看复访。

## 落地动作
- 先确认“首访”的定义，再确认留存统计窗口。
- 用月度 cohort 或周度 cohort 开始，不要一开始就做太复杂的留存矩阵。
- 留存分析要等观察窗口足够长再下结论，例如今天刚来的用户不能立刻判断 7 日留存。
- 如果产品有登录体系，优先用稳定身份或 Distinct ID 辅助解释，否则只能从访问层面理解回访。
- 结合 Goal 和 Funnel 看，避免只把“回来”当作唯一成功标准。

## 对 SimpleTrack 的启发
- SimpleTrack 可以把 Retention 作为“长期价值层”，和首日激活、转化漏斗一起看。
- 留存图最好支持按 cohort 月份切换，并保留默认解释文案。
- 如果未来支持账号型产品，这一页可以直接成为产品健康度主视图之一。
- MVP 阶段可以先不做复杂留存矩阵，但事件和身份模型要提前留好口径，否则后面补 Retention 会很痛。

## 关联现有证据
- 本地证据：`simpletrack/docs/umami/snapshots/phase-06-reports-review/P06-S04-retention-page.png` 已验证 Retention 页面可进入。
- 本地证据：`simpletrack/docs/umami/snapshots/phase-08-growth-and-monetization-insights/README.md` 记录 `P08-S03` 已完成 Retention 截图，留存矩阵已有 cohort 数据。
- 边界说明：当前 Retention 可作为 cohort 留存结构参考，但长期留存结论仍需要更长观察窗口，不能只用当天样本判断 7 日或 30 日留存。
- 官方文档证据：Retention 以月份为输入，使用 cohort chart 展示回访趋势。

## 官方链接
- [Retention](https://docs.umami.is/docs/retention)
- [Insights](https://docs.umami.is/docs/insights)
- [Sessions](https://docs.umami.is/docs/sessions)

## 继续阅读

- [16-Cohorts](./16-Cohorts.md)
- [09-Goals](./09-Goals.md)
- [10-Funnels](./10-Funnels.md)
