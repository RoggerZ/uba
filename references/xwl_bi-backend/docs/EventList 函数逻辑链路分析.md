# EventList 函数逻辑链路分析

本文档详细梳理了后端接口 `EventList` 的逻辑链路，从 Controller 层接收请求到最终生成 ClickHouse SQL 并返回数据的全过程。

## 1. 入口层 (Controller)

- **文件位置**: `controller/behavior_analysis_controller.go`
- **函数**: `EventList(ctx *fiber.Ctx)`

**逻辑流**:
1.  接收前端 POST 请求。
2.  调用 `analysis.NewAnalysisByCommand(analysis.EventComand, ctx.Body())` 初始化分析实例。
3.  调用 `analysis.GetAnalysisRes(i)` 获取分析结果。
4.  返回 JSON 响应。

```go
// 代码片段
func (this BehaviorAnalysisController) EventList(ctx *fiber.Ctx) error {
    i, err := analysis.NewAnalysisByCommand(analysis.EventComand, ctx.Body())
    // ...
    res, err := analysis.GetAnalysisRes(i)
    // ...
    return this.Success(ctx, response.SearchSuccess, res)
}
```

## 2. 调度与初始化 (Service Factory)

- **文件位置**: `platform-basic-libs/service/analysis/interface.go`
- **函数**: `NewAnalysisByCommand`, `NewEvent`

**逻辑流**:
1.  `NewAnalysisByCommand` 根据传入的 `EventComand` (枚举值 4) 查找映射表 `commandMap`。
2.  匹配到 `NewEvent` 函数并执行。

- **文件位置**: `platform-basic-libs/service/analysis/event.go`
- **函数**: `NewEvent(reqData []byte)`

**初始化逻辑**:
1.  **反序列化**: 将请求体解析为 `EventReqData` 结构。
2.  **校验**:
    - 日期范围 (`Date`) 必须包含开始和结束时间。
    - 指标数组 (`ZhibiaoArr`) 不能为空。
    - 分组字段 (`GroupBy`) 不能为空字符串。
3.  **用户分群处理**: 调用 `utils.GetUserGroupSqlAndArgs` 解析 `UserGroup` 参数，生成对应的 SQL 片段 (通常是 `IN` 子查询) 和参数。
4.  返回 `*Event` 实例。

## 3. 执行查询 (Service Execution)

- **文件位置**: `platform-basic-libs/service/analysis/event.go`
- **函数**: `GetList()`

**主逻辑**:
1.  调用 `this.GetExecSql()` 生成最终的 ClickHouse SQL 语句和参数列表。
2.  记录 SQL 日志 (`logs.Logger`).
3.  执行查询: `db.ClickHouseSqlx.Query(SQL, args...)`.
4.  **结果集处理**:
    - 动态获取列名 (`rows.Columns()`).
    - 遍历结果集，将每一行数据映射为 `map[string]interface{}`.
5.  **返回结构**:
    - `alldata`: 查询结果列表。
    - `use_group`: 是否使用了分组。
    - `len`: 指标数量。
    - `groupby`: 分组字段列表。
    - `eventNameDisplayArr`: 指标显示名称列表。

## 4. SQL 生成核心逻辑 (SQL Generation)

这是最复杂的部分，位于 `platform-basic-libs/service/analysis/event.go` 的 `GetExecSql` 及其辅助函数中。

### 4.1 全局过滤条件构建
`GetExecSql` 首先构建适用于所有指标的公共过滤条件：
1.  **通用筛选 (`WhereFilter`)**: 调用 `utils.GetWhereSql` 生成 WHERE 子句 (e.g., `os = 'iOS'`).
2.  **用户属性筛选 (`WhereFilterByUser`)**: 调用 `getUserfilterSqlArgs` 生成基于用户表的筛选条件 (e.g., `xwl_distinct_id IN (...)`).
3.  **日期范围 (`Date`)**: 调用 `GetFilterDateSql` 生成时间范围限制 (`xwl_part_date >= ... AND xwl_part_date <= ...`).

### 4.2 指标 SQL 构建 (Loop)
遍历 `req.ZhibiaoArr`，为每个指标调用 `getSqlByZhibiao` 生成子查询 SQL。

**`getSqlByZhibiao` 逻辑**:
1.  **事件筛选**: `whereInZhibiaoEvent` 限制特定的事件名 (`xwl_part_event IN ('EventName')`).
2.  **时间粒度分组**: `GetGroupDateSql` 根据 `WindowTimeFormat` (按天/小时/月等) 生成 `date_group` 字段的格式化 SQL。
3.  **属性分组**: `GetGroupSql` 处理 `GroupBy` 字段。
4.  **指标计算 (`zhibiao.Typ`)**:
    - **普通指标 (`Zhibiao`)**:
        - 调用 `utils.CountTypMap` 获取聚合函数 (如 `count(1)`, `uniq(xwl_distinct_id)`).
        - 生成 `amount` 列。
    - **公式指标 (`Formula`)**:
        - 处理两个指标的四则运算 (+, -, *, /).
        - **除数无分组处理 (`DivisorNoGrouping`)**: 如果开启，使用 ClickHouse 的 `WITH` 子句先计算不分组的分母，再通过 `mapValues` 和 `indexOf` 映射回分子。
        - 格式化输出 (百分比、两位小数等).
5.  **子查询组装**:
    ```sql
    SELECT [GroupCols], [Amount], [EventName], [SerialNumber]
    FROM xwl_event_appid
    PREWHERE [GlobalFilters] AND [EventFilter]
    GROUP BY [GroupCols]
    ORDER BY date_group
    ```

### 4.3 最终 SQL 组装
`GetExecSql` 将所有指标生成的 SQL 子查询通过 `UNION ALL` 连接，并在外层包裹排序。

```sql
SELECT * FROM (
    (SubQuery_Metric_1)
    UNION ALL
    (SubQuery_Metric_2)
    ...
) ORDER BY [GroupByCols], serial_number
```

## 5. 关键辅助模块

- **Count Utils**: `platform-basic-libs/service/analysis/utils/count.go`
  - 定义了 `CountTypMap`，映射前端选择的聚合类型 (总次数、触发用户数、人均次数等) 到具体的 SQL 实现。
  - 例如 `A1` (AllCount) 映射为 `count()` 或 `count(col)`。

- **Where Utils**: `platform-basic-libs/service/analysis/utils/sql.go`
  - `GetWhereSql`: 递归解析前端传递的复杂筛选树 (COMPOUND/SIMPLE)，生成 SQL WHERE 子句。

## 6. 数据流总结

1.  **Frontend**: 发送 JSON 请求 (含指标、分组、筛选条件)。
2.  **Controller**: 接收并转发。
3.  **Service (NewEvent)**: 解析请求，准备用户分群数据。
4.  **Service (GetExecSql)**:
    - 拼接 Global WHERE (时间、属性、用户)。
    - 循环 Metric: 拼接 Event Filter + Group By + Aggregation -> SubQuery.
    - Union All SubQueries.
5.  **DB**: ClickHouse 执行查询。
6.  **Service (GetList)**: 格式化结果为 JSON 对象。
7.  **Response**: 返回给前端渲染表格/图表。