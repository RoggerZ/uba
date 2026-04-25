# 17-UTM

## 这个能力解决什么问题

UTM 解决的是“流量从哪里来、哪条投放链路更有效”的问题。

它把营销来源拆成可分析字段，例如 `source`、`medium`、`campaign`、`content`、`term`，让你能区分“哪个渠道带来流量”与“哪个创意真正有效”。

## 官方原话

> "Umami automatically captures the UTM parameters"

> "Track your campaigns through UTM parameters."

官方列出的标准口径是：
> "5 standard UTM parameters"

## 中文解读

UTM 在 Umami 里不是孤立存在的功能，而是和事件、报表、过滤器一起工作。

实际使用时通常是这条链路：

1. 给投放链接加 UTM 参数。
2. Umami 自动采集这些参数。
3. 在 Dashboards 或 UTM 报告里看结果。
4. 需要时再用 Filters 或 Segments 缩小范围。

这意味着 UTM 不只是“生成链接”，而是一个完整的归因入口。

UTM 的难点从来不是参数本身，而是团队命名纪律。同一个活动如果一会儿写 `spring`，一会儿写 `Spring2026`，一会儿写 `spring_launch`，最后报表会被拆成多份，看起来像三场活动。

## 通俗例子

如果你给同一个落地页发了两条链接：

- `utm_source=newsletter&utm_medium=email&utm_campaign=spring`
- `utm_source=twitter&utm_medium=social&utm_campaign=spring`

那么 UTM 的作用就是让你知道，这两条链接带来的访问和转化是分开的。

## 它和相邻能力的区别

- UTM 解决“流量标签怎么写、怎么采集”。
- Attribution 解决“转化功劳怎么分配”。
- Links 解决“点击链路怎么记录和重定向”。
- Revenue 解决“收入怎么和事件关联起来”。

## 落地动作

- 统一 UTM 命名规则，避免 `newsletter`、`email-news`、`mail` 混用。
- 固定 `source / medium / campaign` 的最小必填集。
- 把 UTM 命名和投放台账绑定，避免分析时不知道每个参数代表什么。
- 在报表里先做一个最小可用的 UTM 视图，再逐步增加创意和内容维度。
- 对大小写、空格、下划线、日期后缀制定规则，并在链接生成器里自动校验。
- `utm_content` 和 `utm_term` 先作为可选字段，等 source/medium/campaign 稳定后再强推。

## 对 SimpleTrack 的启发

SimpleTrack 如果要做增长分析，UTM 应该是“投放入口标准化”的第一步。

值得借鉴的不是参数名字本身，而是这条思路：

- 把外部投放链路结构化。
- 让报表能直接读出营销含义。
- 让链接生成与后续分析在同一套语义里工作。
- UTM 生成器和报表筛选器应共用同一份 campaign 字典，避免运营填参和分析看数脱节。

## 关联现有证据

### 本地已验证

- `simpletrack/docs/umami/snapshots/phase-05-dashboard-components/P05-C14-filter-dialog-fields.png`：过滤器 UI 已具备按字段切换的能力，UTM 数据可以落入这层筛选。
- `simpletrack/docs/umami/snapshots/phase-05-dashboard-components/P05-C15-filter-match-dropdown.png`：`All / Any` 可以配合 UTM 条件做复合分析。
- `simpletrack/docs/umami/tracking-demo/README.md`：本地 demo 已能稳定造出事件和会话数据，适合后续补 UTM 测试流。
- `simpletrack/docs/umami/snapshots/phase-08-growth-and-monetization-insights/README.md`：`P08-S07` 已完成 UTM 截图，Campaign / Source / Medium 有数据。

### 官方文档补充

- 官方文档已经把 UTM 放进 Growth 和 Campaign 相关路径里，说明它不是附属字段，而是正式分析面。
- 本地当前已验证 UTM 结果态，但参数命名纪律和 campaign 字典仍是 SimpleTrack 落地时最需要治理的部分。

## 官方链接

- [UTM](https://docs.umami.is/docs/utm)
- [Measure campaigns](https://docs.umami.is/docs/guides/measure-campaigns)
- [Filters](https://docs.umami.is/docs/filters)
- [Create a public dashboard](https://docs.umami.is/docs/guides/create-a-public-dashboard)

## 继续阅读

- [18-Revenue](./18-Revenue.md)
- [19-Attribution](./19-Attribution.md)
- [playbooks/05-从渠道到营收到归因](./playbooks/05-从渠道到营收到归因.md)
