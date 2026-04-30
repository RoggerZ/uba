# Stripe、Polar、Lemon Squeezy 对比

> 状态：后置评审，上线前逐项处理  
> 用途：解释这些服务是不是支付，以及它们对 SimpleTrack 的意义。

## Q：Stripe、Polar、Lemon Squeezy 都是支付吗？

A：都和收款有关，但定位不完全一样。

- Stripe：支付基础设施和支付处理平台，能力很强，适合自建更可控的 billing 系统。
- Lemon Squeezy：面向数字产品和 SaaS 的 Merchant of Record 平台，处理销售税、订阅、发票、退款等事务。
- Polar：面向开发者和开源/数字产品商业化的平台，提供订阅、一次性销售、税务、发票、退款等能力，并强调 Merchant of Record。

## 核心差异

| 服务 | 更像什么 | MoR | 适合什么场景 | SimpleTrack 判断 |
| --- | --- | --- | --- | --- |
| Stripe | 支付网关 + billing 基础设施 | 通常不是 | 有主体、有合规能力、想高度自定义支付和订阅 | 技术强，但早期负担较重 |
| Lemon Squeezy | 数字产品/SaaS MoR | 是 | 独立工具、软件订阅、希望平台处理税务和交易责任 | 第一优先支付候选 |
| Polar | 开发者商业化 MoR 平台 | 是 | 开源、API、SaaS、开发者工具订阅 | 值得评估，尤其适合开发者产品 |

## Q：Stripe 的优势是什么？

A：Stripe 的优势是生态成熟、API 强、文档完整、订阅和企业支付能力强。缺点是它通常要求我们自己承担更多商户责任，例如税务、发票、拒付处理和合规流程。

## Q：Lemon Squeezy 的优势是什么？

A：Lemon Squeezy 的优势是 Merchant of Record 模式，适合独立工具出海。它能把销售税、订阅、退款、拒付、发票等交易事务集中到平台侧处理，早期工程和合规负担更轻。

## Q：Polar 的优势是什么？

A：Polar 更偏开发者和开源商业化场景，提供订阅、一次性购买、license key、benefits、webhooks 等能力。它对开发者工具类 SaaS 有吸引力，但需要评估模板支持度和生态成熟度。

## Q：SimpleTrack 该怎么选？

A：当前工程决策是先按 Supastarter 已支持的 provider 接入，不急着在 P0/P1 把最终支付平台拍死。

Supastarter 本地 provider 已包含 Stripe、Lemon Squeezy、Polar、Creem、Dodo Payments。P1 先验证 subscription gate、checkout/webhook 抽象和账单入口能跑通；真实上线收费前，再逐项确认 KYC/KYB、发票税务、退款拒付、费用结构和目标地区支持。

## Q：还要不要看 Paddle？

A：可以后续看。Paddle 也是成熟 Merchant of Record 方案，适合 SaaS 和软件产品。但当前先顺着 Supastarter 已有 provider 做，不把 Paddle 作为 P1 必接项。

## 参考

- Stripe Docs: https://docs.stripe.com/
- Lemon Squeezy Merchant of Record: https://docs.lemonsqueezy.com/help/payments/merchant-of-record
- Polar Docs: https://polar.sh/docs
- Paddle Merchant of Record: https://www.paddle.com/merchant-of-record
