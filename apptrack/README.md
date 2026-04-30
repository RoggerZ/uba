# AppTrack

`AppTrack` 是一个面向独立移动开发者和小型 App 团队的移动分析产品方案，核心目标是用更低门槛、更低成本的方式，提供接近“可行动洞察”的分析能力。

## 产品定位

核心方向：

- 面向移动应用的行为分析
- 5 分钟级别的 SDK 集成体验
- AI 自动分析、异常识别和改进建议
- 固定价格、适合独立开发者预算

目标用户主要包括：

- 独立 iOS / Android 开发者
- 2 到 10 人的小型 App 团队
- 预算有限的初创项目
- 需要比 Firebase 更主动、比 Mixpanel 更便宜的分析方案的团队

## 当前目录

```text
apptrack/
├── docs/
│   ├── AppTrack-AI驱动的移动分析工具方案.md
│   ├── App行为追踪竞品分析报告.md
│   ├── 产品方向决策-移动分析市场定位.md
│   ├── 市场验证与获客策略.md
│   ├── 成本控制与盈利策略.md
│   ├── 技术实现方案-架构与开发指南.md
│   ├── 推广营销策略-移动开发者社区增长.md
│   └── countly/
│       ├── 各平台集成与部署文档
│       └── src/vue-countly-demo/
└── README.md
```

## 文档导航

### 核心方案

- `docs/AppTrack-AI驱动的移动分析工具方案.md`
  产品总方案，包含目标用户、价值主张、核心功能和商业方向。
- `docs/技术实现方案-架构与开发指南.md`
  技术方案草案，包含 Go 后端、移动 SDK、前端控制台和基础部署建议。

### 市场与定位

- `docs/App行为追踪竞品分析报告.md`
  用于理解移动分析赛道的竞品格局。
- `docs/产品方向决策-移动分析市场定位.md`
  用于确定产品切入点和差异化策略。

### 增长与经营

- `docs/市场验证与获客策略.md`
- `docs/成本控制与盈利策略.md`
- `docs/推广营销策略-移动开发者社区增长.md`

## Countly 参考资料

`docs/countly/` 目录不是 `AppTrack` 的正式产品代码，而是参考研究与验证资料，主要用于：

- 拆解 Countly 的产品和能力边界
- 参考多平台 SDK 的集成方式
- 验证移动分析产品的技术可行性

其中：

- `docs/countly/src/README.md` 提供示例目录说明
- `docs/countly/src/vue-countly-demo/` 是一个 Vue 3 + Vite + Countly Web SDK 的参考 demo

如果需要运行该 demo，可在对应目录执行：

```bash
npm install
npm run dev
```

## 建议阅读顺序

1. `docs/AppTrack-AI驱动的移动分析工具方案.md`
2. `docs/产品方向决策-移动分析市场定位.md`
3. `docs/App行为追踪竞品分析报告.md`
4. `docs/技术实现方案-架构与开发指南.md`
5. `docs/countly/` 下的参考文档与 demo

## 当前状态

- 当前以产品方案和竞品研究为主
- 已有较明确的 SDK 与后端架构设想
- 包含一份可参考的 Countly Web demo
- 尚未整理为正式的生产级多模块工程

如果后续开始真正实现，建议先确定：

- 事件模型与埋点规范
- iOS / Android SDK 最小能力集
- 后端事件接收与聚合链路
- 控制台 MVP 页面与数据看板范围
