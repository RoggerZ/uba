# CLIProxyAPI 接入 Codex 说明

## 1. 目标

本文档用于说明当前环境下，如何通过 `CLIProxyAPI` 接入 `Codex`，并作为 `NewAPI` 的上游使用。

当前推荐链路：

`NewAPI -> CLIProxyAPI -> Codex(OpenAI OAuth)`

不推荐的方向：

- 不建议把 `Codex App` 本身当成一个可供 `NewAPI` 反向接入的标准上游服务
- 不建议尝试把网页登录态、Cookie、临时会话直接转换成通用 API Key

## 2. 当前环境

- CLIProxyAPI 路径：`C:\Users\admin\Downloads\CLIProxyAPI_6.9.1_windows_amd64`
- 配置文件：`C:\Users\admin\Downloads\CLIProxyAPI_6.9.1_windows_amd64\config.yaml`
- 建议认证目录：`C:\Users\admin\.cli-proxy-api`
- 推荐监听端口：`8317`
- 当前系统：Windows
- NewAPI 大概率运行在 Docker 容器内

## 3. 配置原则

需要区分两个 Key：

### 3.1 `api-keys`

`config.yaml` 顶层的 `api-keys`，是下游客户端访问 `CLIProxyAPI` 时使用的 Key。

也就是说：

- `NewAPI -> CLIProxyAPI` 时，NewAPI 填的就是这个 Key
- 这个 Key 不是 OpenAI 官方 API Key
- 这个 Key 也不是 Codex 登录态里的 token

### 3.2 `codex-api-key`

`codex-api-key` 仅在“你已经拥有一个第三方 Codex 兼容 API Key 上游”时才需要配置。

当前场景走的是 `OpenAI OAuth 登录 Codex`，因此：

- 不需要配置 `codex-api-key`
- 不建议在当前方案里混用 `codex-api-key`

## 4. 推荐配置文件

适合当前环境的 `config.yaml` 如下：

```yaml
host: "0.0.0.0"
port: 8317

tls:
  enable: false
  cert: ""
  key: ""

remote-management:
  allow-remote: false
  secret-key: ""
  disable-control-panel: true
  panel-github-repository: "https://github.com/router-for-me/Cli-Proxy-API-Management-Center"

auth-dir: "C:/Users/admin/.cli-proxy-api"

api-keys:
  - "sk-cli-proxy-8d5f2ab8d4e74a4fb7d8b8bb6bfa6f27"

debug: false

pprof:
  enable: false
  addr: "127.0.0.1:8316"

commercial-mode: false
logging-to-file: false
logs-max-total-size-mb: 0
error-logs-max-files: 10
usage-statistics-enabled: false

proxy-url: ""
force-model-prefix: false
passthrough-headers: false

request-retry: 3
max-retry-credentials: 0
max-retry-interval: 30

quota-exceeded:
  switch-project: true
  switch-preview-model: true

routing:
  strategy: "round-robin"

ws-auth: true
nonstream-keepalive-interval: 0
codex-instructions-enabled: true

streaming:
  keepalive-seconds: 15
  bootstrap-retries: 1

oauth-excluded-models:
  codex:
    - "gpt-5-codex-mini"
```

## 5. 推荐登录方式

### 5.1 推荐使用 `device flow`

当前环境下，优先推荐：

```powershell
cd C:\Users\admin\Downloads\CLIProxyAPI_6.9.1_windows_amd64
.\cli-proxy-api.exe --config .\config.yaml --codex-device-login
```

原因：

- 不依赖本地浏览器回调端口
- 不依赖 `localhost:1455`
- 对 Windows 本机、代理、浏览器跳转异常场景更稳

执行后通常会出现：

- 一个设备码
- 一个授权地址，通常是 `https://auth.openai.com/codex/device`

然后在浏览器中完成授权，CLIProxyAPI 会自动轮询并保存登录结果。

### 5.2 不推荐优先使用 `--codex-login`

浏览器回调模式命令如下：

```powershell
.\cli-proxy-api.exe --config .\config.yaml --codex-login
```

此模式依赖本地回调：

`http://localhost:1455/auth/callback`

如果浏览器未成功把回调地址打回本机，CLIProxyAPI 就会失败。

## 6. 本次失败的原因

本次日志现象如下：

- 浏览器成功打开 OpenAI 授权页
- 用户已在浏览器中完成认证
- CLIProxyAPI 控制台最终提示：`Authentication failed. Please try again.`

这类问题的本质是：

`OpenAI 已授权成功 -> 浏览器未把 code/state 成功送回 CLIProxyAPI 本地回调`

也就是：

- OpenAI 端大概率已经完成认证
- CLIProxyAPI 没有成功收到本地 `callback`

## 7. 浏览器回调模式的补救办法

如果仍然要使用 `--codex-login`，建议这样执行：

```powershell
.\cli-proxy-api.exe --config .\config.yaml --codex-login --no-browser
```

然后：

1. 手工打开终端打印出来的授权 URL
2. 浏览器授权完成后，如果没有自动跳回成功页
3. 将浏览器地址栏里的完整回调 URL 粘贴回 CLIProxyAPI 终端

正确示例：

```text
http://localhost:1455/auth/callback?code=xxxx&state=yyyy
```

注意：

- 必须粘贴完整 URL
- 不能只粘贴 `code`
- 不能粘贴最开始的 `/oauth/authorize?...` 授权地址
- 少了 `state` 会导致校验失败

## 8. 为什么不建议继续折腾 `oauth-callback-port`

虽然程序支持 `--oauth-callback-port` 参数，但对当前 Codex 浏览器登录流未必有效。

原因是该版本的 Codex OAuth 授权地址里，`redirect_uri` 是固定写成：

`http://localhost:1455/auth/callback`

因此就算你改了本地监听参数，也不一定能改变 OpenAI 最终回调到哪个地址。

所以当前更实际的方案仍然是：

- 优先改用 `--codex-device-login`

## 9. NewAPI 上游配置

CLIProxyAPI 登录成功后，在 NewAPI 中添加上游：

### 9.1 如果 NewAPI 运行在 Docker 容器内

- 上游地址：`http://host.docker.internal:8317/v1`
- 上游 Key：`sk-cli-proxy-8d5f2ab8d4e74a4fb7d8b8bb6bfa6f27`
- 协议：`OpenAI`

### 9.2 如果 NewAPI 运行在 Windows 宿主机

- 上游地址：`http://127.0.0.1:8317/v1`
- 上游 Key：`sk-cli-proxy-8d5f2ab8d4e74a4fb7d8b8bb6bfa6f27`
- 协议：`OpenAI`

### 9.3 模型建议

可优先尝试：

- `gpt-5`
- `gpt-5-codex`

## 10. 登录成功后的检查项

### 10.1 检查认证目录

检查目录：

`C:\Users\admin\.cli-proxy-api`

通常登录成功后，认证文件会保存在这个目录下。

### 10.2 启动服务

完成登录后，正常启动 CLIProxyAPI：

```powershell
cd C:\Users\admin\Downloads\CLIProxyAPI_6.9.1_windows_amd64
.\cli-proxy-api.exe --config .\config.yaml
```

### 10.3 测试模型列表

可以用任意 OpenAI 兼容客户端测试：

```powershell
curl http://127.0.0.1:8317/v1/models ^
  -H "Authorization: Bearer sk-cli-proxy-8d5f2ab8d4e74a4fb7d8b8bb6bfa6f27"
```

如果返回模型列表，说明 `NewAPI -> CLIProxyAPI` 这层已经可用。

## 11. 结论

当前环境下的最佳方案是：

1. 使用上文配置好的 `config.yaml`
2. 使用 `--codex-device-login` 完成 Codex OAuth 登录
3. 启动 CLIProxyAPI 服务
4. 在 NewAPI 中把 CLIProxyAPI 作为 OpenAI 兼容上游接入

一句话总结：

当前问题不是 OpenAI 账号不能用，而是浏览器回调模式不稳定；切换到 `device flow` 是更稳、更省时间的做法。
