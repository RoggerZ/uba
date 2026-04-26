# 20-Boards-Links-Pixels-Teams-API

## 这个能力解决什么问题

这一组能力不是单个功能点，而是 Umami 的“展示层 + 协作层 + 接入层”。

- Boards 负责把指标拼成可分享的看板。
- Links 负责记录外链点击。
- Pixels 负责记录外部来源的访问或曝光。
- Teams 负责共享工作区、权限和网站归属。
- API 负责把所有数据能力开放给程序化调用。

## 官方原话

> "Boards are customizable dashboards"

> "Manage your team members and websites."

> "Any operation you can do through the application is available in the API."

相关页面还明确了外部接入形态：
> "record clicks on specific URLs"

> "pixel code that you can embed"

## 中文解读

这几个能力虽然在产品里分属不同入口，但语义是连在一起的：

- Boards 是“怎么把数据摆出来”。
- Links / Pixels 是“怎么把外部流量接进来”。
- Teams 是“谁可以看、谁可以改、网站归谁管”。
- API 是“怎么让别的系统直接调用这些数据”。

Umami 的设计不是把这些东西塞进同一个报表页，而是把它们拆成不同层次，减少用户在一个界面里承受过多概念。

如果把这四层混在一个入口里，用户会分不清自己是在配“展示”、配“采集”、配“权限”，还是在做“系统集成”。

这组能力最适合按“谁在使用”来理解：Boards 面向看数据的人，Links / Pixels 面向投放和外部页面，Teams 面向组织协作，API 面向开发者和自动化系统。

## 通俗例子

如果你要给老板看公开增长面板：

- 用 Board 组合 Overview、Traffic、World Map。
- 用 Share URL 公开它。

如果你要让外部宣传页也能被记录：

- 用 Link 记录点击跳转。
- 用 Pixel 记录页面访问或曝光。

如果你要让团队协作：

- 用 Team 管成员和网站。

如果你要做自动化：

- 用 API 和 API Key 把数据拉到别的系统。

## 它和相邻能力的区别

- Boards 是看板，不是原始数据采集器。
- Links 是点击跟踪，不是通用仪表盘。
- Pixels 是轻量采集入口，不是完整分析页。
- Teams 是权限和协作入口，不是数据模型。
- API 是程序接口，不是 UI 功能。

## 落地动作

- 先决定哪个能力是 MVP 核心，哪个只能后置。
- 如果要做公开分享，先做 Boards 和 Share URL。
- 如果要做外链投放分析，再补 Links。
- 如果要做埋点之外的轻量采集，再补 Pixels。
- 如果要做团队协作，再补 Teams。
- 如果要做自动化集成，再补 API 和 API Key。
- 对外公开之前，先把“公开分享”和“内部团队协作”分成两套权限语义。
- Links 要明确是“点击和重定向跟踪”，不要和 UTM 生成器混成同一个概念。
- Pixels 要明确适合轻量曝光或外部嵌入，不适合替代完整 tracker。
- API Key 要有创建、轮换、撤销和权限说明，不要只给一个永久密钥。

## 对 SimpleTrack 的启发

SimpleTrack 最值得借鉴的是这条分层：

1. 先把看板和数据采集分开。
2. 再把协作权限和数据访问分开。
3. 最后把程序化接口独立出来。

这比“把所有能力都塞在一个 Dashboard 里”更容易维护，也更适合逐步商业化。
公开分享、团队协作、程序化访问三者一定要分开，否则权限模型会很快失控。

另外，Umami 的 Board 类型很值得参考：

- `Website`
- `Mixed`
- `Pixel`
- `Link`

这说明看板本身也可以有语义，而不只是空容器。

SimpleTrack 可以按这个顺序实现：先做 Website Board，再做 Mixed Board；Links 和 Pixels 只有在营销跟踪需求明确后再进入产品主路径；Teams 和 API 需要和权限模型一起设计，不能只补 UI。

## 关联现有证据

### 本地已验证

- `simpletrack/docs/umami/snapshots/phase-04-dashboard-core/P04-D01-boards-list.png`：Boards 列表已存在。
- `simpletrack/docs/umami/snapshots/phase-04-dashboard-core/P04-D03-add-board-type-dropdown.png`：Board type 下拉里已经出现 `Mixed / Website / Pixel / Link`。
- `simpletrack/docs/umami/snapshots/phase-05-dashboard-components/P05-C01-website-board-config-empty-website-picker.png`：Website board 的配置壳。
- `simpletrack/docs/umami/snapshots/phase-05-dashboard-components/P05-C04-design-board-initial.png`：Design board 初始态。
- `simpletrack/docs/umami/snapshots/phase-05-dashboard-components/P05-C13-board-with-website-chart-component.png`：组件已能落到 board。
- `simpletrack/prototype/simpletrack-umami-inspired/team.html`：本地原型里已有 Team / Invite member / Full access / Pending 等协作语义。
- `simpletrack/docs/umami/tracking-demo/send-event.mjs` 与 `simpletrack/docs/umami/tracking-demo/bulk-send.mjs`：本地已经有 API send 和批量造数脚本，可作为 API 层验证基础。

### 官方文档补充

- Boards 已通过列表、类型下拉和配置壳截图形成较强本地证据；Links / Pixels / Teams 本轮主要作为官方文档和产品设计输入，后续如进入 SimpleTrack 主路径再补管理页截图。
- 官方文档已经把 Share URL、Board share、Link、Pixel、Team、API 分成独立文档，说明这些能力在产品里应该保持清晰边界。

## 官方链接

- [Using boards](https://docs.umami.is/docs/using-boards)
- [Create a public dashboard](https://docs.umami.is/docs/guides/create-a-public-dashboard)
- [Enable Share URL](https://docs.umami.is/docs/enable-share-url)
- [Design a board](https://docs.umami.is/docs/design-a-board)
- [Links](https://docs.umami.is/docs/links)
- [Pixels](https://docs.umami.is/docs/pixels)
- [Manage a team](https://docs.umami.is/docs/manage-a-team)
- [Using teams](https://docs.umami.is/docs/using-teams)
- [API Overview](https://docs.umami.is/docs/api)
- [Teams API](https://docs.umami.is/docs/api/teams)
- [Websites API](https://docs.umami.is/docs/api/websites)
- [Events API](https://docs.umami.is/docs/api/events)
- [Sessions API](https://docs.umami.is/docs/api/sessions)
- [Links API](https://docs.umami.is/docs/api/links)
- [Pixels API](https://docs.umami.is/docs/api/pixels)
- [API client](https://docs.umami.is/docs/api/api-client)
- [API Key](https://docs.umami.is/docs/cloud/api-key)

## 继续阅读

- [README](./README.md)
- [playbooks/07-SimpleTrack能力优先级](./playbooks/07-SimpleTrack能力优先级.md)
