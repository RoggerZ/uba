package controller

import (
	"errors"
	"fmt"

	"github.com/1340691923/xwl_bi/engine/logs"
	"github.com/1340691923/xwl_bi/model"
	"github.com/1340691923/xwl_bi/platform-basic-libs/my_error"
	"github.com/1340691923/xwl_bi/platform-basic-libs/service/report"
	sinkerModel "github.com/1340691923/xwl_bi/platform-basic-libs/sinker/model"
	parser "github.com/1340691923/xwl_bi/platform-basic-libs/sinker/parse"
	"github.com/1340691923/xwl_bi/platform-basic-libs/util"
	"go.uber.org/zap"
)

type parseMetricFunc func(body []byte) (*parser.FastjsonMetric, error)
type loadDimsFunc func(tableName string) ([]*sinkerModel.ColumnWithType, error)
type isDebugDeviceFunc func(debug, distinctID, tableID string) bool
type sendDebugDataFunc func(data map[string]interface{}) error

type debugInspectionHandler struct {
	parseMetric   parseMetricFunc
	loadDims      loadDimsFunc
	isDebugDevice isDebugDeviceFunc
	sendDebugData sendDebugDataFunc
}

// newDebugInspectionHandler 创建 debug 检查处理器。
//
// 这里注入的小函数分别代表不同的变化点：
// 1. parseMetric：如何把上报 body 解析成 FastjsonMetric。
// 2. loadDims：如何读取 CK 列定义。
// 3. isDebugDevice：如何判断当前请求是否命中 debug 设备。
// 4. sendDebugData：如何把 debug 校验账本写入 Kafka。
//
// 这样做的目标是让 debug 检查逻辑本身稳定，而把基础设施接法留给装配层决定。
func newDebugInspectionHandler(
	parseMetric parseMetricFunc,
	loadDims loadDimsFunc,
	isDebugDevice isDebugDeviceFunc,
	sendDebugData sendDebugDataFunc,
) *debugInspectionHandler {
	return &debugInspectionHandler{
		parseMetric:   parseMetric,
		loadDims:      loadDims,
		isDebugDevice: isDebugDevice,
		sendDebugData: sendDebugData,
	}
}

// Handle 负责处理命中 debug 设备后的校验、记账和“是否继续正式上报”的决策。
//
// 关键语义必须保持与旧实现一致：
// 1. 只有命中 debug 设备集合，才会进入这条链路。
// 2. 命中后一定会尝试把校验结果写入 debug topic。
// 3. 校验失败直接返回业务错误，不继续正式 Kafka。
// 4. debug=2 且校验通过时，返回“成功但不入库”，不继续正式 Kafka。
//
// 数据流示例：
// 1. 请求先被 ingress handler 构造成 KafkaData。
// 2. Handle 会读取 KafkaData.ReqData 并解析为 FastjsonMetric。
// 3. 根据 CK dims 做类型校验，并把校验结果整理成 debugData。
// 4. 无论校验成功还是失败，都先把 debugData 写进 debug topic。
// 5. 最后再通过 decision/error 告诉主流程是否继续正式 Kafka。
func (h *debugInspectionHandler) Handle(request DecodedReportRequest, kafkaData model.KafkaData) (debugInspectionDecision, error) {
	if !h.isDebugDevice(request.Debug, request.DistinctID, kafkaData.TableId) {
		return debugInspectionDecision{}, nil
	}

	metric, debugErr := h.parseMetric(kafkaData.ReqData)
	if debugErr != nil {
		logs.Logger.Error("parser.ParseKafkaData", append(reportLogFields(request, kafkaData), zap.Error(debugErr))...)
		return debugInspectionDecision{}, errors.New("服务异常")
	}

	dims, err := h.loadDims(kafkaData.GetTableName())
	if err != nil {
		logs.Logger.Error("load debug dims", append(reportLogFields(request, kafkaData), zap.Error(err))...)
		return debugInspectionDecision{}, errors.New("服务异常")
	}

	obj := metric.GetParseObject()
	debugData := map[string]interface{}{
		"data_name":   kafkaData.EventName,
		"report_data": util.Bytes2str(request.Body),
		"report_time": kafkaData.ReportTime,
		"appid":       kafkaData.TableId,
		"distinct_id": request.DistinctID,
	}
	haveFailAttr := false
	eventType := kafkaData.GetReportTypeErr()

	// 先做字段类型检查。
	//
	// 这里沿用旧逻辑的容错约定：
	// 1. Int <-> Float 之间允许互相兼容。
	// 2. 其他类型不兼容时，立即把错误原因写入 debugData。
	for _, column := range dims {
		if obj.Get(column.Name) == nil {
			continue
		}

		reportType := parser.FjDetectType(obj.Get(column.Name))
		if reportType == column.Type {
			continue
		}
		if (reportType == parser.Int && column.Type == parser.Float) || (reportType == parser.Float && column.Type == parser.Int) {
			continue
		}

		debugData["error_reason"] = fmt.Sprintf(
			"%s的类型错误，正确类型为%v，上报类型为%v(%v)",
			column.Name,
			parser.TypeRemarkMap[column.Type],
			parser.TypeRemarkMap[reportType],
			obj.Get(column.Name).String(),
		)
		debugData["data_judge"] = eventType
		haveFailAttr = true
	}

	applyDebugClockSkewCheck(request.Body, kafkaData.ReportTime, eventType, debugData, &haveFailAttr)
	if !haveFailAttr {
		debugData["data_judge"] = "数据检验通过"
	}

	// 即使校验失败，也必须先把 debugData 写入 debug topic。
	//
	// 这是当前链路很重要的旧语义：
	// 1. debug topic 承担“调试账本”角色。
	// 2. 用户看到接口失败之前，后台已经能查到这条失败记录。
	if err := h.sendDebugData(debugData); err != nil {
		logs.Logger.Error("send debug data failed", append(reportLogFields(request, kafkaData), zap.Error(err))...)
		return debugInspectionDecision{}, errors.New("服务异常")
	}

	if haveFailAttr {
		logs.Logger.Error(
			"debug inspection failed",
			append(reportLogFields(request, kafkaData), zap.String("error_reason", debugData["error_reason"].(string)))...,
		)
		// 校验失败时返回业务错误，主流程必须立即停止，不得继续正式 Kafka。
		return debugInspectionDecision{Stop: true}, my_error.NewError(debugData["error_reason"].(string), 10006)
	}

	if request.Debug == report.DebugNotToDB {
		// debug=2 的含义是“校验通过，但只做调试，不做正式入库”。
		return debugInspectionDecision{
			Stop:    true,
			Message: "上报成功（数据不入库）",
		}, nil
	}

	return debugInspectionDecision{}, nil
}

func reportLogFields(request DecodedReportRequest, kafkaData model.KafkaData) []zap.Field {
	// 这里统一收口日志字段，避免每个错误分支各自拼一遍上下文。
	//
	// 示例：
	// 1. debug 校验失败时，可以直接看到 appid / table_id / event_name / distinct_id / report_time
	// 2. 排障时不用再回头翻原始请求拼接上下文
	return []zap.Field{
		zap.String("appid", request.APPID),
		zap.String("table_id", kafkaData.TableId),
		zap.String("event_name", kafkaData.EventName),
		zap.String("debug", request.Debug),
		zap.String("typ", request.Typ),
		zap.String("distinct_id", request.DistinctID),
		zap.String("report_time", kafkaData.ReportTime),
	}
}
