# Phase 05 Flow

1. 起点状态：进入 `Shields`
2. 页面响应：默认落在 Domains allow list，对应 `P05-S01`
3. 用户动作：切到 `IP addresses`
4. 页面响应：进入 IP 排除标签页，对应 `P05-S02`
5. 用户动作：切到 `Bot traffic`
6. 页面响应：进入 bot 流量治理标签页，对应 `P05-S03`
7. 用户动作：回到 Domains 并点击 `Add domain`
8. 页面响应：出现支持通配符的 allow list 弹窗，对应 `P05-S04`
9. 用户动作：进入 `Analyst`
10. 页面响应：先看到默认聊天态，再展开示例问题，对应 `P05-S05` `P05-S06`
11. 用户动作：进入 `Plans`
12. 页面响应：先看到 Personal 套餐，再切到 Business 并展开 FAQ，对应 `P05-S07` `P05-S08` `P05-S09`
13. 用户动作：进入顶部 `Share links`
14. 页面响应：看到 `Shareable links` 管理页，当前 `0 active`，可选域名、可选 public/protected link、可填描述，对应 `P05-S10`
15. 安全备注：`Create link` 会创建新的只读访问入口，本轮只记录页面态，不触发创建
16. 用户动作：直接访问 `Members`
17. 页面响应：页面长时间停在 `Loading... Please wait. If this takes too long, contact the project owner.`，对应 `P05-S11`
18. 设计启发：Litlyx 把治理、AI、商业化、分享都做成独立层，但 Members 当前不可用态的解释不如 Reports / SEO 清楚
