package analysis

import (
	"strings"
	"testing"

	"github.com/1340691923/xwl_bi/platform-basic-libs/request"
)

func buildRetentionReqForTest(startDate string, endDate string, windowTime int) request.RetentionReqData {
	return request.RetentionReqData{
		ZhibiaoArr: []struct {
			EventName        string                 `json:"eventName"`
			EventNameDisplay string                 `json:"eventNameDisplay"`
			Relation         request.AnalysisFilter `json:"relation"`
		}{
			{
				EventName:        "用户注册",
				EventNameDisplay: "用户注册",
				Relation:         request.AnalysisFilter{FilterType: "COMPOUND"},
			},
			{
				EventName:        "应用启动",
				EventNameDisplay: "应用启动",
				Relation:         request.AnalysisFilter{FilterType: "COMPOUND"},
			},
		},
		WhereFilter:       request.AnalysisFilter{FilterType: "COMPOUND"},
		WhereFilterByUser: request.AnalysisFilter{FilterType: "COMPOUND"},
		WindowTime:        windowTime,
		WindowTimeFormat:  "天",
		Date:              []string{startDate, endDate},
		Appid:             41,
	}
}

func TestRetentionGetExecSqlUsesSingleGroupedQueryForLargeRange(t *testing.T) {
	retention := &Retention{
		sql: " and ( 1 = 1 ) ",
		req: buildRetentionReqForTest("2025-03-12", "2026-03-11", 7),
	}

	sql, args, err := retention.GetExecSql()
	if err != nil {
		t.Fatalf("GetExecSql returned error: %v", err)
	}

	if len(args) != 0 {
		t.Fatalf("expected no SQL args for empty filters, got %d", len(args))
	}

	lowerSQL := strings.ToLower(sql)
	if strings.Contains(lowerSQL, "union all") {
		t.Fatalf("expected grouped query without union all, got SQL: %s", sql)
	}

	if strings.Contains(lowerSQL, "retention(") {
		t.Fatalf("expected retention query without retention() expansion, got SQL: %s", sql)
	}

	if !strings.Contains(sql, "groupUniqArray(toDate(xwl_part_date)) AS return_dates") {
		t.Fatalf("expected SQL to pre-aggregate return dates, got SQL: %s", sql)
	}

	if !strings.Contains(sql, "GROUP BY T1.cohort_date") {
		t.Fatalf("expected SQL to group by cohort date, got SQL: %s", sql)
	}

	if len(sql) > 50000 {
		t.Fatalf("expected compact SQL for long range, got length %d", len(sql))
	}
}

func TestRetentionFillMissingRowsPreservesResponseShape(t *testing.T) {
	retention := &Retention{
		req: buildRetentionReqForTest("2025-03-12", "2025-03-14", 2),
	}

	rows := []retentionResultRow{
		{
			Dates: "2025-03-12",
			Value: []uint64{3, 2, 1, 0},
			UI: [][]string{
				{"u1", "u2", "u3"},
				{"u1", "u2"},
				{"u1"},
				{},
			},
		},
		{
			Dates: "2025-03-14",
			Value: []uint64{1, 1, 0, 0},
			UI: [][]string{
				{"u9"},
				{"u9"},
				{},
				{},
			},
		},
	}

	normalized := retention.fillMissingRows(rows)
	if len(normalized) != 3 {
		t.Fatalf("expected 3 rows after filling dates, got %d", len(normalized))
	}

	middle := normalized[1]
	if middle.Dates != "2025-03-13" {
		t.Fatalf("expected missing row for 2025-03-13, got %s", middle.Dates)
	}

	if len(middle.Value) != 4 {
		t.Fatalf("expected value length 4, got %d", len(middle.Value))
	}

	if len(middle.UI) != 4 {
		t.Fatalf("expected ui length 4, got %d", len(middle.UI))
	}

	for index, value := range middle.Value {
		if value != 0 {
			t.Fatalf("expected filled value[%d] to be 0, got %d", index, value)
		}
	}

	for index, ids := range middle.UI {
		if len(ids) != 0 {
			t.Fatalf("expected filled ui[%d] to be empty, got %v", index, ids)
		}
	}
}

func TestRetentionGetExecSqlRejectsReversedDateRange(t *testing.T) {
	retention := &Retention{
		sql: " and ( 1 = 1 ) ",
		req: buildRetentionReqForTest("2026-03-11", "2025-03-12", 3),
	}

	_, _, err := retention.GetExecSql()
	if err == nil {
		t.Fatal("expected invalid date range error, got nil")
	}
}
