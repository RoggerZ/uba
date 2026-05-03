# 问：identify 在 Umami 是什么概念？

## 答

`identify` 是 Umami tracker 暴露给业务页面的一个用户识别函数，用来把“当前浏览器会话”关联到一个更稳定的用户标识，并写入用户或会话属性。

通俗讲：

- 不调用 `identify` 时，Umami 主要靠匿名 session 判断“同一个访客这一段时间做了什么”。
- 调用 `identify("user_123", { plan: "pro" })` 后，Umami 可以知道这个匿名访问者现在对应业务里的 `user_123`，并记录 `plan=pro` 这样的属性。
- 它不是登录系统，也不是权限系统，只是分析系统里的身份线索。

## Umami 源码中的位置

| 位置 | 作用 |
| --- | --- |
| `references/umami/src/tracker/index.js` | 暴露 `window.umami.identify(id, data)`，发送 `type=identify` |
| `references/umami/src/app/api/send/route.ts` | 在 `identify` 分支调用 `saveSessionData` |
| `references/umami/src/queries/sql/sessions/saveSessionData.ts` | 把 identify 的属性展开并写入 `session_data` |
| `references/umami/prisma/schema.prisma`、`references/umami/db/clickhouse/schema.sql` | 定义 `SessionData` / `session_data` |

## 举例

```js
window.umami.identify('user_123', {
  plan: 'pro',
  signupChannel: 'google',
});
```

这类数据后续可以用于回答：

- 某个付费用户是否完成了关键事件？
- Pro 用户和 Free 用户行为是否不同？
- 某个渠道来的用户是否更容易触发转化？

## 给 SimpleTrack 的启发

SimpleTrack 应在 docs/quickstart 里把 `identify` 解释成“可选的用户身份增强”，不要让用户误以为必须登录后才能采集事件。P1 可以先支持事件采集，identify 作为推荐增强能力。

## 给 analytics-core 的启发

`analytics-core` 的 `EventEnvelope` 已有 `DistinctID` 和 `UserProps`，可以承接 identify 语义。实现时应把用户属性和事件属性区分开：事件属性描述“这次行为”，用户属性描述“这个人”。

