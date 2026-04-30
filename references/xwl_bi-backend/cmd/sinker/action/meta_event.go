package action

import "github.com/1340691923/xwl_bi/model"

// AddMetaEvent 负责维护事件名级别的元数据去重与异步投递。
func AddMetaEvent(kafkaData model.KafkaData) (err error) {
	if kafkaData.ReportType != model.EventReportType {
		return nil
	}

	key := BuildMetaEventKey(kafkaData.TableId, kafkaData.EventName)

	if _, found := MetaEventMap.Load(key); found {
		return nil
	}

	metaEventChan <- metaEventModel{
		EventName: kafkaData.EventName,
		AppId:     kafkaData.TableId,
	}
	MetaEventMap.Store(key, struct{}{})
	return nil
}
