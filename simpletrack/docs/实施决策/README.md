# SimpleTrack 实施决策库

> 目录用途：存放 SimpleTrack 已确定要实施的决定、分阶段实施计划，以及仍需评审的关键事项。后续每次确定新决定，都先更新本目录，再继续改原型或生产代码。

## 文档索引

| 文档 | 用途 | 状态 |
| --- | --- | --- |
| 2026-05-07 | `visit_id` 长期方案定稿：不再依赖 readback 临时派生，改为 collect 写入前确定、ClickHouse event / `_properties` 表持久化、Realtime / Events 直接读取存储字段；同步新增 `session_id / visit_id / distinct_id` Q&A | P1-002C、P1-005C、P1-005D、visit_id |
| 2026-05-06 | `simpletrack/prototype/simpletrack-enterprise-mvp/` 的页面数据契约已重新收口到当前原型页：`/onboarding`、`/dashboard`、`/events`、`/goals`、`/settings`，`team` / `funnels` / `insights` 退回 P2/P3 候选，不再写进 P1 页面 contract | P0-001、页面 contract、enterprise-mvp、P1 收口 |
| 2026-05-04 | 历史过渡：`analytics-core` 的 readback 结果曾先行带上 provisional `visit_id`，用于让 Realtime / Events UI 提前验证展示价值；该口径已被 2026-05-07 的持久化 `visit_id` 方案替代 | P1-002C、P1-005D、visit_id、readback |
| 2026-05-03 | 浏览器 SDK 已补齐 DNT opt-in 隐私开关和 UTM/click id 白名单；随后完成边界纠偏：SDK 不再属于 `analytics-core`，改由 `simpletrack-anaysitics-service` 静态托管，quickstart 改为 `data-write-key` 接入 | P1-002B、P1-002C、P1-004、P1-005、DNT、UTM/click id、Web tracker SDK、docs/quickstart |
| 2026-05-03 | P1-004 Web tracker SDK 最短链路已完成，后续归属调整为产品运行时层：`analytics-core` 只保留 Go library 能力，`simpletrack-anaysitics-service` 承接 `/tracker.js` 和 `/collect` | P1-004、P1-005、Web tracker SDK、analytics-core、analytics-service |
| 2026-05-03 | 在 `analytics-core` 落地 P1-002B/C 第一版：新增 collect pre-queue `Stage` 管道、盐化窗口 `SessionResolver`、client UA/referrer/IP hash 派生属性、bot/internal traffic 过滤；HTTP 层默认不信任 `X-Forwarded-For` / `X-Real-IP`，只有显式 `WithTrustedProxyHeaders()` 才启用可信代理头；子仓提交 `9c06b0d` 已推送，并通过结对审查、全仓 Go 验证和真实 e2e | analytics-core、P1-002B、P1-002C、collect、隐私、bot/internal traffic |
| [分阶段实施计划.md](分阶段实施计划.md) | 记录 P0/P1/P2/P3 的目标、范围、交付物和验收标准 | 已确定，持续更新 |
| [待评审事项.md](待评审事项.md) | 记录还没有拍板的技术栈、模板、支付、数据面复用方案 | 待评审，持续更新 |
| [技术栈底座决策.md](技术栈底座决策.md) | 记录当前已经形成的技术底座方向和边界 | 已确定 + 待评审 |
| [付费SaaS模板本地对比.md](付费SaaS模板本地对比.md) | 基于 `template-src` 本地源码和 MakerKit 官方资料对比付费模板 | 已确定先选 Supastarter，持续更新 |
| [analytics-core实施方案.md](analytics-core实施方案.md) | 记录 `analytics-core` 的 P1 抽取边界、模块草案、EventBus 方案和 xwl_bi 代码评审结论 | 已确定，设计细节持续评审 |
| [SimpleTrack分析服务职责边界.md](SimpleTrack分析服务职责边界.md) | 说明 `simpletrack-saas`、`simpletrack-anaysitics-service` 与 `analytics-core` Go library 的控制面/数据面边界 | 已确定，持续更新 |
| [xwl_bi后端架构参考映射.md](xwl_bi后端架构参考映射.md) | 将 `references/xwl_bi-backend/` 的后端架构设计映射到 `analytics-core`，明确只参考架构不搬旧业务代码 | 已确定，持续更新 |
| [Umami源代码参考映射.md](Umami源代码参考映射.md) | 将 `references/umami/` 的官方源码快照映射到 SimpleTrack P1/P2/P3 实施边界，明确只参考对象体系和实现策略 | 已确定，持续更新 |

## 修订记录

| 日期 | 修订内容 | 影响范围 |
| --- | --- | --- |
| 2026-05-08 | 细化 P1.5-001 ClickHouse 读侧优化取舍：明确 projection、materialized view、小时聚合表和属性治理的适用边界，推荐长期分层路线为“先做属性治理和 query plan 约束，再按稳定查询引入 MV / 小时聚合表 / projection” | P1.5-001、analytics-core、ClickHouse 读侧优化 |
| 2026-05-07 | P1-005D 与 P1-005 收口为已完成：Realtime / Events 内部读接口、query routes、query token 轮换、property filter 白名单、visit_id 透传、SaaS readback helper、Website selector、repeatable `property_filter` 和组合查询回归均已验证；更复杂的聚合分析与 Breakdown/Compare 维持到 P1.5/P2 | P1-005、P1-005D、analytics-service、simpletrack-saas、Events、Realtime |
| 2026-05-07 | P1-005B 与 P1-005C 收口为已完成：collect runtime service 已具备 Fiber `/collect`、write key/source enforcement、Redis durable enqueue、可选同进程 ingestion worker、启动 schema surface 校验和本地/小部署 routed table auto-create；SaaS control-plane runtime enforcement 已具备 HTTP resolver、ETag 条件重验证、disable/delete 后撤权、runtime-source API、Websites CRUD/enable/disable/delete、active source quota 和 runtime source 字段回写 | P1-005B、P1-005C、analytics-service、simpletrack-saas、runtime source、control plane |
| 2026-05-07 | P1-002 与 P1-002C 收口为已完成：session/visit resolver 已覆盖缺失时服务端派生、显式 SDK `visit_id` 保留、ClickHouse event / `_properties` 持久化、Realtime / Events 读侧返回、SaaS runtime-source `visit_salt` / `visit_window_seconds` 输出和 Events URL/readback 查询态；salt 轮换、cookie/no-cookie、server identity、Sessions 专页和 retention 产品化继续放 P1.5/P2 | P1-002、P1-002C、analytics-core、analytics-service、simpletrack-saas、visit_id |
| 2026-05-07 | P1-002B 收口为已完成：`simpletrack-saas` 已验证 Websites 表单会把 additional origins、Bot user agents、Internal CIDRs、Internal IPs 和 fingerprint 开关完整回写 runtime source；`analytics-service` 已验证 internal IP 过滤审计不记录 raw IP；过滤统计与可视化不阻塞 P1，降到 P1.5/P2 评审 | P1-002B、P1-005B、P1-005C、simpletrack-saas、analytics-service、internal traffic |
| 2026-05-07 | `analytics-core` 补齐 Events 组合查询计划回归：同一个 ClickHouse query plan 同时覆盖时间窗、`event_name`、`distinct_id`、`visit_id`、分页、`event_name` 排序、event/user property filters，并验证业务值继续通过 bound args 参数化；子仓提交 `7c29667` 已推送 | P1-005D、analytics-core、ClickHouse、EventQueryBuilder |
| 2026-05-07 | P1-005D 补齐 Events 组合查询回归：`simpletrack-anaysitics-service` 验证 `event_name`、`distinct_id`、`visit_id`、分页、排序和 repeatable `property_filter` 会进入同一个 `EventListQuery`；`simpletrack-saas` 验证同一组条件会序列化为单个 readback 请求；子仓提交 `09656b6`、`bce3335` 已推送 | P1-005D、analytics-service、simpletrack-saas、Events、query combination |
| 2026-05-07 | `simpletrack-anaysitics-service` 的 Events readback API 文档与当前实现对齐：OpenAPI 将 `property_filter` 标为 repeatable array query parameter，并补齐 `sort_field=event_name`；README 示例展示 `event_name` + 多个 `property_filter` 的组合查询；子仓提交 `3672fb9` 已推送 | P1-005D、analytics-service、OpenAPI、Events、property filter |
| 2026-05-07 | `simpletrack-anaysitics-service` README 的本地 runtime source 示例补齐 `visit_salt` 和 `visit_window_seconds`，并明确 `session_salt` / `visit_salt` / `client_hash_salt` 都是 server-only runtime secrets；子仓提交 `958b26c` 已推送 | P1-005B、P1-002C、analytics-service、deployment params、visit_id |
| 2026-05-07 | `simpletrack-saas` 将 P1 quota 从单个 source limit helper 收口为明确的 `SimpleTrackPlanLimits` 策略：当前只限制 active website source，Free=1、Pro/Lifetime=5、Enterprise=不限，事件量、留存期、成员数和用量计费不提前进入 P1；子仓提交 `41f8d4d` 已推送 | P1-005C、simpletrack-saas、Websites、quota、plan limits |
| 2026-05-07 | `simpletrack-saas` 的 Websites 设置区补齐可见标签：Source name、Primary origin、Additional origins、Bot user agents、Internal CIDRs、Internal IPs 都从仅靠 placeholder / aria-label 改为明确配置项，便于 P1 internal traffic 配置在控制面被理解和维护 | P1-002B、P1-005C、simpletrack-saas、Websites、internal traffic、runtime source |
| 2026-05-07 | `analytics-service` 补 internal traffic 过滤审计回归：通过可信代理头触发 internal IP filter，确认日志只记录 event/source 边界和 `reason=internal ip`，不记录原始客户端 IP | P1-002B、P1-005B、analytics-service、internal traffic、audit、privacy |
| 2026-05-07 | 用当前模型复审 repeatable `property_filter` 链路后，补齐 `simpletrack-saas` 前端与 URL normalization 的共享 5 条上限：组件禁用第 6 条新增，query-state 也会截断超额直达 URL，避免页面发出会被 `analytics-service` 拒绝的查询 | P1-005D、simpletrack-saas、Events、property filter、query-state |
| 2026-05-07 | `analytics-service` 的 Events query boundary 补 repeatable `property_filter` 回归：`/v1/events` 现在有服务层测试确认多个 property_filter 参数会按顺序进入 `analytics-core` 查询计划，和 `simpletrack-saas` 的多条件 Events 页面保持同一契约 | P1-005D、analytics-service、query parser、property filter |
| 2026-05-07 | `simpletrack-saas` 的 Events 页从单条 property filter 升级为 repeatable `property_filter` 多条件查询：page/query-state/readback 统一切到数组模型，支持新增/删除过滤行并输出重复 query 参数，同时保留 legacy 单字段回退兼容 | P1-005D、simpletrack-saas、Events、query builder |
| 2026-05-07 | `analytics-core` 的 P1-002B enrich 边界继续收口：新增 browser / OS / device 的可插拔 UA 派生，以及基于离线 MaxMind mmdb 的 geo enrichment 接口；`simpletrack-anaysitics-service` 可通过 `ANALYTICS_SERVICE_GEOIP_MMDB_FILE` 提供地理库文件，collect 侧继续保持框架无关 | P1-002B、collect、analytics-core、analytics-service、geo、browser、device |
| 2026-05-07 | 明确 SimpleTrack 仍处于新建项目阶段，`src/analytics-core`、`src/simpletrack-saas` 和 `src/analytics-service` 的 schema 调整只走初始化 / 建表 / 启动校验路径，不提前引入迁移 SQL、backfill、兼容分支或完整迁移框架；生产级迁移逻辑留到真实上线后单独评估 | 仓库治理、schema、P1-005B、P1-005C |
| 2026-05-07 | `visit_id` 从 readback provisional 字段升级为 canonical analytics visit key：`analytics-core` 增加 VisitResolver、存储字段、writer/reader/query 支持；`simpletrack-anaysitics-service` runtime config 增加 `visit_salt` / `visit_window_seconds`；`simpletrack-saas` runtime-source 输出 visit 配置并让 Realtime / Events 使用真实存储字段 | P1-002C、P1-005C、P1-005D、analytics-core、analytics-service、simpletrack-saas |
| 2026-05-07 | `simpletrack-anaysitics-service` 补充 runtime source resolver 与 HTTP resolver cache 边界注释，作为 `/tracker.js`、`/v1/realtime`、`/v1/events` 源码解读的稳定证据基线；子仓提交 `825e366` 已推送，父仓随后同步 gitlink 和 `docs/analytics-source-reading/collectapi-query-and-tracker-flow.md` 引用 | P1-005B、P1-005C、P1-005D、analytics-service、源码解读、submodule |
| 2026-05-07 | `analytics-core` 的 `collect/httpapi` 和 `simpletrack-anaysitics-service` HTTP runtime 已从 fasthttp 迁移到 Fiber v3：core 保持 `collect.Handler` 框架无关，service 改为 `collectapi.NewApp` 返回 Fiber app，内部 Events / Realtime 路由支持配置且默认保留 `/v1/events`、`/v1/realtime`，CORS 改用 Fiber middleware，并新增 Swagger UI + OpenAPI 文件；子仓提交 `58668c9`、`1e8f283` 已推送，父仓随后同步 gitlink | P1-002、P1-005B、P1-005D、analytics-core、analytics-service、Fiber、Swagger、submodule |
| 2026-05-05 | `simpletrack-saas` 的 Events 页补时间窗口预设：在现有事件名 / distinct id / 单条 property filter / 排序 / 分页基础上，再补 `30m / 6h / 24h / 7d` 时间窗口查询态；窗口会进入服务端 href、分页链接与 readback 请求，不再固定锁死最近 24 小时，子仓提交 `fa822d1` 已推送，父仓随后同步 gitlink | P1-005D、simpletrack-saas、Events、time-window、submodule |
| 2026-05-05 | `simpletrack-anaysitics-service` 补齐 control-plane runtime revoke 回归：新增 HTTP resolver revalidation 驱动的 handler 级测试，验证已缓存 source 在控制面 disable 或 delete 后，会立即拒绝新的 `/collect` 与内部 `/v1/realtime` 访问，不必等待服务重启；子仓提交 `8ae0b78` 已推送，父仓随后同步 gitlink | P1-005C、P1-005D、analytics-service、runtime-revocation、submodule |
| 2026-05-05 | `simpletrack-saas` 的 `Websites` 控制面补齐最小 CRUD 与剩余 runtime source 设置收口：在 create + enable / disable + 编辑基础上，再补 delete 动作和 `include client fingerprint` 编辑通路；`allowedOrigins`、`bot user agents`、`internal CIDRs`、`internal IPs` 和 fingerprint 开关都可回写 Website runtime source config，子仓提交 `5574dd2` 已推送，父仓随后同步 gitlink | P1-005C、simpletrack-saas、Websites、runtime-source、submodule |
| 2026-05-05 | `simpletrack-saas` 的 `Websites` 控制面继续前进到更完整的 runtime source 维护：在 create + enable / disable + origin 编辑基础上，再补 `bot user agents`、`internal CIDRs`、`internal IPs` 的编辑入口；这些配置会直接落回 Website runtime source config，子仓提交 `f0be432` 已推送，父仓随后同步 gitlink | P1-005C、simpletrack-saas、Websites、runtime-source、submodule |
| 2026-05-05 | `simpletrack-saas` 的 `Websites` 控制面继续前进到最小可维护状态：在 create + enable / disable 基础上，再补 name / primary origin / additional origins 的编辑入口；编辑路径会保留 secondary `allowedOrigins`，不会把多 origin runtime config 压回单 origin，子仓提交 `44b7a35` 已推送，父仓随后同步 gitlink | P1-005C、simpletrack-saas、Websites、runtime-source、submodule |
| 2026-05-05 | `simpletrack-saas` 的 `Websites` 控制面继续前进到最小可维护状态：在真实 source 管理与 enable / disable 基础上，再补 name/domain 编辑入口；编辑路径会保留 secondary `allowedOrigins`，不会因保存表单把多 origin runtime config 压成单 origin，子仓提交 `b049751` 已推送，父仓随后同步 gitlink | P1-005C、simpletrack-saas、Websites、runtime-source、submodule |
| 2026-05-05 | `simpletrack-saas` 的 Website source limit 实现与文档口径对齐：disabled source 创建不再占用 active source limit，并补 `createWebsiteWithinLimit()` 直达回归；子仓提交 `74248a6` 已推送，父仓随后同步 gitlink | P1-005C、simpletrack-saas、Websites、quota、submodule |
| 2026-05-05 | `simpletrack-saas` 的 `Websites` 页继续前进到真实 runtime source 管理：在真实组织 Website 数据和最小创建 source 表单基础上，再补 enable / disable 动作，readback 按钮与 enabled 状态保持一致；创建与启用路径都通过 serializable 事务 helper + `P2034` 重试守住 active source limit，子仓提交 `dcc1db8` 已推送，父仓随后同步 gitlink | P1-005C、P1-005D、simpletrack-saas、Websites、runtime-source、submodule |
| 2026-05-05 | `simpletrack-saas` 的 `Websites` 页已从 mock 列表切到真实组织 Website 数据，并补最小创建 source 表单、write key 展示、Quickstart / Realtime / Events 入口；创建路径通过 serializable 事务 helper + `P2034` 重试守住 active source limit，子仓提交 `5d165c6` 已推送，父仓随后同步 gitlink | P1-005C、P1-005D、simpletrack-saas、Websites、runtime-source、submodule |
| 2026-05-04 | `simpletrack-saas` 的 Realtime / Events 页面开始直接展示 provisional `visit_id`，把已确认可用的 readback 字段前移到产品界面；`analytics-readback-core` 同步把 visit 维度纳入 view model，并补页面级回归；子仓提交 `236d179` 已推送，父仓随后同步 gitlink | P1-005D、simpletrack-saas、visit_id、readback UI、submodule |
| 2026-05-04 | `analytics-core` 的 readback `visit_id` 改为独立 128-bit digest，并补 30 分钟 UTC bucket 边界 / 归一化测试；`simpletrack-anaysitics-service` 的 Realtime / Events 继续透传该字段，两个子仓提交已推送，父仓随后同步 gitlink | P1-002C、P1-005D、visit_id、readback、submodule |
| 2026-05-04 | `src/analytics-service` 已接入远端 `simpletrack/anaysitics-service` 仓库：本地 `main` 已推送，父仓开始登记 `src/analytics-service` 子模块；同步把 `AGENTS.md` 的子仓提交顺序和 SSH 规则扩展到第三个子仓库 | 仓库治理、P1-005、analytics-service、submodule、SSH |
| 2026-05-04 | `simpletrack-anaysitics-service` 的内部 Realtime / Events 读回放补 query token 生命周期与审计基础：`ANALYTICS_SERVICE_QUERY_TOKENS_JSON` 现在支持带 `id`、`not_before`、`expires_at` 的结构化凭据；运行时会拒绝未生效或已过期 token，并对命中轮换 token、过期 token、未来 token 和未知 token 记录审计日志 | P1-005D、simpletrack-anaysitics-service、query token lifecycle、audit、Realtime、Events |
| 2026-05-04 | `simpletrack-anaysitics-service` 的内部 Events 读回放补属性过滤入口：runtime source config 新增 `allowed_property_filters`，`/v1/events` 接收 URL 编码 JSON `property_filter`，并在服务层和 `analytics-core` ClickHouse query builder 双重白名单校验后才执行查询 | P1-005D、simpletrack-anaysitics-service、property filters、query whitelist、analytics-core |
| 2026-05-04 | `simpletrack-anaysitics-service` 的内部 Realtime / Events 读回放补 query token 轮换支持：保留 `ANALYTICS_SERVICE_QUERY_TOKEN` 单 token 兼容，同时新增 `ANALYTICS_SERVICE_QUERY_TOKENS_JSON` 短窗口 allowlist，可在 SaaS 切换内部读 token 时接受当前 token 和上一版 token | P1-005D、simpletrack-anaysitics-service、query token rotation、Realtime、Events |
| 2026-05-03 | `simpletrack-saas` 的 Website selector 改为 client-safe DTO：页面只把 `id/name/domain/enabled` 传给浏览器，完整 Website row 仍留在服务端 readback 路径；`website` query 参数切换组织内 enabled source，禁用 source 仍保持 fail-soft | P1-005D、simpletrack-saas、Website selector、server/client boundary |
| 2026-05-03 | `simpletrack-saas` 的 Realtime / Events 页面补真实 Website 选择器初版：页面读取组织下 Website 列表，通过 `website` query 参数切换 enabled source；服务端 readback helper 只接受组织内启用 source，禁用或无 source 仍保持 fail-soft | P1-005D、simpletrack-saas、Website selector、Realtime、Events |
| 2026-05-03 | `simpletrack-saas` 的 Realtime / Events 页面接入 `simpletrack-anaysitics-service` 内部读回放初版：页面服务端按组织查启用 Website，使用内部 query token 调 `/v1/realtime` 与 `/v1/events`，未配置服务、无启用 source 或服务异常时保持 fail-soft 空态 | P1-005D、simpletrack-saas、simpletrack-anaysitics-service、Realtime、Events |
| 2026-05-03 | `simpletrack-saas` 的 Events 页面补齐服务端筛选与分页：`event_name`、`distinct_id`、`sort_field`、`sort_direction`、`limit`、`offset` 均走白名单归一化，`hasMore` 驱动上一页 / 下一页，内部 query token 仍不下发到浏览器 | P1-005D、simpletrack-saas、Events、query whitelist、pagination |
| 2026-05-03 | `simpletrack-saas` 的 Events 页面再补重复 query 参数与空页分页硬化：重复参数取首值后归一化，空页但 `offset > 0` 时仍显示 Previous / Next 控件，避免书签或重复参数把服务器渲染页打崩 | P1-005D、simpletrack-saas、Events、query normalization、pagination hardening |
| 2026-05-03 | `simpletrack-anaysitics-service` 的内部 Events / Realtime 读回放开始落地：新增 `/v1/realtime` 与 `/v1/events`，通过内部 bearer token 保护并复用 `analytics-core` 的 `EventReader`，同时补齐浏览器/SaaS 页面调用需要的 query `OPTIONS` preflight | P1-005D、simpletrack-anaysitics-service、Events、Realtime、query API、CORS |
| 2026-05-03 | `simpletrack-anaysitics-service` 增加 `ANALYTICS_SERVICE_SOURCE_RESOLVER=http` 控制面 HTTP resolver，可用 bearer token 向 SaaS 控制面读取 runtime source config，并用短 TTL 缓存 + `ETag` 条件重验证降低热路径控制面压力且避免 disabled/source salt/origin 变更被陈旧缓存掩盖；控制面 URL 默认必须是 HTTPS，本地 loopback HTTP 需要显式开启 `ANALYTICS_SERVICE_CONTROL_PLANE_ALLOW_INSECURE_LOOPBACK=true`；同进程 ingestion 仍要求 `ANALYTICS_SERVICE_SOURCES_JSON` 作为启动 schema surface，并拒绝 HTTP resolver 返回未在启动阶段校验过的 source | P1-005C、simpletrack-anaysitics-service、SaaS control-plane runtime enforcement |
| 2026-05-03 | `simpletrack-anaysitics-service` 为显式开启的 ingestion worker 增加可选 ClickHouse routed table auto migration；默认仍先校验 schema 并 fail-closed，本地/小部署可用 `ANALYTICS_SERVICE_CLICKHOUSE_AUTO_MIGRATE=true` 创建当前 runtime config 内所有启用 source 所需的 event / `_properties` 表 | P1-005B、simpletrack-anaysitics-service、ClickHouse schema、运行时装配 |
| 2026-05-03 | 按服务命名要求把运行时服务展示名统一为 `simpletrack-anaysitics-service`；继续推进 P1-005B，服务可显式开启同进程 ingestion worker，装配 Redis Stream、MySQL checkpoint guard、ClickHouse native writer 和 typed property indexing；同时强制 session/client salt 由服务端 runtime config 提供，并在 ingestion 启动时校验 ClickHouse 表就绪 | P1-005、P1-005B、simpletrack-anaysitics-service、运行时装配、隐私 salt、ClickHouse schema |
| 2026-05-03 | `simpletrack-saas` 落地内部 runtime-source API，并补齐 Website 运行时配置模型和 write key -> runtime source config 的读取路径；`simpletrack-anaysitics-service` 现在可按 write key 从控制面读取 server-only salts、allowlist 和 traffic 配置 | P1-005C、simpletrack-saas、simpletrack-anaysitics-service、SaaS control-plane runtime enforcement、Website 模型 |
| 2026-05-03 | 建立 `src/analytics-service` 本地仓库，明确 `analytics-core` 作为 Go 第三方库而非独立业务服务；core 公共 API 调整为根目录包，Browser SDK 从 core 移到 `simpletrack-anaysitics-service` 静态交付，docs/quickstart 改为 write key 接入 | P1-005、analytics-service、analytics-core、Web tracker SDK、控制面/数据面边界 |
| 2026-05-03 | 在根目录 `AGENTS.md` 固化代码结对规则：代码改动必须由主代理实现、Codex native 子代理基于 `AGENTS.md` 复审当前 diff，提交前处理阻塞问题；纯文档更新默认不强制子代理 | 仓库治理、代码审查、协作规范 |
| 2026-05-03 | 在 `analytics-core` 把 typed property rows 接入 ingestion 热路径：新增 `PropertyIndexingEventWriter`、`PropertyWriteGuard`、MySQL `property_indexing_status`，属性写入只 reclaim 明确 failed checkpoint，processing 视为结果不明以避免重复追加；同步子仓提交到 `0586ee6` | analytics-core、P1-002A、ClickHouse、属性索引、跨表幂等、代码审查规则 |
| 2026-05-02 | 在 `analytics-core` 为 Events 查询新增 typed property filter：属性 scope/name/type/value 进入 allowlist + 绑定参数，ClickHouse 使用 tuple `IN` 半连接查询属性表；真实 e2e 复验修复 correlated `EXISTS` 外层 alias 卡点；同步子仓提交到 `ae5c21c` | analytics-core、P1-002D、P1-002A、Events 查询、ClickHouse 查询安全 |
| 2026-05-02 | 在根目录 `AGENTS.md` 固化提交规范：按 `$git-commit-cn` 的范围核对、分路径说明、提交后推送流程执行，但提交信息统一使用英文并保留必要 Lore trailer | 仓库治理、提交规范、协作规范 |
| 2026-05-02 | 修复 `analytics-core` e2e 冷启动依赖 readiness 问题，为 Redis/MySQL/ClickHouse 连接增加重试窗口，并新增真实 ClickHouse `PropertyBatchWriter` e2e 写入读取验证；同步子仓提交到 `2cc83e1`，排查记录写入 `docs/开发环境卡壳问题记录.md` | analytics-core、P1-002A、P1-002E、本地运行依赖、ClickHouse |
| 2026-05-02 | 在 `analytics-core` 新增 `EventPropertyWriter` 契约和 ClickHouse `PropertyBatchWriter`，可把 typed property rows 写入同源路由的 `_properties` 物理表；暂不并入事件写入热路径，等待跨表重试/幂等语义评审；同步子仓提交到 `5698f11` | analytics-core、P1-002A、ClickHouse、事件属性、用户属性、属性物理写入 |
| 2026-05-01 | 在 `analytics-core` 新增 storage-neutral typed property rows 和 `FlattenEventProperties`，把 event/user properties 从原始 map 稳定展开为可供 `EventWriter`、属性字典和属性过滤复用的逻辑记录；同步子仓提交到 `f64ed3c` | analytics-core、P1-002A、事件属性、用户属性、属性展开 |
| 2026-05-01 | 在 `analytics-core` 为 collect 阶段新增事件属性/用户属性入口约束：属性 key 形状、数量、标量类型、字符串长度和有限数字校验；同步子仓提交到 `9224961` | analytics-core、P1-002A、collect、事件属性、用户属性 |
| 2026-05-01 | 在 `analytics-core` 为 Events 查询继续补齐类型化过滤字段/operator 白名单、filter 数量上限和 `ErrInvalidEventQuery` 错误分类；同步子仓提交到 `cd9e48f` | analytics-core、P1-002D、Events 查询、ClickHouse 查询安全 |
| 2026-05-01 | 在 `analytics-core` 为 Events 查询新增类型化排序字段和方向白名单，避免后续 UI 排序参数直接穿透 SQL；同步子仓提交到 `3afaf4c` | analytics-core、P1-002D、Events 查询、ClickHouse 查询安全 |
| 2026-05-01 | 安装并补跑 `golint ./...`，修正 `analytics-core` e2e helper 的 `context.Context` 参数顺序；同步子仓提交到 `0538a0b` | analytics-core、P1-002E、本地验证、代码规范 |
| 2026-05-01 | 在 `analytics-core` 新增 opt-in 端到端测试，真实验证 collect -> Redis Stream -> ingestion -> ClickHouse -> Realtime/Events reader，并将 P1-002E 标记为已完成；同步子仓提交到 `4931f15` | analytics-core、P1-002、Realtime、Events、本地运行依赖 |
| 2026-05-01 | 将 Umami 源码深解中可被 `analytics-core` 吸收的优化点排入实施计划和评审表，新增事件属性、client enrich、session/visit、查询白名单、Realtime/Events 验收、Web SDK 和 ClickHouse 读侧优化任务 | analytics-core、Umami 参考资产、P1/P1.5/P2 计划、待评审事项 |
| 2026-05-01 | 为 Umami P1 源码实现参考补充 `Q&A/` 概念解释，覆盖 identify、Prisma schema、字段白名单、Core Web Vitals、SDK 取舍、bot/IP 过滤、Zod、storage dispatch、session 隐私机制和 JSON 属性风险 | Umami 参考资产、analytics-core、SimpleTrack docs/quickstart、文档体系 |
| 2026-05-01 | 在 `simpletrack/docs/umami/docs/源码实现参考/` 落地 Umami P1 数据管道源码分章节深解，补齐整体架构、tracker、collect、写入、模型、Realtime/Events、过滤查询和 SimpleTrack / `analytics-core` 启发 | Umami 参考资产、analytics-core、Realtime、Events、文档体系 |
| 2026-05-01 | 将 Umami 官方 GitHub 源码克隆为 `references/umami/` 只读参考快照，新增源码实现审阅和参考映射文档 | Umami 参考资产、analytics-core、Realtime、Events、仓库治理 |
| 2026-05-01 | 调整 `analytics-core` 本地 compose 默认端口，避开本机 Redis/MySQL 端口冲突，并记录 Docker 卡壳处理；同步子仓提交到 `c7aa2cb` | analytics-core、本地运行依赖、开发环境 |
| 2026-05-01 | 在 `analytics-core` 新增 Redis Stack、MySQL、ClickHouse 本地 `docker-compose.yml` 和 README 运行说明；同步子仓提交到 `0bd1cc4` | analytics-core、本地运行依赖、端到端验证 |
| 2026-05-01 | 在 `analytics-core` 落地 `storage.EventReader` 和 ClickHouse/GORM 查询执行器，让 Realtime/Events query plan 能扫成 `EventRecord`；同步子仓提交到 `a072275` | analytics-core、Realtime、Events、ClickHouse 查询 |
| 2026-05-01 | 明确 `analytics-core` 的 `ingestion.Processor` 是 P1 worker 边界，补充 Run 级测试和 Example；同步子仓提交到 `a22ab6e` | analytics-core、ingestion worker、队列消费 |
| 2026-05-01 | 在 `analytics-core` 落地 `EventQueryBuilder` 查询契约和 ClickHouse/GORM Events、Realtime query plan 边界；同步子仓提交到 `7ab7b12` | analytics-core、Realtime、Events、ClickHouse 查询 |
| 2026-05-01 | 在 `analytics-core` 落地 GORM/MySQL `IngestionStatusGuard` 与 `ingestion_status` 幂等状态表，并强化函数体阶段注释强制规范；同步子仓提交到 `71f5ae3` | analytics-core、MySQL/GORM、代码规范 |
| 2026-05-01 | 澄清 `collect.Handler` 是事件上报核心处理器而非 HTTP 路由函数，并将“污染”表述改为“框架耦合 / 边界穿透”；同步 `analytics-core` 子仓到 `0c6bf8c` | analytics-core、HTTP collect API、协作规范 |
| 2026-05-01 | 在 `analytics-core` 落地 ClickHouse native batch `EventWriter`、`EventWriteGuard` 幂等边界和 Go 结构体/接口注释强制规范 | analytics-core、ClickHouse、代码规范 |
| 2026-04-30 | 评估 xwl_bi 的 Fiber 与 fasthttp/fasthttprouter HTTP 栈；后续已改为 `analytics-core` 的 collect HTTP API 使用 Fiber v3，并保持 `collect.Handler` 框架无关 | analytics-core、HTTP collect API、协作规范 |
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
| 2026-04-30 | 将 Supastarter marketing/docs 占位内容替换为 SimpleTrack 产品介绍、定价语义和 docs/quickstart，并记录 Windows 验证环境要求 | P1 公开产品入口、Supastarter spike、协作规范 |
| 2026-04-30 | 完成 Supastarter marketing/docs/mail-preview 浏览器截图验证，替换可见模板占位，并将本地 PostgreSQL Docker 配置对齐到 SimpleTrack | P1 公开产品入口、邮件预览、SaaS 控制面数据库 |
| 2026-04-30 | 在 `analytics-core` 落地 Redis Stream pending 优先重试、MaxAttempts 死信队列和 ingestion 幂等处理边界，并通过 `go test ./...` | analytics-core、EventBus、ingestion |
| 2026-04-30 | 将本地 `xwl_bi` 后端源码与关键文档复制为 `references/xwl_bi-backend/` 只读临时参考快照，供 `analytics-core` 实现对照使用 | 分析数据面、参考资产、仓库治理 |
| 2026-04-30 | 在 `analytics-core` 落地 collect 请求标准化、`EventWriter` 写入接口和 ClickHouse `TableRouter`，启动 P1-002 数据管道最小闭环实现 | analytics-core、collect、storage、ClickHouse |
| 2026-04-30 | 新增 xwl_bi 后端架构参考映射，明确快照主要参考模块边界、启动装配、消费链路、ClickHouse 写入/查询分层和分析服务拆分 | analytics-core、xwl_bi 参考、后端架构 |
| 2026-04-30 | 在 `analytics-core` 落地 `collect.Handler`，完成 collect 请求标准化到 EventBus 发布的最小链路 | analytics-core、collect、EventBus |

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
| P0-003 | 企业分析控制台 UI 可改造性评审 | 进行中 | `src/simpletrack-saas` 已能在 Supastarter `apps/saas` 组织内导航挂载 Websites、Realtime、Events 页面草案 | 做截图级评审，确认密度、表格、侧边栏和空态是否满足企业分析控制台 |
| P0-004 | Supastarter for Next.js 接入核验 | 进行中 | 已确定先选 Supastarter；`src/simpletrack-saas` 已作为独立子仓库推送；Websites 页已从 UI-only gate 前进到真实 Website source 列表 + 最小创建入口；marketing/docs/mail-preview 已完成浏览器截图验证；支付先按模板已有 Stripe、Lemon Squeezy、Polar、Creem、Dodo Payments provider 接入 | 核验许可证、私有仓库、闭源修改、团队席位，并在需要 authenticated SaaS 流程时用 Docker Postgres 验证 |
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
| P1-003 | 产品官网 / Marketing Site / 公开站点 | 已完成 | 已从 `template-src/ai-supastarter-template` 初始化 `src/simpletrack-saas` 工作副本；marketing 文案、pricing 语义、docs/quickstart、mail-preview 品牌文案和截图级验证已完成；公开站点首屏已露出下一节内容 | 后续只做轻量文案和视觉微调，不阻塞 P1 数据管道 |
| P1-004 | Web tracker SDK 最短链路 | 已完成 | P1 浏览器 SDK 已从 `analytics-core` 迁出，由 `src/analytics-service/public/tracker.js` 作为运行时服务静态资产托管；SDK 继续支持 auto pageview、SPA route pageview、manual track、identify、debug、snippet queue、`localStorage` fallback、非法 event name 拦截、opt-in DNT 和 allowlisted UTM/click id；docs/quickstart 已改为 `data-write-key` 接入，`simpletrack-anaysitics-service` 通过 write key 覆盖 tenant/project/source | React/Next/Node/mobile SDK、多语言 SDK、performance metrics 和 CDN/版本化发布策略放后续阶段评审 |
| P1-005 | SimpleTrack Anaysitics Service 本地仓库 | 已完成 | 已创建并推送 `src/analytics-service` 远端仓库 `simpletrack/anaysitics-service` 并登记父仓子模块，服务名 `simpletrack-anaysitics-service`；当前提供 `/healthz`、`/tracker.js`、`OPTIONS /collect`、`POST /collect`、内部 `/v1/realtime`、内部 `/v1/events`，并用 `MemoryResolver` 或 SaaS HTTP resolver 执行 runtime source config；`simpletrack-saas` 内部 runtime-source API、Websites 控制面 CRUD、Realtime/Events 页面服务端读回放已落地；内部 query token 已支持短窗口轮换 allowlist、结构化生效/过期时间和审计日志，Events 属性过滤已接入 `allowed_property_filters`；P1-005B/C/D 均已完成，复杂聚合分析与 Breakdown/Compare 维持到 P1.5/P2 | 保持现有回归；更复杂的聚合查询与高级分析放 P1.5/P2 |
| P1-005A | `analytics-core` root-level Go library API | 已完成 | `analytics-core` 已调整为可被 Go 服务引用的根目录公共包：`contracts`、`collect`、`eventbus`、`ingestion`、`storage` 等；不再把 Browser SDK 放进 core，也不把 core 作为长期业务服务运行 | 后续公共 API 变更需保持外部服务 import 稳定 |
| P1-005B | collect runtime service | 已完成 | `simpletrack-anaysitics-service` 已实现 write key 解析、source enabled、Origin allowlist、Fiber CORS middleware、客户端 scope 覆盖、bot/internal traffic 过滤、client enrich、session resolver、Redis durable enqueue，以及显式开启的同进程 ingestion worker；HTTP runtime 已迁移为 Fiber app，`collectapi.Handler` 保持业务依赖持有边界，`main.go` 只负责生命周期；worker 复用 `analytics-core` 的 Redis Stream subscribe、MySQL guard、ClickHouse native writer 和 typed property indexing；write key 不再作为 privacy salt fallback；ingestion 启动默认校验启用 source 的 ClickHouse event/property 表存在，也可用 `ANALYTICS_SERVICE_CLICKHOUSE_AUTO_MIGRATE=true` 在本地/小部署创建当前 runtime config 内所有启用 source 的 routed tables 后再校验；README 本地运行示例已对齐当前 SourceConfig，包含 `visit_salt` 和 `visit_window_seconds` | 保持 collect/runtime/worker 回归；部署级 schema migration/rollback 规范留到真实上线后再评审 |
| P1-005C | SaaS control-plane config runtime enforcement | 已完成 | 已新增 HTTP source resolver：`simpletrack-anaysitics-service` 可用 `ANALYTICS_SERVICE_SOURCE_RESOLVER=http`、控制面 URL 和 bearer token 按 write key 读取 runtime source config，并用短 TTL 缓存 + `ETag` 条件重验证降低热路径控制面压力且避免 stale auth state；控制面读取默认要求 HTTPS，只有本地 loopback 明确 opt-in 才允许 HTTP；同进程 ingestion 会把 HTTP 返回的 source 绑定到启动 schema surface；`simpletrack-saas` 内部 runtime-source API 已落地，runtime source config 必须携带 server-only `session_salt`、`visit_salt`、`visit_window_seconds` 和 `client_hash_salt`，不能由公开 write key 派生；`Websites` 页现已使用真实组织 Website 数据并提供最小 source 创建、编辑、enable/disable 和 delete 入口，编辑动作可回写 `allowedOrigins`、`bot user agents`、`internal CIDRs`、`internal IPs` 与 `includeClientFingerprint`，并已将这些 runtime source 字段做成可见设置项；disabled source 不再占用 active source limit，创建与启用动作都通过 serializable 事务 helper + `P2034` 重试守住 active source limit；P1 quota 策略已明确只执行 active website source 上限，事件量、留存期、成员数和用量计费不提前进入 P1；`analytics-service` 现在也有 handler 级回归覆盖，验证 cached source 在控制面 disable/delete 后会立即拒绝 collect/readback | salt 轮换、runtime source 变更审计和更细的配置历史放 P1.5/P2 |
| P1-005D | Events / Realtime 查询 API | 已完成 | `simpletrack-anaysitics-service` 已提供内部 `/v1/realtime` 与 `/v1/events` 读接口，使用 bearer token 保护并映射到 `analytics-core` 的 `EventReader`；两个默认路径保留，同时新增 `ANALYTICS_SERVICE_EVENTS_PATH` 和 `ANALYTICS_SERVICE_REALTIME_PATH` 可配置路由；Swagger UI / OpenAPI 文件已随 Fiber runtime 接入；`simpletrack-saas` Realtime / Events 页面已用 server-side readback helper 和 client-safe Website selector 接入，按组织查启用 Website 后使用 write key 调内部服务，未配置、无启用 source 或服务错误时 fail-soft；Events 现在已补 `event_name`、`distinct_id`、`visit_id`、`limit`、`offset`、`sort_field`、`sort_direction` 白名单、`hasMore` 分页和 `30m / 6h / 24h / 7d` 时间窗口预设；query token 轮换已通过 `ANALYTICS_SERVICE_QUERY_TOKENS_JSON` 支持当前 token + 上一版 token 并行接受，并可附带 `id`、`not_before`、`expires_at` 元数据；运行时会拒绝过期/未来 token，并对轮换命中和拒绝场景打审计日志；属性过滤已由 source runtime config 的 `allowed_property_filters` 提供白名单来源，并映射到 `analytics-core` typed property filters；读侧现在使用持久化 `visit_id`，不再用 `session_id` 兜底成 visit；OpenAPI 已把 `property_filter` 标为 repeatable array，并补齐 `event_name` 排序枚举；当前已补组合查询回归，覆盖 SaaS 请求序列化、service query mapping 和 core ClickHouse query plan 三层的标量过滤、`visit_id`、分页、排序、时间窗口和 repeatable property filters 同时存在 | 更复杂的聚合分析、Breakdown/Compare/Funnels/Journeys 维持到 P1.5/P2 |
| P1.5-001 | ClickHouse 读侧优化与属性治理 | 待完成 | 先做属性治理和 query plan 约束，保持 `EventQueryBuilder` / `EventReader` 为唯一读侧入口；热点明细路径再选择性引入 projection，稳定指标再落 materialized view 或小时聚合表 | P1 闭环后按长期分层路线实施 |
| P2-001 | Performance metrics 采集与查询 | 暂缓 | Umami tracker 可采集 LCP、INP、CLS、FCP、TTFB；SimpleTrack P1 不以性能诊断为阻塞项 | P2 评审是否作为事件类型、属性组或独立 performance 模型进入 `analytics-core` |
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
- `src/analytics-service` 已补 `P1-005D` 内部读回放入口：`/v1/realtime` 和 `/v1/events` 复用 `analytics-core` 的 `EventReader`，通过内部 bearer token 保护，并支持 Fiber CORS preflight；Events 侧已补事件名、distinct id、排序、分页和 typed property filters 白名单；内部 query token 已支持短窗口轮换 allowlist、结构化生效/过期时间和命中/拒绝审计日志；HTTP resolver revalidation 现在还有 handler 级回归，验证控制面 disable/delete 后 collect/readback 立即撤权；query routes 现在可配置，Swagger UI / OpenAPI 文件已接入。
- `src/simpletrack-saas` 已补 Realtime / Events 页面服务端读回放 helper 和 client-safe Website 选择器初版：内部 query token 不下发到浏览器，页面按当前组织启用 Website 的 write key 调 `simpletrack-anaysitics-service`，并覆盖未配置服务、无启用 source、禁用 source 和服务异常空态；Events 页面还支持白名单分页、上一页 / 下一页、`30m / 6h / 24h / 7d` 时间窗口预设，以及重复 query 参数/空页偏移的服务端硬化，不会因重复参数把 RSC 页面打崩。
- Umami 源码深解已经转化为 `analytics-core` 优化计划：事件属性与用户属性模型、client info enrich、bot/IP 过滤、session/visit resolver、查询白名单、Realtime/Events 验收、Web tracker SDK、ClickHouse 读侧优化和 performance metrics 均已进入计划表或评审表。
- `simpletrack-enterprise-mvp` 的页面 contract 已从旧 `simpletrack-umami-inspired` 收口到当前原型页集合，`team` / `funnels` / `insights` 不再作为 P1 页面前置假设。

- `analytics-core` / `simpletrack-anaysitics-service` 的 `visit_id` 已从 readback 临时派生升级为写入前确定、入库存储、读侧直接读取；后续只评审 salt 轮换、cookie/no-cookie、server identity、Sessions 专页和 retention 产品化。

正在推进：

- Supastarter for Next.js 的 1 天 SimpleTrack spike：已创建独立工作副本并推送远端，已完成 Websites、Realtime、Events 组织内页面挂载；其中 Websites 已从 UI-only gate 前进到真实 source 列表 + 最小创建表单，marketing 文案、pricing 语义、docs/quickstart、mail-preview 和浏览器截图验证也已完成。
- `analytics-core` P1 数据管道已收口：collect handler、Fiber `POST /collect` 适配器、属性入口约束、typed property row 逻辑展开、ClickHouse property batch writer、`PropertyIndexingEventWriter` 热路径组合、MySQL `property_indexing_status` guard、表路由契约、ClickHouse native batch writer、GORM/MySQL ingestion status guard、Realtime/Events query builder、typed property filter、ClickHouse query reader、worker 边界、本地运行依赖、最小端到端验证、Events 排序/过滤白名单、P1-002B/C collect pre-queue stage 和持久化 `visit_id` 链路已完成；browser / OS / device 派生、geo enrichment、internal traffic 产品配置和过滤审计边界也已补齐，core 继续保持 Go library 方式被 `simpletrack-anaysitics-service` 引用。
- `simpletrack-anaysitics-service` 主线（本地仓库 `src/analytics-service`）：已完成本地仓库、服务骨架、Fiber runtime app、memory / HTTP runtime config resolver、`/collect` 运行时校验、`/tracker.js` 静态托管、collect 单测、Redis durable enqueue、可选 ingestion worker 装配、本地/小部署 ClickHouse routed table auto migration，以及 `P1-005D` 内部 Events / Realtime 查询入口、query routes 配置、Swagger UI / OpenAPI 文件、query token 轮换 allowlist、结构化生命周期、source-scoped 属性过滤白名单、`visit_id` 持久字段读取和 control-plane revoke handler 回归；`simpletrack-saas` 内部 runtime-source API、Websites 真实 source 管理最小闭环（create/list/update/enable/disable/delete）、visit resolver runtime 配置、Realtime/Events 页面读回放、visit_id 可见列、页面级回归、client-safe Website selector、Events 时间窗口预设、repeatable property_filter 多条件查询和 P1 active-source quota 策略已落地，复杂聚合分析维持到 P1.5/P2。
- `analytics-core`、`simpletrack-anaysitics-service` 和 `simpletrack-saas` 近期又补了 Events 组合查询回归：同一条请求里同时携带 `event_name`、`distinct_id`、`visit_id`、分页、排序、时间窗口和 repeatable `property_filter` 时，SaaS 会保持完整 readback state，服务端会把同一组条件送入 `analytics-core`，core 会生成参数化 ClickHouse query plan。
- `analytics-core` 的 ClickHouse 读侧优化已经从“预研名词”细化成可执行分层路线：属性治理和 query plan 约束先做，热点明细路径再择机引入 projection，稳定指标再落到 materialized view 或小时聚合表。
- `simpletrack/prototype/simpletrack-enterprise-mvp/` 继续作为 P0/P1 评审原型，后续围绕当前页面集合补齐更清晰的 contract / mock / production 映射。
- `xwl_bi` 后端只读临时快照已就位，主要用于参考后端架构设计：模块边界、启动装配、消费链路、ClickHouse 写入/查询分层、元数据流转和分析服务拆分。
- Umami 官方源码只读快照已就位，主要用于参考分析对象体系、tracker 采集、事件属性、Realtime/Events 读侧、ClickHouse 明细与聚合模型；P1 数据管道源码分章节深解和 Q&A 概念解释已落地到 `simpletrack/docs/umami/docs/源码实现参考/`。
- Umami 源码启发的 `analytics-core` 优化项已排期并部分落地：P1 已补属性入库/查询、client enrich 第一版、session resolver 第一版、查询安全、端到端验收、浏览器 SDK 最短链路、DNT opt-in 和 UTM/click id 白名单；P1.5/P2 再评审 ClickHouse 聚合优化、多语言 SDK、visit 扩展、SDK CDN 发布和 performance metrics。
- 企业分析控制台 UI 可改造性确认。
- 产品官网 / Marketing Site / docs 公开站点的信息架构已按 P1 验收完成，后续只做轻量优化。

下一步：

1. 继续推进 P1-005：在组合查询已打通的基础上，继续补更复杂的聚合查询交互，保留 ClickHouse 本地/小部署建表与启动校验边界；生产迁移/回滚逻辑等真实上线后再单独评审；quota 当前只执行 active website source 上限。
2. 把过滤统计、salt 轮换、cookie/no-cookie、server identity、Sessions 专页和 retention 产品化维持在 P1.5/P2，除非明确重开范围。
3. 把 R3-U1/R3-U2 的剩余项降为 P1.5/P2：属性字典治理、ambiguous `property_indexing_status=processing` 恢复策略、ClickHouse 读侧优化中的 projection / materialized view / 小时聚合表落地。
4. 在需要 authenticated SaaS 流程时，用 `src/simpletrack-saas/docker-compose.yml` 启动本地 PostgreSQL，验证登录、组织和真实 subscription gate 依赖。
5. 公开站点继续使用 Supastarter 的 marketing/docs app，后续只做轻量文案和视觉微调。
6. 每次子仓库提交推送后，先提交子仓，再更新父仓 gitlink 和实施进度文档；`src/analytics-service` 现已具备远端和子模块登记条件，后续沿用与 `analytics-core`、`simpletrack-saas` 相同的收口顺序。

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
