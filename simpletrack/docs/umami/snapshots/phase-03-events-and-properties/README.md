# Phase 03 README

## Goal

记录 Events 页的完整布局、Chart / Activity / Properties、Filter 和日期范围，并验证 demo 上报后的页面反馈。

## Preconditions

- demo 页面已可触发 DOM event、track() 和 identify()
- 已完成至少一轮上报请求

## Snapshot Order

1. 触发 DOM event
2. 触发 `track()`
3. 触发 `identify()`
4. 打开 Events 页面完整布局
5. 查看 Activity
6. 查看 Properties
7. 打开 Filter
8. 展开日期范围

## Observation Focus

- 事件列表可读性
- 属性入口显隐方式
- 筛选和明细的操作成本
- Activity log 的使用路径
- 日期范围和 Filter 是否容易联动
