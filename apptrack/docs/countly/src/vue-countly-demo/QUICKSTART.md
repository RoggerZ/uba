# 🚀 快速启动指南

> 5分钟快速运行Vue Countly Demo

## 📋 前置要求

- Node.js 18+ （推荐使用最新LTS版本）
- npm 或 yarn
- Countly服务器（已部署并运行）

## ⚡ 快速开始

### 步骤1：安装依赖

```bash
cd apptrack/docs/countly/src/vue-countly-demo
npm install
```

### 步骤2：配置Countly（可选）

**当前配置已设置好**，可以直接使用：

```javascript
// src/main.js 中的配置
app.use(countlyPlugin, {
  appKey: '2963dcc87fe3f000725a16c9ada61f707ca546dd',
  url: 'http://localhost:8380',
  debug: true
});
```

**如果需要修改**，编辑 `src/main.js` 和 `index.html`：

1. **修改 src/main.js**：
   ```javascript
   appKey: 'YOUR_NEW_APP_KEY',
   url: 'http://YOUR_SERVER:PORT',
   ```

2. **修改 index.html 中的 noscript 标签**：
   ```html
   <img src='http://YOUR_SERVER:PORT/pixel.png?app_key=YOUR_NEW_APP_KEY&begin_session=1' />
   ```

### 步骤3：运行开发服务器

```bash
npm run dev
```

看到以下输出表示成功：

```
  VITE v5.x.x  ready in xxx ms

  ➜  Local:   http://localhost:5173/
  ➜  Network: use --host to expose
```

### 步骤4：打开浏览器

访问：`http://localhost:5173`

## 🧪 测试功能

### 1. 基础测试（必做）

1. **打开浏览器开发者工具**（F12）
2. **查看Console**，应该看到：
   ```
   ✓ Countly plugin installed successfully
   ```
3. **点击"发送简单事件"按钮**
4. **查看事件日志**，应该显示：
   ```
   [时间] ✓ 简单事件已发送: button_clicked
   ```

### 2. 验证数据（必做）

1. **登录Countly管理界面**：`http://localhost:8380`
2. **进入Dashboard > Real-time**
3. **应该看到**：
   - 1个在线用户
   - 刚才发送的事件

### 3. 验证新增功能（推荐）

**滚动追踪**：
- 在页面上下滚动
- 查看：`Analytics > User Behavior > Scrolls`

**表单追踪**：
- 填写并提交表单
- 查看：`Analytics > Events` 中的 `[CLY]_form_*` 事件

### 3. 完整测试（可选）

按照应用界面上的各个功能模块逐一测试：

- ✅ 基础事件（3个按钮）
- ✅ 用户属性（填写表单）
- ✅ 页面浏览（3个页面）
- ✅ 错误追踪（2个按钮）
- ✅ 表单测试（提交表单）

## 🐛 常见问题

### 问题1：npm install失败

**错误**：网络超时或依赖安装失败

**解决**：

```bash
# 使用国内镜像
npm config set registry https://registry.npmmirror.com
npm install
```

### 问题2：CORS跨域错误

**错误**：浏览器Console显示CORS错误

**原因**：Countly服务器未配置CORS

**解决**：

参考 `PORT_CONFIGURATION.md` 中的Nginx配置，添加CORS头。

或者临时解决：在Countly服务器的docker-compose.yml中添加环境变量。

### 问题3：事件未发送

**错误**：点击按钮后没有反应

**排查**：

1. **检查配置**：
   ```bash
   # 查看src/main.js
   # 确认appKey和url已正确替换
   ```

2. **检查Countly服务器**：
   ```bash
   # 确认服务器正在运行
   docker ps
   # 应该看到countly相关容器
   ```

3. **检查网络**：
   - 打开浏览器开发者工具
   - 切换到Network标签
   - 点击按钮
   - 查看是否有请求发送到Countly服务器

### 问题4：端口5173已被占用

**错误**：`Port 5173 is already in use`

**解决**：

修改 `vite.config.js`：

```javascript
export default defineConfig({
  plugins: [vue()],
  server: {
    port: 3000  // 改为其他端口
  }
});
```

## 📊 验证数据

### 在Countly管理界面查看

1. **实时数据**
   ```
   Dashboard > Real-time
   - 查看当前在线用户
   - 查看实时事件
   ```

2. **事件列表**
   ```
   Analytics > Events
   - button_clicked
   - button_clicked_with_params
   - timed_action
   - form_submitted
   ```

3. **用户信息**
   ```
   Analytics > Users
   - 查看用户列表
   - 点击用户查看详细信息
   - 查看自定义属性
   ```

4. **页面浏览**
   ```
   Analytics > Views
   - Home
   - home
   - products
   - about
   ```

5. **错误报告**
   ```
   Crashes > Overview
   - 查看JavaScript错误
   - 查看错误堆栈
   ```

## 🎯 下一步

### 开发自己的应用

1. **复制插件代码**
   ```bash
   # 复制 src/plugins/countly.js 到你的项目
   ```

2. **在main.js中使用**
   ```javascript
   import countlyPlugin from './plugins/countly';
   app.use(countlyPlugin, { /* 配置 */ });
   ```

3. **在组件中使用**
   ```javascript
   this.$countly.recordEvent('my_event');
   ```

### 生产部署

1. **关闭调试模式**
   ```javascript
   app.use(countlyPlugin, {
     appKey: 'YOUR_APP_KEY',
     url: 'https://your-domain.com',
     debug: false  // 关闭调试
   });
   ```

2. **构建生产版本**
   ```bash
   npm run build
   ```

3. **部署dist目录**
   ```bash
   # dist目录包含所有静态文件
   # 可以部署到任何静态服务器
   ```

## 📚 更多资源

- **详细文档**：查看 `README.md`
- **集成指南**：查看 `../Countly-Web集成测试指南.md`
- **安装指南**：查看 `../Countly安装部署完全指南.md`
- **官方文档**：https://support.count.ly

## 💡 提示

- 开发时保持 `debug: true`，可以在Console看到详细日志
- 生产环境务必设置 `debug: false`
- 定期查看Countly管理界面验证数据
- 使用浏览器开发者工具调试问题

---

**遇到问题？** 查看 `README.md` 中的"常见问题"章节，或查看浏览器Console的错误信息。

**祝你使用愉快！** 🎉
