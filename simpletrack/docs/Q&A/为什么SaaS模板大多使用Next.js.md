# 为什么 SaaS 模板大多使用 Next.js

## Q：为什么我看到很多免费或收费 SaaS 技术栈都使用 Next.js？

A：因为 Next.js 同时解决了 SaaS 产品早期最常见的几类问题：营销页、登录页、定价页、控制台、API routes、SEO、部署、React 生态、TypeScript、组件库集成和全栈能力。对模板作者来说，用 Next.js 可以覆盖“从官网到应用后台”的完整产品壳。

## Q：Next.js 对独立工具和 SaaS 模板有什么优势？

A：主要优势是：

- 一个项目里能做 landing、pricing、docs、dashboard。
- React 生态成熟，组件库和模板多。
- App Router 适合组织复杂页面和布局。
- Route Handlers 可以承接轻量后端接口和 webhook。
- Vercel 部署链路顺滑，但也可以自部署。
- shadcn/ui、Tailwind、Radix、React Email、Vercel AI SDK 等生态组合常见。
- AI 编码工具对 Next.js 项目结构更熟悉。

## Q：为什么不是 Vue 或纯 React SPA？

A：Vue 和 React SPA 都可以做控制台，但 SaaS 模板通常还需要 SEO、营销页、服务端渲染、订阅 webhook、auth callback、邮件预览、边缘部署等能力。Next.js 把这些能力打包在一个主流框架里，所以模板生态更集中。

## Q：Next.js 是不是一定适合所有模块？

A：不是。Next.js 适合商业控制面和前端产品层，但 SimpleTrack 的事件采集、实时写入、ClickHouse 查询、队列、分析聚合，不应该强行塞进 Next.js。那些属于分析数据面，应该由 Go/Node 服务、ClickHouse、Redis、Kafka 或其他专门组件承接。

## Q：这对 SimpleTrack 的决策意味着什么？

A：SimpleTrack 可以用 Next.js 做官网、注册、控制台、设置、订阅入口和产品页面，但核心数据管道仍然单独设计。这样既利用 Next.js 模板生态，又不让前端框架绑架分析引擎。

## 参考

- Next.js Docs: https://nextjs.org/docs
- shadcn/ui Docs: https://ui.shadcn.com/
- Tailwind CSS Docs: https://tailwindcss.com/docs

