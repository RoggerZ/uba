# SimpleTrack 研发任务拆解

> 用途：把 [SimpleTrack产品需求草案.md](./SimpleTrack产品需求草案.md) 和 [SimpleTrack实施路线图.md](./SimpleTrack实施路线图.md) 拆成可进入研发排期的 Epic、任务、依赖和验收口径。

## 1. 拆分原则

1. 先做能证明数据进入系统的能力。
2. 先做用户第一天会用到的路径。
3. 先给清楚的空态、示例态、受限态，再做复杂后台能力。
4. 任何创建、删除、分享、邀请类动作都后置到权限和审计边界清楚之后。

## 2. Epic 总览

| Epic | 目标 | 阶段 | 优先级 |
| --- | --- | --- | --- |
| E01 安装与工作区接入 | 用户登录后能立刻完成安装并回查安装信息 | 阶段 1 | P0 |
| E02 事件入库与 Raw Events | 用户能证明真实事件已经进入系统 | 阶段 1 | P0 |
| E03 Product 空态、示例态、真实态 | 用户在没数据和有数据时都能理解 Product 页 | 阶段 1 | P0 |
| E04 Marketing 与 UTM | 用户能从来源分析继续生成投放链接 | 阶段 2 | P1 |
| E05 Reports 模板与样张 | 用户能理解报告价值，但不先做正式生成 | 阶段 2 | P1 |
| E06 AI 任务模板 | 用户知道 AI 可以问哪些分析问题 | 阶段 2 | P2 |
| E07 治理层基础能力 | 用户能理解域名、IP、bot 噪音治理入口 | 阶段 3 | P1 |
| E08 分享、成员与受限态 | 用户能区分外部只读分享和内部协作 | 阶段 3 | P1 |
| E09 商业化门槛表达 | 用户知道哪些能力未解锁以及下一步动作 | 阶段 4 | P1 |

## 3. MVP 任务拆解

### E01 安装与工作区接入

| 任务 | 内容 | 依赖 | 验收 |
| --- | --- | --- | --- |
| T01 默认安装落点 | 登录后默认进入安装页 | 账号 / workspace 基础路由 | 新用户登录后不是空 dashboard |
| T02 Script 安装卡片 | 展示脚本代码、Workspace ID、复制入口 | workspace id 可用 | 用户能复制脚本 |
| T03 GTM 安装卡片 | 展示 GTM 接入代码和说明 | 同 T02 | 用户能切换 Script / GTM |
| T04 Settings 回查 | Settings 中再次展示 Workspace ID 和脚本 | Settings 页面 | 用户能从设置页回查安装信息 |
| T05 Verify Installation | 展示验证入口和状态反馈 | tracker 入库能力 | 成功、失败、等待三种状态清楚 |

测试建议：

- 新 workspace 登录后默认路由正确。
- Script / GTM tab 切换不丢失 Workspace ID。
- 复制按钮复制内容不带敏感账号信息。
- 验证失败时有下一步说明。

### E02 事件入库与 Raw Events

| 任务 | 内容 | 依赖 | 验收 |
| --- | --- | --- | --- |
| T06 事件接收接口 | 接收 pageview 和自定义事件 | tracker / broker | 能保存事件名、metadata、domain、userAgent |
| T07 Raw Events 表 | 展示事件明细表 | T06 | 能看到事件名、metadata、时间、domain |
| T08 事件排障空态 | 无事件时解释如何发送首个事件 | T07 | 空态能引导回安装页或文档 |
| T09 最小 demo 页面 | 提供本地或内置 demo 触发事件 | T06 | demo 能发出至少一个事件 |

测试建议：

- 发送单条事件后 Raw Events 可见。
- metadata JSON 格式错误时有清楚反馈。
- Raw Events 不展示邮箱、token 等敏感字段。

### E03 Product 空态、示例态、真实态

| 任务 | 内容 | 依赖 | 验收 |
| --- | --- | --- | --- |
| T10 Product 空态骨架 | 展示 Top Events、Funnel、User Flow、Metadata 区块 | 基础页面壳 | 没数据时仍能看懂未来结构 |
| T11 Show test data | 提供示例数据开关 | T10 | 示例态有清楚标识 |
| T12 Product 聚合 | 聚合真实事件到 Top Events | T06 | Raw Events 有数据后 Product 能显示 |
| T13 Setup events CTA | 跳转或打开事件接入说明 | T01 / T06 | CTA 文案说明真实去向 |

测试建议：

- 空态、示例态、真实态互不混淆。
- 示例态不会写入真实数据。
- `Setup events` 如果打开外链，文案要明确。

## 4. 增强版任务拆解

### E04 Marketing 与 UTM

| 任务 | 内容 | 依赖 | 验收 |
| --- | --- | --- | --- |
| T14 Marketing 空态 | 展示渠道、来源、社交模块骨架 | 基础事件 / pageview | 无来源数据时结构清楚 |
| T15 Marketing 示例态 | 提供示例渠道数据 | T14 | 能解释来源分析价值 |
| T16 UTM 生成器 | 在 Marketing 页内生成 UTM link | T14 | 不跳文档也能生成链接 |
| T17 UTM 字段采集 | 接收并展示 source / medium / campaign | T06 | 带 UTM 访问能进入来源分析 |

### E05 Reports 模板与样张

| 任务 | 内容 | 依赖 | 验收 |
| --- | --- | --- | --- |
| T18 Reports 模板中心 | 展示报告类型和周期选择 | Product / Marketing 基础数据 | 用户能理解报告类别 |
| T19 Sample 预览 | 打开报告样张 | T18 | 不生成正式报告也能看样张 |
| T20 Generate disabled reason | 正式生成不可用时说明原因 | T18 | disabled 状态解释套餐、权限或缺少选择 |

### E06 AI 任务模板

| 任务 | 内容 | 依赖 | 验收 |
| --- | --- | --- | --- |
| T21 AI 默认页 | 提供分析助手入口 | Product / Marketing 页面 | 不是空聊天框 |
| T22 示例问题 | 提供趋势、漏斗、SEO、UTM、报告 prompt | T21 | 用户能一键使用示例问题 |
| T23 AI 结果占位 | 未接入模型时解释能力边界 | T21 | 不伪装成已可用 |

## 5. 后置任务拆解

### E07 治理层基础能力

| 任务 | 内容 | 依赖 | 验收 |
| --- | --- | --- | --- |
| T24 Domains 治理页 | 展示 domain data 和 sanitization 入口 | 事件域名字段 | 页面态清楚，默认不执行删除 |
| T25 IP 排除页 | 展示 IP 排除配置 | 采集层过滤规则 | 用户能理解用途 |
| T26 Bot traffic 页 | 展示 bot 治理配置 | bot 识别规则 | 用户能理解策略状态 |
| T27 高风险确认 | 删除 / sanitization 前确认 | T24 | 明确影响范围和不可逆风险 |

### E08 分享、成员与受限态

| 任务 | 内容 | 依赖 | 验收 |
| --- | --- | --- | --- |
| T28 Shareable links 空态 | 展示外部只读分享入口 | dashboard / report 可分享对象 | 未创建时显示 0 active |
| T29 Share link 创建确认 | 创建前说明访问范围 | T28 | 不会误创建外部链接 |
| T30 Members 受限态 | 明确是权限、套餐还是系统异常 | 用户权限模型 | 不用 loading 代替原因 |
| T31 Members 邀请 | 邀请成员、角色、状态 | T30 | owner 权限下可邀请 |

### E09 商业化门槛表达

| 任务 | 内容 | 依赖 | 验收 |
| --- | --- | --- | --- |
| T32 Premium gate | 高级能力锁定页 | 计划 / 权限状态 | 有价值说明和 Upgrade CTA |
| T33 Plans / FAQ | 套餐页解释边界 | 计费配置 | 用户能理解 free trial 和 Business 差异 |
| T34 Formal report generation | 正式 PDF 生成 | Reports 模板、权限、生成服务 | 解锁后能生成且可下载 |

## 6. 依赖关系

| 后续能力 | 前置条件 |
| --- | --- |
| Product 聚合 | 事件接收和 Raw Events |
| Marketing 真实来源 | UTM 字段采集 |
| Reports 生成 | Product / Marketing 数据、报告模板、权限 |
| AI 分析 | Product / Marketing / Reports 的结构化数据 |
| Shareable links | 可分享对象和权限模型 |
| Members | workspace 角色模型 |
| Domain sanitization | 域名数据、风险确认、审计记录 |

## 7. Definition of Done

一个任务只有同时满足这些条件，才算完成：

- 页面态、空态、错误态至少覆盖当前任务的主路径。
- 涉及数据的任务有最小真实样本或明确 mock / 示例标识。
- 受限态必须说明原因和下一步动作。
- 不把账号、token、workspace id、邮箱写进文档、日志或截图。
- 新增截图或文档后运行 `python validate_litlyx_assets.py`。

## 8. 不建议提前做

- 在 Raw Events 可用前做复杂 Reports。
- 在 UTM 字段稳定前做渠道归因。
- 在权限模型稳定前做 Members 邀请。
- 在撤销机制清楚前做 Shareable links 创建。
- 在审计和确认机制缺失时做数据删除或 sanitization。

## 9. 关联材料

- [SimpleTrack产品需求草案.md](./SimpleTrack产品需求草案.md)
- [SimpleTrack实施路线图.md](./SimpleTrack实施路线图.md)
- [落地评审清单.md](./落地评审清单.md)
- [功能打通矩阵.md](./功能打通矩阵.md)
- [执行与复验手册.md](./执行与复验手册.md)
