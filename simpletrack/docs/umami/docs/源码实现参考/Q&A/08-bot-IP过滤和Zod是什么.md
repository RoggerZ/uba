# 问：bot/IP 过滤和 Zod 是什么？

## 答

这两个概念都出现在 Umami 的 collect API 入口里，但职责不同。

## bot/IP 过滤是什么

bot 过滤是识别搜索引擎爬虫、自动化脚本或明显非真实用户的访问，避免这些访问污染页面浏览和转化数据。

IP 过滤是屏蔽指定 IP 或 IP 段，比如公司内网、测试机器、恶意来源。

Umami 在 `references/umami/src/app/api/send/route.ts` 中会：

- 用 `isbot(userAgent)` 判断 user agent 是否像 bot。
- 用 `hasBlockedIp(ip)` 判断 IP 是否被屏蔽。

## Zod 是什么

Zod 是 TypeScript/JavaScript 生态里的数据校验库。它用来声明“请求应该长什么样”，然后验证真实请求是否符合。

在 Umami collect API 里，Zod 负责检查：

- `type` 必须是 `event`、`identify` 或 `performance`。
- `payload.website/link/pixel` 必须三选一。
- `url`、`title`、`event name` 等字段不能超出规则。

## 给 SimpleTrack 的启发

SimpleTrack 产品需要给用户提供测试流量过滤或内部流量过滤能力，否则团队自己访问产品会污染数据。bot/IP 过滤可以先作为后台规则或配置项，而不是 P1 UI 大功能。

## 给 analytics-core 的启发

Go 里不会用 Zod，但需要等价的请求校验层。`collect.Normalize` 和 `collect.Handler` 应承担 schema validation；bot/IP 过滤可以作为 collect 前置或 enrichment stage，但不要混进 ClickHouse writer。

