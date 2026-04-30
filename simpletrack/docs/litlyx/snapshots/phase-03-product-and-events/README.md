# Phase 03 README

## Goal

记录 `Product` 模块的默认骨架、示例数据模式、`Setup events` 外链行为，以及本地 demo 回写后的真实出数表现。

## Preconditions

- 已登录并可进入 `Product`

## Snapshot Order

1. Product 默认空态
2. 打开 `Show test data`
3. 观察 Product 下半屏 Funnel / User Flow / Metadata 区域
4. 进入 `Raw Data`
5. 点击 `Setup events` 并确认 Custom Events 文档新标签页
6. 运行本地 `tracking-demo` 与批量事件脚本
7. 回到 Product 和 Raw Events 复核真实事件入库

## Observation Focus

- 空态是否仍然保留完整分析框架
- 示例数据是否足以教育用户理解功能
- Raw Data 路由是否稳定
- `Setup events` 是站内配置入口还是外部文档入口
- 真实自定义事件是否能同时进入聚合页和明细表
