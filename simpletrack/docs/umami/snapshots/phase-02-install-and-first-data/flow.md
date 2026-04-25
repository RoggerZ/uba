# Phase 02 Flow

1. 起点状态：website 已创建但仍为空站点
2. 用户动作：进入 website Overview 和 settings 页面
3. 页面响应：Overview 空态与 settings 中的 tracking code 同时可见
4. 截图编号：`P02-S01` `P02-S02`
5. 用户动作：继续拆看 settings
6. 页面响应：依次看到 `Website basics / Tracking code / Share / Danger zone`
7. 截图编号：`P02-S05` `P02-S06` `P02-S07` `P02-S08`
8. 用户动作：打开本地 demo，输入 `website id` 并加载 tracker
9. 页面响应：demo 页面记录 tracker 已加载
10. 截图编号：`P02-S03` `P02-S04`
11. 观察点：接入链路很短，但首次数据展示存在延迟或未刷出的问题
12. 对产品设计的启发：SimpleTrack 需要更明确的“接入成功但数据处理中”反馈
