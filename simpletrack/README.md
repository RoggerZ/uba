# SimpleTrack

## 项目初始化和依赖安装

SimpleTrack 当前有两个主要实施子仓：

- `src/simpletrack-saas`：Supastarter for Next.js 工作副本，承接 SaaS 控制面、公开站点和控制台 shell。
- `src/analytics-core`：独立业务无关分析核心仓库，承接事件数据面。

首次拉取父仓后先同步子模块：

```powershell
git pull --ff-only
git submodule update --init --recursive
```

验证 `src/simpletrack-saas` 时使用 Node 24.1.0：

```powershell
nvm use 24.1.0
```

网络不稳定时使用本机代理：

```powershell
$env:HTTP_PROXY = "http://localhost:7897"
$env:HTTPS_PROXY = "http://localhost:7897"
$env:npm_config_proxy = "http://localhost:7897"
$env:npm_config_https_proxy = "http://localhost:7897"
$env:npm_config_registry = "https://registry.npmjs.org/"
```

安装 Supastarter 工作副本依赖：

```powershell
cd src/simpletrack-saas
& "$env:USERPROFILE\AppData\Local\nvm\v24.1.0\npm.cmd" exec --yes pnpm@10.9.0 -- --config.manage-package-manager-versions=false --config.package-manager-strict=false --registry=https://registry.npmjs.org/ install --frozen-lockfile
```

生成 Prisma client 并验证当前 SimpleTrack 相关 app：

```powershell
$env:DATABASE_URL = "postgresql://postgres:postgres@127.0.0.1:5432/simpletrack"
& "$env:USERPROFILE\AppData\Local\nvm\v24.1.0\npm.cmd" exec --yes pnpm@10.9.0 -- --config.manage-package-manager-versions=false --config.package-manager-strict=false --filter @repo/database run generate
& "$env:USERPROFILE\AppData\Local\nvm\v24.1.0\npm.cmd" exec --yes pnpm@10.9.0 -- --config.manage-package-manager-versions=false --config.package-manager-strict=false --filter saas --filter marketing --filter docs run type-check
```

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

- 当前处于 P0 产品与底座确认阶段，并已进入部分 P1 前置底座实施。
- 生产 SaaS 模板已确定先选择 Supastarter for Next.js。
- `src/simpletrack-saas/` 已从 Supastarter 初始化为独立子仓库，已挂载 Websites、Realtime、Events 页面草案，完成 Websites 页 UI-only subscription gate，并替换 marketing/docs 占位内容。
- `src/analytics-core/` 已创建为独立业务无关分析核心仓库，已验证 Redis Stream publish / consume / ack 链路。
- `simpletrack/docs/实施决策/` 是当前实施进度和阶段计划的主要来源。

如果后续进入实现阶段，建议优先补齐：

- Supastarter mail-preview、marketing/docs 浏览器预览和截图级验证
- `analytics-core` 失败重试、死信队列和幂等入库边界
- tracker -> collect -> storage -> Realtime/Events 最小闭环
- 产品官网 / Marketing Site / docs/quickstart 截图级验证
