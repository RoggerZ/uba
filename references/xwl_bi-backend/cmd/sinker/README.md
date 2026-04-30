# sinker 命令行说明

## 1. 服务启动

正常启动 `sinker` 服务时，命令和以前一致：

```bash
sinker -configFileDir config -configFileName config -configFileExt json
```

如果不传参数，默认读取：

- `config/config.json`

### 1.1 保护场景配置入口

当前 `sinker` 的保护演练场景已经统一收敛到**单一配置文件入口**：

- `scripts/config/config.json`

旧的场景文件：

- `scripts/config/config.consumer-low.json`
- `scripts/config/config.protect-mock.json`
- `scripts/config/config.consumer-low-soft-backlog.json`

已经移除，不再作为主入口。

现在如果要切换保护相关场景，请在同一份配置里通过：

- `sinker.protection.mock.preset`

来选择。

当前支持的 preset：

1. `""`
   - 空值
   - 不展开任何额外场景覆盖
   - 这是默认生产入口
2. `consumer_low`
   - 把 `reportConsumerPool` / `reportPersistPool` 收缩到低消费单 worker 场景
   - 不启用 mock
3. `protect_mock`
   - 在 `consumer_low` 基础上启用 mock 注入
   - 用于演练 `softLimited / hardPaused`
   - 这个场景要求 `mock.enabled=true`
4. `consumer_low_soft_backlog`
   - 在 `consumer_low` 基础上下调 soft backlog 阈值
   - 用于验证“纯 backlog 足以驱动 soft_limited”的测试画像
   - 这个场景要求 `mock.enabled=false`

示例：

```json
{
  "sinker": {
    "protection": {
      "mock": {
        "preset": "protect_mock",
        "enabled": true
      }
    }
  }
}
```

再例如：

```json
{
  "sinker": {
    "protection": {
      "mock": {
        "preset": "consumer_low_soft_backlog",
        "enabled": false
      }
    }
  }
}
```

说明：

1. `preset` 不是弱提示，而是启动期强约束。
2. 非法 `preset` 会导致启动失败。
3. 如果同时手写了和 `preset` 冲突的 worker pool / threshold / mock 字段，也会导致启动失败，而不是静默覆盖。
4. `preset` 只建议用于测试/演练，不建议直接带进生产主配置。

## 2. 诊断控制模式

`sinker` 还支持诊断控制模式。  
这时它**不会启动消费者服务**，而是作为 HTTP client 调用已经运行中的 `sinker` admin 接口，执行完成后立即退出。

命令前缀固定为：

```bash
sinker diagnostic ...
```

## 3. 显式参数模式

### 开启诊断

```bash
sinker diagnostic enable --duration 3m --trace-offset 123 --report-handler-threshold 5s --admin-addr http://127.0.0.1:8094 --admin-token secret --output text
```

参数说明：

- `--duration`
  - 可选
  - 例如 `3m`、`30s`、`1h`
  - 不传时由服务端使用默认值
- `--trace-offset`
  - 可选
  - 指定后只跟踪这个 offset
  - 不传时只开启诊断日志，不开启 trace
- `--admin-addr`
  - 可选
  - 默认是 `http://127.0.0.1:8094`
- `--admin-token`
  - 可选
  - 远程访问时需要带
- `--report-handler-threshold`
  - 可选
  - 仅 `enable` 可传
  - 使用 Go duration 语法，例如 `500ms`、`2s`、`5s`
  - 不传时，本次诊断会话里的 `report handler stage timing` 慢日志阈值仍使用默认 `2s`
- `--output`
  - 可选
  - 默认是 `text`
  - 可选值：`text`、`json`

### 关闭诊断

```bash
sinker diagnostic disable --admin-addr http://127.0.0.1:8094 --admin-token secret --output text
```

### 查看状态

```bash
sinker diagnostic status --admin-addr http://127.0.0.1:8094 --admin-token secret --output json
```

## 4. JSON 模式

如果希望把一次诊断请求写成一个 JSON 块，可以用 `-json`。

### 开启诊断

```bash
sinker diagnostic -json "{\"action\":\"enable\",\"duration\":\"3m\",\"traceOffset\":123,\"reportHandlerStageTimingThreshold\":\"5s\",\"adminAddr\":\"http://127.0.0.1:8094\",\"adminToken\":\"secret\",\"output\":\"text\"}"
```

### 关闭诊断

```bash
sinker diagnostic -json "{\"action\":\"disable\",\"adminAddr\":\"http://127.0.0.1:8094\",\"adminToken\":\"secret\",\"output\":\"text\"}"
```

### 查看状态

```bash
sinker diagnostic -json "{\"action\":\"status\",\"adminAddr\":\"http://127.0.0.1:8094\",\"adminToken\":\"secret\",\"output\":\"json\"}"
```

JSON 字段说明：

- `action`
  - 必填
  - 只能是 `enable`、`disable`、`status`
- `duration`
  - 仅 `enable` 时可传
  - 使用 Go duration 语法，例如 `3m`
- `traceOffset`
  - 仅 `enable` 时可传
- `reportHandlerStageTimingThreshold`
  - 仅 `enable` 时可传
  - 使用 Go duration 语法，例如 `500ms`、`2s`
- `adminAddr`
  - 可选
- `adminToken`
  - 可选
- `output`
  - 可选
  - 默认是 `text`
  - 可选值：`text`、`json`

## 5. 诊断接口鉴权规则

- 如果 admin 服务绑定在 loopback 地址，例如 `127.0.0.1`，本机请求默认免 token
- 如果 admin 服务绑定成可远程访问地址，远程请求必须带 `X-Admin-Token`
- 如果配置成非 loopback 绑定但没有配置 `adminToken`，服务启动会直接失败

## 6. 说明

- `SINKER_DIAGNOSTIC_LOG` 和 `SINKER_TRACE_OFFSET` 仍兼容启动时环境变量初始化
- 运行中动态开关请优先使用 `sinker diagnostic ...`
- `report handler stage timing` 的慢日志阈值默认是 `2s`，也可以在 `diagnostic enable` 时临时覆盖；`disable` 或会话过期后会自动回到 `2s`
- `SINKER_REPORT_DIRECT_EXEC` 仍然是启动期模式选择，不支持运行中热切换
- 保护状态切换请优先通过 `sinker protect status|enable|disable|set` 观察或控制，不要再依赖旧场景配置文件切换
