# Litlyx 采集、事件与 Product

## 当前结论

Litlyx 已经被这轮调研验证了最小闭环：

1. 浏览器脚本可以接入
2. 自定义事件可以发出
3. Product 聚合页会出数
4. Raw Events 明细表会出现真实事件

## Product 的三层状态

### 空态

即使没数据，Product 也会先摆出：

- `Top Events`
- `Funnel Analysis`
- `Events User Flow`
- `Analyze event metadata`

这不是“空白页”，而是分析骨架。

### 示例态

`Show test data` 能立刻把 Product 变成可讲解状态，适合新手教育。

### 真实态

通过 `tracking-demo` 和 `bulk-send.mjs`，后台已经看到：

- `Total events: 14`
- 4 类真实事件名
- Raw Events 表格里的 14 条 `localhost` 事件

## `Setup events` 的真实含义

`Setup events` 看起来像站内配置入口，但当前实证结果是：

- 点击后会打开 `https://docs.litlyx.com/custom-events`
- 新标签页标题为 `Custom Events - Litlyx Docs`
- 它本质上是文档外链，不是站内事件向导

这是一处很典型的产品误导点：文案像“继续配置”，实际动作是“去看文档”。

## Raw Events 的价值

这轮调研里，Raw Events 不是附属功能，而是证明接入真的成功的关键页面。只看 Product 聚合还不够，明细表才是排障最快入口。

## 对 SimpleTrack 的启发

- Product 应同时支持空态骨架、示例态和真实态
- 一定要保留 Raw Events 这类明细页
- `Setup events` 这类 CTA 必须说明真实去向
- 如果真实数据暂时不足，示例态可以先帮助用户理解系统价值

## 当前证据

- `P03-S01`
- `P03-S02`
- `P03-S03`
- `P03-S04`
- `P03-S05`
- `P03-S06`
- `P03-S07`
- [../tracking-demo/README.md](../tracking-demo/README.md)

## 当前缺口

- 真实 metadata 分析结果还没有单独截图沉淀
- Marketing 来源样本目前还不够厚，后续可以再用带 UTM 的访问补齐
