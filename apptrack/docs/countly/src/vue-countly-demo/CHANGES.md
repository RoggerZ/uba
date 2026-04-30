# 更新日志 - 对齐官方向导

> 根据Countly官方向导进行的配置调整

## 📅 更新时间

2026年1月29日

## 🔄 主要变更

### 1. 新增追踪功能

根据官方向导，在 `src/plugins/countly.js` 中新增了4个追踪功能：

```javascript
// 原有功能
Countly.track_sessions();
Countly.track_pageview();
Countly.track_clicks();
Countly.track_errors();

// ✨ 新增功能
Countly.track_scrolls();      // 滚动深度追踪
Countly.track_links();        // 链接点击追踪
Countly.track_forms();        // 表单交互追踪
Countly.collect_from_forms(); // 表单数据收集
```

### 2. 添加noscript支持

在 `index.html` 中添加了 `<noscript>` 标签：

```html
<noscript>
  <img src='http://localhost:8380/pixel.png?app_key=2963dcc87fe3f000725a16c9ada61f707ca546dd&begin_session=1'
       alt=''
       style='position:absolute; left:-9999px;'/>
</noscript>
```

**作用**：追踪禁用JavaScript的用户

### 3. 更新配置信息

在 `src/main.js` 中使用实际的配置：

```javascript
// 之前（占位符）
appKey: 'YOUR_APP_KEY',
url: 'http://YOUR_IP:8080',

// 现在（实际配置）
appKey: '2963dcc87fe3f000725a16c9ada61f707ca546dd',
url: 'http://localhost:8380',
```

## 📊 功能对比

| 功能 | 之前 | 现在 | 说明 |
|------|------|------|------|
| 会话追踪 | ✅ | ✅ | 无变化 |
| 页面浏览 | ✅ | ✅ | 无变化 |
| 点击追踪 | ✅ | ✅ | 无变化 |
| 错误追踪 | ✅ | ✅ | 无变化 |
| **滚动追踪** | ❌ | ✅ | **新增** |
| **链接追踪** | ❌ | ✅ | **新增** |
| **表单追踪** | ❌ | ✅ | **新增** |
| **表单数据收集** | ❌ | ✅ | **新增** |
| **noscript支持** | ❌ | ✅ | **新增** |

## 🎯 新功能详解

### track_scrolls - 滚动深度追踪

**数据示例**：
```json
{
  "key": "[CLY]_scroll",
  "segmentation": {
    "depth": "75%",
    "page": "/home"
  }
}
```

**应用场景**：
- 内容优化：了解用户阅读深度
- 广告位优化：确定最佳广告位置
- 用户体验：识别用户感兴趣的内容

### track_links - 链接点击追踪

**数据示例**：
```json
{
  "key": "[CLY]_link_click",
  "segmentation": {
    "url": "https://example.com",
    "text": "了解更多"
  }
}
```

**应用场景**：
- 导航优化：了解用户点击路径
- 外链分析：追踪外部链接效果
- 转化分析：优化CTA按钮

### track_forms - 表单交互追踪

**数据示例**：
```json
{
  "key": "[CLY]_form_view",
  "key": "[CLY]_form_start",
  "key": "[CLY]_form_submit",
  "key": "[CLY]_form_abandon"
}
```

**应用场景**：
- 转化优化：降低表单放弃率
- 用户体验：识别问题字段
- A/B测试：测试不同表单设计

### collect_from_forms - 表单数据收集

**数据示例**：
```json
{
  "key": "[CLY]_form_field",
  "segmentation": {
    "field": "email",
    "type": "email",
    "time": 5.2
  }
}
```

**应用场景**：
- 字段优化：识别填写困难的字段
- 时间分析：了解用户填写时间
- 错误分析：追踪验证错误

### noscript支持

**工作原理**：
1. 用户禁用JavaScript
2. 浏览器加载 `<noscript>` 中的图片
3. 图片请求发送到Countly服务器
4. 服务器记录会话开始

**数据收集**：
- 基本会话信息
- 页面访问
- 用户代理

**局限性**：
- 无法追踪事件
- 无法追踪用户交互
- 只能记录页面访问

## 🔧 技术细节

### SDK加载方式对比

**NPM方式（当前使用）**：
```javascript
import Countly from 'countly-sdk-web';
Countly.init({ /* config */ });
```

**官方向导方式（异步加载）**：
```javascript
var Countly = Countly || {};
Countly.q = Countly.q || [];
// 配置
(function() {
  var cly = document.createElement('script');
  cly.src = 'http://localhost:8380/sdk/web/countly.min.js';
  cly.onload = function(){ Countly.init() };
  // 插入脚本
})();
```

**两种方式的区别**：

| 特性 | NPM方式 | 异步加载方式 |
|------|---------|-------------|
| 安装 | 需要npm install | 无需安装 |
| 版本控制 | package.json | 服务器控制 |
| 离线开发 | ✅ 支持 | ❌ 需要服务器 |
| 构建优化 | ✅ Tree-shaking | ❌ 完整SDK |
| TypeScript | ✅ 类型支持 | ❌ 无类型 |
| 更新 | 手动更新 | 自动更新 |

**当前选择NPM方式的原因**：
- ✅ 更好的开发体验
- ✅ 版本稳定可控
- ✅ 支持现代构建工具
- ✅ TypeScript友好

## 📝 配置文件变更

### src/plugins/countly.js

**变更行数**：4行新增

```diff
  Countly.track_sessions();
  Countly.track_pageview();
  Countly.track_clicks();
+ Countly.track_scrolls();
  Countly.track_errors();
+ Countly.track_links();
+ Countly.track_forms();
+ Countly.collect_from_forms();
```

### index.html

**变更行数**：4行新增

```diff
  <body>
+   <!-- Fallback tracking for users with disabled JavaScript -->
+   <noscript>
+     <img src='http://localhost:8380/pixel.png?app_key=2963dcc87fe3f000725a16c9ada61f707ca546dd&begin_session=1' alt='' style='position:absolute; left:-9999px;'/>
+   </noscript>
    <div id="app"></div>
```

### src/main.js

**变更行数**：2行修改

```diff
  app.use(countlyPlugin, {
-   appKey: 'YOUR_APP_KEY',
+   appKey: '2963dcc87fe3f000725a16c9ada61f707ca546dd',
-   url: 'http://YOUR_IP:8080',
+   url: 'http://localhost:8380',
    debug: true
  });
```

## ✅ 验证清单

### 功能验证

```
✅ 会话追踪正常
✅ 页面浏览正常
✅ 点击追踪正常
✅ 错误追踪正常
✅ 滚动追踪正常（新增）
✅ 链接追踪正常（新增）
✅ 表单追踪正常（新增）
✅ 表单数据收集正常（新增）
✅ noscript支持正常（新增）
```

### 配置验证

```
✅ App Key已更新
✅ 服务器URL已更新（端口8380）
✅ Debug模式已启用
✅ noscript标签已添加
✅ 所有追踪功能已启用
```

## 🎓 使用建议

### 开发阶段

1. **保持debug模式**：便于调试
2. **测试所有功能**：确保追踪正常
3. **查看Console日志**：了解SDK行为

### 生产部署

1. **关闭debug模式**：
   ```javascript
   debug: false
   ```

2. **使用HTTPS**：
   ```javascript
   url: 'https://your-domain.com'
   ```

3. **更新noscript URL**：
   ```html
   <img src='https://your-domain.com/pixel.png?...' />
   ```

### 隐私合规

如需GDPR合规，添加用户同意：

```javascript
Countly.require_consent = true;
Countly.add_consent(['sessions', 'events', 'views', 'scrolls', 'clicks', 'forms']);
```

## 📚 新增文档

- ✅ `CONFIGURATION.md` - 详细配置说明
- ✅ `CHANGES.md` - 本文档，变更日志

## 🔗 参考资源

- **官方向导**：Countly管理界面 > Management > Applications > Your App > Setup
- **Web SDK文档**：https://support.count.ly/hc/en-us/articles/360037441932
- **GitHub仓库**：https://github.com/Countly/countly-sdk-web

## 🎉 总结

所有配置已完全对齐Countly官方向导，新增了4个重要的追踪功能和noscript支持，提供了更全面的用户行为分析能力。

**主要改进**：
- ✅ 功能更完整（8个追踪功能）
- ✅ 覆盖更全面（包括禁用JS的用户）
- ✅ 配置更准确（使用实际配置）
- ✅ 文档更详细（新增配置说明）

---

**版本**：v1.1.0
**更新日期**：2026年1月29日
**作者**：Kiro AI
