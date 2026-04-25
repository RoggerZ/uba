# 11-Journeys

## 这个能力解决什么问题
Journey 用来回答“用户接下来通常会去哪里”。它解决的是“用户不是按固定漏斗走，而是有多条真实路径”的问题。

## 官方原话
> “the top user journeys for your website.”

官方文档还说明它会展示起点、掉点和终点：
> “where users start off, where they drop off, and where they end up”

## 中文解读
Journey 是路径探索视图。它不像 Funnel 那样强制每一步都按顺序命中，而是让你查看用户在几个关键步骤之间的真实流动。

## 通俗例子
你想知道用户从首页之后是去价格页、博客页还是注册页，Journey 就会把这些路径摊开给你看。

## 它和相邻能力的区别
- 和 Funnel 的区别：Funnel 是固定步骤的转化链路，Journey 是多路径的流向分析。
- 和 Breakdown 的区别：Breakdown 看按维度拆分的结果，Journey 看按步骤连接起来的路径。
- 和 Compare 的区别：Compare 看时间差，Journey 看行为路径。

## 落地动作
- 先确定关键起点和终点，再让中间步骤保持尽量少而清晰。
- 优先看“高频路径”和“高掉点路径”，再考虑是否要改版。
- 如果路径太散，先用 Breakdown 找出主流入口，再回到 Journey 里看流程。

## 对 SimpleTrack 的启发
- SimpleTrack 可以把 Journey 作为“用户路径解释层”，用于解释为什么某些目标达成率高或低。
- 起点和终点应该支持页面与事件双模式，和 Funnel 保持一致。
- 如果后续要做 AI 辅助分析，Journey 很适合作为“下一步建议”的输入层。

## 关联现有证据
- 本地证据：`simpletrack/docs/umami/snapshots/phase-06-reports-review/P06-S03-journeys-page.png` 和 `P06-S10-journeys-with-data.png` 已验证 Journeys 页面与有数据状态。
- 本地证据：`simpletrack/docs/umami/snapshots/phase-06-reports-review/flow.md` 明确记录了进入 Journeys 和查看路径图的操作路径。
- 官方文档证据：Journey 支持 3 到 7 个步骤，并允许设置 Start Step 与 End Step。

## 官方链接
- [Journey](https://docs.umami.is/docs/journey)
- [Insights](https://docs.umami.is/docs/insights)
- [Build a conversion funnel](https://docs.umami.is/docs/guides/build-a-funnel)

## 继续阅读

- [10-Funnels](./10-Funnels.md)
- [12-Retention](./12-Retention.md)
- [13-Replays](./13-Replays.md)
