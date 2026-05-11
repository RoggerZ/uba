# Umami Cloud 调研工作区

这个目录集中存放 Umami Cloud 的调研、样例、快照和操作流说明。

## 目录结构

- `Umami功能深度分析.md`: 主调研文档
- `docs/21-源代码实现参考.md`: Umami 官方源码实现审阅与 SimpleTrack 映射
- `docs/`: 官方文档双视角中文解读
  - 入口：[docs/README.md](./docs/README.md)
  - 落地评审：[SimpleTrack 落地评审清单](./docs/落地评审清单.md)
  - 实施路线：[SimpleTrack 实施路线图](./docs/SimpleTrack实施路线图.md)
  - 埋点口径：[SimpleTrack 数据模型与事件字典](./docs/数据模型与事件字典.md)
  - 存储模型单点分析：[事件与属性存储模型源码分析](./docs/事件与属性存储模型源码分析.md)
  - 仿真站执行归档：[真实业务数据方案](./docs/真实业务数据方案.md)、[高品质仿真站设计规范](./docs/高品质仿真站设计规范.md)、[功能打通矩阵](./docs/功能打通矩阵.md)、[执行与复验手册](./docs/执行与复验手册.md)
- `快照索引.md`: 全量快照编号索引
- `快照进度.md`: 阶段进度与缺口
- `tracking-demo/`: Umami Cloud 上报验证样例
- `umami-reference/`: 历史参考截图
- `snapshots/`: 按阶段归档的正式快照
- `raw-playwright/`: 原始 Playwright 抓取与控制台文件
- `../../../references/umami/`: Umami 官方 GitHub 源码只读快照，源码边界见 `SIMPLETRACK_REFERENCE.md`

## 资料分层

- `docs/`
  - 面向官方文档的“能力模块型 + 链路型”中文解读
  - 重点回答 Umami 是怎么拆能力、这些能力如何落地
- `Umami功能深度分析.md`
  - 面向 Umami Cloud 实操的主调研文档
  - 重点回答我们已经验证了什么、页面和交互长什么样
- `snapshots/` + `快照索引.md` + `快照进度.md`
  - 面向截图、操作流和阶段证据的归档层
  - Phase 07 / Phase 08 已完成 `growth-baseline-x3` 三倍样本的真实 Cloud 复验和截图；高级报告对象配置、Funnel、Goals、Segments、Cohorts、Attribution 口径已补齐，Segments 已补应用后结果态，当前剩余边界主要是 Replays 套餐限制、Revenue 多轮重跑累积口径和 Retention 长窗口结论

建议先读 `docs/README.md` 建立能力框架，再回到主调研文档和快照证据层做交叉验证。

如果你只是想先分清几个容易混淆的概念，可以直接去看 `docs/README.md` 里的“常见混淆概念速查”。
如果你想知道哪些能力已经有本地 Cloud 实证、哪些还主要是官方文档解读，可以直接看 `docs/README.md` 里的“实证覆盖速览”。
如果你想从某个模块快速跳回具体截图阶段和主分析章节，可以直接看 `docs/README.md` 里的“证据入口速查”。
如果你想直接重跑某个模块的验证，可以直接看 `docs/README.md` 里的“复验入口速查”。
如果你要把调研结果用于 SimpleTrack 产品评审，可以直接看 `docs/落地评审清单.md`。
如果你要进入开发排期和阶段拆分，可以直接看 `docs/SimpleTrack实施路线图.md`。
如果你要开始设计 SimpleTrack 的事件和字段，可以直接看 `docs/数据模型与事件字典.md`。
如果你要直接看 Umami 的事件、事件属性和 identify/session 属性是怎么分层落库的，可以直接看 `docs/事件与属性存储模型源码分析.md`。
如果你要参考 Umami 的真实源码实现，可以直接看 `docs/21-源代码实现参考.md`，再回看 `../../../references/umami/`。
如果你要把 Umami 和 `analytics-core`、`xwl_bi`、Litlyx 放在一起比较，可以直接看 `../事件与属性存储方案对比.md`。

## 快照规范

- 编号格式:
  - `P01-S01`: Phase 01 的通用步骤快照
  - `P04-D03`: Phase 04 的 Dashboard 快照
  - `P05-C07`: Phase 05 的组件矩阵快照
- 所有正式快照都需要进入 `快照索引.md`
- 每个阶段目录都要维护 `README.md` 和 `flow.md`
- `raw-playwright/` 只保存原始产物，不作为评审成品
