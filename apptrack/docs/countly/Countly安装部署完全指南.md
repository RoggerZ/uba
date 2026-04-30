# Countly 安装部署完全指南

> 基于官方文档的开源移动分析平台部署教程
> 日期：2026年1月28日

---

## 📋 目录

1. [系统要求](#系统要求)
2. [安装方式对比](#安装方式对比)
3. [一键安装脚本（推荐）](#一键安装脚本推荐)
4. [Docker安装](#docker安装)
5. [手动安装](#手动安装)
6. [配置优化](#配置优化)
7. [常见问题](#常见问题)

---

## 💻 系统要求

### 最低配置

**硬件要求**：
```
CPU：2核心
内存：4GB RAM
存储：20GB SSD
网络：稳定的互联网连接
```

**软件要求**：
```
操作系统：
- Ubuntu 20.04/22.04 LTS（推荐）
- CentOS 7/8/Stream
- RHEL 7/8
- Debian 10/11

端口要求：
- 80（HTTP）- 必须空闲
- 443（HTTPS）- 必须空闲
- 27017（MongoDB）- 内部使用
```

### 推荐配置

**生产环境**：
```
CPU：4核心+
内存：8GB RAM+
存储：100GB SSD+
带宽：100Mbps+
```

---

## 🔄 安装方式对比

| 方式 | 优点 | 缺点 | 适合场景 |
|------|------|------|---------|
| **一键脚本** | 最简单、自动化 | 定制性较差 | 快速部署、生产环境 |
| **Docker** | 隔离性好、易管理 | 需要Docker知识 | 测试、开发环境 |
| **手动安装** | 完全控制 | 复杂耗时 | 特殊需求 |

**官方推荐**：一键安装脚本（适合生产环境）

---

## ⚡ 一键安装脚本（推荐）

### 方法1：使用官方安装脚本

这是Countly官方推荐的安装方式，适合生产环境。

**前提条件**：
- 全新的Ubuntu/CentOS/RHEL服务器
- 端口80和443必须空闲且对外开放
- Root权限

**安装命令**：
```bash
# 使用wget
wget -qO- https://c.ly/install | bash

# 或使用curl
curl -s https://c.ly/install | bash
```

**安装过程**：
```
1. 检测操作系统类型
2. 安装依赖包（Node.js, MongoDB, Nginx等）
3. 下载Countly最新版本
4. 配置数据库和服务
5. 启动Countly服务

预计时间：10-15分钟
```

**安装完成后**：
```bash
# 访问Countly
http://YOUR_SERVER_IP

# 首次访问会要求创建管理员账号
```


### 方法2：Digital Ocean一键部署

如果你有Digital Ocean账号，可以使用官方的一键部署：

1. 访问：https://marketplace.digitalocean.com/apps/countly
2. 选择服务器配置和数据中心
3. 点击"Create Droplet"
4. 10分钟内完成部署

---

## 🐳 Docker安装

### 重要说明

Docker方式适合开发和测试环境。官方提供了Docker支持，但推荐生产环境使用一键安装脚本。

### 步骤1：安装Docker和Docker Compose

**Ubuntu/Debian**：
```bash
# 更新系统
sudo apt update
sudo apt upgrade -y

# 安装Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# 启动Docker
sudo systemctl start docker
sudo systemctl enable docker

# 安装Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# 验证安装
docker --version
docker-compose --version
```

### 步骤2：下载Countly

```bash
# 克隆Countly仓库
git clone https://github.com/Countly/countly-server.git
cd countly-server
```

### 步骤3：使用Docker Compose启动

Countly官方仓库包含docker-compose.yml文件。

```bash
# 启动Countly（在countly-server目录下）
docker-compose up -d

# 查看运行状态
docker-compose ps

# 查看日志
docker-compose logs -f
```

### 步骤4：访问Countly

```bash
# 打开浏览器访问
http://YOUR_SERVER_IP:6001

# 注意：Docker版本默认端口是6001，不是80
```

### Docker配置说明

**环境变量配置**：

可以通过环境变量配置Countly：

```bash
# 创建.env文件
cat > .env << 'EOF'
# MongoDB配置
MONGODB_ROOT_PASSWORD=your_secure_password
MONGODB_USERNAME=countly
MONGODB_PASSWORD=your_countly_password

# Countly配置
COUNTLY_CONFIG_API_HOST=0.0.0.0
COUNTLY_CONFIG_FRONTEND_HOST=0.0.0.0
EOF
```

**数据持久化**：

Docker Compose会自动创建volumes来持久化数据：
- MongoDB数据
- Countly配置文件
- 日志文件

### Docker管理命令

```bash
# 停止服务
docker-compose down

# 重启服务
docker-compose restart

# 查看日志
docker-compose logs -f countly-api
docker-compose logs -f countly-frontend

# 更新Countly
git pull
docker-compose pull
docker-compose up -d

# 备份数据
docker-compose exec mongodb mongodump --out /backup
```

### Docker注意事项

1. **端口映射**：Docker版本默认使用6001端口，可以在docker-compose.yml中修改
2. **性能**：Docker会有轻微的性能开销
3. **生产环境**：官方推荐生产环境使用一键安装脚本而非Docker
4. **数据备份**：定期备份MongoDB数据卷

---

## 🔧 手动安装

### 方法3：从GitHub手动安装

适合需要完全控制安装过程的场景。

**步骤1：下载Countly**

```bash
# 下载最新版本
wget https://github.com/Countly/countly-server/archive/refs/heads/master.zip
unzip master.zip
cd countly-server-master
```

**步骤2：运行安装脚本**

```bash
# Ubuntu/Debian
sudo su -
cd /path/to/countly-server-master/bin
bash countly.install.sh

# CentOS/RHEL
sudo su -
cd /path/to/countly-server-master/bin
bash countly.install_rhel.sh
```

**步骤3：等待安装完成**

安装时间：6-10分钟，取决于服务器性能。

---

## 🔧 配置优化

### 1. 配置HTTPS（Let's Encrypt）

安装完成后，强烈建议配置HTTPS。

```bash
# 安装Certbot
sudo apt install certbot python3-certbot-nginx -y

# 获取SSL证书
sudo certbot --nginx -d countly.yourdomain.com

# 自动续期测试
sudo certbot renew --dry-run
```

### 2. 配置邮件服务

编辑邮件配置文件：

```bash
# 进入Countly目录
cd /opt/countly

# 复制邮件配置示例
cp extend/mail.example.js extend/mail.js

# 编辑配置
nano extend/mail.js
```

**Gmail配置示例**：

```javascript
module.exports = function(mail){
    var nodemailer = require('nodemailer');
    var smtpTransport = require('nodemailer-smtp-transport');

    mail.smtpTransport = nodemailer.createTransport(smtpTransport({
        host: "smtp.gmail.com",
        secureConnection: true,
        port: 587,
        auth: {
            user: "your-email@gmail.com",
            pass: "your-app-password"  // 使用应用专用密码
        }
    }));

    mail.from = "Countly <your-email@gmail.com>";
};
```

**重启Countly**：

```bash
sudo countly restart
```


### 3. MongoDB优化

编辑MongoDB配置：

```bash
sudo nano /etc/mongod.conf
```

```yaml
# MongoDB优化配置
storage:
  dbPath: /var/lib/mongodb
  journal:
    enabled: true
  wiredTiger:
    engineConfig:
      cacheSizeGB: 2  # 设置为可用RAM的50%

systemLog:
  destination: file
  logAppend: true
  path: /var/log/mongodb/mongod.log

net:
  port: 27017
  bindIp: 127.0.0.1  # 只允许本地连接

security:
  authorization: enabled  # 启用认证
```

**重启MongoDB**：

```bash
sudo systemctl restart mongod
```

### 4. Nginx反向代理配置

如果需要自定义域名和SSL：

```nginx
# /etc/nginx/sites-available/countly
server {
    listen 80;
    server_name countly.yourdomain.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name countly.yourdomain.com;

    ssl_certificate /etc/letsencrypt/live/countly.yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/countly.yourdomain.com/privkey.pem;

    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;

    location / {
        proxy_pass http://localhost:6001;  # Countly默认端口
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;
    }
}
```

**启用配置**：

```bash
sudo ln -s /etc/nginx/sites-available/countly /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl restart nginx
```

---

## 🛠️ Countly命令行工具

安装完成后，可以使用`countly`命令管理服务。

### 基本命令

```bash
# 启动Countly
sudo countly start

# 停止Countly
sudo countly stop

# 重启Countly
sudo countly restart

# 查看状态
sudo countly status

# 查看版本
sudo countly version

# 升级Countly
sudo countly upgrade

# 备份数据
sudo countly backup /path/to/backup

# 恢复数据
sudo countly restore /path/to/backup
```

### 备份和恢复

**完整备份**（包括数据库、配置、文件）：

```bash
# 备份到指定目录
sudo countly backup /var/backups/countly

# 只备份数据库
sudo countly backupdb /var/backups/countly-db

# 只备份文件
sudo countly backupfiles /var/backups/countly-files
```

**恢复备份**：

```bash
# 完整恢复
sudo countly restore /var/backups/countly

# 只恢复数据库
sudo countly restoredb /var/backups/countly-db

# 只恢复文件
sudo countly restorefiles /var/backups/countly-files
```

### 升级Countly

```bash
# 升级到最新版本
sudo countly upgrade

# 升级过程：
# 1. 停止Countly服务
# 2. 备份当前版本
# 3. 下载新版本
# 4. 更新数据库
# 5. 重启服务
```

---

## 📱 集成移动应用

### iOS集成

**安装SDK**：

```ruby
# Podfile
pod 'Countly'
```

```bash
pod install
```

**初始化SDK**：

```swift
// AppDelegate.swift
import Countly

func application(_ application: UIApplication,
                 didFinishLaunchingWithOptions launchOptions: [UIApplication.LaunchOptionsKey: Any]?) -> Bool {

    let config = CountlyConfig()
    config.appKey = "YOUR_APP_KEY"
    config.host = "https://countly.yourdomain.com"  // 或 http://YOUR_SERVER_IP

    Countly.sharedInstance().start(with: config)

    return true
}
```

### Android集成

**添加依赖**：

```gradle
// build.gradle (app)
dependencies {
    implementation 'ly.count.android:sdk:23.8.0'
}
```

**初始化SDK**：

```kotlin
// Application.kt
import ly.count.android.sdk.Countly

class MyApp : Application() {
    override fun onCreate() {
        super.onCreate()

        Countly.sharedInstance().init(
            this,
            "https://countly.yourdomain.com",  // 或 http://YOUR_SERVER_IP
            "YOUR_APP_KEY"
        )
    }
}
```

**获取APP KEY**：

1. 登录Countly管理界面
2. 进入 Management > Applications
3. 创建新应用或选择现有应用
4. 复制App Key

---

## 🔍 验证安装

### 检查服务状态

```bash
# 检查Countly服务
sudo countly status

# 检查MongoDB
sudo systemctl status mongod

# 检查Nginx
sudo systemctl status nginx

# 检查端口监听
sudo netstat -tulpn | grep -E '(80|443|6001|27017)'
```

### 测试API连接

```bash
# 测试API端点
curl http://localhost:6001/o/ping

# 预期输出
{"result":"pong"}
```

### 查看日志

```bash
# Countly API日志
sudo tail -f /var/log/countly/countly-api.log

# Countly Frontend日志
sudo tail -f /var/log/countly/countly-dashboard.log

# MongoDB日志
sudo tail -f /var/log/mongodb/mongod.log

# Nginx日志
sudo tail -f /var/log/nginx/error.log
sudo tail -f /var/log/nginx/access.log
```

---

## ❓ 常见问题

### 1. 无法访问Countly界面

**问题**：浏览器无法访问 http://SERVER_IP

**排查步骤**：

```bash
# 1. 检查Countly是否运行
sudo countly status

# 2. 检查端口是否监听
sudo netstat -tulpn | grep 6001

# 3. 检查防火墙
sudo ufw status
sudo ufw allow 80
sudo ufw allow 443

# 4. 检查Nginx配置
sudo nginx -t

# 5. 查看错误日志
sudo tail -f /var/log/countly/countly-api.log
```

### 2. MongoDB连接失败

**问题**：Countly无法连接MongoDB

**解决方案**：

```bash
# 检查MongoDB状态
sudo systemctl status mongod

# 重启MongoDB
sudo systemctl restart mongod

# 检查MongoDB日志
sudo tail -f /var/log/mongodb/mongod.log

# 测试MongoDB连接
mongo --eval "db.adminCommand('ping')"
```

### 3. 端口80/443被占用

**问题**：安装时提示端口被占用

**解决方案**：

```bash
# 查看占用端口的进程
sudo lsof -i :80
sudo lsof -i :443

# 停止占用端口的服务
sudo systemctl stop apache2  # 如果是Apache
sudo systemctl stop nginx    # 如果是Nginx

# 然后重新安装Countly
```

### 4. 内存不足

**问题**：系统内存不足，服务崩溃

**解决方案**：

```bash
# 添加Swap空间
sudo fallocate -l 4G /swapfile
sudo chmod 600 /swapfile
sudo mkswap /swapfile
sudo swapon /swapfile

# 永久启用
echo '/swapfile none swap sw 0 0' | sudo tee -a /etc/fstab

# 优化MongoDB内存
# 编辑 /etc/mongod.conf
# 设置 wiredTiger.engineConfig.cacheSizeGB = 1
```

### 5. 邮件发送失败

**问题**：无法发送邮件通知

**解决方案**：

```bash
# 检查邮件配置
cat /opt/countly/extend/mail.js

# Gmail用户需要：
# 1. 启用"两步验证"
# 2. 生成"应用专用密码"
# 3. 使用应用专用密码而非账号密码

# 测试SMTP连接
telnet smtp.gmail.com 587
```

### 6. Docker容器无法启动

**问题**：docker-compose up失败

**解决方案**：

```bash
# 查看详细错误
docker-compose logs

# 清理并重新启动
docker-compose down -v
docker-compose up -d

# 检查Docker资源
docker system df
docker system prune  # 清理未使用的资源
```


---

## 📊 性能优化建议

### 1. 数据库索引优化

```javascript
// 连接MongoDB
mongo

// 切换到countly数据库
use countly

// 创建常用索引
db.events.createIndex({app_id: 1, timestamp: -1})
db.events.createIndex({uid: 1, timestamp: -1})
db.app_users.createIndex({uid: 1, app_id: 1})
db.sessions.createIndex({app_id: 1, timestamp: -1})
```

### 2. 数据清理策略

```bash
# 在Countly管理界面设置数据保留期：
# Management > Applications > [Your App] > Data Retention

# 推荐设置：
# - 事件数据：90天
# - 会话数据：180天
# - 用户数据：永久
```

### 3. 系统资源监控

```bash
# 安装监控工具
sudo apt install htop iotop -y

# 实时监控
htop

# 磁盘使用
df -h

# 内存使用
free -m

# MongoDB数据库大小
mongo
> use countly
> db.stats()
```

### 4. 定期维护任务

创建定期维护脚本：

```bash
# 创建维护脚本
sudo nano /opt/countly/maintenance.sh
```

```bash
#!/bin/bash
# Countly定期维护脚本

# 备份数据库
countly backup /var/backups/countly-$(date +%Y%m%d)

# 清理旧备份（保留最近7天）
find /var/backups -name "countly-*" -mtime +7 -delete

# 压缩MongoDB数据库
mongo countly --eval "db.runCommand({compact: 'events'})"

# 重启服务
countly restart

echo "Maintenance completed at $(date)"
```

```bash
# 添加执行权限
sudo chmod +x /opt/countly/maintenance.sh

# 添加到crontab（每周日凌晨3点执行）
sudo crontab -e
# 添加：0 3 * * 0 /opt/countly/maintenance.sh >> /var/log/countly-maintenance.log 2>&1
```

---

## 🔐 安全加固

### 1. 防火墙配置

```bash
# 启用UFW防火墙
sudo ufw enable

# 只开放必要端口
sudo ufw default deny incoming
sudo ufw default allow outgoing
sudo ufw allow 22/tcp    # SSH
sudo ufw allow 80/tcp    # HTTP
sudo ufw allow 443/tcp   # HTTPS

# 查看状态
sudo ufw status verbose
```

### 2. MongoDB安全加固

```bash
# 创建MongoDB管理员用户
mongo
> use admin
> db.createUser({
    user: "admin",
    pwd: "your_secure_password",
    roles: ["root"]
})

# 创建Countly数据库用户
> use countly
> db.createUser({
    user: "countly",
    pwd: "your_countly_password",
    roles: ["readWrite"]
})

# 退出并编辑配置启用认证
> exit

sudo nano /etc/mongod.conf
# 添加：
# security:
#   authorization: enabled

sudo systemctl restart mongod
```

### 3. 定期更新系统

```bash
# 更新系统包
sudo apt update && sudo apt upgrade -y

# 更新Countly
sudo countly upgrade

# 设置自动安全更新
sudo apt install unattended-upgrades -y
sudo dpkg-reconfigure -plow unattended-upgrades
```

### 4. 限制SSH访问

```bash
# 编辑SSH配置
sudo nano /etc/ssh/sshd_config

# 推荐设置：
# PermitRootLogin no
# PasswordAuthentication no  # 使用SSH密钥
# Port 2222  # 更改默认端口

# 重启SSH
sudo systemctl restart sshd
```

---

## 📚 参考资源

### 官方资源

- **官方网站**：https://countly.com
- **官方文档**：https://support.count.ly
- **GitHub仓库**：https://github.com/Countly/countly-server
- **Docker Hub**：https://hub.docker.com/r/countly/countly-server
- **API文档**：https://api.count.ly

### 社区资源

- **Discord社区**：https://discord.gg/countly
- **社区论坛**：https://community.count.ly
- **Stack Overflow**：标签 `countly`
- **GitHub Issues**：https://github.com/Countly/countly-server/issues

### 学习资源

- **官方博客**：https://countly.com/blog
- **YouTube频道**：Countly Analytics
- **文档中心**：https://resources.count.ly

---

## 🎯 快速开始检查清单

### 安装前准备

```
- [ ] 准备服务器（最低4GB RAM）
- [ ] 确保端口80和443空闲
- [ ] 配置域名DNS（可选）
- [ ] 准备邮箱用于SMTP（可选）
```

### 安装步骤

```
- [ ] 选择安装方式（推荐：一键脚本）
- [ ] 执行安装命令
- [ ] 等待安装完成（10-15分钟）
- [ ] 访问Web界面
```

### 初始配置

```
- [ ] 创建管理员账号
- [ ] 创建第一个应用
- [ ] 获取App Key
- [ ] 配置HTTPS（推荐）
- [ ] 配置邮件服务（可选）
```

### 集成测试

```
- [ ] 集成iOS/Android SDK
- [ ] 发送测试事件
- [ ] 验证数据显示
- [ ] 测试推送通知（可选）
```

### 维护设置

```
- [ ] 设置自动备份
- [ ] 配置监控告警
- [ ] 设置数据保留策略
- [ ] 定期更新系统
```

---

## 💡 部署建议

### 小型部署（<10K MAU）

```
服务器配置：
- 2核CPU，4GB RAM
- 50GB SSD
- Ubuntu 22.04 LTS

安装方式：
- 一键安装脚本

预计成本：
- VPS：$10-20/月
```

### 中型部署（10K-100K MAU）

```
服务器配置：
- 4核CPU，8GB RAM
- 200GB SSD
- Ubuntu 22.04 LTS

安装方式：
- 一键安装脚本
- 独立MongoDB服务器（可选）

预计成本：
- VPS：$40-80/月
```

### 大型部署（>100K MAU）

```
服务器配置：
- 8核CPU，16GB RAM
- 500GB SSD
- 负载均衡 + 多节点

安装方式：
- 分布式部署
- MongoDB副本集
- Nginx负载均衡

预计成本：
- 多服务器：$200+/月
```

---

## 🔄 从其他平台迁移

### 从Mixpanel迁移

1. 导出Mixpanel数据
2. 使用Countly导入API
3. 更新SDK配置
4. 验证数据完整性

### 从Firebase迁移

1. 导出Firebase Analytics数据
2. 转换数据格式
3. 导入到Countly
4. 更新应用SDK

### 从Google Analytics迁移

1. 导出GA数据
2. 映射事件和属性
3. 导入到Countly
4. 更新追踪代码

---

## 📞 获取帮助

### 遇到问题？

1. **查看文档**：https://support.count.ly
2. **搜索论坛**：https://community.count.ly
3. **GitHub Issues**：https://github.com/Countly/countly-server/issues
4. **Discord社区**：https://discord.gg/countly

### 商业支持

如需专业支持，可以考虑：
- Countly Enterprise Edition（包含SLA支持）
- 专业服务和咨询
- 定制开发

---

## 🎉 总结

### 推荐安装方式

**生产环境**：
```bash
# 使用官方一键安装脚本（最简单、最可靠）
wget -qO- https://c.ly/install | bash
```

**开发/测试环境**：
```bash
# 使用Docker（隔离性好、易于管理）
git clone https://github.com/Countly/countly-server.git
cd countly-server
docker-compose up -d
```

### 关键要点

1. ✅ **一键脚本是官方推荐的生产环境安装方式**
2. ✅ **Docker适合开发和测试环境**
3. ✅ **确保端口80和443空闲且对外开放**
4. ✅ **安装后立即配置HTTPS**
5. ✅ **定期备份数据库**
6. ✅ **设置数据保留策略**
7. ✅ **监控系统资源使用**

### 下一步

安装完成后：
1. 创建你的第一个应用
2. 集成iOS/Android SDK
3. 发送测试事件
4. 探索Countly功能
5. 配置推送通知
6. 设置自动化报告

---

**恭喜！你已经成功部署Countly！**

现在可以开始追踪和分析你的移动应用数据了。

---

*文档版本：v2.0*
*最后更新：2026年1月28日*
*基于官方文档：https://support.count.ly*
*作者：Kiro AI*
