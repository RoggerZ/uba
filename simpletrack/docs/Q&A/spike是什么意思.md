# spike 是什么意思

## Q：spike 是什么意思？

A：spike 是短周期技术验证。它的目标不是交付完整功能，而是在很短时间内验证一条技术路线是否可行、风险在哪里、成本是否可接受。

## Q：spike 和正式开发有什么区别？

A：正式开发要做完整体验、测试、错误处理、可维护代码和交付质量。spike 可以更粗糙，只验证关键假设。例如“Supastarter 能否新增 SimpleTrack 的 Website 模型并跑通构建”或“xwl_bi 核心能否用 Redis Stream 写入事件”。

## Q：SimpleTrack 现在需要哪些 spike？

A：当前至少需要两个：

1. SaaS 模板 spike：验证 Supastarter for Next.js 是否能承接 SimpleTrack 控制面和企业分析控制台 UI。
2. xwl_bi 核心抽取 spike：验证采集、事件流、ClickHouse 查询、Redis Stream 替代 Kafka 是否能组成业务无关核心。

