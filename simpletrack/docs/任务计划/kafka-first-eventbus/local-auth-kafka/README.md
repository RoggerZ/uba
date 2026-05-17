# Local Kafka SASL_SSL Drill

> 适用范围：只用于本机 disposable Kafka TLS/SASL 演练，补齐 `KAFKA-PROD-003` 的真实认证 broker 证据。
> 生成物位置：repo 根目录 `.tmp/kafka-auth/`，包含私钥、keystore、JAAS 和测试密码，不提交。

## 目标

这个目录提供一个最小的单 broker KRaft `SASL_SSL` Kafka 环境，用于验证 `src/analytics-core/eventbus/kafka` 的认证 broker 路径：

- Go client 使用 CA PEM 验证 TLS server certificate。
- Sarama 使用 `SASL/PLAIN` 用户名和密码连接 broker。
- 既有 gated integration tests 能在认证 broker 下完成 publish、consume、ordered commit 和 replay。

## 启动

在父仓根目录执行：

```powershell
.\simpletrack\docs\任务计划\kafka-first-eventbus\local-auth-kafka\Start-KafkaSaslSslDrill.ps1 -RegenerateMaterial
```

脚本会使用本机已有的 `confluentinc/cp-kafka:7.5.0` 镜像生成证书材料，因此 Windows 主机不需要安装 `openssl` 或 `keytool`。

## 验证 analytics-core

```powershell
. .\.tmp\kafka-auth\certs\analytics-core-auth-env.ps1
Set-Location .\src\analytics-core
go test .\eventbus\kafka -run 'TestKafkaIntegration(PublishConsume|OrderedCommitWaitsForEarlierOffset|UncommittedMessageReplaysAfterRestart)$' -count=1 -v
```

如果已经 `Set-Location` 到 `src/analytics-core`，请用绝对路径加载 env 文件，避免相对路径落到子仓目录后导致测试被 skip：

```powershell
. 'C:\Users\admin\Documents\src\uba\.tmp\kafka-auth\certs\analytics-core-auth-env.ps1'
go test ./eventbus/kafka -run 'TestKafkaIntegration(PublishConsume|OrderedCommitWaitsForEarlierOffset|UncommittedMessageReplaysAfterRestart)$' -count=1 -v
```

通过后，才可以把 `KAFKA-PROD-003` 的 live 证据从“待外部环境回填”更新为“本机 disposable SASL_SSL broker 已验证”。

## 本机验证记录

- 2026-05-17：`analytics-core` commit `d835341` 下，本机 disposable `SASL_SSL` broker `127.0.0.1:39093` 通过 `TestKafkaIntegrationPublishConsume`、`TestKafkaIntegrationOrderedCommitWaitsForEarlierOffset`、`TestKafkaIntegrationUncommittedMessageReplaysAfterRestart`。
- 启动命令：`Start-KafkaSaslSslDrill.ps1 -TimeoutSeconds 120`。
- 验证命令：`. 'C:\Users\admin\Documents\src\uba\.tmp\kafka-auth\certs\analytics-core-auth-env.ps1'; go test ./eventbus/kafka -run 'TestKafkaIntegration(PublishConsume|OrderedCommitWaitsForEarlierOffset|UncommittedMessageReplaysAfterRestart)$' -count=1 -v`。
- 关键约束：readiness probe 在容器内必须使用 internal listener `kafka-auth:9094`，宿主机 Go 测试继续使用 external listener `127.0.0.1:39093`。

## 停止

```powershell
.\simpletrack\docs\任务计划\kafka-first-eventbus\local-auth-kafka\Stop-KafkaSaslSslDrill.ps1
```

如需同时清理容器卷：

```powershell
.\simpletrack\docs\任务计划\kafka-first-eventbus\local-auth-kafka\Stop-KafkaSaslSslDrill.ps1 -RemoveVolumes
```

## 约束

- 不要提交 `.tmp/kafka-auth/`。
- 不要把 `simpletrack-secret` 当成生产密码；它只存在于本地 disposable drill。
- 不要设置 `ANALYTICS_CORE_KAFKA_TLS_INSECURE_SKIP_VERIFY=true` 作为正式证据。
- 这个环境是单 broker，只验证 TLS/SASL 认证路径；多 broker / outage 仍用 `ANALYTICS_CORE_KAFKA_REPLICATED_INTEGRATION=1` 和 `ANALYTICS_CORE_KAFKA_OUTAGE_INTEGRATION=1` 的独立演练。
