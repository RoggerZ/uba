# Countly Flutter 集成测试指南

> 完整的Flutter Demo应用集成Countly SDK教程
> 日期：2026年1月28日

---

## 📋 目录

1. [项目概述](#项目概述)
2. [开发环境准备](#开发环境准备)
3. [创建Flutter项目](#创建flutter项目)
4. [集成Countly SDK](#集成countly-sdk)
5. [核心功能实现](#核心功能实现)
6. [iOS配置](#ios配置)
7. [Android配置](#android配置)
8. [测试验证](#测试验证)
9. [常见问题](#常见问题)

---

## 🎯 项目概述

### Demo应用功能

创建一个跨平台Flutter应用来测试Countly的核心功能：

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
   - 定时事件

3. 用户属性
   - 设置用户信息
   - 自定义属性
   - 用户画像

4. 崩溃报告
   - 自动崩溃捕获
   - 手动异常记录

5. 视图追踪
   - 自动视图追踪
   - 手动视图追踪

6. 设备信息
   - 自动收集设备信息
   - 位置信息（可选）
```

### 技术栈

```
框架：Flutter 3.16+
语言：Dart 3.2+
Countly SDK：countly_flutter 24.1.0+
支持平台：iOS 12+, Android 5.0+ (API 21)
IDE：VS Code / Android Studio / IntelliJ IDEA
```

---

## 💻 开发环境准备

### 1. 安装Flutter

**Windows**：
```bash
# 下载Flutter SDK
# https://docs.flutter.dev/get-started/install/windows

# 解压到目录（如 C:\flutter）
# 添加到环境变量 PATH: C:\flutter\bin

# 验证安装
flutter doctor
```

**macOS**：
```bash
# 使用Homebrew安装
brew install flutter

# 或手动下载
# https://docs.flutter.dev/get-started/install/macos

# 验证安装
flutter doctor
```

**Linux**：
```bash
# 下载Flutter SDK
wget https://storage.googleapis.com/flutter_infra_release/releases/stable/linux/flutter_linux_3.16.0-stable.tar.xz

# 解压
tar xf flutter_linux_3.16.0-stable.tar.xz

# 添加到PATH
export PATH="$PATH:`pwd`/flutter/bin"

# 验证安装
flutter doctor
```

### 2. 安装开发工具

**VS Code**（推荐）：
```
1. 下载安装VS Code
2. 安装Flutter插件
3. 安装Dart插件
```

**Android Studio**：
```
1. 下载安装Android Studio
2. 安装Flutter插件
3. 安装Dart插件
```

### 3. 配置开发环境

```bash
# 检查环境
flutter doctor

# 应该看到：
# ✓ Flutter (Channel stable, 3.16.0)
# ✓ Android toolchain
# ✓ Xcode (macOS only)
# ✓ VS Code / Android Studio
# ✓ Connected device
```

### 4. 准备Countly服务器信息

```
服务器地址：http://YOUR_IP:8080
App Key：在Countly管理界面获取
```

---

## 📱 创建Flutter项目

### 步骤1：创建新项目

```bash
# 创建Flutter项目
flutter create countly_flutter_demo

# 进入项目目录
cd countly_flutter_demo

# 运行项目（测试环境）
flutter run
```

### 步骤2：项目结构

```
countly_flutter_demo/
├── lib/
│   ├── main.dart
│   ├── screens/
│   │   ├── home_screen.dart
│   │   ├── event_test_screen.dart
│   │   ├── user_profile_screen.dart
│   │   └── crash_test_screen.dart
│   ├── services/
│   │   └── countly_service.dart
│   └── widgets/
│       └── custom_button.dart
├── android/
│   └── app/
│       └── src/
│           └── main/
│               └── AndroidManifest.xml
├── ios/
│   └── Runner/
│       └── Info.plist
├── pubspec.yaml
└── README.md
```

---

## 🔧 集成Countly SDK

### 步骤1：添加依赖

编辑`pubspec.yaml`：

```yaml
name: countly_flutter_demo
description: Countly Flutter Demo Application
publish_to: 'none'
version: 1.0.0+1

environment:
  sdk: '>=3.2.0 <4.0.0'

dependencies:
  flutter:
    sdk: flutter

  # Countly SDK
  countly_flutter: ^24.1.0

  # UI组件
  cupertino_icons: ^1.0.6

dev_dependencies:
  flutter_test:
    sdk: flutter
  flutter_lints: ^3.0.0

flutter:
  uses-material-design: true
```

### 步骤2：安装依赖

```bash
# 获取依赖
flutter pub get

# 验证安装
flutter pub deps
```

### 步骤3：创建Countly服务类

创建`lib/services/countly_service.dart`：

```dart
import 'package:countly_flutter/countly_flutter.dart';

class CountlyService {
  static final CountlyService _instance = CountlyService._internal();
  factory CountlyService() => _instance;
  CountlyService._internal();

  // Countly配置
  static const String serverUrl = 'http://YOUR_IP:8080'; // 替换为你的服务器地址
  static const String appKey = 'YOUR_APP_KEY'; // 替换为你的App Key

  // 初始化Countly
  Future<void> init() async {
    // Countly配置
    CountlyConfig config = CountlyConfig(serverUrl, appKey);

    // 启用日志（开发环境）
    config.setLoggingEnabled(true);

    // 启用崩溃报告
    config.enableCrashReporting();

    // 启用自动视图追踪
    config.enableAutomaticViewTracking();

    // 设置设备ID
    config.setDeviceId('flutter_demo_device');

    // 初始化
    await Countly.initWithConfig(config);

    // 开始会话
    await Countly.start();

    print('Countly initialized successfully');
  }

  // 记录简单事件
  Future<void> recordEvent(String eventName) async {
    await Countly.recordEvent({
      'key': eventName,
      'count': 1,
    });
  }

  // 记录带参数的事件
  Future<void> recordEventWithSegmentation(
    String eventName,
    Map<String, Object> segmentation,
  ) async {
    await Countly.recordEvent({
      'key': eventName,
      'count': 1,
      'segmentation': segmentation,
    });
  }

  // 记录定时事件
  Future<void> startEvent(String eventName) async {
    await Countly.startEvent(eventName);
  }

  Future<void> endEvent(String eventName, {Map<String, Object>? segmentation}) async {
    Map<String, Object> event = {
      'key': eventName,
    };
    if (segmentation != null) {
      event['segmentation'] = segmentation;
    }
    await Countly.endEvent(event);
  }

  // 设置用户属性
  Future<void> setUserProfile({
    String? name,
    String? email,
    String? username,
    String? phone,
    String? gender,
    int? birthYear,
  }) async {
    Map<String, Object> userProfile = {};

    if (name != null) userProfile['name'] = name;
    if (email != null) userProfile['email'] = email;
    if (username != null) userProfile['username'] = username;
    if (phone != null) userProfile['phone'] = phone;
    if (gender != null) userProfile['gender'] = gender;
    if (birthYear != null) userProfile['byear'] = birthYear;

    await Countly.setUserData(userProfile);
  }

  // 设置自定义用户属性
  Future<void> setCustomUserProperty(String key, dynamic value) async {
    await Countly.setProperty(key, value.toString());
  }

  // 记录视图
  Future<void> recordView(String viewName) async {
    await Countly.recordView(viewName);
  }

  // 记录异常
  Future<void> recordException(
    String error,
    String stackTrace, {
    bool fatal = false,
  }) async {
    await Countly.logException(error, fatal, {
      'stackTrace': stackTrace,
    });
  }

  // 添加面包屑
  Future<void> addBreadcrumb(String breadcrumb) async {
    await Countly.addCrashLog(breadcrumb);
  }

  // 停止会话
  Future<void> stop() async {
    await Countly.stop();
  }
}
```

**重要提示**：
```dart
// 必须替换的配置
serverUrl: 'http://YOUR_IP:8080'  // 你的Countly服务器地址
appKey: 'YOUR_APP_KEY'  // 从Countly管理界面获取
```

---

## 🎨 核心功能实现

### 1. 主应用入口（main.dart）

```dart
import 'package:flutter/material.dart';
import 'package:countly_flutter_demo/services/countly_service.dart';
import 'package:countly_flutter_demo/screens/home_screen.dart';

void main() async {
  WidgetsFlutterBinding.ensureInitialized();

  // 初始化Countly
  await CountlyService().init();

  runApp(const MyApp());
}

class MyApp extends StatelessWidget {
  const MyApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Countly Flutter Demo',
      theme: ThemeData(
        colorScheme: ColorScheme.fromSeed(seedColor: Colors.blue),
        useMaterial3: true,
      ),
      home: const HomeScreen(),
    );
  }
}
```

### 2. 主界面（home_screen.dart）

创建`lib/screens/home_screen.dart`：

```dart
import 'package:flutter/material.dart';
import 'package:countly_flutter_demo/services/countly_service.dart';
import 'package:countly_flutter_demo/screens/event_test_screen.dart';
import 'package:countly_flutter_demo/screens/user_profile_screen.dart';
import 'package:countly_flutter_demo/screens/crash_test_screen.dart';

class HomeScreen extends StatefulWidget {
  const HomeScreen({super.key});

  @override
  State<HomeScreen> createState() => _HomeScreenState();
}

class _HomeScreenState extends State<HomeScreen> {
  final CountlyService _countly = CountlyService();

  @override
  void initState() {
    super.initState();
    // 记录视图
    _countly.recordView('HomeScreen');
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Countly Flutter Demo'),
        backgroundColor: Theme.of(context).colorScheme.inversePrimary,
      ),
      body: Center(
        child: SingleChildScrollView(
          padding: const EdgeInsets.all(16.0),
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              const Text(
                'Countly 功能测试',
                style: TextStyle(
                  fontSize: 24,
                  fontWeight: FontWeight.bold,
                ),
              ),
              const SizedBox(height: 32),

              // 简单事件
              _buildButton(
                context,
                '发送简单事件',
                Icons.send,
                () async {
                  await _countly.recordEvent('button_clicked');
                  _showSnackBar('简单事件已发送');
                },
              ),

              // 带参数事件
              _buildButton(
                context,
                '发送带参数事件',
                Icons.settings,
                () async {
                  await _countly.recordEventWithSegmentation(
                    'button_clicked_with_params',
                    {
                      'button_name': 'event_with_params',
                      'screen': 'HomeScreen',
                      'timestamp': DateTime.now().millisecondsSinceEpoch,
                    },
                  );
                  _showSnackBar('带参数事件已发送');
                },
              ),

              // 定时事件
              _buildButton(
                context,
                '测试定时事件',
                Icons.timer,
                () async {
                  await _countly.startEvent('timed_event');
                  _showSnackBar('定时事件已开始，3秒后结束');

                  await Future.delayed(const Duration(seconds: 3));

                  await _countly.endEvent('timed_event', segmentation: {
                    'duration': '3_seconds',
                  });
                  _showSnackBar('定时事件已结束');
                },
              ),

              // 事件测试页面
              _buildButton(
                context,
                '事件测试页面',
                Icons.science,
                () {
                  Navigator.push(
                    context,
                    MaterialPageRoute(
                      builder: (context) => const EventTestScreen(),
                    ),
                  );
                },
              ),

              // 用户属性页面
              _buildButton(
                context,
                '用户属性设置',
                Icons.person,
                () {
                  Navigator.push(
                    context,
                    MaterialPageRoute(
                      builder: (context) => const UserProfileScreen(),
                    ),
                  );
                },
              ),

              // 崩溃测试页面
              _buildButton(
                context,
                '崩溃测试',
                Icons.bug_report,
                () {
                  Navigator.push(
                    context,
                    MaterialPageRoute(
                      builder: (context) => const CrashTestScreen(),
                    ),
                  );
                },
                color: Colors.red,
              ),

              const SizedBox(height: 32),
              const Text(
                '点击按钮测试Countly功能\n查看Countly管理界面验证数据',
                textAlign: TextAlign.center,
                style: TextStyle(color: Colors.grey),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildButton(
    BuildContext context,
    String label,
    IconData icon,
    VoidCallback onPressed, {
    Color? color,
  }) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 8.0),
      child: SizedBox(
        width: double.infinity,
        height: 56,
        child: ElevatedButton.icon(
          onPressed: onPressed,
          icon: Icon(icon),
          label: Text(label),
          style: ElevatedButton.styleFrom(
            backgroundColor: color,
            foregroundColor: color != null ? Colors.white : null,
          ),
        ),
      ),
    );
  }

  void _showSnackBar(String message) {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text(message),
        duration: const Duration(seconds: 2),
      ),
    );
  }
}
```

### 3. 事件测试页面（event_test_screen.dart）

创建`lib/screens/event_test_screen.dart`：

```dart
import 'package:flutter/material.dart';
import 'package:countly_flutter_demo/services/countly_service.dart';

class EventTestScreen extends StatefulWidget {
  const EventTestScreen({super.key});

  @override
  State<EventTestScreen> createState() => _EventTestScreenState();
}

class _EventTestScreenState extends State<EventTestScreen> {
  final CountlyService _countly = CountlyService();
  int _counter = 0;
  bool _isTimerRunning = false;

  @override
  void initState() {
    super.initState();
    _countly.recordView('EventTestScreen');
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('事件测试'),
      ),
      body: Padding(
        padding: const EdgeInsets.all(16.0),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            // 计数器
            Card(
              child: Padding(
                padding: const EdgeInsets.all(16.0),
                child: Column(
                  children: [
                    const Text(
                      '点击计数器',
                      style: TextStyle(fontSize: 18),
                    ),
                    const SizedBox(height: 16),
                    Text(
                      '$_counter',
                      style: const TextStyle(
                        fontSize: 48,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                    const SizedBox(height: 16),
                    ElevatedButton(
                      onPressed: _incrementCounter,
                      child: const Text('点击 +1'),
                    ),
                  ],
                ),
              ),
            ),

            const SizedBox(height: 16),

            // 定时事件
            Card(
              child: Padding(
                padding: const EdgeInsets.all(16.0),
                child: Column(
                  children: [
                    const Text(
                      '定时事件测试',
                      style: TextStyle(fontSize: 18),
                    ),
                    const SizedBox(height: 16),
                    ElevatedButton(
                      onPressed: _isTimerRunning ? null : _startTimedEvent,
                      child: Text(_isTimerRunning ? '计时中...' : '开始计时'),
                    ),
                  ],
                ),
              ),
            ),

            const SizedBox(height: 16),

            // 批量事件
            Card(
              child: Padding(
                padding: const EdgeInsets.all(16.0),
                child: Column(
                  children: [
                    const Text(
                      '批量事件测试',
                      style: TextStyle(fontSize: 18),
                    ),
                    const SizedBox(height: 16),
                    ElevatedButton(
                      onPressed: _sendBatchEvents,
                      child: const Text('发送10个事件'),
                    ),
                  ],
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  void _incrementCounter() async {
    setState(() {
      _counter++;
    });

    await _countly.recordEventWithSegmentation(
      'counter_incremented',
      {
        'count': _counter,
        'screen': 'EventTestScreen',
      },
    );
  }

  void _startTimedEvent() async {
    setState(() {
      _isTimerRunning = true;
    });

    await _countly.startEvent('user_action_duration');

    // 模拟用户操作5秒
    await Future.delayed(const Duration(seconds: 5));

    await _countly.endEvent('user_action_duration', segmentation: {
      'action': 'test_action',
      'duration_seconds': 5,
    });

    setState(() {
      _isTimerRunning = false;
    });

    if (mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('定时事件已完成（5秒）')),
      );
    }
  }

  void _sendBatchEvents() async {
    for (int i = 1; i <= 10; i++) {
      await _countly.recordEventWithSegmentation(
        'batch_event',
        {
          'event_number': i,
          'batch_id': DateTime.now().millisecondsSinceEpoch,
        },
      );
      await Future.delayed(const Duration(milliseconds: 100));
    }

    if (mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('已发送10个批量事件')),
      );
    }
  }
}
```

### 4. 用户属性页面（user_profile_screen.dart）

创建`lib/screens/user_profile_screen.dart`：

```dart
import 'package:flutter/material.dart';
import 'package:countly_flutter_demo/services/countly_service.dart';

class UserProfileScreen extends StatefulWidget {
  const UserProfileScreen({super.key});

  @override
  State<UserProfileScreen> createState() => _UserProfileScreenState();
}

class _UserProfileScreenState extends State<UserProfileScreen> {
  final CountlyService _countly = CountlyService();
  final _formKey = GlobalKey<FormState>();

  final _nameController = TextEditingController();
  final _emailController = TextEditingController();
  final _usernameController = TextEditingController();
  final _phoneController = TextEditingController();

  String _selectedGender = 'M';
  int _birthYear = 1990;

  @override
  void initState() {
    super.initState();
    _countly.recordView('UserProfileScreen');
  }

  @override
  void dispose() {
    _nameController.dispose();
    _emailController.dispose();
    _usernameController.dispose();
    _phoneController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('用户属性设置'),
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(16.0),
        child: Form(
          key: _formKey,
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              TextFormField(
                controller: _nameController,
                decoration: const InputDecoration(
                  labelText: '姓名',
                  border: OutlineInputBorder(),
                ),
              ),
              const SizedBox(height: 16),

              TextFormField(
                controller: _emailController,
                decoration: const InputDecoration(
                  labelText: '邮箱',
                  border: OutlineInputBorder(),
                ),
                keyboardType: TextInputType.emailAddress,
              ),
              const SizedBox(height: 16),

              TextFormField(
                controller: _usernameController,
                decoration: const InputDecoration(
                  labelText: '用户名',
                  border: OutlineInputBorder(),
                ),
              ),
              const SizedBox(height: 16),

              TextFormField(
                controller: _phoneController,
                decoration: const InputDecoration(
                  labelText: '电话',
                  border: OutlineInputBorder(),
                ),
                keyboardType: TextInputType.phone,
              ),
              const SizedBox(height: 16),

              DropdownButtonFormField<String>(
                value: _selectedGender,
                decoration: const InputDecoration(
                  labelText: '性别',
                  border: OutlineInputBorder(),
                ),
                items: const [
                  DropdownMenuItem(value: 'M', child: Text('男')),
                  DropdownMenuItem(value: 'F', child: Text('女')),
                ],
                onChanged: (value) {
                  setState(() {
                    _selectedGender = value!;
                  });
                },
              ),
              const SizedBox(height: 16),

              DropdownButtonFormField<int>(
                value: _birthYear,
                decoration: const InputDecoration(
                  labelText: '出生年份',
                  border: OutlineInputBorder(),
                ),
                items: List.generate(
                  60,
                  (index) => DropdownMenuItem(
                    value: 1970 + index,
                    child: Text('${1970 + index}'),
                  ),
                ),
                onChanged: (value) {
                  setState(() {
                    _birthYear = value!;
                  });
                },
              ),
              const SizedBox(height: 24),

              ElevatedButton(
                onPressed: _saveUserProfile,
                style: ElevatedButton.styleFrom(
                  padding: const EdgeInsets.symmetric(vertical: 16),
                ),
                child: const Text('保存用户属性'),
              ),

              const SizedBox(height: 16),

              ElevatedButton(
                onPressed: _setCustomProperties,
                style: ElevatedButton.styleFrom(
                  padding: const EdgeInsets.symmetric(vertical: 16),
                ),
                child: const Text('设置自定义属性'),
              ),
            ],
          ),
        ),
      ),
    );
  }

  void _saveUserProfile() async {
    await _countly.setUserProfile(
      name: _nameController.text.isNotEmpty ? _nameController.text : null,
      email: _emailController.text.isNotEmpty ? _emailController.text : null,
      username: _usernameController.text.isNotEmpty ? _usernameController.text : null,
      phone: _phoneController.text.isNotEmpty ? _phoneController.text : null,
      gender: _selectedGender,
      birthYear: _birthYear,
    );

    if (mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('用户属性已保存')),
      );
    }
  }

  void _setCustomProperties() async {
    await _countly.setCustomUserProperty('app_version', '1.0.0');
    await _countly.setCustomUserProperty('user_level', 'premium');
    await _countly.setCustomUserProperty('last_login', DateTime.now().toIso8601String());

    if (mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('自定义属性已设置')),
      );
    }
  }
}
```


### 5. 崩溃测试页面（crash_test_screen.dart）

创建`lib/screens/crash_test_screen.dart`：

```dart
import 'package:flutter/material.dart';
import 'package:countly_flutter_demo/services/countly_service.dart';

class CrashTestScreen extends StatefulWidget {
  const CrashTestScreen({super.key});

  @override
  State<CrashTestScreen> createState() => _CrashTestScreenState();
}

class _CrashTestScreenState extends State<CrashTestScreen> {
  final CountlyService _countly = CountlyService();

  @override
  void initState() {
    super.initState();
    _countly.recordView('CrashTestScreen');
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('崩溃测试'),
        backgroundColor: Colors.red,
      ),
      body: Padding(
        padding: const EdgeInsets.all(16.0),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            const Card(
              color: Colors.red[50],
              child: Padding(
                padding: EdgeInsets.all(16.0),
                child: Text(
                  '⚠️ 警告：这些按钮会触发异常和崩溃\n仅用于测试Countly的崩溃报告功能',
                  style: TextStyle(
                    color: Colors.red,
                    fontWeight: FontWeight.bold,
                  ),
                  textAlign: TextAlign.center,
                ),
              ),
            ),

            const SizedBox(height: 24),

            ElevatedButton(
              onPressed: _recordHandledException,
              style: ElevatedButton.styleFrom(
                backgroundColor: Colors.orange,
                padding: const EdgeInsets.symmetric(vertical: 16),
              ),
              child: const Text('记录已处理异常'),
            ),

            const SizedBox(height: 16),

            ElevatedButton(
              onPressed: _simulateDivisionByZero,
              style: ElevatedButton.styleFrom(
                backgroundColor: Colors.red,
                padding: const EdgeInsets.symmetric(vertical: 16),
              ),
              child: const Text('模拟除零错误'),
            ),

            const SizedBox(height: 16),

            ElevatedButton(
              onPressed: _simulateNullPointerException,
              style: ElevatedButton.styleFrom(
                backgroundColor: Colors.red,
                padding: const EdgeInsets.symmetric(vertical: 16),
              ),
              child: const Text('模拟空指针异常'),
            ),

            const SizedBox(height: 16),

            ElevatedButton(
              onPressed: _addBreadcrumbs,
              style: ElevatedButton.styleFrom(
                backgroundColor: Colors.blue,
                padding: const EdgeInsets.symmetric(vertical: 16),
              ),
              child: const Text('添加面包屑'),
            ),
          ],
        ),
      ),
    );
  }

  void _recordHandledException() async {
    try {
      throw Exception('这是一个测试异常');
    } catch (e, stackTrace) {
      await _countly
.recordException(
        e.toString(),
        stackTrace.toString(),
        fatal: false,
      );

      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('已处理异常已记录')),
        );
      }
    }
  }

  void _simulateDivisionByZero() async {
    try {
      // 模拟除零错误
      int result = 100 ~/ 0;
      print(result);
    } catch (e, stackTrace) {
      await _countly.recordException(
        e.toString(),
        stackTrace.toString(),
        fatal: false,
      );

      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('除零错误已记录')),
        );
      }
    }
  }

  void _simulateNullPointerException() async {
    try {
      // 模拟空指针异常
      String? nullString;
      print(nullString!.length);
    } catch (e, stackTrace) {
      await _countly.recordException(
        e.toString(),
        stackTrace.toString(),
        fatal: false,
      );

      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('空指针异常已记录')),
        );
      }
    }
  }

  void _addBreadcrumbs() async {
    await _countly.addBreadcrumb('用户进入崩溃测试页面');
    await _countly.addBreadcrumb('用户点击了添加面包屑按钮');
    await _countly.addBreadcrumb('面包屑记录时间: ${DateTime.now()}');

    if (mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('面包屑已添加')),
      );
    }
  }
}
```

---

## 📱 iOS配置

### 步骤1：配置Info.plist

编辑`ios/Runner/Info.plist`，添加网络权限：

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <!-- 其他配置 -->

    <!-- 允许HTTP连接（开发环境） -->
    <key>NSAppTransportSecurity</key>
    <dict>
        <key>NSAllowsArbitraryLoads</key>
        <true/>
    </dict>

    <!-- 或者只允许特定域名 -->
    <!--
    <key>NSAppTransportSecurity</key>
    <dict>
        <key>NSExceptionDomains</key>
        <dict>
            <key>YOUR_IP</key>
            <dict>
                <key>NSExceptionAllowsInsecureHTTPLoads</key>
                <true/>
                <key>NSIncludesSubdomains</key>
                <true/>
            </dict>
        </dict>
    </dict>
    -->
</dict>
</plist>
```

### 步骤2：配置Podfile（如需要）

编辑`ios/Podfile`：

```ruby
# Uncomment this line to define a global platform for your project
platform :ios, '12.0'

# CocoaPods analytics sends network stats synchronously affecting flutter build latency.
ENV['COCOAPODS_DISABLE_STATS'] = 'true'

project 'Runner', {
  'Debug' => :debug,
  'Profile' => :release,
  'Release' => :release,
}

def flutter_root
  generated_xcode_build_settings_path = File.expand_path(File.join('..', 'Flutter', 'Generated.xcconfig'), __FILE__)
  unless File.exist?(generated_xcode_build_settings_path)
    raise "#{generated_xcode_build_settings_path} must exist. If you're running pod install manually, make sure flutter pub get is executed first"
  end

  File.foreach(generated_xcode_build_settings_path) do |line|
    matches = line.match(/FLUTTER_ROOT\=(.*)/)
    return matches[1].strip if matches
  end
  raise "FLUTTER_ROOT not found in #{generated_xcode_build_settings_path}. Try deleting Generated.xcconfig, then run flutter pub get"
end

require File.expand_path(File.join('packages', 'flutter_tools', 'bin', 'podhelper'), flutter_root)

flutter_ios_podfile_setup

target 'Runner' do
  use_frameworks!
  use_modular_headers!

  flutter_install_all_ios_pods File.dirname(File.realpath(__FILE__))
end

post_install do |installer|
  installer.pods_project.targets.each do |target|
    flutter_additional_ios_build_settings(target)
  end
end
```

### 步骤3：安装iOS依赖

```bash
cd ios
pod install
cd ..
```

---

## 🤖 Android配置

### 步骤1：配置AndroidManifest.xml

编辑`android/app/src/main/AndroidManifest.xml`：

```xml
<manifest xmlns:android="http://schemas.android.com/apk/res/android">
    <!-- 网络权限 -->
    <uses-permission android:name="android.permission.INTERNET" />
    <uses-permission android:name="android.permission.ACCESS_NETWORK_STATE" />

    <application
        android:label="countly_flutter_demo"
        android:name="${applicationName}"
        android:icon="@mipmap/ic_launcher"
        android:usesCleartextTraffic="true">

        <activity
            android:name=".MainActivity"
            android:exported="true"
            android:launchMode="singleTop"
            android:theme="@style/LaunchTheme"
            android:configChanges="orientation|keyboardHidden|keyboard|screenSize|smallestScreenSize|locale|layoutDirection|fontScale|screenLayout|density|uiMode"
            android:hardwareAccelerated="true"
            android:windowSoftInputMode="adjustResize">

            <meta-data
              android:name="io.flutter.embedding.android.NormalTheme"
              android:resource="@style/NormalTheme"
              />

            <intent-filter>
                <action android:name="android.intent.action.MAIN"/>
                <category android:name="android.intent.category.LAUNCHER"/>
            </intent-filter>
        </activity>

        <meta-data
            android:name="flutterEmbedding"
            android:value="2" />
    </application>
</manifest>
```

**重要配置**：
```xml
android:usesCleartextTraffic="true"
```
这允许HTTP连接（开发环境）。生产环境应使用HTTPS。

### 步骤2：配置build.gradle

编辑`android/app/build.gradle`：

```gradle
android {
    namespace "com.example.countly_flutter_demo"
    compileSdkVersion 34
    ndkVersion flutter.ndkVersion

    compileOptions {
        sourceCompatibility JavaVersion.VERSION_1_8
        targetCompatibility JavaVersion.VERSION_1_8
    }

    defaultConfig {
        applicationId "com.example.countly_flutter_demo"
        minSdkVersion 21
        targetSdkVersion 34
        versionCode flutterVersionCode.toInteger()
        versionName flutterVersionName
    }

    buildTypes {
        release {
            signingConfig signingConfigs.debug
        }
    }
}
```

---

## 🧪 测试验证

### 步骤1：运行应用

```bash
# 查看可用设备
flutter devices

# 运行在Android设备/模拟器
flutter run

# 运行在iOS设备/模拟器（macOS only）
flutter run -d ios

# 运行在特定设备
flutter run -d <device_id>
```

### 步骤2：测试功能

**测试清单**：

```
✅ 应用启动
   - 启动应用
   - 查看控制台日志
   - 确认"Countly initialized successfully"

✅ 简单事件
   - 点击"发送简单事件"
   - 查看SnackBar提示
   - 在Countly管理界面验证

✅ 带参数事件
   - 点击"发送带参数事件"
   - 查看SnackBar提示
   - 验证事件参数

✅ 定时事件
   - 点击"测试定时事件"
   - 等待3秒
   - 验证事件时长

✅ 事件测试页面
   - 进入事件测试页面
   - 测试计数器
   - 测试定时事件
   - 测试批量事件

✅ 用户属性
   - 进入用户属性页面
   - 填写用户信息
   - 保存属性
   - 在Countly验证

✅ 崩溃测试
   - 进入崩溃测试页面
   - 测试各种异常
   - 在Countly查看崩溃报告
```

### 步骤3：查看控制台日志

```bash
# Flutter日志
flutter logs

# 过滤Countly日志
flutter logs | grep Countly
```

**预期日志**：
```
[Countly] Initializing...
[Countly] SDK initialized
[Countly] Recording event: button_clicked
[Countly] Sending request...
[Countly] Request completed successfully
```

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
   - 查看事件详情
   - 验证事件参数
   ```

3. **用户数据**
   ```
   Analytics > Users
   - 查看用户列表
   - 查看用户属性
   ```

4. **崩溃报告**
   ```
   Crashes > Overview
   - 查看崩溃列表
   - 查看崩溃详情
   - 查看堆栈信息
   ```

5. **视图数据**
   ```
   Analytics > Views
   - 查看页面浏览
   - 查看停留时间
   ```

---

## ❓ 常见问题

### 1. 无法连接到Countly服务器

**问题**：应用无法发送数据

**排查步骤**：

```dart
// 1. 检查服务器地址
// CountlyService中的serverUrl是否正确
static const String serverUrl = 'http://YOUR_IP:8080';

// 2. 检查App Key
// 是否从Countly管理界面正确复制

// 3. 检查网络权限
// Android: AndroidManifest.xml中是否有INTERNET权限
// iOS: Info.plist中是否配置了NSAppTransportSecurity
```

**解决方案**：

```bash
# 测试服务器连接
# Android
adb shell ping YOUR_IP

# iOS
# 在模拟器中打开Safari访问 http://YOUR_IP:8080
```

### 2. iOS无法使用HTTP连接

**问题**：iOS应用无法连接HTTP服务器

**解决方案**：

在`ios/Runner/Info.plist`中添加：

```xml
<key>NSAppTransportSecurity</key>
<dict>
    <key>NSAllowsArbitraryLoads</key>
    <true/>
</dict>
```

**注意**：生产环境应使用HTTPS。

### 3. Android网络安全配置

**问题**：Android 9+无法使用HTTP

**解决方案**：

在`AndroidManifest.xml`中添加：

```xml
android:usesCleartextTraffic="true"
```

或创建网络安全配置文件。

### 4. 数据未显示在Countly管理界面

**问题**：发送了事件但看不到数据

**可能原因**：

```
1. 数据延迟
   - 等待1-2分钟后刷新

2. SDK未正确初始化
   - 检查控制台日志
   - 确认"SDK initialized"消息

3. 选择了错误的应用
   - 在Countly管理界面确认应用选择

4. 网络问题
   - 检查设备网络连接
   - 检查防火墙设置
```

### 5. 热重载后Countly失效

**问题**：热重载后Countly不工作

**解决方案**：

```dart
// 在main.dart中确保初始化在runApp之前
void main() async {
  WidgetsFlutterBinding.ensureInitialized();

  // 初始化Countly
  await CountlyService().init();

  runApp(const MyApp());
}
```

热重载不会重新执行main()，需要完全重启应用。

### 6. 模拟器无法访问localhost

**问题**：使用localhost无法连接

**解决方案**：

```dart
// Android模拟器使用特殊地址
static const String serverUrl = 'http://10.0.2.2:8080';

// iOS模拟器可以使用localhost
static const String serverUrl = 'http://localhost:8080';

// 或使用局域网IP
static const String serverUrl = 'http://192.168.1.100:8080';
```

---

## 📦 构建应用

### Debug版本

```bash
# Android APK
flutter build apk --debug

# iOS App（macOS only）
flutter build ios --debug

# 输出位置
# Android: build/app/outputs/flutter-apk/app-debug.apk
# iOS: build/ios/iphoneos/Runner.app
```

### Release版本

**Android**：

```bash
# 1. 生成签名密钥（首次）
keytool -genkey -v -keystore ~/upload-keystore.jks -keyalg RSA -keysize 2048 -validity 10000 -alias upload

# 2. 配置签名
# 创建 android/key.properties
storePassword=<password>
keyPassword=<password>
keyAlias=upload
storeFile=<path-to-keystore>

# 3. 编辑 android/app/build.gradle
# 添加签名配置（参考Android文档）

# 4. 构建Release APK
flutter build apk --release

# 输出位置
# build/app/outputs/flutter-apk/app-release.apk
```

**iOS**（macOS only）：

```bash
# 1. 配置签名
# 在Xcode中配置开发者账号和证书

# 2. 构建Release版本
flutter build ios --release

# 3. 在Xcode中Archive和上传
open ios/Runner.xcworkspace
```

---

## 🎯 性能优化

### 1. 减少网络请求

```dart
// 在CountlyConfig中配置
config.setEventQueueSizeToSend(10);  // 累积10个事件后发送
config.setUpdateSessionTimerDelay(60);  // 60秒更新一次会话
```

### 2. 离线支持

Countly SDK自动处理离线情况：
- 离线时缓存事件
- 联网后自动发送
- 无需额外配置

### 3. 减少日志输出

```dart
// 生产环境关闭日志
config.setLoggingEnabled(false);
```

### 4. 优化初始化

```dart
// 在应用启动时异步初始化
void main() async {
  WidgetsFlutterBinding.ensureInitialized();

  // 异步初始化，不阻塞UI
  CountlyService().init();

  runApp(const MyApp());
}
```

---

## 🚀 高级功能

### 1. 推送通知（可选）

```dart
// 配置推送通知
config.enablePushNotifications();

// 处理推送点击
Countly.onNotificationClicked((notification) {
  print('Notification clicked: $notification');
});
```

### 2. 远程配置

```dart
// 获取远程配置
await Countly.remoteConfigUpdate();

// 获取配置值
var value = await Countly.getRemoteConfigValueForKey('feature_flag');
```

### 3. A/B测试

```dart
// 获取变体
var variant = await Countly.getRemoteConfigValueForKey('button_color');

// 根据变体显示不同UI
if (variant == 'blue') {
  // 显示蓝色按钮
} else {
  // 显示默认按钮
}
```

### 4. 用户反馈

```dart
// 显示反馈表单
await Countly.showFeedbackWidget('widget_id');

// 获取NPS评分
await Countly.showNPS();
```

---

## 📚 参考资源

### 官方文档

- **Flutter SDK文档**：https://support.count.ly/hc/en-us/articles/360037754511-Flutter
- **GitHub仓库**：https://github.com/Countly/countly-sdk-flutter-bridge
- **Pub.dev**：https://pub.dev/packages/countly_flutter

### 示例代码

- **官方Demo**：https://github.com/Countly/countly-sdk-flutter-bridge/tree/master/example
- **集成示例**：https://github.com/Countly/countly-sdk-flutter-bridge/wiki

### 学习资源

- **Flutter官网**：https://flutter.dev
- **Dart官网**：https://dart.dev
- **Countly博客**：https://countly.com/blog

---

## ✅ 完成检查清单

```
环境准备：
- [ ] 安装Flutter SDK
- [ ] 安装开发工具
- [ ] 配置开发环境
- [ ] 准备Countly服务器

项目创建：
- [ ] 创建Flutter项目
- [ ] 添加Countly依赖
- [ ] 创建服务类

功能实现：
- [ ] 主界面（HomeScreen）
- [ ] 事件测试页面（EventTestScreen）
- [ ] 用户属性页面（UserProfileScreen）
- [ ] 崩溃测试页面（CrashTestScreen）

平台配置：
- [ ] iOS配置（Info.plist）
- [ ] Android配置（AndroidManifest.xml）

测试验证：
- [ ] 运行应用
- [ ] 测试所有功能
- [ ] 在Countly管理界面验证数据
- [ ] 查看控制台日志

构建发布：
- [ ] 构建Debug版本
- [ ] 配置签名
- [ ] 构建Release版本
```

---

## 🎉 总结

### Flutter集成优势

```
✅ 跨平台：一套代码，iOS和Android都支持
✅ 热重载：快速开发和调试
✅ 性能好：接近原生性能
✅ 易维护：统一的代码库
✅ 社区活跃：丰富的插件生态
```

### 与原生对比

| 特性 | Flutter | Android原生 | iOS原生 |
|------|---------|------------|---------|
| **开发语言** | Dart | Kotlin/Java | Swift/ObjC |
| **代码复用** | 100% | 0% | 0% |
| **开发效率** | 高 | 中 | 中 |
| **性能** | 接近原生 | 最优 | 最优 |
| **学习曲线** | 中等 | 中等 | 中等 |

### 下一步

1. **扩展功能**
   - 添加推送通知
   - 实现远程配置
   - 添加A/B测试
   - 集成用户反馈

2. **优化性能**
   - 减少网络请求
   - 优化内存使用
   - 提升启动速度

3. **完善UI**
   - 优化界面设计
   - 添加动画效果
   - 提升用户体验

---

**恭喜！你已经完成了Countly Flutter SDK的集成和测试！**

现在你可以：
1. 在真实Flutter应用中集成Countly
2. 追踪跨平台用户行为
3. 分析iOS和Android数据
4. 优化应用体验

---

*文档版本：v1.0*
*最后更新：2026年1月28日*
*作者：Kiro AI*
