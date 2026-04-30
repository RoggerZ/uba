# SimpleTrack

`SimpleTrack` 是一个面向中小型 SaaS 团队的极简行为分析产品方案，重点不是做“大而全”的通用分析平台，而是聚焦转化漏斗、关键事件追踪和 AI 洞察。

## 产品定位

核心方向：

- 5 分钟集成
- 面向 SaaS 场景的关键行为追踪
- 转化漏斗分析优先
- AI 每周自动输出洞察与建议
- 固定定价，避免按事件量快速涨价

目标用户主要是：

- 10 到 100 人的 SaaS 团队
- 预算有限但需要数据支持的创业公司
- 没有专职数据分析师的产品团队

## 当前目录

```text
simpletrack/
├── docs/
│   ├── SimpleTrack-极简SaaS分析工具方案.md
│   ├── 产品方向决策-竞品调研与差异化分析.md
│   ├── 市场验证与获客策略.md
│   ├── 成本控制与盈利策略.md
│   ├── 技术实现方案-架构设计与开发指南.md
│   ├── 推广营销策略-Product Hunt与社区增长.md
│   ├── fathom/
│   ├── plausible/
│   └── umami/
├── prototype/
│   └── simpletrack-umami-inspired/
└── .tmp/
```

## 文档导航

### 核心方案

- `docs/SimpleTrack-极简SaaS分析工具方案.md`
  产品总方案，适合先看，了解定位、用户、MVP 功能和价值主张。
- `docs/技术实现方案-架构设计与开发指南.md`
  技术落地草案，包含 Go 后端、数据层、服务层以及建议的项目结构。

### 产品决策

- `docs/产品方向决策-竞品调研与差异化分析.md`
  用于判断和主流竞品的差异化空间。
- `docs/成本控制与盈利策略.md`
  关注个人或小团队做这类产品时的成本与变现模型。

### 增长与验证

- `docs/市场验证与获客策略.md`
  关注早期验证和种子用户获取。
- `docs/推广营销策略-Product Hunt与社区增长.md`
  更偏对外增长和启动打法。

### 竞品拆解

- `docs/fathom/Fathom功能深度分析.md`
- `docs/plausible/Plausible功能深度分析.md`
- `docs/umami/Umami功能深度分析.md`

这两份文档更适合在做功能边界取舍时参考。

### Umami 研究资产

- `docs/umami/README.md`
- `docs/umami/快照索引.md`
- `docs/umami/快照进度.md`
- `docs/umami/tracking-demo/`
- `prototype/simpletrack-umami-inspired/`

其中 `docs/umami/` 存放 Umami Cloud 调研、快照与样例，`prototype/simpletrack-umami-inspired/` 继续保留为 SimpleTrack 自己的原型目录。

## 建议阅读顺序

1. `docs/SimpleTrack-极简SaaS分析工具方案.md`
2. `docs/产品方向决策-竞品调研与差异化分析.md`
3. `docs/技术实现方案-架构设计与开发指南.md`
4. `docs/市场验证与获客策略.md`
5. `docs/成本控制与盈利策略.md`

## 当前状态

- 当前以方案文档为主
- 技术架构已有较明确草案
- 目录里暂未落地对应的正式应用代码

如果后续进入实现阶段，建议优先补齐：

- MVP 功能边界
- SDK 与事件模型定义
- 数据表设计与采集链路
- 本地开发与部署说明
