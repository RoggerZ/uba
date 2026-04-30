# Vue Countly Demo

> Vue 3 + Countly SDK 集成示例应用

## 📋 项目简介

这是一个完整的Vue 3应用，演示了如何集成和使用Countly Web SDK。包含了所有核心功能的测试界面。

## ✨ 功能特性

- ✅ 基础事件追踪（简单事件、带参数事件、定时事件）
- ✅ 用户属性设置（姓名、邮箱、年龄、自定义属性）
- ✅ 页面浏览追踪
- ✅ 错误追踪和异常捕获
- ✅ 表单提交追踪
- ✅ 实时事件日志显示
- ✅ 响应式设计

## 🚀 快速开始

### 1. 安装依赖

```bash
npm install
```

### 2. 配置Countly

编辑 `src/main.js`，替换以下配置：

```javascript
app.use(countlyPlugin, {
  appKey: 'YOUR_APP_KEY',        // 替换为你的App Key
  url: 'http://YOUR_IP:8080',    // 替换为你的Countly服务器地址
  debug: true                     // 生产环境设置为false
});
```

**获取App Key**：
1. 登录Countly管理界面：`http://YOUR_IP:8080`
2. 进入 Management > Applications
3. 创建新应用或选择现有应用
4. 复制App Key

### 3. 运行开发服务器

```bash
npm run dev
```

应用将在 `http://localhost:5173` 启动

### 4. 构建生产版本

```bash
npm run build
```

构建产物将生成在 `dist` 目录

### 5. 预览生产版本

```bash
npm run preview
```

## 📁 项目结构

```
vue-countly-demo/
├── src/
│   ├── App.vue              # 主应用组件
│   ├── main.js              # 应用入口
│   └── plugins/
│       └── countly.js       # Countly插件
├── index.html               # HTML入口
├── vite.config.js           # Vite配置
├── package.json             # 项目配置
└── README.md                # 项目说明
```

## 🎯 功能说明

### 1. 基础事件

- **发送简单事件**：记录按钮点击等简单操作
- **发送带参数事件**：记录带有详细信息的事件
- **定时事件**：测量操作持续时间

### 2. 用户属性

- **基础属性**：姓名、邮箱、年龄
- **自定义属性**：订阅类型、登录时间、应用版本等

### 3. 页面浏览

追踪用户访问的页面，包括：
- 首页
- 产品页
- 关于页

### 4. 错误追踪

- **记录错误**：手动记录错误信息
- **触发异常**：捕获并记录JavaScript异常

### 5. 表单测试

追踪表单提交事件，包括：
- 输入内容长度
- 评论内容长度
- 表单完整性

### 6. 事件日志

实时显示所有操作的日志，包括：
- 时间戳
- 事件类型
- 状态（成功/错误/信息）

## 🔧 Countly插件API

### 全局方法

在组件中通过 `this.$countly` 访问：

```javascript
// 记录事件
this.$countly.recordEvent(eventName, segmentation, count);

// 开始定时事件
this.$countly.startEvent(eventName);

// 结束定时事件
this.$countly.endEvent(eventName, segmentation);

// 设置用户属性
this.$countly.setUserDetails(userDetails);

// 设置自定义属性
this.$countly.setCustomProperty(key, value);

// 追踪页面浏览
this.$countly.trackPageView(pageName);

// 记录错误
this.$countly.logError(error);

// 添加崩溃日志
this.$countly.addCrashLog(log);
```

## 🧪 测试验证

### 1. 浏览器开发者工具

打开浏览器开发者工具（F12），查看：

- **Console**：查看Countly初始化和事件日志
- **Network**：查看发送到Countly服务器的请求

### 2. Countly管理界面

登录Countly管理界面验证数据：

```
URL: http://YOUR_IP:8080
```

**验证位置**：

1. **实时数据**：Dashboard > Real-time
2. **事件数据**：Analytics > Events
3. **用户数据**：Analytics > Users
4. **页面浏览**：Analytics > Views
5. **错误报告**：Crashes > Overview

### 3. 测试清单

```
✅ 页面加载
   - 打开应用
   - 查看Console日志
   - 确认SDK初始化

✅ 简单事件
   - 点击"发送简单事件"按钮
   - 查看事件日志
   - 在Countly验证

✅ 带参数事件
   - 点击"发送带参数事件"按钮
   - 验证事件参数

✅ 定时事件
   - 点击"开始定时事件"
   - 等待几秒
   - 点击"结束定时事件"
   - 验证持续时间

✅ 用户属性
   - 填写姓名、邮箱、年龄
   - 点击"设置用户属性"
   - 在Countly查看用户信息

✅ 自定义属性
   - 点击"设置自定义属性"
   - 在Countly查看自定义字段

✅ 页面浏览
   - 点击页面浏览按钮
   - 验证浏览记录

✅ 错误追踪
   - 点击"记录错误"和"触发异常"
   - 在Countly查看错误报告

✅ 表单提交
   - 填写表单内容
   - 点击"提交表单"
   - 验证表单事件
```

## 🐛 常见问题

### 1. CORS跨域错误

**问题**：浏览器报CORS错误

**解决方案**：

在Countly服务器的Nginx配置中添加CORS头：

```nginx
location / {
    add_header 'Access-Control-Allow-Origin' '*';
    add_header 'Access-Control-Allow-Methods' 'GET, POST, OPTIONS';
    add_header 'Access-Control-Allow-Headers' 'Content-Type';
}
```

### 2. 事件未发送

**问题**：点击按钮后事件未发送

**排查步骤**：

1. 检查Console是否有错误
2. 检查Network标签是否有请求
3. 确认App Key和服务器地址正确
4. 确认Countly服务器正常运行

### 3. 开发服务器端口冲突

**问题**：5173端口已被占用

**解决方案**：

修改 `vite.config.js`：

```javascript
export default defineConfig({
  server: {
    port: 3000  // 使用其他端口
  }
});
```

### 4. 构建失败

**问题**：npm run build失败

**解决方案**：

```bash
# 清除缓存
rm -rf node_modules
rm package-lock.json

# 重新安装
npm install

# 重新构建
npm run build
```

## 📚 参考资源

### 官方文档

- **Countly Web SDK**：https://support.count.ly/hc/en-us/articles/360037441932
- **Vue 3文档**：https://vuejs.org/
- **Vite文档**：https://vitejs.dev/

### 相关指南

- **Countly安装部署指南**：`../Countly安装部署完全指南.md`
- **Countly Web集成指南**：`../Countly-Web集成测试指南.md`

## 🎨 自定义开发

### 添加新功能

1. 在 `App.vue` 中添加新的方法
2. 使用 `this.$countly` 调用Countly API
3. 使用 `this.addLog()` 记录日志

### 修改样式

编辑 `App.vue` 中的 `<style>` 部分

### 扩展插件

编辑 `src/plugins/countly.js` 添加新的方法

## 📝 技术栈

- **Vue 3**：渐进式JavaScript框架
- **Vite**：下一代前端构建工具
- **Countly SDK**：Web分析SDK

## 📄 许可证

MIT License

## 👨‍💻 作者

Kiro AI

## 🤝 贡献

欢迎提交Issue和Pull Request！

---

**祝你使用愉快！** 🎉
