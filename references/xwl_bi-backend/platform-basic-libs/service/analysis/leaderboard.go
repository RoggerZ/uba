package analysis

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/1340691923/xwl_bi/engine/db"
	"github.com/1340691923/xwl_bi/engine/logs"
	"github.com/1340691923/xwl_bi/platform-basic-libs/my_error"
	"github.com/1340691923/xwl_bi/platform-basic-libs/request"
	"github.com/1340691923/xwl_bi/platform-basic-libs/service/analysis/utils"
	jsoniter "github.com/json-iterator/go"
)

const (
	leaderboardMetricEventCount  = "event_count"
	leaderboardMetricUserCount   = "user_count"
	leaderboardMetricSum         = "sum"
	leaderboardMetricAvg         = "avg"
	leaderboardMetricDistinct    = "distinct_count"
	leaderboardMetricAvgCount    = "avg_count_by_user"
	leaderboardMetricAvgSum      = "avg_sum_by_user"
	leaderboardMetricSuccessRate = "success_rate"
)

const (
	leaderboardSortByCurrent = "current_value"
	leaderboardSortByDelta   = "delta_value"
	leaderboardSortByRate    = "delta_rate"
)

var safeIdentifierRegexp = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

type Leaderboard struct {
	req          request.LeaderboardReqData
	userGroupSql string
	userGroupArg []interface{}
}

type leaderboardAggRow struct {
	GroupValue  string  `db:"group_value"`
	MetricValue float64 `db:"metric_value"`
}

type LeaderboardRowRes struct {
	Rank              int     `json:"rank"`
	GroupKey          string  `json:"group_key"`
	GroupValue        string  `json:"group_value"`
	GroupDisplayValue string  `json:"group_display_value"`
	CurrentValue      float64 `json:"current_value"`
	ShareRate         float64 `json:"share_rate"`
	PrevValue         float64 `json:"prev_value"`
	DeltaValue        float64 `json:"delta_value"`
	DeltaRate         float64 `json:"delta_rate"`
}

type LeaderboardSummaryRes struct {
	TotalValue  float64 `json:"total_value"`
	RowCount    int     `json:"row_count"`
	TopN        int     `json:"top_n"`
	OthersValue float64 `json:"others_value"`
}

func NewLeaderboard(reqData []byte) (Ianalysis, error) {
	obj := &Leaderboard{}
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	if err := json.Unmarshal(reqData, &obj.req); err != nil {
		return nil, err
	}

	if len(obj.req.Date) < 2 {
		return nil, my_error.NewBusiness(ERROR_TABLE, TimeError)
	}
	if strings.TrimSpace(obj.req.EventName) == "" {
		return nil, my_error.NewBusiness(ERROR_TABLE, EventNameEmptyError)
	}
	if len(obj.req.GroupBy) != 1 {
		return nil, my_error.NewBusiness(ERROR_TABLE, GroupNumError)
	}
	if strings.TrimSpace(obj.req.GroupBy[0]) == "" {
		return nil, my_error.NewBusiness(ERROR_TABLE, GroupEmptyError)
	}
	if _, err := sanitizeIdentifier(obj.req.GroupBy[0]); err != nil {
		return nil, err
	}

	obj.req.Metric.MetricType = strings.TrimSpace(strings.ToLower(obj.req.Metric.MetricType))
	if obj.req.Metric.MetricType == "" {
		obj.req.Metric.MetricType = leaderboardMetricEventCount
	}

	if obj.req.Metric.MetricType == leaderboardMetricSum ||
		obj.req.Metric.MetricType == leaderboardMetricAvg ||
		obj.req.Metric.MetricType == leaderboardMetricDistinct ||
		obj.req.Metric.MetricType == leaderboardMetricAvgSum ||
		obj.req.Metric.MetricType == leaderboardMetricSuccessRate {
		if strings.TrimSpace(obj.req.Metric.ValueField) == "" {
			return nil, errors.New("metric.valueField 不能为空")
		}
		if _, err := sanitizeIdentifier(obj.req.Metric.ValueField); err != nil {
			return nil, err
		}
	}

	obj.req.SortBy = strings.TrimSpace(strings.ToLower(obj.req.SortBy))
	if obj.req.SortBy == "" {
		obj.req.SortBy = leaderboardSortByCurrent
	}
	if obj.req.SortBy != leaderboardSortByCurrent && obj.req.SortBy != leaderboardSortByDelta && obj.req.SortBy != leaderboardSortByRate {
		return nil, errors.New("sortBy 不合法")
	}

	obj.req.SortOrder = strings.TrimSpace(strings.ToLower(obj.req.SortOrder))
	if obj.req.SortOrder == "" {
		obj.req.SortOrder = "desc"
	}
	if obj.req.SortOrder != "desc" && obj.req.SortOrder != "asc" {
		return nil, errors.New("sortOrder 不合法")
	}

	if obj.req.TopN <= 0 {
		obj.req.TopN = 20
	}
	if obj.req.TopN > 100 {
		return nil, errors.New("topN 不能超过 100")
	}

	obj.req.RankingMode = strings.TrimSpace(strings.ToLower(obj.req.RankingMode))
	if obj.req.RankingMode == "" {
		obj.req.RankingMode = "hot"
	}

	userGroupSql, userGroupArg, err := utils.GetUserGroupSqlAndArgs(obj.req.UserGroup, obj.req.Appid)
	if err != nil {
		return nil, err
	}
	obj.userGroupSql = userGroupSql
	obj.userGroupArg = userGroupArg

	return obj, nil
}

func (this *Leaderboard) GetExecSql() (string, []interface{}, error) {
	return this.buildAggSQL(this.req.Date)
}

func (this *Leaderboard) GetList() (interface{}, error) {
	currentMap, err := this.queryAgg(this.req.Date)
	if err != nil {
		return nil, err
	}

	compareMap := map[string]float64{}
	hasCompare := len(this.req.CompareDate) >= 2 &&
		strings.TrimSpace(this.req.CompareDate[0]) != "" &&
		strings.TrimSpace(this.req.CompareDate[1]) != ""

	if hasCompare {
		compareMap, err = this.queryAgg(this.req.CompareDate)
		if err != nil {
			return nil, err
		}
	}

	rows, currentTotal, compareTotal := this.mergeRows(currentMap, compareMap, hasCompare)
	this.sortRows(rows)

	topRows, othersValue := this.applyTopN(rows)
	for idx := range topRows {
		topRows[idx].Rank = idx + 1
	}

	summary := LeaderboardSummaryRes{
		TotalValue:  currentTotal,
		RowCount:    len(topRows),
		TopN:        this.req.TopN,
		OthersValue: othersValue,
	}

	compareSummary := map[string]interface{}{"total_value": compareTotal}
	meta := map[string]interface{}{
		"eventName":    this.req.EventName,
		"metricType":   this.req.Metric.MetricType,
		"groupBy":      this.req.GroupBy[0],
		"rankingMode":  this.req.RankingMode,
		"sortBy":       this.req.SortBy,
		"sortOrder":    this.req.SortOrder,
		"hasCompare":   hasCompare,
		"includeOther": this.req.IncludeOthers,
	}

	return map[string]interface{}{
		"rows":           topRows,
		"summary":        summary,
		"compareSummary": compareSummary,
		"meta":           meta,
	}, nil
}

func (this *Leaderboard) buildAggSQL(date []string) (string, []interface{}, error) {
	if len(date) < 2 {
		return "", nil, my_error.NewBusiness(ERROR_TABLE, TimeError)
	}

	groupByField, err := sanitizeIdentifier(this.req.GroupBy[0])
	if err != nil {
		return "", nil, err
	}
	metricExpr, err := this.getMetricExpr()
	if err != nil {
		return "", nil, err
	}

	globalFilterSql, globalArgs, err := buildAnalysisFilterSql(this.req.WhereFilter)
	if err != nil {
		return "", nil, err
	}

	userFilterSql, userFilterArg, err := getUserfilterSqlArgs(this.req.WhereFilterByUser, this.req.Appid)
	if err != nil {
		return "", nil, err
	}

	startTime := date[0] + " 00:00:00"
	endTime := date[1] + " 23:59:59"

	sqlText := fmt.Sprintf(`
		SELECT
			if(isNull(%s), '', toString(%s)) AS group_value,
			toFloat64(%s) AS metric_value
		FROM xwl_event%d
		PREWHERE xwl_part_event = ?
			AND xwl_part_date >= toDateTime(?)
			AND xwl_part_date <= toDateTime(?)
			AND (%s)%s%s
		GROUP BY group_value
		LIMIT 1000
	`, groupByField, groupByField, metricExpr, this.req.Appid, globalFilterSql, this.userGroupSql, userFilterSql)

	args := []interface{}{this.req.EventName, startTime, endTime}
	args = append(args, globalArgs...)
	args = append(args, this.userGroupArg...)
	args = append(args, userFilterArg...)
	return sqlText, args, nil
}

func (this *Leaderboard) getMetricExpr() (string, error) {
	switch this.req.Metric.MetricType {
	case leaderboardMetricEventCount:
		return "count()", nil
	case leaderboardMetricUserCount:
		return "count(DISTINCT xwl_distinct_id)", nil
	case leaderboardMetricSum:
		field, err := sanitizeIdentifier(this.req.Metric.ValueField)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("sum(toFloat64(ifNull(%s, 0)))", field), nil
	case leaderboardMetricAvg:
		field, err := sanitizeIdentifier(this.req.Metric.ValueField)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("avg(toFloat64(ifNull(%s, 0)))", field), nil
	case leaderboardMetricDistinct:
		field, err := sanitizeIdentifier(this.req.Metric.ValueField)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("count(DISTINCT %s)", field), nil
	case leaderboardMetricAvgCount:
		return "if(count(DISTINCT xwl_distinct_id) = 0, 0, count() / count(DISTINCT xwl_distinct_id))", nil
	case leaderboardMetricAvgSum:
		field, err := sanitizeIdentifier(this.req.Metric.ValueField)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("if(count(DISTINCT xwl_distinct_id) = 0, 0, sum(toFloat64(ifNull(%s, 0))) / count(DISTINCT xwl_distinct_id))", field), nil
	case leaderboardMetricSuccessRate:
		field, err := sanitizeIdentifier(this.req.Metric.ValueField)
		if err != nil {
			return "", err
		}
		successValue := this.req.Metric.SuccessValue
		if strings.TrimSpace(successValue) == "" {
			successValue = "success"
		}
		successValue = escapeSQLStringLiteral(successValue)
		return fmt.Sprintf("if(count() = 0, 0, sum(if(toString(%s) = '%s', 1, 0)) / count())", field, successValue), nil
	default:
		return "", errors.New("metric.metricType 不支持")
	}
}

func (this *Leaderboard) queryAgg(date []string) (map[string]float64, error) {
	sqlText, args, err := this.buildAggSQL(date)
	if err != nil {
		return nil, err
	}

	logs.Logger.Sugar().Infof("leaderboard sql", sqlText, args)

	var rows []leaderboardAggRow
	if err := db.ClickHouseSqlx.Select(&rows, sqlText, args...); err != nil {
		return nil, err
	}

	result := make(map[string]float64, len(rows))
	for _, row := range rows {
		result[row.GroupValue] += row.MetricValue
	}
	return result, nil
}

func (this *Leaderboard) mergeRows(currentMap map[string]float64, compareMap map[string]float64, hasCompare bool) ([]LeaderboardRowRes, float64, float64) {
	groupSet := map[string]struct{}{}
	for k := range currentMap {
		groupSet[k] = struct{}{}
	}
	if hasCompare {
		for k := range compareMap {
			groupSet[k] = struct{}{}
		}
	}

	rows := make([]LeaderboardRowRes, 0, len(groupSet))
	currentTotal := 0.0
	compareTotal := 0.0

	for groupValue := range groupSet {
		if this.req.ExcludeEmpty && strings.TrimSpace(groupValue) == "" {
			continue
		}

		currentValue := currentMap[groupValue]
		prevValue := 0.0
		if hasCompare {
			prevValue = compareMap[groupValue]
		}

		currentTotal += currentValue
		compareTotal += prevValue

		displayValue := groupValue
		if strings.TrimSpace(displayValue) == "" {
			displayValue = "(空值)"
		}

		deltaValue := currentValue - prevValue
		deltaRate := safeDivide(deltaValue, prevValue)

		rows = append(rows, LeaderboardRowRes{
			GroupKey:          this.req.GroupBy[0],
			GroupValue:        groupValue,
			GroupDisplayValue: displayValue,
			CurrentValue:      currentValue,
			PrevValue:         prevValue,
			DeltaValue:        deltaValue,
			DeltaRate:         deltaRate,
		})
	}

	for idx := range rows {
		rows[idx].ShareRate = safeDivide(rows[idx].CurrentValue, currentTotal)
	}

	return rows, currentTotal, compareTotal
}

func (this *Leaderboard) sortRows(rows []LeaderboardRowRes) {
	isDesc := this.req.SortOrder != "asc"

	sort.SliceStable(rows, func(i, j int) bool {
		left := this.getSortValue(rows[i])
		right := this.getSortValue(rows[j])
		if left == right {
			if rows[i].CurrentValue == rows[j].CurrentValue {
				return rows[i].GroupDisplayValue < rows[j].GroupDisplayValue
			}
			if isDesc {
				return rows[i].CurrentValue > rows[j].CurrentValue
			}
			return rows[i].CurrentValue < rows[j].CurrentValue
		}
		if isDesc {
			return left > right
		}
		return left < right
	})
}

func (this *Leaderboard) getSortValue(row LeaderboardRowRes) float64 {
	switch this.req.SortBy {
	case leaderboardSortByDelta:
		return row.DeltaValue
	case leaderboardSortByRate:
		return row.DeltaRate
	case leaderboardSortByCurrent:
		fallthrough
	default:
		return row.CurrentValue
	}
}

func (this *Leaderboard) applyTopN(rows []LeaderboardRowRes) ([]LeaderboardRowRes, float64) {
	if len(rows) <= this.req.TopN {
		return rows, 0
	}

	result := rows[:this.req.TopN]
	if !this.req.IncludeOthers {
		return result, 0
	}

	othersRows := rows[this.req.TopN:]
	others := LeaderboardRowRes{
		GroupKey:          this.req.GroupBy[0],
		GroupValue:        "__others__",
		GroupDisplayValue: "其他",
	}

	for _, row := range othersRows {
		others.CurrentValue += row.CurrentValue
		others.PrevValue += row.PrevValue
		others.DeltaValue += row.DeltaValue
	}
	others.DeltaRate = safeDivide(others.DeltaValue, others.PrevValue)
	others.ShareRate = safeDivide(others.CurrentValue, others.CurrentValue+sumCurrentValue(result))

	return append(result, others), others.CurrentValue
}

func sumCurrentValue(rows []LeaderboardRowRes) float64 {
	total := 0.0
	for _, row := range rows {
		total += row.CurrentValue
	}
	return total
}

func safeDivide(dividend, divisor float64) float64 {
	if divisor == 0 {
		return 0
	}
	return dividend / divisor
}

func sanitizeIdentifier(name string) (string, error) {
	trimmed := strings.TrimSpace(name)
	if !safeIdentifierRegexp.MatchString(trimmed) {
		return "", errors.New("字段名不合法")
	}
	return trimmed, nil
}

func escapeSQLStringLiteral(value string) string {
	return strings.ReplaceAll(value, "'", "''")
}
