# xwl_bi 后端架构参考映射

> 状态：已确定为架构设计参考，不作为代码搬运来源  
> 最近更新：2026-05-17
> 参考目录：`references/xwl_bi-backend/`

## 结论

`references/xwl_bi-backend/` 主要用于参考后端架构设计，不直接复用旧业务代码、旧包名、旧字段名或旧 Vue2 后台。

对 `analytics-core` 有价值的是这些架构思路：

- 采集入口把 HTTP / 请求上下文转换成统一事件消息。
- 队列消费和入库解耦，消费成功语义由真正写入完成决定。
- 批量写入抽象出公共 batch core，失败批次要恢复，不能静默丢数据。
- 明细写入、实时写入、验收状态分别建模，避免一张表承载所有语义。
- ClickHouse 高吞吐明细写入优先走 native batch writer。
- 分析能力通过统一 command / service 边界进入，后续再落到统一 query builder。

## 参考点映射

| xwl_bi 架构点 | 参考文件 | analytics-core 映射 |
| --- | --- | --- |
| report_server 运行时装配 | `cmd/report_server/runtime.go` | 后续 `cmd/collect-api` 只负责装配 collect handler、EventBus、metadata resolver 和观测能力。 |
| 采集主流程编排 | `controller/report_ingress_handler.go` | `internal/collect` 采用“标准化请求 -> 校验 -> 形成 EventEnvelope -> 发布 EventBus”的编排。 |
| sinker 运行时装配 | `cmd/sinker/internal/runner/runtime.go` | 后续 `cmd/worker` 装配 RedisStreamBus/KafkaBus、ingestion processor、ClickHouse writer、Realtime writer 和 status writer。 |
| Kafka mark offset | `platform-basic-libs/sinker/kafka_sarama.go` | `analytics-core` 已在 Kafka provider 内保持“处理完成或 DLQ 成功后才 mark”的语义；Redis Stream 继续用 pending/ack 模拟同类 provider-owned completion。 |
| 顺序提交与完成门 | `cmd/sinker/internal/runner/ordered_commit.go`、`report_completion_gate.go` | Kafka provider 已参考该思路实现 per-partition ordered commit 与 message completion gate，后续只继续补生产硬化和 rebalance 集成验证。 |
| 批量器公共行为 | `platform-basic-libs/service/consumer_data/batch_core.go` | 后续实现通用 batch writer：swap buffer、失败恢复、定时 flush、大小触发、异步 flush 去重。 |
| ClickHouse 明细写入 | `platform-basic-libs/service/consumer_data/reportdata2ck.go` | `EventWriter` 默认使用 `clickhouse-go/v2 PrepareBatch`，按 TableRouter 分表后批量写入。 |
| 实时写入 | `platform-basic-libs/service/consumer_data/real_time_warehousing.go` | 后续 `RealtimeWriter` 单独建模，不与明细 writer 混成一个职责。 |
| 验收状态 | `platform-basic-libs/service/consumer_data/report_accpet_status.go` | 后续 `ingestion_status` 记录成功/失败、错误原因、消息 id 和处理阶段，用于排障。 |
| 分析命令入口 | `platform-basic-libs/service/analysis/interface.go` | `internal/analysis` 保留 events、funnels、retention、paths 等服务边界，但查询统一进入 query builder。 |

## 不采用的部分

- 不采用 `xwl_` 字段命名。
- 不采用 `appid` 单维分表模型，改为 `tenant_id / project_id / source_id`。
- 不采用全局数据库变量和全局 producer。
- 不采用旧后台权限、菜单、页面、控制器叙事。
- 不在 `references/xwl_bi-backend/` 内改代码。

## 对 P1-002 的影响

P1-002 继续按以下顺序推进：

1. `collect` handler：接收标准请求，调用 `Normalize`，发布到 `EventBus`。
2. ClickHouse `EventWriter`：复用 `TableRouter`，优先实现 native batch writer 边界。
3. Realtime / Events 查询：先提供最小查询接口，能证明数据已经入库。
4. `ingestion_status`：记录成功、失败和死信前后的排障信息。

这个顺序来自 xwl_bi 的架构启发，但实现必须保持 `analytics-core` 的业务无关命名和接口边界。
