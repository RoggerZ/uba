# Phase 03 Flow

1. 起点状态：进入 `Product`
2. 页面响应：展示 `Top Events / Top 5 / Events / Funnel / User Flow / Metadata` 等空态骨架，对应 `P03-S01`
3. 用户动作：点击 `Show test data`
4. 页面响应：Product 立即进入 demo mode，对应 `P03-S02`
5. 用户动作：观察下半屏
6. 页面响应：`Funnel Analysis` 空态已在下半屏留档，对应 `P03-S04`
7. 用户动作：进入 `Raw Data`
8. 页面响应：早期自动化访问里多次停留在 loading flame，对应 `P03-S03`
9. 额外验证：DOM 探测确认 Raw Data 页面底层确实存在 `Domain / Name / Metadata / Date / Session` 表格
10. 用户动作：点击 `Setup events`
11. 页面响应：当前页 URL 不变，但新标签页打开 `https://docs.litlyx.com/custom-events`，标题为 `Custom Events - Litlyx Docs`，对应 `P03-S07`
12. 代码补验：登录态 Nuxt chunk 将 `Setup events` 渲染为 `target="_blank"` 的 Custom Events 文档外链；它不是站内配置向导
13. 用户动作：运行本地 `tracking-demo` 并补充批量事件
14. 页面响应：Product 出现 `Total events: 14` 与四类真实事件，对应 `P03-S05`
15. 页面响应：`Raw Events` 表格展示 14 条 `localhost` 事件，对应 `P03-S06`
