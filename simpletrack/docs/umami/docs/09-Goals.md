# 09-Goals

## 这个能力解决什么问题
Goals 用来定义“什么算成功”。它解决的是“流量很多，但到底有没有完成业务目标”的问题。

## 官方原话
> “valuable insights into how well your website is meeting its objectives.”

页面也明确了目标类型：
> “Viewed page or Triggered event”

转化率口径上，官方强调：
> “out of total users within that date range”

## 中文解读
Goals 是把页面访问或事件点击，变成一个可统计的转化目标。你先定义成功动作，再看有多少用户真的完成了它。

这里最容易被忽略的是分母。Goals 不是只数“成功事件发生了几次”，而是把成功用户放回日期范围内的总体用户里计算转化率。所以 Goal 的价值在于把一个业务动作变成可比较的比例。

## 通俗例子
比如访问 `/pricing`、点击 `signup`、触发 `checkout-complete`，都可以成为 Goal。这样你就能知道“访问量高”是不是等于“业务目标完成得好”。

## 它和相邻能力的区别
- 和 Compare 的区别：Compare 看趋势变化，Goals 看是否达成转化。
- 和 Breakdown 的区别：Breakdown 看分布结构，Goals 看单一目标达成率。
- 和 Funnel 的区别：Goals 是单个成功点，Funnel 是多个连续步骤。

## 落地动作
- 先列出 3 到 5 个最重要的业务动作。
- 给每个动作选一个稳定的页面条件或事件名。
- 为每个 Goal 写清楚分母和适用范围，例如“访问过定价页的用户”还是“站点全部用户”。
- 页面型 Goal 适合 URL 稳定的动作，事件型 Goal 适合按钮、表单、支付这类交互动作。
- 把 Goal 结果和漏斗、路径分析放在一起看，避免只看单点数字。

## 对 SimpleTrack 的启发
- SimpleTrack 可以把 Goals 设计成“业务指标定义层”，而不是单纯的事件列表。
- 对页面型目标和事件型目标要用同一套配置表单，减少学习成本。
- 后续如果支持模板，默认模板可以直接覆盖“注册、下单、留资、下载”几类常见目标。
- 目标列表里应显示“目标条件 + 当前转化率 + 日期范围”，否则团队很容易只记住目标名，却忘了它到底在算谁。

## 关联现有证据

### 本地已验证

- 本地证据：`simpletrack/docs/umami/snapshots/phase-03-events-and-properties/` 已验证事件、属性和 `track()` / `identify()` 这类事件采集路径，适合作为 Goal 的底层素材。
- 本地证据：`simpletrack/docs/umami/tracking-demo/` 和 `phase-03` 里的事件截图已经说明仓库里有可用于转化定义的示例事件，如 `demo_track_call`、`demo_signup_click`。
- 本地证据：`simpletrack/docs/umami/snapshots/phase-07-traffic-and-behavior-insights/README.md` 记录 `P07-S06` 已创建 `Checkout Completed Goal`，截图显示 `49 / 1.73k`、转化率 `3%`。

### 官方文档补充

- 官方文档证据：Goals 可以基于页面或事件创建，转化率按日期范围计算。

## 官方链接
- [Goals](https://docs.umami.is/docs/goals)
- [Set up conversion goals](https://docs.umami.is/docs/guides/setup-conversion-goals)
- [Track events](https://docs.umami.is/docs/track-events)

## 继续阅读

- [10-Funnels](./10-Funnels.md)
- [11-Journeys](./11-Journeys.md)
- [playbooks/04-从目标到漏斗到旅程](./playbooks/04-从目标到漏斗到旅程.md)
