# xwl_bi 的 consumer offset 和 acceptance status 是什么

## Q：consumer offset 是什么？

A：在 Kafka 里，consumer offset 是消费组对某个 topic 分区的消费进度。它表示“这个消费组已经确认处理到哪里了”。

你的理解是对的：业务上可以把它理解成消费端 commit offset，也就是 ack。

原则是：

1. 消费端拿到消息。
2. 业务处理成功。
3. 提交 offset。
4. 下次从新 offset 继续消费。

如果业务没成功，不应该提交 offset。

## Q：consumer offset 还需要吗？

A：需要。只是它不应该进入业务事件协议。

在 `analytics-core` 里建议这样处理：

- KafkaBus 内部保留 Kafka offset。
- RedisStreamBus 内部保留 stream id / pending / ack 状态。
- 对上层统一抽象成 `checkpoint` 或 `ack`。
- `EventEnvelope` 这种业务事件结构里不放 `source_offset`。

也就是说：**消费进度需要保留，但它属于队列 adapter 内部和 ingestion 状态，不属于事件本身。**

## Q：acceptance status 是什么？

A：xwl_bi 的 acceptance status 可以理解为“事件入库验收状态”。它记录某条消息或某批消息是否成功完成写入，以及失败原因、消费位置、处理阶段等排障信息。

它的价值是：

- 判断事件是否真的进入 ClickHouse。
- 排查为什么某条事件没进来。
- 记录失败原因。
- 帮助对账消费进度和入库状态。

## Q：acceptance status 还需要吗？

A：需要。尤其 SimpleTrack / analytics-core P1 要做接入验收和 Raw Events，必须有一个能回答“事件为什么没进来”的状态表。

建议在 `analytics-core` 中保留为 `ingestion_status`：

| 字段 | 含义 |
| --- | --- |
| `event_id` | 事件唯一 ID |
| `tenant_id` | 租户 |
| `project_id` | 项目 |
| `source_id` | 数据源 |
| `status` | `accepted` / `failed` / `dead_lettered` |
| `stage` | 失败或成功发生在哪个阶段 |
| `error_reason` | 失败原因 |
| `attempt_count` | 尝试次数 |
| `queue_type` | `redis_stream` / `kafka` / `direct` |
| `checkpoint` | 队列内部消费进度，调试用 |
| `created_at` / `updated_at` | 状态时间 |

## Q：那为什么之前说 `source_offset` 可以去掉？

A：去掉的是事件协议里的 `source_offset` 字段，不是去掉 offset / checkpoint 能力。

原因是：

- `source_offset` 对 Kafka 有意义，但对 Redis Stream 是 stream id，对 DirectBus 可能不存在。
- 如果写进事件模型，会把事件协议和队列实现绑死。
- 更好的位置是 `ingestion_status.checkpoint` 或 KafkaBus / RedisStreamBus adapter 内部。

## Q：提出这两个问题，是不是有别的想法？

A：有，核心想法是分层：

- `EventEnvelope`：只放业务事件事实。
- `EventBus adapter`：管理 Kafka offset、Redis stream id、ack。
- `ingestion_status`：记录事件处理验收和排障状态。

这样既保留 xwl_bi 中有价值的 offset / acceptance status 思路，又不会把 Kafka 专属字段污染到通用事件协议。

## Q：当前确认的结论是什么？

A：确认采用这个分层思路。

- consumer offset / Redis Stream id / ack 仍然需要，但只属于 EventBus adapter 和消费 checkpoint。
- acceptance status 仍然需要，但在 `analytics-core` 中改名为 `ingestion_status`。
- `EventEnvelope` 不放 `source_offset`，避免事件协议绑定 Kafka 或 Redis 的内部进度模型。
- `ingestion_status` 同时承担接入验收、失败排障和幂等入库的职责。
