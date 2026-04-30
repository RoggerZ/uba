# Phase 04 Flow

1. 起点状态：进入 `Marketing`
2. 页面响应：渠道、来源、社交模块全部可见，但数据为 0，对应 `P04-S01`
3. 用户动作：点击 `Show test data`
4. 页面响应：Marketing 进入 demo mode，曲线和来源榜单被激活，对应 `P04-S02`
5. 用户动作：点击 `Generate UTM link`
6. 页面响应：出现完整 UTM 生成弹窗，对应 `P04-S03`
7. 用户动作：进入 `Reports`
8. 页面响应：展示按周期和模板生成报告的中心页，对应 `P04-S04`
9. 用户动作：点击报告卡片右上角的 `Sample`
10. 页面响应：在当前页上层打开 `Easy report sample` PDF 预览，对应 `P04-S05`
11. 额外验证：6 类正式报告卡片当前都处于 disabled 状态，底部 `Generate report` 也不可点击，对应 `P04-S07`
12. 代码补验：前端存在 `generate_pdf / generate_pdf_cust / generate_pdf_adv / generate_pdf_seo / generate_pdf_product / generate_pdf_marketing` 等 POST 生成路径，但当前账户只能预览样张，不能触发正式生成
13. 用户动作：进入 `SEO`
14. 页面响应：看到 premium gate 和升级 CTA，对应 `P04-S06`
