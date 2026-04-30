# Litlyx 专题解读与 Playbooks

> 目标：把 `simpletrack/docs/litlyx/` 现有截图和主分析，继续压缩成更适合产品评审、方案讨论和范围拆分的专题文档。

## 怎么读这套材料

这套 Litlyx 资料现在分成三层：

- 主证据层：
  - [../Litlyx功能深度分析.md](../Litlyx功能深度分析.md)
  - [../Litlyx功能矩阵.md](../Litlyx功能矩阵.md)
  - [../快照索引.md](../快照索引.md)
- 阶段证据层：
  - `../snapshots/phase-*`
  - 每个阶段目录里的 `README.md` 和 `flow.md`
- 决策与落地层：
  - 当前 `docs/`
  - 更适合回答“这套能力怎么翻译成 SimpleTrack 的产品决策”

如果你现在的目标是：

| 问题 | 先看什么 |
| --- | --- |
| Litlyx 现在到底验证了哪些能力 | [../快照进度.md](../快照进度.md) |
| 某个模块解决什么问题、证据在哪 | [../Litlyx功能矩阵.md](../Litlyx功能矩阵.md) |
| 需要完整理解产品结构和交互意图 | [../Litlyx功能深度分析.md](../Litlyx功能深度分析.md) |
| 需要把能力翻译成 SimpleTrack 的实施路径 | `playbooks/` |
| 需要看 Litlyx 和 Umami 的取舍差异 | [00-Litlyx与Umami取舍对照.md](./00-Litlyx与Umami取舍对照.md) |
| 需要直接开范围评审会 | [落地评审清单.md](./落地评审清单.md) |
| 需要把能力排成实施阶段 | [SimpleTrack实施路线图.md](./SimpleTrack实施路线图.md) |
| 需要把结论转成 PRD 草案 | [SimpleTrack产品需求草案.md](./SimpleTrack产品需求草案.md) |
| 需要拆成研发任务和验收口径 | [SimpleTrack研发任务拆解.md](./SimpleTrack研发任务拆解.md) |
| 需要快速判断哪些能力已经打通 | [功能打通矩阵.md](./功能打通矩阵.md) |
| 需要继续接力执行和复验 | [执行与复验手册.md](./执行与复验手册.md) |
| 需要看当前事件和字段到底长什么样 | [数据模型与事件字典.md](./数据模型与事件字典.md) |
| 需要规划下一阶段怎么补更像业务的数据 | [真实业务数据方案.md](./真实业务数据方案.md) |

## 推荐阅读顺序

1. 先看 [00-Litlyx与Umami取舍对照.md](./00-Litlyx与Umami取舍对照.md)
2. 再看模块专题：
   - [00-产品主张与能力地图.md](./00-产品主张与能力地图.md)
   - [01-安装与接入.md](./01-安装与接入.md)
   - [02-采集、事件与 Product.md](<./02-采集、事件与 Product.md>)
   - [03-Marketing、Reports 与 AI.md](<./03-Marketing、Reports 与 AI.md>)
   - [04-治理、分享、协作与套餐.md](<./04-治理、分享、协作与套餐.md>)
   - [05-门槛、受限态与升级表达.md](<./05-门槛、受限态与升级表达.md>)
3. 再看 [../Litlyx功能矩阵.md](../Litlyx功能矩阵.md)
4. 然后按问题进入对应 playbook：
   - 首次接入和首批数据：`01`
   - 分析到增长闭环：`02`
   - 治理、分享、协作边界：`03`
   - SimpleTrack 能力优先级：`04`
   - 门槛和升级转化：`05`

## 文档边界

- 本目录只写当前仓库里已经有截图、操作流或真实验证支撑的结论。
- 受当前 Free trial 限制，`Reports` 正式 PDF 生成链路、`Shareable links` 创建、`Members` 真正可用态仍未实测。
- 这些限制不会阻止我们做产品判断，但要明确它们属于“已看到入口和门槛”而不是“完整跑通”。

## Playbooks

- [00-Litlyx与Umami取舍对照.md](./00-Litlyx与Umami取舍对照.md)
- [00-产品主张与能力地图.md](./00-产品主张与能力地图.md)
- [01-安装与接入.md](./01-安装与接入.md)
- [02-采集、事件与 Product.md](<./02-采集、事件与 Product.md>)
- [03-Marketing、Reports 与 AI.md](<./03-Marketing、Reports 与 AI.md>)
- [04-治理、分享、协作与套餐.md](<./04-治理、分享、协作与套餐.md>)
- [05-门槛、受限态与升级表达.md](<./05-门槛、受限态与升级表达.md>)
- [落地评审清单.md](./落地评审清单.md)
- [SimpleTrack实施路线图.md](./SimpleTrack实施路线图.md)
- [SimpleTrack产品需求草案.md](./SimpleTrack产品需求草案.md)
- [SimpleTrack研发任务拆解.md](./SimpleTrack研发任务拆解.md)
- [功能打通矩阵.md](./功能打通矩阵.md)
- [执行与复验手册.md](./执行与复验手册.md)
- [数据模型与事件字典.md](./数据模型与事件字典.md)
- [真实业务数据方案.md](./真实业务数据方案.md)
- [playbooks/01-从安装到首批数据.md](./playbooks/01-从安装到首批数据.md)
- [playbooks/02-从分析到增长闭环.md](./playbooks/02-从分析到增长闭环.md)
- [playbooks/03-从治理到协作边界.md](./playbooks/03-从治理到协作边界.md)
- [playbooks/04-SimpleTrack能力优先级.md](./playbooks/04-SimpleTrack能力优先级.md)
- [playbooks/05-从门槛到升级转化.md](./playbooks/05-从门槛到升级转化.md)
