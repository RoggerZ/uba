# LoadPropQuotas 函数及属性查询逻辑分析

本文档针对 `controller/behavior_analysis_controller.go` 中 `LoadPropQuotas` 函数的 SQL 查询逻辑进行分析，并回答关于自定义属性支持及 `attribute_type` 使用场景的问题。

## 1. 问题背景

用户发现在 `LoadPropQuotas` 函数中执行的 SQL 语句包含以下条件：
```sql
select attribute_name, show_name, data_type 
from attribute 
where app_id = 42 
  and (status = 1 or attribute_type = 1) 
  and attribute_source = 2 
  ...
```
用户疑问：
1.  为何事件分析似乎不支持自定义属性（认为 `attribute_type = 1` 限制了查询）？
2.  这是设计如此还是 Bug？
3.  `attribute_type = 2` 在哪些业务逻辑中使用？

## 2. 代码与数据库结构分析

### 2.1 数据库表结构 (`attribute`)
根据 `cmd/init_app/mysql/bi.sql` 中的定义：
```sql
CREATE TABLE `attribute`  (
  ...
  `attribute_type` tinyint(4) NULL DEFAULT 1 COMMENT '默认为1 （1为预置属性，2为自定义属性）',
  `attribute_source` tinyint(4) NULL DEFAULT 1 COMMENT '默认为1 （1为用户属性，2为事件属性）',
  `status` tinyint(4) NULL DEFAULT 0 COMMENT '是否显示 0为不显示 1为显示 默认不显示',
  ...
)
```
-   **attribute_type**: `1` = 预置属性 (Preset), `2` = 自定义属性 (Custom).
-   **status**: `0` = 不显示 (Hidden), `1` = 显示 (Visible).
-   **attribute_source**: `2` = 事件属性 (Event Attribute).

### 2.2 SQL 查询逻辑解析
`LoadPropQuotas` 用于获取事件分析中的可选属性列表。查询条件为：
```sql
(status = 1 or attribute_type = 1)
```
该逻辑的含义是：
1.  **如果是预置属性 (`attribute_type = 1`)**：无论 `status` 是什么，**总是显示**（或者说总是被查询出来）。
2.  **如果是自定义属性 (`attribute_type != 1`，即 `2`)**：必须满足 **`status = 1`** 才能被查询出来。

## 3. 问题解答

### 3.1 为什么事件分析不支持使用自定义属性？
**答：事件分析实际上是支持自定义属性的。**

用户之所以感觉“不支持”，通常是因为自定义属性的 **默认状态 (`status`) 为 0 (不显示)**。
根据 SQL 定义 `DEFAULT 0`，当新的自定义字段通过数据上报自动入库时，其 `status` 默认为 0。
因此，在 `LoadPropQuotas` 的查询条件 `(status = 1 or attribute_type = 1)` 下，这些未被手动开启显示的自定义属性会被过滤掉。

**解决方法**：
需要在“元数据管理”或“属性管理”页面中，找到对应的自定义属性，将其状态修改为“显示” (`status = 1`)，之后该属性即可在事件分析中被选用。

### 3.2 这是设计如此还是 Bug？
**答：这是有意为之的设计 (By Design)。**

*   **设计意图**：避免系统自动采集的大量临时或脏数据字段直接暴露在分析选项中，污染分析界面。只有经过管理员确认并“开启显示”的自定义属性，才会被视为有效业务属性供分析使用。
*   **预置属性**：通常是系统核心字段（如 OS, AppVersion 等），因此默认总是可见。

### 3.3 在哪些业务逻辑会使用到 attribute_type = 2 的查询？

`attribute_type = 2` (自定义属性) 主要在以下场景涉及：

1.  **数据接入 (Sinker)**:
    *   在 `cmd/sinker/action/action.go` 中，当处理上报数据时，系统会检查字段名是否属于 `SysColumn` (系统预置字段)。
    *   如果不属于预置字段，则标记为 `attribute_type = 2` (CustomAttribute) 并写入数据库。
    ```go
    // cmd/sinker/action/action.go
    if _, ok := parser.SysColumn[columnName]; ok {
        attributeType = PresetAttribute // 1
    } else {
        attributeType = CustomAttribute // 2
    }
    ```

2.  **属性查询过滤 (LoadPropQuotas)**:
    *   即当前讨论的 `controller/behavior_analysis_controller.go`。
    *   逻辑：`where ... (status = 1 or attribute_type = 1)`。
    *   此处虽然没有显式写 `attribute_type = 2`，但该逻辑通过排除法确立了 **自定义属性必须 `status=1`** 的规则。

3.  **元数据管理 (MetaDataController)**:
    *   在 `platform-basic-libs/service/meta_data/meta_data.go` 的 `AttrManager` 方法中，会查询所有属性并返回给前端管理界面。前端可能会根据 `attribute_type` 展示不同的图标或分类（例如区分预置属性和自定义属性）。

## 4. 总结
*   **现象**：LoadPropQuotas 过滤掉了部分自定义属性。
*   **原因**：自定义属性默认 `status=0` (隐藏)，而查询条件要求非预置属性必须 `status=1` 才显示。
*   **操作建议**：请在属性管理页面将需要的自定义属性设置为“显示”。
