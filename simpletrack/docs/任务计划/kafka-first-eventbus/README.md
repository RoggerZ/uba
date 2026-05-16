# Kafka-First EventBus 任务计划

> 状态：实现已完成，阶段评审与正式文档同步中
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

Kafka-first EventBus 的代码实现已恢复并推进到阶段性交付点：`analytics-core` 已完成公共契约调整、Redis/direct 适配和 Sarama Kafka provider；`analytics-service` 已完成 Kafka 配置与 runtime 装配；本地 compose 已加入单节点 Kafka。当前剩余工作是处理阶段性子代理评审结论，并在任务全部完成后把精简结论同步到正式实施决策文档。
