# Litlyx 功能矩阵

> 用途：把 Litlyx 的后台能力拆成可对标、可复核、可决策的产品矩阵。主文档负责叙事分析，本文件负责快速查证和 SimpleTrack 借鉴取舍。

## 1. 总体功能地图

| 层级 | Litlyx 模块 | 主要任务 | 关键证据 | 对 SimpleTrack 的判断 |
| --- | --- | --- | --- | --- |
| 账号层 | Workspaces | 项目切换、套餐状态、升级入口 | `P01-S04`、`REF-02` | 值得保留账号级项目列表，不要把项目切换塞进分析侧栏 |
| 接入层 | Web / Install | 脚本接入、GTM 接入、AI 辅助接入、安装验证 | `P01-S03`、`P02-S01`、`P02-S02`、`REF-01` | 强烈建议把首次登录默认落点放到安装页 |
| 配置层 | Settings / General | 工作区名、ID、脚本、删除工作区 | `P02-S03`、`REF-06` | 安装信息需要二次暴露，便于用户回查 |
| 数据清理层 | Settings / Domains | 按域名删除 visits/events，执行域名 sanitization | `P02-S05` | 可借鉴，但必须强化危险操作提示和审计语义 |
| 产品分析层 | Product | 事件聚合、Top Events、漏斗、用户流、元数据分析 | `P03-S01`、`P03-S02`、`P03-S05` | 应同时支持空态骨架、示例态和真实态 |
| 明细层 | Raw Events | 事件明细表，验证采集是否入库 | `P03-S03`、`P03-S06` | Raw 表是排障刚需，不应只给聚合图 |
| 增长分析层 | Marketing | 渠道、来源、社交、UTM 生成 | `P04-S01`、`P04-S02`、`P04-S03` | UTM 生成器应和营销分析放在同一工作流里 |
| 报告层 | Reports | 模板报告、样张预览、正式 PDF 生成入口 | `P04-S04`、`P04-S05`、`P04-S07` | 报告中心比单个导出按钮更有产品空间 |
| 付费门槛层 | SEO | 高级 SEO 能力入口和升级 CTA | `P04-S06`、`REF-07` | 锁定功能要明确说明是 premium，不要伪装成空数据 |
| 治理层 | Shields | 域名 allow list、IP 排除、bot 流量过滤 | `P05-S01` 到 `P05-S04` | Analytics 产品需要独立治理层，避免测试和内部流量污染 |
| AI 层 | Analyst | 数据解释、趋势、漏斗、SEO、报告 prompt | `P05-S05`、`P05-S06`、`REF-08` | AI 应做任务入口，不只是聊天框 |
| 分享层 | Shareable links | 外部只读 dashboard 分享、public/protected link | `P05-S10` | 外部只读分享应和团队成员邀请分开 |
| 协作层 | Members | 内部成员协作入口 | `P05-S11` | 当前 Litlyx 表达弱，SimpleTrack 应给清楚门槛或错误原因 |
| 商业化层 | Plans | Personal / Business 套餐、FAQ、能力边界 | `P05-S07` 到 `P05-S09`、`REF-09` | 套餐页应解释功能边界，而不仅是价格表 |

## 2. 核心工作流矩阵

| 工作流 | Litlyx 路径 | 关键动作 | 当前验证状态 | 设计结论 |
| --- | --- | --- | --- | --- |
| 首次接入 | Login -> Web | 登录后直接进入安装页，选择 Script 或 GTM | 已截图验证 | 首屏先解决“怎么接入”，比空 dashboard 更适合新用户 |
| 脚本回查 | Settings -> General | 在设置里再次查看 Workspace ID 和脚本 | 已截图验证 | 安装信息应可重复找到 |
| 自定义事件学习 | Product -> Setup events | 点击后打开 Custom Events 文档新标签页 | 已复核外链行为 | CTA 文案要提示会打开文档，避免误判无响应 |
| 真实事件验证 | tracking-demo -> Product / Raw Events | 本地页面和批量脚本发送 14 条事件 | 已真实验证 | 必须同时检查聚合页和明细页，才算采集链路跑通 |
| 示例数据教育 | Product / Marketing -> Show test data | 切换到 demo mode | 已截图验证 | 示例数据是低成本产品教育机制 |
| 营销闭环 | Marketing -> Generate UTM link | 在分析页内生成 UTM | 已截图验证 | 分析和投放工具应互相靠近 |
| 报告体验 | Reports -> Sample | 点击 Sample 打开 PDF 预览 | 已截图验证 | 付费前先给样张，能解释高级能力价值 |
| 报告生成 | Reports -> disabled cards | 正式报告卡片与 Generate report 不可用 | 已确认受限 | 需要解锁后再验证 POST 生成链路 |
| 数据治理 | Settings / Domains、Shields | 删除域名数据、allow list、IP 排除、bot 过滤 | 已页面态验证 | 这些是高风险或高影响配置，交互上要给更强确认 |
| 外部分享 | Shareable links | 创建只读链接 | 只验证空态，没有创建 | 创建链接会改变访问状态，后续需授权再测 |

## 3. 空态、示例态、真实态的分层

| 模块 | 空态表现 | 示例态 | 真实态 | 产品启发 |
| --- | --- | --- | --- | --- |
| Product | 保留 Top Events、Funnel、User Flow、Metadata 骨架 | `Show test data` 后填充示例事件 | `Total events: 14` 和 Raw Events 明细可见 | 三态都要有，不能只做空态提示 |
| Marketing | 渠道、来源、社交模块空数值 | 示例曲线和来源榜单 | 本轮未专门制造营销来源数据 | 示例态能展示未来价值 |
| Reports | 模板中心可见 | Sample PDF 可预览 | 正式生成被 Free trial 锁定 | 先展示样张，再引导升级 |
| SEO | 不做空数据伪装，直接 premium gate | 无 | 当前未解锁 | 付费门槛要透明 |
| Members | loading / contact owner | 无 | 当前不可用原因不清 | 受限态不应只有 loading |

## 4. 权限与风险边界

| 行为 | 是否已执行 | 原因 | 后续条件 |
| --- | --- | --- | --- |
| 真实事件发送 | 已执行 | 只写入当前调研工作区的测试事件，且用于验证接入链路 | 如需清理，需要明确授权后走 Domains 删除流程 |
| Reports 正式生成 | 未执行 | 当前 Free trial 下按钮 disabled | 账号解锁后再测 |
| Shareable link 创建 | 未执行 | 会创建新的外部只读访问入口 | 需要用户明确授权 |
| Domain data 删除 | 未执行 | 带删除 visits/events 语义 | 需要用户明确授权 |
| Domain sanitization | 未执行 | 会改变域名数据归属和清理状态 | 需要用户明确授权 |
| Members 邀请 | 未执行 | 当前入口受限且可能改变协作状态 | 需要账号权限和用户授权 |

## 5. SimpleTrack 借鉴优先级

| 优先级 | 模式 | 为什么优先 | 建议落点 |
| --- | --- | --- | --- |
| P0 | 登录后默认安装入口 | 新用户第一需求是接入，不是看空图表 | Onboarding 首页 |
| P0 | Show test data | 立刻解释产品价值，降低空态挫败 | Product、Marketing、Reports |
| P0 | Raw Events 明细表 | 采集排障必须有可查证的事件明细 | Product 下钻或独立 Raw Data |
| P1 | UTM 生成器内嵌 Marketing | 把分析和投放动作连起来 | Marketing 工具区 |
| P1 | Reports 模板中心 | 高级能力可以产品化，而不只是导出 | Reports 中心 |
| P1 | Shields 治理层 | 埋点产品需要过滤内部流量、bot、错误域名 | Settings 之外的独立治理入口 |
| P1 | Shareable links 与 Members 分离 | 外部客户只读和内部团队协作是两种权限模型 | 顶部分享入口 + 团队设置 |
| P2 | AI prompt 模板 | 让 AI 变成分析动作入口 | Analyst / Insights |
| P2 | 套餐 FAQ 同页 | 直接解释限制，减少用户猜测 | Plans 页面 |

## 6. 当前仍需后续验证的点

- 正式 PDF 报告生成链路：需要账号权限解锁。
- Members 入口的真实限制原因：当前只看到 loading / contact owner，没有明确是权限、套餐还是接口问题。
- Marketing 真实来源数据表现：当前主要验证了 Product 事件链路，营销来源可以后续用带 UTM 的页面访问补测。
- 数据清理闭环：Domain data 删除和 sanitization 都是高风险动作，需要明确授权后才能实测。
