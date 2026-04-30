package controller

import (
	"math"

	"github.com/1340691923/xwl_bi/platform-basic-libs/util"
	"github.com/tidwall/gjson"
)

// applyDebugClockSkewCheck 负责执行 debug 校验里的“客户端时间偏差是否超过十分钟”检查。
//
// 这里保留旧语义：
// 1. 客户端时间来自 xwl_update_time。
// 2. 一旦偏差超过十分钟，就把错误原因写入 debug 账本。
// 3. 校验结果通过输出 map 回传给上层，而不是在函数内部直接返回 HTTP 响应。
func applyDebugClockSkewCheck(body []byte, reportTime, eventType string, data map[string]interface{}, haveFailAttr *bool) {
	xwlUpdateTime := gjson.GetBytes(body, "xwl_update_time").String()
	clientTime := util.Str2Time(xwlUpdateTime, util.TimeFormat)
	serverTime := util.Str2Time(reportTime, util.TimeFormat)
	if math.Abs(serverTime.Sub(clientTime).Minutes()) > 10 {
		*haveFailAttr = true
		data["error_reason"] = "客户端上报时间误差大于十分钟"
		data["data_judge"] = eventType
	}
}
