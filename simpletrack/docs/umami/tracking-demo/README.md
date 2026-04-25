# Umami Tracking Demo

这个目录现在包含两层验证资产：

- `debug.html` / `index.html`：保留 tracker、track()、identify() 的单点调试能力。
- `site/`：高品质 SimpleTrack SaaS 仿真站，通过真实页面跳转、按钮点击和业务事件喂入 Umami Cloud。

## Files

- `site/`：多页 SaaS 仿真站。
- `site/tracker.js`：共享上报 helper，暴露 `initTracker`、`identifyBusinessUser`、`trackBusinessEvent`、`applySessionProfile`、`runPersonaStep`。
- `site/main.js`：页面识别、点击事件绑定、persona 初始化和表单跳转。
- `debug.html` / `debug.js`：单独验证 tracker、track()、identify()。
- `bulk-send.mjs`：`growth-baseline-x3` 批量事件发送脚本。
- `run-browser-flows.mjs`：使用真实浏览器跑 72 个 persona 的三段式流量。
- `send-event.mjs`：单事件服务端发送脚本。
- `probe-cloud-read-apis.mjs`：登录态下批量抓取 Cloud 读接口和页面空态证据。
- `generate-site.mjs`：仿真站和归档文档生成器。

## Start

```powershell
cd C:\Users\admin\Documents\src\uba\simpletrack\docs\umami\tracking-demo
python -m http.server 4173
```

打开：

```text
http://localhost:4173/site/index.html?websiteId=YOUR_WEBSITE_ID&utm_source=producthunt&utm_medium=launch&utm_campaign=producthunt_launch&cohort=spring_launch&plan=trial&performance=1
```

## Growth Baseline X3

- 648 个逻辑用户
- 1944 个 session
- 72 个真实浏览器 persona
- 576 个批量事件用户
- 每 session 使用 8/10/12 个事件模板之一
- 6 组 campaign、3 组 cohort、4 个 plan
- 54 条收入转化，其中浏览器 6 条，批量 48 条

## Browser Flows

```powershell
npx --yes -p playwright node run-browser-flows.mjs --website-id YOUR_WEBSITE_ID --base-url http://localhost:4173/site
```

## Bulk Send

先 dry run：

```powershell
node bulk-send.mjs --website-id YOUR_WEBSITE_ID --dry-run
```

真实发送：

```powershell
node bulk-send.mjs --website-id YOUR_WEBSITE_ID --hostname localhost --base-url http://localhost:4173/site
```

## Cloud Snapshots

```powershell
npx --yes -p playwright node capture-cloud-snapshots.mjs --website-id YOUR_WEBSITE_ID --headed --login-wait-ms 120000
```

脚本会打开 Umami Cloud analytics 页面；如果当前浏览器没有登录态，会在 headed 模式下等待登录完成，再尝试采集 P07-S01 到 P08-S09，并补采 P08-S05A / P08-S06A 配置态。Attribution 截图会先切到 `Triggered event / checkout_completed` 口径。不要把 `website id`、账号、cookie、token 或 storage state 写入仓库；如需 `--storage-state`，路径必须放在仓库外。
如果检测到 Cloud 登录页，脚本会停止并不写入正式 P07-Sxx / P08-Sxx 截图。
默认输出目录固定为本目录上一级的 `snapshots/`。如果只是复验脚本动作，请用 `--output-root` 指向仓库外临时目录，复验后删除临时截图。
默认会对页面里可见的邮箱账号文本做截图 mask；如果只想复验单张，可追加 `--only P08-S09` 或 `--only attribution`。只有在本地临时排查时才使用 `--no-redact`。

## Cloud Read Probe

```powershell
npx --yes -p playwright node probe-cloud-read-apis.mjs --website-id YOUR_WEBSITE_ID --user-data-dir "%TEMP%\\simpletrack-umami-playwright-profile"
```

这个脚本用于在已登录的浏览器 profile 上抓取 `Overview / Realtime / Events / Sessions / Revenue` 等读接口返回值，帮助判断“写入成功但 Cloud 查询仍为 0”到底是不是前端展示问题。

## API Endpoint Note

- `send-event.mjs` 和 `bulk-send.mjs` 现在默认直发到 `https://api-gateway.umami.dev/api/send`。
- 这是根据真实浏览器里 `https://cloud.umami.is/script.js` 的网络请求校正出来的；它实际命中的也是 `api-gateway.umami.dev`。
- `https://cloud.umami.is/api/send` 即使回 `200 {"beep":"boop"}`，也不应再被当成“已经入库”的充分证据。

## Notes

- 不要把账号、cookie、token 写入仓库。
- `website id` 只通过 URL 参数、localStorage 或命令行参数传入。
- 公开 `/api/send` 不承诺支持历史时间戳回填；42 天跨度使用 `logical_day`、`synthetic_date`、`cohort` 等业务属性表达。
