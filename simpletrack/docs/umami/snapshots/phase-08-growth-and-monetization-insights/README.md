# Phase 08: Growth And Monetization Insights

## 目标

记录 Funnels、Journeys、Retention、Replays、Segments、Cohorts、UTM、Revenue、Attribution 在三倍样本下的页面证据。

## 截图清单

| 编号 | 页面 | 说明 | 当前状态 |
| --- | --- | --- | --- |
| P08-B00 | Cloud Login | analytics URL 被重定向到登录页的阻塞证据 | 已截图 |
| P08-S01 | Funnels | pricing -> signup -> install -> first_event -> checkout | 已截图；`Growth Baseline Checkout Funnel` 显示 `1.68k -> 45 -> 45 visitors` |
| P08-S02 | Journeys | 高意图流量路径图 | 已截图；路径流有真实非零数据 |
| P08-S03 | Retention | 三组 cohort 的回访表现 | 已截图；留存矩阵已有 cohort 数据 |
| P08-S04 | Replays | 真实浏览器会话回放 | 已截图；Business plan 限制 |
| P08-S05 | Segments | 命名流量切片 | 已截图；已保存 `Producthunt Launch Segment` |
| P08-S05A | Segments / Config | `producthunt_launch` campaign 切片 | 已截图；配置为 `UTM Campaign is producthunt_launch` |
| P08-S06 | Cohorts | spring_launch / self_serve_wave / paid_pilot | 已截图；已保存 `Paid Checkout Cohort` |
| P08-S06A | Cohorts / Config | checkout cohort 定义 | 已截图；配置为 `Triggered event checkout_completed`，日期范围 `Last 90 days` |
| P08-S07 | UTM | 6 组 campaign 的流量表现 | 已截图；Campaign / Source / Medium 有数据 |
| P08-S08 | Revenue | 54 条收入转化 | 已截图；当前累积站点数据显示 `$11.86k / 355 orders` |
| P08-S09 | Attribution | campaign 到 checkout revenue 的归因 | 已截图；切换到 `Triggered event / checkout_completed` 后显示 `49 visitors / 49 visits / 57 views` |

## 备注

本阶段已完成正式截图。若后续新增更多 Segments / Cohorts 保存对象、调整 Funnel step 或配置其它 Attribution 转化条件，需要继续补结果态截图并同步索引。

2026-04-25 复核结论：P08-B00 到 P08-S09 均已采集并重新完成邮箱脱敏，且补充 `P08-S05A / P08-S06A` 两张配置态截图。修正 User-Agent 后，`Journeys / Retention / UTM / Revenue` 已出现真实结果；随后通过 Cloud UI 补建 `Growth Baseline Checkout Funnel`、`Producthunt Launch Segment`、`Paid Checkout Cohort`，并将 Attribution 转化条件切到 `checkout_completed`。当前只有 `Replays` 页面明确提示需要 `Business plan`，且网站元数据为 `replayEnabled=false`。
