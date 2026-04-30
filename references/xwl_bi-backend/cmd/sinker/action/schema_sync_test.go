package action

import (
	"errors"
	"sync"
	"testing"

	"github.com/1340691923/xwl_bi/engine/logs"
	"github.com/1340691923/xwl_bi/model"
	"github.com/1340691923/xwl_bi/platform-basic-libs/service/consumer_data"
	model2 "github.com/1340691923/xwl_bi/platform-basic-libs/sinker/model"
	parser "github.com/1340691923/xwl_bi/platform-basic-libs/sinker/parse"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

func TestAddTableColumnUsesCachedDimsFastPath(t *testing.T) {
	oldGetDims := getDimsFn
	oldChangeSchema := changeSchemaFn
	oldClearDimsCache := clearDimsCacheFn
	defer func() {
		getDimsFn = oldGetDims
		changeSchemaFn = oldChangeSchema
		clearDimsCacheFn = oldClearDimsCache
	}()

	getDimsCalled := false
	changeSchemaCalled := false
	getDimsFn = func(database, table string, excludedColumns []string, conn *sqlx.DB, onlyRedis bool) ([]*model2.ColumnWithType, error) {
		getDimsCalled = true
		return nil, nil
	}
	changeSchemaFn = func(newKeys *sync.Map, database, table string, dims []*model2.ColumnWithType) ([]*model2.ColumnWithType, error) {
		changeSchemaCalled = true
		return dims, nil
	}

	consumer_data.ResetTableColumnsForTest()
	consumer_data.StoreTableColumns("xwl_event52", []*model2.ColumnWithType{
		{Name: "xwl_part_date", Type: parser.DateTime},
		{Name: "xwl_distinct_id", Type: parser.String},
	})

	var p parser.FastjsonParser
	metric, err := p.Parse([]byte(`{"xwl_part_date":"2026-04-16 15:45:53","xwl_distinct_id":"abc"}`))
	if err != nil {
		t.Fatalf("Parse err: %v", err)
	}
	err = AddTableColumn(model.KafkaData{
		TableId:    "52",
		EventName:  "pay",
		ReportType: model.EventReportType,
		ReportTime: "2026-04-16 15:45:53",
		ReqData:    []byte(`{"xwl_part_date":"2026-04-16 15:45:53","xwl_distinct_id":"abc"}`),
	}, func(data consumer_data.ReportAcceptStatusData) {
		t.Fatalf("failFunc should not be called, got %+v", data)
	}, "xwl_event52", metric)
	if err != nil {
		t.Fatalf("AddTableColumn err: %v", err)
	}
	if getDimsCalled {
		t.Fatal("getDimsFn should not be called on cached fast path")
	}
	if changeSchemaCalled {
		t.Fatal("changeSchemaFn should not be called when there is no new key")
	}
}

func TestAddTableColumnRefreshesCachedDimsAfterSchemaChangeFailure(t *testing.T) {
	oldLogger := logs.Logger
	logs.Logger = zap.NewNop()
	defer func() {
		logs.Logger = oldLogger
	}()

	oldGetDims := getDimsFn
	oldChangeSchema := changeSchemaFn
	oldClearDimsCache := clearDimsCacheFn
	defer func() {
		getDimsFn = oldGetDims
		changeSchemaFn = oldChangeSchema
		clearDimsCacheFn = oldClearDimsCache
	}()
	clearDimsCacheFn = func(tableName string) {}

	getDimsCalls := 0
	getDimsFn = func(database, table string, excludedColumns []string, conn *sqlx.DB, onlyRedis bool) ([]*model2.ColumnWithType, error) {
		getDimsCalls++
		switch getDimsCalls {
		case 1:
			return []*model2.ColumnWithType{
				{Name: "xwl_part_date", Type: parser.DateTime},
				{Name: "xwl_distinct_id", Type: parser.String},
			}, nil
		case 2:
			return []*model2.ColumnWithType{
				{Name: "xwl_part_date", Type: parser.DateTime},
				{Name: "xwl_distinct_id", Type: parser.String},
			}, nil
		default:
			t.Fatalf("unexpected getDimsFn call count: %d", getDimsCalls)
			return nil, nil
		}
	}
	changeSchemaFn = func(newKeys *sync.Map, database, table string, dims []*model2.ColumnWithType) ([]*model2.ColumnWithType, error) {
		return append(dims, &model2.ColumnWithType{Name: "new_attr", Type: parser.String}), errors.New("schema change failed")
	}

	consumer_data.ResetTableColumnsForTest()

	var p parser.FastjsonParser
	metric, err := p.Parse([]byte(`{"xwl_part_date":"2026-04-16 15:45:53","xwl_distinct_id":"abc","new_attr":"v"}`))
	if err != nil {
		t.Fatalf("Parse err: %v", err)
	}

	err = AddTableColumn(model.KafkaData{
		TableId:    "52",
		EventName:  "pay",
		ReportType: model.EventReportType,
		ReportTime: "2026-04-16 15:45:53",
		ReqData:    []byte(`{"xwl_part_date":"2026-04-16 15:45:53","xwl_distinct_id":"abc","new_attr":"v"}`),
	}, func(data consumer_data.ReportAcceptStatusData) {
		t.Fatalf("failFunc should not be called, got %+v", data)
	}, "xwl_event52", metric)
	if err == nil || err.Error() != "schema change failed" {
		t.Fatalf("AddTableColumn err = %v, want schema change failed", err)
	}

	dims, ok := consumer_data.LoadTableColumns("xwl_event52")
	if !ok {
		t.Fatal("table columns should be stored after refresh")
	}
	if len(dims) != 2 {
		t.Fatalf("cached dims len = %d, want 2", len(dims))
	}
	if dims[0].Name != "xwl_part_date" || dims[1].Name != "xwl_distinct_id" {
		t.Fatalf("unexpected cached dims after refresh: %+v", dims)
	}
}
