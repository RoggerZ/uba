# 15-Segments

## 这个能力解决什么问题

Segments 解决的是“常用过滤条件如何保存并重复使用”的问题。

Filters 适合临时分析，Segments 适合把某一套分析口径固化下来，例如“北美新访客”“付费来源用户”“来自某个活动页的访问者”。

## 官方原话

> "Segments let you save commonly used filters in Umami."

> "quickly reapply them without setting criteria each time."

官方还把操作讲成：
> "Save Segment"

## 中文解读

Segments 本质上是“可保存的过滤器模板”。

它的使用方式通常是：

1. 在网站页面先把 Filters 配好。
2. 保存为 Segment。
3. 之后在其他页面直接切换到 Segments tab 复用。

这样做的好处是，分析口径不必每次重新拼装，也减少了团队成员之间口径不一致的问题。

Segments 的产品价值不是“多一个收藏夹”，而是让团队把常用口径变成共享资产。一个好 Segment 应该有业务化名字、清楚的条件说明和明确的维护人，否则保存越多，越难知道哪个还能用。

## 通俗例子

你每周都要看“美国 + 移动端 + 来自邮件渠道”的流量。

如果没有 Segments，你每次都要重新点 3 次条件。
如果有 Segments，你保存一次，下次直接点这个分组就行。

## 它和相邻能力的区别

- Filters 是当前视图的临时条件。
- Segments 是可保存、可复用的过滤条件集。
- Cohorts 是按行为和时间分出来的用户组，不等同于一般的筛选条件。
- Breakdown 是把数据按维度切开看，不是保存筛选集合。

## 落地动作

- 给常见分析口径提供“保存为分组”的按钮。
- 给 Segment 命名时要求业务化，而不是技术化。
- 允许团队复用 Segment，而不是每个人各建一套。
- 让 Segment 也能出现在报表和 Dashboard 的共享状态里。
- 在 Segment 列表中展示条件摘要和最近更新时间，避免历史分组失效后没人发现。
- 对团队共享 Segment 增加命名规范，例如 `渠道-地区-设备` 或 `生命周期-动作-时间`。

## 对 SimpleTrack 的启发

SimpleTrack 如果要做团队级分析，Segments 是很值得优先补的能力。

原因很直接：

- 它把高频筛选变成稳定资产。
- 它比“复制一堆过滤器条件”更适合协作。
- 它是 Boards 之外的另一层复用结构。
- Segment 要有治理规则，不能只允许无限保存；否则它会从“复用资产”变成“筛选垃圾桶”。

对 MVP 来说，先做“保存过滤器模板”，再做更复杂的 cohort/attribution，路径会更稳。

## 关联现有证据

### 本地已验证

- `simpletrack/docs/umami/snapshots/phase-05-dashboard-components/P05-C16-filter-segments-tab-empty.png`：早期 Filter 弹窗里的 Segments tab 空态，保留为入口历史证据。
- `simpletrack/docs/umami/snapshots/phase-08-growth-and-monetization-insights/P08-S05-segments.png`：Segments 独立页已显示 `Producthunt Launch Segment` 保存对象。
- `simpletrack/docs/umami/snapshots/phase-08-growth-and-monetization-insights/P08-S05A-segment-config.png`：该 Segment 配置为 `UTM Campaign is producthunt_launch`。
- `simpletrack/docs/umami/snapshots/phase-05-dashboard-components/P05-C14-filter-dialog-fields.png`：Segments 的前置条件是先有 Filters。

### 官方文档补充

- 官方文档说明了保存、应用、编辑和删除 Segment 的流程。
- 本地当前已经截到“已保存 Segment 列表”和配置态；还没有单独截“应用 Segment 后的结果页”，后续可在 Compare / Breakdown 或 Dashboard Filter 里补交互证据。

## 官方链接

- [Segments](https://docs.umami.is/docs/segments)
- [Filters](https://docs.umami.is/docs/filters)
- [Cohorts](https://docs.umami.is/docs/cohorts)
- [Sessions API filters](https://docs.umami.is/docs/api/sessions)

## 继续阅读

- [14-Filters](./14-Filters.md)
- [16-Cohorts](./16-Cohorts.md)
- [12-Retention](./12-Retention.md)
