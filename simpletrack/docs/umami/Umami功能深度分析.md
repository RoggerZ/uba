# Umami 功能深度分析

> 目的：把 Umami Cloud 真正用起来，并把页面、交互和能力拆成可供 SimpleTrack 评审直接复用的证据链。

## 0. 官方文档解读入口

本文件继续承担“Cloud 实操证据链”的角色。

如果你想看另一层材料，请配合 [docs/README.md](./docs/README.md) 一起阅读：

- `docs/`：官方文档双视角中文解读
  - `能力模块型`：解释 Umami 怎么拆系统能力、数据对象和分析体系
  - `链路型`：把多个能力串成最佳实践和落地路径
- 本文：Cloud 实操与产品评审证据
  - 重点保留已经跑通过的页面、操作、截图编号和产品判断

建议的搭配方式：

1. 先在 `docs/` 理解能力边界和最佳实践
2. 再回到本文查看已经验证过的 Cloud 页面和截图证据
3. 如果需要页面级证据，再继续看 `快照索引.md`、`快照进度.md` 和各阶段 `flow.md`

## 1. 调研边界

- 运行方式：Umami Cloud
- 目标：
  - 跑通登录、建站、安装代码、首批数据
  - 跑通事件、属性和 API send
  - 细粒度记录 Dashboard 和 Reports 关键交互
- 资产边界：
  - Umami 相关研究资产统一放在 `docs/umami/`
  - `prototype/simpletrack-umami-inspired/` 继续作为 SimpleTrack 自己的原型目录

## 2. 最短使用链路

对应官方解读：

- [docs/01-安装与接入.md](./docs/01-安装与接入.md)
- [docs/playbooks/01-从0到首批数据.md](./docs/playbooks/01-从0到首批数据.md)

Umami Cloud 的最短使用链路可以拆成 4 步：

1. 登录 Cloud 并进入站点入口
2. 创建 website，拿到 `website id`
3. 在目标页面注入 tracker script 或用 API send 直接发送事件
4. 回到 Overview、Events、Reports 验证数据已经入库

对应快照：

- `P01-*`: 登录与建站
- `P02-*`: 安装代码与首批数据
- `P03-*`: 事件与属性验证

本轮实操结果：

- 登录后直接进入 `Websites` 列表页
- `Add website` 是一个轻量弹窗，只要求 `Name` 和 `Domain`
- 创建 `simpletrack-local / localhost` 后，设置页直接暴露 `Website ID` 与可复制的 tracking code
- 从建站到拿到可用 tracking snippet 的链路很短，适合作为 SimpleTrack onboarding 参考

## 3. 数据采集方式

对应官方解读：

- [docs/02-采集与事件.md](./docs/02-采集与事件.md)
- [docs/03-指标对象与Distinct-ID.md](./docs/03-指标对象与Distinct-ID.md)
- [docs/playbooks/02-从采集到事件分析.md](./docs/playbooks/02-从采集到事件分析.md)

### 3.1 Tracker script

按官方文档，最基础的接入方式是在页面中加载 Umami tracker script，并通过 `data-website-id` 关联具体网站。

本轮样例把 tracker script 动态注入到 demo 页面，避免把敏感配置写死到仓库。

实际验证：

- Cloud 设置页提供的 snippet 形态为：
  - `<script defer src="https://cloud.umami.is/script.js" data-website-id="..."></script>`
- demo 页面成功加载 tracker，浏览器网络层能看到 `POST https://api-gateway.umami.dev/api/send => 200`

### 3.2 HTML 属性埋点

Umami 支持直接在元素上使用 `data-umami-event` 来声明事件，这是最适合做按钮点击类验证的路径。

本轮 demo 用它验证：

- 事件是否入库
- 事件名称是否进入 Events 列表
- 事件属性是否能在 Properties 或明细视图中看到

### 3.3 JavaScript 埋点

对于动态事件，使用 `window.umami.track(name, data)` 更直接。它适合：

- 需要附加属性的事件
- 页面内部状态变化
- 非 DOM 直接触发的业务行为

### 3.4 Identify

`window.umami.identify(id, data)` 适合把匿名访问和业务身份做更稳定的关联。

本轮只验证：

- identify 是否成功执行
- identify 后的附加属性是否进入数据视图

不把用户体系设计或会话建模扩展成额外方案。

### 3.5 API send

官方还提供 `POST /api/send` 的无 UI 上报路径。这适合：

- 服务端模拟事件
- 批量补数据
- 不方便嵌入 tracker script 的上下文

本轮用一个最小 Node 脚本做验证，不引入额外依赖。

实际验证：

- `send-event.mjs` 对 `https://cloud.umami.is/api/send` 的请求返回 `200`
- 响应体为 `{"beep":"boop"}`
- `bulk-send.mjs` 进一步验证了 Cloud `/api/send` 当前接受的主要类型是 `event` 和 `identify`
- 在把数据量放大后，`Overview / Events / Realtime` 都已经开始稳定出数

## 4. 验证样例设计

`tracking-demo/` 会覆盖这四类行为：

- 自动 pageview
- `data-umami-event`
- `window.umami.track(...)`
- `window.umami.identify(...)`

另配一个 `send-event.mjs`，直接向 Cloud API 发事件。

验证目标不是做完整 SDK，而是确认：

- Cloud 接入流程是可跑通的
- Events 与 Properties 视图能承接这些数据
- Dashboard / Reports 是否能消费这些数据

本轮结果分层如下：

- 已确认：
  - tracker script 成功加载
  - 浏览器端上报返回 200
  - Node API send 返回 200
  - 批量 `identify + event` 上报后，`Realtime / Events / Overview` 已进入有数据态
  - `Funnels` 通过手动配置后已经能显示结果
  - `Journeys` 已经出现有解释力的路径图
  - `Properties` 已经可以按 `pricePlan` 等属性展示分布
- 仍可继续扩展：
  - `Properties` 还可以继续补 `variant / surface` 等更多样本

因此文档、索引和操作流里保留了两类证据：

- 早期“请求成功但页面仍是空态”的状态
- 数据量放大后进入稳定展示的状态

这能更真实地反映 Umami Cloud 从接入成功到聚合展示之间存在时间差。

## 5. Dashboard 和 Boards 的区别、联系、以及对 SimpleTrack 的意义

对应官方解读：

- [docs/20-Boards-Links-Pixels-Teams-API.md](./docs/20-Boards-Links-Pixels-Teams-API.md)

这部分是这轮评审最容易混淆的点。结论先说：

- `Dashboard`：系统级默认看板
- `Boards`：用户自定义看板系统

### 5.1 Dashboard 是什么

从顶层导航进入 `Dashboard`，可以看到：

- 一个默认的看板页面
- 一个 `Design` 入口
- 更像“系统预置的主看板”而不是“看板列表”

对应快照：

- `P04-D00`
- `P04-D04`

它的特点是：

- 入口在顶层导航第一位
- 语义更接近“默认工作台”
- 它也能编辑，但编辑对象是这张默认 Dashboard 本身

### 5.2 Boards 是什么

从顶层导航进入 `Boards`，可以看到：

- 一个 board 列表
- `Add board`
- `Design / Edit`

对应快照：

- `P04-D01`
- `P04-D02`
- `P04-D03`

它的特点是：

- 这是一个“用户自建看板系统”
- 先建 board，再选 type，再选 component
- 允许一套账户下维护多张不同用途的 board

### 5.3 二者的联系

- `Dashboard` 是默认看板
- `Boards` 是可扩展看板集合
- 两者底层都使用“组件化布局 + 过滤条件 + 时间范围”
- 但语义不同：
  - `Dashboard` 偏“默认主工作台”
  - `Boards` 偏“专题化自定义看板”

### 5.4 对 SimpleTrack 的指导意义

对 SimpleTrack 来说，最值得借鉴的不是 UI 外观，而是这层产品分工：

- 如果 SimpleTrack 只有一张首页总看板，就可以类比 `Dashboard`
- 如果 SimpleTrack 允许按产品、团队、目标、专题做多张看板，就需要类比 `Boards`

更具体地说：

- MVP 可以先做一张默认 `Dashboard`
- 当用户开始需要“按主题拆看板”时，再引入 `Boards`
- 一开始不要把两者混成一个入口，否则用户会分不清“修改默认首页”和“新建专题看板”到底是什么关系

## 6. Board Type 是什么、怎么用、对 SimpleTrack 有什么意义

对应官方解读：

- [docs/20-Boards-Links-Pixels-Teams-API.md](./docs/20-Boards-Links-Pixels-Teams-API.md)

`Add board` 里当前能看到 4 个 `Board type`：

- `Mixed`
- `Website`
- `Pixel`
- `Link`

对应快照：

- `P04-D02`
- `P04-D03`

### 6.1 Mixed

作用：

- 混合型看板
- 不强绑定某一种数据对象
- 更适合做“综合观察板”

使用方式：

- 创建 board 时选 `Mixed`
- 后续在设计器里自由组合组件

对 SimpleTrack 的意义：

- 这相当于“通用看板”
- 适合给团队负责人或产品经理做综合页

### 6.2 Website

作用：

- 把 board 绑定到某个 website
- 所有组件默认围绕这个 website 的数据展开

使用方式：

- 选 `Website`
- 再选具体 website
- 进入设计器后再加组件

对应快照：

- `P05-C01`
- `P05-C02`
- `P05-C03`

对 SimpleTrack 的意义：

- 这是一种“上下文先绑定，再配组件”的方式
- 对 SimpleTrack 很有帮助，因为它能减少组件配置时的上下文选择成本

### 6.3 Pixel

作用：

- 面向 Pixel 维度的数据做看板
- 更适合跨站点或跨页面的 pixel 采集场景

对 SimpleTrack 的意义：

- 如果后续 SimpleTrack 有自己的采集脚本 / pixel 概念，这类 board type 值得保留
- MVP 阶段可以先不做

### 6.4 Link

作用：

- 面向链接或短链等 link 对象的数据看板

对 SimpleTrack 的意义：

- 只有在产品中存在“链接对象”且需要统计时才有价值
- 对当前 SimpleTrack MVP 可以直接后置

### 6.5 总结

`Board type` 的核心价值不是枚举这 4 个名词，而是建立一条规则：

- 先确定“这张看板服务什么对象”
- 再进入组件配置

这比直接让用户面对一堆组件更清晰。

## 7. 组件选择弹窗里的 7 个组件分别是什么

`Select component` 里当前至少看到 7 个组件：

- `Events chart`
- `Metrics bar`
- `Metrics table`
- `Text`
- `Website chart`
- `Weekly traffic`
- `World map`

对应快照：

- `P05-C05`
- `P05-C06` 到 `P05-C12`

### 7.1 Events chart

作用：

- 看自定义事件随时间的变化

适合：

- 注册点击
- 付费点击
- CTA 触发

### 7.2 Metrics bar

作用：

- 用一排摘要指标展示关键 KPI

适合：

- Views
- Visitors
- Bounce rate
- Time on site

### 7.3 Metrics table

作用：

- 用表格按维度拆指标

适合：

- 按页面
- 按来源
- 按国家
- 按浏览器

### 7.4 Text

作用：

- 放自由文本说明

适合：

- 写解释
- 标注本周关注点
- 写看板用途

### 7.5 Website chart

作用：

- 展示 website 的访问量和访客变化

这一项已经实际落板：

- `P05-C13`

### 7.6 Weekly traffic

作用：

- 以热力图方式看一周内每天每小时的流量分布

适合：

- 看高峰时段
- 看工作日 / 周末差异

### 7.7 World map

作用：

- 用地图展示地域分布

适合：

- 看国际流量来源
- 看区域差异

### 7.8 对 SimpleTrack 的指导意义

这 7 个组件说明 Umami 的组件思路不是“很多复杂图表”，而是围绕 4 类基本信息表达：

- 趋势图
- 摘要指标
- 维度表格
- 辅助说明

SimpleTrack 完全可以先只做这 4 类表达，再慢慢扩张。

## 8. Filter 里的 Fields 是怎么用的

对应官方解读：

- [docs/14-Filters.md](./docs/14-Filters.md)
- [docs/08-Breakdown.md](./docs/08-Breakdown.md)

对应快照：

- `P05-C14`
- `P03-S07`

`Fields` 可以理解成：

- 你要按哪个字段筛数据

当前 UI 里按组展示：

- `URL`
  - `Path`
  - `Query`
  - `Page title`
- `Sources`
  - `Referrer`
- `Location`
  - `Country`
  - `Region`
  - `City`
- `Environment`
  - `Browser`
  - `OS`
  - `Device`
- `UTM`
  - `Source`
  - `Medium`
  - `Campaign`
  - `Content`
  - `Term`
- `Other`
  - `Hostname`
  - `Distinct ID`
  - `Tag`
  - 在 Events 页还能看到 `Event`

怎么理解它的使用：

1. 先选字段
2. 再设匹配逻辑
3. 再限定值
4. 把结果应用到当前页面或组件

`Match` 下拉：

- `All`：所有条件都满足
- `Any`：满足其中一个即可

对应快照：

- `P05-C15`

对 SimpleTrack 的指导意义：

- Fields 不要一开始就铺太多
- 但字段分组方式很值得借鉴，因为它降低了筛选器的认知负担

## 9. Filter 里的 Segments 和 Cohorts 是怎么用的

对应官方解读：

- [docs/15-Segments.md](./docs/15-Segments.md)
- [docs/16-Cohorts.md](./docs/16-Cohorts.md)
- [docs/playbooks/03-从过滤到细分用户.md](./docs/playbooks/03-从过滤到细分用户.md)

### 9.1 Segments

对应快照：

- `P05-C16`

Segments 可以理解成：

- 一组已经命名好的用户或流量切片

比如：

- 付费用户
- 新访客
- 来自某渠道的用户

它的使用方式通常是：

- 先预定义 segment
- 在分析页或 board 里直接套用

意义：

- 适合复用常见筛选条件
- 减少每次都重新配 Fields 的成本

### 9.2 Cohorts

在 Umami 里，Cohorts 更像：

- 按某个起始时间或共同特征形成的一组用户

常见用途：

- 看某天/某周进入的那批用户，后续是否还回来
- 看不同批次用户的留存差异

从产品层面上讲：

- Segments 更像“当前切片”
- Cohorts 更像“带时间维度的一批人”

对 SimpleTrack 的指导意义：

- Segments 可以比 Cohorts 更早进入 MVP
- Cohorts 和 Retention 往往应当配套出现

## 10. Website 编辑页各部分都是什么

对应官方解读：

- [docs/01-安装与接入.md](./docs/01-安装与接入.md)

对应快照：

- `P02-S02`
- `P02-S05`
- `P02-S06`
- `P02-S07`
- `P02-S08`

Website settings 可以拆成 4 块：

### 10.1 Website basics

- `Website ID`
- `Name`
- `Domain`

作用：

- 识别网站
- 修改站点基础配置

### 10.2 Tracking code

作用：

- 告诉你如何把 Umami tracker 接到页面里

这是接入链路里最关键的一块。

### 10.3 Share

作用：

- 把这份统计结果分享给其他人

如果当前为空，说明还没添加共享对象。

### 10.4 Transfer / Reset / Delete

作用：

- `Transfer`: 转移所有权
- `Reset`: 清空数据但保留配置
- `Delete`: 删除站点及其数据

这块本质上是 danger zone，SimpleTrack 后续也应该独立分组。

## 11. Events、Funnels、Journeys、Retention、Realtime 分别在看什么

对应官方解读：

- [docs/05-Realtime.md](./docs/05-Realtime.md)
- [docs/09-Goals.md](./docs/09-Goals.md)
- [docs/10-Funnels.md](./docs/10-Funnels.md)
- [docs/11-Journeys.md](./docs/11-Journeys.md)
- [docs/12-Retention.md](./docs/12-Retention.md)
- [docs/13-Replays.md](./docs/13-Replays.md)
- [docs/playbooks/04-从目标到漏斗到旅程.md](./docs/playbooks/04-从目标到漏斗到旅程.md)
- [docs/playbooks/06-从实时异常到性能与回放排查.md](./docs/playbooks/06-从实时异常到性能与回放排查.md)

### 11.1 Events

对应快照：

- `P03-S04`
- `P03-S05`
- `P03-S06`
- `P03-S07`
- `P03-S08`

页面功能拆解：

- 顶部指标：
  - `Visitors`
  - `Visits`
  - `Events`
  - `Unique events`
- Tabs:
  - `Chart`
    - 看事件量趋势
  - `Activity`
    - 看日志式活动列表
    - 当前列表有 `All / Views / Events`
  - `Properties`
    - 看事件属性维度
- 上方工具：
  - `Filter`
  - 日期范围

更贴近实际的使用方式是：

1. 先在 `Realtime` 验证数据有没有真的进来
2. 再到 `Events` 看聚合后的事件总量和事件名分布
3. 再切 `Activity` 看最近发生了哪些具体事件
4. 再切 `Properties` 看事件属性分析

本轮在数据量放大后，`Events` 已经开始稳定展示：

- `demo_track_call`
- `demo_signup_click`

对应快照：

- `P03-S09`

进一步下钻后，`Properties` 的真实使用路径也已经跑通：

1. 先选 `Event`
2. 再选 `Property`
3. 才会看到属性分布

当前已经确认可见的属性包括：

- `pricePlan`
- `variant`
- `surface`

并且已经拍到 `pricePlan=starter` 的分布结果。

对应快照：

- `P03-S10`
- `P03-S11`
- `P03-S12`

你问的 `Properties` 是什么：

- 它是“事件属性分析视图”
- 用来看某个事件附带的字段分布，比如 plan、role、surface

你问的 `Activity log` 是什么：

- 当前页面里实际叫 `Activity`
- 它更像按时间排布的事件/访问日志视图

### 11.2 Funnels

对应快照：

- `P06-S02`

Funnels 是什么：

- 用一组步骤看转化流失

页面上当前关键元素：

- `Filter`
- 日期范围
- `Funnel` 按钮

说明：

- 这页通常先定义 funnel，再看每一步转化率
- 它不会像 `Realtime` 那样“有数据就自动出结果”
- 必须先配置 funnel 对象

这轮实操里我配置了一个 `Signup to Track Funnel`，用来说明漏斗页的真实使用方式：

- 第一步：`demo_signup_click`
- 第二步：期望追踪 `demo_track_call`

配置过程对应快照：

- `P06-S07`
- `P06-S08`

结果页对应快照：

- `P06-S09`

这轮结果也顺带说明了一个很重要的产品事实：

- 漏斗结果高度依赖 step type 和 step value
- 如果 step type 配错，后续步骤会直接掉到 0

这不是噪音，反而是很好的产品教育案例。

### 11.3 Journeys

对应快照：

- `P06-S03`

Journeys 是什么：

- 看用户从起点到终点的行为路径

页面上当前关键元素：

- `Steps` 数量选择
- `Start Step`
- `End Step`
- 下方类型过滤：`All / Views / Events`

它比 Funnel 更像“路径探索”，而不是固定转化。

这轮在数据量拉大并叠加 demo 会话后，Journeys 已经出现一条可解释路径：

- `/uba/simpletrack/docs/umami/tracking-demo/index.html`
- `demo_track_call`
- `demo_signup_click`

这张图很好地说明了：

- Journeys 更适合回答“用户下一步去了哪里”
- Funnel 更适合回答“用户有没有按既定步骤转化”

对应快照：

- `P06-S10`

### 11.4 Retention

对应快照：

- `P06-S04`

Retention 是什么：

- 看某一批用户在后续天数里还有多少回来

页面上当前关键元素：

- 月份下拉
- 年份下拉
- cohort 表格

这页是本轮唯一明确拍到“有数据态”的高级页，因为页面里出现了 cohort 行数据。

### 11.5 Realtime

对应快照：

- `P06-S05`

Realtime 是什么：

- 看当前在线和即时活动

页面上当前关键元素：

- 即时指标：
  - `Views`
  - `Visitors`
  - `Events`
  - `Countries`
- `Activity`
  - `All / Views / Visitors / Events`
- 即时列表：
  - `Pages`
  - `Referrers`
  - `Countries`

更贴近实际的理解方式是：

- `Realtime` 是接入验收页
- 它最适合回答“数据到底进没进来”
- 它不一定等价于“聚合报表已经完全准备好”

这轮里，正是 `Realtime` 先稳定出数，帮助我们确认：

- 实时指标已经增长
- 活动流已经出现 `demo_track_call`、`demo_signup_click`
- 页面路径、浏览器、国家都已经被 Umami 正确识别

对应快照：

- `P06-S06`

## 12. 关于“上传数据后有数据截图”的本轮结论

这是这轮里最需要诚实记录的地方。

### 已验证成功

- 浏览器 tracker 请求多次返回 `200`
- Node `api/send` 请求多次返回 `200`
- demo 已能稳定触发：
  - pageview
  - DOM event
  - `track()`
  - `identify()`

### 当前已经进入有数据态的页面

经过放大数据量之后，这些页面已经进入有数据态：

- `Overview`
- `Events`
- `Properties`
- `Realtime`
- `Retention`
- `Funnels` 结果页
- `Journeys`

### 已拍到有数据的页面

- `Overview`
- `Events`
- `Properties`
- `Realtime`
- `Retention`
- `Funnels` 结果页
- `Journeys`

这说明当前 Cloud UI 至少不是“完全不处理数据”，但它对不同分析页的展示节奏并不一致。

### 这对评审意味着什么

- 我们已经拿到了“接入成功证据”
- 也拿到了“页面结构证据”
- 还拿到了 `Overview / Events / Properties / Realtime / Retention / Funnel result / Journeys` 的有数据态
- 目前主要剩下的是继续丰富样本，而不是再去追“页面为什么不出数”

因此文档里我会明确标成：

- 已完成：结构评审、功能扫盲、接入证据，以及 `Overview / Events / Properties / Realtime / Retention / Funnels / Journeys` 的关键有数据态
- 未完全完成：如果要做更漂亮的 Journeys 和 Properties 演示，还可以继续补更丰富的样本

## 13. 对 SimpleTrack 的借鉴与取舍

对应官方解读：

- [docs/playbooks/07-SimpleTrack能力优先级.md](./docs/playbooks/07-SimpleTrack能力优先级.md)
- [docs/README.md](./docs/README.md)

本轮基于已完成快照，形成这组结论：

- 值得直接借鉴：
  - `Add website` 只收集最小字段，建站链路很短
  - 设置页同时放 `Website ID` 和 tracking code，减少来回跳转
  - `Boards` 把“board type”放在“component type”之前，降低复杂度
  - 组件弹窗提供列表 + 预览双栏，有利于产品评审
- 适合保留但简化：
  - `Filter` 可以保留 `Fields / Segments / Cohorts` 三层，但 MVP 阶段不必一开始就把字段分组做得这么细
  - 日期范围下拉可以先保留高频项，再逐步扩展到 `All time / Custom range`
- MVP 阶段建议后置：
  - `Replays`
  - `Cohorts`
  - `Revenue`
  - `Attribution`
- 不建议照搬：
  - website 内部左侧报告导航项过多，SimpleTrack MVP 不应该一次摊开这么多入口
  - 当前 Umami Cloud 的空态反馈较弱，接入成功但 UI 不显示时，用户不容易判断是“还没数据”还是“处理延迟”

## 14. `growth-baseline-x3` 仿真站打通计划

本轮新增的 `tracking-demo/site/` 不复用 `prototype/simpletrack-umami-inspired/` 的旧页面壳，而是作为 SimpleTrack 自己的 SaaS 产品站和产品内工作台。它的目标不是展示静态 UI，而是让页面真实打开、跳转、点击，从而向 Umami Cloud 产生可解释的 pageview、session、performance、replay 和业务事件。

### 样本当量

默认预设为 `growth-baseline-x3`：

- 648 个逻辑用户，3 个 session / 用户，共 1944 个 session
- 72 个用户走真实浏览器三段式流量，576 个用户走批量事件灌数
- 每个 session 使用 8 / 10 / 12 个事件模板之一，预计总业务事件约 19440 条
- 6 组 campaign：`producthunt_launch`、`google_brand`、`google_competitor`、`docs_seo`、`linkedin_founder`、`email_nurture`
- 3 组 cohort：`spring_launch`、`self_serve_wave`、`paid_pilot`
- 4 个 plan：`free`、`trial`、`pro_monthly`、`pro_annual`
- 54 个付费 workspace，其中 36 个 `pro_monthly`，18 个 `pro_annual`

### 页面与事件

仿真站页面分为获客、转化、激活、产品内、收入五组，对应 `site/index.html`、`pricing.html`、`compare.html`、`docs.html`、`signup.html`、`install*.html`、`app/*.html`、`checkout*.html`、`billing.html`。所有页面通过普通多页链接跳转，不依赖 SPA 路由。

共享上报接口固定为：

- `initTracker({ websiteId, scriptUrl, enablePerformance, enableReplays })`
- `identifyBusinessUser(profile)`
- `trackBusinessEvent(name, data)`
- `applySessionProfile(sessionProfile)`
- `runPersonaStep(stepName, metadata)`

固定业务事件为 `pricing_viewed`、`compare_opened`、`signup_started`、`signup_completed`、`workspace_created`、`install_started`、`sdk_install_completed`、`first_event_sent`、`dashboard_viewed`、`filter_applied`、`segment_opened`、`cohort_opened`、`checkout_started`、`checkout_completed`、`subscription_upgraded`、`billing_viewed`。

### 功能覆盖矩阵

| 功能 | 数据来源 | 目标截图 | 当前状态 |
| --- | --- | --- | --- |
| Session | 72 个真实浏览器 persona 的多页跳转与 identify | P07-S01 | 已截图；session count 进入 1.7k 量级，有会话列表 |
| RealTime | 浏览器实时 pageview 与 CTA 事件 | P07-S02 | 已截图；`Views / Visitors / Events / Countries` 有数据，活动流有命名事件 |
| Performance | `performance=1` 的真实页面加载 | P07-S03 | 已截图；LCP / FCP / TTFB 等指标有数据 |
| Compare | plan / campaign / cohort 属性 | P07-S04 | 已截图；当前周期指标与路径表有数据 |
| BreakDown | source / medium / plan / role / cohort 属性 | P07-S05 | 已截图；路径 breakdown 表有数据 |
| Goals | signup / first event / checkout 目标事件 | P07-S06 | 已截图；`Checkout Completed Goal` 显示 `49 / 1.73k`，转化率 `3%` |
| Filter | plan / campaign / cohort / role / workspaceSize 字段 | P07-S07, P07-S08 | 已截图；Filter 弹窗可用，Segment 应用后 Compare 指标会按切片重算 |
| Funnels | pricing -> signup -> install -> first_event -> checkout | P08-S01 | 已截图；`Growth Baseline Checkout Funnel` 显示 `1.68k -> 45 -> 45 visitors` |
| Journeys | 获客、激活、产品内、收入路径 | P08-S02 | 已截图；路径流有真实非零数据 |
| Retention | 三组 cohort 和 42 天逻辑跨度属性 | P08-S03 | 已截图；留存矩阵已有 cohort 数据 |
| Replays | 真实浏览器 session，需账号支持 recorder | P08-S04 | 已截图；Business plan 限制 |
| Segments | `segment_opened` 与命名切片属性 | P07-S08, P08-S05, P08-S05A | 已截图；已保存 `Producthunt Launch Segment`，配置为 `UTM Campaign is producthunt_launch`，并已应用到 Compare 形成结果态 |
| Cohorts | `spring_launch` / `self_serve_wave` / `paid_pilot` | P08-S06, P08-S06A | 已截图；已保存 `Paid Checkout Cohort`，配置为 `Triggered event checkout_completed` / `Last 90 days` |
| UTM | 6 组 campaign 的 URL 参数与事件属性 | P08-S07 | 已截图；6 组 campaign 已出现非零 views |
| Revenue | 54 条 `checkout_completed` 收入事件 | P08-S08 | 已截图；当前累积站点结果为 `$11.86k / 355 orders` |
| Attribution | campaign -> signup -> checkout revenue 链路 | P08-S09 | 已截图；`checkout_completed` 归因显示 `49 visitors / 49 visits / 57 views`，并展示 referrer 与 UTM 分布 |

### 当前边界

本地仿真站、浏览器脚本、批量灌数脚本和四份执行文档已经准备好。2026-04-25 已完成真实 Cloud 上报：72 个浏览器 persona 形成 216 个 browser session，批量 run id `growth-x3-full-20260425-1843` 接受 17,280 条事件且 0 失败；正式批量前还额外做过 1 条单事件和 100 条小批量 smoke。仓库不记录真实 website id。

2026-04-25 19:40 已追加本地复核：`tracking-demo/site` 在 `127.0.0.1:49173` 下完成 9 个核心页面的桌面和移动首屏检查，共 18 个页面态，全部 200、无横向溢出、无脚本错误；另用阻断 Umami 外网脚本的方式跑通 2 个 persona、6 个 session、36 次页面点击，验证普通链接跳转和 CTA 交互可用。

当前剩余工作已经从“进入 Umami Cloud UI 做截图”和“解释读侧全空”切换为“记录套餐限制与累积口径”。P07 / P08 的 18 张正式截图已经补齐并重新脱敏，`P07-B00` 和 `P08-B00` 作为登录阻塞历史证据保留。

这轮新增的关键证据有五条：

1. 浏览器网络证据：本地仿真站加载 `https://cloud.umami.is/script.js` 后，真实事件并不是发到 `cloud.umami.is/api/send`，而是发到 `https://api-gateway.umami.dev/api/send`。命名事件 `pricing_viewed` 和 `performance` 事件都能在浏览器请求里看到。
2. 直发脚本修正：`send-event.mjs` 与 `bulk-send.mjs` 的默认端点已经改成 `https://api-gateway.umami.dev/api/send`，与真实浏览器写入链路保持一致。
3. 根因修正：早期空态主要来自脚本使用自定义 User-Agent，Cloud 将这类请求按 bot / 非人类流量处理；切换为普通 Chrome UA 后，写入响应开始带 `sessionId / visitId`，Cloud 读接口也开始返回真实数据。
4. 当前读侧证据：登录态探针显示 Overview stats 为 `pageviews=118 / visitors=2 / visits=2`，Sessions count 为 `1733`，Realtime 返回 112 条活动项，Events 已按事件名聚合，Revenue sessions count 为 `100`；Revenue 页面当前显示 `$11.86k / 355 orders`，这是多轮 smoke、旧 UA、修正 UA 和重跑后的累积站点结果，不等同于单次默认样本只跑一遍。
5. 高级对象证据：通过 Cloud UI 创建 `Checkout Completed Goal`、`Growth Baseline Checkout Funnel`、`Producthunt Launch Segment`、`Paid Checkout Cohort`，并将 Attribution conversion step 切到 `Triggered event / checkout_completed`；Goals、Funnels、Attribution 已显示非零结果，Segments 已补充应用后结果态，Cohorts 已保留配置态截图。

另外，`Replays` 页面现在有明确产品发现：页面可达，但当前账号/套餐提示 `This feature requires a Business plan subscription.`，而网站元数据里 `replayEnabled=false`。这个能力应当按套餐限制记录，而不是按“未验证”处理。

仍需明确的边界是：`Replays` 受当前账号套餐限制，无法验证播放态；`Cohorts` 当前证明的是可保存、可复用的人群对象，列表页本身不直接展示人数；Revenue 与 Attribution 的当前数值来自多轮重跑后的累积站点结果。

## 15. 关联文件

- `快照索引.md`
- `快照进度.md`
- `tracking-demo/README.md`
- `snapshots/phase-*/README.md`
- `snapshots/phase-*/flow.md`
- `docs/落地评审清单.md`
- `docs/SimpleTrack实施路线图.md`
- `docs/数据模型与事件字典.md`
- `docs/真实业务数据方案.md`
- `docs/高品质仿真站设计规范.md`
- `docs/功能打通矩阵.md`
- `docs/执行与复验手册.md`

## 16. 参考文档

> 官方当前公开文档为 Umami v3；下面只保留与本仓库实操、双视角解读和后续复验直接相关的入口。

- [Introduction](https://docs.umami.is/docs)
- [About](https://docs.umami.is/docs/about)
- [Cloud overview](https://docs.umami.is/docs/cloud)
- [Installation](https://docs.umami.is/docs/install)
- [Collect data](https://docs.umami.is/docs/collect-data)
- [Track events](https://docs.umami.is/docs/track-events)
- [Tracker functions](https://docs.umami.is/docs/tracker-functions)
- [Sending stats](https://docs.umami.is/docs/api/sending-stats)
- [Insights](https://docs.umami.is/docs/insights)
- [Reports API](https://docs.umami.is/docs/api/reports)
- [Website statistics API](https://docs.umami.is/docs/api/website-stats)
