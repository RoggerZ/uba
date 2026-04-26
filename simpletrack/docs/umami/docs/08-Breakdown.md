# 08-Breakdown

## 这个能力解决什么问题
Breakdown 用来把数据按维度拆开看。它解决的是“数据都混在一起，不知道哪些来源、路径或属性贡献了结果”的问题。

## 官方原话
> “aggregate and view your data in a variety of ways.”

官方页面还说明它依赖分段和过滤：
> “utilization of segments and filters”

## 中文解读
Breakdown 是一种“切片”视图。你选一个字段，它就帮你把页面数据、会话数据或其他可分组的数据拆开，看看每一类各占多少。

## 通俗例子
比如你想看访问量是来自哪个路径、哪个国家、哪个浏览器，或者某个自定义属性的分布情况。Breakdown 就像把总表按某个字段分组统计。

## 它和相邻能力的区别
- 和 Compare 的区别：Compare 看时间差，Breakdown 看维度差。
- 和 Goals 的区别：Breakdown 关注分布结构，Goals 关注是否完成转化。
- 和 Funnels 的区别：Breakdown 不要求步骤顺序，也不关心掉漏斗，只关心拆分后的结果。

## 落地动作
- 把最常用的分析字段预设进去，比如 Path、Country、Browser、Event property。
- 为每个字段提供清晰的筛选器和分组项。
- 先让用户看到“结构”，再让用户决定要不要继续下钻。

## 对 SimpleTrack 的启发
- SimpleTrack 可以把 Breakdown 作为“分析底座”，优先支持路径、来源、设备和自定义属性拆分。
- 如果有事件属性系统，Breakdown 应该直接支持事件字段聚合。
- 对初级用户，默认字段不要太多，避免一上来就暴露复杂度。

## 关联现有证据

### 本地已验证

- 本地证据：`simpletrack/docs/umami/snapshots/phase-03-events-and-properties/` 中的 Properties 页面、属性选择器和属性分布图，已经验证了“按属性拆分”的分析需求。
- 本地证据：`simpletrack/docs/umami/snapshots/phase-05-dashboard-components/P05-C14-filter-dialog-fields.png`、`P05-C15-filter-match-dropdown.png`、`P05-C16-filter-segments-tab-empty.png` 说明仓库里已经记录了字段、过滤与 Segments 的交互。

### 官方文档补充

- 官方文档证据：Breakdown 的参数就是 `Fields`，并默认从 Path 开始。

## 官方链接
- [Breakdown](https://docs.umami.is/docs/breakdown)
- [Compare traffic across time periods](https://docs.umami.is/docs/guides/compare-traffic-periods)
- [Insights](https://docs.umami.is/docs/insights)

## 继续阅读

- [07-Compare](./07-Compare.md)
- [14-Filters](./14-Filters.md)
- [15-Segments](./15-Segments.md)
