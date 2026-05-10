# SimpleTrack 实施决策库

> 目录用途：存放 SimpleTrack 已确定要实施的决定、分阶段实施计划，以及仍需评审的关键事项。后续每次确定新决定，都先更新本目录，再继续改原型或生产代码。

## 文档索引

| 文档 | 用途 | 状态 |
| --- | --- | --- |
| [修订记录.md](修订记录.md) | 记录实施决策库的完整修订流水，避免 README 过大 | 持续更新 |
| [分阶段实施计划.md](分阶段实施计划.md) | 记录 P0/P1/P2/P3 的目标、范围、交付物和验收标准 | 已确定，持续更新 |
| [待评审事项.md](待评审事项.md) | 记录还没有拍板的技术栈、模板、支付、数据面复用方案 | 待评审，持续更新 |
| [技术栈底座决策.md](技术栈底座决策.md) | 记录当前已经形成的技术底座方向和边界 | 已确定 + 待评审 |
| [付费SaaS模板本地对比.md](付费SaaS模板本地对比.md) | 基于 `template-src` 本地源码和 MakerKit 官方资料对比付费模板 | 已确定先选 Supastarter，持续更新 |
| [analytics-core实施方案.md](analytics-core实施方案.md) | 记录 `analytics-core` 的 P1 抽取边界、模块草案、EventBus 方案和 xwl_bi 代码评审结论 | 已确定，设计细节持续评审 |
| [SimpleTrack分析服务职责边界.md](SimpleTrack分析服务职责边界.md) | 说明 `simpletrack-saas`、`simpletrack-anaysitics-service` 与 `analytics-core` Go library 的控制面/数据面边界 | 已确定，持续更新 |
| [xwl_bi后端架构参考映射.md](xwl_bi后端架构参考映射.md) | 将 `references/xwl_bi-backend/` 的后端架构设计映射到 `analytics-core`，明确只参考架构不搬旧业务代码 | 已确定，持续更新 |
| [Umami源代码参考映射.md](Umami源代码参考映射.md) | 将 `references/umami/` 的官方源码快照映射到 SimpleTrack P1/P2/P3 实施边界，明确只参考对象体系和实现策略 | 已确定，持续更新 |

## 修订记录

完整修订流水已迁移到 [修订记录.md](修订记录.md)。README 只保留当前索引、实施计划状态、当前进度和维护规则。

## 实施计划完成列表

状态取值：`待完成`、`进行中`、`已完成`、`暂缓`、`已否决`。

| 编号 | 任务 | 状态 | 当前说明 | 下一步 |
| --- | --- | --- | --- | --- |
| PLAN-001 | 建立实施决策库目录 | 已完成 | 已创建 `simpletrack/docs/实施决策/` | 持续维护 |
| PLAN-002 | 形成分阶段实施计划 | 已完成 | 已明确 P0/P1/P2/P3，P1 收窄为“数据管道活了 + 公开产品入口” | 随评审更新阶段边界 |
| PLAN-003 | 建立 Q&A 目录 | 已完成 | 已创建概念解释和评审问答目录 | 新问题继续追加 |
| PLAN-004 | 建立支付服务说明目录 | 已完成 | 已创建 Stripe / Polar / Lemon Squeezy 对比，并明确先按 Supastarter provider 接入 | 上线收费前补 Paddle/Creem/Dodo 和个人开发者收款检查 |
| P0-001 | Next.js 可交互原型 | 进行中 | `simpletrack-enterprise-mvp` 的当前页面集合已收口到 `onboarding`、`dashboard`、`events`、`goals`、`settings`，并作为 P1 页面 contract 的最新依据；需继续按生产可迁移骨架推进 | 完成 Next.js 迁移和页面验证 |
| P0-002 | SaaS 模板选型评估 | 已完成 | 已确定先选择 Supastarter for Next.js；MakerKit 保留为 B2B 对照和备选 | 进入 Supastarter 1 天 SimpleTrack spike |
| P0-003 | 企业分析控制台 UI 可改造性评审 | 进行中 | `src/simpletrack-saas` 已能在 Supastarter `apps/saas` 组织内导航挂载 Websites、Realtime、Events 页面草案；`simpletrack-saas` `9ebe9f4` 把 Next dev loopback 登录链路收口为共享 helper，真实验证 `localhost:3005` / `127.0.0.1` 都能完成 authenticated 登录跳转；`ffdb254` 已把 Websites 页从“行内整块设置表单”重构为“紧凑来源行 + write key 面板 + 右侧动作 + accordion 设置区”，并补上“编辑成功/失败后重新展开目标来源”的上下文保持行为；`08b9acc` 又把本地 auth/mail runtime 收口为 development 默认不强制 email verification、未配置 mail provider 时仅在 development 走 console fallback；`2a33f00` 继续把 authenticated Realtime / Events / Goals 收口成更诚实的 P1 体验；`d48660e`、`aff924e` 和 `cb643e3` 分别把 Quickstart 真实跳转、三页 Quickstart 覆盖、三页 disabled-source fallback 覆盖纳入 Playwright；最新 `6a14cd9` 又用真实 server-side readback outage 验证 Realtime / Events / Goals 在 analytics-service 不可用时 fail-soft | 继续做 authenticated 页面截图级评审，并补更多错误态浏览器级回归 |
| P0-004 | Supastarter for Next.js 接入核验 | 进行中 | 已确定先选 Supastarter；`src/simpletrack-saas` 已作为独立子仓库推送；Websites 页已从 UI-only gate 前进到真实 Website source 列表 + 最小创建入口，并在 `ffdb254` 收口成企业控制台 accordion 行布局；marketing/docs/mail-preview 已完成浏览器截图验证；privacy/terms 已替换为 SimpleTrack 基线版本；authenticated readback 页面已完成 selector、空态、错误态、Quickstart、disabled-source fallback 和 service outage 的真实浏览器回归，其中最新 `simpletrack-saas` `6a14cd9` 通过未监听本地 analytics-service 端口验证三页 fail-soft | 继续核验许可证、私有仓库、闭源修改、团队席位，并补更多 docs/readback 错误态回归 |
| P0-005 | xwl_bi 分析数据面抽核方案 | 已完成 | 已确认 P1 新建独立业务无关仓库 `analytics-core`，不复用旧 Vue2 后台，不整仓改名 | 进入 P1-000 实施设计 |
| P1-000A | 输出 `analytics-core` 实施方案 | 已完成 | 已新增 `analytics-core实施方案.md`，并补充方案 B 物理分表、原生 ClickHouse batch writer、入库幂等去重、tenant/project/source 映射 | 根据评审继续细化接口和表模型 |
| P1-000 | 创建 `analytics-core` 独立核心仓库 | 已完成 | `src/analytics-core` 已初始化为独立 Git 仓库，远端为 `git@github-simpletrack:simpletrack/analytics-core.git`，并已挂载到父仓子模块 | 后续按独立仓库推进数据面实现 |
| P1-001 | EventBus 抽象设计 | 已完成 | 已落地 `EventEnvelope`、`EventBus`、`DirectBus`、`RedisStreamBus` 和 `KafkaBus` 包边界；Redis Stream 已支持 pending 优先重试、`MaxAttempts` 死信队列和消费成功后 ack；ingestion processor 已把重复事件写入视为成功处理 | 进入 P1-002，继续实现 collect、ClickHouse `EventWriter`、`TableRouter` 和 Realtime/Events 最小闭环 |
| P1-000B | 引入 xwl_bi 后端参考快照 | 已完成 | 已将本地 `xwl_bi` 后端代码和顶层关键文档复制到 `references/xwl_bi-backend/`，并明确为只读架构设计参考快照，不包含 Vue2 前端、日志和二进制 | 仅按需 refresh 快照；主要参考模块边界、启动装配、消费链路、ClickHouse 写入/查询分层和元数据流转，不直接在快照中开发 |
| P1-000C | 引入 Umami 官方源码参考快照 | 已完成 | 已将 Umami 官方 GitHub 源码克隆到 `references/umami/`，删除上游 `.git` 元数据并记录 commit；已新增源码审阅、实施映射、`simpletrack/docs/umami/docs/源码实现参考/` 分章节深解文档和 Q&A 概念解释 | 仅按需 refresh 快照；主要参考 tracker、collect、事件/会话模型、Realtime/Events 查询、ClickHouse schema，并用分章节文档和 Q&A 对照 SimpleTrack 与 `analytics-core`，不直接复制代码 |
| P1-002 | 数据管道最小闭环 | 已完成 | 已完成 collect 请求标准化、字段校验、属性入口约束、typed property row 逻辑展开、`EventPropertyWriter` 契约、ClickHouse `PropertyBatchWriter`、`PropertyIndexingEventWriter` 热路径组合、MySQL `property_indexing_status` guard、`collect.Handler`、Fiber `POST /collect` 适配入口、storage `EventWriter` 接口、ClickHouse `TableRouter`、native batch `BatchWriter`、`EventWriteGuard` 幂等边界、GORM/MySQL `IngestionStatusGuard`、`EventQueryBuilder` 查询边界、typed 属性过滤、`storage.EventReader` 查询执行器、`ingestion.Processor` worker 边界、本地 Redis/MySQL/ClickHouse compose、opt-in e2e 验收测试、Events 排序/过滤 typed 白名单，以及 P1-002B/C collect pre-queue stage；`visit_id` 已按长期方案进入事件契约、ClickHouse event / `_properties` 表、writer、reader 和 query builder，不再依赖 readback 临时派生 | 保持 e2e 回归；ambiguous `property_indexing_status=processing` 自动恢复、过滤统计、复杂身份合并和 ClickHouse 读侧优化放 P1.5/P2 |
| P1-002A | 事件属性与用户属性模型优化 | 已完成 | P1 范围已完成：collect 只接受有界数量、合法 key、标量值、有限数字和有限长度字符串；storage 提供 `EventPropertyRecord` 与 `FlattenEventProperties`；ClickHouse `PropertyBatchWriter` 写入同源路由 `_properties` 表；`PropertyIndexingEventWriter` 将属性索引接入 ingestion 热路径；MySQL `property_indexing_status` 独立防重复，failed 可原子 reclaim，processing 视为 ambiguous 不自动重试；nested object/array 暂不进入 P1 | 属性字典治理、ambiguous processing 运维恢复、ClickHouse 去重/物化视图/projection 放 P1.5/P2；P1 主线转入 P1-002B/C |
| P1-002B | client info enrich 与 bot/IP 过滤 stage | 已完成 | 第一版已落地：`collect.Stage` 在 EventBus publish 前执行，`ClientEnrichmentStage` 可补 UA、referrer 和盐化 `client.ip_hash`，`TrafficFilterStage` 可按 bot UA、internal CIDR/IP 过滤；Fiber adapter 默认只用连接端 IP，需显式 `WithTrustedProxyHeaders()` 才信任代理头；浏览器 SDK 已补 `data-do-not-track="true"`，DNT active 时不发送且不持久化 `distinct_id`，并自动收集 allowlisted UTM/click id；bot 与 internal IP filter 都已有审计回归，日志只记录 event/source 边界和过滤 reason，不记录原始 IP；Websites 设置区已将 Bot user agents、Internal CIDRs、Internal IPs 做成可见配置项，表单提交和 runtime source 回写已有回归覆盖；逻辑不进入 ClickHouse writer | 过滤统计、过滤原因报表和配置变更审计放 P1.5/P2；若这些进入 P1，需要按状态重置规则重开 |
| P1-002C | session/visit resolver 隐私友好识别 | 已完成 | `NewSessionResolverStage` 在缺失 `session_id` 时按 tenant/project/source/distinct_id/时间窗口生成盐化匿名 `ses_` 标识，可选把 transient UA/IP 仅作为 hash 输入；`NewVisitResolverStage` 已按 server-only visit salt、默认 30 分钟窗口和最终 `session_id` 派生缺失的 canonical `vis_` 标识；显式 SDK `visit_id` 会被保留；`visit_id` 已写入 ClickHouse event 表和 `_properties` 表，Realtime / Events 直接读取持久字段；`simpletrack-saas` runtime-source 会输出 server-only `visit_salt` 与 `visit_window_seconds`；原始 IP/UA 不写入事件契约或存储；浏览器 DNT opt-in 已避免 DNT active 时创建持久本地身份 | salt 轮换、cookie/no-cookie、server identity、Sessions 专页和 retention 产品化放 P1.5/P2；若这些进入 P1，需要按状态重置规则重开 |
| P1-002D | 查询白名单与过滤构建硬化 | 已完成 | 已有 `EventQueryBuilder` 边界；已新增 Events 类型化排序字段/方向白名单、过滤字段/operator 白名单、filter 数量上限、typed property filter allowlist、非法属性字段测试和 `ErrInvalidEventQuery` 错误分类；属性过滤采用 ClickHouse 可执行的 tuple `IN` 子查询并已通过真实 e2e | 后续新增 Breakdown/Funnel/Retention 查询时复用同一 allowlist 思路，并为复杂 ClickHouse SQL 补真实 e2e |
| P1-002E | Realtime/Events 最小端到端验收 | 已完成 | 已新增 `internal/e2e` opt-in 测试，使用本地 Redis/MySQL/ClickHouse 验证 collect -> Redis Stream -> ingestion -> ClickHouse -> Realtime/Events reader；测试覆盖 pageview、自定义事件属性、user properties、ClickHouse property writer、属性索引热路径和 allowlisted property filter 精确排除非匹配事件；冷启动依赖 readiness 已通过重试窗口修复 | 后续保持该 e2e 作为回归入口，并在 P1.5/P2 扩展属性治理、聚合表和复杂查询场景 |
| P1-003 | 产品官网 / Marketing Site / 公开站点 | 已完成 | 已从 `template-src/ai-supastarter-template` 初始化 `src/simpletrack-saas` 工作副本；marketing 文案、pricing 语义、docs/quickstart、mail-preview 品牌文案和截图级验证已完成；privacy/terms 已替换模板占位；公开站点首屏已露出下一节内容 | 后续只做轻量文案和视觉微调，不阻塞 P1 数据管道 |
| P1-004 | Web tracker SDK 最短链路 | 已完成 | P1 浏览器 SDK 已从 `analytics-core` 迁出，由 `src/analytics-service/public/tracker.js` 作为运行时服务静态资产托管；SDK 继续支持 auto pageview、SPA route pageview、manual track、identify、debug、snippet queue、`localStorage` fallback、非法 event name 拦截、opt-in DNT 和 allowlisted UTM/click id；docs/quickstart 已改为 `data-write-key` 接入，`simpletrack-anaysitics-service` 通过 write key 覆盖 tenant/project/source | React/Next/Node/mobile SDK、多语言 SDK、performance metrics 和 CDN/版本化发布策略放后续阶段评审 |
| P1-005 | SimpleTrack Anaysitics Service 本地仓库 | 已完成 | 已创建并推送 `src/analytics-service` 远端仓库 `simpletrack/anaysitics-service` 并登记父仓子模块，服务名 `simpletrack-anaysitics-service`；当前提供 `/healthz`、`/tracker.js`、`OPTIONS /collect`、`POST /collect`、内部 `/v1/realtime`、内部 `/v1/events`，并用 `MemoryResolver` 或 SaaS HTTP resolver 执行 runtime source config；`simpletrack-saas` 内部 runtime-source API、Websites 控制面 CRUD、Realtime/Events 页面服务端读回放已落地；内部 query token 已支持短窗口轮换 allowlist、结构化生效/过期时间、route scopes 和审计日志，Events 属性过滤已接入 `allowed_property_filters`；P1-005B/C/D 均已完成，复杂聚合分析与 Breakdown/Compare 维持到 P1.5/P2 | 保持现有回归；更复杂的聚合查询与高级分析放 P1.5/P2 |
| P1-005A | `analytics-core` root-level Go library API | 已完成 | `analytics-core` 已调整为可被 Go 服务引用的根目录公共包：`contracts`、`collect`、`eventbus`、`ingestion`、`storage` 等；不再把 Browser SDK 放进 core，也不把 core 作为长期业务服务运行 | 后续公共 API 变更需保持外部服务 import 稳定 |
| P1-005B | collect runtime service | 已完成 | `simpletrack-anaysitics-service` 已实现 write key 解析、source enabled、Origin allowlist、Fiber CORS middleware、客户端 scope 覆盖、bot/internal traffic 过滤、client enrich、session resolver、Redis durable enqueue，以及显式开启的同进程 ingestion worker；HTTP runtime 已迁移为 Fiber app，`collectapi.Handler` 保持业务依赖持有边界，`main.go` 只负责生命周期；worker 复用 `analytics-core` 的 Redis Stream subscribe、MySQL guard、ClickHouse native writer 和 typed property indexing；write key 不再作为 privacy salt fallback；ingestion 启动默认校验启用 source 的 ClickHouse event/property 表存在，也可用 `ANALYTICS_SERVICE_CLICKHOUSE_AUTO_MIGRATE=true` 在本地/小部署创建当前 runtime config 内所有启用 source 的 routed tables 后再校验；README 本地运行示例已对齐当前 SourceConfig，包含 `visit_salt` 和 `visit_window_seconds` | 保持 collect/runtime/worker 回归；部署级 schema migration/rollback 规范留到真实上线后再评审 |
| P1-005C | SaaS control-plane config runtime enforcement | 已完成 | 已新增 HTTP source resolver：`simpletrack-anaysitics-service` 可用 `ANALYTICS_SERVICE_SOURCE_RESOLVER=http`、控制面 URL 和 bearer token 按 write key 读取 runtime source config，并用短 TTL 缓存 + `ETag` 条件重验证降低热路径控制面压力且避免 stale auth state；控制面读取默认要求 HTTPS，只有本地 loopback 明确 opt-in 才允许 HTTP；同进程 ingestion 会把 HTTP 返回的 source 绑定到启动 schema surface；`simpletrack-saas` 内部 runtime-source API 已落地，runtime source config 必须携带 server-only `session_salt`、`visit_salt`、`visit_window_seconds` 和 `client_hash_salt`，不能由公开 write key 派生；`Websites` 页现已使用真实组织 Website 数据并提供最小 source 创建、编辑、enable/disable 和 delete 入口，编辑动作可回写 `allowedOrigins`、`bot user agents`、`internal CIDRs`、`internal IPs` 与 `includeClientFingerprint`，并已将这些 runtime source 字段做成可见设置项；disabled source 不再占用 active source limit，创建与启用动作都通过 serializable 事务 helper + `P2034` 重试守住 active source limit；P1 quota 策略已明确只执行 active website source 上限，事件量、留存期、成员数和用量计费不提前进入 P1；`analytics-service` 现在也有 handler 级回归覆盖，验证 cached source 在控制面 disable/delete 后会立即拒绝 collect/readback | salt 轮换、runtime source 变更审计和更细的配置历史放 P1.5/P2 |
| P1-005D | Events / Realtime 查询 API | 已完成 | `simpletrack-anaysitics-service` 已提供内部 `/v1/realtime` 与 `/v1/events` 读接口，使用 bearer token 保护并映射到 `analytics-core` 的 `EventReader`；两个默认路径保留，同时新增 `ANALYTICS_SERVICE_EVENTS_PATH` 和 `ANALYTICS_SERVICE_REALTIME_PATH` 可配置路由；Swagger UI / OpenAPI 文件已随 Fiber runtime 接入；`simpletrack-saas` Realtime / Events 页面已用 server-side readback helper 和 client-safe Website selector 接入，按组织查启用 Website 后使用 write key 调内部服务，未配置、无启用 source 或服务错误时 fail-soft；Events 现在已补 `event_name`、`distinct_id`、`visit_id`、`limit`、`offset`、`sort_field`、`sort_direction` 白名单、`hasMore` 分页和 `30m / 6h / 24h / 7d` 时间窗口预设；query token 轮换已通过 `ANALYTICS_SERVICE_QUERY_TOKENS_JSON` 支持当前 token + 上一版 token 并行接受，并可附带 `id`、`not_before`、`expires_at` 元数据；运行时会拒绝过期/未来 token，并对轮换命中和拒绝场景打审计日志；属性过滤已由 source runtime config 的 `allowed_property_filters` 提供白名单来源，并映射到 `analytics-core` typed property filters；读侧现在使用持久化 `visit_id`，不再用 `session_id` 兜底成 visit；OpenAPI 已把 `property_filter` 标为 repeatable array，并补齐 `event_name` 排序枚举；当前已补组合查询回归，覆盖 SaaS 请求序列化、service query mapping 和 core ClickHouse query plan 三层的标量过滤、`visit_id`、分页、排序、时间窗口和 repeatable property filters 同时存在 | 更复杂的聚合分析、Breakdown/Compare/Funnels/Journeys 维持到 P1.5/P2 |
| P1-005E | Goal 最小闭环读回放 | 已完成 | 已完成 `analytics-core` 事件 count 读侧契约、`simpletrack-anaysitics-service` 内部 `/v1/goals`、OpenAPI、query route 配置和测试；`simpletrack-saas` 已新增 Goal 数据模型、组织内 Goals 页面、事件名校验、按 Website 列表和 24h count readback；Goal 创建会把并发唯一约束冲突映射为 `duplicate_event`；Goals readback UI 已把 true no-data 和 service failure 拆成不同状态；事件名契约复制风险已用 Go/TS 两边一致的 accepted/rejected 样例测试锁住；`9c56e96` 把 Goal readback fan-out 收口到 25；`d48660e`、`aff924e`、`cb643e3` 和最新 `6a14cd9` 继续把 Goals empty state、Quickstart、disabled-source fallback 与 analytics-service outage 串进真实 Playwright 回归；版本标记：`analytics-core` `1ea78f3`，`analytics-service` `3b8d27c`，`simpletrack-saas` `6a14cd9`，父仓同步 commit `pending commit` | 后续关注更多 docs/readback 错误态回归，以及更大规模 Goal 批量 count 放 P1.5/P2 |
| P1.5-001 | ClickHouse 读侧优化与属性治理 | 进行中 | 已完成属性治理和 query plan 约束雏形：`readSidePolicy`、`EventQueryEvidence`、source-scoped `PropertyCatalog`、内部 `/v1/properties`、SaaS Events filter builder 属性建议、真实 ClickHouse EventReader/BatchWriter benchmark、Redis Stream benchmark、`collect.Handler` benchmark、opt-in ClickHouse explain，以及 `QueryEvidence()` 的 `PropertyFilters` 快照语义；500k 行复测已把 Realtime 拆为短窗口 `low_realtime_recent_window` 和宽时间窗 `low_realtime_wide_since`，又把 Events 拆为 recent-window 与 wide-window scalar/property 场景；`caf314d` 已补真实 `EventQueryPlan` 时间上下界和 bound args 断言，`7c31eb2` 已把该口径同步到子仓 README 和 read-side policy；`f84024a` 又把 typed property filters 收口到 query-builder guardrail：必须显式带 `from/to`，且 direct fact-table 窗口默认不超过 7 天；`a99147f` 继续把 bounded scalar benchmark / explain 扩成 24h / 72h / 7d 多窗口套件，并补到真实 500k / 1,000,000 行证据：`24h` 与 `72h` 仍在中等观察区，`7d` 才第一次进入 `46-52ms/op` 压力区；因此 `analytics-service` `c08e1da` 撤回 `24h => high` 的 bounded scalar triage 规则仍然成立。当前仍以 direct fact table 为默认，不新增 projection、materialized view 或小时聚合表 | 继续按策略文档观察宽时间窗 scalar Events 明细查询和更大 row-volume 下的 bounded scalar 形状；typed property 过滤的 7 天护栏已在 core 与 service 双层收口，如果要支持更宽的 property 历史窗口，必须先补新的 query evidence、benchmark、explain 和物理结构评审 |
| P2-001 | Performance metrics 采集与查询 | 暂缓 | Umami tracker 可采集 LCP、INP、CLS、FCP、TTFB；SimpleTrack P1 不以性能诊断为阻塞项 | P2 评审是否作为事件类型、属性组或独立 performance 模型进入 `analytics-core` |
| INFRA-001 | SimpleTrack GitHub SSH 与子仓库推送配置 | 已完成 | 已生成并记录 `id_ed25519_simpletrack` 专用 key 流程；用户交互式 PowerShell 已验证 `ssh -F config_simpletrack` 可认证到 `RoggerZ`；`analytics-core` `a99147f`、`analytics-service` `580c547`、`simpletrack-saas` `6a14cd9` 均已通过 `github-simpletrack` Host 与仓库级 `core.sshCommand` 推送到远端；父仓本轮 gitlink 与文档同步待本次提交回填 | 后续新建子仓或换机器时继续复用 `docs/Q&A/Windows-SSH仓库权限怎么配置.md` 和 `AGENTS.md` 中的专用 SSH 配置流程 |

P1.5-001 补充进度：2026-05-08 的 500k 行复测已纠正 Realtime benchmark 口径，`low_realtime_recent_window` 代表产品短窗口 Realtime，`low_realtime_wide_since` 代表宽时间窗压力查询；2026-05-09 的 `caf314d` 进一步把 Events benchmark / explain 拆成 recent-window 与 wide-window scalar/property 场景，并在执行前断言真实 `EventQueryPlan` 的时间上下界。随后 `analytics-core` commit `f84024a` 又把 typed property filters 收口到 query-builder guardrail：必须显式带 `from/to`，且 direct fact-table 查询窗口默认不超过 7 天；`analytics-core` commit `a99147f` 再把 bounded scalar benchmark / explain 扩成 24h / 72h / 7d 多窗口套件：`24h` 与 `72h` 在 100k / 500k 行夹具下仍接近 direct fact-table 的中等观察区，`7d` 在 1,000,000 行夹具下才第一次进入 `46-52ms/op` 压力区。因此 `analytics-service` commit `c08e1da` 撤回 `24h => high` 的 bounded scalar triage 规则仍然成立，而后续如果要重引入 bounded scalar heuristic，也必须把时间窗和 row volume 一起纳入判断。当前观察候选仍集中在“宽时间窗 scalar Events 明细查询”；如果要支持更宽的 property 历史窗口，先做方案评审而不是直接放开。当前仍不新增 ClickHouse projection、materialized view 或小时聚合表。

### 状态重置规则

- `已完成` 只表示当前验收口径下完成。
- 如果已完成任务发生功能重构、范围重开、验收失败、底座替换或实现被废弃，必须将状态重置为 `待完成`。
- 重置时必须同步更新“当前说明”和“下一步”，并在“修订记录”新增一条说明。

## 当前进度

本节版本标记：`analytics-core` commit `a99147f`；`analytics-service` commit `580c547`；`simpletrack-saas` commit `6a14cd9`；父仓同步 commit `pending commit`。

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
- `analytics-core` 已完成 EventBus 抽象、Redis Stream pending 优先重试、死信队列和幂等 ingestion processor。
- `references/xwl_bi-backend/` 已加入为只读临时参考快照，供 `analytics-core` 实现映射时查阅。
- `references/umami/` 已加入为 Umami 官方源码只读参考快照，供 tracker、collect、事件/会话模型、Realtime/Events 查询和 ClickHouse schema 对照。
- `analytics-core` 已完成 collect 请求标准化、storage `EventWriter` 接口和 ClickHouse `TableRouter` 契约。
- `analytics-core` 已完成 `collect.Handler` 和 Fiber `POST /collect` 入口，能把 JSON 请求转换为 `EventEnvelope` 并发布到 EventBus。
- `analytics-core` 已完成 ClickHouse native batch `BatchWriter`，使用 `clickhouse-go/v2 PrepareBatch` 通过 `EventWriter` 接口写入动态物理事件表，并预留 `EventWriteGuard` 幂等边界。
- `analytics-core` 已完成 GORM/MySQL `IngestionStatusGuard`，通过 `ingestion_status` 表对 `(tenant_id, project_id, source_id, event_id)` 做 `processing / inserted / failed` 状态占用、提交、失败回滚和重复写入跳过。
- `analytics-core` 已完成 `storage.EventQueryBuilder` 契约和 ClickHouse/GORM query plan builder，Events 与 Realtime 查询共用同一套字段白名单、表路由、时间范围和分页限制。
- `analytics-core` 已完成 `storage.EventReader` 契约和 ClickHouse/GORM 查询执行器，执行 query plan 后返回 storage-neutral `EventRecord`。
- `analytics-core` 已明确 `ingestion.Processor` 是 P1 worker 边界，EventBus adapter 负责 ack/nack，Processor 只把消息写入 `storage.EventWriter` 并把错误交回队列重试/死信策略。
- `analytics-core` 已新增本地 `docker-compose.yml`，包含 Redis Stack、MySQL 8.4、ClickHouse 25.3，并在 README 记录启动、避开冲突的高位端口和 Redis Stream 集成测试入口。
- `analytics-core` 已新增 opt-in e2e 测试，真实跑通 collect -> Redis Stream -> ingestion -> ClickHouse -> Realtime/Events reader；验证 pageview、自定义事件属性和 user properties 可被读侧查出，并已补齐 Redis/MySQL/ClickHouse 冷启动 readiness 重试。
- `analytics-core` 已从“可能自带服务/SDK”的实现形态纠偏为 Go 第三方库：外部服务通过根目录公共包引用 core，Browser SDK 由 `simpletrack-anaysitics-service` 托管。
- `src/analytics-service` 已作为本地 Go 仓库创建，服务展示名为 `simpletrack-anaysitics-service`，负责 SimpleTrack 分析数据面的 runtime enforcement：write key、Origin、CORS、internal traffic、bot 过滤、collect 调用 core，以及可选同进程 ingestion worker 装配。
- `src/analytics-service` 已推送远端 `simpletrack/anaysitics-service`，并开始登记父仓子模块；仓库级 `core.sshCommand` 已对齐 `config_simpletrack`。
- `src/analytics-service` 已补 `P1-005D/E` 内部读回放入口：`/v1/realtime`、`/v1/events` 和 `/v1/goals` 复用 `analytics-core` 的 `EventReader` / count 读侧契约，通过内部 bearer token 保护，并支持 Fiber CORS preflight；Events 侧已补事件名、distinct id、排序、分页和 typed property filters 白名单；Goal 侧按 write key 解析 source 后做精确事件 count；内部 query token 已支持短窗口轮换 allowlist、结构化生效/过期时间、按 `realtime / events / properties / goals` 拆分的 route scopes，以及命中/拒绝审计日志；HTTP resolver revalidation 现在还有 handler 级回归，验证控制面 disable/delete 后 collect/readback 立即撤权；现在 readback 还会按 runtime-source `readback_policy` fail closed，Swagger UI / OpenAPI 文件已接入；`6f6f742` 先把 typed property filter 的 7 天窗口护栏锁到 service 边界回归，`c08e1da` 又根据 100k / 500k bounded 24h scalar 证据撤回了 `24h => high` 的 bounded scalar triage 规则，而 `580c547` 再把 query token 的路由粒度权限前移到 source resolution 之前。
- `src/simpletrack-saas` 已补 Realtime / Events / Goals 页面服务端读回放 helper 和 client-safe Website 选择器初版：内部 query token 不下发到浏览器，页面按当前组织启用 Website 的 write key 调 `simpletrack-anaysitics-service`，并覆盖未配置服务、无启用 source、禁用 source 和服务异常空态；Events 页面还支持白名单分页、上一页 / 下一页、`30m / 6h / 24h / 7d` 时间窗口预设，以及重复 query 参数/空页偏移的服务端硬化，不会因重复参数把 RSC 页面打崩；Goals 页面可创建一个关键事件目标，并通过 24h `/v1/goals` readback 显示是否已有匹配事件；`simpletrack-saas` 的 runtime-source `readback_policy` 契约已经在 `9c8ba37` 中稳定存在，`9ebe9f4` 把 Next dev `allowedDevOrigins` 与 Better Auth 本地 loopback trusted origins 收口到共享 helper，`b1db9d5` 又让 server-side readback helper 按 `realtime / events / properties / goals` 优先选择 route-scoped query token，`08b9acc` 继续把本地 auth/mail runtime 收口为 development-only verification policy + development-only console mail fallback，`2a33f00` 把 authenticated Realtime / Events / Goals 在 no-source 与 disabled-requested-source 场景下收口为更诚实的 P1 产品态，`075828d` 又进一步补齐了 `Websites + Quickstart` 双动作和统一的 `buildWebsiteSettingsHref()` 深链 helper，`d48660e` 再把 signup -> sign-in -> onboarding -> org create -> Websites / Realtime / Events / Goals empty state 串成真实 Playwright 回归，并验证 Quickstart CTA 会真实跳到独立 docs app `/quickstart`；`aff924e` 又把这条真实点击验证从 Realtime 扩展到 Events 和 Goals；最新 `cb643e3` 再把 disabled-source fallback 从 Realtime 扩展到 Events 和 Goals，三个分析页现在都锁住了 notice + `edit_target` 深链契约；真实验证 sign-up / sign-in 已可直接打到 authenticated `/onboarding`，不再被 `EMAIL_NOT_VERIFIED` 阻断。
- `simpletrack-saas` `d53de11` 已把 marketing 公开站点的 `privacy-policy` 和 `terms` 从 Supastarter 模板占位替换为 SimpleTrack 基线文案；本轮真实外部审查以 DeepSeek + Codex 子代理为准，Gemini/Claude CLI 失败原因已记录到 `docs/开发环境卡壳问题记录.md`。
- Umami 源码深解已经转化为 `analytics-core` 优化计划：事件属性与用户属性模型、client info enrich、bot/IP 过滤、session/visit resolver、查询白名单、Realtime/Events 验收、Web tracker SDK、ClickHouse 读侧优化和 performance metrics 均已进入计划表或评审表。
- `simpletrack-enterprise-mvp` 的页面 contract 已从旧 `simpletrack-umami-inspired` 收口到当前原型页集合，`team` / `funnels` / `insights` 不再作为 P1 页面前置假设。

- `analytics-core` / `simpletrack-anaysitics-service` 的 `visit_id` 已从 readback 临时派生升级为写入前确定、入库存储、读侧直接读取；`readback_policy` 控制面闭环也已完成；后续只评审 salt 轮换、cookie/no-cookie、server identity、Sessions 专页、内部读权限粒度和 retention 产品化。

正在推进：

- Supastarter for Next.js 的 1 天 SimpleTrack spike：已创建独立工作副本并推送远端，已完成 Websites、Realtime、Events、Goals 组织内页面挂载；其中 Websites 已从 UI-only gate 前进到真实 source 列表 + 最小创建表单，并在 `simpletrack-saas` `ffdb254` 进一步重构为“紧凑来源行 + write key 面板 + accordion 设置区”的企业控制台布局，同时补上 settings server action 成功/失败后的上下文保持。`b1db9d5` 把 readback helper 收口到 route-scoped query token 选择逻辑，优先匹配 `realtime / events / properties / goals` 四类 token，再回退到 shared token；`08b9acc` 把本地 authenticated review 从“必须依赖 live mail verification”改成 development 可直接闭环；`2a33f00` 把 authenticated Realtime / Events / Goals 收口为 honest P1 state：无启用 source 统一 setup empty state、Realtime/Events 不再做无效 readback、disabled requested source 会回退到 enabled source 并显式提示；`075828d` 又把 empty state 收口为 `Websites + Quickstart` 双动作，并通过 `buildWebsiteSettingsHref()` 统一 disabled-source 深链；`d48660e` 再把这条链路钉进真实 Playwright e2e，覆盖 signup、login、onboarding、org create、Websites create、disabled-source fallback、`edit_target` 展开，以及 Quickstart -> 独立 docs `/quickstart` 的真实跨 origin 跳转；`aff924e` 又把 setup empty state 的 Quickstart 真实点击覆盖扩展到 Realtime / Events / Goals 三页；最新 `cb643e3` 再把 disabled-source fallback 的浏览器覆盖扩展到 Realtime / Events / Goals 三页。目标 `oxfmt --check`、Vitest（142 tests）、`saas` type-check、Playwright 正常运行和 `CI=1` 运行均已通过；相关子代理均已关闭。
- `analytics-core` P1 数据管道已收口：collect handler、Fiber `POST /collect` 适配器、属性入口约束、typed property row 逻辑展开、ClickHouse property batch writer、`PropertyIndexingEventWriter` 热路径组合、MySQL `property_indexing_status` guard、表路由契约、ClickHouse native batch writer、GORM/MySQL ingestion status guard、Realtime/Events query builder、typed property filter、ClickHouse query reader、worker 边界、本地运行依赖、最小端到端验证、Events 排序/过滤白名单、P1-002B/C collect pre-queue stage 和持久化 `visit_id` 链路已完成；browser / OS / device 派生、geo enrichment、internal traffic 产品配置和过滤审计边界也已补齐，core 继续保持 Go library 方式被 `simpletrack-anaysitics-service` 引用。
- `simpletrack-anaysitics-service` 主线（本地仓库 `src/analytics-service`）：已完成本地仓库、服务骨架、Fiber runtime app、memory / HTTP runtime config resolver、`/collect` 运行时校验、`/tracker.js` 静态托管、collect 单测、Redis durable enqueue、可选 ingestion worker 装配、本地/小部署 ClickHouse routed table auto migration，以及 `P1-005D/E` 内部 Events / Realtime / Goals 查询入口、query routes 配置、Swagger UI / OpenAPI 文件、query token 轮换 allowlist、结构化生命周期、token route scopes、source-scoped 属性过滤白名单、`visit_id` 持久字段读取、control-plane revoke handler 回归和内部 `/v1/properties` 属性目录读回接口；`6f6f742` 先把 typed property filter 的 7 天窗口护栏锁到 service boundary regression，用真实 `analytics-core` query builder 校验 `/v1/events` 的 `168h` success 和 `192h` rejection；`c08e1da` 又根据 100k / 500k bounded 24h scalar Events 证据撤回了 `24h => high` 的 service-side triage 规则，让 bucket 再次回到 filter-count 主导；最新 `580c547` 再把 query token 的 readback route scope 判定前移到 source resolution 之前，使 token 除了生效期和 source 边界外，还具备 `realtime / events / properties / goals` 级别的第二道权限层，不依赖本地 ClickHouse 也能通过 handler 回归钉住行为；`simpletrack-saas` 内部 runtime-source API、Websites 真实 source 管理最小闭环（create/list/update/enable/disable/delete）、visit resolver runtime 配置、Realtime/Events/Goals 页面读回放、visit_id 可见列、页面级回归、client-safe Website selector、Events 时间窗口预设、repeatable property_filter 多条件查询、Goal event-name contract samples、Goal fan-out 25 条上限、P1 active-source quota 策略和 `readback_policy` 显式契约已落地，复杂聚合分析维持到 P1.5/P2。父仓文档版本以本页顶部版本标记为准。
- `analytics-core`、`simpletrack-anaysitics-service` 和 `simpletrack-saas` 近期又补了 Events 组合查询回归：同一条请求里同时携带 `event_name`、`distinct_id`、`visit_id`、分页、排序、时间窗口和 repeatable `property_filter` 时，SaaS 会保持完整 readback state，服务端会把同一组条件送入 `analytics-core`，core 会生成参数化 ClickHouse query plan。
- `analytics-core` 的 ClickHouse 读侧优化已经从“预研名词”细化成可执行分层路线：属性治理和 query plan 约束先做，热点明细路径再择机引入 projection，稳定指标再落到 materialized view 或小时聚合表；`f84024a` 进一步把 typed property filters 收口到 query-builder guardrail，要求显式 `from/to` 且 direct fact-table 历史窗口默认不超过 7 天，`93cff0f` 又把 bounded 24h scalar Events benchmark / explain 收口成“只有真实 distinct 时才纳入证据”。
- `analytics-core` 的 ClickHouse 读侧长期规范已写入实施方案、`AGENTS.md` 和 `src/analytics-core/docs/read-side-optimization-policy.md`：后续读侧实现必须保持 `EventQueryBuilder` / `EventReader` / `TableRouter` 边界，不允许把 ClickHouse SQL 或物理表名扩散到 service、handler 或产品页面。
- `analytics-core` 已开始 P1.5-001 第一段实现：ClickHouse query builder 内部新增 `readSidePolicy`，把 query limit、filter cap、typed property filter bounded window 和 property allowlist 作为 adapter-owned guardrails 管理；`pressure` 已被明确成 read-side triage 桶，并在 Q&A 里说明不代表 SLA；当前已同时具备 builder-only query shape benchmark、真实 ClickHouse EventReader benchmark、真实 ClickHouse BatchWriter benchmark、GORM `CreateInBatches` 对照 benchmark、Redis Stream publish / subscribe+ack benchmark 和 `collect.Handler` 热路径 benchmark 基线，并确认 ClickHouse 事件热路径继续优先 native `PrepareBatch`；属性目录治理已从 ingestion upsert 补到 source-scoped reader，并已接入 SaaS Events filter builder 作为字段建议来源。
- `analytics-core` 已开始输出结构化 query plan evidence：计划会记录 query family、read path、optimization、scalar/property filter 数量、是否使用属性表和排序证据，后续是否上 projection / MV / 小时聚合表必须基于这类证据继续判断。
- `analytics-core` 已把 ClickHouse explain 证据纳入仓库级验证：`TestEventReaderClickHouseExplain` 复用 reader benchmark 的 routed table fixture 和 sealed query plan，当前 high property 场景已观测到 value-free property filter shape、`CreatingSets`、3 个 `event_id in ... set` 和包含 `visit_id` 的主键条件；500k 行复测已把 Realtime 拆成短窗口 `low_realtime_recent_window` 和宽时间窗 `low_realtime_wide_since`，也把 Events 拆成 recent-window 与 wide-window scalar/property 场景；最新 `93cff0f` 又保证 bounded 24h scalar Events 只在 fixture 真大于一天时进入 benchmark / explain，避免默认 10k 基线制造伪分支。当前确认产品短窗口 Realtime 和近期 Events 不构成当前读侧压力，后续重点观察宽时间窗 Events 和 typed property 过滤。
- `analytics-core` 的 query plan evidence 已补齐 effective limit、offset、time lower/upper bound、bounded time window 和 value-free property filter shape；`simpletrack-anaysitics-service` 已把这些字段随 `query_evidence` 透出到内部 readback API 和 OpenAPI，避免后续从 SQL 或 URL 参数反推优化依据，同时不回显属性过滤值；`QueryEvidence()` 已补快照语义，避免调用方修改返回 slice 后污染计划内证据；`analytics-service` `6f6f742` 把 core 的 7 天 typed property window 护栏锁到 handler 回归，`c08e1da` 则根据 100k / 500k bounded scalar 证据撤回了 `24h => high` 的 service-side 时间窗 heuristic，避免用不足够重的 bounded 窗口误报高压力。
- `docs/analytics-source-reading/read-side-benchmark-baseline.md` 已同步 `f84024a` 之后的读侧基线结论：typed property filters 默认只允许显式 `from/to` 且 direct fact-table 窗口不超过 7 天；后续观察候选收窄为宽时间窗 scalar Events 明细查询和 7 天内 typed property 过滤读路径。
- `docs/analytics-source-reading/collectapi-query-and-tracker-flow.md`、`data-flow-analysis.md`、`interfaces-and-formats.md` 和 `read-side-benchmark-baseline.md` 已统一补齐同一条 read-side 口径：typed property filters 必须显式 `from/to`，direct fact-table 默认只允许 7 天内窗口，`query_evidence` 只回显 value-free 形状。
- `simpletrack/prototype/simpletrack-enterprise-mvp/` 继续作为 P0/P1 评审原型，后续围绕当前页面集合补齐更清晰的 contract / mock / production 映射。
- `xwl_bi` 后端只读临时快照已就位，主要用于参考后端架构设计：模块边界、启动装配、消费链路、ClickHouse 写入/查询分层、元数据流转和分析服务拆分。
- Umami 官方源码只读快照已就位，主要用于参考分析对象体系、tracker 采集、事件属性、Realtime/Events 读侧、ClickHouse 明细与聚合模型；P1 数据管道源码分章节深解和 Q&A 概念解释已落地到 `simpletrack/docs/umami/docs/源码实现参考/`。
- Umami 源码启发的 `analytics-core` 优化项已排期并部分落地：P1 已补属性入库/查询、client enrich 第一版、session resolver 第一版、查询安全、端到端验收、浏览器 SDK 最短链路、DNT opt-in 和 UTM/click id 白名单；P1.5/P2 再评审 ClickHouse 聚合优化、多语言 SDK、visit 扩展、SDK CDN 发布和 performance metrics。
- 企业分析控制台 UI 可改造性评审继续推进，当前已确认 Supastarter shell 能承载 Websites、Realtime、Events 页面骨架。
- 产品官网 / Marketing Site / docs 公开站点的信息架构已按 P1 验收完成，privacy/terms 法务基线页也已从模板占位收口为 SimpleTrack 版本；后续只做轻量优化。

下一步：

1. 继续推进 P1.5-001：围绕宽时间窗 scalar Events 明细查询继续做稳定 query pattern 和回归计划评审；typed property 过滤的 7 天窗口护栏已在 core query-builder 与 analytics-service handler 两层收口，而原先 bounded `24h` 的 `pressure=high` heuristic 已被 `c08e1da` 撤回。当前 bounded scalar 已经有 `24h -> 72h -> 7d` 的证据梯度；下一步继续评审是否需要把 row-volume + time-window 作为更长期的 triage 输入，或者继续长期保持不做 bounded scalar 特判。`docs/analytics-source-reading/read-side-benchmark-baseline.md` 已同步这条结论。
2. 收口 Goal 最小闭环剩余 `WATCH` 项：事件名契约复制已由 `analytics-core` `1ea78f3` 与 `simpletrack-saas` `9c56e96` 的一致样例测试和 fan-out 上限缓解；readback policy 控制面闭环已完成，而 `analytics-service` `580c547` 又把内部 query token 粒度推进到 readback route scopes；下一步继续评审 live browser / e2e 缺口，以及是否还需要比 route scope 更细的内部读权限；更大规模批量 Goal count 放 P1.5/P2。
4. 把过滤统计、salt 轮换、cookie/no-cookie、server identity、Sessions 专页和 retention 产品化维持在 P1.5/P2，除非明确重开范围。
5. 把 R3-U1/R3-U2 的剩余项降为 P1.5/P2：ambiguous `property_indexing_status=processing` 恢复策略、ClickHouse 读侧优化中的 projection / materialized view / 小时聚合表落地。
5. 在需要 authenticated SaaS 流程时，用 `src/simpletrack-saas/docker-compose.yml` 启动本地 PostgreSQL，验证登录、组织和真实 subscription gate 依赖。
6. 公开站点继续使用 Supastarter 的 marketing/docs app，后续只做轻量文案和视觉微调。
7. 每次子仓库提交推送后，先提交子仓，再更新父仓 gitlink 和实施进度文档；`src/analytics-service` 现已具备远端和子模块登记条件，后续沿用与 `analytics-core`、`simpletrack-saas` 相同的收口顺序。

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
14. Umami 官方源码只作为只读实现参考：可参考 tracker、collect、事件/会话模型、Realtime/Events 查询和 ClickHouse schema，不直接复制 Umami 源码或采用其 Next.js API route 热路径架构。
15. `analytics-core` 的 ClickHouse 表策略直接采用方案 B，按 `tenant_id / project_id / source_id` 路由到物理事件表，上层仍使用统一 `events` 逻辑模型。
16. ClickHouse 事件写入热路径优先使用原生 batch writer，入库按 `(tenant_id, project_id, source_id, event_id)` 做幂等去重。
17. Umami 源码启发进入 `analytics-core` 的方式是“落到既有边界”：属性模型进入 `EventWriter` / `EventQueryBuilder`，session/visit 与 client enrich 进入 collect/ingestion stage，Realtime/Events 验收进入 `EventReader`，ClickHouse 读侧优化进入 P1.5/P2 的长期分层路线，先做属性治理和 query plan 约束，再按稳定查询引入 projection / materialized view / 小时聚合表。
18. `analytics-core` 是 Go 第三方库，不是 SimpleTrack 业务服务；`simpletrack-anaysitics-service` 才是运行时数据面服务，负责 write key、domain/CORS、internal traffic、quota 等 runtime enforcement；配置 CRUD 仍在 `simpletrack-saas`。

## 当前待评审的总方向

1. Supastarter 的 1 天 SimpleTrack spike 是否顺畅通过。
2. 支付 provider 的具体上线顺序和平台配置，放到上线前处理。
3. `analytics-core` 的表模型、接口分层、存储模型和从 xwl_bi 抽取的具体代码步骤。
4. 企业级控制台 UI 是否直接跟随 Supastarter UI 栈，还是在模板 shell 内重做业务控制台。
5. Supastarter 的 marketing/docs app 是否满足产品官网和 docs 需要；若不满足再轻量定制。
6. Umami 源码启发的 `analytics-core` 优化项如何取舍：事件属性 typed storage、session/visit 隐私策略、client enrich、bot/IP 过滤、ClickHouse 读侧优化的具体落地顺序、SDK 分阶段路线和 performance metrics 是否进入 P1、P1.5 或 P2。

## 维护规则

- 确定了就写入 `已确定`。
- 还没有最终拍板但已经进入讨论，就写入 `待评审`。
- 被明确排除的方案，写入 `已否决`，并说明原因，避免后续重复评估。
- 每条决定都要能回答三个问题：为什么这样做、影响哪些模块、下一步怎么验证。
- 每次任务完成后，必须更新上方“实施计划完成列表”。
- 已完成任务如果被重构或重新打开，状态必须重置为 `待完成`。
