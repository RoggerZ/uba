# xwl_bi 能否作为 SimpleTrack 核心引擎

## Q：SimpleTrack 的核心引擎能否考虑使用 xwl_bi 现成技术？

A：可以，但现在已经不是“直接拿 xwl_bi 当 SimpleTrack 核心引擎”，而是 **P1 确定抽取 xwl_bi 的分析数据面核心，形成独立业务无关仓库 `analytics-core`**。

`analytics-core` 只关注分析功能本身，不处理 SimpleTrack 的定价、团队、订阅、账单、onboarding、营销页和企业控制台。

## Q：`analytics-core` 具体要包含哪些能力？

A：P1 的底座设计要覆盖这些分析数据面能力：

- 事件分析。
- 漏斗分析。
- 留存分析。
- 路径分析。
- LTV 分析。
- 归因分析。
- 分群。
- 会话。
- 元数据。
- 实时数据。
- 事件属性。
- 用户属性。

这些能力不代表 P1 产品界面全部开放。P1 产品闭环仍优先 tracker、collect、Realtime、Events、Website settings 和 Goal 最小闭环；但 `analytics-core` 的仓库边界、模块命名和接口设计要从一开始容纳这些分析能力。

## Q：Umami 和 Litlyx 对 `analytics-core` 有什么参考价值？

A：它们是两个不同层面的参考。

- **Umami** 更适合参考分析对象体系：Realtime、Events、Filters、Segments、Funnels、Journeys、Retention、Attribution 这些能力边界比较清楚，也有稳定的事件语义和 distinct id 思路。
- **Litlyx** 更适合参考新用户首价值：短接入链路、Raw Events 验收、Product 空态/示例态/真实态、`Show test data` 和 docs 引导更利于让用户快速理解“数据已经进来了”。

所以 `analytics-core` 的底层能力边界参考 Umami 的体系化分析模型；P1 的接入验收和排障体验参考 Litlyx 的 Raw Events 与示例数据教育方式。

落实到 SimpleTrack：

1. Realtime 参考 Umami，作为“现在有没有数据”的最快反馈。
2. Raw Events / Events 参考 Litlyx，作为“事件到底有没有入库”的排障入口。
3. Funnels、Journeys、Retention、Segments 的模块边界参考 Umami，但产品页面放到 P2 以后。
4. Product / Overview 的空态、示例态和真实态参考 Litlyx，但示例数据不能替代真实接入验收。

## Q：为什么仓库名叫 `analytics-core`，不能带 SimpleTrack 或 xwl？

A：因为这个仓库应该是业务无关的分析核心，不是 SimpleTrack 的私有业务层，也不是 xwl_bi 的整仓改名。

- 不叫 `simpletrack-analytics-core`：避免核心仓库被 SimpleTrack 的套餐、团队、订阅、UI 和客户模型绑定。
- 不叫 `xwl-bi-core`：避免和原作者仓库、旧业务后台、旧命名习惯混在一起。
- 只叫 `analytics-core`：强调它是可复用、高性能、业务无关的数据面核心。

## Q：为什么不建议把 xwl_bi 整仓直接改名？

A：`github.com/1340691923/xwl_bi` 是原作者的上游仓库名称。整仓改名会把上游来源、旧 Vue2 后台、旧业务模型、SimpleTrack 新业务和核心引擎改造混在一起，后续很难维护。

更合适的做法是：

1. 保留 xwl_bi 作为上游来源和参考。
2. 新建独立仓库 `analytics-core`。
3. 抽取采集、事件、漏斗、留存、路径、LTV、归因、分群、会话、元数据、实时数据、事件属性、用户属性等分析数据面核心。
4. SimpleTrack 引用 `analytics-core`，在 SimpleTrack 自己的产品层处理 workspace、site、定价、团队、订阅、账单和 onboarding。

## Q：代码层面要怎么处理旧命名？

A：需要做深度优化设计，而不是机械搬运。

- 剔除 `xwl_` 变量前缀。
- 剔除 xwl 相关包名、函数名、模块名。
- 将命名改成更专业、业务无关的分析领域语言，例如 `event`、`session`、`cohort`、`funnel`、`retention`、`attribution`、`metadata`。
- 重新梳理目录结构，让采集、队列、存储、查询、分析模型、元数据边界清楚。
- 删除或隔离旧业务后台、旧权限、旧菜单、旧页面逻辑。

## Q：旧的 Vue2 后台界面还要复用吗？

A：不复用。旧 Vue2 后台界面明确不作为 SimpleTrack 产品界面，也不进入 `analytics-core`。

最多只参考它背后的字段、查询意图和数据转换逻辑。SimpleTrack 控制台应由 Next.js + SaaS 模板承接，保持低装饰、高密度、信息层次稳定的企业分析控制台风格。

## Q：为什么说 xwl_bi 适合做数据面？

A：行为分析产品最难的是事件采集、事件入库、查询聚合和高吞吐写入。xwl_bi 已经围绕这些问题使用了比较典型的分析型技术栈：

- ClickHouse：适合大规模事件明细和聚合查询。
- MySQL：适合元数据、配置和事务型数据。
- Redis：适合缓存、限流、会话和轻量队列。
- Kafka：适合高吞吐事件流和异步写入。

这些能力适合抽成 `analytics-core`，但不适合连同旧后台一起照搬成 SimpleTrack。

## Q：P1 是否必须启用 Kafka？

A：不必须。Kafka 很适合高吞吐，但会增加早期运维复杂度。当前确定的方向是：**前期使用 Redis Stream 替代 Kafka，但不删除 Kafka 代码，保留 KafkaBus 作为后续高吞吐实现。**

建议在 `analytics-core` 中抽象 `EventBus`：

- `DirectBus`：本地开发和低流量场景，直接处理。
- `RedisStreamBus`：P1 优先实现，适合早期自部署和轻量异步写入。
- `KafkaBus`：保留高吞吐实现，后续事件量上来后启用。

## Q：`analytics-core` 支持哪些技术栈？

A：底层技术栈确认支持：

- Kafka。
- MySQL。
- Redis。
- ClickHouse。

前期部署可以只启用 Redis Stream + MySQL + ClickHouse，降低运维复杂度；Kafka 作为可插拔实现保留。

## Q：spike 是什么意思？

A：spike 是短周期技术验证，通常用 0.5 天到 2 天确认某条路线是否可行。现在 `analytics-core` 已经确定要做，spike 的目的不再是判断“要不要做”，而是降低实施风险。

`analytics-core` spike 可以这样定义：

1. 从 xwl_bi 映射一个最小采集入口。
2. 抽象 `EventBus`。
3. 用 Redis Stream 写一条事件流。
4. 写入 ClickHouse 或 mock event store。
5. 用简化查询接口读出事件。
6. 不接入旧 Vue2 后台，不做套餐、团队、支付。

## Q：这件事当前是什么状态？

A：已确定，P1 执行。

确定方向是：新建独立业务无关仓库 `analytics-core`；从 xwl_bi 抽取分析数据面核心；前期 Redis Stream 替代 Kafka；KafkaBus 保留；旧 Vue2 后台界面不复用；SimpleTrack 在自己的产品层承接定价、团队、订阅、账单、onboarding 和企业控制台。

具体实施方案见 `simpletrack/docs/实施决策/analytics-core实施方案.md`。
