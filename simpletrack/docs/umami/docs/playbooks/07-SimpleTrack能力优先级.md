# SimpleTrack 能力优先级

## 目标

把 Umami 的能力拆成适合 SimpleTrack 演进的三档路线，避免一次性把系统做得过重。

## 适用场景

- 做 MVP 范围规划
- 准备从“基础统计”升级到“行为分析”
- 讨论哪些能力可以后置

## 涉及能力

- MVP 能力：[01-安装与接入](../01-安装与接入.md)、[02-采集与事件](../02-采集与事件.md)、[05-Realtime](../05-Realtime.md)、[14-Filters](../14-Filters.md)、[09-Goals](../09-Goals.md)
- 增强版能力：[04-Sessions](../04-Sessions.md)、[07-Compare](../07-Compare.md)、[08-Breakdown](../08-Breakdown.md)、[10-Funnels](../10-Funnels.md)、[11-Journeys](../11-Journeys.md)、[15-Segments](../15-Segments.md)、[17-UTM](../17-UTM.md)、[20-Boards-Links-Pixels-Teams-API](../20-Boards-Links-Pixels-Teams-API.md)
- 后置能力：[06-Performance](../06-Performance.md)、[12-Retention](../12-Retention.md)、[13-Replays](../13-Replays.md)、[16-Cohorts](../16-Cohorts.md)、[18-Revenue](../18-Revenue.md)、[19-Attribution](../19-Attribution.md)

## 推荐顺序

1. 先做接入与基础验证
2. 再做事件和基础分析
3. 然后做分群和转化分析
4. 最后补营销归因、协作和高级诊断

## 每步判断点

1. 没有稳定采集，就不要先做复杂洞察
2. 没有统一事件语义，就不要先做 Funnels 和 Attribution
3. 没有明确团队协作场景，就不要急着做复杂 board/pixel/link 体系

## 常见坑

- 把高级报表做出来了，但没有稳定的基础数据供给
- 先上很多配置页面，结果新手根本不知道从哪里开始
- 把 Boards、Teams、Attribution 这些进阶能力放进 MVP，拖慢基础体验

## SimpleTrack 可直接借鉴的决策

### MVP

- Website 创建与最短接入链路
- Tracker snippet 与最小事件埋点
- Overview、Realtime、Events
- 基础 Filters
- 简单 Goal

### 增强版

- Breakdown、Compare
- Sessions
- Segments
- Funnels、Journeys
- 基础 UTM
- 默认 Dashboard + 可保存看板

### 后置能力

- Cohorts、Retention
- Revenue、Attribution
- Performance、Replays
- Pixels、Links、Teams、公开分享

## 依赖关系

- `Funnels` 依赖稳定的事件命名和 step 定义
- `Journeys` 依赖足够丰富的路径和事件数据
- `Retention`、`Cohorts` 依赖稳定的人群定义与时间维度
- `Revenue`、`Attribution` 依赖 UTM、转化事件和金额数据
- `Boards` 的价值要建立在已有可复用分析组件之上

## 分层建议表

| 阶段 | 目标 | 建议优先实现 | 先不要急着做 |
| --- | --- | --- | --- |
| MVP | 让用户尽快接入并相信数据是活的 | Website、Tracking code、Overview、Realtime、Events、基础 Filters、简单 Goal | Replays、Attribution、复杂 Team 权限 |
| 增强版 | 让用户开始回答“为什么”和“哪里掉了” | Breakdown、Compare、Sessions、Segments、Funnels、Journeys、基础 UTM、默认 Dashboard | Revenue、多模型归因、复杂公开分享 |
| 后置能力 | 让系统支持商业分析、体验诊断和协作扩展 | Cohorts、Retention、Revenue、Attribution、Performance、Replays、Pixels、Links、Teams | 一开始就做全量入口堆叠 |

## 什么时候该升级阶段

- 当用户已经能稳定看到 Overview / Realtime / Events，而且事件命名开始收敛时，可以从 `MVP` 升到 `增强版`
- 当用户已经明确提出“我要看渠道效果、收入归因、留存和性能问题”时，再从 `增强版` 升到 `后置能力`
- 如果还在反复修“数据为什么不进来”，就不要急着做高级分析页
