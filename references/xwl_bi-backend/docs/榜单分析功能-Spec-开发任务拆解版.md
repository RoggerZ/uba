# 榜单分析功能 Spec（开发任务拆解版）

## 1. 文档目标

本文档用于将 `xwl_bi` 的“榜单分析”能力从产品想法收敛为可执行的开发方案，覆盖：

- PRD 目标与范围
- 页面与交互方案
- API 设计
- 报表存储协议
- 后端改动点
- 前端改动点
- DB / 协议改动点
- 验收测试用例清单

本文档默认基于当前 `xwl_bi` 已有的分析架构推进：

- 行为分析统一入口：`/api/analysis/*`
- 报表保存统一模型：`report_table`
- 看板挂载统一机制：`panel + report_table`
- 已有分析类型：事件 / 留存 / 漏斗 / 路径 / 归因 / LTV

---

## 2. 背景与问题

当前系统已有较成熟的事件分析能力，支持：

- 事件筛选
- 用户筛选
- 用户分群
- 单维或多维分组
- 表格 / 图形展示

但缺少“榜单”这一高频分析形态。现有事件分析虽然可以输出表格，但不能完整承接以下需求：

1. 以排名为中心的表达方式
2. `Top N / Bottom N`
3. 榜单专用指标，如占比、升降、变化率
4. 专用的榜单看板卡片
5. 常用榜单模板
6. 统一的榜单配置和保存语义

因此需要新增独立的 `榜单分析` 功能。

---

## 3. 产品定位

### 3.1 功能名称

- 中文名：`榜单分析`
- 英文名：`Leaderboard Analysis`

### 3.2 所属模块

- 一级模块：`行为分析`

### 3.3 设计原则

1. 前台独立
   - 作为独立分析类型出现在行为分析菜单中

2. 后台复用
   - 底层复用现有事件分析的筛选、分群、聚合能力

3. 口径稳定
   - 排序、截断、占比、升降全部由后端统一计算

4. 看板友好
   - 榜单结果可保存、可加入看板、可做专用卡片

### 3.4 推荐方案

建议将榜单分析做成新的分析类型，而不是简单挂靠在事件分析页下的一个“表格排序模式”。

原因：

1. 榜单有独立的业务语义
   - 排名、TopN、占比、升降、长尾

2. 榜单需要独立模板与卡片

3. 榜单需要独立的报表类型与路由

4. 底层查询仍可复用 `Event` 分析引擎

---

## 4. 范围定义

### 4.1 本期范围

本期交付以下能力：

1. 新增 `榜单分析` 页面
2. 支持单维度榜单查询
3. 支持 `Top N / Bottom N`
4. 支持当前周期与对比周期
5. 支持热门榜 / 增长榜 / 下滑榜
6. 支持事件筛选、用户筛选、用户分群
7. 支持保存报表
8. 支持加入看板
9. 支持导出 Excel / CSV
10. 支持从榜单行钻取到事件分析
11. 支持系统预置榜单模板

### 4.2 非本期范围

本期不做：

1. 秒级实时榜单
2. 跨应用汇总榜单
3. 多维透视榜单
4. 订阅推送与异常告警
5. 对外分享页
6. 榜单历史快照中心

---

## 5. 用户场景

### 5.1 内容运营

- 最近 7 天最热剧集榜
- 最近 7 天播放增长最快剧集榜
- 最近 7 天互动最高剧集榜

### 5.2 增长运营

- 注册渠道榜
- 绑定渠道榜
- 渠道新增增长榜

### 5.3 产品分析

- 内容入口榜
- 搜索关键词榜
- 不同支付方式贡献榜

### 5.4 商业化分析

- 充值商品榜
- 剧集解锁榜
- 付费转化增长榜

---

## 6. 信息架构

### 6.1 菜单结构

建议新增：

- 行为分析
- 事件分析
- 留存分析
- 漏斗分析
- 智能路径分析
- 归因分析
- LTV分析
- 榜单分析

### 6.2 路由建议

- 前端路由：`/behavior-analysis/ranking/:id`

### 6.3 报表类型建议

- `rt_type = 11`

说明：

- 1：事件分析
- 2：留存分析
- 3：漏斗分析
- 4：智能路径分析
- 5：用户属性分析
- 6：归因分析
- 9：LTV分析
- 11：榜单分析

---

## 7. 页面与交互设计

### 7.1 页面结构

页面布局建议整体保持与现有行为分析页面一致，采用：

- 顶部标题操作区
- 左侧配置区
- 右侧结果区

即沿用现有 `event.vue` / `retention.vue` 这一类页面的主交互心智：

- 左边负责“定义分析对象和规则”
- 右边负责“查看结果并做结果态调整”

但榜单页不要求与现有事件分析页面完全 1:1 对齐，允许在以下部分做榜单化增强：

- 左侧配置区顶部可增加“榜单模板 / 榜单模块”快捷入口
- 右侧结果区顶部可增加“概览条 + 榜单工具栏 + 视图切换”
- 结果主体采用“横向条形图 + 榜单表格”的组合展现，并支持切换纯图/纯表

具体分为 3 个主区域：

1. 标题操作区
   - 页面标题
   - 已存报表名称
   - 导出
   - 已存报表列表
   - 保存报表
   - 加入看板

2. 左侧配置区
   - 榜单模板 / 榜单模块快捷区
   - 事件
   - 主指标
   - 榜单维度
   - 事件筛选
   - 用户筛选
   - 用户分群
   - 榜单模式
   - 排序字段
   - 排序方向
   - TopN
   - 长尾处理
   - 空值处理
   - 查询按钮
   - 重置按钮

3. 右侧结果区
   - 时间范围
   - 对比时间范围
   - 结果概览条
   - 视图切换：图表 / 表格 / 图+表
   - 横向条形图主视图
   - 榜单表格
   - 迷你趋势图
   - 行级钻取操作

### 7.2 与现有行为分析页面的对齐策略

#### 1. 保持对齐的部分

1. 保持 `split-pane` 双栏布局
2. 保持顶部标题 + 报表名称 + 导出 / 报表列表入口的使用习惯
3. 保持左侧配置、右侧结果的交互心智
4. 保持报表保存、看板挂载、结果导出这套公共动作

#### 2. 明确差异的部分

1. 榜单页结果区默认采用“图+表”组合，且图表位于主视区上方，视觉上以图为主
2. 榜单页允许切换 `图表 / 表格 / 图+表`
3. 榜单页允许在左侧配置区增加模板 / 模块快捷区
4. 榜单页允许在右侧结果区增加概览条、排序说明、升降说明

#### 3. 时间范围与对比时间的放置结论

对照现有 `事件分析` 页面，时间范围和对比时间范围更适合放在右侧结果区工具栏，而不是左侧配置区。

原因：

1. 现有事件分析已经形成这一交互习惯，用户成本更低
2. 时间范围更接近“结果重算条件”，放在结果区更符合操作语义
3. 榜单分析也会有结果态切换，放在右侧便于快速比较

因此本方案调整为：

- 左侧配置区不放 `date`、`compareDate`
- 右侧结果区顶部工具栏放 `date`、`compareDate`
- 左侧点击“查询”后进入初次结果，右侧调整时间后支持快捷重算

### 7.3 榜单模式

#### 1. 热门榜

- 默认按 `current_value desc`
- 用于查看当前周期头部内容 / 入口 / 渠道 / 商品

#### 2. 增长榜

- 默认按 `delta_rate desc`
- 若未传对比时间，前端自动推荐上一周期

#### 3. 下滑榜

- 默认按 `delta_rate asc`
- 用于识别明显下滑对象

### 7.4 榜单模板与榜单模块

#### 7.4.1 定义

1. 系统模板
   - 由系统预置的一组抽象配置
   - 目的是降低配置成本
   - 模板本身不直接绑定具体业务名称

2. 榜单模块
   - 由用户保存的榜单配置实例
   - 可以使用业务相关名称
   - 例如：`热门剧榜`、`高增长剧榜`、`搜索热词榜`

#### 7.4.2 设计原则

`xwl_bi` 本身应保持与业务解耦，因此系统层不直接写死强业务名称作为核心产品结构。

建议将快捷区拆成两类来源：

1. 系统模板
   - 使用抽象命名
   - 例如：
     - `按内容维度热门榜`
     - `按入口维度热门榜`
     - `按渠道维度增长榜`
     - `按关键词维度热门榜`
     - `按支付方式维度贡献榜`

2. 我的榜单模块
   - 来自用户已保存的榜单报表
   - 允许使用强业务名称
   - 例如：
     - `热门剧榜`
     - `重点渠道增长榜`
     - `TikTok 拉新榜`

#### 7.4.3 应用行为

1. 应用系统模板
   - 将模板里的预配置直接写入左侧配置区
   - 同时写入右侧结果区的默认时间范围配置
   - 默认不自动发起查询，由用户确认后点击 `查询`

2. 打开榜单模块
   - 直接恢复一份已保存的榜单报表配置
   - 可直接进入结果态

#### 7.4.4 首版实现建议

首版不单独新增 `leaderboard_module` 表。

直接复用现有 `report_table`：

- 系统模板：前端静态配置或后端固定模板配置
- 榜单模块：本质上是 `rt_type = 11` 的已保存榜单报表

因此，“保存为榜单模块”在首版的落地含义是：

- 保存为一份榜单报表
- 在榜单页快捷区可按“我的榜单模块”展示

### 7.5 查询配置项

#### 7.5.1 左侧配置区字段

1. 事件 `eventName`
   - 单选

2. 主指标 `metric`
   - 单选

3. 榜单维度 `groupBy`
   - 首版仅支持 1 个字段

4. 事件筛选 `whereFilter`
   - 选填

5. 用户筛选 `whereFilterByUser`
   - 选填

6. 用户分群 `userGroup`
   - 选填

7. 榜单模式 `rankingMode`
   - `hot`
   - `growth`
   - `decline`

8. 排序字段 `sortBy`
   - `current_value`
   - `delta_value`
   - `delta_rate`

9. 排序方向 `sortOrder`
   - `desc`
   - `asc`

10. 结果条数 `topN`
   - 默认 20

11. 长尾聚合 `includeOthers`
   - bool

12. 空值过滤 `excludeEmpty`
   - bool

#### 7.5.2 右侧结果区工具栏字段

1. 时间范围 `date`
   - 必填
   - 放在右侧结果区顶部工具栏

2. 对比时间 `compareDate`
   - 选填
   - 放在右侧结果区顶部工具栏

3. 快捷重算
   - 时间变更后支持重新查询

4. 结果导出
   - 与结果区绑定，便于导出当前视图结果

5. 视图切换 `displayMode`
   - `chart`
   - `table`
   - `chart_table`

### 7.6 结果展示方式

#### 7.6.1 默认展示策略

榜单分析结果区默认使用 `chart` 模式：

- 默认先展示横向条形图
- 由图形承担“快速理解榜单结构”的主职责

`chart_table` 作为增强模式保留，用于：

- 桌面端同时看图和精确数值
- 需要边看排名边看占比、变化值、变化率的分析场景

因此本方案的展示心智是：

1. 默认图表优先
2. 表格作为精确查看和导出的补充
3. 图+表是可选增强模式，不强制作为唯一默认态

#### 7.6.2 条形图要求

建议使用横向条形图作为榜单主图：

1. 横向更适合展示长维度值
2. 更符合“排行榜”从上到下阅读的习惯
3. 更适合直接映射 TopN 排名

图形建议规则：

1. 默认展示 TopN 当前值
2. 开启对比周期时，可支持双系列或当前值 + 变化标识
3. 排序顺序与表格完全一致

#### 7.6.3 视图模式

1. `chart`
   - 仅展示横向条形图

2. `table`
   - 仅展示榜单表格

3. `chart_table`
   - 同时展示条形图与榜单表格
   - 为桌面端增强模式

### 7.7 结果表格字段

统一字段建议：

- `rank`
- `group_display_value`
- `current_value`
- `share_rate`
- `prev_value`
- `delta_value`
- `delta_rate`
- `trend_points`

### 7.8 行级操作

每一行支持：

1. 查看事件分析
2. 查看用户列表
3. 复制维度值

#### 7.8.1 什么是“点击行钻取”

通俗讲，就是：

- 你在榜单里看到一行结果
- 想继续看这一行背后的明细
- 直接点击这一行，系统自动把这行对应的条件带到下一个分析页面

而不是让用户自己重新手动配置一遍。

#### 7.8.2 举例说明

例如你在“播放热门榜”里看到：

- 第 1 名：`霸总再爱我一次`

这时点击这一行，系统可以直接跳到事件分析页面，并自动带上：

- 事件：`WatchVideo`
- 时间范围：当前榜单使用的时间范围
- 筛选条件：`drama_code = 当前这行的 drama_code`

这样用户就可以继续看：

- 这部剧每天播放趋势如何
- 是哪些入口带来的播放
- 是哪些用户群在看

同样，如果点击“搜索热词榜”中的某个关键词，也可以自动跳到事件分析并带上：

- 事件：`Search`
- 筛选：`keyword = 当前点击的关键词`

### 7.9 原型草图

```text
┌──────────────────────────────────────────────────────────────────────────────────────────────┐
│ 榜单分析                                                     [导出] [已存报表] [保存] [加看板] │
├───────────────────────────────┬──────────────────────────────────────────────────────────────┤
│ 左侧配置区                    │ 右侧结果区                                                   │
│                               │                                                              │
│ [系统模板] [我的榜单模块]     │ 时间 [2026-03-25 ~ 2026-03-31]                              │
│ [按内容维度热门榜]            │ 对比 [2026-03-18 ~ 2026-03-24] [重算]                       │
│ [按入口维度热门榜]            │ 视图 [图表][表格][图+表]                                    │
│                               │                                                              │
│ 事件 [WatchVideo]             │ 总量 1376200 | Top20 占比 82.18% | 环比 +12.9%              │
│ 指标 [播放次数]               │                                                              │
│ 维度 [drama_code]             │ ┌──────────────── 横向条形图（TopN） ────────────────┐        │
│ 事件筛选 [筛选器]             │ 1  霸总再爱我一次   ████████████████████              │        │
│ 用户筛选 [筛选器]             │ 2  重生后我封神了   █████████████████                 │        │
│ 分群 [分群]                   │                                                              │
│                               │ 排名 | 维度值 | 当前值 | 占比 | 对比值 | 变化值 | 变化率 | 趋势 │
│ 模式 [热门榜]                 │ 行操作：[查看事件分析] [查看用户列表] [复制维度值]          │
│ 排序 [当前值] [降序]          │                                                              │
│ TopN [20]                     │                                                              │
│ 空值 [排除]                   │                                                              │
│ 长尾 [聚合其他]               │                                                              │
│ [查询] [重置]                 │                                                              │
└───────────────────────────────┴──────────────────────────────────────────────────────────────┘
```

### 7.10 “保存为报表”弹窗

榜单分析页左上或顶部操作区的 `保存为报表` 按钮，首版建议沿用现有 `AddReportTable.vue` 的交互模型。

弹窗建议包含以下内容：

1. 目标文件夹
   - 下拉选择
   - 支持“新增文件夹”

2. 目标看板
   - 下拉选择
   - 支持“新增看板”

3. 报表名称
   - 用户输入
   - 支持“检测是否有同名”

4. 当前时间配置预览
   - 显示当前 `date`
   - 显示当前 `compareDate`
   - 只读预览即可，不在弹窗里再次编辑

5. 展示模式
   - `图表`
   - `表格`
   - `图+表`

6. 图表类型
   - 首版默认 `横向条形图`
   - 先不开放太多图表类型

7. 看板卡片尺寸
   - `small`
   - `medium`
   - `large`

8. 备注
   - 多行文本

9. 操作按钮
   - `返回`
   - `添加/更新`

#### 7.10.1 弹窗目的

这个弹窗在首版不是一个单纯的“本地保存名字”动作，而是同时完成两件事：

1. 保存一份榜单报表配置
2. 指定这份报表要挂到哪个数据看板里

这与当前系统 `AddReportTable` 的交互习惯一致。

### 7.11 配置到数据看板

#### 7.11.1 首版交互结论

首版建议继续沿用当前系统习惯：

- 保存报表时，就同时要求选择 `目标文件夹` 和 `目标看板`
- 保存成功后，这个榜单报表会被挂到对应的数据看板中

也就是说，首版“保存报表”和“配置到数据看板”不是两套完全分离的动作，而是在同一个弹窗里一次完成。

#### 7.11.2 用户视角下的流程

1. 用户在榜单分析页完成配置并点击 `查询`
2. 用户确认结果可用，点击 `保存为报表`
3. 在弹窗中选择：
   - 放到哪个文件夹
   - 放到哪个看板
   - 用什么名称保存
   - 卡片用什么尺寸
   - 默认展示图表还是表格
4. 点击 `添加/更新`
5. 系统完成：
   - 保存 `report_table`
   - 绑定到对应 `panel`
   - 在数据看板中生成一张榜单卡片

#### 7.11.3 看板卡片展示建议

看板中的榜单卡片应至少支持：

1. 卡片标题
2. 时间范围摘要
3. Top3 / Top5 榜单摘要
4. 点击进入完整榜单分析页

#### 7.11.4 后续可优化项

如果后面用户明确需要“只保存报表，不立即挂到看板”，可以在后续版本拆分成两步：

1. `保存报表`
2. `加入看板`

但首版不建议拆开，避免偏离当前系统已有的操作习惯。

---

## 8. API Spec

### 8.1 新增接口

- `POST /api/analysis/LeaderboardList`

### 8.2 前端 API 命名

- `LeaderboardList(data)`

### 8.3 请求体

```json
{
  "appid": 612912612744648805,
  "eventName": "WatchVideo",
  "eventNameDisplay": "观看视频",
  "metric": {
    "metricType": "event_count",
    "valueField": "",
    "displayName": "播放次数"
  },
  "groupBy": ["drama_code"],
  "date": ["2026-03-25", "2026-03-31"],
  "compareDate": ["2026-03-18", "2026-03-24"],
  "whereFilter": {
    "filterType": "COMPOUND",
    "filts": [],
    "relation": "且"
  },
  "whereFilterByUser": {
    "filterType": "COMPOUND",
    "filts": [],
    "relation": "且"
  },
  "userGroup": [],
  "rankingMode": "hot",
  "sortBy": "current_value",
  "sortOrder": "desc",
  "topN": 20,
  "includeOthers": true,
  "excludeEmpty": true,
  "dashboard_config": {
    "folderId": 12,
    "pannelId": 88,
    "windowSize": "medium",
    "displayMode": "chart",
    "chartType": "horizontal_bar"
  }
}
```

### 8.4 请求字段定义

| 字段 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `appid` | int | 是 | 应用 ID |
| `eventName` | string | 是 | 事件名 |
| `eventNameDisplay` | string | 否 | 事件展示名 |
| `metric.metricType` | string | 是 | 指标类型 |
| `metric.valueField` | string | 否 | 求和/均值类字段 |
| `metric.displayName` | string | 否 | 指标展示名 |
| `groupBy` | string[] | 是 | 首版必须长度为 1 |
| `date` | string[] | 是 | 主时间范围 |
| `compareDate` | string[] | 否 | 对比时间范围 |
| `whereFilter` | object | 否 | 事件筛选 |
| `whereFilterByUser` | object | 否 | 用户筛选 |
| `userGroup` | int[] | 否 | 分群 ID 列表 |
| `rankingMode` | string | 是 | 榜单模式 |
| `sortBy` | string | 是 | 排序字段 |
| `sortOrder` | string | 是 | 排序方向 |
| `topN` | int | 是 | 返回条数 |
| `includeOthers` | bool | 否 | 是否聚合其他 |
| `excludeEmpty` | bool | 否 | 是否过滤空值 |
| `dashboard_config` | object | 否 | 保存看板相关配置 |

### 8.5 指标类型枚举

| `metricType` | 说明 |
| :--- | :--- |
| `event_count` | 事件次数 |
| `user_count` | 去重用户数 |
| `sum` | 某数值字段求和 |
| `avg` | 某数值字段均值 |
| `success_rate` | 成功率 |

### 8.6 返回体

```json
{
  "rows": [
    {
      "rank": 1,
      "group_key": "drama_code",
      "group_value": "D12345",
      "group_display_value": "霸总再爱我一次",
      "current_value": 182345,
      "share_rate": 0.1325,
      "prev_value": 151220,
      "delta_value": 31125,
      "delta_rate": 0.2058,
      "trend_points": [23000, 25000, 26000, 24000, 27000, 29000, 28345],
      "extra": {
        "drama_title": "霸总再爱我一次"
      }
    }
  ],
  "summary": {
    "total_value": 1376200,
    "row_count": 20,
    "top_n": 20,
    "others_value": 245300
  },
  "compareSummary": {
    "total_value": 1218700
  },
  "meta": {
    "eventName": "WatchVideo",
    "metricType": "event_count",
    "groupBy": "drama_code",
    "rankingMode": "hot",
    "sortBy": "current_value",
    "sortOrder": "desc",
    "hasCompare": true
  }
}
```

### 8.7 返回字段定义

| 字段 | 说明 |
| :--- | :--- |
| `rows` | 榜单行数据 |
| `rank` | 排名 |
| `group_key` | 维度字段名 |
| `group_value` | 维度原始值 |
| `group_display_value` | 展示值 |
| `current_value` | 当前周期值 |
| `share_rate` | 当前占比 |
| `prev_value` | 对比周期值 |
| `delta_value` | 差值 |
| `delta_rate` | 差异率 |
| `trend_points` | 趋势图数据 |
| `summary.total_value` | 当前总量 |
| `summary.others_value` | 长尾合并量 |
| `meta` | 查询元信息 |

### 8.8 错误码建议

| 错误码 | 说明 |
| :--- | :--- |
| `LeaderboardGroupByRequired` | 必须选择榜单维度 |
| `LeaderboardOnlyOneGroupBySupported` | 首版仅支持一个分组字段 |
| `LeaderboardMetricRequired` | 必须选择主指标 |
| `LeaderboardTopNTooLarge` | TopN 超过允许上限 |
| `LeaderboardCompareDateRequired` | 增长榜/下滑榜缺少对比周期 |
| `LeaderboardFieldNotAllowed` | 字段不允许用于出榜 |
| `LeaderboardResultTooLarge` | 查询结果规模过大 |

---

## 9. 报表存储协议

### 9.1 存储方式

沿用当前 `report_table`：

- `rt_type = 11`
- `data = JSON`

### 9.2 推荐保存结构

```json
{
  "eventName": "WatchVideo",
  "eventNameDisplay": "观看视频",
  "metric": {
    "metricType": "event_count",
    "valueField": "",
    "displayName": "播放次数"
  },
  "groupBy": ["drama_code"],
  "date": ["2026-03-25", "2026-03-31"],
  "compareDate": ["2026-03-18", "2026-03-24"],
  "whereFilter": {
    "filterType": "COMPOUND",
    "filts": [],
    "relation": "且"
  },
  "whereFilterByUser": {
    "filterType": "COMPOUND",
    "filts": [],
    "relation": "且"
  },
  "userGroup": [],
  "rankingMode": "hot",
  "sortBy": "current_value",
  "sortOrder": "desc",
  "topN": 20,
  "includeOthers": true,
  "excludeEmpty": true,
  "dashboard_config": {
    "folderId": 12,
    "pannelId": 88,
    "windowSize": "medium",
    "displayMode": "chart",
    "chartType": "horizontal_bar",
    "displayColumns": ["rank", "group_display_value", "current_value", "share_rate", "delta_rate"],
    "showTrend": true
  }
}
```

### 9.3 字段白名单建议

首版建议只开放以下适合榜单的字段：

- `drama_code`
- `drama_title`
- `drama_type`
- `source`
- `episode_index`
- `referrer`
- `provider`
- `payment_method`
- `product_id`
- `keyword`
- `video_id`

默认不开放：

- 原始长文本字段
- UUID / DistinctID
- JSON 字段
- 超高基数字段

---

## 10. 口径规则

### 10.1 排名规则

1. 排名必须由后端生成
2. 前端不得根据展示列再排序
3. 排名基于 `sortBy + sortOrder` 的最终结果

### 10.2 占比规则

- `share_rate = current_value / total_value`

### 10.3 对比规则

- `delta_value = current_value - prev_value`
- `delta_rate = delta_value / prev_value`
- 若 `prev_value = 0`，则 `delta_rate = null`

### 10.4 长尾规则

- 若 `includeOthers = true`
- TopN 之外所有行合并为 `其他`

### 10.5 空值规则

- 空维度值统一展示为 `（空值）`
- 若 `excludeEmpty = true`，则空值不参与排序和占比

### 10.6 榜单模式默认排序

| 模式 | 默认排序字段 | 默认排序方向 |
| :--- | :--- | :--- |
| `hot` | `current_value` | `desc` |
| `growth` | `delta_rate` | `desc` |
| `decline` | `delta_rate` | `asc` |

---

## 11. 后端改动点

### 11.1 控制器层

新增：

- `controller/behavior_analysis_controller.go`
  - 新增 `LeaderboardList(ctx *fiber.Ctx) error`

作用：

- 统一处理榜单查询请求
- 走现有 `/api/analysis` 分析体系

### 11.2 路由层

新增：

- `router/analysis.go`
  - 注册 `LeaderboardList`

建议接口：

- `POST /api/analysis/LeaderboardList`

### 11.3 分析命令层

新增：

- `platform-basic-libs/service/analysis/interface.go`
  - 新增 `LeaderboardCommand`
  - 注册 `NewLeaderboard`

### 11.4 分析实现层

新增文件建议：

- `platform-basic-libs/service/analysis/leaderboard.go`

职责：

1. 解析请求参数
2. 校验榜单配置合法性
3. 复用现有事件分析筛选 / 分群 / 日期过滤逻辑
4. 生成按单维聚合的 SQL
5. 计算当前值
6. 计算对比周期值
7. 计算差值与变化率
8. 执行排序与 TopN 截断
9. 计算占比
10. 处理 `其他` 聚合
11. 返回统一结果结构

### 11.5 请求结构层

新增：

- `platform-basic-libs/request/Model.go`
  - 新增 `LeaderboardReqData`
  - 新增 `LeaderboardMetric`

### 11.6 响应结构层

如需要强类型化，建议新增：

- `platform-basic-libs/response/Model.go`
  - `LeaderboardRow`
  - `LeaderboardSummary`
  - `LeaderboardResponse`

### 11.7 元数据层

可选增强：

- 复用现有 `GetConfigs`
- 增加榜单字段白名单过滤逻辑

目的：

- 防止无意义字段用于出榜

### 11.8 SQL 与性能层

后端需要重点处理：

1. 单维分组聚合
2. 排序字段选择
3. TopN 截断
4. 对比期 join / merge
5. 聚合 `其他`
6. 防高基数与大结果集

建议约束：

1. `topN <= 100`
2. 首版仅允许一个 `groupBy`
3. 查询返回不超过 `1000` 行中间结果
4. 高基数字段走白名单控制

---

## 12. 前端改动点

### 12.1 路由

修改：

- `vue/src/utils/router.js`
  - 增加 `ranking/:id`

- `vue/src/store/modules/user.js`
  - 动态路由补充 `榜单分析`

### 12.2 API 层

修改：

- `vue/src/api/analysis.js`
  - 新增 `LeaderboardList(data)`

### 12.3 页面层

新增：

- `vue/src/views/behavior-analysis/ranking.vue`

职责：

1. 沿用现有行为分析的双栏 `split-pane` 布局
2. 左侧展示查询表单
3. 管理系统模板与我的榜单模块
4. 右侧结果区顶部承载 `date`、`compareDate`、`displayMode` 与结果工具栏
5. 发起榜单查询
6. 保存榜单报表 / 榜单模块
7. 承接钻取逻辑

### 12.4 结果组件层

建议新增：

- `vue/src/views/behavior-analysis/components/LeaderboardResult.vue`
- `vue/src/views/behavior-analysis/components/LeaderboardBarChart.vue`
- `vue/src/views/dashboard/components/analysis/LeaderboardCard.vue`

职责：

1. 榜单横向条形图展示
2. 榜单表格展示
3. 榜单概览条展示
4. 图表 / 表格 / 图+表切换
5. 榜单卡片展示
6. 导出逻辑
7. 行点击钻取

### 12.5 报表列表层

修改：

- `vue/src/views/behavior-analysis/components/ReportTableList.vue`

需要补充：

1. `tableTypeMap[11] = '榜单分析'`
2. `routerMap[11] = '/behavior-analysis/ranking/'`

### 12.6 看板层

修改：

- `vue/src/views/dashboard/index.vue`

需要补充：

1. 识别 `rt_type = 11`
2. 加载榜单卡片渲染器
3. 支持榜单卡片下载导出

### 12.7 报表设置弹窗层

修改：

- `vue/src/views/behavior-analysis/components/AddReportTable.vue`

需要补充：

1. `isLeaderboardReport(rtType)` 判定
2. 榜单报表保存配置解析
3. 榜单卡片尺寸与显示列设置
4. 榜单默认视图模式与图表类型设置

### 12.8 工具层

建议新增：

- `vue/src/utils/leaderboard-dashboard.js`

职责：

1. 默认配置规范化
2. 榜单 dashboard_config 构造
3. 显示列与窗口尺寸默认值
4. 榜单模板默认值
5. 视图模式与图表类型默认值

---

## 13. DB / 协议改动点

### 13.1 数据库

本期不强制新增物理表。

沿用：

- `report_table`
- `panel`
- 现有行为分析来源表

### 13.2 `report_table` 协议扩展

需要扩展的是：

1. 新增 `rt_type = 11`
2. `data` 中新增榜单配置 JSON 协议

首版约束：

1. 不新增单独的 `leaderboard_module` 表
2. “榜单模块”本质上仍是 `rt_type = 11` 的已保存榜单报表
3. 系统模板不落库，使用前端固定配置或后端固定模板配置

### 13.3 请求协议新增

新增：

- `LeaderboardReqData`

### 13.4 响应协议新增

新增：

- `rows`
- `summary`
- `compareSummary`
- `meta`

### 13.5 可选的后续 DB 优化

若后续榜单查询量很大，可在下一期考虑：

1. 预聚合明细表
2. 榜单物化视图
3. 榜单查询缓存表
4. 榜单模板配置表

本期不建议提前引入，避免过度设计。

### 13.6 看板挂载相关协议说明

首版榜单保存沿用现有报表保存协议，因此需要在 `dashboard_config` 中明确以下字段：

| 字段 | 说明 |
| :--- | :--- |
| `folderId` | 目标文件夹 |
| `pannelId` | 目标看板 |
| `windowSize` | 看板卡片尺寸 |
| `displayMode` | 默认展示模式 |
| `chartType` | 默认图表类型 |

其中：

- `folderId`、`pannelId`、`windowSize` 对应“配置到数据看板”
- `displayMode`、`chartType` 对应看板卡片默认展示方式

---

## 14. 开发任务拆解

### 14.1 后端任务拆解

#### P0

1. 定义 `LeaderboardReqData`
2. 新增 `LeaderboardCommand`
3. 新增 `LeaderboardList` 控制器
4. 新增 `leaderboard.go`
5. 实现单维聚合查询
6. 实现排序与 TopN
7. 实现对比周期合并
8. 实现占比与变化率
9. 实现 `其他` 聚合
10. 返回统一协议

#### P1

1. 维度白名单校验
2. 高基数限制
3. 默认对比周期补全
4. 榜单模板后端配置化

### 14.2 前端任务拆解

#### P0

1. 新增路由入口
2. 新增 `ranking.vue`
3. 新增 `LeaderboardList` API
4. 新增榜单查询表单
5. 新增系统模板 / 我的榜单模块快捷区
6. 新增榜单结果组件
7. 新增横向条形图主视图
8. 支持 `图表 / 表格 / 图+表` 切换
9. 支持报表保存
10. 支持加入看板
11. 支持导出
12. 支持钻取到事件分析

#### P1

1. 榜单卡片
2. 榜单显示列自定义
3. 榜单结果趋势迷你图
4. 榜单模板后端配置化

#### 14.2.1 保存报表弹窗专项任务

1. 榜单类型接入 `AddReportTable.vue`
2. 弹窗中支持展示当前时间配置预览
3. 弹窗中支持设置 `displayMode`
4. 弹窗中支持设置 `chartType`
5. 弹窗中支持设置 `windowSize`
6. 保存时正确写入 `folderId`、`pannelId`
7. 保存时正确写入 `displayMode`、`chartType`

#### 14.2.2 配置到数据看板专项任务

1. 榜单保存成功后正确挂到目标看板
2. 看板中正确渲染榜单卡片
3. 榜单卡片正确使用保存时配置的 `windowSize`
4. 榜单卡片正确使用保存时配置的 `displayMode`
5. 榜单卡片点击可回到完整榜单分析页

### 14.3 测试任务拆解

#### P0

1. 热门榜正确性验证
2. 增长榜正确性验证
3. 下滑榜正确性验证
4. 对比周期口径验证
5. TopN 截断验证
6. `其他` 聚合验证
7. 空值过滤验证
8. 模板应用与榜单模块恢复验证
9. 图表 / 表格 / 图+表切换验证
10. 导出一致性验证
11. 保存报表与重新打开验证
12. 看板挂载验证

---

## 15. 验收测试用例清单

### 15.1 查询能力

1. 使用 `ViewDramaDetail + drama_title + event_count` 可以返回热门剧榜
2. 使用 `Search + keyword + event_count` 可以返回热词榜
3. 使用 `Purchase + payment_method + sum(amount)` 可以返回支付方式榜

### 15.2 排序能力

1. `hot` 模式默认按当前值降序
2. `growth` 模式默认按变化率降序
3. `decline` 模式默认按变化率升序
4. 手动切换排序字段后结果正确

### 15.3 对比能力

1. 存在 `compareDate` 时返回 `prev_value`
2. `delta_value` 计算正确
3. `delta_rate` 计算正确
4. `prev_value = 0` 时 `delta_rate` 返回空值

### 15.4 TopN 与长尾

1. `topN = 10` 只返回前 10 条
2. `includeOthers = true` 时返回 `其他`
3. `其他` 的值等于 TopN 之外剩余值汇总
4. TopN + 其他的占比加总为 100%

### 15.5 空值规则

1. `excludeEmpty = true` 时不返回空维度值
2. `excludeEmpty = false` 时空值显示为 `（空值）`

### 15.6 报表能力

1. 榜单可以保存为已存报表
2. 已存报表列表能显示 `榜单分析`
3. 点击已存报表可以重新打开榜单页面
4. 榜单可以加入看板

### 15.7 模板与模块能力

1. 应用系统模板后，配置区字段被正确填充
2. 应用系统模板默认不自动查询
3. 打开“我的榜单模块”可以恢复一份已保存榜单配置
4. 榜单模块名称允许使用业务相关名称
5. 系统模板名称保持抽象，不直接绑定某个具体业务

### 15.8 图表与表格展示能力

1. 默认结果模式为 `图+表`
2. 可以切换为纯图模式
3. 可以切换为纯表模式
4. 条形图顺序与表格顺序一致
5. 对比周期开启后图表与表格口径一致

### 15.9 导出能力

1. 榜单 Excel 导出字段完整
2. 导出顺序与页面顺序一致
3. 导出值与页面展示值一致

### 15.10 钻取能力

1. 点击榜单行可跳转事件分析
2. 跳转后自动带上事件、时间范围、当前维度筛选值

### 15.11 性能与边界

1. 常规 Top20 榜单查询在合理时间内返回
2. 高基数字段被正确拦截
3. 非法 `groupBy` 数量被正确拦截
4. `topN > 100` 被正确拦截

---

## 16. 推荐实现顺序

### 第一阶段：协议与后端主链路

1. 定义请求结构
2. 实现后端榜单查询
3. 打通接口返回

### 第二阶段：前端查询与结果页

1. 新增榜单页面
2. 接入查询 API
3. 渲染榜单结果

### 第三阶段：报表与看板

1. 保存报表
2. 已存报表列表接入
3. 看板卡片接入

### 第四阶段：模板与体验优化

1. 榜单模板
2. 结果趋势迷你图
3. 长尾与空值体验优化

---

## 17. 当前推进建议

建议按以下顺序推进：

1. 先完成后端协议和查询实现
2. 再完成前端基础页和结果展示
3. 然后补报表保存和看板卡片
4. 最后补模板与交互细节

当前 worktree 建议作为榜单功能专用开发现场：

- 工作目录：`C:\Users\admin\Documents\src\xwl_bi-ranking`
- 分支：`codex/ranking-spec`
