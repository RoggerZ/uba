# Wasp 是什么

## Q：Wasp 是什么？

A：Wasp 是一个全栈 Web 应用框架。它让开发者用一个配置文件描述 app、route、page、query、action、auth、job 等能力，然后生成和组织 React、Node.js、Prisma 等代码。Open SaaS 就是基于 Wasp 做的 SaaS 模板。

## Q：它和 Next.js 是同一类东西吗？

A：不完全是。Next.js 是 React 框架，核心围绕页面、路由、渲染、Server Components、Route Handlers 和部署生态。Wasp 更像一个全栈应用编排层，它把前端、后端、数据库、任务等用自己的 DSL 连接起来。

## Q：Wasp 的好处是什么？

A：它能把全栈 SaaS 常见的结构约定好，减少样板代码。对 Open SaaS 这类模板来说，Wasp 能快速把 auth、payments、email、jobs、Prisma、React 页面组织成一套可运行系统。

## Q：为什么不建议 SimpleTrack 直接走 Wasp？

A：不是因为 Wasp 不好，而是因为 SimpleTrack 当前更适合 Next.js 主线。候选付费模板、前端生态、可迁移原型、未来招聘和 AI 编码上下文都更集中在 Next.js。采用 Wasp 会引入额外框架心智和部署方式，生产路线会偏离我们当前的主路径。

## Q：Open SaaS 还有参考价值吗？

A：有。Open SaaS 可以作为免费对照，帮我们确认一个完整 SaaS 控制面应该覆盖哪些能力，例如登录、支付、邮件、后台任务、文件上传、AI 示例和 E2E 测试。但它不一定作为 SimpleTrack 的最终生产底座。

## 参考

- Wasp Docs: https://wasp.sh/docs
- Open SaaS: https://github.com/wasp-lang/open-saas/

