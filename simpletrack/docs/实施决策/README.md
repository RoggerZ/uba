# SimpleTrack 实施决策库

> 目录用途：存放 SimpleTrack 已确定要实施的决定、分阶段实施计划，以及仍需评审的关键事项。后续每次确定新决定，都先更新本目录，再继续改原型或生产代码。

## 文档索引

| 文档 | 用途 | 状态 |
| --- | --- | --- |
| [分阶段实施计划.md](分阶段实施计划.md) | 记录 P0/P1/P2/P3 的目标、范围、交付物和验收标准 | 已确定，持续更新 |
| [待评审事项.md](待评审事项.md) | 记录还没有拍板的技术栈、模板、支付、数据面复用方案 | 待评审，持续更新 |
| [技术栈底座决策.md](技术栈底座决策.md) | 记录当前已经形成的技术底座方向和边界 | 已确定 + 待评审 |
| [付费SaaS模板本地对比.md](付费SaaS模板本地对比.md) | 基于 `template-src` 本地源码和 MakerKit 官方资料对比付费模板 | 已确定先选 Supastarter，持续更新 |
| [analytics-core实施方案.md](analytics-core实施方案.md) | 记录 `analytics-core` 的 P1 抽取边界、模块草案、EventBus 方案和 xwl_bi 代码评审结论 | 已确定，设计细节持续评审 |

## 修订记录

| 日期 | 修订内容 | 影响范围 |
| --- | --- | --- |
| 2026-04-29 | 创建实施决策库，写入总方向、待评审方向和维护规则 | 决策管理 |
| 2026-04-29 | 增加修订记录、实施计划完成列表、当前进度和下一步动作 | 决策管理、阶段推进 |
| 2026-04-29 | 明确 xwl_bi 不整仓改名，改为抽取分析数据面核心仓库的待评审方向 | 数据面架构 |
| 2026-04-29 | 确认 P1 新建 `analytics-core` 独立核心仓库，前期 Redis Stream 先行，KafkaBus 保留 | 分析数据面、P1 底座 |
| 2026-04-29 | 将 P1 范围补充为包含产品官网 / Marketing Site / 公开站点 | P1 产品交付 |
| 2026-04-29 | 增加 `template-src` 本地付费 SaaS 模板对比，收敛为 Supastarter + MakerKit 两个核心候选 | SaaS 模板选型 |
| 2026-04-29 | 确认生产 SaaS 模板先选择 Supastarter for Next.js，MakerKit 降为 B2B 对照和备选 | SaaS 模板选型 |
| 2026-04-29 | 新增 `analytics-core` 实施方案，纳入 xwl_bi analyze/code-review 证据，并补充 Umami、Litlyx 参考边界 | 分析数据面、P1 实施 |
| 2026-04-29 | 支付路线改为先按 Supastarter 支持的 provider 接入，KYC/KYB、发票税务、退款拒付等放到上线前后置检查 | 支付与商业化 |
| 2026-04-30 | 补齐 analytics-core 评审 Q&A：GitHub 组织、tenant/project/source、ack/重试/死信、consumer offset、acceptance status、GORM query builder 和 UI 策略 | 分析数据面、协作规范、Q&A |
| 2026-04-30 | 确认 ClickHouse 表策略直接采用方案 B，事件写入热路径使用原生 batch writer，入库必须按 event_id 幂等去重 | analytics-core、ClickHouse、数据入库 |
| 2026-04-30 | 本地创建 `src/analytics-core` 独立仓库骨架，并从 Supastarter 初始化 `src/simpletrack-saas` 工作副本；远端推送受 GitHub 权限或仓库创建状态阻塞 | P1 底座、SaaS 工作副本 |
| 2026-04-30 | 在 `src/simpletrack-saas` 完成 Supastarter P1 页面草案 spike：挂载 Websites、Realtime、Events 到组织内导航并通过 saas type-check | Supastarter spike、P1 产品层 |
| 2026-04-30 | 为 `analytics-core` 增加 Redis Stream 集成测试，并用 `redis/redis-stack:latest` 验证 publish / consume / ack / pending=0 | analytics-core、EventBus、Redis Stream |
| 2026-04-30 | 修复 SimpleTrack 专用 SSH 身份，成功推送 `analytics-core` 与 `simpletrack-saas` 远端，并将二者作为父仓子模块挂载 | 仓库治理、P1 底座 |
| 2026-04-30 | 固化 Windows SSH 仓库权限配置、专用 key 初始化流程和 `core.sshCommand` 规则，并补充 SimpleTrack 目录 README | 仓库治理、协作规范、文档入口 |
| 2026-04-30 | 在 Supastarter Websites 草案页接入 UI-only subscription gate，验证 Free plan source 限制、锁定态和升级入口 | Supastarter spike、订阅限制、P1 产品层 |

## 实施计划完成列表

状态取值：`待完成`、`进行中`、`已完成`、`暂缓`、`已否决`。

| 编号 | 任务 | 状态 | 当前说明 | 下一步 |
| --- | --- | --- | --- | --- |
| PLAN-001 | 建立实施决策库目录 | 已完成 | 已创建 `simpletrack/docs/实施决策/` | 持续维护 |
| PLAN-002 | 形成分阶段实施计划 | 已完成 | 已明确 P0/P1/P2/P3，P1 收窄为“数据管道活了 + 公开产品入口” | 随评审更新阶段边界 |
| PLAN-003 | 建立 Q&A 目录 | 已完成 | 已创建概念解释和评审问答目录 | 新问题继续追加 |
| PLAN-004 | 建立支付服务说明目录 | 已完成 | 已创建 Stripe / Polar / Lemon Squeezy 对比，并明确先按 Supastarter provider 接入 | 上线收费前补 Paddle/Creem/Dodo 和个人开发者收款检查 |
| P0-001 | Next.js 可交互原型 | 进行中 | 原型方向已确定，需继续按生产可迁移骨架推进 | 完成 Next.js 迁移和页面验证 |
| P0-002 | SaaS 模板选型评估 | 已完成 | 已确定先选择 Supastarter for Next.js；MakerKit 保留为 B2B 对照和备选 | 进入 Supastarter 1 天 SimpleTrack spike |
| P0-003 | 企业分析控制台 UI 可改造性评审 | 进行中 | `src/simpletrack-saas` 已能在 Supastarter `apps/saas` 组织内导航挂载 Websites、Realtime、Events 页面草案 | 做截图级评审，确认密度、表格、侧边栏和空态是否满足企业分析控制台 |
| P0-004 | Supastarter for Next.js 接入核验 | 进行中 | 已确定先选 Supastarter；`src/simpletrack-saas` 已作为独立子仓库推送；Websites 页已完成 UI-only subscription gate；支付先按模板已有 Stripe、Lemon Squeezy、Polar、Creem、Dodo Payments provider 接入 | 核验许可证、私有仓库、闭源修改、团队席位，并继续 mail-preview / marketing/docs spike |
| P0-005 | xwl_bi 分析数据面抽核方案 | 已完成 | 已确认 P1 新建独立业务无关仓库 `analytics-core`，不复用旧 Vue2 后台，不整仓改名 | 进入 P1-000 实施设计 |
| P1-000A | 输出 `analytics-core` 实施方案 | 已完成 | 已新增 `analytics-core实施方案.md`，并补充方案 B 物理分表、原生 ClickHouse batch writer、入库幂等去重、tenant/project/source 映射 | 根据评审继续细化接口和表模型 |
| P1-000 | 创建 `analytics-core` 独立核心仓库 | 已完成 | `src/analytics-core` 已初始化为独立 Git 仓库，远端为 `git@github-simpletrack:simpletrack/analytics-core.git`，并已挂载到父仓子模块 | 后续按独立仓库推进数据面实现 |
| P1-001 | EventBus 抽象设计 | 进行中 | 已落地 `EventEnvelope`、`EventBus`、`DirectBus`、`RedisStreamBus` 和 `KafkaBus` 包边界；已用 Redis Stack 验证 publish / consume / ack / pending=0 | 继续设计失败重试、死信队列和幂等入库边界 |
| P1-002 | 数据管道最小闭环 | 待完成 | P1 目标已确定 | 实现 tracker -> collect -> storage -> Realtime/Events |
| P1-003 | 产品官网 / Marketing Site / 公开站点 | 进行中 | 已从 `template-src/ai-supastarter-template` 初始化 `src/simpletrack-saas` 工作副本，远端为 `git@github-simpletrack:simpletrack/simpletrack-saas.git`；SaaS app 页面挂载和 subscription gate 草案已验证，marketing/docs 还未截图级验证 | 启动 Supastarter marketing/docs 验证 |
| INFRA-001 | SimpleTrack GitHub SSH 与子仓库推送配置 | 已完成 | 已生成并记录 `id_ed25519_simpletrack` 专用 key 流程，`src/analytics-core` 和 `src/simpletrack-saas` 固定使用 `config_simpletrack + core.sshCommand`，父仓已提交相关 Q&A 和 AGENTS 规则 | 后续新机器按 Q&A 复现；默认 SSH config ACL 可暂不阻塞主线 |

### 状态重置规则

- `已完成` 只表示当前验收口径下完成。
- 如果已完成任务发生功能重构、范围重开、验收失败、底座替换或实现被废弃，必须将状态重置为 `待完成`。
- 重置时必须同步更新“当前说明”和“下一步”，并在“修订记录”新增一条说明。

## 当前进度

当前处于 **P0：产品与底座确认**，并已经明确部分 **P1 前置底座任务**。

已经完成：

- 建立实施决策库。
- 建立 Q&A 目录。
- 建立支付服务说明目录。
- 初步确定 P1 范围。
- 初步确定 Next.js 主线和成熟 SaaS 模板优先路线。
- 确认 `analytics-core` 作为 P1 独立核心仓库建设方向。
- `src/analytics-core` 和 `src/simpletrack-saas` 已作为独立子仓库推送，并挂载到父仓子模块。
- SimpleTrack 专用 SSH key、`github-simpletrack` Host、`config_simpletrack` 和仓库级 `core.sshCommand` 规则已固化到 Q&A 与 `AGENTS.md`。
- `simpletrack/README.md` 已作为 SimpleTrack 资料入口提交。

正在推进：

- Supastarter for Next.js 的 1 天 SimpleTrack spike：已创建独立工作副本并推送远端，已完成 Websites、Realtime、Events 组织内页面挂载和 UI-only subscription gate；下一步验证 mail-preview、marketing/docs。
- `analytics-core` 模块边界、命名规范、队列抽象和依赖设计：已完成 Redis Stream publish / consume / ack 验证，下一步进入失败重试、死信队列和幂等入库边界。
- 企业分析控制台 UI 可改造性确认。
- 产品官网 / Marketing Site / docs 公开站点的信息架构。

下一步：

1. 在 `src/simpletrack-saas` 继续 Supastarter 1 天 spike：邮件模板、mail-preview、marketing/docs 预览。
2. 继续完善 `analytics-core` 的失败重试、死信队列和幂等入库边界。
3. 先使用 Supastarter 的 marketing/docs app 做产品官网和 docs；效果满足就直接使用，否则轻量定制。
4. 每次子仓库提交推送后，先提交子仓，再更新父仓 gitlink 和实施进度文档。

## 当前已确定的总方向

1. SimpleTrack 是一个正式项目，不再只做一次性评审图。
2. 前端原型继续按可交互、可跳转、可迁移到生产代码的方向推进。
3. 生产路线优先使用成熟框架、成熟组件库和成熟 SaaS 模板，不自造登录、支付、邮件、组织、后台、AI 基础设施。
4. 技术主线优先 Next.js，因为候选 SaaS 模板、独立工具生态和可生产化前端骨架都集中在 Next.js 上。
5. P1 的产品目标先收窄为“数据管道活了 + 公开产品入口”：页面浏览和自定义事件进入 Realtime 与 Events，同时具备产品官网、定价/订阅入口和 docs/quickstart。
6. 视觉风格采用低装饰、高密度、信息层次稳定的企业分析控制台风格。
7. 旧的 xwl_bi Vue2 后台界面不考虑复用。
8. P1 新建独立核心仓库 `analytics-core`，从 xwl_bi 抽取分析数据面核心；仓库名、包名、函数名、变量名不带 `simpletrack` 或 `xwl` 前缀。
9. `analytics-core` 前期用 Redis Stream 替代 Kafka 以降低运维复杂度，但保留 KafkaBus 作为后续高吞吐实现。
10. P1 包含产品官网 / Marketing Site / 公开站点，覆盖产品介绍、定价/订阅入口、docs/quickstart。
11. 生产 SaaS 模板先选择 Supastarter for Next.js；MakerKit 只保留为 B2B 企业控制面对照和备选。
12. 支付先按 Supastarter 已支持的 Stripe、Lemon Squeezy、Polar、Creem、Dodo Payments provider 接入；KYC/KYB、退款、拒付、发票、税务和费用结构放到上线前逐项处理。
13. `analytics-core` 参考 Umami 的分析对象体系和 Litlyx 的首价值、Raw Events、Show test data 经验。
14. `analytics-core` 的 ClickHouse 表策略直接采用方案 B，按 `tenant_id / project_id / source_id` 路由到物理事件表，上层仍使用统一 `events` 逻辑模型。
15. ClickHouse 事件写入热路径优先使用原生 batch writer，入库按 `(tenant_id, project_id, source_id, event_id)` 做幂等去重。

## 当前待评审的总方向

1. Supastarter 的 1 天 SimpleTrack spike 是否顺畅通过。
2. 支付 provider 的具体上线顺序和平台配置，放到上线前处理。
3. `analytics-core` 的表模型、接口分层、存储模型和从 xwl_bi 抽取的具体代码步骤。
4. 企业级控制台 UI 是否直接跟随 Supastarter UI 栈，还是在模板 shell 内重做业务控制台。
5. Supastarter 的 marketing/docs app 是否满足产品官网和 docs 需要；若不满足再轻量定制。

## 维护规则

- 确定了就写入 `已确定`。
- 还没有最终拍板但已经进入讨论，就写入 `待评审`。
- 被明确排除的方案，写入 `已否决`，并说明原因，避免后续重复评估。
- 每条决定都要能回答三个问题：为什么这样做、影响哪些模块、下一步怎么验证。
- 每次任务完成后，必须更新上方“实施计划完成列表”。
- 已完成任务如果被重构或重新打开，状态必须重置为 `待完成`。
