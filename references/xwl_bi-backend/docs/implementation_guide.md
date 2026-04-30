# 实施指南 (Implementation Guide)

本文档旨在指导开发人员如何将 [行为分析规划](behavior_analysis.md) 和 [用户分析规划](user_analysis.md) 落地到现有系统中。

## 1. 数据库元数据初始化 (Database Initialization)

请在 MySQL 数据库中执行以下 SQL 语句，以注册新的事件和属性元数据。
*注意：请将 `@appid` 替换为实际的项目 AppID。*

```sql
SET @appid = 1; -- 请修改为实际 AppID

-- 1. 注册事件 (meta_event)
INSERT IGNORE INTO `meta_event` (`appid`, `event_name`, `show_name`) VALUES
(@appid, 'app_start', '应用启动'),
(@appid, 'login_success', '登录成功'),
(@appid, 'guest_upgrade', '游客升级'),
(@appid, 'view_home', '浏览首页'),
(@appid, 'view_foryou', '浏览推荐页'),
(@appid, 'view_fav', '浏览收藏页'),
(@appid, 'view_drama_detail', '浏览剧集详情'),
(@appid, 'play_start', '开始播放'),
(@appid, 'play_end', '播放结束'),
(@appid, 'click_unlock', '点击解锁'),
(@appid, 'view_checkout', '浏览收银台'),
(@appid, 'select_sku', '选中商品'),
(@appid, 'initiate_pay', '发起支付'),
(@appid, 'pay_success', '支付成功'),
(@appid, 'like_drama', '点赞'),
(@appid, 'favorite_drama', '收藏'),
(@appid, 'share_drama', '分享');

-- 2. 注册属性 (attribute)
-- 属性类型 data_type: 1=String, 2=Int, 3=Float
-- 属性来源 attribute_source: 1=用户属性, 2=事件属性
-- 属性分类 attribute_type: 2=自定义属性

-- 2.1 公共/事件属性
INSERT IGNORE INTO `attribute` (`app_id`, `attribute_name`, `show_name`, `data_type`, `attribute_type`, `attribute_source`, `status`) VALUES
(@appid, 'is_guest', '是否游客', '1', 2, 2, 1),
(@appid, 'vip_status', '会员状态', '1', 2, 2, 1),
(@appid, 'drama_id', '剧集ID', '1', 2, 2, 1),
(@appid, 'drama_name', '剧集名称', '1', 2, 2, 1),
(@appid, 'episode_num', '集数', '2', 2, 2, 1),
(@appid, 'is_free', '是否免费', '1', 2, 2, 1),
(@appid, 'play_duration', '播放时长(秒)', '2', 2, 2, 1),
(@appid, 'progress', '播放进度(%)', '2', 2, 2, 1),
(@appid, 'is_finish', '是否看完', '1', 2, 2, 1),
(@appid, 'sku_id', '商品ID', '1', 2, 2, 1),
(@appid, 'pay_amount', '支付金额', '3', 2, 2, 1),
(@appid, 'pay_method', '支付方式', '1', 2, 2, 1),
(@appid, 'order_id', '订单ID', '1', 2, 2, 1),
(@appid, 'rec_strategy_id', '推荐策略ID', '1', 2, 2, 1),
(@appid, 'source', '来源', '1', 2, 2, 1);

-- 2.2 用户属性
INSERT IGNORE INTO `attribute` (`app_id`, `attribute_name`, `show_name`, `data_type`, `attribute_type`, `attribute_source`, `status`) VALUES
(@appid, 'coin_balance', '剩余金币', '2', 2, 1, 1),
(@appid, 'total_pay_amount', '累计充值金额', '3', 2, 1, 1),
(@appid, 'vip_expire_date', '会员到期日', '1', 2, 1, 1),
(@appid, 'last_watch_drama', '最近观看剧集', '1', 2, 1, 1),
(@appid, 'channel', '渠道来源', '1', 2, 1, 1);
```

## 2. 前端 SDK 接入指南 (Frontend Integration)

### 2.1 初始化与公共属性设置

在应用启动时（如 `main.js` 或 `App.vue`），初始化 SDK 并设置全局状态。

```javascript
// 假设已引入 EventReport
const report = new EventReport("YOUR_SERVER_URL", "YOUR_APPID", "YOUR_APPKEY", 0); // debug=0, 1=log, 2=alert

// 1. 设置公共属性 (所有事件都会带上)
// 根据当前用户状态判断是否游客
const isGuest = !localStorage.getItem('token'); 
const vipStatus = localStorage.getItem('vip_status') || 'none';

report.setSuperProperties({
    "is_guest": isGuest ? "true" : "false",
    "vip_status": vipStatus,
    "channel": "douyin_ad" // 示例渠道
});

// 2. 上报应用启动
report.track("app_start", {
    "is_first_time": "false" // 需自行判断是否首次
});
```

### 2.2 用户登录与状态切换

当用户从游客状态登录成功后，必须调用 `login` 方法关联账号。

```javascript
function onLoginSuccess(userId, userInfo) {
    // 1. 关联账号 (将之前的游客行为归因到该账号)
    report.login(userId);

    // 2. 更新公共属性
    report.setSuperProperties({
        "is_guest": "false",
        "vip_status": userInfo.vipStatus
    });

    // 3. 上报登录事件
    report.track("login_success", {
        "method": "wechat",
        "is_new_user": "false"
    });

    // 4. 更新用户属性 (User Profile)
    report.userSet({
        "coin_balance": userInfo.coinBalance,
        "vip_expire_date": userInfo.vipExpireDate,
        "last_login_time": new Date().toISOString()
    });
    // 触发用户属性上报
    report.trackUserData();
}
```

### 2.3 核心业务埋点示例

#### 浏览剧集详情
```javascript
report.track("view_drama_detail", {
    "drama_id": "10086",
    "drama_name": "霸道总裁爱上我",
    "source": "homepage_banner"
});
```

#### 播放结束 (关键事件)
```javascript
// 当视频暂停、播放结束或页面销毁时调用
report.track("play_end", {
    "drama_id": "10086",
    "episode_num": 5,
    "play_duration": 120, // 秒
    "progress": 85,       // %
    "is_finish": "true",  // 是否看完关键剧情
    "is_free": "true"
});
```

#### 发起支付
```javascript
report.track("initiate_pay", {
    "sku_id": "coin_pack_30",
    "pay_amount": 29.9,
    "pay_method": "alipay",
    "scene": "unlock_modal" // 触发场景
});
```

#### 支付成功
```javascript
report.track("pay_success", {
    "order_id": "202310270001",
    "pay_amount": 29.9,
    "sku_id": "coin_pack_30"
});

// 同时更新用户累计充值金额
report.userAdd({
    "total_pay_amount": 29.9,
    "coin_balance": 3000
});
report.trackUserData();
```

## 3. 注意事项
1.  **字符串类型**: SDK 传输属性值时，建议基础类型尽量保持一致。布尔值推荐使用字符串 `"true"/"false"` 或整数 `1/0`，避免兼容性问题。
2.  **埋点验证**: 在开发阶段，将 SDK 初始化参数 `debug` 设置为 `1`，可以在浏览器控制台看到上报日志。
3.  **用户属性更新**: `userSet` 只是更新本地状态，必须调用 `trackUserData()` 才会真正发送到服务器。建议在关键属性变更时（如充值后、升级会员后）立即调用。
