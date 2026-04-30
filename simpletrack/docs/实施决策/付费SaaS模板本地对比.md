# 付费 SaaS 模板本地对比

> 状态：已确定先选 Supastarter，持续更新  
> 最近更新：2026-04-29  
> 评审范围：`template-src/` 已有本地源码的 Supastarter、ShipAny、MkSaaS、ShipFast、Mkdirs，以及当前锁定对照项 MakerKit。

## 当前结论

SimpleTrack 的生产控制面先选择 **Supastarter for Next.js**。

MakerKit Next.js 保留为 B2B/企业治理强对照和备选；只有后续拿到更强证据，证明它在组织/RBAC、Super Admin、测试、文档、企业控制台 UI 和支付路线上的综合适配度明显超过 Supastarter，才重开选型。

本地 `template-src` 里已经有 Supastarter、ShipAny、MkSaaS、ShipFast、Mkdirs 源码。MakerKit 当前没有本地源码，因此 MakerKit 只能先基于官方资料和购买前问询评审，不能和 Supastarter 一样做源码级确认。

## 本地源码证据

| 模板 | 本地路径 | 关键证据 | 对 SimpleTrack 的判断 |
| --- | --- | --- | --- |
| Supastarter for Next.js | `template-src/ai-supastarter-template` | `apps/saas`、`apps/marketing`、`apps/docs`、`apps/mail-preview`；`packages/auth`、`api`、`database`、`payments`、`mail`、`storage`、`ui`、`ai`；SaaS 模块含 admin、auth、onboarding、organizations、payments、settings；payments provider 含 Stripe、Lemon Squeezy、Polar、Creem、Dodo Payments。 | 已确定先选。它同时覆盖产品官网、docs、SaaS 控制台、邮件预览、多支付、组织、Admin 和 AI，最贴合当前路线。 |
| MkSaaS | `template-src/ai-mksaas-template` | Next 16、React 19、Better Auth、Drizzle、Stripe、Resend、React Email、Fumadocs；有 protected admin/dashboard/payment/settings；`src/components/data-table/` 组件比较完整。 | 可作为 Events table、docs 和后台组件参考；但支付偏 Stripe，组织/多租户/企业治理证据不如 Supastarter 明确。 |
| ShipAny template two | `template-src/ai-shipany-template-two` | Next 16、Better Auth、RBAC 初始化脚本、Stripe、Creem、PayPal、Fumadocs。 | 偏 AI SaaS 快速启动，可参考 RBAC、支付和 docs 组织方式；不优先作为企业分析控制台底座。 |
| ShipAny template one | `template-src/ai-shipany-template-one` | Next 15、Stripe、Creem、Fumadocs。 | 作为 ShipAny 另一版本参考，不进入第一梯队。 |
| ShipFast | `template-src/ai-ship-fast-template` | Next 14、NextAuth、Stripe、Mongo/Mongoose、Mailgun/Crisp。 | 适合快速商业化和营销页，企业分析控制台、组织治理和现代 App Router 证据较弱。 |
| Mkdirs | `template-src/ai-mkdirs-template` | Next 14、Stripe，整体更像内容/目录站模板。 | 可参考内容站或目录站结构，不作为 SimpleTrack SaaS 控制面底座。 |

## Supastarter 为什么当前领先

Supastarter 的优势不是单点功能，而是结构更完整：

- **应用分离清楚**：marketing、saas、docs、mail-preview 分开，正好对应 SimpleTrack 的产品官网、控制台、文档和邮件模板预览。
- **控制面覆盖广**：本地源码能看到 auth、organizations、payments、admin、settings、onboarding。
- **多支付路线更适合出海**：本地 provider 覆盖 Stripe、Lemon Squeezy、Polar、Creem、Dodo Payments，不被单一 Stripe 路线锁死。
- **P1 可落地性强**：可以把 SimpleTrack 的 Realtime、Events、Website settings、Goal 页面迁入 `apps/saas`，公开站点迁入 `apps/marketing` 和 `apps/docs`。

## MakerKit 可能强在哪里

MakerKit 当前作为 B2B 对照，不是因为它一定比 Supastarter 更适合，而是因为它在企业治理叙事上通常更强，需要重点核验：

- **更强组织/账户模型**：如果它的 personal account、team account、邀请、成员管理和数据隔离更成熟，会利于后续企业客户。
- **更细 RBAC/权限体系**：如果支持自定义角色、权限字符串、权限检查和数据库级隔离，会比简单 owner/admin/member 更适合 B2B。
- **Super Admin 和支持操作**：如果有更完整的用户管理、组织管理、封禁、审计、impersonation 或支持入口，企业售后会更顺。
- **测试和文档体系**：如果 Playwright、seed、开发文档、升级指南、AI agent rules 更完整，会降低长期维护成本。
- **团队计费能力**：如果按席位、按用量、订阅变更和团队账单更成熟，会支撑 P3 企业套餐。

但 MakerKit 要替代 Supastarter，需要满足一个更高门槛：**必须在 B2B 企业控制面上明显领先，并且支付路线、许可证和 UI 可改造性都满足 SimpleTrack。** 如果只是“也能做”，不值得放弃已经有本地源码证据的 Supastarter。

## 取舍规则

| 判断项 | Supastarter 更有利时 | MakerKit 更有利时 |
| --- | --- | --- |
| 本地可验证性 | 已有本地源码，可以直接 spike | 需要购买或获取源码后才能验证 |
| 支付路线 | 需要 Lemon Squeezy / Polar / Creem / Dodo 等多支付抽象 | MakerKit 对 Lemon Squeezy / Paddle 支持确认更完整 |
| 企业治理 | Supastarter 的 organizations/admin 足够 P1-P2 | MakerKit 的 RBAC、Super Admin、审计、团队计费明显更强 |
| UI 改造 | Supastarter dashboard shell 能快速改成分析控制台 | MakerKit 控制台组件、表格、设置页更适合分析产品 |
| P1 速度 | 想最快进入产品官网、docs、控制台、支付和数据面开发 | 愿意为更强 B2B 治理接受更高切换成本 |

## 当前执行建议

1. P0/P1 先以 Supastarter for Next.js 为生产商业控制面底座。
2. MakerKit 只保留为付费 B2B 强对照和备选。
3. ShipAny、MkSaaS、ShipFast、Mkdirs 作为参考模板，不进入生产底座第一梯队。
4. 下一步先在 Supastarter 本地源码里做 1 天 spike：新增 Website 资源、Realtime/Events 页面草案、subscription gate 和一封邮件模板。
5. 支付先按 Supastarter 已有 Stripe、Lemon Squeezy、Polar、Creem、Dodo Payments provider 接入；KYC/KYB、发票税务、退款拒付和费用结构放到上线前逐项检查。

## 参考

- 本地 Supastarter：`template-src/ai-supastarter-template`
- 本地 MkSaaS：`template-src/ai-mksaas-template`
- 本地 ShipAny：`template-src/ai-shipany-template-one`、`template-src/ai-shipany-template-two`
- 本地 ShipFast：`template-src/ai-ship-fast-template`
- 本地 Mkdirs：`template-src/ai-mkdirs-template`
- Supastarter 官方文档：https://supastarter.dev/docs/nextjs
- MakerKit Next.js SaaS Boilerplate：https://makerkit.dev/nextjs-saas-boilerplate
