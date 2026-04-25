# Phase 06 Flow

1. 起点状态：Board 设计页已保存一个组件
2. 用户动作：展开日期范围下拉
3. 页面响应：出现 `Today` 到 `Custom range` 的完整时间范围列表
4. 截图编号：`P06-S01`
5. 用户动作：直接访问高级报告页
6. 页面响应：`Funnels / Journeys / Retention / Realtime` 页面都可以进入并截图
7. 截图编号：`P06-S02` `P06-S03` `P06-S04` `P06-S05`
8. 用户动作：在 demo 侧加大数据上报后回刷 `Realtime`
9. 页面响应：Realtime 已开始展示实时指标、活动流、页面列表和国家分布
10. 截图编号：`P06-S06`
11. 用户动作：打开 `Funnel` 配置弹窗并填写一个示例漏斗
12. 页面响应：出现配置态和结果态
13. 截图编号：`P06-S07` `P06-S08` `P06-S09`
14. 用户动作：继续查看 `Journeys`
15. 页面响应：出现 `/tracking-demo/index.html -> demo_track_call -> demo_signup_click` 的路径图
16. 截图编号：`P06-S10`
17. 观察点：Realtime 是最快的接入验收页；Funnels 需要显式配置 step type 与 step value；Journeys 用来回答“用户下一步去了哪里”
18. 对产品设计的启发：SimpleTrack 可以把 Realtime 作为接入验收页，把 Funnel 设计成“先配对象、再看结果”的明确流程，把 Journeys 放在更靠后的深度分析层
