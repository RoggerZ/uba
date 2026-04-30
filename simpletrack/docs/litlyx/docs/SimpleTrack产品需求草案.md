# SimpleTrack 产品需求草案

> 用途：把 Litlyx 调研结论转成 SimpleTrack 可讨论的 PRD 草案。它不是最终需求冻结版，而是用于产品评审、研发拆分和设计确认的起点。

## 1. 背景

Litlyx 这轮调研显示，一个 analytics 产品的第一价值不应该是“展示很多报表入口”，而应该是让用户尽快完成三件事：

1. 成功接入
2. 确认首批真实数据
3. 明白下一步可以怎么分析和行动

SimpleTrack 的第一版可以优先借鉴 Litlyx 的短接入链路、示例数据、Raw Events 验收、Marketing + UTM 闭环、Reports 样张，以及清楚的受限态表达。

## 2. 目标

### 产品目标

- 让新用户在首次进入后能快速找到安装入口。
- 让用户能通过 Raw Events 和 Product 确认真实事件已经入库。
- 让空态页面也能解释未来价值，而不是只显示“暂无数据”。
- 让增长、报告、AI 和治理能力有清晰但不过度膨胀的演进路径。

### 非目标

- 不在第一版实现完整 PDF 报告生成。
- 不在第一版实现复杂团队权限。
- 不在第一版自动创建外部分享链接。
- 不在第一版执行高风险数据删除或 sanitization。

## 3. 目标用户

| 用户 | 主要诉求 | 第一版要满足什么 |
| --- | --- | --- |
| 创始人 / 独立开发者 | 快速接入并知道有没有数据 | 安装入口、验证状态、Raw Events |
| 增长负责人 | 看来源、做 UTM、看报告样张 | Marketing、UTM 生成器、Reports 模板 |
| 产品经理 | 看事件、漏斗骨架、用户行为线索 | Product 骨架、示例数据、事件明细 |
| 工程师 | 排查脚本是否工作、事件是否入库 | Script / GTM、Workspace ID、Raw Events |

## 4. 范围

### MVP 范围

| 模块 | 需求 | 优先级 |
| --- | --- | --- |
| Onboarding | 登录后默认进入安装入口 | P0 |
| Install | 支持 Script 和 GTM 两种接入方式 | P0 |
| Settings | 可回查 Workspace ID 和安装代码 | P0 |
| Verify | 有安装验证入口和明确状态反馈 | P0 |
| Product | 空态展示 Top Events、Funnel、User Flow、Metadata 骨架 | P0 |
| Raw Events | 展示事件明细，用于确认真实入库 | P0 |
| Demo mode | 支持 `Show test data`，但必须标记为示例数据 | P0 |
| Marketing | 展示来源和渠道结构的基础页 | P1 |
| UTM | 在 Marketing 页内提供 UTM 生成器 | P1 |
| Reports | 展示报告模板和 Sample 样张 | P1 |
| AI | 提供任务型 prompt 模板 | P2 |
| Gate | 对未解锁能力给出明确原因和下一步动作 | P0 |

### 后置范围

| 模块 | 后置原因 |
| --- | --- |
| 正式 PDF 生成 | 依赖报告模板、权限和生成链路稳定 |
| Shareable links 创建 | 会改变外部访问状态，需要权限和撤销模型 |
| Members 邀请 | 依赖团队权限模型 |
| Domain data 删除 | 高风险动作，需要确认、审计和回滚策略 |
| Domain sanitization | 高风险治理动作，需要明确授权和执行日志 |

## 5. 用户故事

| 编号 | 用户故事 | 验收标准 |
| --- | --- | --- |
| US-01 | 作为新用户，我登录后想立刻看到安装方式 | 默认落点是安装页，不是空 dashboard |
| US-02 | 作为工程师，我想复制脚本或 GTM 代码 | 页面提供 Script / GTM 两种入口 |
| US-03 | 作为工程师，我想稍后回查 Workspace ID | Settings 中能再次看到安装信息 |
| US-04 | 作为新用户，我想知道事件是否真的进来了 | Raw Events 能显示至少一条真实事件 |
| US-05 | 作为产品经理，我想在没数据时理解未来页面价值 | Product 空态展示分析骨架和说明 |
| US-06 | 作为增长负责人，我想生成 UTM 链接 | Marketing 页内可打开 UTM 生成器 |
| US-07 | 作为业务负责人，我想先看报告样张 | Reports 中可打开 Sample 预览 |
| US-08 | 作为用户，我想知道为什么某功能不能用 | disabled 或 premium gate 必须说明原因 |

## 6. 功能需求

### 6.1 安装入口

必须支持：

- Script 安装代码
- GTM 安装代码
- Workspace ID 展示
- 安装验证入口
- 回到 Settings 后可再次找到安装信息

验收方式：

- 首次登录后直接进入安装页。
- 用户能在 2 个页面内找到 Workspace ID。
- 安装验证失败时不只显示空状态，要给下一步提示。

### 6.2 Product 与 Raw Events

必须支持：

- Product 空态骨架
- 示例数据开关
- 真实事件聚合
- Raw Events 明细表

验收方式：

- 没数据时用户仍能看到未来分析结构。
- 打开示例数据后，页面能解释图表含义。
- 真实事件入库后，Raw Events 和 Product 都能看到结果。

### 6.3 Marketing 与 UTM

必须支持：

- 来源 / 渠道基础结构
- Marketing 示例态
- UTM 生成弹窗或面板

验收方式：

- 用户能从来源分析页面直接进入 UTM 生成。
- UTM 生成器不需要跳到文档才能使用。

### 6.4 Reports 与 Sample

必须支持：

- 报告模板列表
- Sample 样张预览
- 正式生成受限时的原因说明

验收方式：

- 用户能在不生成正式报告的情况下看到样张。
- disabled 状态说明是权限、套餐还是缺少选择条件。

### 6.5 受限态与升级

必须支持：

- premium gate
- disabled reason
- Upgrade CTA
- Plans / FAQ 对能力边界的解释

验收方式：

- 不用 loading 代替权限说明。
- 不用灰态按钮代替原因说明。
- 用户能知道下一步该升级、联系 owner，还是补充配置。

## 7. 数据与隐私要求

- 不在 analytics payload 中写入邮箱、密码、token。
- Workspace ID 不写入仓库文档或截图。
- 示例数据必须和真实数据有清楚标识。
- 高风险治理动作必须有确认和审计边界。

## 8. 关键指标

| 指标 | 解释 |
| --- | --- |
| 安装完成率 | 用户从进入安装页到 tracker 成功加载的比例 |
| 首事件达成率 | 用户完成首个真实事件入库的比例 |
| Raw Events 使用率 | 用户在接入期查看 Raw Events 的比例 |
| 示例数据开启率 | 用户用示例数据理解产品的比例 |
| UTM 生成率 | Marketing 页内生成 UTM 的比例 |
| Sample 预览率 | Reports 样张被打开的比例 |
| 升级 CTA 点击率 | premium gate 和 disabled 态后的下一步动作 |

## 9. 依赖

- 事件命名和 metadata 字段必须先稳定。
- Raw Events 明细能力要早于复杂报告。
- Marketing 样本需要 UTM 字段输入。
- Reports 正式生成依赖模板、权限和生成服务。
- Shareable links 和 Members 依赖权限模型。

## 10. 风险

| 风险 | 缓解 |
| --- | --- |
| 用户把示例数据误认为真实数据 | 示例态必须有明确标识 |
| 用户以为外链文档入口无响应 | CTA 文案说明会打开文档 |
| disabled 状态造成困惑 | 给出具体原因和下一步动作 |
| 分享链接误创建 | 创建前必须确认影响范围 |
| 治理动作误删数据 | 删除和 sanitization 必须有确认、审计和权限限制 |

## 11. 关联材料

- [SimpleTrack实施路线图.md](./SimpleTrack实施路线图.md)
- [落地评审清单.md](./落地评审清单.md)
- [功能打通矩阵.md](./功能打通矩阵.md)
- [数据模型与事件字典.md](./数据模型与事件字典.md)
- [05-门槛、受限态与升级表达.md](<./05-门槛、受限态与升级表达.md>)
