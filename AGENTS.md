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
- `src/analytics-core` 或 `src/simpletrack-saas` 有变更时，必须先提交并推送子仓库，再更新父仓子模块 gitlink、相关文档和父仓提交。
- 已标记 `已完成` 的任务如果进入功能重构、范围重开、验收失败或实现被替换，必须把状态重置为 `待完成`，并在修订记录中说明原因。
- P1 已确定包含 `analytics-core` 独立核心仓库建设：仓库名只用 `analytics-core`，不得带 `simpletrack` 或 `xwl`；从 xwl_bi 抽取分析数据面核心，保留 KafkaBus，前期优先 Redis Stream，不复用旧 Vue2 后台界面。
- P1 已确定包含产品官网 / Marketing Site / 公开站点：需要产品介绍、定价/订阅入口、docs/quickstart；不要把它仅理解为单张 landing page。
- SimpleTrack 生产 SaaS 模板已确定先选择 Supastarter for Next.js；MakerKit 只作为 B2B 企业控制面对照和备选，除非用户明确重开选型，不要在两者之间反复摇摆。
- SimpleTrack 支付路线先按 Supastarter 已支持的 Stripe、Lemon Squeezy、Polar、Creem、Dodo Payments provider 接入；KYC/KYB、退款、拒付、发票、税务和费用结构放到上线收费前逐项处理，不作为 P0/P1 早期阻塞。
- `analytics-core` 的实施方案维护在 `simpletrack/docs/实施决策/analytics-core实施方案.md`；每次修改其模块边界、EventBus、命名映射、存储模型或验收标准时，必须同步更新实施决策 README 的修订记录和实施计划完成列表。
- `analytics-core` 和 SimpleTrack 分析产品参考采用“双参考”：Umami 用于分析对象体系、事件语义、Realtime/Events/Funnels/Journeys/Retention/Segments 边界；Litlyx 用于短接入链路、Raw Events 验收、Product 空态/示例态/真实态和 Show test data 教育方式。

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
- `src/analytics-core` 和 `src/simpletrack-saas` 是独立子仓库，推送到 `simpletrack` GitHub 组织时必须使用专用 SSH 配置：
  - `git -C "C:\Users\admin\Documents\src\uba\src\analytics-core" config core.sshCommand "ssh -F C:/Users/admin/.ssh/config_simpletrack"`
  - `git -C "C:\Users\admin\Documents\src\uba\src\simpletrack-saas" config core.sshCommand "ssh -F C:/Users/admin/.ssh/config_simpletrack"`
- 两个子仓库的 `origin` 必须使用 `github-simpletrack` Host 别名：
  - `git@github-simpletrack:simpletrack/analytics-core.git`
  - `git@github-simpletrack:simpletrack/simpletrack-saas.git`
- 不要依赖默认 `C:\Users\admin\.ssh\config` 推送这两个仓库；该文件曾因 Windows ACL 权限异常导致 OpenSSH 报 `Bad owner or permissions`。
- 相关说明维护在 `simpletrack/docs/Q&A/Windows-SSH仓库权限怎么配置.md`。

## 截图评审标准

- 优先同时保留页面态和关键交互态
- 只要弹窗、下拉、tab 切换、筛选条件变化会影响理解，就要单独截图
- 如果请求已经成功但页面仍然空白，必须把它作为产品发现明确记录，不能模糊处理

## 中文文档安全

- 修改中文文档时保持 UTF-8
- 文本编辑优先使用 `apply_patch`
- 避免使用不安全的 PowerShell 中文读写方式
