# 付费 SaaS 模板的付费模式

## Q：这些付费 SaaS 模板通常怎么收费？

A：常见模式是一次性购买源码访问权，按个人、团队、项目数或产品线分档。购买后通常可以把模板代码用于自己的商业项目，但许可证会限制转售模板本身、公开源码、多人共享账号或超出授权项目使用。

## Q：它们是按月订阅吗？

A：有些是一次性付费，有些会提供年度更新、会员、团队版或企业版。具体要看每个模板的许可证和 pricing 页面。不能只看“买一次”，还要看是否包含长期更新和支持。

## Q：买的是服务还是代码？

A：多数 SaaS boilerplate 买的是代码、文档、更新和支持。它不是托管服务，不会替我们运营 SimpleTrack。买回来后仍然要自己部署、改业务模型、接数据库、接支付、做安全和测试。

## Q：付费后能不能商用？

A：通常可以商用，但必须看许可证。关键要确认：

- 是否允许用于 SimpleTrack 这种商业 SaaS。
- 是否限制项目数量。
- 是否限制开发者席位。
- 是否允许客户私有化部署。
- 是否允许修改后闭源。
- 是否允许把模板代码放进公司的私有仓库。

这些问题的详细解释见：[一次性购买模板为什么还要确认许可证.md](一次性购买模板为什么还要确认许可证.md)。

## Q：为什么买模板还要做 spike？

A：因为模板能省时间，但也可能引入复杂度。1 天 spike 能确认真实二次开发手感：新增一个 SimpleTrack 的 Website 资源、受保护页面、mock subscription gate、邮件模板和 Realtime 页面，是否能保持构建和测试通过。

## Q：SimpleTrack 现在最该问什么？

A：当前已经先选择 Supastarter for Next.js，所以最该问的是“能不能顺畅接入和长期合规使用”，而不是继续比较模板。优先确认：

1. 许可证是否覆盖我们的商业使用。
2. 是否允许私有仓库、闭源修改、团队协作和未来可能的私有化交付。
3. Organizations、RBAC、Admin、Email、AI 是否是完整能力，不只是 demo。
4. UI 和代码结构是否能改成企业分析控制台。
5. 支付先按 Supastarter 已有 provider 跑通；KYC/KYB、发票税务、退款拒付和费用结构放到上线收费前逐项确认。

## 参考

- Supastarter Pricing: https://supastarter.dev/pricing
- MakerKit Pricing: https://makerkit.dev/pricing
- ShipFast: https://shipfa.st/
- ShipAny: https://shipany.ai/
- MkSaaS: https://mksaas.com/
