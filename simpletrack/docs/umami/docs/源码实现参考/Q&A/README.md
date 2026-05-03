# 源码实现参考 Q&A

> 状态：持续补充  
> 用途：把 Umami 源码解读里的概念问题拆成一问一答，供 SimpleTrack 产品评审和 `analytics-core` 实现对照时快速引用。

## 索引

| 问题 | 回答 |
| --- | --- |
| identify 在 Umami 是什么概念？ | [01-identify是什么](./01-identify是什么.md) |
| Umami 的事件存储是 PostgreSQL 和 ClickHouse 二选一吗？ | [02-Umami事件存储是PostgreSQL和ClickHouse二选一吗](./02-Umami事件存储是PostgreSQL和ClickHouse二选一吗.md) |
| Prisma schema 是什么？ | [03-Prisma-schema是什么](./03-Prisma-schema是什么.md) |
| 字段白名单是什么意思？ | [04-字段白名单是什么](./04-字段白名单是什么.md) |
| LCP、INP、CLS、FCP、TTFB 是什么，如何采集？ | [05-LCP-INP-CLS-FCP-TTFB是什么](./05-LCP-INP-CLS-FCP-TTFB是什么.md) |
| 字段白名单、过滤参数、Realtime 短窗口和 Events 分页模型怎么理解？ | [06-过滤参数-Realtime短窗口-Events分页是什么](./06-过滤参数-Realtime短窗口-Events分页是什么.md) |
| 自动采集和 SDK 能否借鉴到 SimpleTrack？ | [07-自动采集和SDK如何借鉴到SimpleTrack](./07-自动采集和SDK如何借鉴到SimpleTrack.md) |
| bot/IP 过滤和 Zod 是什么？ | [08-bot-IP过滤和Zod是什么](./08-bot-IP过滤和Zod是什么.md) |
| runQuery 和 storage dispatch 是什么？ | [09-runQuery和storage-dispatch是什么](./09-runQuery和storage-dispatch是什么.md) |
| session/visit 是否就是隐私友好的用户识别？ | [10-session隐私机制是什么](./10-session隐私机制是什么.md) |
| 只存 JSON 属性的风险如何解决？ | [11-只存JSON属性的风险如何解决](./11-只存JSON属性的风险如何解决.md) |

