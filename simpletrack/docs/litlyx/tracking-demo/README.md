# Litlyx Tracking Demo

这个目录现在已经补成一个最小可运行验证工件，用于后续复查 Litlyx 的真实接入链路。

## 文件

- `index.html`：本地 demo 页面
- `app.js`：运行时加载浏览器脚本，并触发两个示例自定义事件
- `send-event.mjs`：从终端发送单条测试事件
- `bulk-send.mjs`：批量发送多用户、多 session 的示例事件

## 设计原则

- 不把真实 workspace ID 写死到仓库文件里
- 运行时再手动粘贴 `Workspace ID`
- 页面端只使用官方文档明确给出的接入方式：
  - 浏览器脚本：`<script defer data-workspace="workspace_id" src="https://cdn.jsdelivr.net/npm/litlyx-js@latest/browser/litlyx.js"></script>`
- 自定义事件：`Lit.event("your_event_name", { metadata: ... })`

本轮已实测通过：

- 浏览器脚本成功加载
- `Lit.event(...)` 成功发送自定义事件
- 批量事件脚本成功写入 `https://broker.litlyx.com/event`
- 后台 `Product` 与 `Raw Events` 均已看到回写结果

## 启动方式

在当前目录启动一个静态服务：

```bash
python -m http.server 4174
```

然后打开：

```text
http://localhost:4174/index.html
```

## Demo 步骤

1. 在 Litlyx 后台 `Web` 或 `Settings / General` 里复制 `Workspace ID`
2. 打开本地 demo 页面
3. 粘贴 `Workspace ID`
4. 点击 `Load tracker`
5. 刷新一次页面，生成自动 pageview
6. 分别点击 `Fire event A` 和 `Fire event B`
7. 回到 Litlyx 后台观察：
   - `Product`
   - `Raw Data`
   - `Reports / Product Report`

## 单条事件发送

```bash
node send-event.mjs ^
  --pid YOUR_WORKSPACE_ID ^
  --name demo_terminal_event ^
  --metadata "{\"source\":\"terminal\",\"surface\":\"tracking-demo\"}"
```

默认会请求官方文档里的事件入口：

```text
https://broker.litlyx.com/event
```

## 批量事件发送

```bash
node bulk-send.mjs ^
  --pid YOUR_WORKSPACE_ID ^
  --users 6 ^
  --sessions-per-user 2
```

它会发送多组示例事件，帮助 Product 和 Reports 更快进入可观察状态。

## 本轮状态

- 已完成：demo 页面和终端脚本落盘
- 已完成：使用当前工作区完成一轮真实验证，后台 `Product` 与 `Raw Events` 均已看到 14 条事件
- 未完成：正式 Reports 生成链路仍受当前 Free trial 权限限制，不能用本 demo 继续触发 PDF 生成

## 参考资料

- [Litlyx Easy Setup](https://docs.litlyx.com/universal)
- [Litlyx Custom Events](https://docs.litlyx.com/custom-events)
