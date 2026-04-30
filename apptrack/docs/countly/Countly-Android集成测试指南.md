# Countly Android 集成测试指南

> 完整的Android Demo应用集成Countly SDK教程
> 日期：2026年1月28日

---

## 📋 目录

1. [项目概述](#项目概述)
2. [开发环境准备](#开发环境准备)
3. [创建Android项目](#创建android项目)
4. [集成Countly SDK](#集成countly-sdk)
5. [核心功能实现](#核心功能实现)
6. [测试验证](#测试验证)
7. [常见问题](#常见问题)

---

## 🎯 项目概述

### Demo应用功能

创建一个简单的Android应用来测试Countly的核心功能：

**功能列表**：
```
1. 基础追踪
   - 应用启动/退出
   - 屏幕浏览
   - 会话追踪

2. 自定义事件
   - 按钮点击事件
   - 用户操作事件
   - 带参数的事件

3. 用户属性
   - 设置用户信息
   - 自定义属性

4. 崩溃报告
   - 模拟崩溃
   - 异常捕获

5. 推送通知（可选）
   - 接收推送
   - 点击处理
```

### 技术栈

```
语言：Kotlin
最低SDK：API 21 (Android 5.0)
目标SDK：API 34 (Android 14)
构建工具：Gradle 8.0+
IDE：Android Studio Hedgehog (2023.1.1) 或更高
```

---

## 💻 开发环境准备

### 1. 安装Android Studio

下载地址：https://developer.android.com/studio

### 2. 配置JDK

```bash
# 推荐使用JDK 17
# Android Studio会自动下载，或手动配置
```

### 3. 创建Android虚拟设备（AVD）

```
1. 打开Android Studio
2. Tools > Device Manager
3. Create Device
4. 选择设备型号（推荐：Pixel 6）
5. 选择系统镜像（推荐：API 34）
6. 完成创建
```

### 4. 准备Countly服务器信息

```
服务器地址：http://YOUR_IP:8080
App Key：在Countly管理界面获取
```

**获取App Key步骤**：
1. 登录Countly管理界面
2. Management > Applications
3. 创建新应用或选择现有应用
4. 复制App Key

---

## 📱 创建Android项目

### 步骤1：创建新项目

```
1. 打开Android Studio
2. File > New > New Project
3. 选择 "Empty Activity"
4. 配置项目：
   - Name: CountlyDemo
   - Package name: com.example.countlydemo
   - Save location: 选择保存位置
   - Language: Kotlin
   - Minimum SDK: API 21
5. Finish
```

### 步骤2：项目结构

```
CountlyDemo/
├── app/
│   ├── src/
│   │   ├── main/
│   │   │   ├── java/com/example/countlydemo/
│   │   │   │   ├── CountlyDemoApp.kt
│   │   │   │   ├── MainActivity.kt
│   │   │   │   ├── EventTestActivity.kt
│   │   │   │   └── UserProfileActivity.kt
│   │   │   ├── res/
│   │   │   │   ├── layout/
│   │   │   │   │   ├── activity_main.xml
│   │   │   │   │   ├── activity_event_test.xml
│   │   │   │   │   └── activity_user_profile.xml
│   │   │   │   └── values/
│   │   │   │       ├── strings.xml
│   │   │   │       └── colors.xml
│   │   │   └── AndroidManifest.xml
│   │   └── build.gradle.kts
│   └── build.gradle.kts
└── build.gradle.kts
```

---

## 🔧 集成Countly SDK

### 步骤1：添加依赖

编辑`app/build.gradle.kts`：

```kotlin
plugins {
    id("com.android.application")
    id("org.jetbrains.kotlin.android")
}

android {
    namespace = "com.example.countlydemo"
    compileSdk = 34

    defaultConfig {
        applicationId = "com.example.countlydemo"
        minSdk = 21
        targetSdk = 34
        versionCode = 1
        versionName = "1.0"

        testInstrumentationRunner = "androidx.test.runner.AndroidJUnitRunner"
    }

    buildTypes {
        release {
            isMinifyEnabled = false
            proguardFiles(
                getDefaultProguardFile("proguard-android-optimize.txt"),
                "proguard-rules.pro"
            )
        }
    }

    compileOptions {
        sourceCompatibility = JavaVersion.VERSION_17
        targetCompatibility = JavaVersion.VERSION_17
    }

    kotlinOptions {
        jvmTarget = "17"
    }

    buildFeatures {
        viewBinding = true
    }
}

dependencies {
    // Countly SDK
    implementation("ly.count.android:sdk:24.1.0")

    // Android基础库
    implementation("androidx.core:core-ktx:1.12.0")
    implementation("androidx.appcompat:appcompat:1.6.1")
    implementation("com.google.android.material:material:1.11.0")
    implementation("androidx.constraintlayout:constraintlayout:2.1.4")

    // 测试库
    testImplementation("junit:junit:4.13.2")
    androidTestImplementation("androidx.test.ext:junit:1.1.5")
    androidTestImplementation("androidx.test.espresso:espresso-core:3.5.1")
}
```

### 步骤2：配置权限

编辑`AndroidManifest.xml`：

```xml
<?xml version="1.0" encoding="utf-8"?>
<manifest xmlns:android="http://schemas.android.com/apk/res/android"
    xmlns:tools="http://schemas.android.com/tools">

    <!-- 必需权限 -->
    <uses-permission android:name="android.permission.INTERNET" />
    <uses-permission android:name="android.permission.ACCESS_NETWORK_STATE" />

    <!-- 可选权限（用于更详细的设备信息） -->
    <uses-permission android:name="android.permission.READ_PHONE_STATE"
        tools:ignore="ProtectedPermissions" />

    <application
        android:name=".CountlyDemoApp"
        android:allowBackup="true"
        android:dataExtractionRules="@xml/data_extraction_rules"
        android:fullBackupContent="@xml/backup_rules"
        android:icon="@mipmap/ic_launcher"
        android:label="@string/app_name"
        android:roundIcon="@mipmap/ic_launcher_round"
        android:supportsRtl="true"
        android:theme="@style/Theme.CountlyDemo"
        tools:targetApi="31">

        <activity
            android:name=".MainActivity"
            android:exported="true">
            <intent-filter>
                <action android:name="android.intent.action.MAIN" />
                <category android:name="android.intent.category.LAUNCHER" />
            </intent-filter>
        </activity>

        <activity
            android:name=".EventTestActivity"
            android:exported="false" />

        <activity
            android:name=".UserProfileActivity"
            android:exported="false" />
    </application>

</manifest>
```

### 步骤3：创建Application类

创建`CountlyDemoApp.kt`：

```kotlin
package com.example.countlydemo

import android.app.Application
import ly.count.android.sdk.Countly
import ly.count.android.sdk.CountlyConfig
import ly.count.android.sdk.DeviceId

class CountlyDemoApp : Application() {

    override fun onCreate() {
        super.onCreate()

        // 初始化Countly
        initCountly()
    }

    private fun initCountly() {
        // Countly配置
        val config = CountlyConfig(
            this,
            "YOUR_APP_KEY",  // 替换为你的App Key
            "http://YOUR_IP:8080"  // 替换为你的Countly服务器地址
        ).apply {
            // 启用日志（开发环境）
            setLoggingEnabled(true)
            enableCrashReporting()

            // 自动追踪视图
            setViewTracking(true)

            // 自动追踪会话
            setRequiresConsent(false)

            // 设置设备ID模式
            setDeviceId(DeviceId.Type.OPEN_UDID)

            // 启用自动会话追踪
            enableAutomaticViewTracking()
        }

        // 初始化Countly
        Countly.sharedInstance().init(config)

        // 开始会话
        Countly.sharedInstance().onStart(this)
    }
}
```

**重要配置说明**：

```kotlin
// 必须替换的配置
YOUR_APP_KEY: 从Countly管理界面获取
YOUR_IP:8080: 你的Countly服务器地址

// 开发环境配置
setLoggingEnabled(true): 启用日志，方便调试

// 生产环境配置
setLoggingEnabled(false): 关闭日志
```

---

## 🎨 核心功能实现

### 1. 主界面（MainActivity）

创建`activity_main.xml`：

```xml
<?xml version="1.0" encoding="utf-8"?>
<androidx.constraintlayout.widget.ConstraintLayout
    xmlns:android="http://schemas.android.com/apk/res/android"
    xmlns:app="http://schemas.android.com/apk/res-auto"
    xmlns:tools="http://schemas.android.com/tools"
    android:layout_width="match_parent"
    android:layout_height="match_parent"
    android:padding="16dp"
    tools:context=".MainActivity">

    <TextView
        android:id="@+id/tvTitle"
        android:layout_width="wrap_content"
        android:layout_height="wrap_content"
        android:text="Countly Demo"
        android:textSize="24sp"
        android:textStyle="bold"
        app:layout_constraintEnd_toEndOf="parent"
        app:layout_constraintStart_toStartOf="parent"
        app:layout_constraintTop_toTopOf="parent"
        android:layout_marginTop="32dp" />

    <Button
        android:id="@+id/btnSimpleEvent"
        android:layout_width="0dp"
        android:layout_height="wrap_content"
        android:text="发送简单事件"
        app:layout_constraintEnd_toEndOf="parent"
        app:layout_constraintStart_toStartOf="parent"
        app:layout_constraintTop_toBottomOf="@id/tvTitle"
        android:layout_marginTop="32dp" />

    <Button
        android:id="@+id/btnEventWithParams"
        android:layout_width="0dp"
        android:layout_height="wrap_content"
        android:text="发送带参数事件"
        app:layout_constraintEnd_toEndOf="parent"
        app:layout_constraintStart_toStartOf="parent"
        app:layout_constraintTop_toBottomOf="@id/btnSimpleEvent"
        android:layout_marginTop="16dp" />

    <Button
        android:id="@+id/btnSetUserProfile"
        android:layout_width="0dp"
        android:layout_height="wrap_content"
        android:text="设置用户属性"
        app:layout_constraintEnd_toEndOf="parent"
        app:layout_constraintStart_toStartOf="parent"
        app:layout_constraintTop_toBottomOf="@id/btnEventWithParams"
        android:layout_marginTop="16dp" />

    <Button
        android:id="@+id/btnSimulateCrash"
        android:layout_width="0dp"
        android:layout_height="wrap_content"
        android:text="模拟崩溃"
        android:backgroundTint="@android:color/holo_red_dark"
        app:layout_constraintEnd_toEndOf="parent"
        app:layout_constraintStart_toStartOf="parent"
        app:layout_constraintTop_toBottomOf="@id/btnSetUserProfile"
        android:layout_marginTop="16dp" />

    <Button
        android:id="@+id/btnEventTest"
        android:layout_width="0dp"
        android:layout_height="wrap_content"
        android:text="事件测试页面"
        app:layout_constraintEnd_toEndOf="parent"
        app:layout_constraintStart_toStartOf="parent"
        app:layout_constraintTop_toBottomOf="@id/btnSimulateCrash"
        android:layout_marginTop="16dp" />

    <TextView
        android:id="@+id/tvInfo"
        android:layout_width="0dp"
        android:layout_height="wrap_content"
        android:text="点击按钮测试Countly功能\n查看Countly管理界面验证数据"
        android:textAlignment="center"
        android:layout_marginTop="32dp"
        app:layout_constraintEnd_toEndOf="parent"
        app:layout_constraintStart_toStartOf="parent"
        app:layout_constraintTop_toBottomOf="@id/btnEventTest" />

</androidx.constraintlayout.widget.ConstraintLayout>
```

创建`MainActivity.kt`：

```kotlin
package com.example.countlydemo

import android.content.Intent
import android.os.Bundle
import android.widget.Toast
import androidx.appcompat.app.AppCompatActivity
import com.example.countlydemo.databinding.ActivityMainBinding
import ly.count.android.sdk.Countly

class MainActivity : AppCompatActivity() {

    private lateinit var binding: ActivityMainBinding

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        binding = ActivityMainBinding.inflate(layoutInflater)
        setContentView(binding.root)

        setupListeners()

        // 记录屏幕浏览
        Countly.sharedInstance().views().startAutoStoppedView("MainActivity")
    }

    private fun setupListeners() {
        // 简单事件
        binding.btnSimpleEvent.setOnClickListener {
            Countly.sharedInstance().events().recordEvent("button_clicked")
            showToast("简单事件已发送")
        }

        // 带参数的事件
        binding.btnEventWithParams.setOnClickListener {
            val segmentation = HashMap<String, Any>()
            segmentation["button_name"] = "event_with_params"
            segmentation["screen"] = "MainActivity"
            segmentation["timestamp"] = System.currentTimeMillis()

            Countly.sharedInstance().events().recordEvent(
                "button_clicked_with_params",
                segmentation,
                1
            )
            showToast("带参数事件已发送")
        }

        // 设置用户属性
        binding.btnSetUserProfile.setOnClickListener {
            val userProfile = HashMap<String, Any>()
            userProfile["name"] = "Test User"
            userProfile["email"] = "test@example.com"
            userProfile["age"] = 25
            userProfile["gender"] = "M"

            Countly.sharedInstance().userProfile().setProperties(userProfile)
            Countly.sharedInstance().userProfile().save()

            showToast("用户属性已设置")
        }

        // 模拟崩溃
        binding.btnSimulateCrash.setOnClickListener {
            try {
                throw RuntimeException("这是一个测试崩溃")
            } catch (e: Exception) {
                Countly.sharedInstance().crashes().recordHandledException(e)
                showToast("崩溃已记录")
            }
        }

        // 跳转到事件测试页面
        binding.btnEventTest.setOnClickListener {
            startActivity(Intent(this, EventTestActivity::class.java))
        }
    }

    private fun showToast(message: String) {
        Toast.makeText(this, message, Toast.LENGTH_SHORT).show()
    }

    override fun onStart() {
        super.onStart()
        Countly.sharedInstance().onStart(this)
    }

    override fun onStop() {
        Countly.sharedInstance().onStop()
        super.onStop()
    }
}
```


### 2. 事件测试页面（EventTestActivity）

创建`activity_event_test.xml`：

```xml
<?xml version="1.0" encoding="utf-8"?>
<ScrollView xmlns:android="http://schemas.android.com/apk/res/android"
    xmlns:app="http://schemas.android.com/apk/res-auto"
    android:layout_width="match_parent"
    android:layout_height="match_parent"
    android:padding="16dp">

    <androidx.constraintlayout.widget.ConstraintLayout
        android:layout_width="match_parent"
        android:layout_height="wrap_content">

        <TextView
            android:id="@+id/tvTitle"
            android:layout_width="wrap_content"
            android:layout_height="wrap_content"
            android:text="事件测试"
            android:textSize="20sp"
            android:textStyle="bold"
            app:layout_constraintEnd_toEndOf="parent"
            app:layout_constraintStart_toStartOf="parent"
            app:layout_constraintTop_toTopOf="parent" />

        <Button
            android:id="@+id/btnPurchaseEvent"
            android:layout_width="0dp"
            android:layout_height="wrap_content"
            android:text="模拟购买事件"
            app:layout_constraintEnd_toEndOf="parent"
            app:layout_constraintStart_toStartOf="parent"
            app:layout_constraintTop_toBottomOf="@id/tvTitle"
            android:layout_marginTop="24dp" />

        <Button
            android:id="@+id/btnLoginEvent"
            android:layout_width="0dp"
            android:layout_height="wrap_content"
            android:text="模拟登录事件"
            app:layout_constraintEnd_toEndOf="parent"
            app:layout_constraintStart_toStartOf="parent"
            app:layout_constraintTop_toBottomOf="@id/btnPurchaseEvent"
            android:layout_marginTop="16dp" />

        <Button
            android:id="@+id/btnShareEvent"
            android:layout_width="0dp"
            android:layout_height="wrap_content"
            android:text="模拟分享事件"
            app:layout_constraintEnd_toEndOf="parent"
            app:layout_constraintStart_toStartOf="parent"
            app:layout_constraintTop_toBottomOf="@id/btnLoginEvent"
            android:layout_marginTop="16dp" />

        <Button
            android:id="@+id/btnTimedEvent"
            android:layout_width="0dp"
            android:layout_height="wrap_content"
            android:text="开始计时事件"
            app:layout_constraintEnd_toEndOf="parent"
            app:layout_constraintStart_toStartOf="parent"
            app:layout_constraintTop_toBottomOf="@id/btnShareEvent"
            android:layout_marginTop="16dp" />

        <Button
            android:id="@+id/btnEndTimedEvent"
            android:layout_width="0dp"
            android:layout_height="wrap_content"
            android:text="结束计时事件"
            android:enabled="false"
            app:layout_constraintEnd_toEndOf="parent"
            app:layout_constraintStart_toStartOf="parent"
            app:layout_constraintTop_toBottomOf="@id/btnTimedEvent"
            android:layout_marginTop="16dp" />

        <TextView
            android:id="@+id/tvEventLog"
            android:layout_width="0dp"
            android:layout_height="wrap_content"
            android:text="事件日志："
            android:textSize="16sp"
            android:layout_marginTop="24dp"
            app:layout_constraintEnd_toEndOf="parent"
            app:layout_constraintStart_toStartOf="parent"
            app:layout_constraintTop_toBottomOf="@id/btnEndTimedEvent" />

        <TextView
            android:id="@+id/tvLog"
            android:layout_width="0dp"
            android:layout_height="wrap_content"
            android:text=""
            android:textSize="14sp"
            android:layout_marginTop="8dp"
            android:background="@android:color/darker_gray"
            android:padding="8dp"
            app:layout_constraintEnd_toEndOf="parent"
            app:layout_constraintStart_toStartOf="parent"
            app:layout_constraintTop_toBottomOf="@id/tvEventLog" />

    </androidx.constraintlayout.widget.ConstraintLayout>
</ScrollView>
```

创建`EventTestActivity.kt`：

```kotlin
package com.example.countlydemo

import android.os.Bundle
import androidx.appcompat.app.AppCompatActivity
import com.example.countlydemo.databinding.ActivityEventTestBinding
import ly.count.android.sdk.Countly
import java.text.SimpleDateFormat
import java.util.*

class EventTestActivity : AppCompatActivity() {

    private lateinit var binding: ActivityEventTestBinding
    private val logBuilder = StringBuilder()
    private var timedEventStarted = false

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        binding = ActivityEventTestBinding.inflate(layoutInflater)
        setContentView(binding.root)

        // 记录屏幕浏览
        Countly.sharedInstance().views().startAutoStoppedView("EventTestActivity")

        setupListeners()
    }

    private fun setupListeners() {
        // 购买事件
        binding.btnPurchaseEvent.setOnClickListener {
            val segmentation = HashMap<String, Any>()
            segmentation["product_id"] = "prod_123"
            segmentation["product_name"] = "Premium Plan"
            segmentation["price"] = 99.99
            segmentation["currency"] = "USD"
            segmentation["payment_method"] = "credit_card"

            Countly.sharedInstance().events().recordEvent(
                "purchase_completed",
                segmentation,
                1,
                99.99
            )

            addLog("购买事件已发送: Premium Plan - $99.99")
        }

        // 登录事件
        binding.btnLoginEvent.setOnClickListener {
            val segmentation = HashMap<String, Any>()
            segmentation["method"] =

### 2. 事件测试页面（EventTestActivity）

创建`activity_event_test.xml`：

```xml
<?xml version="1.0" encoding="utf-8"?>
<ScrollView xmlns:android="http://schemas.android.com/apk/res/android"
    xmlns:app="http://schemas.android.com/apk/res-auto"
    android:layout_width="match_parent"
    android:layout_height="match_parent">

    <androidx.constraintlayout.widget.ConstraintLayout
        android:layout_width="match_parent"
        android:layout_height="wrap_content"
        android:padding="16dp">

        <TextView
            android:id="@+id/tvTitle"
            android:layout_width="wrap_content"
            android:layout_height="wrap_content"
            android:text="事件测试"
            android:textSize="20sp"
            android:textStyle="bold"
            app:layout_constraintEnd_toEndOf="parent"
            app:layout_constraintStart_toStartOf="parent"
            app:layout_constraintTop_toTopOf="parent" />

        <Button
            android:id="@+id/btnPurchase"
            android:layout_width="0dp"
            android:layout_height="wrap_content"
            android:text="模拟购买事件"
            app:layout_constraintEnd_toEndOf="parent"
            app:layout_constraintStart_toStartOf="parent"
            app:layout_constraintTop_toBottomOf="@id/tvTitle"
            android:layout_marginTop="24dp" />

        <Button
            android:id="@+id/btnLogin"
            android:layout_width="0dp"
            android:layout_height="wrap_content"
            android:text="模拟登录事件"
            app:layout_constraintEnd_toEndOf="parent"
            app:layout_constraintStart_toStartOf="parent"
            app:layout_constraintTop_toBottomOf="@id/btnPurchase"
            android:layout_marginTop="16dp" />

        <Button
            android:id="@+id/btnShare"
            android:layout_width="0dp"
            android:layout_height="wrap_content"
            android:text="模拟分享事件"
            app:layout_constraintEnd_toEndOf="parent"
            app:layout_constraintStart_toStartOf="parent"
            app:layout_constraintTop_toBottomOf="@id/btnLogin"
            android:layout_marginTop="16dp" />

        <Button
            android:id="@+id/btnTimedEvent"
            android:layout_width="0dp"
            android:layout_height="wrap_content"
            android:text="开始计时事件"
            app:layout_constraintEnd_toEndOf="parent"
            app:layout_constraintStart_toStartOf="parent"
            app:layout_constraintTop_toBottomOf="@id/btnShare"
            android:layout_marginTop="16dp" />

        <Button
            android:id="@+id/btnEndTimedEvent"
            android:layout_width="0dp"
            android:layout_height="wrap_content"
            android:text="结束计时事件"
            android:enabled="false"
            app:layout_constraintEnd_toEndOf="parent"
            app:layout_constraintStart_toStartOf="parent"
            app:layout_constraintTop_toBottomOf="@id/btnTimedEvent"
            android:layout_marginTop="16dp" />

        <TextView
            android:id="@+id/tvEventLog"
            android:layout_width="0dp"
            android:layout_height="wrap_content"
            android:text="事件日志："
            android:textSize="16sp"
            android:layout_marginTop="24dp"
            app:layout_constraintEnd_toEndOf="parent"
            app:layout_constraintStart_toStartOf="parent"
            app:layout_constraintTop_toBottomOf="@id/btnEndTimedEvent" />

        <TextView
            android:id="@+id/tvLog"
            android:layout_width="0dp"
            android:layout_height="wrap_content"
            android:text=""
            android:textSize="14sp"
            android:layout_marginTop="8dp"
            android:background="@android:color/darker_gray"
            android:padding="8dp"
            app:layout_constraintEnd_toEndOf="parent"
            app:layout_constraintStart_toStartOf="parent"
            app:layout_constraintTop_toBottomOf="@id/tvEventLog" />

    </androidx.constraintlayout.widget.ConstraintLayout>
</ScrollView>
```

创建`EventTestActivity.kt`：

```kotlin
package com.example.countlydemo

import android.os.Bundle
import android.widget.Toast
import androidx.appcompat.app.AppCompatActivity
import com.example.countlydemo.databinding.ActivityEventTestBinding
import ly.count.android.sdk.Countly
import java.text.SimpleDateFormat
import java.util.*

class EventTestActivity : AppCompatActivity() {

    private lateinit var binding: ActivityEventTestBinding
    private val logBuilder = StringBuilder()
    private val dateFormat = SimpleDateFormat("HH:mm:ss", Locale.getDefault())

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        binding = ActivityEventTestBinding.inflate(layoutInflater)
        setContentView(binding.root)

        // 记录屏幕浏览
        Countly.sharedInstance().views().startAutoStoppedView("EventTestActivity")

        setupListeners()
    }

    private fun setupListeners() {
        // 购买事件
        binding.btnPurchase.setOnClickListener {
            val segmentation = HashMap<String, Any>()
            segmentation["product_id"] = "prod_123"
            segmentation["product_name"] = "Premium Plan"
            segmentation["price"] = 99.99
            segmentation["currency"] = "USD"
            segmentation["payment_method"] = "credit_card"

            Countly.sharedInstance().events().recordEvent(
                "purchase_completed",
                segmentation,
                1,
                99.99
            )

            addLog("购买事件已发送: Premium Plan $99.99")
            showToast("购买事件已发送")
        }

        // 登录事件
        binding.btnLogin.setOnClickListener {
            val segmentation = HashMap<String, Any>()
            segmentation["method"] = "email"
            segmentation["success"] = true
            segmentation["user_type"] = "premium"

            Countly.sharedInstance().events().recordEvent(
                "user_login",
                segmentation,
                1
            )

            addLog("登录事件已发送: email登录成功")
            showToast("登录事件已发送")
        }

        // 分享事件
        binding.btnShare.setOnClickListener {
            val segmentation = HashMap<String, Any>()
            segmentation["platform"] = "twitter"
            segmentation["content_type"] = "article"
            segmentation["content_id"] = "article_456"

            Countly.sharedInstance().events().recordEvent(
                "content_shared",
                segmentation,
                1
            )

            addLog("分享事件已发送: Twitter分享文章")
            showToast("分享事件已发送")
        }

        // 开始计时事件
        binding.btnTimedEvent.setOnClickListener {
            Countly.sharedInstance().events().startEvent("video_watch")

            binding.btnTimedEvent.isEnabled = false
            binding.btnEndTimedEvent.isEnabled = true

            addLog("开始计时事件: video_watch")
            showToast("开始计时")
        }

        // 结束计时事件
        binding.btnEndTimedEvent.setOnClickListener {
            val segmentation = HashMap<String, Any>()
            segmentation["video_id"] = "video_789"
            segmentation["video_title"] = "Tutorial Video"

            Countly.sharedInstance().events().endEvent(
                "video_watch",
                segmentation,
                1,
                0.0
            )

            binding.btnTimedEvent.isEnabled = true
            binding.btnEndTimedEvent.isEnabled = false

            addLog("结束计时事件: video_watch")
            showToast("计时结束")
        }
    }

    private fun addLog(message: String) {
        val timestamp = dateFormat.format(Date())
        logBuilder.append("[$timestamp] $message\n")
        binding.tvLog.text = logBuilder.toString()
    }

    private fun showToast(message: String) {
        Toast.makeText(this, message, Toast.LENGTH_SHORT).show()
    }

    override fun onStart() {
        super.onStart()
        Countly.sharedInstance().onStart(this)
    }

    override fun onStop() {
        Countly.sharedInstance().onStop()
        super.onStop()
    }
}
```

### 3. 用户属性页面（UserProfileActivity）

创建`activity_user_profile.xml`：

```xml
<?xml version="1.0" encoding="utf-8"?>
<ScrollView xmlns:android="http://schemas.android.com/apk/res/android"
    xmlns:app="http://schemas.android.com/apk/res-auto"
    android:layout_width="match_parent"
    android:layout_height="match_parent">

    <androidx.constraintlayout.widget.ConstraintLayout
        android:layout_width="match_parent"
        android:layout_height="wrap_content"
        android:padding="16dp">

        <TextView
            android:id="@+id/tvTitle"
            android:layout_width="wrap_content"
            android:layout_height="wrap_content"
            android:text="用户属性设置"
            android:textSize="20sp"
            android:textStyle="bold"
            app:layout_constraintEnd_toEndOf="parent"
            app:layout_constraintStart_toStartOf="parent"
            app:layout_constraintTop_toTopOf="parent" />

        <com.google.android.material.textfield.TextInputLayout
            android:id="@+id/tilName"
            android:layout_width="0dp"
            android:layout_height="wrap_content"
            android:hint="姓名"
            app:layout_constraintEnd_toEndOf="parent"
            app:layout_constraintStart_toStartOf="parent"
            app:layout_constraintTop_toBottomOf="@id/tvTitle"
            android:layout_marginTop="24dp">

            <com.google.android.material.textfield.TextInputEditText
                android:id="@+id/etName"
                android:layout_width="match_parent"
                android:layout_height="wrap_content" />
        </com.google.android.material.textfield.TextInputLayout>

        <com.google.android.material.textfield.TextInputLayout
            android:id="@+id/tilEmail"
            android:layout_width="0dp"
            android:layout_height="wrap_content"
            android:hint="邮箱"
            app:layout_constraintEnd_toEndOf="parent"
            app:layout_constraintStart_toStartOf="parent"
            app:layout_constraintTop_toBottomOf="@id/tilName"
            android:layout_marginTop="16dp">

            <com.google.android.material.textfield.TextInputEditText
                android:id="@+id/etEmail"
                android:layout_width="match_parent"
                android:layout_height="wrap_content"
                android:inputType="textEmailAddress" />
        </com.google.android.material.textfield.TextInputLayout>

        <com.google.android.material.textfield.TextInputLayout
            android:id="@+id/tilAge"
            android:layout_width="0dp"
            android:layout_height="wrap_content"
            android:hint="年龄"
            app:layout_constraintEnd_toEndOf="parent"
            app:layout_constraintStart_toStartOf="parent"
            app:layout_constraintTop_toBottomOf="@id/tilEmail"
            android:layout_marginTop="16dp">

            <com.google.android.material.textfield.TextInputEditText
                android:id="@+id/etAge"
                android:layout_width="match_parent"
                android:layout_height="wrap_content"
                android:inputType="number" />
        </com.google.android.material.textfield.TextInputLayout>

        <Button
            android:id="@+id/btnSaveProfile"
            android:layout_width="0dp"
            android:layout_height="wrap_content"
            android:text="保存用户属性"
            app:layout_constraintEnd_toEndOf="parent"
            app:layout_constraintStart_toStartOf="parent"
            app:layout_constraintTop_toBottomOf="@id/tilAge"
            android:layout_marginTop="24dp" />

        <Button
            android:id="@+id/btnSetCustomProperty"
            android:layout_width="0dp"
            android:layout_height="wrap_content"
            android:text="设置自定义属性"
            app:layout_constraintEnd_toEndOf="parent"
            app:layout_constraintStart_toStartOf="parent"
            app:layout_constraintTop_toBottomOf="@id/btnSaveProfile"
            android:layout_marginTop="16dp" />

    </androidx.constraintlayout.widget.ConstraintLayout>
</ScrollView>
```

创建`UserProfileActivity.kt`：

```kotlin
package com.example.countlydemo

import android.os.Bundle
import android.widget.Toast
import androidx.appcompat.app.AppCompatActivity
import com.example.countlydemo.databinding.ActivityUserProfileBinding
import ly.count.android.sdk.Countly

class UserProfileActivity : AppCompatActivity() {

    private lateinit var binding: ActivityUserProfileBinding

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        binding = ActivityUserProfileBinding.inflate(layoutInflater)
        setContentView(binding.root)

        // 记录屏幕浏览
        Countly.sharedInstance().views().startAutoStoppedView("UserProfileActivity")

        setupListeners()
    }

    private fun setupListeners() {
        // 保存用户属性
        binding.btnSaveProfile.setOnClickListener {
            val name = binding.etName.text.toString()
            val email = binding.etEmail.text.toString()
            val age = binding.etAge.text.toString().toIntOrNull()

            if (name.isNotEmpty() && email.isNotEmpty() && age != null) {
                val userProfile = HashMap<String, Any>()
                userProfile["name"] = name
                userProfile["email"] = email
                userProfile["byear"] = 2024 - age  // 出生年份

                Countly.sharedInstance().userProfile().setProperties(userProfile)
                Countly.sharedInstance().userProfile().save()

                showToast("用户属性已保存")
            } else {
                showToast("请填写完整信息")
            }
        }

        // 设置自定义属性
        binding.btnSetCustomProperty.setOnClickListener {
            val customProperties = HashMap<String, Any>()
            customProperties["subscription_type"] = "premium"
            customProperties["registration_date"] = System.currentTimeMillis()
            customProperties["favorite_category"] = "technology"
            customProperties["notification_enabled"] = true

            Countly.sharedInstance().userProfile().setProperties(customProperties)
            Countly.sharedInstance().userProfile().save()

            showToast("自定义属性已设置")
        }
    }

    private fun showToast(message: String) {
        Toast.makeText(this, message, Toast.LENGTH_SHORT).show()
    }

    override fun onStart() {
        super.onStart()
        Countly.sharedInstance().onStart(this)
    }

    override fun onStop() {
        Countly.sharedInstance().onStop()
        super.onStop()
    }
}
```

---

## 🧪 测试验证

### 步骤1：运行应用

```bash
1. 连接Android设备或启动模拟器
2. 在Android Studio中点击Run按钮
3. 等待应用安装并启动
```

### 步骤2：测试基础功能

**测试清单**：

```
✅ 应用启动
   - 打开应用
   - 检查Logcat是否有Countly初始化日志

✅ 简单事件
   - 点击"发送简单事件"按钮
   - 查看Toast提示

✅ 带参数事件
   - 点击"发送带参数事件"按钮
   - 查看Toast提示

✅ 用户属性
   - 点击"设置用户属性"按钮
   - 查看Toast提示

✅ 崩溃报告
   - 点击"模拟崩溃"按钮
   - 查看Toast提示

✅ 屏幕浏览
   - 点击"事件测试页面"按钮
   - 跳转到新页面
```

### 步骤3：在Countly管理界面验证

**登录Countly**：
```
URL: http://YOUR_IP:8080
用户名: 你创建的管理员账号
密码: 你设置的密码
```

**验证数据**：

1. **实时数据**
   ```
   Dashboard > Real-time
   - 查看实时用户数
   - 查看实时事件
   ```

2. **事件数据**
   ```
   Analytics > Events
   - 查看事件列表
   - 查看事件详情
   - 查看事件参数
   ```

3. **用户数据**
   ```
   Analytics > Users
   - 查看用户列表
   - 查看用户属性
   - 查看用户会话
   ```

4. **崩溃报告**
   ```
   Crashes > Overview
   - 查看崩溃列表
   - 查看崩溃详情
   - 查看堆栈信息
   ```

5. **屏幕浏览**
   ```
   Analytics > Views
   - 查看页面浏览量
   - 查看页面停留时间
   ```

### 步骤4：查看Logcat日志

在Android Studio的Logcat中过滤Countly日志：

```
过滤器: Countly

预期日志：
[Countly] Initializing...
[Countly] SDK initialized
[Countly] Recording event: button_clicked
[Countly] Sending request...
[Countly] Request completed successfully
```

---

## 🔍 常见问题

### 1. 无法连接到Countly服务器

**问题**：应用无法发送数据到Countly

**排查步骤**：

```kotlin
// 1. 检查网络权限
// AndroidManifest.xml中是否有：
<uses-permission android:name="android.permission.INTERNET" />

// 2. 检查服务器地址
// CountlyDemoApp.kt中的URL是否正确
"http://YOUR_IP:8080"  // 注意：使用http而非https

// 3. 检查App Key
// 是否从Countly管理界面正确复制

// 4. 检查网络连接
// 模拟器/设备是否能访问服务器
```

**解决方案**：

```bash
# 测试服务器连接
adb shell ping YOUR_IP

# 或在应用中添加测试代码
val url = URL("http://YOUR_IP:8080/o/ping")
val connection = url.openConnection() as HttpURLConnection
val response = connection.inputStream.bufferedReader().readText()
Log.d("Countly", "Server response: $response")
```

### 2. 数据未显示在Countly管理界面

**问题**：发送了事件但管理界面看不到

**可能原因**：

```
1. 数据延迟
   - Countly有数据处理延迟（通常几秒到几分钟）
   - 等待1-2分钟后刷新页面

2. 应用未正确初始化
   - 检查Application类是否在AndroidManifest中注册
   - 检查初始化代码是否执行

3. 会话未开始
   - 确保调用了onStart()和onStop()

4. 选择了错误的应用
   - 在Countly管理界面确认选择了正确的应用
```

### 3. 崩溃报告未记录

**问题**：模拟崩溃后看不到报告

**解决方案**：

```kotlin
// 确保启用了崩溃报告
val config = CountlyConfig(this, appKey, serverUrl).apply {
    enableCrashReporting()  // 必须启用
}

// 记录异常
try {
    throw RuntimeException("Test crash")
} catch (e: Exception) {
    Countly.sharedInstance().crashes().recordHandledException(e)
}

// 注意：崩溃数据可能需要几分钟才能显示
```

### 4. 模拟器无法访问localhost

**问题**：使用localhost无法连接

**解决方案**：

```kotlin
// 不要使用localhost，使用以下地址：

// 方案1：使用10.0.2.2（Android模拟器特殊地址）
"http://10.0.2.2:8080"

// 方案2：使用局域网IP
"http://192.168.1.100:8080"  // 替换为你的实际IP

// 方案3：使用公网IP或域名
"http://YOUR_PUBLIC_IP:8080"
```

### 5. ProGuard混淆问题

**问题**：Release版本无法正常工作

**解决方案**：

在`proguard-rules.pro`中添加：

```proguard
# Countly SDK
-keep class ly.count.android.sdk.** { *; }
-dontwarn ly.count.android.sdk.**
```

---

## 📦 构建APK

### Debug版本

```bash
# 在Android Studio中
Build > Build Bundle(s) / APK(s) > Build APK(s)

# 或使用命令行
./gradlew assembleDebug

# APK位置
app/build/outputs/apk/debug/app-debug.apk
```

### Release版本

```bash
# 1. 生成签名密钥（首次）
keytool -genkey -v -keystore countly-demo.jks -keyalg RSA -keysize 2048 -validity 10000 -alias countly-demo

# 2. 配置签名
# 在app/build.gradle.kts中添加：
android {
    signingConfigs {
        create("release") {
            storeFile = file("countly-demo.jks")
            storePassword = "your_password"
            keyAlias = "countly-demo"
            keyPassword = "your_password"
        }
    }
    buildTypes {
        release {
            signingConfig = signingConfigs.getByName("release")
        }
    }
}

# 3. 构建Release APK
./gradlew assembleRelease

# APK位置
app/build/outputs/apk/release/app-release.apk
```

---

## 🎯 下一步

### 扩展功能

1. **推送通知**
   - 集成Firebase Cloud Messaging
   - 配置Countly推送通知
   - 测试推送接收和点击

2. **A/B测试**
   - 使用远程配置
   - 实现功能开关
   - 测试不同变体

3. **用户反馈**
   - 添加反馈表单
   - 集成NPS评分
   - 收集用户意见

4. **深度链接**
   - 配置Deep Links
   - 追踪链接来源
   - 分析转化率

### 性能优化

```kotlin
// 1. 批量发送事件
val config = CountlyConfig(this, appKey, serverUrl).apply {
    setEventQueueSizeToSend(10)  // 累积10个事件后发送
}

// 2. 减少网络请求
val config = CountlyConfig(this, appKey, serverUrl).apply {
    setUpdateSessionTimerDelay(60)  // 60秒更新一次会话
}

// 3. 离线模式
// Countly SDK自动处理离线情况，会缓存数据并在联网后发送
```

---

## 📚 参考资源

### 官方文档

- **Android SDK文档**：https://support.count.ly/hc/en-us/articles/360037754031-Android
- **API参考**：https://countly.github.io/countly-sdk-android/
- **GitHub仓库**：https://github.com/Countly/countly-sdk-android

### 示例代码

- **官方Demo**：https://github.com/Countly/countly-sdk-android/tree/master/app
- **集成示例**：https://github.com/Countly/countly-sdk-android/wiki

### 学习资源

- **视频教程**：https://www.youtube.com/c/CountlyAnalytics
- **博客文章**：https://countly.com/blog
- **社区论坛**：https://community.count.ly

---

## ✅ 完成检查清单

```
项目创建：
- [ ] 创建Android项目
- [ ] 配置Gradle依赖
- [ ] 添加必要权限

SDK集成：
- [ ] 创建Application类
- [ ] 初始化Countly SDK
- [ ] 配置服务器地址和App Key

功能实现：
- [ ] 主界面（MainActivity）
- [ ] 事件测试页面（EventTestActivity）
- [ ] 用户属性页面（UserProfileActivity）

测试验证：
- [ ] 运行应用
- [ ] 测试所有功能
- [ ] 在Countly管理界面验证数据
- [ ] 查看Logcat日志

构建发布：
- [ ] 生成Debug APK
- [ ] 配置签名
- [ ] 生成Release APK
```

---

**恭喜！你已经完成了Countly Android SDK的集成和测试！**

现在你可以：
1. 在真实应用中集成Countly
2. 追踪用户行为和事件
3. 分析应用数据
4. 优化用户体验

---

*文档版本：v1.0*
*最后更新：2026年1月28日*
*作者：Kiro AI*
