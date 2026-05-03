# 问：LCP、INP、CLS、FCP、TTFB 是什么，如何采集？

## 答

这五个指标是网页性能和用户体验指标，Umami 把它们作为 performance event 上报。

| 指标 | 全称 | 通俗解释 |
| --- | --- | --- |
| LCP | Largest Contentful Paint | 页面最大主要内容多久显示出来 |
| INP | Interaction to Next Paint | 用户点击、输入后，页面多久有响应 |
| CLS | Cumulative Layout Shift | 页面加载过程中布局是否乱跳 |
| FCP | First Contentful Paint | 页面第一个文字或图片多久出现 |
| TTFB | Time to First Byte | 浏览器多久收到服务器返回的第一个字节 |

## Umami 如何采集

Umami tracker 在浏览器里读取 performance entries 和相关 Web Vitals 数据，组成 payload 后发送：

```text
type = performance
payload = { lcp, inp, cls, fcp, ttfb, url, title, website, ... }
```

服务端在 `references/umami/src/app/api/send/route.ts` 的 `performance` 分支里把它写成 `EVENT_TYPE.performance`，也就是 `event_type = 5`。

## 源码位置

| 位置 | 作用 |
| --- | --- |
| `references/umami/src/tracker/index.js` | 采集并发送 performance metrics |
| `references/umami/src/lib/constants.ts` | 定义 `WEB_VITALS_THRESHOLDS` 和 `EVENT_TYPE.performance` |
| `references/umami/src/app/api/send/route.ts` | 接收 `lcp/inp/cls/fcp/ttfb` 并写入 performance event |
| `references/umami/prisma/schema.prisma` | `WebsiteEvent` 里有对应字段 |
| `references/umami/db/clickhouse/schema.sql` | `website_event` 里有对应字段 |

## 给 SimpleTrack 的启发

Performance 对用户价值很高，但不是 SimpleTrack P1 必须项。P1 先保证 pageview、自定义事件、Realtime、Events；性能指标可以在 P2 作为诊断增强能力引入。

## 给 analytics-core 的启发

`analytics-core` 可以把 performance 作为事件类型或事件属性承接。表模型应预留扩展能力，但 P1 不必因为 performance 牺牲 collect 主链路稳定性。

