# Countly Web 集成测试指南

> 完整的Web应用集成Countly JavaScript SDK教程
> 日期：2026年1月28日

---

## 📋 目录

1. [项目概述](#项目概述)
2. [开发环境准备](#开发环境准备)
3. [快速开始](#快速开始)
4. [集成方式](#集成方式)
5. [核心功能实现](#核心功能实现)
6. [React集成](#react集成)
7. [Vue集成](#vue集成)
8. [测试验证](#测试验证)
9. [常见问题](#常见问题)

---

## 🎯 项目概述

### Demo应用功能

创建一个Web应用来测试Countly的核心功能：

**功能列表**：
```
1. 基础追踪
   - 页面浏览
   - 会话追踪
   - 用户追踪

2. 自定义事件
   - 按钮点击事件
   - 表单提交事件
   - 带参数的事件

3. 用户属性
   - 设置用户信息
   - 自定义属性

4. 错误追踪
   - JavaScript错误捕获
   - 自定义错误记录

5. 性能监控
   - 页面加载时间
   - AJAX请求监控

6. 热力图（可选）
   - 点击热力图
   - 滚动热力图
```

### 技术栈

```
SDK：Countly Web SDK (最新版本)
支持浏览器：Chrome, Firefox, Safari, Edge
框架支持：原生JS, React, Vue, Angular
最低要求：支持ES5的现代浏览器
```

---

## 💻 开发环境准备

### 1. 基础工具

```bash
# Node.js（可选，用于本地开发服务器）
node --version  # 推荐 v18+

# 简单HTTP服务器（可选）
npm install -g http-server
# 或
python -m http.server 8000
```

### 2. 代码编辑器

推荐使用：
- VS Code
- WebStorm
- Sublime Text

### 3. 准备Countly服务器信息

```
服务器地址：http://YOUR_IP:8080
App Key：在Countly管理界面获取
```

**获取App Key**：
1. 登录Countly管理界面
2. Management > Applications
3. 创建新应用或选择现有应用
4. 复制App Key

---

## 🚀 快速开始

### 最简单的集成（5分钟）

创建`index.html`：

```html
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Countly Web Demo</title>
</head>
<body>
    <h1>Countly Web 测试</h1>
    <button onclick="sendEvent()">发送事件</button>

    <!-- Countly SDK -->
    <script type="text/javascript">
        // Countly配置
        var Countly = Countly || {};
        Countly.q = Countly.q || [];

        // 初始化
        Countly.app_key = 'YOUR_APP_KEY';  // 替换为你的App Key
        Countly.url = 'http://YOUR_IP:8080';  // 替换为你的服务器地址

        // 启用功能
        Countly.q.push(['track_sessions']);
        Countly.q.push(['track_pageview']);
        Countly.q.push(['track_clicks']);
        Countly.q.push(['track_errors']);

        // 加载SDK
        (function() {
            var cly = document.createElement('script');
            cly.type = 'text/javascript';
            cly.async = true;
            cly.src = 'https://cdn.jsdelivr.net/npm/countly-sdk-web@latest/lib/countly.min.js';
            cly.onload = function() {
                Countly.init();
            };
            var s = document.getElementsByTagName('script')[0];
            s.parentNode.insertBefore(cly, s);
        })();

        // 发送自定义事件
        function sendEvent() {
            Countly.q.push(['add_event', {
                key: 'button_clicked',
                count: 1,
                segmentation: {
                    button_name: 'test_button'
                }
            }]);
            alert('事件已发送！');
        }
    </script>
</body>
</html>
```

**运行**：
```bash
# 使用http-server
http-server -p 8000

# 或使用Python
python -m http.server 8000

# 访问
http://localhost:8000
```



---

## 🔧 集成方式

### 方式1：CDN引入（推荐）

最简单的方式，直接从CDN加载：

```html
<script type="text/javascript">
    var Countly = Countly || {};
    Countly.q = Countly.q || [];

    // 配置
    Countly.app_key = 'YOUR_APP_KEY';
    Countly.url = 'http://YOUR_IP:8080';

    // 功能配置
    Countly.q.push(['track_sessions']);
    Countly.q.push(['track_pageview']);
    Countly.q.push(['track_clicks']);
    Countly.q.push(['track_errors']);

    // 加载SDK
    (function() {
        var cly = document.createElement('script');
        cly.type = 'text/javascript';
        cly.async = true;
        cly.src = 'https://cdn.jsdelivr.net/npm/countly-sdk-web@latest/lib/countly.min.js';
        cly.onload = function() { Countly.init(); };
        var s = document.getElementsByTagName('script')[0];
        s.parentNode.insertBefore(cly, s);
    })();
</script>
```

### 方式2：NPM安装

适合使用构建工具的项目：

```bash
# 安装
npm install countly-sdk-web --save
```

```javascript
// 引入
import Countly from 'countly-sdk-web';

// 初始化
Countly.init({
    app_key: 'YOUR_APP_KEY',
    url: 'http://YOUR_IP:8080',
    debug: true
});

// 启用功能
Countly.track_sessions();
Countly.track_pageview();
Countly.track_clicks();
Countly.track_errors();
```

### 方式3：自托管SDK

从Countly服务器加载SDK：

```html
<script type="text/javascript">
    var Countly = Countly || {};
    Countly.q = Countly.q || [];

    Countly.app_key = 'YOUR_APP_KEY';
    Countly.url = 'http://YOUR_IP:8080';

    Countly.q.push(['track_sessions']);
    Countly.q.push(['track_pageview']);

    (function() {
        var cly = document.createElement('script');
        cly.type = 'text/javascript';
        cly.async = true;
        // 从Countly服务器加载
        cly.src = 'http://YOUR_IP:8080/sdk/web/countly.min.js';
        cly.onload = function() { Countly.init(); };
        var s = document.getElementsByTagName('script')[0];
        s.parentNode.insertBefore(cly, s);
    })();
</script>
```

---

## 🎨 核心功能实现

### 完整Demo页面

创建`demo.html`：

```html
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Countly Web 完整Demo</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }

        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            padding: 20px;
        }

        .container {
            max-width: 800px;
            margin: 0 auto;
            background: white;
            border-radius: 20px;
            padding: 40px;
            box-shadow: 0 20px 60px rgba(0,0,0,0.3);
        }

        h1 {
            color: #333;
            margin-bottom: 10px;
            font-size: 32px;
        }

        .subtitle {
            color: #666;
            margin-bottom: 30px;
            font-size: 16px;
        }

        .section {
            margin-bottom: 30px;
            padding: 20px;
            background: #f8f9fa;
            border-radius: 10px;
        }

        .section h2 {
            color: #667eea;
            margin-bottom: 15px;
            font-size: 20px;
        }

        button {
            background: #667eea;
            color: white;
            border: none;
            padding: 12px 24px;
            border-radius: 8px;
            cursor: pointer;
            font-size: 14px;
            margin: 5px;
            transition: all 0.3s;
        }

        button:hover {
            background: #5568d3;
            transform: translateY(-2px);
            box-shadow: 0 5px 15px rgba(102, 126, 234, 0.4);
        }

        button.danger {
            background: #e74c3c;
        }

        button.danger:hover {
            background: #c0392b;
        }

        input, textarea {
            width: 100%;
            padding: 10px;
            margin: 10px 0;
            border: 2px solid #e0e0e0;
            border-radius: 8px;
            font-size: 14px;
        }

        input:focus, textarea:focus {
            outline: none;
            border-color: #667eea;
        }

        .log {
            background: #2c3e50;
            color: #2ecc71;
            padding: 15px;
            border-radius: 8px;
            font-family: 'Courier New', monospace;
            font-size: 12px;
            max-height: 200px;
            overflow-y: auto;
            margin-top: 15px;
        }

        .log-entry {
            margin: 5px 0;
        }

        .status {
            display: inline-block;
            padding: 5px 10px;
            border-radius: 5px;
            font-size: 12px;
            margin-left: 10px;
        }

        .status.success {
            background: #2ecc71;
            color: white;
        }

        .status.error {
            background: #e74c3c;
            color: white;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🚀 Countly Web Demo</h1>
        <p class="subtitle">完整功能测试页面</p>

        <!-- 基础事件 -->
        <div class="section">
            <h2>📊 基础事件</h2>
            <button onclick="sendSimpleEvent()">发送简单事件</button>
            <button onclick="sendEventWithParams()">发送带参数事件</button>
            <button onclick="sendTimedEvent()">发送定时事件</button>
        </div>

        <!-- 用户属性 -->
        <div class="section">
            <h2>👤 用户属性</h2>
            <input type="text" id="userName" placeholder="姓名">
            <input type="email" id="userEmail" placeholder="邮箱">
            <input type="number" id="userAge" placeholder="年龄">
            <button onclick="setUserProfile()">设置用户属性</button>
            <button onclick="setCustomProperty()">设置自定义属性</button>
        </div>

        <!-- 页面浏览 -->
        <div class="section">
            <h2>📄 页面浏览</h2>
            <button onclick="trackPageView('home')">追踪首页</button>
            <button onclick="trackPageView('products')">追踪产品页</button>
            <button onclick="trackPageView('about')">追踪关于页</button>
        </div>

        <!-- 错误追踪 -->
        <div class="section">
            <h2>🐛 错误追踪</h2>
            <button onclick="logError()" class="danger">记录错误</button>
            <button onclick="throwError()" class="danger">触发异常</button>
        </div>

        <!-- 表单测试 -->
        <div class="section">
            <h2>📝 表单测试</h2>
            <input type="text" id="formInput" placeholder="输入内容">
            <textarea id="formTextarea" placeholder="输入评论" rows="3"></textarea>
            <button onclick="submitForm()">提交表单</button>
        </div>

        <!-- 日志 -->
        <div class="section">
            <h2>📋 事件日志</h2>
            <div id="log" class="log">
                <div class="log-entry">[系统] Countly初始化中...</div>
            </div>
        </div>
    </div>

    <!-- Countly SDK -->
    <script type="text/javascript">
        // Countly配置
        var Countly = Countly || {};
        Countly.q = Countly.q || [];

        // 基础配置
        Countly.app_key = 'YOUR_APP_KEY';  // 替换为你的App Key
        Countly.url = 'http://YOUR_IP:8080';  // 替换为你的服务器地址
        Countly.debug = true;  // 开发环境启用调试

        // 启用功能
        Countly.q.push(['track_sessions']);
        Countly.q.push(['track_pageview']);
        Countly.q.push(['track_clicks']);
        Countly.q.push(['track_errors']);
        Countly.q.push(['track_links']);
        Countly.q.push(['track_forms']);

        // 加载SDK
        (function() {
            var cly = document.createElement('script');
            cly.type = 'text/javascript';
            cly.async = true;
            cly.src = 'https://cdn.jsdelivr.net/npm/countly-sdk-web@latest/lib/countly.min.js';
            cly.onload = function() {
                Countly.init();
                addLog('Countly初始化成功', 'success');
            };
            var s = document.getElementsByTagName('script')[0];
            s.parentNode.insertBefore(cly, s);
        })();

        // 日志函数
        function addLog(message, type = 'info') {
            var log = document.getElementById('log');
            var time = new Date().toLocaleTimeString();
            var entry = document.createElement('div');
            entry.className = 'log-entry';
            entry.innerHTML = `[${time}] ${message}`;
            log.appendChild(entry);
            log.scrollTop = log.scrollHeight;
        }

        // 1. 发送简单事件
        function sendSimpleEvent() {
            Countly.q.push(['add_event', {
                key: 'button_clicked',
                count: 1
            }]);
            addLog('✓ 简单事件已发送: button_clicked', 'success');
        }

        // 2. 发送带参数事件
        function sendEventWithParams() {
            Countly.q.push(['add_event', {
                key: 'button_clicked_with_params',
                count: 1,
                segmentation: {
                    button_name: 'event_with_params',
                    page: 'demo',
                    timestamp: Date.now()
                }
            }]);
            addLog('✓ 带参数事件已发送', 'success');
        }

        // 3. 发送定时事件
        var timedEventStartTime;
        function sendTimedEvent() {
            if (!timedEventStartTime) {
                // 开始计时
                timedEventStartTime = Date.now();
                Countly.q.push(['start_event', 'timed_action']);
                addLog('⏱ 定时事件已开始', 'success');

                // 3秒后结束
                setTimeout(function() {
                    var duration = (Date.now() - timedEventStartTime) / 1000;
                    Countly.q.push(['end_event', {
                        key: 'timed_action',
                        segmentation: {
                            duration: duration.toFixed(2) + 's'
                        }
                    }]);
                    addLog('✓ 定时事件已结束: ' + duration.toFixed(2) + '秒', 'success');
                    timedEventStartTime = null;
                }, 3000);
            }
        }

        // 4. 设置用户属性
        function setUserProfile() {
            var name = document.getElementById('userName').value;
            var email = document.getElementById('userEmail').value;
            var age = document.getElementById('userAge').value;

            if (name || email || age) {
                var userData = {};
                if (name) userData.name = name;
                if (email) userData.email = email;
                if (age) userData.age = parseInt(age);

                Countly.q.push(['user_details', userData]);
                addLog('✓ 用户属性已设置', 'success');
            } else {
                addLog('✗ 请填写至少一个字段', 'error');
            }
        }

        // 5. 设置自定义属性
        function setCustomProperty() {
            Countly.q.push(['user_details', {
                custom: {
                    subscription: 'premium',
                    last_login: new Date().toISOString(),
                    app_version: '1.0.0'
                }
            }]);
            addLog('✓ 自定义属性已设置', 'success');
        }

        // 6. 追踪页面浏览
        function trackPageView(pageName) {
            Countly.q.push(['track_pageview', pageName]);
            addLog('✓ 页面浏览已追踪: ' + pageName, 'success');
        }

        // 7. 记录错误
        function logError() {
            Countly.q.push(['log_error', 'This is a test error']);
            addLog('✓ 错误已记录', 'success');
        }

        // 8. 触发异常
        function throwError() {
            try {
                throw new Error('Test exception for Countly');
            } catch (e) {
                Countly.q.push(['log_error', e]);
                addLog('✓ 异常已捕获并记录', 'success');
            }
        }

        // 9. 提交表单
        function submitForm() {
            var input = document.getElementById('formInput').value;
            var textarea = document.getElementById('formTextarea').value;

            Countly.q.push(['add_event', {
                key: 'form_submitted',
                count: 1,
                segmentation: {
                    has_input: input ? 'yes' : 'no',
                    has_comment: textarea ? 'yes' : 'no',
                    input_length: input.length,
                    comment_length: textarea.length
                }
            }]);

            addLog('✓ 表单提交事件已发送', 'success');

            // 清空表单
            document.getElementById('formInput').value = '';
            document.getElementById('formTextarea').value = '';
        }

        // 页面加载完成
        window.addEventListener('load', function() {
            addLog('页面加载完成', 'success');
        });

        // 页面卸载
        window.addEventListener('beforeunload', function() {
            Countly.q.push(['end_session']);
        });
    </script>
</body>
</html>
```

**使用说明**：
1. 替换`YOUR_APP_KEY`和`YOUR_IP:8080`
2. 用浏览器打开文件
3. 点击各个按钮测试功能
4. 在Countly管理界面查看数据



---

## ⚛️ React集成

### 创建React项目

```bash
# 创建项目
npx create-react-app countly-react-demo
cd countly-react-demo

# 安装Countly SDK
npm install countly-sdk-web --save
```

### 创建Countly服务

创建`src/services/countlyService.js`：

```javascript
import Countly from 'countly-sdk-web';

class CountlyService {
  constructor() {
    this.initialized = false;
  }

  init() {
    if (this.initialized) return;

    Countly.init({
      app_key: 'YOUR_APP_KEY',
      url: 'http://YOUR_IP:8080',
      debug: true
    });

    // 启用功能
    Countly.track_sessions();
    Countly.track_pageview();
    Countly.track_clicks();
    Countly.track_errors();

    this.initialized = true;
    console.log('Countly initialized');
  }

  // 记录事件
  recordEvent(eventName, segmentation = {}, count = 1) {
    Countly.add_event({
      key: eventName,
      count: count,
      segmentation: segmentation
    });
  }

  // 设置用户属性
  setUserDetails(userDetails) {
    Countly.user_details(userDetails);
  }

  // 追踪页面浏览
  trackPageView(pageName) {
    Countly.track_pageview(pageName);
  }

  // 记录错误
  logError(error) {
    Countly.log_error(error);
  }
}

export default new CountlyService();
```

### 修改App.js

```javascript
import React, { useEffect, useState } from 'react';
import countlyService from './services/countlyService';
import './App.css';

function App() {
  const [eventCount, setEventCount] = useState(0);
  const [userName, setUserName] = useState('');
  const [userEmail, setUserEmail] = useState('');

  useEffect(() => {
    // 初始化Countly
    countlyService.init();
    countlyService.trackPageView('Home');
  }, []);

  const handleSimpleEvent = () => {
    countlyService.recordEvent('button_clicked');
    setEventCount(eventCount + 1);
    alert('简单事件已发送！');
  };

  const handleEventWithParams = () => {
    countlyService.recordEvent('button_clicked_with_params', {
      button_name: 'event_with_params',
      page: 'App',
      count: eventCount
    });
    alert('带参数事件已发送！');
  };

  const handleSetUserProfile = () => {
    if (userName || userEmail) {
      countlyService.setUserDetails({
        name: userName,
        email: userEmail
      });
      alert('用户属性已设置！');
    } else {
      alert('请填写至少一个字段');
    }
  };

  const handleError = () => {
    try {
      throw new Error('Test error from React');
    } catch (error) {
      countlyService.logError(error);
      alert('错误已记录！');
    }
  };

  return (
    <div className="App">
      <header className="App-header">
        <h1>🚀 Countly React Demo</h1>
        <p>事件计数: {eventCount}</p>

        <div className="section">
          <h2>基础事件</h2>
          <button onClick={handleSimpleEvent}>发送简单事件</button>
          <button onClick={handleEventWithParams}>发送带参数事件</button>
        </div>

        <div className="section">
          <h2>用户属性</h2>
          <input
            type="text"
            placeholder="姓名"
            value={userName}
            onChange={(e) => setUserName(e.target.value)}
          />
          <input
            type="email"
            placeholder="邮箱"
            value={userEmail}
            onChange={(e) => setUserEmail(e.target.value)}
          />
          <button onClick={handleSetUserProfile}>设置用户属性</button>
        </div>

        <div className="section">
          <h2>错误追踪</h2>
          <button onClick={handleError} className="danger">
            记录错误
          </button>
        </div>
      </header>
    </div>
  );
}

export default App;
```

### 运行React应用

```bash
npm start
```

---

## 🎨 Vue集成

### 创建Vue项目

```bash
# 创建项目
npm create vue@latest countly-vue-demo
cd countly-vue-demo

# 安装依赖
npm install

# 安装Countly SDK
npm install countly-sdk-web --save
```

### 创建Countly插件

创建`src/plugins/countly.js`：

```javascript
import Countly from 'countly-sdk-web';

export default {
  install(app, options) {
    // 初始化Countly
    Countly.init({
      app_key: options.appKey || 'YOUR_APP_KEY',
      url: options.url || 'http://YOUR_IP:8080',
      debug: options.debug || true
    });

    // 启用功能
    Countly.track_sessions();
    Countly.track_pageview();
    Countly.track_clicks();
    Countly.track_errors();

    // 添加到全局属性
    app.config.globalProperties.$countly = {
      // 记录事件
      recordEvent(eventName, segmentation = {}, count = 1) {
        Countly.add_event({
          key: eventName,
          count: count,
          segmentation: segmentation
        });
      },

      // 设置用户属性
      setUserDetails(userDetails) {
        Countly.user_details(userDetails);
      },

      // 追踪页面浏览
      trackPageView(pageName) {
        Countly.track_pageview(pageName);
      },

      // 记录错误
      logError(error) {
        Countly.log_error(error);
      }
    };

    console.log('Countly plugin installed');
  }
};
```

### 修改main.js

```javascript
import { createApp } from 'vue';
import App from './App.vue';
import countlyPlugin from './plugins/countly';

const app = createApp(App);

// 使用Countly插件
app.use(countlyPlugin, {
  appKey: 'YOUR_APP_KEY',
  url: 'http://YOUR_IP:8080',
  debug: true
});

app.mount('#app');
```

### 修改App.vue

```vue
<template>
  <div id="app">
    <header>
      <h1>🚀 Countly Vue Demo</h1>
      <p>事件计数: {{ eventCount }}</p>

      <div class="section">
        <h2>基础事件</h2>
        <button @click="sendSimpleEvent">发送简单事件</button>
        <button @click="sendEventWithParams">发送带参数事件</button>
      </div>

      <div class="section">
        <h2>用户属性</h2>
        <input v-model="userName" type="text" placeholder="姓名" />
        <input v-model="userEmail" type="email" placeholder="邮箱" />
        <button @click="setUserProfile">设置用户属性</button>
      </div>

      <div class="section">
        <h2>错误追踪</h2>
        <button @click="logError" class="danger">记录错误</button>
      </div>
    </header>
  </div>
</template>

<script>
export default {
  name: 'App',
  data() {
    return {
      eventCount: 0,
      userName: '',
      userEmail: ''
    };
  },
  mounted() {
    // 追踪页面浏览
    this.$countly.trackPageView('Home');
  },
  methods: {
    sendSimpleEvent() {
      this.$countly.recordEvent('button_clicked');
      this.eventCount++;
      alert('简单事件已发送！');
    },
    sendEventWithParams() {
      this.$countly.recordEvent('button_clicked_with_params', {
        button_name: 'event_with_params',
        page: 'App',
        count: this.eventCount
      });
      alert('带参数事件已发送！');
    },
    setUserProfile() {
      if (this.userName || this.userEmail) {
        this.$countly.setUserDetails({
          name: this.userName,
          email: this.userEmail
        });
        alert('用户属性已设置！');
      } else {
        alert('请填写至少一个字段');
      }
    },
    logError() {
      try {
        throw new Error('Test error from Vue');
      } catch (error) {
        this.$countly.logError(error);
        alert('错误已记录！');
      }
    }
  }
};
</script>

<style>
#app {
  font-family: Avenir, Helvetica, Arial, sans-serif;
  text-align: center;
  color: #2c3e50;
  margin-top: 60px;
}

.section {
  margin: 30px 0;
  padding: 20px;
  background: #f5f5f5;
  border-radius: 10px;
}

button {
  margin: 5px;
  padding: 10px 20px;
  background: #42b983;
  color: white;
  border: none;
  border-radius: 5px;
  cursor: pointer;
}

button:hover {
  background: #35a372;
}

button.danger {
  background: #e74c3c;
}

button.danger:hover {
  background: #c0392b;
}

input {
  margin: 5px;
  padding: 10px;
  border: 1px solid #ddd;
  border-radius: 5px;
  width: 200px;
}
</style>
```

### 运行Vue应用

```bash
npm run dev
```

---

## 🧪 测试验证

### 步骤1：打开浏览器开发者工具

```
Chrome/Edge: F12 或 Ctrl+Shift+I
Firefox: F12
Safari: Cmd+Option+I (macOS)
```

### 步骤2：查看Console日志

**预期日志**：
```
[Countly] Initializing...
[Countly] SDK initialized
[Countly] Recording event: button_clicked
[Countly] Sending request...
[Countly] Request completed successfully
```

### 步骤3：查看Network请求

在Network标签中过滤`countly`或服务器地址，查看：
- 请求URL
- 请求参数
- 响应状态

### 步骤4：在Countly管理界面验证

**登录Countly**：
```
URL: http://YOUR_IP:8080
```

**验证数据**：

1. **实时数据**
   ```
   Dashboard > Real-time
   - 查看实时用户
   - 查看实时事件
   ```

2. **事件数据**
   ```
   Analytics > Events
   - 查看事件列表
   - 查看事件参数
   ```

3. **用户数据**
   ```
   Analytics > Users
   - 查看用户列表
   - 查看用户属性
   ```

4. **页面浏览**
   ```
   Analytics > Views
   - 查看页面浏览量
   - 查看停留时间
   ```

5. **错误报告**
   ```
   Crashes > Overview
   - 查看JavaScript错误
   - 查看错误详情
   ```

### 测试清单

```
✅ 页面加载
   - 打开页面
   - 查看Console日志
   - 确认SDK初始化

✅ 简单事件
   - 点击按钮
   - 查看Console日志
   - 在Countly验证

✅ 带参数事件
   - 点击按钮
   - 验证事件参数

✅ 用户属性
   - 填写表单
   - 提交数据
   - 在Countly查看用户信息

✅ 页面浏览
   - 切换页面
   - 验证浏览记录

✅ 错误追踪
   - 触发错误
   - 在Countly查看错误报告

✅ 会话追踪
   - 刷新页面
   - 验证会话数据
```

---

## ❓ 常见问题

### 1. CORS跨域问题

**问题**：浏览器报CORS错误

**解决方案**：

在Countly服务器的Nginx配置中添加CORS头：

```nginx
# /etc/nginx/sites-available/countly
location / {
    # 添加CORS头
    add_header 'Access-Control-Allow-Origin' '*';
    add_header 'Access-Control-Allow-Methods' 'GET, POST, OPTIONS';
    add_header 'Access-Control-Allow-Headers' 'Content-Type';

    # 其他配置...
}
```

或在Countly配置中启用CORS。

### 2. 无法加载SDK

**问题**：SDK加载失败

**解决方案**：

```javascript
// 方案1：使用备用CDN
cly.src = 'https://unpkg.com/countly-sdk-web@latest/lib/countly.min.js';

// 方案2：使用NPM安装
npm install countly-sdk-web

// 方案3：从Countly服务器加载
cly.src = 'http://YOUR_IP:8080/sdk/web/countly.min.js';
```

### 3. 事件未发送

**问题**：点击按钮后事件未发送

**排查步骤**：

```javascript
// 1. 检查SDK是否初始化
console.log(Countly);

// 2. 启用调试模式
Countly.debug = true;

// 3. 检查网络请求
// 在Network标签查看请求

// 4. 检查配置
console.log(Countly.app_key);
console.log(Countly.url);
```

### 4. 本地文件无法使用

**问题**：直接打开HTML文件无法工作

**解决方案**：

使用本地服务器：

```bash
# 方案1：http-server
npm install -g http-server
http-server -p 8000

# 方案2：Python
python -m http.server 8000

# 方案3：PHP
php -S localhost:8000

# 方案4：VS Code Live Server插件
# 安装Live Server插件，右键选择"Open with Live Server"
```

### 5. 混合内容警告

**问题**：HTTPS页面加载HTTP资源

**解决方案**：

```javascript
// 使用HTTPS连接Countly
Countly.url = 'https://YOUR_DOMAIN';

// 或配置Countly服务器支持HTTPS
// 参考安装部署指南中的SSL配置
```

### 6. 广告拦截器阻止

**问题**：广告拦截器阻止Countly请求

**解决方案**：

```javascript
// 使用自定义域名
Countly.url = 'https://analytics.yourdomain.com';

// 或提示用户禁用广告拦截器
```


<!--  -->
---

## 🚀 高级功能

### 1. 热力图

```javascript
// 启用点击热力图
Countly.q.push(['track_clicks']);

// 启用滚动热力图
Countly.q.push(['track_scrolls']);

// 在Countly管理界面查看：
// Analytics > Heatmaps
```

### 2. 表单分析

```javascript
// 自动追踪表单
Countly.q.push(['track_forms']);

// 或手动追踪特定表单
Countly.q.push(['collect_from_forms', ['#myForm']]);

// 追踪表单字段
Countly.q.push(['add_event', {
    key: 'form_field_interaction',
    segmentation: {
        field_name: 'email',
        action: 'focus'
    }
}]);
```

### 3. 性能监控

```javascript
// 自动追踪页面加载时间
Countly.q.push(['track_performance']);

// 手动记录性能指标
Countly.q.push(['report_performance', {
    name: 'page_load',
    duration: 1234,  // 毫秒
    network: 500,
    server: 300,
    render: 434
}]);
```

### 4. 用户同意管理（GDPR）

```javascript
// 需要用户同意
Countly.require_consent = true;

// 初始化
Countly.init();

// 用户同意后
Countly.q.push(['give_consent', ['sessions', 'events', 'views']]);

// 撤销同意
Countly.q.push(['remove_consent', ['sessions']]);

// 检查同意状态
Countly.q.push(['check_consent', 'sessions']);
```

### 5. 远程配置

```javascript
// 获取远程配置
Countly.fetch_remote_config(function(err, remoteConfig) {
    if (!err) {
        console.log('Remote config:', remoteConfig);

        // 使用配置
        var featureEnabled = remoteConfig.feature_flag;
        var buttonColor = remoteConfig.button_color;
    }
});

// 获取特定键值
Countly.get_remote_config_value('feature_flag', function(err, value) {
    if (!err) {
        console.log('Feature flag:', value);
    }
});
```

### 6. A/B测试

```javascript
// 获取A/B测试变体
Countly.fetch_remote_config(function(err, config) {
    if (!err) {
        var variant = config.button_variant;

        // 根据变体显示不同内容
        if (variant === 'blue') {
            document.getElementById('myButton').style.background = 'blue';
        } else {
            document.getElementById('myButton').style.background = 'green';
        }
    }
});
```

### 7. 用户反馈

```javascript
// 显示反馈小部件
Countly.show_feedback_widget('widget_id');

// 显示NPS调查
Countly.show_nps_widget();

// 显示评分小部件
Countly.show_rating_widget('widget_id');
```

### 8. 自定义设备ID

```javascript
// 设置自定义设备ID
Countly.device_id = 'custom_user_id_123';

// 或在初始化时设置
Countly.q.push(['set_id', 'custom_user_id_123']);

// 更改设备ID
Countly.q.push(['change_id', 'new_user_id_456', true]);
```

---

## 📊 最佳实践

### 1. 事件命名规范

```javascript
// ✅ 好的命名
'button_clicked'
'form_submitted'
'video_played'
'purchase_completed'

// ❌ 不好的命名
'click'
'submit'
'play'
'buy'

// 使用一致的命名风格
// 推荐：snake_case 或 camelCase
```

### 2. 事件参数设计

```javascript
// ✅ 结构化的参数
Countly.q.push(['add_event', {
    key: 'product_viewed',
    segmentation: {
        product_id: 'prod_123',
        product_name: 'Premium Plan',
        category: 'subscription',
        price: 99.99,
        currency: 'USD'
    }
}]);

// ❌ 混乱的参数
Countly.q.push(['add_event', {
    key: 'event',
    segmentation: {
        data: 'prod_123,Premium Plan,99.99'
    }
}]);
```

### 3. 性能优化

```javascript
// 批量发送事件
Countly.q.push(['set_event_queue_size', 10]);

// 减少会话更新频率
Countly.session_update = 60;  // 60秒

// 离线模式
Countly.q.push(['enable_offline_mode']);
```

### 4. 错误处理

```javascript
// 全局错误捕获
window.addEventListener('error', function(event) {
    Countly.q.push(['log_error', event.error]);
});

// Promise错误捕获
window.addEventListener('unhandledrejection', function(event) {
    Countly.q.push(['log_error', event.reason]);
});

// 自定义错误处理
try {
    // 可能出错的代码
} catch (error) {
    Countly.q.push(['log_error', error]);
    // 其他错误处理
}
```

### 5. 隐私保护

```javascript
// 不收集IP地址
Countly.q.push(['disable_ip_tracking']);

// 不收集位置信息
Countly.q.push(['disable_location']);

// 匿名化用户
Countly.q.push(['set_id', 'anonymous_' + Math.random()]);

// 遵守Do Not Track
if (navigator.doNotTrack === '1') {
    // 不初始化Countly
}
```

---

## 📚 参考资源

### 官方文档

- **Web SDK文档**：https://support.count.ly/hc/en-us/articles/360037441932-Web-analytics-JavaScript
- **GitHub仓库**：https://github.com/Countly/countly-sdk-web
- **NPM包**：https://www.npmjs.com/package/countly-sdk-web
- **API参考**：https://countly.github.io/countly-sdk-web/

### 示例代码

- **官方Demo**：https://github.com/Countly/countly-sdk-web/tree/master/examples
- **CDN示例**：https://try.count.ly
- **集成示例**：https://github.com/Countly/countly-sdk-web/wiki

### 框架集成

- **React集成**：https://github.com/Countly/countly-sdk-react-native
- **Vue集成**：https://github.com/Countly/countly-sdk-web/wiki/Vue-integration
- **Angular集成**：https://github.com/Countly/countly-sdk-web/wiki/Angular-integration

### 学习资源

- **Countly博客**：https://countly.com/blog
- **视频教程**：https://www.youtube.com/c/CountlyAnalytics
- **社区论坛**：https://community.count.ly

---

## ✅ 完成检查清单

```
环境准备：
- [ ] 准备开发环境
- [ ] 准备Countly服务器
- [ ] 获取App Key

基础集成：
- [ ] 选择集成方式（CDN/NPM）
- [ ] 添加SDK代码
- [ ] 配置App Key和服务器地址
- [ ] 初始化SDK

功能实现：
- [ ] 基础事件追踪
- [ ] 用户属性设置
- [ ] 页面浏览追踪
- [ ] 错误追踪
- [ ] 表单追踪（可选）

框架集成（可选）：
- [ ] React集成
- [ ] Vue集成
- [ ] Angular集成

测试验证：
- [ ] 浏览器Console测试
- [ ] Network请求验证
- [ ] Countly管理界面验证
- [ ] 跨浏览器测试

优化完善：
- [ ] 性能优化
- [ ] 错误处理
- [ ] 隐私保护
- [ ] 代码优化
```

---

## 🎯 平台对比

### Web vs 移动端

| 特性 | Web | Android | iOS | Flutter |
|------|-----|---------|-----|---------|
| **集成难度** | ⭐ 最简单 | ⭐⭐ 简单 | ⭐⭐ 简单 | ⭐⭐ 简单 |
| **开发时间** | 5分钟 | 1-2小时 | 1-2小时 | 1-2小时 |
| **代码量** | 最少 | 中等 | 中等 | 中等 |
| **跨平台** | ✅ | ❌ | ❌ | ✅ |
| **性能** | 依赖浏览器 | 原生 | 原生 | 接近原生 |
| **热更新** | ✅ 即时 | ❌ | ❌ | ❌ |

### Web集成优势

```
✅ 无需编译：修改即生效
✅ 跨平台：所有浏览器都支持
✅ 易调试：浏览器开发者工具
✅ 快速部署：上传即可使用
✅ 易维护：统一的代码库
✅ 低门槛：只需HTML/JavaScript
```

---

## 🎉 总结

### Web集成特点

**最简单的集成方式**：
- 只需添加一段JavaScript代码
- 5分钟即可完成基础集成
- 无需编译和构建
- 支持所有现代浏览器

**功能完整**：
- 事件追踪
- 用户属性
- 页面浏览
- 错误追踪
- 性能监控
- 热力图
- A/B测试

**框架友好**：
- 原生JavaScript
- React
- Vue
- Angular
- 其他框架

### 下一步

1. **基础集成**
   - 添加SDK代码
   - 测试基础功能
   - 验证数据

2. **功能扩展**
   - 添加自定义事件
   - 设置用户属性
   - 启用高级功能

3. **优化完善**
   - 性能优化
   - 错误处理
   - 隐私保护

4. **生产部署**
   - 关闭调试模式
   - 配置HTTPS
   - 监控数据质量

---

**恭喜！你已经完成了Countly Web SDK的集成和测试！**

现在你可以：
1. 在任何网站中集成Countly
2. 追踪用户行为和事件
3. 分析网站数据
4. 优化用户体验

Web集成是最简单快速的方式，非常适合快速验证和测试Countly功能！

---

*文档版本：v1.0*
*最后更新：2026年1月28日*
*作者：Kiro AI*
