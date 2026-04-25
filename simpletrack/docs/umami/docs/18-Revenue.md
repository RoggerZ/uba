# 18-Revenue

## 这个能力解决什么问题

Revenue 解决的是“某个事件到底带来了多少收入”这个问题。

在 Umami 里，它不是单独录一个金额表，而是把收入作为事件数据的一部分，和事件名、币种、时间范围一起分析，从而看出哪些页面、产品或转化动作真的在赚钱。

## 官方原话

> "track financial performance"

> "The insight works by aggregating Revenue and Currency data"

官方还说明了默认币种行为：
> "default to `USD`"

## 中文解读

Revenue 的关键点有三个：

- 它依附在事件上。
- 它需要 `revenue` 和 `currency` 两个字段。
- 它会按币种聚合。

也就是说，Revenue 不是“手动填一张财务表”，而是把交易语义写进事件数据里，然后在报表中汇总。

如果币种代码不被识别，官方文档说明会回落到 `USD`，这点对多币种产品很重要。

这意味着 Revenue 的产品规则必须提前处理币种。否则一个拼错的币种不会只是“显示异常”，还可能被默认归到 USD，导致后续收入解释偏差。

## 通俗例子

用户点击 `checkout-cart` 事件时，顺手带上：

- `revenue: 19.99`
- `currency: USD`

这样 Umami 就不只是知道“有人点了按钮”，还知道“这次点击对应了多少收入”。

## 它和相邻能力的区别

- Revenue 看的是金额。
- Attribution 看的是收入或转化该归功给谁。
- UTM 看的是流量来源标签。
- Links 看的是外链点击和跳转链路。

## 落地动作

- 先统一哪个事件算收入事件，例如购买成功、订阅完成、升级完成。
- 明确币种默认值和多币种策略。
- 让收入事件和业务订单号能在内部系统对上，方便排查。
- 不要把“收入事件”做成纯前端展示，最好能回传到真实订单状态。
- 对金额单位做统一约定，例如统一使用主币种金额而不是 cents，避免同一报表里混用 `19.99` 和 `1999`。
- 收入事件应该只在支付成功或订阅状态确认后产生，不要在点击结账按钮时就记收入。

## 对 SimpleTrack 的启发

SimpleTrack 如果要做收入分析，应该把 Revenue 当成“事件的财务扩展”。

这会比单独加一个收入面板更稳，因为：

- 它和事件模型天然共用。
- 它可以直接进入筛选、分群和归因。
- 它更容易和后续 attribution 联动。
- Revenue 的 MVP 不必先做复杂财务对账，但一定要把“收入事件、币种、订单状态、计划类型”四个字段定义清楚。

## 关联现有证据

### 本地已验证

- `simpletrack/docs/umami/tracking-demo/send-event.mjs`：本地已有可复用的 API send 脚本，说明事件上报链路已经被验证过。
- `simpletrack/docs/umami/tracking-demo/bulk-send.mjs`：已经在 `checkout_completed`、`subscription_upgraded` 等收入类事件上带 `revenue` 和 `currency` 字段，可作为收入事件输入证据。
- `simpletrack/docs/umami/Umami功能深度分析.md`：已记录 2026-04-25 批量上报接受 17,280 条事件，并把 54 条 `checkout_completed` 收入事件列为 Phase 08 Revenue 复验对象。
- `simpletrack/docs/umami/snapshots/phase-08-growth-and-monetization-insights/README.md`：`P08-S08` 已完成 Revenue 页面截图，当前累积站点结果显示 `$11.86k / 355 orders`。

### 官方文档补充

- 官方文档明确要求把收入作为事件动态数据上报；本地现有证据已经覆盖输入侧和 Cloud UI 结果态。
- 当前 Revenue 数值包含多轮 smoke、旧 UA、修正 UA 和重跑后的累积结果，不能等同于一次干净默认样本跑出来的财务结论。
- 当前仍未接真实订单系统，所以只能作为收入事件与报表聚合能力验证，不应写成真实财务对账完成。

## 官方链接

- [Revenue](https://docs.umami.is/docs/revenue)
- [Track events](https://docs.umami.is/docs/track-events)
- [Tracker functions](https://docs.umami.is/docs/tracker-functions)

## 继续阅读

- [17-UTM](./17-UTM.md)
- [19-Attribution](./19-Attribution.md)
- [09-Goals](./09-Goals.md)
