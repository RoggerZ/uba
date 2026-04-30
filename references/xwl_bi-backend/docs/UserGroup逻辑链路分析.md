# UserGroup (用户分群) 逻辑链路分析

本文档详细梳理了用户分群（User Group）功能的逻辑链路，包括分群的创建来源、数据存储方式以及在分析查询中的使用流程。

## 1. 核心概念

与事件数据（Event Data）不同，用户分群**不是**通过设备直接上报生成的，而是基于**分析结果的快照**。

*   **来源**: 用户在行为分析（如漏斗、留存）结果中，手动选中特定群体并保存。
*   **本质**: 一个静态的 User ID 列表（快照）。
*   **存储**: 使用 Gzip 压缩后的二进制数据存储在 MySQL 中。
*   **用途**: 在其他分析中作为筛选条件（SQL `IN` 子句）。

## 2. 数据流概览

### 2.1 创建分群 (Create)
**数据流向**: `Frontend (Vue)` -> `Backend API` -> `Service (Compression)` -> `MySQL`

1.  **触发**: 用户在 `FunnelResult.vue` (漏斗) 或 `RetentionResult.vue` (留存) 等组件中点击“创建分群”。
2.  **传输**: 前端将分析得出的 `uids` (用户ID数组) 和 `name` 发送给后端。
3.  **处理**: 后端接收 ID 列表，拼接为字符串并进行 **Gzip 压缩**。
4.  **落库**: 将压缩后的二进制数据 (`[]byte`) 存入 MySQL 的 `user_group` 表。

### 2.2 使用分群 (Read/Use)
**数据流向**: `Frontend (Filter)` -> `Backend Service` -> `MySQL (Fetch & Decompress)` -> `ClickHouse SQL`

1.  **筛选**: 用户在分析页面（如事件分析）的全局筛选器中选择某个“用户分群”。
2.  **获取**: 后端根据分群 ID 从 MySQL 读取压缩数据。
3.  **解压**: 将数据解压还原为 User ID 列表。
4.  **构建 SQL**: 将 ID 列表转换为 ClickHouse 的 `IN` 查询条件，如 `xwl_distinct_id IN ('id1', 'id2', ...)`。

---

## 3. 详细逻辑链路

### 3.1 创建分群 (Create)

*   **前端入口**: `vue/src/views/behavior-analysis/components/AddUserGroup.vue`
*   **API**: `/api/user_group/AddUserGroup`
*   **Controller**: `controller/user_group_controller.go` -> `AddUserGroup`
*   **Service**: `platform-basic-libs/service/user_group/user_group_service.go`

**关键代码逻辑**:

```go
// platform-basic-libs/service/user_group/user_group_service.go

func (this *UserGroupService) AddUserGroup(userCount int, uids []string, groupRemark, groupName string) (err error) {
    // 1. 拼接 ID 并进行 Gzip 压缩，节省 MySQL 存储空间
    b, err := util.GzipCompress(strings.Join(uids, ","))
    if err != nil {
        return
    }
    
    // 2. 存入 MySQL
    userGroup := model.UserGroup{}
    // ... 设置属性
    return userGroup.Insert(this.ManagerID, this.Appid, userCount, b)
}
```

### 3.2 使用分群 (Query Injection)

当在 `EventList` 等分析接口中使用分群时：

*   **逻辑入口**: `platform-basic-libs/service/analysis/event.go` -> `NewEvent`
*   **SQL 生成**: `platform-basic-libs/service/analysis/utils/user_group.go` -> `GetUserGroupSqlAndArgs`

**关键代码逻辑**:

```go
// platform-basic-libs/service/analysis/utils/user_group.go

func GetUserGroupSqlAndArgs(ids []int, appid int) (SQL string, Args []interface{}, err error) {
    // 1. 从 MySQL 查询分群记录
    // SELECT user_list FROM user_group WHERE id IN (...)
    
    // 2. 遍历结果并解压
    for index := range userGroupList {
        idStr, err := util.GzipUnCompress(userGroupList[index].UserList)
        // ...
        id := strings.Split(idStr, ",")
        // 3. 构建 OR 条件 (通常每个分群是一个 IN 条件，多个分群之间是 OR 关系)
        or = append(or, db.Eq{"xwl_distinct_id": [][]string{id}})
    }

    // 4. 返回 SQL 片段，例如: AND (xwl_distinct_id IN ('u1', 'u2') OR xwl_distinct_id IN ('u3'))
    SQL, Args, err = or.ToSql()
    SQL = " and " + SQL
    return SQL, Args, err
}
```

### 3.3 管理 (Update/Delete)

*   **前端列表**: `vue/src/views/user-analysis/group.vue`
*   **功能**:
    *   **修改**: 仅支持修改分群名称 (`group_name`) 和备注 (`group_remark`)。**不支持**修改分群内的用户 ID 列表（因为是快照）。
    *   **删除**: 物理删除 MySQL 中的记录。

## 4. 数据库设计

**Table**: `user_group` (MySQL)

| 字段 | 类型 | 说明 |
| :--- | :--- | :--- |
| `id` | int | 主键 |
| `group_name` | varchar | 分群名称 |
| `user_list` | blob/binary | **Gzip 压缩后的 User ID 列表** |
| `user_count` | int | 分群包含的用户数 |
| `create_by` | int | 创建人 ID |
| `appid` | int | 应用 ID |

## 5. 总结

用户分群的数据流是一个 **"ClickHouse -> Frontend -> MySQL -> ClickHouse"** 的闭环：

1.  **ClickHouse**: 通过 SQL 查询计算出满足特定行为的用户列表。
2.  **Frontend**: 用户在界面上确认并提交保存。
3.  **MySQL**: 后端压缩存储这份用户列表快照。
4.  **ClickHouse**: 后续分析时，从 MySQL 取出列表，再次注入回 ClickHouse SQL 中作为筛选条件。
