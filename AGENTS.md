# 工作区约束

## SimpleTrack 外部参考仓库

- `Awesome-independent-tools` 作为 SimpleTrack 工具选型、独立产品参考与竞品素材来源：
  - https://github.com/yaolifeng0629/Awesome-independent-tools
- 当进行 SimpleTrack 产品定位、功能拆解、原型评审或工具型 SaaS 竞品调研时，应把该仓库作为可回看的参考资产。
- SimpleTrack 可延续到生产代码的前端原型优先使用成熟框架、成熟组件库和成熟模板，不要手写基础 UI 轮子。
- 当原型评审通过后可能继续演进为生产前端时，优先采用 Next.js 方向；若免费成熟方案达不到评审质量，可以评估收费成熟框架或模板。

## SimpleTrack 实施决策库

- SimpleTrack 已确定和待评审的实施决定统一维护在 `simpletrack/docs/实施决策/`。
- 每次确认新的阶段、范围、技术选型、功能边界或排期时，必须同步更新：
  - `simpletrack/docs/实施决策/README.md`
  - `simpletrack/docs/实施决策/分阶段实施计划.md`
  - 如仍未最终确认，同步写入 `simpletrack/docs/实施决策/待评审事项.md`
- 概念解释、选型问答和评审澄清统一写入 `simpletrack/docs/Q&A/`，保持一问一答格式。
- 支付服务、订阅计费、Merchant of Record 等商业化基础设施说明统一写入 `simpletrack/docs/支付服务/`。
- 决策文档必须标注状态：`已确定`、`待评审`、`已否决` 或 `暂缓`，并写明依据、影响范围和下一步动作。
- SimpleTrack 当前已确定的 P1 核心目标是“数据管道活了 + 公开产品入口”：页面浏览和自定义事件能够进入 Realtime 与 Events，同时具备产品官网 / Marketing Site / docs/quickstart。不要把 P1 产品层扩成团队/RBAC、收入归因页面、Replay/Performance、Boards/Share/API Key、Funnels/Journeys 的大而全版本。
- `simpletrack/docs/实施决策/README.md` 必须维护修订记录、实施计划完成列表、当前进度和下一步动作。
- 实施计划状态统一使用：`待完成`、`进行中`、`已完成`、`暂缓`、`已否决`。
- 每次完成任务、确认决策、改变阶段范围或发现实现偏离计划时，必须同步更新实施计划完成列表。
- 仓库治理变化也算实施进度变化；创建或推送子仓库、修改子模块 gitlink、调整远端地址、SSH key/Host 或 `core.sshCommand` 时，也必须同步更新 `simpletrack/docs/实施决策/README.md` 的修订记录、实施计划完成列表、当前进度和下一步动作。
- `src/analytics-core`、`src/simpletrack-saas` 或 `src/analytics-service` 有变更时，必须先提交并推送子仓库，再更新父仓子模块 gitlink、相关文档和父仓提交。
- 已标记 `已完成` 的任务如果进入功能重构、范围重开、验收失败或实现被替换，必须把状态重置为 `待完成`，并在修订记录中说明原因。
- P1 已确定包含 `analytics-core` 独立核心仓库建设：仓库名只用 `analytics-core`，不得带 `simpletrack` 或 `xwl`；从 xwl_bi 抽取分析数据面核心，保留 KafkaBus，前期优先 Redis Stream，不复用旧 Vue2 后台界面。
- P1 已确定包含产品官网 / Marketing Site / 公开站点：需要产品介绍、定价/订阅入口、docs/quickstart；不要把它仅理解为单张 landing page。
- SimpleTrack 生产 SaaS 模板已确定先选择 Supastarter for Next.js；MakerKit 只作为 B2B 企业控制面对照和备选，除非用户明确重开选型，不要在两者之间反复摇摆。
- SimpleTrack 支付路线先按 Supastarter 已支持的 Stripe、Lemon Squeezy、Polar、Creem、Dodo Payments provider 接入；KYC/KYB、退款、拒付、发票、税务和费用结构放到上线收费前逐项处理，不作为 P0/P1 早期阻塞。
- SimpleTrack 当前仍处于新建项目阶段，尚未进入已部署系统的历史迁移周期；`src/analytics-core`、`src/simpletrack-saas` 和 `src/analytics-service` 的 schema 调整只允许走初始化/建表/同步路径，不要提前引入迁移 SQL、backfill、兼容分支或完整迁移框架。只有在真实上线并出现历史数据升级需求后，再单独评估迁移逻辑。
- `analytics-core` 的实施方案维护在 `simpletrack/docs/实施决策/analytics-core实施方案.md`；每次修改其模块边界、EventBus、命名映射、存储模型或验收标准时，必须同步更新实施决策 README 的修订记录和实施计划完成列表。
- `analytics-core` 和 SimpleTrack 分析产品参考采用“双参考”：Umami 用于分析对象体系、事件语义、Realtime/Events/Funnels/Journeys/Retention/Segments 边界；Litlyx 用于短接入链路、Raw Events 验收、Product 空态/示例态/真实态和 Show test data 教育方式。
- `analytics-core` 的 P1-001 EventBus 抽象已完成：Redis Stream 采用 pending 优先重试，写入成功后 ack，超过 `MaxAttempts` 进入死信队列；下一步主线是 P1-002 的 collect、ClickHouse `EventWriter`、`TableRouter` 和 Realtime/Events 最小闭环。
- `analytics-core` 的 P1-002 已启动：collect 请求标准化、`collect.Handler`、Fiber `POST /collect`、storage `EventWriter` 接口、ClickHouse `TableRouter`、native batch `BatchWriter`、GORM/MySQL `IngestionStatusGuard`、`EventQueryBuilder` query plan、`storage.EventReader` 查询执行器、`ingestion.Processor` worker 边界和本地 Redis/MySQL/ClickHouse compose 已落地；P1-002B 的 browser / OS / device 派生和 geo enrich 边界已落在 collect stage，`simpletrack-anaysitics-service` 通过 `ANALYTICS_SERVICE_GEOIP_MMDB_FILE` 装配离线 MaxMind mmdb；默认使用高位端口，Redis 集成测试使用 `127.0.0.1:26379`；下一步不要绕开这些契约，端到端运行入口必须复用它们。
- `src/simpletrack-saas` 在 Windows 下验证 Supastarter 时使用 Node 24.1.0 或其他满足 Prisma 要求的版本（Node 20.19+、22.12+、24.0+）；Node 22.10.0 会导致 Prisma preinstall 失败。
- `src/simpletrack-saas` 如果 npm/pnpm 网络失败，优先设置 `HTTP_PROXY`、`HTTPS_PROXY`、`npm_config_proxy`、`npm_config_https_proxy` 为 `http://localhost:64320`；如果仍失败，再切到 `http://localhost:7897`，并设置 `npm_config_registry=https://registry.npmjs.org/`，避免落到不稳定镜像源。
- `src/simpletrack-saas` 的 `saas` type-check 如果报 `packages/database/prisma/generated/client` 缺失，先运行 `pnpm --filter @repo/database run generate`，再重跑 type-check。
- 后续遇到依赖安装、网络代理、SSH 权限、子仓库推送、构建验证、数据库连接等卡壳问题时，不要把排障细节继续写进 README；统一记录到 `docs/开发环境卡壳问题记录.md`，README 只保留初始化和常用命令入口。
- `references/xwl_bi-backend/` 是从本地 `xwl_bi` 复制进来的只读临时参考快照，主要用于参考后端架构设计：模块边界、启动装配、消费链路、ClickHouse 写入/查询分层、元数据流转和分析服务拆分；不要把它当作活跃模块开发，不要直接照搬旧业务代码或旧命名。
- 如需刷新 `references/xwl_bi-backend/`，必须按“重新快照”的方式整体替换，并在 `references/xwl_bi-backend/README.md` 与实施决策文档中记录新的来源 commit。

## 源码分析引用规范

- 进行源码分析、源码解读、数据流分析、架构拆解或审查说明时，引用具体代码必须同时标注对应仓库的 `commit id` 和行号；只写文件路径或只写行号不算完整证据。
- 分析 `src/analytics-core`、`src/analytics-service`、`src/simpletrack-saas` 等子仓库代码时，必须使用子仓库自己的 `HEAD` commit id；不要用父仓 commit id 代替子仓代码版本。
- 如果被分析仓库存在未提交改动，必须在分析文档或回复中明确标注“基于 `HEAD <commit id>` + 工作区未提交改动”，并说明行号来自当前工作区文件。
- Markdown 文档中的推荐引用格式为：`仓库: <repo>, commit: <short-or-full-sha>, file: <path>:<line>`；跨多行代码可写成 `<path>:<start>-<end>`。

## Git 提交规范

- 代码改动必须采用结对编程流程：主代理负责实现，另起一个 Codex native 子代理负责代码审查；子代理必须在本 `AGENTS.md` 约束下审查当前 diff、测试覆盖、godoc 注释、边界风险和是否混入无关改动。
- 子代理代码审查应在提交前完成；发现必须修复的问题时，主代理先修复并重新验证，再按需追加复审。最终交付说明需要简要写明审查结果或剩余风险。
- 该规则适用于代码文件和会影响代码行为的配置、脚本、迁移、测试文件；纯文档更新不强制启动子代理，除非用户明确要求审查。
- 提交和推送默认遵循 `$git-commit-cn` 的流程：先核对 `git status --short --branch`、`git diff --stat`、`git diff --name-status`，只 stage 本次任务相关文件，禁止用 `git add .` 混入 IDE 配置、日志、缓存、临时目录或无关未跟踪文件。
- 提交信息使用英文，不使用中文提交正文；但正文结构沿用 `$git-commit-cn` 的分路径说明方式，按文件路径分组列出每个文件的具体修改点。
- 如果仓库或上层 AGENTS 要求 Lore Commit Protocol，英文提交信息仍需保留有价值的 `Constraint:`、`Rejected:`、`Confidence:`、`Scope-risk:`、`Directive:`、`Tested:`、`Not-tested:` 等 trailer。
- 用户调用 `$git-commit-cn` 或明确要求提交时，默认在提交后继续 push；只有需要 force push、rebase、merge、解决冲突、远端不明确或会推送明显无关历史提交时，才停下来说明风险并等待确认。
- 涉及 `src/analytics-core`、`src/simpletrack-saas` 或 `src/analytics-service` 的改动，必须先在子仓库按上述英文提交规范 commit/push，再回到父仓更新 submodule gitlink、实施决策文档和父仓提交。

## Go 代码注释与 godoc 规范

- Go HTTP 服务入口优先使用成熟第三方框架或活跃第三方 HTTP 库；只有没有合适成熟方案时才考虑标准库 `net/http` 直接作为服务入口。`analytics-core` 的 HTTP 适配层和 `simpletrack-anaysitics-service` 运行时入口当前使用 Fiber v3；不使用标准库 router，也不沿用 xwl_bi 中低活跃的 `buaazp/fasthttprouter` 路由层。
- 修改 Go 代码时必须按 Go 标准库 `$GOROOT/src` 的 godoc 质量作为唯一参照，尤其适用于 `src/analytics-core`。
- 所有导出的函数、类型、接口、常量、变量和结构体字段必须有英文 godoc 注释，100% 覆盖；注释必须以被声明对象名称开头，例如 `// EventBus publishes validated events ...`。
- 结构体字段和接口方法/字段注释是强制项：新增或修改任何 struct/interface 时，每个字段、方法、嵌入成员都必须说明职责、输入输出语义或边界约束；即使是非导出类型，只要属于核心链路、adapter、测试假对象或容易误用的配置，也必须补英文注释。
- 包级注释必须以 `Package xxx ...` 开头，说明包职责、使用场景和边界。
- 注释统一使用英文；禁止在 Go 代码注释中写中文解释。
- godoc 注释使用标准格式：单行使用 `// Name ...`，多行可用 `/* */`，但仍需保持 godoc 可渲染、可读。
- 句首大写；简短单句注释通常不加句号，若后续有多句说明、NOTE、WARNING 或 Example 段落，则按英文段落正常使用标点。
- 结构体字段优先使用行末注释，例如 `TenantID string // tenant boundary key`；同类字段说明不要拆到字段上方单独成行。
- 接口方法优先在方法上一行写注释，例如 `// Publish appends one validated event to the queue.`；不要只在接口类型总注释里笼统说明后省略方法语义。
- 同一结构体内连续字段默认不留空行；只有需要表达明确语义分组时才允许空行，并在分组起始处添加英文分组注释，例如 `// Group: ingestion metadata`。
- 非导出标识符只要业务含义、边界条件、副作用或性能特征无法让同类 Go 开发者在 3 秒内看懂，就必须补英文注释；简单常量或自解释局部变量可豁免。
- 复杂路径必须在关键行上方或行尾补英文注释，解释为什么这样做，而不是复述代码做了什么；范围包括阶段切换、状态机、关键依赖装配、option pattern、plugin load、多层条件分支、早期 return、降级、熔断、重试、缓存回源、并发原语、goroutine、init、background task、metric 注册、反射、unsafe、cgo 和性能优化 trick。
- 函数体内部注释是强制项：凡是函数体超过约 10 行、包含外部副作用、状态变更、数据库/队列/网络调用、事务/幂等/重试/回滚/ack/死信、动态 SQL/动态表、并发或多分支错误处理，必须在每个关键阶段前写英文注释，说明该阶段的意图、边界和失败语义；简单 getter、纯校验小函数和自解释 one-liner 可豁免。
- 核心链路函数即使不足 10 行也必须写函数体阶段注释，范围包括 collect、ingestion、EventBus、Redis Stream、Kafka、ClickHouse、MySQL/GORM、TableRouter、query builder、worker、checkpoint、幂等、死信队列和动态分表路由。
- 函数体注释必须按“阶段”而不是“逐行复述”组织，例如先说明 request normalization / validation，再说明 claim / idempotency，再说明 durable append / batch insert，再说明 commit / rollback / ack / dead-letter；禁止用低价值注释凑数量。
- 函数体注释验收按逻辑块检查：一个非平凡函数如果同时存在输入整理、依赖调用、状态写入、错误分类、回滚或提交等多个逻辑块，却没有在每个逻辑块前写清意图和失败语义，视为注释不合格，不得提交或宣称完成。
- 注释强度必须随架构风险提高：HTTP/队列/存储/查询/幂等/重试/ack/死信/批量写入/动态表路由等边界代码，至少要在包注释、核心类型注释和核心函数注释中说明职责、不负责什么、为什么依赖只能停留在当前层。
- 框架适配层必须写清楚边界：例如 `httpapi` 可以认识 `fiber.Ctx`，但 `collect.Handler`、`EventBus`、`ingestion`、`storage` 不应接收 HTTP 框架对象；注释中优先使用“framework coupling”“boundary crossing”等明确表述，避免使用容易误解的“pollution”。
- 核心处理器必须说明输入、输出、副作用和错误分类。例如 collect handler 要写明输入是 `collect.Request`，输出是 `EventEnvelope`，副作用是发布到 `EventBus`，validation error 与 publish error 的语义不同。
- 新增 HTTP/队列/存储入口时，必须同时补齐：包级边界注释、核心构造函数注释、主处理函数注释、至少一个可编译 Example，以及覆盖正常路径和关键错误路径的测试。
- 禁止为了满足注释数量写低价值注释，例如“set status code”“return error”这类复述代码的注释；注释应解释命名无法承载的边界、原因、约束、风险和长期维护意图。
- 公共 API 或容易误用的函数/类型必须提供可编译的 `Example{Name}` 示例，覆盖正常路径、错误路径和常见配置场景；示例应通过 `go test` 校验，并在需要时包含期望输出。
- 示例密度不得低于每 10 个导出标识符 1 个完整 Example；新增公共 API 时优先随代码一起补示例，而不是事后集中补。
- 存在并发安全、性能陷阱或资源泄露风险时，godoc 首段必须使用 `NOTE:` 或 `WARNING:` 标注关键风险。
- 注释整改应随功能修改小步完成；如果需要全库注释整改，先提交独立整改计划，避免一次性机械刷大量低价值注释。
- 上述注释规范是强制性的，不是建议。凡是新增或修改架构边界、框架适配层、核心处理器、队列、存储、查询、幂等、重试、ack、死信、批量写入、动态表路由等代码，如果缺少对应边界注释、核心类型/函数 godoc、结构体字段注释、接口方法注释、函数体关键阶段注释和必要 Example，任务视为未完成，不得提交或宣称完成。
- Go 代码整改完成后，按可用工具执行 `go doc -all`、`golint`、`go vet`、`go test -run Example` 和 `go test ./...`；如果仓库启用 `golangci-lint`，CI 必须包含并通过 `godox`、`gomnd`、`exhaustive` 等相关检查，未通过不得合并。

## Umami 调研资产规范

当修改 `simpletrack/docs/umami/` 下的文件时，交付物必须保持三层结构：

1. 截图
2. 截图索引
3. 操作流说明

必须遵守这些规则：

- 新截图统一放在 `simpletrack/docs/umami/snapshots/phase-*/`
- 每张新增截图都必须同步更新到：
  - `simpletrack/docs/umami/快照索引.md`
  - 对应阶段的 `flow.md`
  - 如果截图范围发生变化，还要更新对应阶段的 `README.md`
- `simpletrack/docs/umami/快照进度.md` 必须反映真实完成状态和当前缺口
- `simpletrack/docs/umami/Umami功能深度分析.md` 需要解释截图代表什么，不能只是贴路径

## 目录边界

- `simpletrack/prototype/simpletrack-umami-inspired/` 保持在 `prototype/` 下，不迁移
- Umami Cloud 调研资产统一放在 `simpletrack/docs/umami/`
- 不要把账号、cookie、token 或其他敏感信息写入仓库文件

## GitHub SSH 仓库权限

- 如果新机器或新环境缺少 `id_ed25519_simpletrack`，先生成专用 key、把公钥添加到 GitHub 对应账号或组织授权，再验证 `github-simpletrack`：
  - `ssh-keygen -t ed25519 -C "simpletrack" -f "$env:USERPROFILE\.ssh\id_ed25519_simpletrack"`
  - `Get-Content "$env:USERPROFILE\.ssh\id_ed25519_simpletrack.pub"`
  - `ssh -T git@github-simpletrack`
- `src/analytics-core`、`src/simpletrack-saas` 和 `src/analytics-service` 是独立子仓库，推送到 `simpletrack` GitHub 组织时必须使用专用 SSH 配置：
  - `$sshConfig = "$($env:USERPROFILE -replace '\\','/')/.ssh/config_simpletrack"`
  - `git -C ".\src\analytics-core" config core.sshCommand "ssh -F $sshConfig"`
  - `git -C ".\src\simpletrack-saas" config core.sshCommand "ssh -F $sshConfig"`
  - `git -C ".\src\analytics-service" config core.sshCommand "ssh -F $sshConfig"`
- 三个子仓库的 `origin` 必须使用 `github-simpletrack` Host 别名：
  - `git@github-simpletrack:simpletrack/analytics-core.git`
  - `git@github-simpletrack:simpletrack/simpletrack-saas.git`
  - `git@github-simpletrack:simpletrack/anaysitics-service.git`
- 不要依赖默认 `$env:USERPROFILE\.ssh\config` 推送这三个仓库；该文件曾因 Windows ACL 权限异常导致 OpenSSH 报 `Bad owner or permissions`。
- 在 PowerShell 中读取、验证或引用带括号/空格的路径时，优先使用单引号加 `-LiteralPath`，不要让 PowerShell 重新解释路径片段；例如 `Get-Content -LiteralPath 'C:\Users\admin\.ssh\config_simpletrack'`。否则像 `(authenticated)` 这样的路径片段可能被拆成裸标识符并触发 `The term 'authenticated' is not recognized` 这类解析错误。
- 相关说明维护在 `simpletrack/docs/Q&A/Windows-SSH仓库权限怎么配置.md`。

## 截图评审标准

- 优先同时保留页面态和关键交互态
- 只要弹窗、下拉、tab 切换、筛选条件变化会影响理解，就要单独截图
- 如果请求已经成功但页面仍然空白，必须把它作为产品发现明确记录，不能模糊处理
- in-app browser 或 Playwright MCP 只在截图、交互验证和本地预览期间保持开启；暂时不使用时先关闭，等下次需要再开启，避免浏览器长时间挂着出现断流或会话失效。

## 中文文档安全

- 修改中文文档时保持 UTF-8
- 文本编辑优先使用 `apply_patch`
- 避免使用不安全的 PowerShell 中文读写方式
