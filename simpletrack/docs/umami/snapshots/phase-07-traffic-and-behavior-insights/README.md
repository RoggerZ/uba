# Phase 07: Traffic And Behavior Insights

## 目标

记录 Sessions、RealTime、Performance、Compare、BreakDown、Goals、Filter 在增长基线三倍样本下的页面证据。

## 截图清单

| 编号 | 页面 | 说明 | 当前状态 |
| --- | --- | --- | --- |
| P07-B00 | Cloud Login | analytics URL 被重定向到登录页的阻塞证据 | 已截图 |
| P07-S01 | Sessions | 真实浏览器 persona 形成会话列表 | 已截图；会话列表和 session count 有数据 |
| P07-S02 | Realtime | 浏览器流量和 CTA 事件实时出现 | 已截图；`Views / Visitors / Events / Countries` 有数据 |
| P07-S03 | Performance | 页面加载性能数据进入视图 | 已截图；LCP / FCP / TTFB 等指标有数据 |
| P07-S04 | Compare | 按 plan、campaign、cohort 对比 | 已截图；当前周期指标和路径表有数据 |
| P07-S05 | BreakDown | 按来源、计划、cohort 拆分 | 已截图；路径拆分表有数据 |
| P07-S06 | Goals | signup、first event、checkout 目标 | 已截图；`Checkout Completed Goal` 显示 `49 / 1.73k`、转化率 `3%` |
| P07-S07 | Filter | Fields、Segments、Cohorts 过滤入口 | 已截图；Filter 弹窗可用，底层 Compare 有数据 |

## 备注

本阶段已完成正式截图。若后续调整筛选口径，需要继续补交互态截图并同步索引。

2026-04-25 复核结论：P07-B00 到 P07-S07 均已采集并重新完成邮箱脱敏。最初的空态来自批量脚本使用自定义 User-Agent，切换到普通 Chrome UA 并补跑后，`Sessions / Realtime / Performance / Compare / Breakdown / Filter` 都已经有可解释数据。`Goals` 已创建 `Checkout Completed Goal`，当前截图显示 `checkout_completed` 目标为 `49 / 1.73k`，转化率 `3%`。
