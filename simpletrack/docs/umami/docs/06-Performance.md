# 06-Performance

> 说明：`官方原话` 只放短英文摘录；`关联现有证据` 只写本地已验证内容。这里的 Performance 指的是官方当前文档里的真实用户 Web 性能监测，不是普通的时间趋势对比。

## 这个能力解决什么问题

Performance 解决的是“页面是不是慢、慢在哪里、哪些用户受影响更严重”。

它面向的是浏览器真实用户体验，而不是纯业务趋势：

1. 页面首屏是否够快
2. 布局是否稳定
3. 交互是否流畅
4. 不同浏览器、设备、页面之间是否有明显差异

## 官方原话

> "Performance tracks real-user Core Web Vitals"

> "identify slow pages, spot regressions over time"

> "Add the data-performance attribute"

> "p75 means 75% of your visitors experience this value or better"

## 中文解读

Umami 这里讲的 Performance 是真实用户性能监控。

它不是压测结果，也不是单纯服务器指标，而是直接从访客浏览器收集 Core Web Vitals 和相关性能信号。

所以它回答的问题是：

- 哪个页面的 LCP 特别慢
- 哪个页面 CLS 抖动明显
- 哪些浏览器或设备的体验更差
- 改版后性能有没有回退

## 通俗例子

比如首页轮播图改版后，用户反馈“页面感觉更卡了”。

这时你不会先去看注册漏斗，而会先看：

- LCP 是否变慢
- CLS 是否变差
- 问题是否集中在移动端
- 哪个页面 URL 最严重

这就是 Performance 的实际语境。

## 它和相邻能力的区别

- `Performance` 看页面体验和 Web Vitals
- `Compare` 看业务指标的时间对比
- `Realtime` 看现在有没有数据
- `Replays` 看用户在页面上到底怎么操作
- `Funnel` 看步骤转化和流失

如果你怀疑“慢”是原因，先看 Performance。
如果你怀疑“流程设计”是原因，再看 Funnel 或 Replays。

## 落地动作

1. 在 tracking script 上显式开启 `data-performance="true"`
2. 先看 p75，再按页面和环境维度下钻
3. 把 LCP、CLS、INP 当成第一层指标，而不是只看平均值
4. 性能异常时，把 Performance 和 Replays、Sessions 联合使用

## 对 SimpleTrack 的启发

SimpleTrack 如果未来要做更强的体验诊断，Performance 很值得单独成页。

最值得借鉴的是：

- 用真实用户数据，而不是只看实验室跑分
- 默认展示 p75 和 Google 阈值颜色
- 同时提供 Pages breakdown 和 Environment breakdown

这样性能问题才能从“感觉慢”变成“知道哪一页、哪一类设备、哪一个指标慢”。

## 关联现有证据

### 本地已验证

- `../snapshots/phase-07-traffic-and-behavior-insights/README.md`：`P07-S03` 已完成独立 Performance 页面截图，LCP / FCP / TTFB 等指标有数据。
- `../快照进度.md`：Phase 07 记录修正为真实 Chrome UA 后，Performance 已出现可解释数据。
- `../tracking-demo/README.md` 与 `../tracking-demo/app.js` 提供了可重复访问的 demo 页面，适合继续复验 `data-performance="true"` 后的性能数据。

### 官方文档补充

- 官方把 Performance 定义成真实用户 Core Web Vitals 收集
- 启用方式是给 tracker 增加 `data-performance="true"`
- 结果页支持看指标卡、Google 阈值、p50/p75/p95、页面拆分和环境拆分；本仓库当前已验证结果态，但不是压测或实验室跑分证据。

## 官方链接

- [Performance](https://docs.umami.is/docs/performance)
- [Monitor Core Web Vitals](https://docs.umami.is/docs/guides/monitor-web-vitals)
- [Tracker configuration](https://docs.umami.is/docs/tracker-configuration)

## 继续阅读

- [05-Realtime](./05-Realtime.md)
- [13-Replays](./13-Replays.md)
- [playbooks/06-从实时异常到性能与回放排查](./playbooks/06-从实时异常到性能与回放排查.md)
