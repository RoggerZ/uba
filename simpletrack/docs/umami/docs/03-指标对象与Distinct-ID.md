# 03-指标对象与Distinct-ID

> 说明：`官方原话` 只放短英文摘录；`关联现有证据` 只写本地已验证内容。这里会刻意区分“事件属性”“会话属性”“Distinct ID”。

## 这个能力解决什么问题

指标对象和 Distinct-ID 解决的是“同一个人、同一个会话、同一条事件，应该怎么串起来”。

如果没有这层，页面只知道“来了多少条数据”；有了这层，系统才知道：

1. 这条事件属于哪个会话
2. 这个会话是不是同一个人
3. 事件上的属性该不该和会话属性分开看

## 官方原话

> "A Distinct ID is a unique identifier assigned to a user"

> "across multiple sessions"

> "Add the id property to the payload."

> "This property has a 50 character limit."

> "Save data about the current session"

> "Session data"

官方 guide 也强调登录用户识别：
> "pass in your own unique user identifier"

## 中文解读

Distinct-ID 是用户级别的稳定标识，适合跨会话串联。

Session data 是会话级别的数据，适合描述“这一趟访问里发生了什么”。

Event data 则是事件级别的数据，适合描述“这一次动作带了什么上下文”。

三者的边界要分清，不然报表会越来越乱。

Distinct-ID 的关键不是“能不能识别出真实姓名”，而是“能不能稳定、可控地把同一个业务用户串起来”。如果直接使用邮箱、手机号这类个人信息，后续隐私和合规压力会很大；更稳的做法是用内部用户 id、哈希 id 或不可反查的业务标识。

## 通俗例子

同一个用户在不同设备上：

- 用 Distinct-ID 你可以把他串成同一个人
- 用 Session data 你可以知道某一次访问的浏览器、城市、邮箱或计划类型
- 用 Event data 你可以知道这次点击属于哪个按钮、哪个 surface、哪个 variant

## 它和相邻能力的区别

- `Distinct ID` 解决跨会话身份
- `Sessions` 解决单个会话和访问者详情
- `Events` 解决单次动作和动作属性
- `Metric definitions` 解决指标本身怎么定义

不要把“人”“会话”“事件”都叫成一个“用户标签”。

## 落地动作

1. 先定义哪个字段是稳定身份，比如登录 id 或业务 id
2. 再定义哪些字段是会话属性，比如 plan、cohort、region
3. 最后定义哪些字段是事件属性，比如 surface、variant、step
4. 给 Distinct-ID 保留 50 字符限制意识
5. 让 Events、Sessions、Reports 都能读到同一套语义
6. 不要把邮箱、姓名、手机号、token 直接当作 Distinct-ID 或 session data 上报

## 对 SimpleTrack 的启发

SimpleTrack 最值得学的是“分层建模”：

- `event` 层看动作
- `session` 层看一次访问
- `distinct` 层看同一个人

这样后面不管是做漏斗、留存还是路径分析，都不会把字段语义搅乱。

SimpleTrack 的身份模型应该允许“匿名访问者 -> 登录用户”的升级，但不要要求新手在首日接入时就完成复杂账号打通；先让匿名会话和事件稳定，再逐步加身份。

## 关联现有证据

### 本地已验证

- `../tracking-demo/bulk-send.mjs` 里已经同时使用了 `distinctId`、`plan`、`cohort`、`userType`
- `../tracking-demo/app.js` 已经支持 `identify()` 的 id + data 组合
- `../snapshots/phase-03-events-and-properties/flow.md` 记录了 `identify()` 之后再看 `Properties`
- `../snapshots/phase-03-events-and-properties/P03-S10-properties-event-picker.png`
- `../snapshots/phase-03-events-and-properties/P03-S11-properties-property-picker.png`
- `../snapshots/phase-03-events-and-properties/P03-S12-properties-with-data.png`

### 官方文档补充

- Distinct IDs 页把 `id` 定义成跨多个 sessions 的用户标识
- Sessions 页把会话属性放在 `Properties` tab 下
- Tracker functions 页把 `umami.identify()` 说明成“给当前 session 赋值/保存 session data”

## 官方链接

- [Distinct IDs](https://docs.umami.is/docs/distinct-ids)
- [Tracker functions - Session data](https://docs.umami.is/docs/tracker-functions#session-data)
- [Event data](https://docs.umami.is/docs/event-data)
- [Metric definitions](https://docs.umami.is/docs/metric-definitions)
- [Identify logged in users](https://docs.umami.is/docs/guides/identify-logged-in-users)

## 继续阅读

- [04-Sessions](./04-Sessions.md)
- [09-Goals](./09-Goals.md)
- [playbooks/03-从过滤到细分用户](./playbooks/03-从过滤到细分用户.md)
