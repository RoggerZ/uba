# Litlyx 功能深度分析

> 目标：把 Litlyx 后台跑起来，并把 IA、页面态、关键交互、付费分层与可借鉴点整理成可直接用于 SimpleTrack 评审的证据链。

## 1. 调研边界

- 调研对象：`https://dashboard.litlyx.com/`
- 调研方式：真实账号登录后，结合自动化截图、真实事件回写和后台复核
- 本轮已完成：
  - 登录、工作区、安装入口
  - `Web / Product / Marketing / Reports / SEO / Shields / Analyst / Plans / Shareable links / Members`
  - 本地 `tracking-demo` 工件落盘并完成一轮真实验证
- 本轮未执行：
  - Reports 各类正式报告的 POST 生成动作，因为当前 Free trial 下正式报告卡片和 `Generate report` 都是 disabled

## 2. 顶层 IA：Workspaces 与工作区内壳分离

Litlyx 顶层不是一个单页 dashboard，而是两层结构：

1. `Workspaces`：账号级项目入口
2. 工作区内部侧栏：`Web / Product / Marketing / SEO / Reports / Members / Shields / Settings / Analyst`

这套结构的价值在于：

- 项目切换和分析导航不混在一起
- 协作、治理、AI、商业化都能在工作区壳内单独建层
- `Members` 灰态也更容易被理解为套餐限制，而不是导航异常

## 3. Onboarding：首屏先解决“怎么接入”

Litlyx 登录后的默认落点是 `Web` 安装页，而不是总览仪表盘。

这个页面同屏提供：

- `Script`
- `Tag (GTM)`
- 多个框架快捷入口
- `Workspace ID`
- `Use Cursor / Use Claude / Use ChatGPT`
- `Verify Installation`

这说明它把 onboarding 设计成一条很短的链：

1. 复制脚本或 GTM 代码
2. 如有需要，跳去对应框架文档
3. 用 AI prompt 接入现有项目
4. 回后台验证安装

对 SimpleTrack 来说，这比“先给你一个空 dashboard”更符合首次使用心智。

## 4. Settings：安装信息会被重复暴露

`Settings / General` 里再次出现：

- Workspace name
- Workspace ID
- Script
- Delete workspace

这和 `Web` 安装页形成重复暴露。好处是回查方便，风险是未来脚本形态变多时一致性维护成本会上来。

`Settings / Domains` 复核后不是简单空态，而是一个数据治理入口，分成两块：

- `Domain data`：选择特定 domain，删除对应 visits 和 events，用于清理测试数据或响应隐私请求
- `Sanitize Domains`：复核被追踪到的 origins，移除不属于项目的域名，文案明确提到第三方工具、预览环境、iframe、配置平台，以及 DNS hijacking / domain stealing

这类入口带有删除语义，本轮只做页面态记录，没有点击 `Start Domain Sanitization`。

## 5. Product：分析骨架、示例态、真实出数态都已经验证

### 5.1 默认空态不是空白，而是完整分析骨架

即使没数据，Product 也会先摆出：

- `Top Events`
- `Top 5 events`
- `Events`
- `Funnel Analysis`
- `Events User Flow`
- `Analyze event metadata`

这让用户能提前理解“未来这里会长成什么样”。

### 5.2 `Show test data` 很有价值

打开 `Show test data` 后，Product 能立刻进入可演示状态。它既能教育用户，也能降低“接入前全是空页”的挫败感。

### 5.3 `Setup events` 是外链文档入口，不是站内配置向导

我对 `Setup events` 做了额外复核：

- 按钮本身不是 disabled
- HTML 上也没有 `disabled` 或 `aria-disabled`
- 登录态实际加载的 Nuxt chunk 里，这个入口被渲染成 `to="https://docs.litlyx.com/custom-events"`、`target="_blank"` 的链接
- DOM 结构是一个外层 `<a href="https://docs.litlyx.com/custom-events" target="_blank">` 包住 `Setup events` 按钮
- 点击后会打开新标签页，标题为 `Custom Events - Litlyx Docs`
- 新增截图 `P03-S07` 已记录这个落点

更重要的是，这个结论不是因为“系统本身不支持事件”。Litlyx 官方文档已经明确给出两条可用路径：

- 浏览器脚本接入：`<script defer data-workspace="workspace_id" src="https://cdn.jsdelivr.net/npm/litlyx-js@latest/browser/litlyx.js"></script>`
- 自定义事件：`Lit.event('click-on-buy')`，并支持 `metadata`

而本轮本地 `tracking-demo` 也已经用这两条路径跑通，所以 `Setup events` 的真实产品意图可以确认：它不是站内事件配置向导，而是把用户送到官方 Custom Events 文档。需要警惕的是，这个按钮文案很像“继续在产品里配置事件”，但实际是外链文档，用户如果没有捕捉到新标签页，会误以为当前页没有响应。

### 5.4 真实验证后，Product 与 Raw Events 都已经出数

这轮我用本地 `tracking-demo` 加载 Litlyx 浏览器脚本，并触发两条自定义事件，再用 `bulk-send.mjs` 额外补了 12 条批量事件。

后台结果已经明确：

- `Product` 顶部出现 `Total events: 14`
- `Top Events` 中出现：
  - `demo_signup_click`
  - `demo_checkout_started`
  - `demo_report_requested`
  - `demo_upgrade_viewed`
- `Raw Events` 表格出现 14 条 `localhost` 事件记录

这说明下面这条链已经被真实跑通：

1. 本地静态页加载官方浏览器脚本
2. 触发自定义事件
3. 事件进入 Litlyx 后台
4. `Product` 聚合页和 `Raw Events` 明细页都能看到结果

## 6. Marketing：分析页与 UTM 生成器形成闭环

Marketing 模块同屏包含：

- 顶部 KPI
- `Traffic by channel`
- `Traffic by referrer`
- `Social Channels`
- `Show test data`
- `Learn about UTM`
- `Generate UTM link`

这里最值得借鉴的是：UTM 生成器不是藏在文档里，而是直接放在分析页上，形成“看结果 + 继续投放”的闭环。

## 7. Reports：模板中心 + 样张预览

Reports 不是简单的“导出按钮”，而是按周期和模板生成报告的中心。

模板包括：

- `Easy Report`
- `Custom Report`
- `Advanced Report`
- `SEO Report`
- `Product Report`
- `Marketing Report`

### 7.1 `Sample` 预览实际上可以打开

继续深挖后确认：

- 报告卡片本体是 disabled
- 但卡片右上角单独有一个可点击的 `Sample` 按钮
- 点击后会在当前页上层打开一个带 iframe 的 PDF 预览弹窗

这说明 Litlyx 的报告模板支持“先看样张，再决定是否生成”，这比只列模板说明更完整。

### 7.2 正式生成在当前账户下被锁定

继续复核正式报告生成链路后确认：

- `Easy Report / Custom Report / Advanced Report / SEO Report / Product Report / Marketing Report` 6 类报告卡片当前都是 disabled
- 卡片按钮的运行态为 `disabledProp: true`，样式里有 `opacity-70 cursor-not-allowed`
- 底部 `Generate report` 按钮也带 `disabled`，页面提示 `Select a report type to continue`
- 前三类报告的 `Sample` 按钮仍然可点，SEO / Product / Marketing 三类没有样张按钮
- 前端 chunk 里确实存在正式生成接口分支：`generate_pdf`、`generate_pdf_cust`、`generate_pdf_adv`、`generate_pdf_seo`、`generate_pdf_product`、`generate_pdf_marketing`

所以 Reports 的产品意图不是“只有样张”，而是存在完整 PDF 生成链路；只是当前 Free trial 权限下只能看模板价值和样张，不能进入正式生成。

## 8. SEO：清晰的付费门槛

SEO 页很直接：

- 锁图标
- premium 文案
- `Upgrade` CTA

它不会假装成“暂时没数据”，而是明确告诉用户这是一项存在但未解锁的能力。

## 9. Shields：治理层被单独产品化

Shields 分成三类：

- `Domains`
- `IP addresses`
- `Bot traffic`

说明 Litlyx 不只是关心“怎么采集”，也关心“怎么过滤噪音”。

特别是 `Add domain` 弹窗支持通配符，明显是 allow list 心智，而不是简单黑名单。

## 10. Analyst：AI 更像分析解释器

AI 助手的重点不是安装引导，而是分析型 prompt：

- 趋势图
- 漏斗流失
- SEO 机会
- 跳出率
- UTM 效果
- 周报

这让它更像数据解释层，而不是聊天壳。

## 11. Shareable links 与 Members：外部只读分享和协作成员被拆开

Litlyx 顶部有独立的 `Share links` 入口，路由为 `/shareable_links`。当前状态是 `0 active`，没有暴露现有分享 URL。

页面结构说明它支持：

- 选择分享域名范围
- 选择 public / protected link
- 添加可选描述
- 创建 read-only 访问链接
- 随时 revoke

这和 `Members` 是两种协作模型：`Shareable links` 给外部人只读看数据，不邀请进 workspace；`Members` 则是团队成员协作入口。

本轮没有点击 `Create link`，因为它会创建新的访问入口，属于改变云端访问状态的动作。

`Members` 的可见表现不够清晰：侧栏里它是灰态且没有有效 `href`，直接访问 `/members` 后会长时间停在 `Loading... Please wait. If this takes too long, contact the project owner.`。这不像 SEO 的 premium gate 或 Reports 的 disabled 卡片那样明确解释原因。

## 12. Plans：能力分层和 IA 绑定得很紧

Personal 与 Business 两套套餐，不只是价格不同，也直接影响：

- workspace 数
- reports
- AI
- members

FAQ 又和套餐页同屏，所以它本身就是转化链的一部分，不只是售后补充。

## 13. 最值得借鉴的模式

### 可以直接借鉴

- 登录后默认先去安装入口
- `Show test data` 作为教育机制
- Marketing 页内嵌 `Generate UTM link`
- Reports 做成模板中心
- Shields 作为治理层独立建模
- Shareable links 与 Members 分别承载外部只读分享和内部协作
- AI 助手提供任务导向的 prompt 模板

### 需要警惕的问题

- `Setup events` 看起来像站内设置入口，但实际打开外部文档，新标签页反馈不明显时容易造成“无响应”的错觉
- `Settings / Domains` 带删除语义，需要比普通设置页更强的风险提示
- `Members` 当前 loading / contact owner 态不如 SEO / Reports 的门槛解释清楚
- 顶层 `0 live / 1 live` 状态偶尔有波动，说明某些状态更新仍不够稳

## 14. tracking-demo 工件状态

当前目录下已经有一套可直接复用的最小验证工件：

- [index.html](<C:/Users/admin/Documents/src/uba/simpletrack/docs/litlyx/tracking-demo/index.html>)
- [app.js](<C:/Users/admin/Documents/src/uba/simpletrack/docs/litlyx/tracking-demo/app.js>)
- [send-event.mjs](<C:/Users/admin/Documents/src/uba/simpletrack/docs/litlyx/tracking-demo/send-event.mjs>)
- [bulk-send.mjs](<C:/Users/admin/Documents/src/uba/simpletrack/docs/litlyx/tracking-demo/bulk-send.mjs>)

它已经从“工件已就位”升级为“浏览器脚本、自定义事件、Raw Events 均已完成一轮真实验证”。

## 15. 后续建议

1. 如果后续账号解锁 Reports，再逐一实跑 `Marketing Report / Product Report / SEO Report` 的正式 PDF 生成链路。
2. 如果继续做产品借鉴，优先抽象三块：
   - 安装首屏
   - 示例数据模式
   - UTM / Reports / AI 的组合闭环
