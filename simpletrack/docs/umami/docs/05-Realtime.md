# 05-Realtime

> 说明：`官方原话` 只放短英文摘录；`关联现有证据` 只写本地已验证内容。Realtime 是这批资产里“最适合验收接入是否成功”的页面。

## 这个能力解决什么问题

Realtime 解决的是“数据现在有没有进来”。

它不是最终分析层，而是接入验收层。它最擅长回答：

1. 现在有没有活跃访客
2. 页面和来源有没有被识别
3. 事件是否正在进入系统

## 官方原话

> "Realtime data for your website"

> "GET /api/realtime/:websiteId"

> "urls"

> "countries"

> "events"

API 返回字段还包括：
> "series" / "referrers" / "totals"

## 中文解读

Realtime 的价值不在“报表最完整”，而在“反馈最快”。

它把当前时刻的 URL、国家、事件、来源和数值串在一起，适合用来判断：

- tracker 有没有成功发送
- 事件是不是已经进入后端
- 站点是不是已经有真实流量

Realtime 的意义是“快照式确认”，不是长期趋势判断。它适合发布后、接入后、活动刚上线后的即时检查；如果要解释一整周或一个月的变化，应该回到 Compare、Breakdown 或其他报表。

## 通俗例子

你刚把埋点接到页面上时，最先要看的不是留存，而是：

- 现在有没有人访问
- 刚刚点的按钮有没有触发
- 国家、来源、页面列表是不是开始变化

这时候看 Realtime，比看完整报表更直接。

## 它和相邻能力的区别

- `Realtime` 是“接入验收页”
- `Events` 是“事件聚合分析页”
- `Sessions` 是“访问者历史页”
- `Reports` 是“深度分析页”

Realtime 能快，不代表它替代所有分析页。

## 落地动作

1. 把 Realtime 作为接入后的第一道验收
2. 先确认 totals、events、countries、urls 有没有变化
3. 再去看 Events / Sessions / Reports
4. 把“有数据”与“分析结果充分”拆成两个状态
5. 对刚接入的新站，先用单事件或真实浏览器流量制造可观察变化，再判断 tracker 是否正常

## 对 SimpleTrack 的启发

SimpleTrack 很适合把 Realtime 做成“接入健康页”：

- 数据进来了没
- 最近一条事件是什么
- 页面、来源、国家有没有识别出来

这类页面应该比复杂报表更早出现，因为它直接决定用户会不会相信接入成功。

Realtime 页最好提供“最后收到数据时间”和“最近事件样例”，这样比只显示一个在线人数更能帮助用户排查。

## 关联现有证据

### 本地已验证

- `../snapshots/phase-06-reports-review/flow.md` 明确记录了 `Realtime` 页面在加大 demo 数据后开始展示实时指标、活动流、页面列表和国家分布
- `../snapshots/phase-06-reports-review/P06-S05-realtime-page.png`
- `../snapshots/phase-06-reports-review/P06-S06-realtime-with-data.png`
- `../tracking-demo/send-event.mjs` 和 `../tracking-demo/bulk-send.mjs` 都能把数据直接送进 Cloud，用于验证 Realtime 是否开始变化

### 官方文档补充

- Realtime API 返回 `urls`、`countries`、`events`、`series`、`referrers`、`totals`
- `GET /api/realtime/:websiteId` 说明它本质上就是当前网站的实时数据接口

## 官方链接

- [Realtime API](https://docs.umami.is/docs/api/realtime)
- [Insights](https://docs.umami.is/docs/insights)
- [Introduction](https://docs.umami.is/docs)

## 继续阅读

- [02-采集与事件](./02-采集与事件.md)
- [04-Sessions](./04-Sessions.md)
- [06-Performance](./06-Performance.md)
- [playbooks/06-从实时异常到性能与回放排查](./playbooks/06-从实时异常到性能与回放排查.md)
