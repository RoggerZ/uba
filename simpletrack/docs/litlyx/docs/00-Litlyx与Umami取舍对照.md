# Litlyx 与 Umami 取舍对照

> 目标：不是比较谁更强，而是回答“这两种产品思路分别强调什么，以及 SimpleTrack 现在更该借哪一种”。

## 一句话判断

- Litlyx 更强调：更短的接入链路、更强的产品教育、更早暴露商业化与治理能力。
- Umami 更强调：分析对象体系、更完整的报表模型、更细粒度的分析与工作台抽象。

所以如果 SimpleTrack 当前更接近“让用户尽快接入并看到价值”，Litlyx 的参考价值更高；如果更接近“做完整分析工作台”，Umami 的参考价值更高。

## 1. 首次体验

| 维度 | Litlyx | Umami | 对 SimpleTrack 的建议 |
| --- | --- | --- | --- |
| 登录后默认落点 | 直接进入 `Web` 安装页 | 更像标准 analytics 控制台起点 | 早期优先借 Litlyx，把“接入成功”放在第一位 |
| 安装方式表达 | Script、GTM、框架入口、AI prompt 同屏 | 更偏 tracker 和配置本身 | 如果目标是降低首用门槛，Litlyx 的首屏更强 |
| 首次教育 | `Show test data`、模板样张、AI 示例问题 | 更偏正式分析模型和对象体系 | MVP 阶段先借 Litlyx 的教育方式，再补 Umami 的体系化分析 |

## 2. 分析产品形态

| 维度 | Litlyx | Umami | 对 SimpleTrack 的建议 |
| --- | --- | --- | --- |
| Product 主体 | 事件聚合、Top Events、Funnel、User Flow、Metadata 同屏骨架 | Reports/Events/Filters 等对象更细分 | 起步阶段可以先做 Litlyx 式“单页骨架”，减少导航成本 |
| 示例数据 | 内建 `Show test data` | 更依赖真实数据和分析对象 | Litlyx 这点值得优先借鉴 |
| 深度分析 | 当前看到 Reports 模板中心、SEO 门槛、AI 解释层 | Funnels、Journeys、Retention、Filters、Segments 更成熟 | 进阶阶段要补 Umami 那种对象化分析能力 |

## 3. 增长与商业化

| 维度 | Litlyx | Umami | 对 SimpleTrack 的建议 |
| --- | --- | --- | --- |
| Marketing | 自带 UTM 生成器，分析页直接闭环 | UTM/归因更多是报表体系里的能力 | 如果要服务增长团队，Litlyx 的闭环更适合先落地 |
| Reports | 模板中心 + Sample 样张 + 生成门槛 | 报表能力更分析导向 | 可以先做 Litlyx 的模板中心，再逐步补 Umami 式深分析 |
| 套餐表达 | Plans、FAQ、SEO 门槛、disabled reports 都很直接 | 商业化露出没有 Litlyx 这么前台化 | 如果产品本身带商业化目标，Litlyx 更像现成模板 |

## 4. 治理、分享、协作

| 维度 | Litlyx | Umami | 对 SimpleTrack 的建议 |
| --- | --- | --- | --- |
| 数据治理 | `Settings / Domains` + `Shields` 两层治理 | 更偏分析体系和工作台能力 | Litlyx 在“数据干净”这件事上的产品化更直接 |
| 外部分享 | `Shareable links` 独立页 | Umami 更强调 dashboards/boards 组合 | 如果先做对外只读分享，Litlyx 路径更短 |
| 内部协作 | `Members` 入口存在但当前表达弱 | Umami 有更强的工作台与组织抽象潜力 | 协作能力要借 Umami 的结构化思路，但避免 Litlyx 现在这种 loading 态表达 |

## 5. 对 SimpleTrack 的取舍建议

### 现在更值得直接借 Litlyx 的

- 登录后默认安装入口
- `Show test data`
- `Raw Events` 作为验收页
- Marketing 页内嵌 `Generate UTM link`
- Reports 模板中心与样张预览
- 治理层单独产品化
- Shareable links 与 Members 分离
- AI 用任务模板引导，而不是空聊天框

### 现在更值得直接借 Umami 的

- 分析对象分层
- Filters / Segments / Cohorts 这种可复用分析口径
- Funnels / Journeys / Retention 的正式分析模型
- Dashboard / Boards 这种可扩展工作台

### 不建议直接照搬的

- Litlyx 的 `Members` 受限态表达
- Litlyx 把 `Setup events` 做成看起来像站内配置、实际却是外链文档的入口
- Umami 那种把太多高级分析对象过早放进 MVP 导航的复杂度

## 6. 最实际的组合路线

对 SimpleTrack 来说，更稳的一条路不是二选一，而是：

1. 先用 Litlyx 的思路做“更容易接入、更容易看到价值”的壳
2. 再用 Umami 的思路补“更完整的分析对象体系”
3. 最后再决定是否需要 Umami 那种更重的看板、细分与扩展层

一句话收束：

- Litlyx 更像“先让你成功”
- Umami 更像“再让你变强”
