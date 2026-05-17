# Kafka-First EventBus 任务计划

> 状态：已完成
> 创建时间：2026-05-17
> 适用范围：`src/analytics-core`、`src/analytics-service` 以及本任务相关的本地运行依赖与文档同步。

## 目录用途

这个目录是 Kafka-First EventBus 改造的轻量任务计划目录，用来承接本任务的阶段计划、进度表、踩坑记录和执行规范。

原 `AGENTS.md` 指定的实施决策文档仍是长期知识库，但本任务执行期间不再“改一点代码就写一大段总文档”。所有过程性记录先写在本目录；任务全部执行完成后，再把本目录的结论精简汇总到根目录 / `AGENTS.md` 指定的正式文档：

- `simpletrack/docs/实施决策/README.md`
- `simpletrack/docs/实施决策/分阶段实施计划.md`
- `simpletrack/docs/实施决策/analytics-core实施方案.md`
- 如仍有未拍板事项，再同步到 `simpletrack/docs/实施决策/待评审事项.md`

## 文档索引

| 文档 | 用途 | 更新频率 |
| --- | --- | --- |
| [分阶段实施计划.md](分阶段实施计划.md) | 记录 Kafka-First EventBus 的阶段、交付物、验证和评审门槛 | 阶段开始或范围变化时更新 |
| [进度表.md](进度表.md) | 记录当前阶段状态、证据、下一步动作 | 阶段节点和明显状态变化时更新 |
| [踩坑记录.md](踩坑记录.md) | 记录执行过程中遇到的真实问题、根因和处理动作 | 遇到坑或解除 blocker 时更新 |
| [执行规范.md](执行规范.md) | 记录本任务覆盖 AGENTS.md 的轻量执行节奏 | 规则变化时更新 |
| [待办事项.md](待办事项.md) | 记录本任务完成后拆出的 Kafka 生产硬化和验证后续项 | 新增、关闭或调整后续任务时更新 |
| [生产级认证与压测计划.md](生产级认证与压测计划.md) | 记录生产级多 broker `SASL_SSL`、认证 benchmark、Kafka exporter/SLO 和 SCRAM/OAuth/Kerberos 的下一阶段计划 | 阶段 10-13 推进时更新 |

## 当前原则

1. 代码推进优先，文档低频汇总。
2. 父代理负责思考、拆分、合并和最终判断。
3. 需要并行执行时再使用子代理；子代理只做明确分工内的执行和轻度思考。
4. 全程保持 goal 模式；阶段目标不完成就不标记完成。
5. `code-review`、`code-simplifier`、`ai-slop-cleaner` 放在阶段性交付、准备提交或核心边界变更收口时使用，不在每个小代码改动后机械触发。
6. 阶段性交付完成后，必须把本目录的精简结论同步到 AGENTS.md 指定的正式实施决策文档。
7. commit 和 push 都按阶段节奏降频；可本地保留阶段内改动，阶段完成或改动较多时再 commit，push 更低频。
8. push 遇到网络问题时，优先使用本机代理 `http://localhost:64320` 重试多次。

## 当前状态

Kafka-first EventBus 的代码实现、生产硬化、服务级 diagnostics / metrics 出口、本机 `SASL_SSL` drill、生产级认证 drill 和认证 broker benchmark 均已有阶段记录：`analytics-core` `d835341` 已完成 Kafka provider、生产默认、rebalance-safe ordered commit、SASL/TLS options、真实 broker integration、replicated/outage gate、本机 disposable `SASL_SSL` integration 和真实 broker publish benchmark 入口；`analytics-service` `c6d1139` 已完成 Kafka runtime、diagnostics JSON route 和 Prometheus text metrics route。父仓 `c31750e` 已记录并推送本机认证 drill 证据；父仓 `d444717` 新增 3 broker KRaft `SASL_SSL` drill，并在 `127.0.0.1:39193,39194,39195` 下跑通五条目标 integration tests；父仓 `pending commit` 记录同一认证环境下 `BenchmarkKafkaIntegrationPublish` 的默认三次和固定 `10x` 复测结果。下一步按 [生产级认证与压测计划.md](生产级认证与压测计划.md) 继续推进 broker/exporter SLO 指标源和 SCRAM/OAuth/Kerberos 路线评审。

后续待办统一维护在 [待办事项.md](待办事项.md)。
