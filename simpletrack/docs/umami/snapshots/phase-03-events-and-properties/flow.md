# Phase 03 Flow

1. 起点状态：demo 页面已成功加载 tracker
2. 用户动作：依次触发 `data-umami-event`、`track()` 和 `identify()`
3. 页面响应：demo 状态区记录触发完成
4. 截图编号：`P03-S01`
5. 用户动作：回到 Events 页面，查看完整布局
6. 页面响应：指标区、tabs、filter、date range 都可见
7. 截图编号：`P03-S02` `P03-S04`
8. 用户动作：依次切换 `Activity`、`Properties`，再打开 `Filter` 和日期范围
9. 页面响应：Activity log、Properties、Filter Fields、日期范围都可见
10. 截图编号：`P03-S05` `P03-S06` `P03-S07` `P03-S08`
11. 用户动作：继续放大数据量并回刷 Events
12. 页面响应：事件图表和事件明细开始展示 `demo_track_call / demo_signup_click`
13. 截图编号：`P03-S09`
14. 观察点：Realtime 比 Events 更快，但 Events 在数据量足够时会进入稳定展示
15. 对产品设计的启发：SimpleTrack 可以把 Realtime 作为接入验收页，把 Events 作为聚合分析页
