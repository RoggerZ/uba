# 10-Funnels

## 这个能力解决什么问题
Funnel 用来分析用户在一串步骤里的流失位置。它解决的是“用户明明来了，为什么在某一步掉下去了”的问题。

## 官方原话
> “conversion and drop-off rate of users.”

官方文档还强调步骤必须按顺序完成：
> “required order”

## 中文解读
Funnel 是顺序型转化分析。它会统计用户是否按你设定的步骤依次走完，并计算每一步的掉落率和整体转化率。

## 通俗例子
比如“进入商品页 -> 加购 -> 进入结账 -> 支付成功”就是一个典型漏斗。你能一眼看出用户大多卡在哪一步。

## 它和相邻能力的区别
- 和 Compare 的区别：Compare 看前后时间变化，Funnel 看流程步骤流失。
- 和 Breakdown 的区别：Breakdown 看拆分结构，Funnel 看顺序转化。
- 和 Goals 的区别：Goals 只看是否达到某个结果，Funnel 看达成结果之前每一步发生了什么。
- 和 Journeys 的区别：Funnel 是强顺序约束，Journeys 更强调路径探索。

## 落地动作
- 先定义 2 到 7 个清晰步骤。
- 给每一步指定页面或事件条件，确保顺序真实存在。
- 如果某一步掉落异常高，先检查埋点是否准确，再考虑产品体验。

## 对 SimpleTrack 的启发
- SimpleTrack 可以把 Funnel 作为“转化诊断工具”，用于注册、下单、留资、安装完成等关键链路。
- 配置器要同时支持页面和事件，而且要清楚提示顺序约束。
- 结果页最好直接提示“哪一步掉得最多”，不要只给一张图。

## 关联现有证据

### 本地已验证

- 本地证据：`simpletrack/docs/umami/snapshots/phase-06-reports-review/P06-S02-funnels-page.png`、`P06-S07-funnel-config-dialog.png`、`P06-S08-funnel-config-filled.png`、`P06-S09-funnel-result-with-data.png` 已验证 Funnels 页面、配置弹窗和结果页。
- 本地证据：`simpletrack/docs/umami/snapshots/phase-06-reports-review/flow.md` 已记录从入口到 Funnel 配置和结果的操作路径。
- 本地证据：`simpletrack/docs/umami/snapshots/phase-08-growth-and-monetization-insights/README.md` 记录 `P08-S01` 已创建 `Growth Baseline Checkout Funnel`，当前显示 `1.68k -> 45 -> 45 visitors`。

### 官方文档补充

- 官方文档证据：Funnel 必须按顺序完成步骤，窗口时间由 `Window` 控制。

## 官方链接
- [Funnel](https://docs.umami.is/docs/funnel)
- [Build a conversion funnel](https://docs.umami.is/docs/guides/build-a-funnel)
- [Insights](https://docs.umami.is/docs/insights)

## 继续阅读

- [09-Goals](./09-Goals.md)
- [11-Journeys](./11-Journeys.md)
- [12-Retention](./12-Retention.md)
