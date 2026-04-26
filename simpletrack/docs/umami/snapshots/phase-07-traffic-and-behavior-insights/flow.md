# Phase 07 Flow

1. 启动 tracking-demo 静态服务并打开 site/index.html。
2. 携带 websiteId、UTM、plan、cohort、performance 参数运行 72 个浏览器 persona。
3. 在 Umami Cloud 中依次打开 Sessions、Realtime、Performance。
4. 运行 growth-baseline-x3 批量事件后打开 Compare、Breakdown、Goals、Filter。
5. 记录每个页面是否进入有数据状态，以及是否需要额外 Cloud 配置。
6. 2026-04-25 19:40 复核时，Cloud analytics URL 曾跳转登录页；已归档 `P07-B00` 作为历史阻塞证据。
7. 恢复登录态后重新采集 P07-S01 到 P07-S07，并对顶部账号和左下账号区做邮箱脱敏。
8. 2026-04-25 晚间修正为普通 Chrome User-Agent 后补跑真实流量和批量事件，Sessions、Realtime、Performance、Compare、Breakdown、Filter 已进入有数据态。
9. 通过 Cloud UI 创建 `Checkout Completed Goal`，类型为 `Triggered event`，value 为 `checkout_completed`，重新采集 `P07-S06`。
10. 在 Compare 页面应用已保存的 `Producthunt Launch Segment`，采集 `P07-S08`，确认命名 Segment 可以回到分析页复用并触发指标重算。
