# 16-Cohorts

## 这个能力解决什么问题

Cohorts 解决的是“把满足某个行为条件的用户，按时间窗口分成一组，然后看这组人后续表现”的问题。

它比 Filters 更接近“用户分群”，比 Segments 更强调时间边界和行为定义。常见用途是看留存、重复访问、某个动作后的后续行为趋势。

## 官方原话

> "Cohorts let you group users based on specific actions"

> "choose a Custom Range"

官方入口动作包括：
> "Create Cohort"

## 中文解读

Umami 的 Cohorts 不是简单按标签分组，而是要求你先定义：

- 时间范围。
- 行为条件，例如访问某个 URL 或触发某个事件。
- 是否要做静态 cohort。

官方文档特别强调，如果你要静态 cohort，就应该选 Custom Range，这样这组用户不会随着时间滑动而漂移。

Cohort 的核心是“先圈人，再观察”。它不是每次看报表时临时加一个条件，而是把某段时间里做过某件事的人固定下来，后面再看这批人的留存、转化或复访。

## 通俗例子

比如你想看“参加过 4 月活动页的用户，之后一周有没有再回来”。

这不是普通筛选，而是 cohort：

- 先圈定“4 月活动页访问者”。
- 再看这批人在后续时间段里的行为变化。

## 它和相邻能力的区别

- Filters 是当下分析条件。
- Segments 是可保存的过滤条件模板。
- Cohorts 是基于时间和动作定义的用户群。
- Attribution 是在 conversion 之上看“功劳归谁”。

## 落地动作

- 先明确 cohort 必须绑定时间范围，避免组别随时间漂移。
- 把“页面访问”和“事件触发”作为两类最小 cohort 定义。
- 给 cohort 提供可读命名，例如“Signup Apr 2026 cohort”。
- 在结果页里同时显示样本量、留存和后续动作，避免只有一个总数。
- 创建 cohort 时展示“是否静态”和“时间范围”，让用户知道这群人以后会不会变化。
- 不要把 cohort 命名成纯技术条件，最好包含行为和时间，例如 `2026-04_signup_completed`。

## 对 SimpleTrack 的启发

SimpleTrack 如果未来要做留存、回访、激活用户分析，Cohorts 应该独立于 Segments 实现。

原因是：

- Segments 适合复用筛选口径。
- Cohorts 适合复用人群定义。
- Cohorts 的时间稳定性比普通筛选更重要。
- Cohort 适合放在 Retention、Lifecycle、Attribution 这类长期分析前面，而不是作为普通筛选器的替代品。

如果把两者混在一起，后面做留存和生命周期分析会很难收口。

## 关联现有证据

### 本地已验证

- `simpletrack/docs/umami/tracking-demo/bulk-send.mjs`：批量数据生成脚本里已经有 `cohort` 字段，说明本地 demo 可以稳定造出 cohort 相关数据。
- `simpletrack/docs/umami/snapshots/phase-06-reports-review/P06-S01-date-range-dropdown-on-board.png`：日期范围选择器已单独被截取，说明时间窗口是当前报告层的重要控制项。

### 官方文档补充

- 本地当前已经补到单独的 Cohorts 页面截图：`P08-S06` 显示 `Paid Checkout Cohort` 保存对象，`P08-S06A` 显示配置为 `Triggered event checkout_completed`、日期范围 `Last 90 days`。
- 官方文档明确说明了静态 cohort 需要 Custom Range，这一点要写进后续 SimpleTrack 的产品规则里。

## 官方链接

- [Cohorts](https://docs.umami.is/docs/cohorts)
- [Retention](https://docs.umami.is/docs/retention)
- [Sessions API filters](https://docs.umami.is/docs/api/sessions)
- [Measure campaigns](https://docs.umami.is/docs/guides/measure-campaigns)

## 继续阅读

- [15-Segments](./15-Segments.md)
- [12-Retention](./12-Retention.md)
- [playbooks/03-从过滤到细分用户](./playbooks/03-从过滤到细分用户.md)
