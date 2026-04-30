# Gemini Artifact: SimpleTrack Enterprise Start

## Original User Task

基于 `simpletrack/docs/umami/docs/` 中对 Umami 官方文档的解读，结合现有 `simpletrack/prototype/simpletrack-umami-inspired/` 静态原型，重新规划一套面向企业级生产级方向的 SimpleTrack 原型与功能阶段方案。项目先圈定功能、形成总纲、分阶段推进，后端逐步按阶段细化。

## Final Prompt Sent To Gemini CLI

你是一个资深企业级 SaaS 产品架构顾问。请基于仓库中的 Umami 解读、现有原型和阶段路线，为 SimpleTrack 重新规划一套“企业级、生产级、可逐步落地”的产品原型与功能阶段方案。输出产品定位、信息架构、P0-P4 阶段总纲、后端阶段路线、第一套新原型页面清单、防坑指南和一个最高杠杆问题。

## Gemini Output (Raw)

Gemini 的核心判断是：SimpleTrack 最大风险是被现有静态原型拖入功能堆砌，企业级产品的核心不是页面多，而是数据流严谨、模型可扩展、边界清晰。

产品定位建议：面向现代 SaaS 团队与独立开发者的高确定性轻量化分析中枢，主打 5 分钟极速接入和数据可信。早期不是 CDP、不做全量用户画像、不碰 PII，不做录屏回放，不做营销自动化，不做复杂 BI 或 SQL 暴露。

现有原型建议：

- `login` 保留并加固。
- `create-site`、`install`、`first-data` 合并为 onboarding wizard。
- `index` 保留并收敛为站点级 dashboard。
- `events` 与 `insights` 合并为行为与属性分析。
- `funnels` 后置到 P2。
- `team` 后置到 P4。
- `settings` 重写，加入数据字典与 ingestion rules。

阶段建议：

- P0 契约：冻结 Event Schema、Property Dictionary、UTM、隐私基线，不写 UI/API 代码。
- P1 基石：注册登录、新建站点、核心 dashboard、Realtime、单事件分析、属性分布、Simple Goal；不做漏斗、归因、团队权限、复杂联表过滤。
- P2 诊断：Breakdown、Compare、Funnels、Journeys、Segments。
- P3 价值：Cohorts、Retention、Revenue、基础 UTM 归因。
- P4 扩展：Teams/RBAC、Open API Key、Share URL、Performance。

后端建议：

- P1 领域对象：Workspace、Website、EventData、Session。
- 写入侧：SDK -> `/api/collect` -> Validator -> Queue。
- 查询侧：通过 BFF 获取图表数据，不让前端拼复杂聚合。
- 配置侧：账号、站点、鉴权。
- P1 可先用 PostgreSQL JSONB 跑通链路，P2 再考虑 ClickHouse 等列式存储。

第一套新原型建议：

- Auth & Onboarding：Stepper、snippet、listening 状态。
- Dashboard：时间选择、核心 KPI、趋势、Top Pages、Top Referrers、Realtime 活动。
- Events Analytics：事件字典列表，事件趋势和属性拆解。
- Settings：tracker 设置、Data Dictionary、Ingestion Rules。

防坑：

- 不要收集无结构化 anything，事件属性保持扁平且类型强约束。
- V1 原型克制，P1 后端给不出的按钮不要画成正式功能。
- 不要过早纠结跨端 Distinct ID 归一化。

最高杠杆问题：

> 目标用户成功部署 SimpleTrack 代码后的前 5 分钟里，最迫切想确认的唯一一个最重要业务数据状态是什么？

## Concise Summary

Gemini 建议把 SimpleTrack 新原型收敛为“接入可信链路”而不是功能菜单扩写。P1 只证明数据管道通了，并通过 Realtime、Events、属性分布和 Simple Goal 建立信任。Funnels、Journeys、Revenue、Attribution、Team、Replays、Performance、Boards、Share URL、API Key 都应后置。

## Action Items / Next Steps

1. 新原型写入 `simpletrack/prototype/simpletrack-enterprise-mvp/`，不覆盖旧原型。
2. P1 页面收敛到 onboarding、dashboard/realtime、events/properties、goals、settings/data dictionary。
3. 视觉风格采用企业级生产工作台：去 UI 风、去营销感、去炫技装饰，强调表格、状态、审计、配置和操作边界。
4. 后端规划先按采集、验证、队列、查询 BFF、配置、目标定义拆模块，不写具体实现。
