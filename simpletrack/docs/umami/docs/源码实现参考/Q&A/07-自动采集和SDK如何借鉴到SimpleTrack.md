# 问：自动采集和 SDK 能否借鉴到 SimpleTrack，如何取舍？

## 答

可以借鉴，但要分阶段。

Umami tracker 的能力包括：

- 自动 pageview。
- SPA 路由变化自动 track。
- DOM attribute 事件。
- 手动 `track()`。
- 手动 `identify()`。
- 可选 performance。
- `beforeSend` 修改或取消 payload。
- DNT / domain allowlist 等隐私和安全配置。

## SimpleTrack P1 推荐取舍

| 能力 | P1 是否做 | 原因 |
| --- | --- | --- |
| 自动 pageview | 做 | 接入后马上能在 Realtime 看到数据，是首价值 |
| 手动 custom event | 做 | Events 和 Goal 最小闭环需要 |
| DOM attribute 事件 | 可做轻量版 | 降低无代码埋点成本，但不必一开始做太复杂 |
| identify | 可选增强 | 对用户属性有价值，但不能阻塞 pageview/custom event |
| SPA 路由监听 | 做 | Next.js / React 产品常见，否则单页应用 pageview 不准 |
| performance | 暂缓到 P2 | 有价值，但不是 P1 数据管道活了的必要条件 |
| beforeSend | 谨慎做 | 灵活但增加调试成本，服务端仍必须校验 |
| 多语言 SDK | 后续阶段 | P1 先 Web SDK；server/mobile SDK 放 P2/P3 |

## analysis-core 是否需要浏览器自动采集？

严格说，`analytics-core` 不负责浏览器自动采集；它负责接收标准事件、队列、写入和查询。

浏览器自动采集应属于 SimpleTrack Web SDK 或 tracker 包。这个 SDK 把浏览器数据转换成 `analytics-core` 能理解的 collect 请求。

## 给 SimpleTrack 的启发

SimpleTrack 需要提供 SDK，但 P1 优先 Web tracker。后续可以补：

- React / Next.js helper。
- Node.js server SDK。
- Mobile SDK。
- 后端事件上报 SDK。

## 给 analytics-core 的启发

`analytics-core` 不应依赖浏览器 SDK 的实现细节。它只定义稳定 collect 协议和 `EventEnvelope`，让 Web SDK、server SDK、mobile SDK 都能接入同一数据面。

