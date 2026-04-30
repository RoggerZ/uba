package action

import (
	"strconv"
	"strings"
)

// BuildMetaEventKey 生成“表 + 事件”的去重键。
//
// 示例：
// 1. tableID=51, eventName=AppLaunch
// 2. 输出 "51_AppLaunch"
func BuildMetaEventKey(tableID, eventName string) string {
	var builder strings.Builder
	builder.Grow(len(tableID) + 1 + len(eventName))
	builder.WriteString(tableID)
	builder.WriteString("_")
	builder.WriteString(eventName)
	return builder.String()
}

// BuildMetaAttrRelationKey 生成“表 + 事件 + 属性”的去重键。
//
// 示例：
// 1. tableID=51, eventName=AppLaunch, columnName=xwl_browser
// 2. 输出 "51_AppLaunch_xwl_browser"
func BuildMetaAttrRelationKey(tableID, eventName, columnName string) string {
	var builder strings.Builder
	builder.Grow(len(tableID) + len(eventName) + len(columnName) + 2)
	builder.WriteString(tableID)
	builder.WriteString("_")
	builder.WriteString(eventName)
	builder.WriteString("_")
	builder.WriteString(columnName)
	return builder.String()
}

// BuildAttributeKey 生成“表 + reportType + 属性”的去重键。
//
// 示例：
// 1. tableID=51, reportType=2, columnName=xwl_browser
// 2. 输出 "51_xwl_2_xwl_browser"
func BuildAttributeKey(tableID string, reportType int, columnName string) string {
	reportTypeText := strconv.Itoa(reportType)
	var builder strings.Builder
	builder.Grow(len(tableID) + len(reportTypeText) + len(columnName) + 6)
	builder.WriteString(tableID)
	builder.WriteString("_xwl_")
	builder.WriteString(reportTypeText)
	builder.WriteString("_")
	builder.WriteString(columnName)
	return builder.String()
}
