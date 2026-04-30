package action

import (
	"strings"

	"github.com/1340691923/xwl_bi/engine/db"
	"github.com/1340691923/xwl_bi/engine/logs"
)

var (
	attributeChan        = make(chan attributeModel, 10000)
	metaEventChan        = make(chan metaEventModel, 10000)
	metaAttrRelationChan = make(chan metaAttrRelationModel, 10000)
)

type metaAttrRelationModel struct {
	EventName string
	EventAttr string
	AppId     string
}

type attributeModel struct {
	AttributeName   string
	DataType        int
	AttributeType   int
	attributeSource int
	AppId           string
}

type metaEventModel struct {
	EventName string
	AppId     string
}

// MysqlConsumer 专门消费元数据异步写库请求。
//
// Kafka 主链路只负责把请求投递进 channel，
// 真正的 MySQL INSERT 在这里串行完成。
func MysqlConsumer() {
	for {
		select {
		case m := <-metaAttrRelationChan:
			if _, err := db.Sqlx.Exec(`insert into  meta_attr_relation(app_id,event_name,event_attr) values (?,?,?);`,
				m.AppId, m.EventName, m.EventAttr); err != nil && !strings.Contains(err.Error(), "1062") {
				logs.Logger.Sugar().Errorf("meta_attr_relation insert %v %v", m, err)
			}
		case m := <-attributeChan:
			if _, err := db.Sqlx.Exec(`insert into  attribute(app_id,attribute_source,attribute_type,data_type,attribute_name) values (?,?,?,?,?);`,
				m.AppId, m.attributeSource, m.AttributeType, m.DataType, m.AttributeName); err != nil && !strings.Contains(err.Error(), "1062") {
				logs.Logger.Sugar().Errorf("attribute insert %v %v", m, err)
			}
		case m := <-metaEventChan:
			if _, err := db.Sqlx.Exec(`insert into  meta_event(appid,event_name) values (?,?);`, m.AppId, m.EventName); err != nil && !strings.Contains(err.Error(), "1062") {
				logs.Logger.Sugar().Errorf("metaEvent insert %v %v", m, err)
			}
		}
	}
}
