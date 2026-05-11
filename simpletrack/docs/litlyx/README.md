# Litlyx Research Workspace

这个目录用于沉淀 Litlyx 后台的功能调研资产，目标不是“照抄 Umami”，而是沿用那套可复核的工作方式，把页面态、关键交互态、结构化索引和产品解读放在同一处。

## 结构

- `Litlyx功能深度分析.md`：主调研文档，解释截图代表什么、每个模块解决什么问题、哪些地方值得给 SimpleTrack 借鉴。
- `Litlyx功能矩阵.md`：把 IA、模块、关键工作流、权限边界和 SimpleTrack 借鉴优先级整理成查表式决策材料。
- `docs/`：专题解读和 playbooks，把现有证据压成更适合范围评审与方案讨论的材料。
  - 其中 `docs/落地评审清单.md` 可直接作为产品评审核对表。
  - `docs/SimpleTrack实施路线图.md` 把 Litlyx 结论转成 SimpleTrack 的阶段实施路径。
  - `docs/SimpleTrack产品需求草案.md` 把调研结论整理成可评审的 PRD 起点。
  - `docs/SimpleTrack研发任务拆解.md` 把 PRD 拆成 Epic、任务、依赖和验收口径。
  - `docs/功能打通矩阵.md` 适合快速判断哪些能力已经打通、哪些仍停在页面态或受限态。
  - `docs/执行与复验手册.md` 适合后续继续接力这轮调研。
  - `docs/数据模型与事件字典.md` 记录当前真实发送的事件、metadata 和 payload 边界。
  - `docs/事件与属性存储模型源码分析.md` 直接回答 Litlyx 的 event / metadata / visit / session 怎么存。
  - 如果你要把 Litlyx 和 `analytics-core`、`xwl_bi`、Umami 的存储模型并排比较，直接看 `../事件与属性存储方案对比.md`。
  - `docs/真实业务数据方案.md` 规划下一阶段如何补厚 Marketing / Reports / AI 样本。
  - `docs/00` 到 `docs/05` 这组专题文档按模块拆开了能力地图、安装接入、Product、增长闭环、治理协作和门槛升级表达。
- `快照索引.md`：正式快照和参考快照的统一入口。
- `快照进度.md`：当前采集完成度、缺口和产品发现。
- `snapshots/phase-*`：按阶段归档的正式快照，每个阶段目录都维护 `README.md` 和 `flow.md`。
- `litlyx-reference/`：首轮稳定全页基线截图，用于补齐那些重复访问时会偶发 loading 的页面。
- `tracking-demo/`：本轮用于真实回写 Litlyx 浏览器事件和批量事件的最小验证工件。
- `capture_litlyx_research.cjs`：本轮用于批量补拍的辅助脚本，运行时从环境变量读取登录信息。
- `validate_litlyx_assets.py`：本地静态校验脚本，用于检查 Markdown 链接、快照索引、UTF-8 和敏感信息。
- `snapshot-contact-sheet.png`：用于快速人工复核截图脱敏和覆盖面的检查图，不计入正式快照编号。

## 快照规则

- 正式快照以 `快照索引.md` 为准；只有进入索引的编号才算正式资产。
- 阶段快照命名使用 `Pxx-Sxx`，参考快照使用 `REF-xx`。
- 对理解有帮助的弹窗、切换态、示例态，单独截图，不和页面默认态混在一张图里。
- 如果路由已进入目标页面，但界面长时间只显示 loading、空白或无反馈，要把它明确记录为产品发现，而不是略过。
- 截图入档前要遮挡登录邮箱、密码输入内容、workspace id、分享链接 token 和侧栏账号区域。

## 历史散图说明

`snapshots/` 根目录下已有的 `img*.png` 属于早期探索性抓图，本轮不作为正式编号资产使用。它们已经完成遮挡，并在 `快照索引.md` 的“历史非正式文件”里登记，当前正式资产仍统一收敛到 `snapshots/phase-*`。

## 本轮范围

- 已完成：登录、工作区列表、安装入口、Product / Marketing / Reports / SEO / Shields / AI / Plans / Shareable links / Members 的主要页面采集。
- 已完成：本地 `tracking-demo` 真实埋点回放、`Reports` 的 Sample 预览态，以及 `Product` 页 `Setup events` 外链行为确认。
- 已完成：`Settings / Domains` 的域名数据清理入口复核，以及 `Shareable links` 的 0 active 只读分享页复核。
- 受限未执行：当前 Free trial 下正式报告卡片和 `Generate report` 均为 disabled，因此未触发 `Marketing Report / Product Report / SEO Report` 等 PDF 生成 POST 动作。

## 本地校验

每次补充截图或文档后，优先运行：

```bash
python validate_litlyx_assets.py
```
