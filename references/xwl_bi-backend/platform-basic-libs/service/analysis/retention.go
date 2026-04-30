package analysis

import (
	"fmt"
	"strings"
	"time"

	"github.com/1340691923/xwl_bi/engine/db"
	"github.com/1340691923/xwl_bi/engine/logs"
	"github.com/1340691923/xwl_bi/platform-basic-libs/my_error"
	"github.com/1340691923/xwl_bi/platform-basic-libs/request"
	"github.com/1340691923/xwl_bi/platform-basic-libs/service/analysis/utils"
	"github.com/1340691923/xwl_bi/platform-basic-libs/util"

	jsoniter "github.com/json-iterator/go"
)

type Retention struct {
	sql  string
	args []interface{}
	req  request.RetentionReqData
}

type retentionResultRow struct {
	Dates string     `json:"dates" db:"dates"`
	Value []uint64   `json:"value" db:"value"`
	UI    [][]string `json:"ui" db:"ui"`
}

const maxRetentionRangeDays = 720

func (this *Retention) GetList() (interface{}, error) {
	sqls, args, err := this.GetExecSql()
	if err != nil {
		return nil, err
	}

	logs.Logger.Sugar().Infof("sql", sqls, args, err)

	var rows []retentionResultRow
	err = db.ClickHouseSqlx.Select(&rows, sqls, args...)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{"alldata": this.fillMissingRows(rows)}, nil
}

func (this *Retention) appendFilterSQL(whereSQL *string, allArgs *[]interface{}, filter request.AnalysisFilter) error {
	if len(filter.Filts) == 0 {
		return nil
	}

	sqlPart, args, _, err := utils.GetWhereSql(filter)
	if err != nil {
		return err
	}

	*whereSQL += " AND " + sqlPart
	*allArgs = append(*allArgs, args...)
	return nil
}

func (this *Retention) appendUserFilterSQL(whereSQL *string, allArgs *[]interface{}) error {
	if len(this.req.WhereFilterByUser.Filts) == 0 {
		return nil
	}

	sqlPart, args, colArr, err := utils.GetWhereSql(this.req.WhereFilterByUser)
	if err != nil {
		return err
	}

	*whereSQL += fmt.Sprintf(
		" AND xwl_distinct_id IN (SELECT xwl_distinct_id FROM %s WHERE %s)",
		utils.GetUserTableView(this.req.Appid, colArr),
		sqlPart,
	)
	*allArgs = append(*allArgs, args...)
	return nil
}

func (this *Retention) buildEventWhereSQL(eventName string, dateCondition string, relation request.AnalysisFilter) (string, []interface{}, error) {
	whereSQL := fmt.Sprintf("xwl_part_event = '%s' AND %s", eventName, dateCondition)
	allArgs := make([]interface{}, 0)

	if err := this.appendFilterSQL(&whereSQL, &allArgs, relation); err != nil {
		return "", nil, err
	}

	if err := this.appendFilterSQL(&whereSQL, &allArgs, this.req.WhereFilter); err != nil {
		return "", nil, err
	}

	if err := this.appendUserFilterSQL(&whereSQL, &allArgs); err != nil {
		return "", nil, err
	}

	whereSQL += this.sql
	allArgs = append(allArgs, this.args...)
	return whereSQL, allArgs, nil
}

func (this *Retention) parseReqDateRange() (time.Time, time.Time, error) {
	if len(this.req.Date) < 2 {
		return time.Time{}, time.Time{}, my_error.NewBusiness(ERROR_TABLE, TimeError)
	}

	startTime := util.Str2Time(this.req.Date[0], util.TimeFormatDay2)
	endTime := util.Str2Time(this.req.Date[1], util.TimeFormatDay2)
	if startTime.After(endTime) {
		return time.Time{}, time.Time{}, my_error.NewBusiness(ERROR_TABLE, TimeError)
	}

	return startTime, endTime, nil
}

func (this *Retention) zeroValueSlots() int {
	return this.req.WindowTime + 2
}

func (this *Retention) newZeroRow(date string) retentionResultRow {
	slots := this.zeroValueSlots()
	row := retentionResultRow{
		Dates: date,
		Value: make([]uint64, slots),
		UI:    make([][]string, slots),
	}
	for index := 0; index < slots; index++ {
		row.UI[index] = []string{}
	}
	return row
}

func (this *Retention) normalizeRow(row retentionResultRow) retentionResultRow {
	target := this.newZeroRow(row.Dates)
	copy(target.Value, row.Value)
	for index := 0; index < len(target.UI) && index < len(row.UI); index++ {
		if row.UI[index] == nil {
			continue
		}
		target.UI[index] = append(target.UI[index], row.UI[index]...)
	}
	return target
}

func (this *Retention) fillMissingRows(rows []retentionResultRow) []retentionResultRow {
	rowMap := make(map[string]retentionResultRow, len(rows))
	for _, row := range rows {
		rowMap[row.Dates] = this.normalizeRow(row)
	}

	dateList := this.parseReqDate()
	result := make([]retentionResultRow, 0, len(dateList))
	for _, date := range dateList {
		dateText := date.Format(util.TimeFormatDay2)
		if row, ok := rowMap[dateText]; ok {
			result = append(result, row)
			continue
		}
		result = append(result, this.newZeroRow(dateText))
	}
	return result
}

func (this *Retention) parseReqDate() []time.Time {
	if len(this.req.Date) < 2 {
		return nil
	}

	startTimeFormat := this.req.Date[0]
	endTimeFormat := this.req.Date[1]

	if startTimeFormat == endTimeFormat {
		t := make([]time.Time, 1)
		t[0] = util.Str2Time(startTimeFormat, util.TimeFormatDay2)
		return t
	}
	startT := util.Str2Time(startTimeFormat, util.TimeFormatDay2)
	endT := util.Str2Time(endTimeFormat, util.TimeFormatDay2)

	t := []time.Time{}
	for ; startT.Before(endT.AddDate(0, 0, 1)); startT = startT.AddDate(0, 0, 1) {
		t = append(t, startT)
	}

	return t
}

func (this *Retention) validateDateRange() error {
	startDate, endDate, err := this.parseReqDateRange()
	if err != nil {
		return err
	}

	dateList := this.parseReqDate()
	if len(dateList) == 0 {
		return my_error.NewBusiness(ERROR_TABLE, TimeError)
	}

	if startDate.After(endDate) {
		return my_error.NewBusiness(ERROR_TABLE, TimeError)
	}

	if len(dateList) > maxRetentionRangeDays {
		return my_error.NewBusiness(ERROR_TABLE, RetentionRangeError)
	}

	return nil
}

func (this *Retention) GetExecSql() (SQL string, allArgs []interface{}, err error) {
	startDate, endDate, err := this.parseReqDateRange()
	if err != nil {
		return "", nil, err
	}

	cohortStartTime := startDate.Format(util.TimeFormat)
	cohortEndTime := endDate.AddDate(0, 0, 1).Format(util.TimeFormat)
	returnEndTime := endDate.AddDate(0, 0, this.req.WindowTime+1).Format(util.TimeFormat)

	startEvent := this.req.ZhibiaoArr[0]
	cohortWhereSQL, cohortArgs, err := this.buildEventWhereSQL(
		startEvent.EventName,
		fmt.Sprintf("xwl_part_date >= toDateTime('%s') AND xwl_part_date < toDateTime('%s')", cohortStartTime, cohortEndTime),
		startEvent.Relation,
	)
	if err != nil {
		return "", nil, err
	}
	allArgs = append(allArgs, cohortArgs...)

	returnEvent := this.req.ZhibiaoArr[1]
	returnWhereSQL, returnArgs, err := this.buildEventWhereSQL(
		returnEvent.EventName,
		fmt.Sprintf("xwl_part_date >= toDateTime('%s') AND xwl_part_date < toDateTime('%s')", cohortStartTime, returnEndTime),
		returnEvent.Relation,
	)
	if err != nil {
		return "", nil, err
	}
	allArgs = append(allArgs, returnArgs...)

	valueExprs := []string{"count()"}
	uiExprs := []string{"groupArray(T1.xwl_distinct_id)"}
	for offset := 0; offset <= this.req.WindowTime; offset++ {
		returnDateExpr := "T1.cohort_date"
		if offset > 0 {
			returnDateExpr = fmt.Sprintf("addDays(T1.cohort_date, %d)", offset)
		}
		condition := fmt.Sprintf("has(T2.return_dates, %s)", returnDateExpr)
		valueExprs = append(valueExprs, fmt.Sprintf("countIf(%s)", condition))
		uiExprs = append(uiExprs, fmt.Sprintf("groupArrayIf(T1.xwl_distinct_id, %s)", condition))
	}

	SQL = fmt.Sprintf(`
		SELECT
			formatDateTime(T1.cohort_date, '%%Y-%%m-%%d') AS dates,
			array(%s) AS value,
			array(%s) AS ui
		FROM
		(
			SELECT
				xwl_distinct_id,
				toDate(xwl_part_date) AS cohort_date
			FROM xwl_event%d
			WHERE %s
			GROUP BY xwl_distinct_id, cohort_date
		) AS T1
		ALL LEFT JOIN
		(
			SELECT
				xwl_distinct_id,
				groupUniqArray(toDate(xwl_part_date)) AS return_dates
			FROM xwl_event%d
			WHERE %s
			GROUP BY xwl_distinct_id
		) AS T2
		USING xwl_distinct_id
		GROUP BY T1.cohort_date
		ORDER BY T1.cohort_date
	`, strings.Join(valueExprs, ","), strings.Join(uiExprs, ","), this.req.Appid, cohortWhereSQL, this.req.Appid, returnWhereSQL)

	return SQL, allArgs, nil
}

func NewRetention(reqData []byte) (Ianalysis, error) {
	obj := &Retention{}
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	err := json.Unmarshal(reqData, &obj.req)
	if err != nil {
		return nil, err
	}
	if len(obj.req.Date) < 2 {
		return nil, my_error.NewBusiness(ERROR_TABLE, TimeError)
	}
	if len(obj.req.ZhibiaoArr) != 2 {
		return nil, my_error.NewBusiness(ERROR_TABLE, ZhiBiaoNumError)
	}
	if err := obj.validateDateRange(); err != nil {
		return nil, err
	}

	obj.sql, obj.args, err = utils.GetUserGroupSqlAndArgs(obj.req.UserGroup, obj.req.Appid)
	if err != nil {
		return nil, err
	}

	return obj, nil
}
