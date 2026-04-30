package report

import "strings"

// BuildDebugDeviceIDKey 统一构造 debug 设备集合键。
//
// 之所以把它放在 report 包内，而不是 util 或其他公共目录，是因为：
// 1. 这个键只服务 report/debug 领域。
// 2. 目前没有其他模块依赖它。
// 3. 放在领域包内，后续如果 debug 设备语义变化，修改范围更清晰。
//
// 示例：
// 1. tableID="51" -> "DebugDeviceID_51"
// 2. tableID="xwl_demo" -> "DebugDeviceID_xwl_demo"
func BuildDebugDeviceIDKey(tableID string) string {
	var builder strings.Builder
	builder.Grow(len(tableID) + len("DebugDeviceID_"))
	builder.WriteString("DebugDeviceID_")
	builder.WriteString(tableID)
	return builder.String()
}
