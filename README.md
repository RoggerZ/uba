# UBA

## 项目初始化和依赖安装

首次拉取工作区后，先同步子模块：

```powershell
git pull --ff-only
git submodule update --init --recursive
```

`src/simpletrack-saas` 是 SimpleTrack 的 Supastarter 工作副本。Windows 本地验证建议使用 Node 24.1.0，避免 Prisma 在 Node 22.10.0 下安装失败：

```powershell
nvm use 24.1.0
```

如果 npm / pnpm 下载依赖失败，优先使用本机代理和官方 npm registry：

```powershell
$env:HTTP_PROXY = "http://localhost:7897"
$env:HTTPS_PROXY = "http://localhost:7897"
$env:npm_config_proxy = "http://localhost:7897"
$env:npm_config_https_proxy = "http://localhost:7897"
$env:npm_config_registry = "https://registry.npmjs.org/"
```

进入 `src/simpletrack-saas` 后安装依赖：

```powershell
& "$env:USERPROFILE\AppData\Local\nvm\v24.1.0\npm.cmd" exec --yes pnpm@10.9.0 -- --config.manage-package-manager-versions=false --config.package-manager-strict=false --registry=https://registry.npmjs.org/ install --frozen-lockfile
```

如果 `saas` type-check 报 Prisma generated client 缺失，先生成 Prisma client：

```powershell
$env:DATABASE_URL = "postgresql://postgres:postgres@127.0.0.1:5432/simpletrack"
& "$env:USERPROFILE\AppData\Local\nvm\v24.1.0\npm.cmd" exec --yes pnpm@10.9.0 -- --config.manage-package-manager-versions=false --config.package-manager-strict=false --filter @repo/database run generate
```

当前 SimpleTrack Supastarter 验证命令：

```powershell
$env:DATABASE_URL = "postgresql://postgres:postgres@127.0.0.1:5432/simpletrack"
& "$env:USERPROFILE\AppData\Local\nvm\v24.1.0\npm.cmd" exec --yes pnpm@10.9.0 -- --config.manage-package-manager-versions=false --config.package-manager-strict=false --filter saas --filter marketing --filter docs run type-check
```

这个仓库目前是一个以产品研究和方案沉淀为主的工作区，围绕行为分析产品方向拆成两个子项目：

- `simpletrack`：面向中小型 SaaS 团队的极简 Web / SaaS 行为分析方案
- `apptrack`：面向独立移动开发者的小型移动分析方案

仓库重点不是现成可运行的统一工程，而是产品定位、竞品研究、技术方案、获客与盈利策略等文档资产。

## 目录结构

```text
uba/
├── docs/             # 顶层可行性与竞品研究
├── simpletrack/      # Web / SaaS 分析方向
├── apptrack/         # 移动分析方向
└── uba.code-workspace
```

## 顶层文档

- `docs/个人独立开发者可行性分析.md`
  用于判断个人开发者是否适合从通用行为分析平台切入，以及更现实的产品路径。
- `docs/竞品分析报告.md`
  汇总通用行为分析市场的主要竞品、定价、优劣势与差异化机会。

## 子项目说明

### `simpletrack`

聚焦中小型 SaaS 团队，强调：

- 5 分钟集成
- 转化漏斗优先
- AI 周报洞察
- 固定价格、低学习成本

适合查看 Web / SaaS 行为分析工具的轻量化切入方案。

### `apptrack`

聚焦独立移动开发者和小型 App 团队，强调：

- 移动端 SDK 集成
- AI 自动分析与提醒
- 低成本替代 Firebase / Mixpanel 的部分使用场景
- 从移动开发者生态切入

目录中还包含 `Countly` 的参考研究和示例项目，用于验证移动分析 SDK 与落地方式。

## 建议阅读顺序

1. 先看顶层 `docs/`，理解为什么要避开“通用大而全分析平台”。
2. 再看 `simpletrack/docs/`，理解 Web / SaaS 方向的轻量化切口。
3. 最后看 `apptrack/docs/`，对比移动端方向在 SDK、竞品和商业化上的差异。

## 当前状态

- 已有较完整的产品与技术文档
- 尚未形成统一的生产级代码仓结构
- `apptrack/docs/countly/src/vue-countly-demo` 含参考性质的示例代码，可单独查看或运行

如果后续开始落地开发，建议先在对应子项目下补齐：

- 目标版本范围
- MVP 功能清单
- 技术选型定稿
- 代码目录与运行方式
