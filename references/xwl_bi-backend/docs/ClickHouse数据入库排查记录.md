# ClickHouse 数据入库“假成功”问题排查记录

## 1. 问题现象
在 `platform-basic-libs/service/consumer_data/reportdata2ck.go` 服务中，日志明确打印 ClickHouse 入库成功，且显示插入了 6 条数据。
```log
CK入库成功，{"所花时间": "...", "数据长度为": 6}
```
但通过 SQL 查询验证时：
```sql
select count() from xwl_event41 where xwl_part_event = '用户注册'
```
查询结果为 0，数据并未在表中持久化。

## 2. 排查过程
1. **代码单步调试**：确认内存中的 `rowsMap` 在执行 SQL 前确实包含 6 条完整数据。
2. **增加调试日志**：在 `reportdata2ck.go:115` 附近增加日志，打印生成的 `insertSql` 和具体的 `rowsMap` 内容。
3. **分析表结构**：
   检查 `xwl_event41` 的建表语句，重点关注 TTL 设置：
   ```sql
   create table xwl_event41 (
       xwl_part_date DateTime default now(),
       ...
   )
   engine = MergeTree PARTITION BY toYYYYMM(xwl_part_date)
   ORDER BY (toYYYYMM(xwl_part_date), xwl_part_event)
   TTL xwl_part_date + toIntervalMonth(1) -- 关键点：数据保留1个月
   SETTINGS index_granularity = 8192;
   ```

## 3. 原因分析
问题根源在于 **ClickHouse 的 TTL (Time To Live) 机制** 与 **代码中的默认时间处理逻辑** 发生冲突。

1. **TTL 策略**：
   表定义了 `TTL xwl_part_date + toIntervalMonth(1)`。这意味着 ClickHouse 会自动删除 `xwl_part_date` 早于（当前时间 - 1个月）的数据行。

2. **默认解析逻辑**：
   在 `platform-basic-libs/sinker/parse/fastjson.go` 中，当上报的 JSON 数据中缺少日期字段或字段为 null 时，`getDefaultDateTime` 函数会返回 `Epoch` 时间（即 `1970-01-01 00:00:00 UTC`）。

3. **问题复现流程**：
   - 上报的数据中缺失 `xwl_part_date` 字段。
   - Go 程序将其解析默认值 `1970-01-01`。
   - 程序将数据发送给 ClickHouse，操作本身是合法的，ClickHouse 返回成功。
   - ClickHouse 接收数据后，立即触发 TTL 检查：`1970-01-01` + 1个月 < `2026年`。
   - 数据被判定为过期，ClickHouse 立即将其丢弃/删除。
   - 结果：日志显示“入库成功”，但数据库查不到数据。

## 4. 解决方案
在数据入库前进行拦截修正。在 `platform-basic-libs/service/consumer_data/reportdata2ck.go` 中添加逻辑：

当检测到 `xwl_part_date` 字段的值为 `Epoch` (1970-01-01) 时，强制将其更新为当前系统时间 `time.Now()`。

**代码修改示例**：
```go
// platform-basic-libs/service/consumer_data/reportdata2ck.go

val := parser.GetValueByType(obj.FastjsonMetric, dim)
// 增加特殊处理逻辑
if dim.Name == "xwl_part_date" {
    // 如果解析出的时间是 1970-01-01，说明原数据缺失该字段
    if t, ok := val.(time.Time); ok && t.Equal(parser.Epoch) {
        val = time.Now() // 修正为当前时间
    }
}
rowArr = append(rowArr, val)
```

通过此修改，确保了入库数据的 `xwl_part_date` 有效，避免了被 ClickHouse TTL 机制误删。

## 5. 新建应用时数据保留月数报错

### 问题描述
在 `vue/src/views/app/index.vue` 新增应用时，如果填写的“数据保留月数”大于 1，后端会报 500 错误：
```
model.App.SaveMonth: readUint64: unexpected character: \ufffd, error found in #10 byte of ...
```

### 原因分析
前端传递参数时，尽管 `<el-input>` 设置了 `type="number"`，但 Vue 的 `v-model` 默认将输入值作为字符串处理（例如 `"9"`）。
后端 `model.App` 结构体定义 `SaveMonth` 为 `int` 类型，但使用的 JSON 解析库（如 `jsoniter`）在严格模式下或处理某些类型转换时，未能将字符串 `"9"` 自动转换为数字类型，导致解析失败。

### 解决方案
#### 1. 前端修复
在提交表单前，手动将 `save_mouth` 字段转换为数字类型。
```javascript
// vue/src/views/app/index.vue

async addForm() {
  this.form.save_mouth = Number(this.form.save_mouth) // 显式转换为数字
  const res = await Create(this.form)
  // ...
}
```

#### 2. SQL 修正数据保留月数
如果需要修改已创建表的 TTL（数据保留月数），可以使用以下 ClickHouse SQL：

```sql
-- 假设应用的 id 为 <AppId> (注意是整数 ID，不是 UUID)
-- 将数据保留时间修改为 <NewMonth> 个月
ALTER TABLE xwl_event<AppId> MODIFY TTL xwl_part_date + toIntervalMonth(<NewMonth>);
```

示例：将 ID 为 41 的应用数据保留时间改为 6 个月：
```sql
ALTER TABLE xwl_event41 MODIFY TTL xwl_part_date + toIntervalMonth(6);
```
