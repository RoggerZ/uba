# Supastarter 和 MakerKit 如何取舍

## Q：现在是不是已经选择 Supastarter for Next.js？

A：是。当前生产控制面先选择 **Supastarter for Next.js**。

MakerKit Next.js 保留为 B2B/企业治理强对照和备选，不再和 Supastarter 作为同等候选反复摇摆。

ShipAny、MkSaaS、ShipFast、Mkdirs 可以继续作为参考，但不作为第一梯队生产底座。

## Q：为什么当前更倾向 Supastarter？

A：因为本地 `template-src/ai-supastarter-template` 已经能看到比较完整的生产结构：

- `apps/marketing`：适合产品官网 / Marketing Site。
- `apps/saas`：适合 SimpleTrack 控制台。
- `apps/docs`：适合 docs/quickstart。
- `apps/mail-preview`：适合邮件模板预览。
- `packages/payments`：本地 provider 覆盖 Stripe、Lemon Squeezy、Polar、Creem、Dodo Payments。
- `apps/saas/modules`：包含 admin、auth、onboarding、organizations、payments、settings 等模块。

这和 SimpleTrack 需要的“公开站点 + SaaS 控制面 + 分析数据面”组合更贴。

## Q：MakerKit 在“更强 B2B 企业控制面”上可能强在哪里？

A：重点看这些能力是否比 Supastarter 明显更成熟：

- **组织和账户模型**：personal account、team account、邀请、成员管理、组织级数据隔离是否更完整。
- **RBAC 和权限体系**：是否支持自定义角色、权限字符串、权限检查和更细粒度的权限边界。
- **Super Admin / 支持操作**：是否有用户管理、组织管理、封禁、审计、impersonation、支持入口等后台能力。
- **团队计费**：是否对按席位、按用量、订阅变更、团队账单更成熟。
- **测试和文档**：是否有更完整 Playwright、seed、升级文档、AI agent rules 和维护指南。

如果这些能力只是“有”，不一定足以替代 Supastarter；必须明显领先，才值得切换。

## Q：MakerKit 如果可能更强，为什么现在不直接选它？

A：因为当前 SimpleTrack 的约束不是只追求最强 B2B，而是要兼顾：

- 本地能否立刻 spike。
- 是否明确支持 Lemon Squeezy / Paddle / MoR 方向。
- 是否容易改成企业分析控制台。
- 是否不会拖慢 P1 的 `analytics-core` 和数据管道建设。
- 是否能同时承接产品官网、docs、控制台和邮件。

Supastarter 现在有本地源码证据，切入成本更低。MakerKit 需要通过官方资料、购买前问询或拿到源码后验证，才有资格在未来作为替代方案重开评审。

## Q：未来什么情况下才重开 MakerKit 评审？

A：只有同时满足以下条件，才考虑从 Supastarter 改选 MakerKit：

1. MakerKit 的组织/RBAC/Super Admin/团队计费明显强于 Supastarter。
2. MakerKit 的支付路线满足当前商业计划，尤其 Lemon Squeezy、Paddle 或其他 MoR 方案。
3. MakerKit 的 UI 和组件结构更适合低装饰、高密度的企业分析控制台。
4. MakerKit 的许可证允许 SimpleTrack 商业 SaaS、私有仓库、闭源修改和团队协作。
5. 1 天 spike 能顺利新增 Website、Realtime、Events、subscription gate 和一封邮件模板。

## Q：这对 P1 有什么影响？

A：P1 不应该因为模板选择反复停住。当前执行策略是：

- 商业控制面：在 Supastarter 上验证。
- B2B 对照：MakerKit 保留资料级跟踪，不阻塞当前路线。
- 分析数据面：按已确定方案推进 `analytics-core`。
- 公开站点：优先复用模板的 marketing/docs app，做产品介绍、定价/订阅入口和 quickstart。
