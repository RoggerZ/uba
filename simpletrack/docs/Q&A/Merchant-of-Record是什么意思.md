# Merchant of Record 是什么意思

## Q：Merchant of Record 是什么？

A：Merchant of Record，简称 MoR，直译是“登记商户”或“交易责任商户”。在软件出海收费场景里，它指的是由一个支付平台作为面向客户的卖方，负责收款、处理订单、税务、发票、退款、拒付等交易责任，然后再把扣除费用后的收入结算给开发者或公司。

## Q：为什么 MoR 能减少税务和资质负担？

A：因为跨境销售数字产品时，卖方通常要面对不同国家或地区的销售税、VAT/GST、发票、退款、拒付、支付合规等问题。如果你自己用普通支付网关收款，很多责任会落到你自己的主体上。MoR 平台通常会作为交易卖方处理这些事务，开发者主要面对平台结算和平台审核。

## Q：MoR 是不是等于完全没有合规问题？

A：不是。MoR 能减少交易侧和税务侧负担，但不等于所有法律、隐私、公司、收入申报问题都消失。你仍然需要满足平台审核、当地收入申报、产品合规、隐私政策和服务条款要求。

## Q：Lemon Squeezy、Paddle、Polar 都是 MoR 吗？

A：Lemon Squeezy 和 Paddle 明确以 Merchant of Record 方式服务数字产品和 SaaS。Polar 也主打为开发者处理全球销售税、发票、订阅和退款等商业化流程。具体责任边界要以各平台最新官方条款为准。

## Q：Stripe 是 MoR 吗？

A：一般不是。Stripe 更常见的定位是支付基础设施和支付处理平台。它能力很强，但通常需要商户自己承担更多税务、开票、合规和销售责任。Stripe 也有税务相关产品，但它和 MoR 模式不是一回事。

## Q：为什么我看到很多产品都使用 Stripe 支付？

A：因为 Stripe 的开发者体验、API、文档、Checkout、Billing、Customer Portal、webhook、订阅、企业支付和生态集成都非常成熟。很多欧美公司本身已经有公司主体、税务和财务流程，Stripe 对它们来说灵活、稳定、可控，适合长期做复杂计费。

对模板作者来说，Stripe 也很容易做成通用示例，因为它文档清晰、SDK 完整、用户认知度高。所以很多 SaaS 模板默认先支持 Stripe。

但这不代表 Stripe 一定最适合个人开发者早期出海。个人开发者通常更需要降低税务、发票、拒付和主体资质负担，所以 Lemon Squeezy、Paddle、Polar 这类 MoR 路线仍然值得优先评估。

## Q：SimpleTrack 为什么倾向 MoR 路线？

A：SimpleTrack 早期更需要验证产品和收费意愿，而不是先投入大量精力处理跨境税务、发票、拒付、各国税率和支付资质。MoR 路线可以把这些复杂度前置交给平台，让团队更快验证市场。

## Q：当前支付路线要马上把这些问题都确认完吗？

A：不用。当前工程决策是先按 Supastarter 已支持的 provider 接入，也就是 Stripe、Lemon Squeezy、Polar、Creem、Dodo Payments 这些路线先保留能力。

KYC/KYB、退款、拒付、发票、税务处理和费用结构，放到产品准备真实上线收费前逐项确认。现在先保证模板支付抽象能接入 subscription gate，不把合规细节变成 P0/P1 的早期阻塞。

## 参考

- Lemon Squeezy Merchant of Record: https://docs.lemonsqueezy.com/help/payments/merchant-of-record
- Paddle Merchant of Record: https://www.paddle.com/merchant-of-record
- Polar Docs: https://polar.sh/docs
- Stripe Docs: https://docs.stripe.com/
