# 04-Sessions

> 说明：`官方原话` 只放短英文摘录；`关联现有证据` 只写本地已验证内容。Sessions 当前已有 Phase 07 独立截图和会话列表结果态。

## 这个能力解决什么问题

Sessions 解决的是“把一次访问者的行为串成可读的访问历史”。

它不是简单的列表，而是一个可继续下钻的访问者页面：

1. 先看最近访客
2. 再点进单个访客
3. 再看他的活动历史和会话属性

## 官方原话

> "The Sessions screen displays detailed information about your visitors"

> "Explore your most recent visitors"

> "shows details about a particular visitor"

> "activity history over time"

> "View session properties"

官方同时说明：
> "no cookies or fingerprints"

## 中文解读

Sessions 页本质上是在回答一个问题：

“这个人这次来站上，到底做了什么？”

所以它会同时包含：

- 最近访问者的汇总视图
- 单个访问者的详情页
- 会话属性的结构化展示

因为 Umami 的默认定位是无 cookie、无指纹，所以 Sessions 里的“访客”更适合理解为匿名访问上下文，而不是天然等于 CRM 里的实名客户。要把它升级成业务用户，需要配合 Distinct ID 或 identify 数据。

## 通俗例子

如果一个人先看了首页、再点了注册、又回来看了价格页，Sessions 页会把这条路径串起来。

它比 Events 更像“人”的视角，比 Distinct-ID 更像“这一段访问历史”的视角。

## 它和相邻能力的区别

- `Sessions` 看单次访问者和其历史
- `Distinct ID` 看跨会话身份串联
- `Events` 看所有事件流
- `Realtime` 看当前是否有新数据

如果你要找“最近谁访问了站点”，去 Sessions。
如果你要找“同一个人跨设备还是不是同一个”，去 Distinct-ID。

## 落地动作

1. 把 Sessions 页做成“访客卡片 + 详情下钻”结构
2. 让最近访客和活动历史放在同一条阅读路径里
3. 把会话属性单独放在 Properties 区域
4. 允许按 Distinct-ID 搜索后查看关联会话
5. 在 UI 文案里区分 `Visitor`、`Session`、`User`，避免团队误把匿名访客当成实名账号

## 对 SimpleTrack 的启发

SimpleTrack 如果要做会话视图，最好不要直接上成“原始事件表”。

更好的结构是：

- 先看会话列表
- 再看会话详情
- 再从会话详情跳到事件流

这样用户能从“一个访问者”自然过渡到“一个会话”再过渡到“一个事件”。

如果后续接入回放，Sessions 应该成为“聚合指标 -> 单个访问 -> 回放细节”的中转站，而不是另起一个孤立入口。

## 关联现有证据

### 本地已验证

- `../tracking-demo/bulk-send.mjs` 已经批量发送了 `identify + event`，能为 Sessions / Distinct-ID 链路提供数据基础
- `../tracking-demo/app.js` 已经可在浏览器里触发 `identify()`
- `../snapshots/phase-03-events-and-properties/flow.md` 已经验证 `identify()` 和 Events / Properties 的联动
- `../snapshots/phase-07-traffic-and-behavior-insights/README.md`：`P07-S01` 已完成 Sessions 独立截图，真实浏览器 persona 形成会话列表，session count 有数据。

### 官方文档说明

- Sessions 页提供 visitor activity、visitor profile、session properties 三层信息
- 官方 API 还把 `sessions`, `sessions/:sessionId`, `sessions/:sessionId/activity`, `sessions/:sessionId/properties` 拆成了独立接口

## 官方链接

- [Sessions](https://docs.umami.is/docs/sessions)
- [Sessions API](https://docs.umami.is/docs/api/sessions)
- [Distinct IDs](https://docs.umami.is/docs/distinct-ids)
- [Tracker functions - Session data](https://docs.umami.is/docs/tracker-functions#session-data)

## 继续阅读

- [03-指标对象与Distinct-ID](./03-指标对象与Distinct-ID.md)
- [05-Realtime](./05-Realtime.md)
- [13-Replays](./13-Replays.md)
