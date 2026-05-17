# Local Kafka Production SASL_SSL Drill

> 适用范围：只用于本机 disposable 3 broker Kafka `SASL_SSL` 演练，补齐 `KAFKA-PROD-005` 的生产形态认证拓扑证据。
> 生成物位置：repo 根目录 `.tmp/kafka-auth-prod/`，包含私钥、keystore、JAAS 和测试密码，不提交。

## 目标

这个目录把单 broker `local-auth-kafka` drill 扩展为 3 broker `SASL_SSL` 环境，用于验证 `src/analytics-core/eventbus/kafka` 在更接近生产的认证拓扑下仍然满足：

- Go client 使用 CA PEM 验证 TLS server certificate。
- Sarama 使用 `SASL/PLAIN` 用户名和密码连接 broker。
- broker 间通信也走 `SASL_SSL` internal listener。
- replicated topic 使用 3 partitions、replication factor 3、`min.insync.replicas=2`。
- 既有 gated integration tests 能完成 publish / consume、ordered commit、restart replay、replicated publish / consume 和 broker outage。

## 启动

在父仓根目录执行：

```powershell
.\simpletrack\docs\任务计划\kafka-first-eventbus\local-auth-kafka-prod\Start-KafkaSaslSslProdDrill.ps1 -RegenerateMaterial
```

脚本会使用 `confluentinc/cp-kafka:7.5.0` 镜像生成 disposable 证书材料，并启动 3 broker KRaft controller quorum：

- `simpletrack-kafka-auth-prod-1` -> `127.0.0.1:39193`
- `simpletrack-kafka-auth-prod-2` -> `127.0.0.1:39194`
- `simpletrack-kafka-auth-prod-3` -> `127.0.0.1:39195`

## 验证 analytics-core

```powershell
. .\.tmp\kafka-auth-prod\certs\analytics-core-auth-prod-env.ps1
Set-Location .\src\analytics-core
go test .\eventbus\kafka -run 'TestKafkaIntegration(PublishConsume|OrderedCommitWaitsForEarlierOffset|UncommittedMessageReplaysAfterRestart)$|TestKafkaReplicatedIntegration(PublishConsume|SurvivesBrokerOutage)$' -count=1 -v
```

如果已经 `Set-Location` 到 `src/analytics-core`，请使用绝对路径加载 env 文件：

```powershell
. 'C:\Users\admin\Documents\src\uba\.tmp\kafka-auth-prod\certs\analytics-core-auth-prod-env.ps1'
go test ./eventbus/kafka -run 'TestKafkaIntegration(PublishConsume|OrderedCommitWaitsForEarlierOffset|UncommittedMessageReplaysAfterRestart)$|TestKafkaReplicatedIntegration(PublishConsume|SurvivesBrokerOutage)$' -count=1 -v
```

这个 env 会设置：

- `ANALYTICS_CORE_KAFKA_BROKERS=127.0.0.1:39193,127.0.0.1:39194,127.0.0.1:39195`
- `ANALYTICS_CORE_KAFKA_REPLICATED_INTEGRATION=1`
- `ANALYTICS_CORE_KAFKA_OUTAGE_INTEGRATION=1`
- `ANALYTICS_CORE_KAFKA_OUTAGE_STOP_CONTAINER=simpletrack-kafka-auth-prod-3`
- `ANALYTICS_CORE_KAFKA_TOPIC_PARTITIONS=1`
- `ANALYTICS_CORE_KAFKA_TOPIC_REPLICATION_FACTOR=3`
- `ANALYTICS_CORE_KAFKA_TOPIC_MIN_INSYNC_REPLICAS=2`

不要把 `ANALYTICS_CORE_KAFKA_TOPIC_PARTITIONS` 改成 `3` 后再跑 ordered commit 测试；该测试需要单 partition 来证明同一 partition 内 offset 不会乱序提交。replicated tests 会在测试内部强制 3 partitions。

## 本机验证记录

- 2026-05-17：`analytics-core` commit `d835341` 下，本机 disposable 3 broker KRaft `SASL_SSL` 环境 `127.0.0.1:39193,39194,39195` 通过 `TestKafkaIntegrationPublishConsume`、`TestKafkaIntegrationOrderedCommitWaitsForEarlierOffset`、`TestKafkaIntegrationUncommittedMessageReplaysAfterRestart`、`TestKafkaReplicatedIntegrationPublishConsume`、`TestKafkaReplicatedIntegrationSurvivesBrokerOutage`。
- 启动命令：`Start-KafkaSaslSslProdDrill.ps1 -TimeoutSeconds 300`。
- 验证命令：`. 'C:\Users\admin\Documents\src\uba\.tmp\kafka-auth-prod\certs\analytics-core-auth-prod-env.ps1'; go test ./eventbus/kafka -run 'TestKafkaIntegration(PublishConsume|OrderedCommitWaitsForEarlierOffset|UncommittedMessageReplaysAfterRestart)$|TestKafkaReplicatedIntegration(PublishConsume|SurvivesBrokerOutage)$' -count=1 -v`。
- 验证结果：五项 integration test 全部 `PASS`，包级输出 `ok github.com/simpletrack/analytics-core/eventbus/kafka 19.394s`。
- Topic detail：`publish-consume`、`ordered-commit`、`restart-replay` 三类 topic 为 1 partition、RF=3、`min.insync.replicas=2`；`replicated-publish-consume`、`replicated-outage` 两类 topic 为 3 partitions、RF=3、`min.insync.replicas=2`。实际 topic 名包含 run id，例如 `analytics.events.integration.replicated-outage.20260517103841.759026800`。
- 清理命令：`Stop-KafkaSaslSslProdDrill.ps1 -RemoveVolumes`。
- 清理结果：`docker ps -a --filter name=simpletrack-kafka-auth-prod` 和 `docker network ls --filter name=simpletrack-kafka-auth-prod` 均为空。
- 关键约束：最初 Zookeeper 方案被 broker JAAS 影响预检，最终改用 KRaft controller quorum；不要把 Zookeeper client auth 问题混进本轮 broker `SASL_SSL` 验收。

## 停止

```powershell
.\simpletrack\docs\任务计划\kafka-first-eventbus\local-auth-kafka-prod\Stop-KafkaSaslSslProdDrill.ps1
```

如需同时清理容器卷：

```powershell
.\simpletrack\docs\任务计划\kafka-first-eventbus\local-auth-kafka-prod\Stop-KafkaSaslSslProdDrill.ps1 -RemoveVolumes
```

## 约束

- 不要提交 `.tmp/kafka-auth-prod/`。
- 不要把运行时生成的 disposable 密码当成生产密码；它只存在于 `.tmp/kafka-auth-prod/`。
- 不要把这个本机 3 broker KRaft drill 等同于最终生产拓扑；它用于验证认证链路、replication 语义和 outage gate。
- 如果 readiness probe 失败，先看 internal listener、advertised listener、证书 SAN 和 Docker 端口占用，不要直接改 Go integration test 放宽语义。
