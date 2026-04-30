<template>
  <div id="app">
    <header class="app-header">
      <h1>🚀 Countly Vue Demo</h1>
      <p class="subtitle">完整功能测试应用</p>
      <div class="event-counter">
        <span>事件计数: </span>
        <span class="count">{{ eventCount }}</span>
      </div>

      <!-- Basic Events Section -->
      <div class="section">
        <h2>📊 基础事件</h2>
        <button @click="sendSimpleEvent">发送简单事件</button>
        <button @click="sendEventWithParams">发送带参数事件</button>
        <button @click="toggleTimedEvent">
          {{ timedEventActive ? '结束定时事件' : '开始定时事件' }}
        </button>
      </div>

      <!-- User Profile Section -->
      <div class="section">
        <h2>👤 用户属性</h2>
        <input
          v-model="userName"
          type="text"
          placeholder="姓名"
        />
        <input
          v-model="userEmail"
          type="email"
          placeholder="邮箱"
        />
        <input
          v-model="userAge"
          type="number"
          placeholder="年龄"
        />
        <button @click="setUserProfile">设置用户属性</button>
        <button @click="setCustomProperty">设置自定义属性</button>
      </div>

      <!-- Page View Section -->
      <div class="section">
        <h2>📄 页面浏览</h2>
        <button @click="trackPageView('home')">追踪首页</button>
        <button @click="trackPageView('products')">追踪产品页</button>
        <button @click="trackPageView('about')">追踪关于页</button>
      </div>

      <!-- Error Tracking Section -->
      <div class="section">
        <h2>🐛 错误追踪</h2>
        <button @click="logError" class="danger">记录错误</button>
        <button @click="throwError" class="danger">触发异常</button>
      </div>

      <!-- Form Test Section -->
      <div class="section">
        <h2>📝 表单测试</h2>
        <input
          v-model="formInput"
          type="text"
          placeholder="输入内容"
        />
        <textarea
          v-model="formTextarea"
          placeholder="输入评论"
          rows="3"
        ></textarea>
        <button @click="submitForm">提交表单</button>
      </div>

      <!-- Event Log Section -->
      <div class="section">
        <h2>📋 事件日志</h2>
        <div class="log" ref="logContainer">
          <div
            v-for="(log, index) in logs"
            :key="index"
            class="log-entry"
            :class="log.type"
          >
            [{{ log.time }}] {{ log.message }}
          </div>
        </div>
        <button @click="clearLogs" class="secondary">清空日志</button>
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
      userEmail: '',
      userAge: '',
      formInput: '',
      formTextarea: '',
      logs: [],
      timedEventActive: false,
      timedEventStartTime: null
    };
  },
  mounted() {
    // Track page view on mount
    this.$countly.trackPageView('Home');
    this.addLog('Countly初始化成功', 'success');
    this.addLog('页面加载完成', 'info');
  },
  methods: {
    // Add log entry
    addLog(message, type = 'info') {
      const time = new Date().toLocaleTimeString();
      this.logs.push({ time, message, type });

      // Auto scroll to bottom
      this.$nextTick(() => {
        const container = this.$refs.logContainer;
        if (container) {
          container.scrollTop = container.scrollHeight;
        }
      });
    },

    // Clear logs
    clearLogs() {
      this.logs = [];
      this.addLog('日志已清空', 'info');
    },

    // 1. Send simple event
    sendSimpleEvent() {
      this.$countly.recordEvent('button_clicked');
      this.eventCount++;
      this.addLog('✓ 简单事件已发送: button_clicked', 'success');
    },

    // 2. Send event with parameters
    sendEventWithParams() {
      this.$countly.recordEvent('button_clicked_with_params', {
        button_name: 'event_with_params',
        page: 'App',
        count: this.eventCount,
        timestamp: Date.now()
      });
      this.eventCount++;
      this.addLog('✓ 带参数事件已发送', 'success');
    },

    // 3. Toggle timed event
    toggleTimedEvent() {
      if (!this.timedEventActive) {
        // Start timed event
        this.$countly.startEvent('timed_action');
        this.timedEventActive = true;
        this.timedEventStartTime = Date.now();
        this.addLog('⏱ 定时事件已开始', 'success');
      } else {
        // End timed event
        const duration = ((Date.now() - this.timedEventStartTime) / 1000).toFixed(2);
        this.$countly.endEvent('timed_action', {
          duration: duration + 's'
        });
        this.timedEventActive = false;
        this.timedEventStartTime = null;
        this.addLog(`✓ 定时事件已结束: ${duration}秒`, 'success');
      }
    },

    // 4. Set user profile
    setUserProfile() {
      if (!this.userName && !this.userEmail && !this.userAge) {
        this.addLog('✗ 请填写至少一个字段', 'error');
        return;
      }

      const userData = {};
      if (this.userName) userData.name = this.userName;
      if (this.userEmail) userData.email = this.userEmail;
      if (this.userAge) userData.age = parseInt(this.userAge);

      this.$countly.setUserDetails(userData);
      this.addLog('✓ 用户属性已设置', 'success');
    },

    // 5. Set custom property
    setCustomProperty() {
      this.$countly.setUserDetails({
        custom: {
          subscription: 'premium',
          last_login: new Date().toISOString(),
          app_version: '1.0.0',
          platform: 'web'
        }
      });
      this.addLog('✓ 自定义属性已设置', 'success');
    },

    // 6. Track page view
    trackPageView(pageName) {
      this.$countly.trackPageView(pageName);
      this.addLog(`✓ 页面浏览已追踪: ${pageName}`, 'success');
    },

    // 7. Log error
    logError() {
      const errorMessage = 'This is a test error from Vue';
      this.$countly.logError(errorMessage);
      this.addLog('✓ 错误已记录', 'success');
    },

    // 8. Throw error
    throwError() {
      try {
        throw new Error('Test exception for Countly');
      } catch (error) {
        this.$countly.logError(error);
        this.addLog('✓ 异常已捕获并记录', 'success');
      }
    },

    // 9. Submit form
    submitForm() {
      if (!this.formInput && !this.formTextarea) {
        this.addLog('✗ 表单为空', 'error');
        return;
      }

      this.$countly.recordEvent('form_submitted', {
        has_input: this.formInput ? 'yes' : 'no',
        has_comment: this.formTextarea ? 'yes' : 'no',
        input_length: this.formInput.length,
        comment_length: this.formTextarea.length
      });

      this.addLog('✓ 表单提交事件已发送', 'success');

      // Clear form
      this.formInput = '';
      this.formTextarea = '';
    }
  },
  beforeUnmount() {
    // End session before unmount
    this.addLog('页面即将卸载', 'info');
  }
};
</script>

<style scoped>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

#app {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  min-height: 100vh;
  padding: 20px;
}

.app-header {
  max-width: 800px;
  margin: 0 auto;
  background: white;
  border-radius: 20px;
  padding: 40px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
}

h1 {
  color: #333;
  margin-bottom: 10px;
  font-size: 32px;
  text-align: center;
}

.subtitle {
  color: #666;
  margin-bottom: 20px;
  font-size: 16px;
  text-align: center;
}

.event-counter {
  text-align: center;
  padding: 15px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  border-radius: 10px;
  margin-bottom: 30px;
  font-size: 18px;
  font-weight: bold;
}

.event-counter .count {
  font-size: 24px;
  color: #ffd700;
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
  font-weight: 500;
}

button:hover {
  background: #5568d3;
  transform: translateY(-2px);
  box-shadow: 0 5px 15px rgba(102, 126, 234, 0.4);
}

button:active {
  transform: translateY(0);
}

button.danger {
  background: #e74c3c;
}

button.danger:hover {
  background: #c0392b;
  box-shadow: 0 5px 15px rgba(231, 76, 60, 0.4);
}

button.secondary {
  background: #95a5a6;
}

button.secondary:hover {
  background: #7f8c8d;
}

input,
textarea {
  width: 100%;
  padding: 10px;
  margin: 10px 0;
  border: 2px solid #e0e0e0;
  border-radius: 8px;
  font-size: 14px;
  transition: border-color 0.3s;
}

input:focus,
textarea:focus {
  outline: none;
  border-color: #667eea;
}

textarea {
  resize: vertical;
  font-family: inherit;
}

.log {
  background: #2c3e50;
  color: #2ecc71;
  padding: 15px;
  border-radius: 8px;
  font-family: 'Courier New', monospace;
  font-size: 12px;
  max-height: 300px;
  overflow-y: auto;
  margin-bottom: 10px;
}

.log-entry {
  margin: 5px 0;
  padding: 2px 0;
}

.log-entry.success {
  color: #2ecc71;
}

.log-entry.error {
  color: #e74c3c;
}

.log-entry.info {
  color: #3498db;
}

/* Scrollbar styling */
.log::-webkit-scrollbar {
  width: 8px;
}

.log::-webkit-scrollbar-track {
  background: #34495e;
  border-radius: 4px;
}

.log::-webkit-scrollbar-thumb {
  background: #667eea;
  border-radius: 4px;
}

.log::-webkit-scrollbar-thumb:hover {
  background: #5568d3;
}

/* Responsive design */
@media (max-width: 768px) {
  .app-header {
    padding: 20px;
  }

  h1 {
    font-size: 24px;
  }

  button {
    padding: 10px 20px;
    font-size: 13px;
  }
}
</style>
