# signup_time 类型错误分析

## 1. 问题描述
用户上报 `signup_time` (或 `RegisterTime`) 字段时报错，提示：
`类型错误，正确类型为字符串类型，上报类型为时间类型("2025-12-17 14:50:50")`

## 2. 代码分析

### 2.1 数据上报端
在 `tools/event-reporter/main.go` 中，`RegisterTime` 被格式化为字符串发送：
```go
loginProps["RegisterTime"] = user.StartTime.Format("2006-01-02 15:04:05")
```
发送的数据实际上是 JSON 字符串，例如 `"2025-12-17 14:50:50"`。

### 2.2 类型检测逻辑
在 `platform-basic-libs/sinker/parse/fastjson.go` 的 `FjDetectType` 函数中：
```go
	case fastjson.TypeString:
		typ = String
		if val, err := v.StringBytes(); err == nil {
			if _, err := parseInLocation(util.Bytes2str(val), time.Local); err == nil {
				typ = DateTime
			}
		}
```
当字符串符合时间格式 (`2006-01-02 15:04:05`) 时，解析器会将其识别为 `DateTime` 类型，而不是普通的 `String` 类型。

### 2.3 校验逻辑
在 `cmd/sinker/action/action.go` 的 `AddTableColumn` 函数中：
```go
			if reportType != column.Type {
				if !(reportType == parser.Int && column.Type == parser.Float) && !(reportType == parser.Float && column.Type == parser.Int) {
                    // 报错...
                }
            }
```
这里进行了严格的类型比对。
- **ClickHouse 表结构 (`column.Type`)**: `String` (字符串类型)
- **上报数据检测类型 (`reportType`)**: `DateTime` (时间类型)

由于 `DateTime != String`，且不在豁免列表（Int/Float 互转）中，因此抛出错误。

## 3. 根本原因
Sink 端的类型校验逻辑过于严格，没有考虑到 **时间格式的字符串本质上也是字符串**。当目标列类型为 `String` 时，应该允许写入时间格式的字符串。

## 4. 解决方案
修改 `cmd/sinker/action/action.go` 中的类型校验逻辑，增加对 `reportType == DateTime` 且 `column.Type == String` 的兼容处理。

```go
if reportType != column.Type {
    // 允许 DateTime 写入 String 列
    if reportType == parser.DateTime && column.Type == parser.String {
        // 合法，不做处理
    } else if !(reportType == parser.Int && column.Type == parser.Float) && !(reportType == parser.Float && column.Type == parser.Int) {
        // 报错
    }
}
```

## 5. 补充：为什么 signup_time 字段被判定为字符串类型

ClickHouse 中的列类型是由**第一条到达的数据**决定的。`signup_time` 字段之所以被创建为 `String` 类型，是因为：

1.  **Guest 用户的上报逻辑**：
    在 `tools/event-reporter/main.go` 中，当模拟 Guest 用户（未注册用户）时，`RegisterTime` 字段被显式赋值为空字符串 `""`。
    ```go
    // tools/event-reporter/main.go
    } else {
        // ...
        loginProps["RegisterTime"] = "" 
    }
    ```

2.  **Sinker 的类型推断**：
    在 Sinker 处理数据时，如果是空字符串 `""`，`FjDetectType` 函数尝试将其解析为时间会失败（空字符串不符合时间格式），因此回退判定为普通 `String` 类型。

3.  **结果**：
    如果第一条到达的数据来自 Guest 用户，ClickHouse 表中的 `signup_time` 列就会被创建为 `String` 类型。
    后续当注册用户上报带有时间格式（如 `"2025-12-17..."`）的数据时，解析器将其识别为 `DateTime`，从而导致与表中已存在的 `String` 类型不匹配，触发报错。
