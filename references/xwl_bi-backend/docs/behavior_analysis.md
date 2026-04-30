# 短剧项目行为分析规划

## 1. 概述
本文档旨在为短剧项目提供用户行为分析的规划方案。针对 App 支持免登录观看、包含金币与套餐体系的特点，重点关注游客转化、内容消费及付费转化路径。

## 2. 核心事件设计 (Event Design)

### 2.1 基础与账号事件

| 事件显示名 | 事件名 (Event Name) | 触发时机 | 关键属性 (Properties) |
| :--- | :--- | :--- | :--- |
| 应用启动 | `app_start` | 用户打开 App 或小程序冷启动时 | 启动来源(source), 是否首次启动(is_first_time), 是否登录(is_login) |
| 登录成功 | `login_success` | 用户完成登录流程 | 登录方式(method), 是否新注册(is_new_user) |
| 游客升级 | `guest_upgrade` | 游客绑定账号或注册成功 | 原游客ID(guest_id), 新用户ID(user_id) |

### 2.2 浏览与播放事件

| 事件显示名 | 事件名 (Event Name) | 触发时机 | 关键属性 (Properties) |
| :--- | :--- | :--- | :--- |
| 浏览首页 | `view_home` | 进入首页 Tab | 推荐策略ID(rec_strategy_id) |
| 浏览推荐页 | `view_foryou` | 进入 For You Tab (沉浸式推荐) | 视频ID(video_id), 来源算法(algo_ver) |
| 浏览收藏页 | `view_fav` | 进入 Fav Tab | 收藏数量(fav_count) |
| 浏览剧集详情 | `view_drama_detail` | 进入剧集详情页 | 剧集ID(drama_id), 来源(source) |
| 开始播放 | `play_start` | 视频开始播放 | 剧集ID(drama_id), 集数(episode_num), 是否免费(is_free), 观看模式(mode: guest/login) |
| 播放结束 | `play_end` | 视频播放结束/暂停/退出 | 播放时长(play_duration), 进度百分比(progress), 是否看完(is_finish) |

### 2.3 付费与解锁事件 (金币 & 套餐)

| 事件显示名 | 事件名 (Event Name) | 触发时机 | 关键属性 (Properties) |
| :--- | :--- | :--- | :--- |
| 点击解锁 | `click_unlock` | 点击付费解锁按钮 | 剧集ID(drama_id), 集数(episode_num), 所需金币(cost_coins) |
| 浏览收银台 | `view_checkout` | 弹出充值/套餐选择弹窗 | 触发场景(scene: unlock/profile/banner), 剩余金币(current_coins) |
| 选中商品 | `select_sku` | 用户点击选中某个金币包或VIP套餐 | 商品ID(sku_id), 商品类型(type: coin/vip), 价格(price) |
| 发起支付 | `initiate_pay` | 点击确认支付 | 商品ID(sku_id), 金额(amount), 支付方式(pay_method) |
| 支付成功 | `pay_success` | 支付回调成功 | 订单ID(order_id), 金额(amount), 获得金币/天数(content_received) |

### 2.4 互动事件

| 事件显示名 | 事件名 (Event Name) | 触发时机 | 关键属性 (Properties) |
| :--- | :--- | :--- | :--- |
| 点赞 | `like_drama` | 用户点赞 | 剧集ID(drama_id), 集数(episode_num) |
| 收藏 | `favorite_drama` | 用户添加收藏 | 剧集ID(drama_id) |
| 分享 | `share_drama` | 用户点击分享 | 剧集ID(drama_id), 渠道(channel) |

## 3. 关键分析模型

### 3.1 转化漏斗 (Funnel Analysis)

#### 3.1.1 游客转注册漏斗
衡量免登录用户转化为注册用户的效率。
1. `app_start` (is_login=false)
2. `play_end` (累计观看 > 5集)
3. `view_checkout` / `view_fav` (触发登录引导点)
4. `login_success`

#### 3.1.2 付费转化漏斗
1. `click_unlock` (点击解锁)
2. `view_checkout` (浏览收银台)
3. `select_sku` (选择套餐/金币包)
4. `pay_success` (支付成功)

### 3.2 留存分析
*   **游客留存**：仅基于设备ID (Device ID) 追踪未登录用户的活跃留存。
*   **登录留存**：追踪注册用户的长期留存。

### 3.3 路径分析
*   **For You 页流向**：用户在 For You 页滑动多少次后会点击进入剧集详情，或者直接流失。

### 3.4 LTV (生命周期价值) 分析
衡量用户在生命周期内的累计付费金额，评估投放 ROI。
*   **计算逻辑**：以用户首次访问或注册为起始日期（Cohort），统计该群组在后续 N 天内的累计付费总额 / 初始用户数。
*   **应用场景**：
    *   **渠道 ROI 评估**：对比不同渠道来源用户的 LTV，优化投放策略。
    *   **付费习惯分析**：观察用户付费增长曲线（如首日付费占比 vs 长尾付费），调整运营节奏（首充优惠 vs 长期会员）。
