package analysis

import (
	"strings"
	"testing"

	"github.com/1340691923/xwl_bi/platform-basic-libs/request"
)

func buildLTVReqForTest(startDate string, endDate string) request.LTVReqData {
	return request.LTVReqData{
		ZhibiaoArr: []struct {
			EventName        string                 `json:"eventName"`
			EventNameDisplay string                 `json:"eventNameDisplay"`
			Relation         request.AnalysisFilter `json:"relation"`
			ValueField       string                 `json:"valueField"`
		}{
			{
				EventName:        "用户注册",
				EventNameDisplay: "用户注册",
				Relation: request.AnalysisFilter{
					FilterType: "COMPOUND",
					Filts: []struct {
						FilterType string `json:"filterType"`
						Filts      []struct {
							ColumnName string      `json:"columnName"`
							Comparator string      `json:"comparator"`
							FilterType string      `json:"filterType"`
							Ftv        interface{} `json:"ftv"`
						} `json:"filts,omitempty"`
						Relation   string      `json:"relation,omitempty"`
						ColumnName string      `json:"columnName,omitempty"`
						Comparator string      `json:"comparator,omitempty"`
						Ftv        interface{} `json:"ftv,omitempty"`
					}{},
					Relation: "且",
				},
			},
			{
				EventName:        "支付订单",
				EventNameDisplay: "支付订单",
				ValueField:       "total_fee",
				Relation: request.AnalysisFilter{
					FilterType: "COMPOUND",
					Filts: []struct {
						FilterType string `json:"filterType"`
						Filts      []struct {
							ColumnName string      `json:"columnName"`
							Comparator string      `json:"comparator"`
							FilterType string      `json:"filterType"`
							Ftv        interface{} `json:"ftv"`
						} `json:"filts,omitempty"`
						Relation   string      `json:"relation,omitempty"`
						ColumnName string      `json:"columnName,omitempty"`
						Comparator string      `json:"comparator,omitempty"`
						Ftv        interface{} `json:"ftv,omitempty"`
					}{},
					Relation: "且",
				},
			},
		},
		WhereFilter: request.AnalysisFilter{
			FilterType: "COMPOUND",
			Filts: []struct {
				FilterType string `json:"filterType"`
				Filts      []struct {
					ColumnName string      `json:"columnName"`
					Comparator string      `json:"comparator"`
					FilterType string      `json:"filterType"`
					Ftv        interface{} `json:"ftv"`
				} `json:"filts,omitempty"`
				Relation   string      `json:"relation,omitempty"`
				ColumnName string      `json:"columnName,omitempty"`
				Comparator string      `json:"comparator,omitempty"`
				Ftv        interface{} `json:"ftv,omitempty"`
			}{},
			Relation: "且",
		},
		WhereFilterByUser: request.AnalysisFilter{
			FilterType: "COMPOUND",
			Filts: []struct {
				FilterType string `json:"filterType"`
				Filts      []struct {
					ColumnName string      `json:"columnName"`
					Comparator string      `json:"comparator"`
					FilterType string      `json:"filterType"`
					Ftv        interface{} `json:"ftv"`
				} `json:"filts,omitempty"`
				Relation   string      `json:"relation,omitempty"`
				ColumnName string      `json:"columnName,omitempty"`
				Comparator string      `json:"comparator,omitempty"`
				Ftv        interface{} `json:"ftv,omitempty"`
			}{},
			Relation: "且",
		},
		WindowTime:       1,
		WindowTimeFormat: "天",
		Date:             []string{startDate, endDate},
		Appid:            41,
	}
}

func TestLTVGetExecSqlUsesSingleGroupedQueryForLargeRange(t *testing.T) {
	ltv := &LTV{
		sql:              " and ( 1 = 1 ) ",
		revenueValueExpr: "toFloat64(ifNull(total_fee, 0))",
		req:              buildLTVReqForTest("2025-03-12", "2026-03-11"),
	}

	sql, args, err := ltv.GetExecSql()
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

	if !strings.Contains(sql, "GROUP BY T1.cohort_date") {
		t.Fatalf("expected SQL to group by cohort date, got SQL: %s", sql)
	}

	if !strings.Contains(sql, "ALL LEFT JOIN") {
		t.Fatalf("expected SQL to keep all revenue rows, got SQL: %s", sql)
	}

	if len(sql) > 50000 {
		t.Fatalf("expected compact SQL for long range, got length %d", len(sql))
	}
}

func TestLTVGetExecSqlRejectsReversedDateRange(t *testing.T) {
	ltv := &LTV{
		sql:              " and ( 1 = 1 ) ",
		revenueValueExpr: "toFloat64(ifNull(total_fee, 0))",
		req:              buildLTVReqForTest("2026-03-11", "2025-03-12"),
	}

	_, _, err := ltv.GetExecSql()
	if err == nil {
		t.Fatal("expected invalid date range error, got nil")
	}

	if !strings.Contains(err.Error(), "invalid date range") {
		t.Fatalf("expected invalid date range error, got %v", err)
	}
}
