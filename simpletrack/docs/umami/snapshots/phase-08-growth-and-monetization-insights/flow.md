# Phase 08 Flow

1. 先跑浏览器 persona，保证 pageview、session、replay、performance 有真实入口。
2. 再跑 growth-baseline-x3 批量事件，补足高密度转化、收入和属性样本。
3. 配置或打开 Funnels、Goals、Segments、Cohorts 等需要对象的页面。
4. 按 UTM -> signup -> install -> first event -> checkout 顺序复验 Revenue 和 Attribution。
5. 把受 Cloud 账号能力限制的项目明确记录为限制，不写成已完成。
6. 2026-04-25 19:40 复核时，Cloud analytics URL 曾跳转登录页；已归档 `P08-B00` 作为历史阻塞证据。
7. 恢复登录态后重新采集 P08-S01 到 P08-S09，并对顶部账号和左下账号区做邮箱脱敏。
8. 2026-04-25 晚间修正为普通 Chrome User-Agent 后补跑真实流量和批量事件，Journeys、Retention、UTM、Revenue 已进入有数据态。
9. 通过 Cloud UI 创建 `Growth Baseline Checkout Funnel`，steps 为 `pricing_viewed -> checkout_started -> checkout_completed`，重新采集 `P08-S01`。
10. 通过 Cloud UI 创建 `Producthunt Launch Segment` 和 `Paid Checkout Cohort`，并分别补采 `P08-S05 / P08-S05A` 与 `P08-S06 / P08-S06A`。
11. 在 Attribution 页面将类型切到 `Triggered event`、conversion step 填为 `checkout_completed`，重新采集 `P08-S09`；`Replays` 仍按 Business plan 限制记录。
