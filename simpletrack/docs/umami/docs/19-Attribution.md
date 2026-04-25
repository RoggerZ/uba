# 19-Attribution

## 这个能力解决什么问题

Attribution 解决的是“转化到底该算给哪一次触点”的问题。

它不是在问“流量从哪里来”，而是在问“在一个转化路径里，哪个渠道、哪次访问、哪条 UTM 链路更该被记功”。

## 官方原话

> "helps track the effectiveness of marketing channels"

> "displays referrer, paid ads, and UTM data"

官方参数里最关键的是：
> "Model" / "Type" / "Conversion Step"

## 中文解读

Umami Attribution 的核心是归因模型。

官方目前提供的模型里，最关键的是：

- First-Click：把功劳给最早触点。
- Last-Click：把功劳给最终触点。

你还需要先选一个 conversion step，例如某个页面或某个事件。也就是说，Attribution 不是独立看流量，而是围绕一个转化目标来分配贡献。

这也是归因最容易被误解的地方：它不是“自动告诉你哪个渠道一定创造了收入”，而是在你定义好的转化事件和模型规则下，重新分配路径上的触点功劳。

## 通俗例子

用户先点了广告，后来看了博客，又通过邮件链接完成注册。

- First-Click 会更偏向最早的广告。
- Last-Click 会更偏向最后的邮件。

这就是 Attribution 和 UTM 最大的差别：UTM 只是在标记来源，Attribution 在决定“功劳怎么分”。

## 它和相邻能力的区别

- UTM 负责给链接打来源标签。
- Links 负责记录链接点击和重定向。
- Revenue 负责把收入挂到事件上。
- Attribution 负责把转化结果分配给不同触点。

## 落地动作

- 先定义一个统一的 conversion step 清单。
- 先只做 First-Click / Last-Click 两种最小模型。
- 把 UTM、referrer、paid channel 作为归因输入，而不是各自独立看。
- 让归因报表能回溯到具体事件或页面，不要只给汇总数。
- 在报表标题或说明里写明当前模型，例如 `First-Click` 或 `Last-Click`，避免团队用不同模型讨论同一个结论。
- 归因结果只能作为决策输入，不要直接等同于因果 ROI；真正要证明因果仍需要实验或更严格的对照。

## 对 SimpleTrack 的启发

SimpleTrack 如果要做营销分析，Attribution 应该晚于 UTM 和 Revenue 落地。

原因很简单：

- 没有稳定的 UTM，归因输入不干净。
- 没有收入或转化事件，归因输出没有业务价值。
- 没有转化步骤定义，模型就没有锚点。

所以它适合放在第二阶段做，而不是第一天就做。

第一个版本可以只支持“选择转化事件 + 选择归因模型 + 按渠道展示贡献”，先把解释口径做清楚，再扩多触点模型。

## 关联现有证据

### 本地已验证

- `simpletrack/docs/umami/snapshots/phase-03-events-and-properties/P03-S09-events-chart-with-data.png`：事件数据已经能在本地进入聚合视图。
- `simpletrack/docs/umami/snapshots/phase-03-events-and-properties/P03-S12-properties-with-data.png`：属性分布已经可视化，说明后续归因所需的事件属性基础是存在的。
- `simpletrack/docs/umami/tracking-demo/bulk-send.mjs`：脚本里已经批量造出了稳定的 event / identify 数据，适合用来验证归因输出。
- `simpletrack/docs/umami/snapshots/phase-08-growth-and-monetization-insights/README.md`：`P08-S09` 已完成 Attribution 页面截图；转化条件切到 `Triggered event / checkout_completed` 后，页面显示 `49 visitors / 49 visits / 57 views`，并能按 referrer、UTM source、medium 展示分布。

### 官方文档补充

- 官方文档明确把 referrer、paid ads 和 UTM 放在同一个归因框架里，这一点是 SimpleTrack 后续设计的关键参考。
- 本地当前已经证明归因配置入口、conversion step 校准和非零结果态；但归因结果仍只能作为渠道贡献视角，不能直接写成因果 ROI。

## 官方链接

- [Attribution](https://docs.umami.is/docs/attribution)
- [UTM](https://docs.umami.is/docs/utm)
- [Measure campaigns](https://docs.umami.is/docs/guides/measure-campaigns)
- [Filters](https://docs.umami.is/docs/filters)

## 继续阅读

- [17-UTM](./17-UTM.md)
- [18-Revenue](./18-Revenue.md)
- [playbooks/05-从渠道到营收到归因](./playbooks/05-从渠道到营收到归因.md)
