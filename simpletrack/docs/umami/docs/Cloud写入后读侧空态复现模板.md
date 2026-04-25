# Cloud 自定义 UA 过滤复盘模板

## 用途

这份文档保留早期“写入请求成功，但 Cloud 查询接口始终返回 0/空数组”的复盘路径。最新结论已经不是 Cloud 系统性不可读，而是自定义 `User-Agent` 触发了 Cloud 对 bot / 非人类流量的过滤；改用普通 Chrome UA 后，写入结果可以进入 Cloud 读侧。

注意：

- 不要把真实 `website id`、账号密码、cookie、token、storage state 写入仓库。
- 仓库内只保留复现步骤、命令、观察结论和需要后续确认的问题。
- 真实站点标识请在本地执行时通过命令行参数临时传入。

## 最小结论

- 本地页面真实打开、跳转、点击已经触发浏览器写入请求。
- 命名事件和性能事件都能在浏览器网络层看到。
- 浏览器脚本实际把数据发到 `https://api-gateway.umami.dev/api/send`。
- 早期自定义 UA 路径返回 `200 {"beep":"boop"}`，但读侧长期保持 0 或空数组。
- 改用普通 Chrome UA 后，写入响应开始带 `sessionId / visitId`，Cloud authenticated read API 开始返回真实数据。
- 最新复验中 Overview、Realtime、Events、Sessions、UTM、Revenue 等页面已经有可解释结果；Replays 仍受 `Business plan` 限制。

## 建议附上的截图

- `P07-S01`：Sessions 有真实会话列表。
- `P07-S02`：Realtime 有 views、visitors、events 和活动流。
- `P07-S03`：Performance 有 LCP / FCP / TTFB 等指标。
- `P08-S04`：Replays 套餐限制。
- `P08-S07`：UTM 有 6 组 campaign 数据。
- `P08-S08`：Revenue 有累积收入和订单数据。
- `P08-S09`：Attribution 已在 `Triggered event / checkout_completed` 口径下显示非零结果，可作为“默认 conversion step 不等于业务转化事件”的修正证据。

## 建议复验步骤

1. 启动本地站点

```powershell
cd C:\Users\admin\Documents\src\uba\simpletrack\docs\umami\tracking-demo
python -m http.server 49173
```

2. 用普通 Chrome UA 发送最小单事件

```powershell
node send-event.mjs --website-id YOUR_WEBSITE_ID --hostname localhost --url http://localhost:49173/site/index.html --title "Probe Event" --name probe_event
```

3. 发送一小轮真实浏览器流量

```powershell
npx --yes -p playwright node run-browser-flows.mjs --website-id YOUR_WEBSITE_ID --base-url http://localhost:49173/site --users 3 --settle-ms 1500
```

4. 在已登录的浏览器 profile 上抓 Cloud 读接口

```powershell
npx --yes -p playwright node probe-cloud-read-apis.mjs --website-id YOUR_WEBSITE_ID --user-data-dir "%TEMP%\\simpletrack-umami-playwright-profile"
```

## 关键观察

### 1. 写入侧观察

- 浏览器请求里能看到 `pricing_viewed` 等命名事件。
- 浏览器请求里能看到 `type=performance` 的性能上报。
- 写入目标是 `https://api-gateway.umami.dev/api/send`。
- 普通 Chrome UA 路径会返回可关联的 `sessionId / visitId` 信息。

### 2. 读侧观察

- `GET /analytics/<region>/api/websites/:websiteId/stats?...` 已返回非零 `pageviews / visitors / visits`。
- `GET /analytics/<region>/api/realtime/:websiteId` 已返回活动项、URL、referrer 和 country 数据。
- `GET /analytics/<region>/api/websites/:websiteId/events/...` 已返回事件聚合。
- `GET /analytics/<region>/api/websites/:websiteId/sessions?...` 已返回会话列表和 count。
- `GET /analytics/<region>/api/websites/:websiteId/revenue/sessions?...` 已返回收入相关 session 列表。

### 3. 已确认的产品边界

- `Replays` 当前受套餐限制，站点元数据 `replayEnabled=false`。
- `Funnels` 需要先保存对象并确保 step 事件和时间范围匹配；当前 `Growth Baseline Checkout Funnel` 已显示非零结果。
- `Goals / Segments / Cohorts` 都依赖已保存的配置对象；当前已分别保存 checkout goal、producthunt segment 和 checkout cohort。
- `Attribution` 需要把 conversion step 明确指向业务事件；当前 `checkout_completed` 口径已显示 referrer 与 UTM 分布。

## 可直接发送的说明草稿

```text
We previously reproduced zero/empty authenticated Cloud reads after writes, but the current evidence points to a User-Agent filtering issue rather than a general Cloud read-side failure.

What we verified:
1. Real browser traffic is generated against a localhost site and the browser network shows named events and performance events being sent.
2. The actual write endpoint used by cloud.umami.is/script.js is https://api-gateway.umami.dev/api/send.
3. When our scripts used a custom User-Agent, writes returned a generic 200 response and Cloud read APIs stayed at zero/empty.
4. After switching scripts to a normal Chrome User-Agent, write responses included session/visit identifiers and authenticated read APIs started returning real data for overview, realtime, events, sessions, UTM, and revenue.
5. Replays still appears gated by plan because the website metadata has replayEnabled=false and the UI says Business plan is required.

Could you confirm whether custom User-Agent traffic is intentionally filtered as bot/non-human traffic in Umami Cloud, and whether there is a documented way to mark trusted synthetic QA traffic without losing query visibility?
```

## 当前最值得继续确认的问题

1. Umami Cloud 是否明确按 `User-Agent` 过滤 bot / synthetic traffic。
2. 自定义 UA 的 `200 {"beep":"boop"}` 是否只代表请求被接收，而不代表会进入报表。
3. 是否有官方推荐方式让内部 QA / synthetic traffic 保持可查询。
4. `localhost` 域名是否存在额外过滤规则，还是本轮主要由 UA 触发。
