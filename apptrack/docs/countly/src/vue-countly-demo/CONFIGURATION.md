# 配置说明

> 根据Countly官方向导调整的配置说明

## 🔧 已应用的官方配置

### 1. 启用的功能（与官方向导一致）

```javascript
Countly.track_sessions();        // 会话追踪
Countly.track_pageview();        // 页面浏览追踪
Countly.track_clicks();          // 点击追踪
Countly.track_scrolls();         // ✨ 滚动深度追踪（新增）
Countly.track_errors();          // 错误追踪
Countly.track_links();           // ✨ 链接点击追踪（新增）
Countly.track_forms();           // ✨ 表单交互追踪（新增）
Countly.collect_from_forms();    // ✨ 表单数据收集（新增）
```

### 2. noscript支持

在 `index.html` 中添加了 `<noscript>` 标签，用于追踪禁用JavaScript的用户：

```html
<noscript>
  <img src='http://localhost:8380/pixel.png?app_key=YOUR_APP_KEY&begin_session=1'
       alt=''
       style='position:absolute; left:-9999px;'/>
</noscript>
```

**工作原理**：
- 当用户禁用JavaScript时，会加载这个1x1像素的透明图片
- 图片请求会发送到Countly服务器，记录会话开始
- 这样即使JavaScript不可用，也能追踪基本的访问数据

### 3. 当前配置

**文件**：`src/main.js`

```javascript
app.use(countlyPlugin, {
  appKey: '2963dcc87fe3f000725a16c9ada61f707ca546dd',
  url: 'http://localhost:8380',
  debug: true
});
```

**说明**：
- `appKey`：已设置为你的实际App Key
- `url`：已设置为你的实际服务器地址（端口8380）
- `debug`：开发模式启用，生产环境需改为false

## 📊 新增功能说明

### track_scrolls - 滚动深度追踪

**功能**：追踪用户在页面上的滚动行为

**数据收集**：
- 滚动深度百分比（25%, 50%, 75%, 100%）
- 页面停留时间
- 滚动速度

**用途**：
- 了解用户阅读习惯
- 优化内容布局
- 识别用户感兴趣的内容区域

### track_links - 链接点击追踪

**功能**：自动追踪所有外部链接的点击

**数据收集**：
- 链接URL
- 链接文本
- 点击位置

**用途**：
- 了解用户导航路径
- 追踪外部链接点击率
- 优化链接布局

### track_forms - 表单交互追踪

**功能**：追踪表单的交互行为

**数据收集**：
- 表单查看次数
- 表单开始填写次数
- 表单提交次数
- 表单放弃率

**用途**：
- 优化表单设计
- 降低表单放弃率
- 提高转化率

### collect_from_forms - 表单数据收集

**功能**：收集表单字段的详细数据

**数据收集**：
- 字段名称
- 字段类型
- 填写时间
- 错误信息

**用途**：
- 识别问题字段
- 优化表单流程
- 提升用户体验

**⚠️ 隐私提示**：
- 不会收集密码字段
- 不会收集敏感信息（如信用卡号）
- 遵守GDPR等隐私法规

## 🔄 与CDN版本的对比

### NPM版本（当前使用）

**优势**：
- ✅ 版本固定，稳定可控
- ✅ 离线开发支持
- ✅ 构建优化
- ✅ TypeScript支持

**配置方式**：
```javascript
import Countly from 'countly-sdk-web';
Countly.init({ /* config */ });
```

### CDN版本（官方向导）

**优势**：
- ✅ 无需安装依赖
- ✅ 自动更新
- ✅ CDN加速

**配置方式**：
```html
<script>
var Countly = Countly || {};
Countly.q = Countly.q || [];
// ... 配置
</script>
<script src="http://localhost:8380/sdk/web/countly.min.js"></script>
```

### 自托管版本（官方推荐）

**优势**：
- ✅ 完全控制
- ✅ 无外部依赖
- ✅ 隐私保护

**SDK地址**：
```
http://localhost:8380/sdk/web/countly.min.js
```

## 🎯 推荐配置

### 开发环境

```javascript
app.use(countlyPlugin, {
  appKey: '2963dcc87fe3f000725a16c9ada61f707ca546dd',
  url: 'http://localhost:8380',
  debug: true  // 启用调试
});
```

### 生产环境

```javascript
app.use(countlyPlugin, {
  appKey: '2963dcc87fe3f000725a16c9ada61f707ca546dd',
  url: 'https://your-domain.com',  // 使用HTTPS
  debug: false  // 关闭调试
});
```

## 📝 配置检查清单

```
✅ App Key已设置（从Countly管理界面获取）
✅ 服务器URL已设置（http://localhost:8380）
✅ 启用了所有官方推荐的追踪功能
✅ 添加了noscript支持
✅ 开发环境启用debug模式
```

## 🔐 隐私和安全

### 数据收集范围

**自动收集**：
- 页面URL（不含敏感参数）
- 用户代理信息
- 屏幕分辨率
- 会话时长

**不收集**：
- 密码字段
- 信用卡信息
- 个人身份信息（除非明确设置）

### GDPR合规

如需GDPR合规，添加用户同意管理：

```javascript
// 需要用户同意
Countly.require_consent = true;

// 用户同意后
Countly.add_consent(['sessions', 'events', 'views', 'scrolls', 'clicks', 'forms']);

// 撤销同意
Countly.remove_consent(['forms']);
```

## 🧪 测试验证

### 1. 验证滚动追踪

1. 打开应用
2. 慢慢向下滚动页面
3. 在Countly管理界面查看：`Analytics > User Behavior > Scrolls`

### 2. 验证链接追踪

1. 添加外部链接到页面
2. 点击链接
3. 在Countly查看：`Analytics > Events > [CLY]_link_click`

### 3. 验证表单追踪

1. 与表单交互（聚焦、填写、提交）
2. 在Countly查看：`Analytics > Events > [CLY]_form_*`

### 4. 验证noscript

1. 禁用浏览器JavaScript
2. 访问页面
3. 在Countly查看会话数据

## 🔧 故障排除

### 滚动追踪不工作

**检查**：
- 页面是否有足够的内容可滚动
- Console是否有错误

### 表单追踪不工作

**检查**：
- 表单是否使用标准HTML `<form>` 标签
- 是否有JavaScript错误阻止追踪

### noscript不工作

**检查**：
- 图片URL是否正确
- Countly服务器是否可访问
- 网络请求是否被拦截

## 📚 参考资源

- **官方文档**：https://support.count.ly/hc/en-us/articles/360037441932
- **SDK GitHub**：https://github.com/Countly/countly-sdk-web
- **API参考**：https://countly.github.io/countly-sdk-web/

---

**配置已完全对齐官方向导！** ✅
