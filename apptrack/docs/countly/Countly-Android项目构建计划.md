# Countly Android Demo 项目构建计划

> Android测试应用完整开发计划
> 日期：2026年1月28日

---

## 📋 目录

1. [项目概述](#项目概述)
2. [开发阶段](#开发阶段)
3. [详细任务清单](#详细任务清单)
4. [时间规划](#时间规划)
5. [技术要点](#技术要点)
6. [测试计划](#测试计划)

---

## 🎯 项目概述

### 项目目标

创建一个功能完整的Android Demo应用，用于测试和演示Countly SDK的所有核心功能。

### 项目信息

```
项目名称：CountlyDemo
包名：com.example.countlydemo
语言：Kotlin
最低SDK：API 21 (Android 5.0)
目标SDK：API 34 (Android 14)
预计工期：2-3天
```

### 核心功能

```
1. 基础追踪
   - 应用生命周期追踪
   - 屏幕浏览追踪
   - 会话管理

2. 事件追踪
   - 简单事件
   - 带参数事件
   - 计时事件

3. 用户管理
   - 用户属性设置
   - 自定义属性
   - 用户画像

4. 崩溃报告
   - 异常捕获
   - 崩溃记录
   - 堆栈追踪

5. 高级功能（可选）
   - 推送通知
   - 远程配置
   - A/B测试
```

---

## 🚀 开发阶段

### 阶段1：环境准备（0.5天）

**任务**：
- [ ] 安装Android Studio
- [ ] 配置JDK环境
- [ ] 创建Android虚拟设备
- [ ] 准备Countly服务器信息

**产出**：
- 完整的开发环境
- 可用的测试设备

---

### 阶段2：项目初始化（0.5天）

**任务**：
- [ ] 创建Android项目
- [ ] 配置项目结构
- [ ] 添加Countly SDK依赖
- [ ] 配置AndroidManifest
- [ ] 创建Application类

**产出**：
- 项目骨架
- SDK集成完成

**关键代码**：

```kotlin
// build.gradle.kts
dependencies {
    implementation("ly.count.android:sdk:24.1.0")
}

// CountlyDemoApp.kt
class CountlyDemoApp : Application() {
    override fun onCreate() {
        super.onCreate()
        initCountly()
    }
}
```

---

### 阶段3：主界面开发（0.5天）

**任务**：
- [ ] 设计主界面布局
- [ ] 实现MainActivity
- [ ] 添加基础事件按钮
- [ ] 实现事件发送逻辑
- [ ] 添加Toast提示

**产出**：
- 完整的主界面
- 基础事件功能

**界面元素**：
```
- 标题文本
- 发送简单事件按钮
- 发送带参数事件按钮
- 设置用户属性按钮
- 模拟崩溃按钮
- 跳转测试页面按钮
- 信息提示文本
```

---

### 阶段4：事件测试页面（0.5天）

**任务**：
- [ ] 设计事件测试界面
- [ ] 实现EventTestActivity
- [ ] 添加各类事件测试
- [ ] 实现计时事件
- [ ] 添加事件日志显示

**产出**：
- 事件测试页面
- 完整的事件追踪功能

**测试事件**：
```
- 购买事件（带金额）
- 登录事件（带方法）
- 分享事件（带平台）
- 计时事件（视频观看）
```

---

### 阶段5：用户属性页面（0.5天）

**任务**：
- [ ] 设计用户属性界面
- [ ] 实现UserProfileActivity
- [ ] 添加表单输入
- [ ] 实现属性保存
- [ ] 添加自定义属性

**产出**：
- 用户属性页面
- 用户数据管理功能

**用户属性**：
```
基础属性：
- 姓名
- 邮箱
- 年龄
- 性别

自定义属性：
- 订阅类型
- 注册日期
- 偏好分类
- 通知设置
```

---

### 阶段6：测试验证（0.5天）

**任务**：
- [ ] 功能测试
- [ ] 数据验证
- [ ] 日志检查
- [ ] 性能测试
- [ ] 修复Bug

**产出**：
- 测试报告
- Bug修复

**测试清单**：
```
✅ 应用启动正常
✅ SDK初始化成功
✅ 事件发送成功
✅ 数据显示正确
✅ 崩溃记录正常
✅ 用户属性保存成功
✅ 屏幕浏览追踪正常
```

---

### 阶段7：打包发布（0.5天）

**任务**：
- [ ] 生成签名密钥
- [ ] 配置签名信息
- [ ] 构建Debug APK
- [ ] 构建Release APK
- [ ] 编写使用文档

**产出**：
- Debug APK
- Release APK
- 使用说明

---

## 📝 详细任务清单

### Day 1：基础搭建

**上午（4小时）**：
```
09:00-10:00  环境准备
             - 安装Android Studio
             - 配置开发环境
             - 创建虚拟设备

10:00-11:00  项目初始化
             - 创建项目
             - 添加依赖
             - 配置权限

11:00-12:00  Application类
             - 创建CountlyDemoApp
             - 初始化Countly SDK
             - 测试连接

12:00-13:00  午休
```

**下午（4小时）**：
```
13:00-15:00  主界面开发
             - 设计布局
             - 实现MainActivity
             - 添加按钮事件

15:00-17:00  基础功能测试
             - 运行应用
             - 测试事件发送
             - 验证数据
```

---

### Day 2：功能完善

**上午（4小时）**：
```
09:00-11:00  事件测试页面
             - 设计界面
             - 实现EventTestActivity
             - 添加各类事件

11:00-13:00  用户属性页面
             - 设计界面
             - 实现UserProfileActivity
             - 添加表单功能
```

**下午（4小时）**：
```
13:00-15:00  功能测试
             - 测试所有功能
             - 验证数据
             - 修复Bug

15:00-17:00  优化完善
             - 代码优化
             - UI优化
             - 添加注释
```

---

### Day 3：测试发布

**上午（4小时）**：
```
09:00-11:00  完整测试
             - 功能测试
             - 性能测试
             - 兼容性测试

11:00-13:00  打包发布
             - 生成签名
             - 构建APK
             - 测试安装
```

**下午（2小时）**：
```
13:00-15:00  文档编写
             - 使用说明
             - 测试报告
             - 问题记录
```

---

## ⏱️ 时间规划

### 总体时间分配

| 阶段 | 任务 | 预计时间 | 优先级 |
|------|------|---------|--------|
| 1 | 环境准备 | 0.5天 | P0 |
| 2 | 项目初始化 | 0.5天 | P0 |
| 3 | 主界面开发 | 0.5天 | P0 |
| 4 | 事件测试页面 | 0.5天 | P1 |
| 5 | 用户属性页面 | 0.5天 | P1 |
| 6 | 测试验证 | 0.5天 | P0 |
| 7 | 打包发布 | 0.5天 | P1 |
| **总计** | | **3.5天** | |

### 里程碑

```
Day 1 结束：
✅ 项目创建完成
✅ SDK集成完成
✅ 主界面完成
✅ 基础功能可用

Day 2 结束：
✅ 所有页面完成
✅ 所有功能实现
✅ 基础测试通过

Day 3 结束：
✅ 完整测试通过
✅ APK构建完成
✅ 文档编写完成
✅ 项目交付
```

---

## 🔧 技术要点

### 1. Countly SDK初始化

**关键配置**：

```kotlin
val config = CountlyConfig(this, appKey, serverUrl).apply {
    // 日志
    setLoggingEnabled(BuildConfig.DEBUG)

    // 崩溃报告
    enableCrashReporting()

    // 视图追踪
    setViewTracking(true)
    enableAutomaticViewTracking()

    // 设备ID
    setDeviceId(DeviceId.Type.OPEN_UDID)

    // 会话
    setRequiresConsent(false)
}
```

**注意事项**：
- 在Application的onCreate中初始化
- 使用正确的服务器地址（http://IP:8080）
- 从Countly管理界面获取正确的App Key
- 开发环境启用日志，生产环境关闭

### 2. 生命周期管理

**Activity生命周期**：

```kotlin
override fun onStart() {
    super.onStart()
    Countly.sharedInstance().onStart(this)
}

override fun onStop() {
    Countly.sharedInstance().onStop()
    super.onStop()
}
```

**重要**：
- 每个Activity都要调用onStart()和onStop()
- 确保会话正确追踪
- 避免内存泄漏

### 3. 事件追踪

**事件类型**：

```kotlin
// 1. 简单事件
Countly.sharedInstance().events().recordEvent("event_name")

// 2. 带参数事件
val segmentation = HashMap<String, Any>()
segmentation["key"] = "value"
Countly.sharedInstance().events().recordEvent(
    "event_name",
    segmentation,
    count,
    sum
)

// 3. 计时事件
Countly.sharedInstance().events().startEvent("timed_event")
// ... 执行操作 ...
Countly.sharedInstance().events().endEvent("timed_event")
```

### 4. 用户属性

**设置属性**：

```kotlin
val userProfile = HashMap<String, Any>()
userProfile["name"] = "John Doe"
userProfile["email"] = "john@example.com"
userProfile["byear"] = 1990

Countly.sharedInstance().userProfile().setProperties(userProfile)
Countly.sharedInstance().userProfile().save()
```

### 5. 崩溃报告

**记录异常**：

```kotlin
try {
    // 可能抛出异常的代码
} catch (e: Exception) {
    Countly.sharedInstance().crashes().recordHandledException(e)
}
```

---

## 🧪 测试计划

### 功能测试

**测试用例1：应用启动**
```
步骤：
1. 启动应用
2. 查看Logcat日志

预期结果：
- 应用正常启动
- Countly SDK初始化成功
- 日志显示"SDK initialized"
```

**测试用例2：简单事件**
```
步骤：
1. 点击"发送简单事件"按钮
2. 查看Toast提示
3. 登录Countly管理界面
4. 查看Events页面

预期结果：
- Toast显示"简单事件已发送"
- Countly管理界面显示事件
- 事件名称为"button_clicked"
```

**测试用例3：带参数事件**
```
步骤：
1. 点击"发送带参数事件"按钮
2. 查看Toast提示
3. 在Countly管理界面查看事件详情

预期结果：
- Toast显示"带参数事件已发送"
- 事件包含正确的参数
- 参数值正确显示
```

**测试用例4：用户属性**
```
步骤：
1. 点击"设置用户属性"按钮
2. 查看Toast提示
3. 在Countly管理界面查看用户列表

预期结果：
- Toast显示"用户属性已设置"
- 用户属性正确保存
- 管理界面显示用户信息
```

**测试用例5：崩溃报告**
```
步骤：
1. 点击"模拟崩溃"按钮
2. 查看Toast提示
3. 在Countly管理界面查看Crashes页面

预期结果：
- Toast显示"崩溃已记录"
- 崩溃报告已上传
- 堆栈信息完整
```

**测试用例6：屏幕浏览**
```
步骤：
1. 点击"事件测试页面"按钮
2. 跳转到新页面
3. 返回主页面
4. 在Countly管理界面查看Views页面

预期结果：
- 页面跳转正常
- 屏幕浏览已记录
- 显示正确的页面名称
```

### 性能测试

**测试项目**：
```
1. 应用启动时间
   - 冷启动：<3秒
   - 热启动：<1秒

2. 内存占用
   - 空闲状态：<50MB
   - 运行状态：<100MB

3. 网络流量
   - 每次事件：<1KB
   - 每小时：<100KB

4. 电池消耗
   - 后台运行：<1%/小时
```

### 兼容性测试

**测试设备**：
```
1. Android 5.0 (API 21)
2. Android 8.0 (API 26)
3. Android 11 (API 30)
4. Android 14 (API 34)

设备类型：
- 手机
- 平板
- 模拟器
```

---

## 📦 交付物

### 代码文件

```
CountlyDemo/
├── app/
│   ├── src/main/
│   │   ├── java/com/example/countlydemo/
│   │   │   ├── CountlyDemoApp.kt
│   │   │   ├── MainActivity.kt
│   │   │   ├── EventTestActivity.kt
│   │   │   └── UserProfileActivity.kt
│   │   ├── res/
│   │   │   └── layout/
│   │   │       ├── activity_main.xml
│   │   │       ├── activity_event_test.xml
│   │   │       └── activity_user_profile.xml
│   │   └── AndroidManifest.xml
│   └── build.gradle.kts
└── build.gradle.kts
```

### APK文件

```
1. app-debug.apk
   - 用于开发测试
   - 包含调试信息
   - 文件大小：~5MB

2. app-release.apk
   - 用于生产发布
   - 已签名
   - 已混淆
   - 文件大小：~3MB
```

### 文档

```
1. README.md
   - 项目说明
   - 安装步骤
   - 使用方法

2. 集成指南.md
   - SDK集成步骤
   - 配置说明
   - 代码示例

3. 测试报告.md
   - 测试用例
   - 测试结果
   - 问题记录

4. 构建说明.md
   - 构建步骤
   - 签名配置
   - 发布流程
```

---

## ✅ 验收标准

### 功能完整性

```
✅ 所有计划功能已实现
✅ 所有页面正常工作
✅ 所有按钮响应正常
✅ 数据正确发送到Countly
✅ 管理界面能看到数据
```

### 代码质量

```
✅ 代码结构清晰
✅ 命名规范统一
✅ 注释完整
✅ 无明显Bug
✅ 无内存泄漏
```

### 用户体验

```
✅ 界面美观
✅ 操作流畅
✅ 提示清晰
✅ 无卡顿
✅ 无崩溃
```

### 文档完整性

```
✅ 代码注释完整
✅ 使用文档清晰
✅ 测试报告详细
✅ 问题记录完整
```

---

## 🎯 后续计划

### 功能扩展

```
1. 推送通知
   - 集成FCM
   - 实现推送接收
   - 测试推送功能

2. 远程配置
   - 实现功能开关
   - 测试配置更新
   - 验证生效时间

3. A/B测试
   - 实现多变体
   - 测试分流
   - 分析结果

4. 用户反馈
   - 添加反馈表单
   - 实现NPS评分
   - 收集用户意见
```

### 性能优化

```
1. 减少网络请求
2. 优化内存使用
3. 减少电池消耗
4. 提升启动速度
```

### 代码优化

```
1. 重构代码结构
2. 提取公共方法
3. 优化UI布局
4. 添加单元测试
```

---

## 📞 支持与反馈

### 遇到问题？

1. **查看文档**：先查看集成指南和常见问题
2. **查看日志**：检查Logcat中的Countly日志
3. **测试连接**：确认能访问Countly服务器
4. **社区求助**：在Countly社区论坛提问

### 反馈渠道

- **GitHub Issues**：报告Bug和建议
- **Discord社区**：实时交流
- **官方论坛**：深度讨论

---

**祝你构建顺利！**

如有任何问题，随时查阅文档或寻求帮助。

---

*文档版本：v1.0*
*最后更新：2026年1月28日*
*作者：Kiro AI*
