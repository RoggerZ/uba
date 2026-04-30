# Deep Interview Spec: SimpleTrack Enterprise Start

## Metadata

- Profile: standard
- Context type: brownfield
- Final ambiguity: 0.16
- Threshold: 0.20
- Context snapshot: `.omx/context/simpletrack-enterprise-start-20260428T084500Z.md`
- Interview transcript: `.omx/interviews/simpletrack-enterprise-start-20260428T094705Z.md`

## Intent

SimpleTrack 正式进入项目启动阶段。当前目标不是直接写后端，而是先圈定功能、形成总纲、明确阶段和边界，并在新目录重写一套企业级生产级 P1 原型。

## Desired Outcome

第一版要围绕“数据管道活了”建立信任：用户安装 tracker 后，应能快速确认 pageview/event 已进入系统，并能在 Realtime 和 Events/Properties 中看到可信证据。

## In Scope For P1 Prototype

- 新建站点和安装 tracker 的连贯接入流程。
- Realtime 数据进入状态。
- Overview 中的基础趋势和健康状态。
- Events / Properties 事件列表、事件详情、属性分布。
- Simple Goal，强调成功动作和分母口径。
- Settings / Data Dictionary / Ingestion Rules，用于体现生产级数据契约。

## Out Of Scope / Non-goals For P1

- Team / RBAC / 邀请 / 站点归属权限。
- Funnels / Journeys，后置到 P2。
- Revenue / Attribution。
- Replays / Performance。
- Boards / Share URL / API Key。
- 复杂 BI、自定义 SQL、任意 JSON 深层属性、PII 收集。

## Decision Boundaries

已授权直接定稿：

- 页面清单与导航。
- 视觉与交互风格。
- Mock 数据口径。
- 后端模块边界。
- 文案与中英文命名。

## Visual And Interaction Constraint

视觉方向必须去 UI 风，采用企业级生产工作台风格：

- 不做营销感 hero，不做炫彩渐变，不做装饰性大卡片。
- 优先呈现状态、表格、日志、配置、规则、时间戳和健康检查。
- 信息密度中高，但保持可读和稳定布局。
- 控件语义直接，强调“当前系统状态”和“下一步操作”。
- 色彩克制：以中性背景、深色文字、有限状态色和单一主操作色为主。

## Testable Acceptance Criteria

- 新原型位于 `simpletrack/prototype/simpletrack-enterprise-mvp/`，不覆盖旧原型。
- 打开原型第一屏即为产品工作台，不是营销 landing page。
- 页面导航只包含 P1 能力：接入、概览/实时、事件、目标、设置。
- UI 不出现 Team、Revenue、Attribution、Replay、Performance、Board、Share URL、API Key 的正式功能入口。
- Funnels/Journeys 不作为 P1 正式能力出现。
- 至少提供 onboarding -> dashboard -> events -> goals -> settings 的可点击流程。
- Mock 数据字段来自既有事件/字段字典：`plan`、`campaign`、`cohort`、`role`、`workspaceSize`，事件包括 `signup_completed`、`install_started`、`sdk_install_completed`、`first_event_sent` 等。

## Backend Phase Framing

P1 后端只做最小可信链路：

- Collect API: 接收 pageview/event。
- Validator: 校验 website、域名、事件名、属性类型、隐私规则。
- Event Store: 保存事件事实。
- Realtime Read Model: 支撑最近活动和活跃访问。
- Events Query: 事件列表、事件详情、属性分布。
- Goal Definition: 简单页面型或事件型目标。
- Settings / Data Dictionary: 站点配置、允许域名、事件字典、属性规则。

P2 之后再进入 Funnels/Journeys、Breakdown、Compare、Segments 等诊断能力。
