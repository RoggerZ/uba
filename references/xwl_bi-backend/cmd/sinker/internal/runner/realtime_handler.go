package runner

import (
	"strconv"

	"github.com/1340691923/xwl_bi/engine/logs"
	"github.com/1340691923/xwl_bi/platform-basic-libs/service/consumer_data"
	"go.uber.org/zap"
)

// realTimeMessageHandler 只负责“原始消息 -> 实时表批次”这条轻量链路。
// 这条链路不做动态补列、不做复杂校验，只保证快速入批并尽快 ack。
type realTimeMessageHandler struct {
	sink           realTimeSink
	historyBlocker *historyReplayBlocker
	notifyWrite    func(tableName string)
}

// newRealTimeMessageHandler 创建实时链路处理器。
//
// 这里的依赖刻意很少，只保留实时表批次写入端，
// 因为公共解码已经在上游统一完成了。
func newRealTimeMessageHandler(sink realTimeSink, historyBlocker *historyReplayBlocker, notifyWrite func(tableName string)) *realTimeMessageHandler {
	return &realTimeMessageHandler{
		sink:           sink,
		historyBlocker: historyBlocker,
		notifyWrite:    notifyWrite,
	}
}

// HandleDecoded 负责执行实时链路的一次消息处理。
//
// 处理顺序很短：
// 1. 读取已经公共解码好的 KafkaData
// 2. 转成实时表结构
// 3. 放入批量器
// 4. 无论 sink 是否报错，都执行 markFn
//
// 这里选择“始终 mark”的原因是：
// 1. 实时链路定位为旁路快速落地，不承担严格重试职责。
// 2. 如果实时表临时失败，也不应拖死整个消费分区。
func (h *realTimeMessageHandler) HandleDecoded(decoded DecodedMessage) {
	defer decoded.MarkFn()

	appid, err := strconv.Atoi(decoded.KafkaData.TableId)
	if err != nil {
		logs.Logger.Error("strconv.Atoi(kafkaData.TableId) Err", zap.Error(err))
		return
	}

	if h.historyBlocker != nil && h.historyBlocker.ShouldSkip(consumer_data.TableNameRealTimeWarehousing, decoded.KafkaData.ReportTime) {
		return
	}
	if h.notifyWrite != nil {
		h.notifyWrite(consumer_data.TableNameRealTimeWarehousing)
	}

	if err := h.sink.Add(&consumer_data.RealTimeWarehousingData{
		Appid:      int64(appid),
		IngestTime: resolveIngestTime(decoded),
		EventName:  decoded.KafkaData.EventName,
		EventTime:  decoded.KafkaData.ReportTime,
		Data:       decoded.KafkaData.ReqData,
	}); err != nil {
		if consumer_data.IsDeferredFlushError(err) {
			return
		}
		logs.Logger.Error("AddRealTimeData err", zap.Error(err))
	}
}
