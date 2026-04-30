# 为什么说 Umami 无 cookie、无指纹？

> 问题来源：[04-Sessions](../04-Sessions.md) 第 42 行提到“因为 Umami 的默认定位是无 cookie、无指纹”。这里解释这句话的准确含义、实现方式和 SimpleTrack 落地边界。

## 结论

Umami 的“无 cookie、无指纹”不是说它完全不识别会话，也不是说它不读取任何浏览器或请求信息。

更准确的说法是：Umami 默认不在访客浏览器里写入用于跟踪的 cookie，也不把一组高维浏览器特征拼成长期稳定的设备指纹；它用服务端可见的基础请求信息和站点信息生成匿名、带时间边界的 session / visit 标识，用来做去重、会话串联和统计汇总。

所以 Sessions 里的 `visitor` 更像“匿名访问上下文”，不是 CRM 里的实名客户。只有业务方主动传入 Distinct ID 或 session data 时，它才会升级成可按业务身份检索的访问记录。

## “无 cookie”指什么

这里的 cookie 指“在被统计网站的访客浏览器上放一个持久 visitor id，再靠这个 id 跨页面、跨访问持续识别同一个人”的跟踪 cookie。

Umami 默认不靠这种 cookie 工作：

- 官方介绍页把产品定位写成：`No cookies, no fingerprinting, no personal data.`
- Sessions 文档说明 session 来自 IP、User-Agent、website ID 等信息生成的 unique hash，因此访客是匿名跟踪，且不需要 cookie。
- 登录用户识别 guide 也强调 session data 存在服务端，默认匿名访客不使用 cookie。

需要注意两层边界：

- 这里说的是统计 tracker 对被统计访客的默认采集方式，不是 Umami 管理后台登录态一定不使用 cookie。
- “不用 cookie”不等于“完全没有短期运行态”。当前源码里 tracker 会把服务端返回的 cache token 放在脚本内存变量里，并在后续请求头 `x-umami-cache` 里带回，用来减少重复建 session / visit；它不是写入浏览器 cookie 的长期访客 ID。

## “无指纹”指什么

这里的“指纹”主要指浏览器指纹识别：收集大量设备、浏览器、字体、Canvas、插件、硬件、时区等特征，拼成一个尽量长期稳定、可跨会话甚至跨站识别的设备画像。

Umami 的做法更克制：

- 它会记录或解析基础分析字段，例如 URL、referrer、title、language、screen、browser、OS、device、country 等。
- 它会用 IP 和 User-Agent 参与匿名 session hash 的生成。
- 它不会把原始 IP 存进 session 表；官方 metric definitions 明确说 IP 会用于位置识别，但不会存储。
- 它使用 rotating salt，让 hash 结果有时间边界。官方 metric definitions 说明 visitor/session 相关 hash 默认按月轮换，visit 相关 hash 按小时轮换。

所以“无指纹”不是“完全不用 IP/User-Agent”。它的意思是：不把这些信息组合成长期稳定、可反查或可跨站跟踪的浏览器指纹，而是把它们压成匿名、轮换、站点内使用的统计标识。

## 它是怎么做到的

按当前官方文档和 2026-04-28 复核的 Umami GitHub `main` 快照，可以理解成这条链路：

1. 浏览器端 tracker 默认发送页面上下文字段：`website`、`hostname`、`url`、`referrer`、`title`、`screen`、`language` 等。
2. 服务端从请求中读取 IP、User-Agent，并解析出 browser、OS、device、location 等统计字段。
3. 如果没有业务方传入的 Distinct ID，服务端用 `website/source id + IP + User-Agent + session salt` 生成 `sessionId`。
4. `session salt` 默认按月轮换；`visitId` 则由 `sessionId + hourly salt` 生成，并且普通实时请求里 30 分钟后会过期成新 visit。
5. 数据库存的是生成后的 `session_id`、`visit_id` 和分析维度字段，例如 browser、OS、device、screen、language、country、region、city、distinct_id；session 表本身没有原始 IP 字段。
6. 如果业务方调用 `umami.identify()` 传入 `id`，Umami 会把 session 与这个业务 ID 关联。这个能力是显式增强身份识别，不属于默认匿名模式。

这套设计的核心不是“完全无法关联任何访问”，而是“关联只发生在站点内、统计目的内、匿名 hash 内，并且默认不依赖访客端持久 cookie 或长期设备指纹”。

## 和 Sessions 第 42 行的关系

这就是为什么 [04-Sessions](../04-Sessions.md) 里说 Sessions 的“访客”更适合理解为匿名访问上下文。

同一个 session 里，Umami 可以告诉你：

- 这个访问上下文来自哪个国家、浏览器、设备。
- 它访问过哪些页面，触发过哪些事件。
- 它有哪些 session properties。

但默认情况下，它不能告诉你：

- 这个人真实姓名是谁。
- 这个匿名访客是否一定等于某个 CRM 客户。
- 这个人换设备、换浏览器或跨较长时间窗口后是否仍是同一个自然人。

如果 SimpleTrack 需要“业务用户”视角，就应该显式接入 Distinct ID，而且优先使用内部用户 ID、哈希 ID 或不可反查的业务标识，不要直接上传邮箱、手机号、姓名、token、cookie 或原始订单明细。

## 对 SimpleTrack 的落地规则

1. UI 文案里区分 `visitor`、`session`、`distinct user`、`customer`，不要把匿名访客直接叫成客户。
2. 默认接入只做匿名 session 与事件分析，不要求首日就打通登录身份。
3. 需要跨会话或跨设备识别时，再使用 `umami.identify()` 或等价的 Distinct ID 机制。
4. Distinct ID 优先使用内部 ID 或不可反查 hash，不直接用邮箱、手机号、姓名。
5. `before-send` / `data-exclude-search` / `data-exclude-hash` 可以作为隐私闸门，过滤 query、hash、token、表单内容等敏感 payload。
6. 评审 Sessions 报表时，把它标成“匿名访问上下文证据”，不要写成“实名用户画像证据”。

## 常见误解

| 误解 | 更准确的理解 |
| --- | --- |
| 无 cookie = 完全不识别访客 | 它仍会生成匿名 session / visit id，用于去重和会话串联 |
| 无指纹 = 完全不用 IP / User-Agent | 它会用这些基础请求信息生成匿名 hash，但不做长期稳定设备画像 |
| Sessions 里的 visitor = 注册用户 | 默认只是匿名访问上下文；注册用户需要 Distinct ID |
| 匿名 hash = 一定没有合规风险 | 仍应避免上传 PII，并根据实际部署地区做合规评估 |
| 接入 identify 后仍是纯匿名 | identify 是业务主动身份关联，隐私边界已经变成“可控业务标识” |

## 来源与复核

- [Umami Introduction](https://docs.umami.is/docs)：官方 privacy-first 定位。
- [Umami Sessions](https://docs.umami.is/docs/sessions)：session hash 与匿名 cookie-free tracking 说明。
- [Umami Metric definitions](https://docs.umami.is/docs/metric-definitions)：session、visit、IP 不存储与 rotating salt 说明。
- [Umami Distinct IDs](https://docs.umami.is/docs/distinct-ids)：Distinct ID 跨 session 关联能力。
- [Identify logged-in users](https://docs.umami.is/docs/guides/identify-logged-in-users)：登录用户识别与隐私注意事项。
- [Tracker functions](https://docs.umami.is/docs/tracker-functions)：`umami.identify()` 和 session data 用法。
- [Tracker configuration](https://docs.umami.is/docs/tracker-configuration)：`data-exclude-search`、`data-exclude-hash`、`data-do-not-track`、`data-before-send` 等隐私相关配置。
- [Umami source: send route](https://github.com/umami-software/umami/blob/c78ff36db0c82e13c86e5073020472c6546313a3/src/app/api/send/route.ts#L143-L148)：`sessionId` / `visitId` 的 salt 与 hash 生成路径。
- [Umami source: crypto helpers](https://github.com/umami-software/umami/blob/c78ff36db0c82e13c86e5073020472c6546313a3/src/lib/crypto.ts#L60-L78)：`uuid()` 与 `getSalt()` 实现。
- [Umami source: session schema](https://github.com/umami-software/umami/blob/c78ff36db0c82e13c86e5073020472c6546313a3/prisma/schema.prisma#L34-L64)：session 表字段。
- [Umami source: tracker cache](https://github.com/umami-software/umami/blob/c78ff36db0c82e13c86e5073020472c6546313a3/src/tracker/index.js#L175-L190)：cache token 通过请求头回传，不是 cookie 写入。
