package action

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"sync"

	"github.com/1340691923/xwl_bi/engine/db"
	"github.com/1340691923/xwl_bi/engine/logs"
	"github.com/1340691923/xwl_bi/model"
	"github.com/1340691923/xwl_bi/platform-basic-libs/service/consumer_data"
	"github.com/1340691923/xwl_bi/platform-basic-libs/sinker"
	model2 "github.com/1340691923/xwl_bi/platform-basic-libs/sinker/model"
	parser "github.com/1340691923/xwl_bi/platform-basic-libs/sinker/parse"
	"github.com/1340691923/xwl_bi/platform-basic-libs/util"
	"github.com/valyala/fastjson"
	"go.uber.org/zap"
)

var (
	getDimsFn        = sinker.GetDims
	changeSchemaFn   = sinker.ChangeSchema
	clearDimsCacheFn = clearDimsCache
)

// AddTableColumn 负责校验现有列类型、发现新列并在必要时触发补列。
//
// 关键优化点：
//  1. 如果内存里的 tableColumnMap 已经有当前表的列定义，并且本条消息没有新列，
//     就直接走快路径返回，不再每条消息都同步 GetDims。
//  2. 只有缓存缺失或确实发现新列时，才回退到原来的慢路径读取 ClickHouse schema。
func AddTableColumn(kafkaData model.KafkaData, failFunc func(data consumer_data.ReportAcceptStatusData), tableName string, reqDataObject *parser.FastjsonMetric) (err error) {
	obj, err := reqDataObject.GetParseObject().Object()
	if err != nil {
		logs.Logger.Error("ReqDataObject.GetParseObject().Object()", zap.Error(err))
		return err
	}

	tableID, _ := strconv.Atoi(kafkaData.TableId)
	reportTypeErr := kafkaData.GetReportTypeErr()

	if dims, ok := loadCachedTableDims(tableName); ok {
		foundNewKey, newKeys, err := validateAndTrackColumns(kafkaData, obj, dims, tableID, reportTypeErr, failFunc)
		if err != nil {
			return err
		}
		if !foundNewKey {
			return nil
		}

		dims, err = getDimsFn(model.GlobConfig.Comm.ClickHouse.DbName, tableName, nil, db.ClickHouseSqlx, false)
		if err != nil {
			logs.Logger.Error("sinker.GetDims", zap.Error(err))
			return err
		}

		dims, err = changeSchemaFn(newKeys, model.GlobConfig.Comm.ClickHouse.DbName, tableName, dims)
		if err != nil {
			logs.Logger.Error("err", zap.Error(err))
		}
		clearDimsCacheFn(tableName)
		if refreshErr := refreshTableColumnsFromSource(tableName); refreshErr != nil {
			if err == nil {
				return refreshErr
			}
		}
		return err
	}

	dims, err := getDimsFn(model.GlobConfig.Comm.ClickHouse.DbName, tableName, nil, db.ClickHouseSqlx, false)
	if err != nil {
		logs.Logger.Error("sinker.GetDims", zap.Error(err))
		return err
	}

	foundNewKey, newKeys, err := validateAndTrackColumns(kafkaData, obj, dims, tableID, reportTypeErr, failFunc)
	if err != nil {
		return err
	}
	if foundNewKey {
		dims, err = changeSchemaFn(newKeys, model.GlobConfig.Comm.ClickHouse.DbName, tableName, dims)
		if err != nil {
			logs.Logger.Error("err", zap.Error(err))
		}
		clearDimsCacheFn(tableName)
		if refreshErr := refreshTableColumnsFromSource(tableName); refreshErr != nil {
			if err == nil {
				return refreshErr
			}
		}
		return err
	}

	consumer_data.StoreTableColumns(tableName, dims)
	return err
}

// refreshTableColumnsFromSource 在补列后强制回源 ClickHouse，避免把“推测中的列集合”留在内存缓存里。
//
// 典型场景：
// 1. AddTableColumn 发现 5 个新字段，先调用 ChangeSchema 尝试补列。
// 2. 如果补列只成功了一部分，或者线上表结构与本地推断不一致，旧代码会把推测出的 dims 直接写回 tableColumnMap。
// 3. 后续 reportdata2ck 再按这份缓存组装 row，就会出现“表里 30 列，但内存里认为有 35 列”的参数个数错位。
func refreshTableColumnsFromSource(tableName string) error {
	dims, err := getDimsFn(model.GlobConfig.Comm.ClickHouse.DbName, tableName, nil, db.ClickHouseSqlx, false)
	if err != nil {
		logs.Logger.Error("refresh table columns from source failed",
			zap.String("table", tableName),
			zap.Error(err),
		)
		return err
	}

	consumer_data.StoreTableColumns(tableName, dims)
	return nil
}

func loadCachedTableDims(tableName string) ([]*model2.ColumnWithType, bool) {
	dims, ok := consumer_data.LoadTableColumns(tableName)
	if !ok || len(dims) == 0 {
		return nil, false
	}
	return dims, true
}

func validateAndTrackColumns(
	kafkaData model.KafkaData,
	obj *fastjson.Object,
	dims []*model2.ColumnWithType,
	tableID int,
	reportTypeErr string,
	failFunc func(data consumer_data.ReportAcceptStatusData),
) (foundNewKey bool, newKeys *sync.Map, err error) {
	knownKeys := make([]string, 0, len(dims))
	newKeys = new(sync.Map)

	for _, column := range dims {
		knownKeys = append(knownKeys, column.Name)
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

		errorReason := fmt.Sprintf("%s的类型错误，正确类型为%v，上报类型为%v(%v)", column.Name, parser.TypeRemarkMap[column.Type], parser.TypeRemarkMap[reportType], obj.Get(column.Name).String())
		failFunc(consumer_data.ReportAcceptStatusData{
			PartDate:       kafkaData.ReportTime,
			TableId:        tableID,
			ReportType:     reportTypeErr,
			DataName:       kafkaData.EventName,
			ErrorReason:    errorReason,
			ErrorHandling:  "丢弃数据",
			ReportData:     util.Bytes2str(kafkaData.ReqData),
			XwlKafkaOffset: kafkaData.Offset,
			Status:         consumer_data.FailStatus,
		})
		return false, nil, errors.New(errorReason)
	}

	builder := bytes.Buffer{}
	obj.Visit(func(key []byte, v *fastjson.Value) {
		columnName := string(key)
		enqueueMetaAttrRelationIfNeeded(kafkaData, columnName, builder)
		registerAttributeAndTrackNewKey(kafkaData, columnName, obj, knownKeys, newKeys, &foundNewKey, builder)
	})

	return foundNewKey, newKeys, nil
}

func enqueueMetaAttrRelationIfNeeded(kafkaData model.KafkaData, columnName string, builder bytes.Buffer) {
	builder.Reset()
	builder.WriteString(BuildMetaAttrRelationKey(kafkaData.TableId, kafkaData.EventName, columnName))
	key := builder.String()

	if _, found := MetaAttrRelationSet.Load(key); found {
		return
	}

	metaAttrRelationChan <- metaAttrRelationModel{
		EventName: kafkaData.EventName,
		EventAttr: columnName,
		AppId:     kafkaData.TableId,
	}
	MetaAttrRelationSet.Store(key, struct{}{})
}

func registerAttributeAndTrackNewKey(kafkaData model.KafkaData, columnName string, obj *fastjson.Object, knownKeys []string, newKeys *sync.Map, foundNewKey *bool, builder bytes.Buffer) {
	if !util.InstrArr(knownKeys, columnName) {
		*foundNewKey = true
		newKeys.Store(columnName, parser.FjDetectType(obj.Get(columnName)))
	}

	builder.Reset()
	builder.WriteString(BuildAttributeKey(kafkaData.TableId, kafkaData.ReportType, columnName))
	attributeMapKey := builder.String()

	if _, found := AttributeMap.Load(attributeMapKey); found {
		return
	}

	var attributeType, attributeSource int
	if _, ok := parser.SysColumn[columnName]; ok {
		attributeType = PresetAttribute
	} else {
		attributeType = CustomAttribute
	}

	switch kafkaData.ReportType {
	case model.UserReportType:
		attributeSource = IsUserAttribute
	case model.EventReportType:
		attributeSource = IsEventAttribute
	}

	attributeChan <- attributeModel{
		AttributeName:   columnName,
		DataType:        parser.FjDetectType(obj.Get(columnName)),
		AttributeType:   attributeType,
		attributeSource: attributeSource,
		AppId:           kafkaData.TableId,
	}
	AttributeMap.Store(attributeMapKey, struct{}{})
}

func clearDimsCache(tableName string) {
	redisConn := db.RedisPool.Get()
	defer redisConn.Close()

	dimsCacheKey := sinker.GetDimsCachekey(model.GlobConfig.Comm.ClickHouse.DbName, tableName)
	_, err := redisConn.Do("unlink", dimsCacheKey)
	if err != nil {
		if _, err = redisConn.Do("del", dimsCacheKey); err != nil {
			logs.Logger.Error("err", zap.Error(err))
		}
	}
	sinker.ClearDimsCacheByKey(dimsCacheKey)
}
