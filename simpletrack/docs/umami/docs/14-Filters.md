# 14-Filters

## 这个能力解决什么问题

Filters 是 Umami 最基础的分析入口。它解决的是“我要把同一份数据按条件切开看”的问题，例如只看某个页面、某个来源、某个国家、某个设备，或者只看某一组 UTM 数据。

在 Umami 里，Filters 不是只服务某一个页面，而是可以作为通用条件跨屏复用。这样做的价值是：同一条分析口径可以同时作用于 Overview、Reports、Sessions 等视图，减少“每个页面都重新筛一遍”的重复操作。

## 官方原话

> "With universal filtering, you can apply conditions across all screens/reports."

> "The visitor must match every condition."

官方过滤分类包括：
> "Page" / "UTM" / "Event Data" / "Session Data"

## 中文解读

Umami 的过滤器分成几类：URL、来源、地理位置、环境、UTM 和其他自定义维度。

它的核心不是“做一个高级搜索框”，而是把分析条件标准化：

- 先选字段，再选匹配方式。
- 再决定是 `All` 还是 `Any`。
- 最后把这个条件复用到同一站点的不同分析页。

这意味着 Filters 更像分析底层的“切片规则”，而不是某个单页里的临时筛选器。

过滤器设计的关键是字段来源要清楚。页面 URL、Referrer、UTM、事件属性、会话属性看起来都像“字段”，但它们对应的数据层不同；如果 UI 不区分来源，用户很容易把“事件上的 plan”和“会话上的 plan”混成一个东西。

## 通俗例子

如果你想看“美国地区、Chrome 浏览器、来源是 newsletter 的用户”，Filters 就是把这三个条件同时挂上去，然后所有相关图表都按这组条件重算。

## 它和相邻能力的区别

- Filters 是一次性的条件组合，重点是“现在按什么条件看”。
- Segments 是把常用 Filters 存起来，重点是“下次直接复用”。
- Cohorts 是按时间和行为把用户分组，重点是“这群人在一段时间内做过什么”。
- Breakdown 更偏聚合分析，不是条件配置本身。

## 落地动作

- 先定义一套统一的过滤字段字典，避免不同页面出现同义不同名。
- 把 `Match`、日期范围、字段选择做成共享组件。
- 让过滤条件可保存、可分享、可回放，减少口头说明成本。
- 对高频分析维度优先提供快捷入口，例如来源、国家、设备、UTM。
- 对字段标注来源，例如 `Page`、`UTM`、`Event Data`、`Session Data`。
- 对 `All / Any` 用自然语言解释成“必须都满足 / 满足任意一个”，不要只暴露技术词。

## 对 SimpleTrack 的启发

SimpleTrack 如果要做自己的分析底座，Filters 应该先于图表复杂度存在。

最值得借鉴的是两点：

- 过滤器状态要能跨页面保留。
- 过滤器语义要稳定，不要把“临时筛选”和“长期分群”混在一个入口里。
- 字段字典必须先于高级图表稳定下来；否则图表越多，筛选口径越容易分裂。

如果后面要做 Boards、Segments、Cohorts，这层过滤器会是共同底座。

## 关联现有证据

### 本地已验证

- `simpletrack/docs/umami/snapshots/phase-03-events-and-properties/P03-S07-events-filter-dialog.png`：Filter 弹窗。
- `simpletrack/docs/umami/snapshots/phase-05-dashboard-components/P05-C14-filter-dialog-fields.png`：Fields 结构。
- `simpletrack/docs/umami/snapshots/phase-05-dashboard-components/P05-C15-filter-match-dropdown.png`：`All / Any` 匹配模式。
- `simpletrack/docs/umami/snapshots/phase-07-traffic-and-behavior-insights/README.md`：`P07-S07` 已完成 Filter 截图，过滤弹窗可用，底层 Compare 有数据。
- `simpletrack/docs/umami/snapshots/phase-07-traffic-and-behavior-insights/P07-S08-filter-segment-applied.png`：已保存 Segment 应用到 Compare 后，指标按切片重算。

### 官方文档补充

- 官方文档明确列出过滤类别、匹配模式和操作符；本地当前已验证过滤器 UI、字段结构和带数据视图中的应用入口，但某个具体 UTM 或地理筛选组合的结果页仍可在后续补交互态。

## 官方链接

- [Filters](https://docs.umami.is/docs/filters)
- [Segments](https://docs.umami.is/docs/segments)
- [Cohorts](https://docs.umami.is/docs/cohorts)
- [Breakdown](https://docs.umami.is/docs/breakdown)
- [Sessions API filters](https://docs.umami.is/docs/api/sessions)

## 继续阅读

- [15-Segments](./15-Segments.md)
- [16-Cohorts](./16-Cohorts.md)
- [playbooks/03-从过滤到细分用户](./playbooks/03-从过滤到细分用户.md)
