# SimpleTrack MVP 执行 Backlog

> 用途：把 Litlyx 调研结论压成可以直接进入研发排期、验收和复验的最小执行清单。它不是完整 PRD 的替代品，而是从 [SimpleTrack产品需求草案.md](./SimpleTrack产品需求草案.md)、[SimpleTrack实施路线图.md](./SimpleTrack实施路线图.md) 和 [SimpleTrack研发任务拆解.md](./SimpleTrack研发任务拆解.md) 中截出的第一阶段交付面。

## 1. 截断原则

MVP 只解决一个核心问题：用户能证明数据已经进入 SimpleTrack，并能理解第一批数据下一步怎么分析。

因此第一阶段只收敛到四类能力：

| 能力 | 是否进入 MVP | 原因 |
| --- | --- | --- |
| 安装与接入 | 是 | 没有接入成功，后续分析全部失去基础 |
| Raw Events | 是 | 需要一条最可信的“事件已入库”证据链 |
| Product 基础分析 | 是 | 用户需要看到事件聚合、空态、示例态和真实态之间的关系 |
| 轻量治理提示 | 是 | 内部流量、测试域名和 bot 会影响早期信任 |
| Marketing / UTM | 延后到增强版 | 依赖稳定 pageview 与 UTM 字段输入 |
| Reports 正式生成 | 延后 | Litlyx 当前也处于受限态，MVP 先做样张和 disabled reason |
| AI Analyst | 延后 | 需要结构化分析数据作为输入 |
| Shareable links / Members | 延后 | 涉及权限、外部访问和撤销模型 |
| Domain 删除 / Sanitization | 延后 | 属于高风险治理动作，需要审计与确认机制 |

## 2. 第一批可排期任务

| ID | 任务 | 交付内容 | 依赖 | 验收口径 | 参考证据 |
| --- | --- | --- | --- | --- | --- |
| MVP-01 | 登录后默认进入安装页 | 新 workspace 登录后落到 Web / Install，而不是空 dashboard | workspace 路由 | 新账号第一屏可以直接复制接入脚本 | `P02-S01`、`P02-S02` |
| MVP-02 | Script / GTM 安装卡片 | Script 与 GTM 两种安装方式、复制入口、workspace 标识回查 | workspace id | tab 切换稳定，复制内容不暴露账号或 token | `P02-S02`、`P02-S03` |
| MVP-03 | Settings 安装信息回查 | Settings / General 中能再次找到脚本和 workspace 信息 | MVP-02 | 用户离开 onboarding 后仍能找回安装信息 | `P02-S03`、`REF-06` |
| MVP-04 | 事件接收与 Raw Events 表 | 接收 pageview / 自定义事件并展示事件名、时间、domain、metadata | tracker 入库 | 本地 demo 能发送至少 1 条事件并在 Raw Events 可见 | `tracking-demo/`、`P03-S07` |
| MVP-05 | Raw Events 空态与排障 | 无事件时说明如何发送第一条事件 | MVP-04 | 空态能引导回安装页或事件文档，不是单纯空白 | `P03-S01` |
| MVP-06 | Product 空态骨架 | Top Events、Funnels、User Flow、Metadata 的占位结构 | 页面框架 | 没数据时仍能理解未来分析面 | `P03-S01`、`P03-S02` |
| MVP-07 | Product 示例态 | Show test data / 示例数据，不写入真实数据 | MVP-06 | 示例态有明确标识，能解释分析价值 | `P03-S03`、`P03-S04` |
| MVP-08 | Product 真实态聚合 | Raw Events 有数据后，Product 能显示 Top Events 和 metadata 入口 | MVP-04 | 真实 demo 事件能进入 Product 聚合 | `P03-S08`、`P03-S09` |
| MVP-09 | Setup events 去向说明 | CTA 明确打开文档或站内事件说明 | MVP-06 | 用户知道按钮会去哪里，不误以为页面无响应 | `P03-S05`、`P03-S06` |
| MVP-10 | 受限态说明组件 | 对 Reports / SEO / 高级能力给出 disabled 原因和下一步动作 | 套餐/权限状态 | disabled 不是灰掉了事，能解释是权限、套餐还是缺少选择 | `P04-S07`、`P04-S09`、`P08-S01` |
| MVP-11 | 轻量 Shields 信息页 | Domains、IP addresses、Bot traffic 的只读说明或占位入口 | 采集字段规划 | 用户知道后续如何过滤内部和异常流量 | `P05-S01` 到 `P05-S04` |
| MVP-12 | 资产安全校验脚本 | 新增截图 / 文档后检查链接、索引、UTF-8、敏感信息 | 文档规范 | 本地校验脚本通过，且不写入账号、token、workspace id | `validate_litlyx_assets.py` |

## 3. 不进入 MVP 的任务

这些能力不删除，但不进入第一阶段排期：

| 能力 | 延后原因 | 进入条件 |
| --- | --- | --- |
| Reports 正式 PDF 生成 | Litlyx 当前 Free trial 下正式生成也未解锁，缺少端到端证据 | 已有报告模板、样张、权限模型和生成服务 |
| Shareable links 创建 | 会改变外部访问状态，需要撤销和审计 | 只读分享对象、有效期、撤销、访问日志完成 |
| Members 邀请 | 当前 Litlyx 入口仍表现为受限或 loading，不适合作为 MVP 标准 | workspace 角色模型清楚 |
| Domain data 删除 | 删除 visits/events，风险高 | 审计日志、二次确认、影响范围预览完成 |
| Domain sanitization | 会改变域名数据归属或清理状态 | 域名归并规则、回滚策略、操作日志完成 |
| AI Analyst | 没有足够结构化结果时容易变成空聊天 | Product / Marketing / Reports 输出结构稳定 |

## 4. 推荐实施顺序

1. 先做 `MVP-01` 到 `MVP-03`，确保新用户第一天能完成安装并回查配置。
2. 再做 `MVP-04` 到 `MVP-05`，建立最小可信事件证据链。
3. 接着做 `MVP-06` 到 `MVP-09`，把 Raw Events 转成用户能理解的 Product 分析面。
4. 最后做 `MVP-10` 到 `MVP-11`，补足受限态和治理入口，避免用户把未解锁或未实现误判为故障。
5. 每一批完成后运行 `MVP-12` 对文档和截图资产做静态复核。

## 5. 每个任务的 Definition of Done

- 至少覆盖页面默认态、空态和主要成功态。
- 涉及事件数据的任务必须有最小真实样本，不能只靠 mock。
- 示例数据必须标明是 sample / test data，不能混入真实统计。
- 受限态必须说明原因和下一步动作。
- 高风险动作只展示入口和说明，不在 MVP 中实际执行。
- 截图、文档、复验手册要同步更新，避免结论和证据脱节。
- 不把账号、邮箱、token、workspace id、分享链接密钥写入仓库。

## 6. 研发评审时的硬问题

评审 MVP 范围时建议直接问这些问题：

| 问题 | 通过标准 |
| --- | --- |
| 新用户是否能在 5 分钟内找到脚本并完成第一条事件？ | 安装页、复制入口、验证反馈完整 |
| 事件是否能从 demo 走到 Raw Events？ | 能看到事件名、metadata、domain 和时间 |
| Product 页在没数据时是否仍然有意义？ | 空态不是空白，而是未来分析结构 |
| Product 页在有数据后是否能证明聚合有效？ | Top Events 或 metadata 入口能反映真实事件 |
| disabled 功能是否会被误判为 bug？ | 有原因、套餐或下一步说明 |
| 是否避免了外部分享、删除、邀请等副作用？ | 这些动作只作为后续能力，不在 MVP 中触发 |

## 7. 关联材料

- [Litlyx功能深度分析.md](../Litlyx功能深度分析.md)
- [Litlyx功能矩阵.md](../Litlyx功能矩阵.md)
- [功能打通矩阵.md](./功能打通矩阵.md)
- [执行与复验手册.md](./执行与复验手册.md)
- [SimpleTrack产品需求草案.md](./SimpleTrack产品需求草案.md)
- [SimpleTrack实施路线图.md](./SimpleTrack实施路线图.md)
- [SimpleTrack研发任务拆解.md](./SimpleTrack研发任务拆解.md)
