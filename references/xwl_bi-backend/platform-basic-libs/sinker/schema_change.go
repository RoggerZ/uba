package sinker

import (
	"fmt"
	"sync"

	"github.com/1340691923/xwl_bi/engine/db"
	"github.com/1340691923/xwl_bi/engine/logs"
	model2 "github.com/1340691923/xwl_bi/platform-basic-libs/sinker/model"
	parser "github.com/1340691923/xwl_bi/platform-basic-libs/sinker/parse"
	"github.com/1340691923/xwl_bi/platform-basic-libs/util"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

func ChangeSchema(newKeys *sync.Map, dbname, table string, dims []*model2.ColumnWithType) ([]*model2.ColumnWithType, error) {
	// ChangeSchema 负责把新字段同步到 ClickHouse 表结构。
	//
	// 输入的 newKeys 是“字段名 -> 解析出的字段类型”。
	// 这里会把业务类型映射成 ClickHouse 类型，再逐条执行 ALTER TABLE。
	//
	// 示例：
	// 1. newKeys["price"] = parser.Float
	// 2. 最终生成 `ADD COLUMN IF NOT EXISTS price Float64`
	var queries []string
	var err error

	newKeys.Range(func(key, value interface{}) bool {
		strKey, _ := key.(string)
		intVal := value.(int)
		var strVal string
		switch intVal {
		case parser.Int, parser.Float:
			strVal = "Float64"
		case parser.String:
			strVal = "String"
		case parser.DateTime:
			strVal = "Nullable(DateTime)"
		case parser.IntArray:
			strVal = "Array(Int64)"
		case parser.FloatArray:
			strVal = "Array(Float64)"
		case parser.StringArray:
			strVal = "Array(String)"
		case parser.DateTimeArray:
			strVal = "Array(DateTime)"
		default:
			err = errors.Errorf("BUG: unsupported column type %s", strVal)
			return false
		}

		query := fmt.Sprintf("ALTER TABLE %s.%s %s ADD COLUMN IF NOT EXISTS `%s` %s", dbname, table, GetClusterSql(), strKey, strVal)
		queries = append(queries, query)
		tp, nullable := parser.WhichType(strVal)
		dims = append(dims, &model2.ColumnWithType{
			Name:       strKey,
			Type:       tp,
			Nullable:   nullable,
			SourceName: GetSourceName(strKey),
		})
		return true
	})

	for _, query := range queries {
		logs.Logger.Info(fmt.Sprintf("executing sql=> %s", query), zap.String("table", table))
		if _, err = db.ClickHouseSqlx.Exec(query); err != nil {
			util.RecordPersistenceError("clickhouse_schema_change_failed", err)
			return dims, errors.Wrapf(err, "%s", query)
		}
	}

	return dims, err
}
