# Countly 集成示例项目

> 包含Web、Android、Flutter等平台的Countly SDK集成示例

## 📁 项目结构

```
src/
└── vue-countly-demo/          # Vue 3 + Countly SDK 示例项目
    ├── src/
    │   ├── App.vue            # 主应用组件（完整功能演示）
    │   ├── main.js            # 应用入口
    │   └── plugins/
    │       └── countly.js     # Countly Vue插件
    ├── index.html             # HTML入口
    ├── vite.config.js         # Vite配置
    ├── package.json           # 项目依赖
    ├── README.md              # 详细文档
    ├── QUICKSTART.md          # 快速启动指南
    └── .gitignore             # Git忽略文件
```

## 🚀 快速开始

### Vue项目

```bash
# 进入项目目录
cd vue-countly-demo

# 安装依赖
npm install

# 配置Countly（编辑 src/main.js）
# 替换 YOUR_APP_KEY 和 YOUR_IP:8080

# 运行开发服务器
npm run dev

# 访问 http://localhost:5173
```

**详细说明**：查看 `vue-countly-demo/QUICKSTART.md`

## 📚 相关文档

### 集成指南

- **Web集成**：`../Countly-Web集成测试指南.md`
- **Android集成**：`../Countly-Android集成测试指南.md`
- **Flutter集成**：`../Countly-Flutter集成测试指南.md`

### 部署指南

- **Countly安装部署**：`../Countly安装部署完全指南.md`
- **端口配置**：`../../../../countly-server/PORT_CONFIGURATION.md`

## ✨ 功能特性

### Vue Demo应用

- ✅ 基础事件追踪（简单事件、带参数事件、定时事件）
- ✅ 用户属性设置（基础属性、自定义属性）
- ✅ 页面浏览追踪
- ✅ 错误追踪和异常捕获
- ✅ 表单提交追踪
- ✅ 实时事件日志显示
- ✅ 响应式设计
- ✅ 完整的测试界面

## 🎯 项目说明

### Vue项目特点

**技术栈**：
- Vue 3（Composition API）
- Vite（快速构建）
- Countly Web SDK

**适用场景**：
- Web应用集成Countly
- 学习Countly SDK使用
- 快速验证Countly功能
- 作为项目模板使用

**优势**：
- 开箱即用
- 代码简洁
- 功能完整
- 易于扩展

## 🧪 测试验证

### 1. 本地测试

```bash
# 运行Vue项目
cd vue-countly-demo
npm run dev
```

### 2. 功能测试

在浏览器中测试所有功能：
- 基础事件
- 用户属性
- 页面浏览
- 错误追踪
- 表单提交

### 3. 数据验证

登录Countly管理界面验证数据：
```
URL: http://YOUR_IP:8080
Dashboard > Real-time
Analytics > Events
Analytics > Users
```

## 📖 使用指南

### 快速集成到现有项目

1. **复制Countly插件**
   ```bash
   cp vue-countly-demo/src/plugins/countly.js your-project/src/plugins/
   ```

2. **在main.js中使用**
   ```javascript
   import countlyPlugin from './plugins/countly';

   app.use(countlyPlugin, {
     appKey: 'YOUR_APP_KEY',
     url: 'http://YOUR_IP:8080',
     debug: true
   });
   ```

3. **在组件中使用**
   ```javascript
   // 记录事件
   this.$countly.recordEvent('button_clicked');

   // 设置用户属性
   this.$countly.setUserDetails({ name: 'John' });

   // 追踪页面浏览
   this.$countly.trackPageView('Home');
   ```

### 自定义开发

参考 `vue-countly-demo/src/App.vue` 中的示例代码，根据需求修改和扩展。

## 🔧 配置说明

### Countly服务器配置

**Docker部署**（推荐）：
```bash
# 在countly-server目录
docker-compose up -d

# 访问管理界面
http://YOUR_IP:8080
```

**端口配置**：
- 默认端口：8080（已修改，避免与系统Nginx冲突）
- 可以使用系统Nginx反向代理
- 详见：`PORT_CONFIGURATION.md`

### 应用配置

**必需配置**：
- `appKey`：从Countly管理界面获取
- `url`：Countly服务器地址

**可选配置**：
- `debug`：调试模式（开发环境true，生产环境false）

## 🐛 常见问题

### 1. CORS跨域错误

**解决方案**：配置Countly服务器的Nginx添加CORS头

### 2. 事件未发送

**排查步骤**：
1. 检查App Key和服务器地址
2. 检查Countly服务器是否运行
3. 查看浏览器Console和Network

### 3. 依赖安装失败

**解决方案**：
```bash
# 使用国内镜像
npm config set registry https://registry.npmmirror.com
npm install
```

更多问题查看各项目的README.md文档。

## 📊 项目状态

| 项目 | 状态 | 说明 |
|------|------|------|
| Vue Demo | ✅ 完成 | 完整的功能演示应用 |
| Android Demo | 📝 计划中 | 参考Android集成指南 |
| Flutter Demo | 📝 计划中 | 参考Flutter集成指南 |

## 🎓 学习路径

### 新手入门

1. **阅读安装部署指南**
   - 了解Countly架构
   - 部署Countly服务器

2. **运行Vue Demo**
   - 快速体验Countly功能
   - 理解SDK使用方法

3. **查看集成指南**
   - 学习各平台集成方法
   - 了解最佳实践

### 进阶开发

1. **自定义事件**
   - 设计事件命名规范
   - 定义事件参数结构

2. **用户分析**
   - 设置用户属性
   - 用户分群策略

3. **数据分析**
   - 使用Countly管理界面
   - 创建自定义报表

## 🔗 相关链接

### 官方资源

- **Countly官网**：https://countly.com
- **官方文档**：https://support.count.ly
- **GitHub**：https://github.com/Countly

### 项目文档

- **功能分析**：`../Countly功能深度分析.md`
- **产品方向**：`../../产品方向决策-移动分析市场定位.md`
- **技术方案**：`../../技术实现方案-架构与开发指南.md`

## 📝 更新日志

### v1.0.0 (2026-01-29)

- ✅ 创建Vue 3 + Countly SDK示例项目
- ✅ 实现所有核心功能演示
- ✅ 添加完整的文档和指南
- ✅ 提供快速启动指南

## 👨‍💻 贡献

欢迎提交Issue和Pull Request！

## 📄 许可证

MIT License

---

**开始你的Countly集成之旅！** 🚀
