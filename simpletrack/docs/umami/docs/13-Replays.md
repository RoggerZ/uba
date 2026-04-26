# 13-Replays

## 这个能力解决什么问题
Replays 用来把用户行为重新放回上下文里看。它解决的是“知道发生了什么，但不知道当时页面上到底发生了什么”的问题。

## 官方原话
> “every click, scroll, and navigation captured as it happened”

官方配置项里还出现了这些关键控制：
> “Sample Rate” / “Mask Level” / “Max Duration”

## 中文解读
Replays 本质上是会话回放。它把用户在页面上的操作轨迹重建出来，帮助你理解用户为什么点击、为什么卡住、为什么离开。

它不是传统意义上的“录一个视频文件”，而是把用户交互、DOM 状态和时间线组合成可回看的行为证据。因此 Replays 的产品设计一定会碰到三个问题：采多少、遮哪些、留多久。

## 通俗例子
比如某个表单总是提交失败。回放可以帮你看到用户是否反复修改字段、卡在某个错误提示上，或者在某一步直接退出。

## 它和相邻能力的区别
- 和 Sessions 的区别：Sessions 更偏“会话列表和摘要”，Replays 更偏“上下文回放”。
- 和 Journeys 的区别：Journey 看路径，Replays 看过程细节。
- 和 Goals 的区别：Goals 看是否达成，Replays 看为什么没达成。

## 落地动作
- 优先把回放用于高价值页面，如注册、支付、下单、留资。
- 在回放中同时保留页面状态和关键交互事件，避免只看点击轨迹。
- 对隐私字段做脱敏，确保回放不会暴露敏感数据。
- 上线前先决定采样率，不要默认全量记录所有会话。
- 启用回放后再产生的会话才有意义，不能期待它回填历史行为。
- 对登录、支付、个人资料、表单输入这类区域，默认使用遮罩或排除规则。

## 对 SimpleTrack 的启发
- SimpleTrack 如果要做 Replays，核心不是“录屏”，而是“可解释的交互事件重建”。
- 应该把回放和事件、目标、路径分析连起来，形成一条排障链。
- MVP 可以先从关键表单与关键转化页开始，不必一开始覆盖全站。
- 产品配置页至少要暴露采样率、遮罩策略、最长回放时长和保留周期，否则后续隐私与成本都会失控。

## 关联现有证据

### 本地已验证

- 本地已验证：`simpletrack/docs/umami/snapshots/phase-08-growth-and-monetization-insights/README.md` 记录 `P08-S04` 已完成 Replays 页面截图，页面可达。
- 本地边界：当前账号/套餐提示 `This feature requires a Business plan subscription.`，网站元数据为 `replayEnabled=false`，所以只能证明入口和套餐限制，不能写成“回放播放态已验证”。
- 本地可借鉴证据：`simpletrack/docs/umami/snapshots/phase-07-traffic-and-behavior-insights/` 已有 Sessions、Realtime、Performance 结果态，可作为后续把 Replays 接入排障链路的前置证据。

### 官方文档补充

- 官方文档证据：`/docs/replays` 是独立页面，明确说明了启用方式、采样率、遮罩级别、最长时长、保留周期和回放访问入口。

## 官方链接
- [Replays](https://docs.umami.is/docs/replays)
- [Sessions](https://docs.umami.is/docs/sessions)
- [About](https://docs.umami.is/docs/about)
- [Insights](https://docs.umami.is/docs/insights)

## 继续阅读

- [04-Sessions](./04-Sessions.md)
- [05-Realtime](./05-Realtime.md)
- [06-Performance](./06-Performance.md)
