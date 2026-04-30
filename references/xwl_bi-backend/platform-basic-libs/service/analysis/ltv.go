package analysis

import (
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/1340691923/xwl_bi/engine/db"
	"github.com/1340691923/xwl_bi/engine/logs"
	"github.com/1340691923/xwl_bi/platform-basic-libs/request"
	"github.com/1340691923/xwl_bi/platform-basic-libs/service/analysis/utils"
	parser "github.com/1340691923/xwl_bi/platform-basic-libs/sinker/parse"
	"github.com/1340691923/xwl_bi/platform-basic-libs/util"

	jsoniter "github.com/json-iterator/go"
)

var ltvValueFieldPattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

type ltvAttrMeta struct {
	DataType int `db:"data_type"`
}

type ltvFloat64Array []float64

func (this *ltvFloat64Array) Scan(src interface{}) error {
	switch values := src.(type) {
	case []float64:
		*this = append((*this)[:0], values...)
		return nil
	case []int64:
		result := make([]float64, len(values))
		for index, value := range values {
			result[index] = float64(value)
		}
		*this = result
		return nil
	case []uint64:
		result := make([]float64, len(values))
		for index, value := range values {
			result[index] = float64(value)
		}
		*this = result
		return nil
	case []interface{}:
		result := make([]float64, len(values))
		for index, value := range values {
			switch number := value.(type) {
			case float64:
				result[index] = number
			case int64:
				result[index] = float64(number)
			case uint64:
				result[index] = float64(number)
			case int32:
				result[index] = float64(number)
			case uint32:
				result[index] = float64(number)
			default:
				return fmt.Errorf("unsupported ltv value type %T", value)
			}
		}
		*this = result
		return nil
	default:
		return fmt.Errorf("unsupported ltv values scan type %T", src)
	}
}

type LTV struct {
	sql              string
	args             []interface{}
	revenueValueExpr string
	req              request.LTVReqData
}

func (this *LTV) GetList() (interface{}, error) {
	sqls, args, err := this.GetExecSql()
	if err != nil {
		return nil, err
	}

	logs.Logger.Sugar().Infof("sql", sqls, args, err)

	type Res struct {
		Dates      string          `json:"dates" db:"dates"`
		CohortSize uint64          `json:"cohort_size" db:"cohort_size"`
		Values     ltvFloat64Array `json:"values" db:"values"`
	}

	var res []Res
	err = db.ClickHouseSqlx.Select(&res, sqls, args...)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{"alldata": res}, nil
}

func (this *LTV) appendFilterSQL(whereSQL *string, allArgs *[]interface{}, filter request.AnalysisFilter) error {
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

func (this *LTV) appendUserFilterSQL(whereSQL *string, allArgs *[]interface{}) error {
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

func (this *LTV) buildEventWhereSQL(eventName string, dateCondition string, relation request.AnalysisFilter) (string, []interface{}, error) {
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

func (this *LTV) getRevenueValueExpr() (string, error) {
	payEvent := this.req.ZhibiaoArr[1]
	valueField := strings.TrimSpace(payEvent.ValueField)
	if valueField == "" {
		return "", fmt.Errorf("please select a revenue metric field")
	}

	if !ltvValueFieldPattern.MatchString(valueField) {
		return "", fmt.Errorf("invalid revenue metric field")
	}

	var meta ltvAttrMeta
	err := db.Sqlx.Get(
		&meta,
		`SELECT a.data_type
		FROM attribute a
		INNER JOIN meta_attr_relation mar ON mar.app_id = a.app_id AND mar.event_attr = a.attribute_name
		WHERE a.app_id = ?
		  AND a.attribute_source = 2
		  AND (a.status = 1 OR a.attribute_type = 1)
		  AND mar.event_name = ?
		  AND a.attribute_name = ?
		LIMIT 1`,
		this.req.Appid,
		payEvent.EventName,
		valueField,
	)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return "", err
		}

		err = db.Sqlx.Get(
			&meta,
			`SELECT data_type
			FROM attribute
			WHERE app_id = ?
			  AND attribute_source = 2
			  AND (status = 1 OR attribute_type = 1)
			  AND attribute_name = ?
			LIMIT 1`,
			this.req.Appid,
			valueField,
		)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return "", fmt.Errorf("revenue metric field %s is not available for the selected event", valueField)
			}
			return "", err
		}
	}

	if meta.DataType != parser.Int && meta.DataType != parser.Float {
		return "", fmt.Errorf("revenue metric field must be numeric")
	}

	return fmt.Sprintf("toFloat64(ifNull(%s, 0))", valueField), nil
}

func (this *LTV) parseReqDateRange() (time.Time, time.Time, error) {
	if len(this.req.Date) < 2 {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid date range")
	}

	startTime := util.Str2Time(this.req.Date[0], util.TimeFormatDay2)
	endTime := util.Str2Time(this.req.Date[1], util.TimeFormatDay2)
	if startTime.After(endTime) {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid date range")
	}

	return startTime, endTime, nil
}

func (this *LTV) GetExecSql() (SQL string, allArgs []interface{}, err error) {
	startDate, endDate, err := this.parseReqDateRange()
	if err != nil {
		return "", nil, err
	}

	cohortStartTime := startDate.Format(util.TimeFormat)
	cohortEndTime := endDate.AddDate(0, 0, 1).Format(util.TimeFormat)
	revenueEndTime := endDate.AddDate(0, 0, this.req.WindowTime+1).Format(util.TimeFormat)

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

	payEvent := this.req.ZhibiaoArr[1]
	revenueWhereSQL, revenueArgs, err := this.buildEventWhereSQL(
		payEvent.EventName,
		fmt.Sprintf("xwl_part_date >= toDateTime('%s') AND xwl_part_date < toDateTime('%s')", cohortStartTime, revenueEndTime),
		payEvent.Relation,
	)
	if err != nil {
		return "", nil, err
	}
	allArgs = append(allArgs, revenueArgs...)

	sumArr := make([]string, this.req.WindowTime+1)
	for i := 0; i <= this.req.WindowTime; i++ {
		sumArr[i] = fmt.Sprintf(
			"sumIf(ifNull(T2.revenue_value, 0), dateDiff('day', T1.cohort_date, T2.revenue_date) = %d)",
			i,
		)
	}

	SQL = fmt.Sprintf(`
		SELECT
			formatDateTime(T1.cohort_date, '%%Y-%%m-%%d') AS dates,
			count(distinct T1.xwl_distinct_id) AS cohort_size,
			array(%s) AS values
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
				toDate(xwl_part_date) AS revenue_date,
				sum(%s) AS revenue_value
			FROM xwl_event%d
			WHERE %s
			GROUP BY xwl_distinct_id, revenue_date
		) AS T2
		USING xwl_distinct_id
		GROUP BY T1.cohort_date
		ORDER BY T1.cohort_date
	`, strings.Join(sumArr, ","), this.req.Appid, cohortWhereSQL, this.revenueValueExpr, this.req.Appid, revenueWhereSQL)

	return SQL, allArgs, nil
}

func NewLTV(reqData []byte) (Ianalysis, error) {
	var req request.LTVReqData
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	err := json.Unmarshal(reqData, &req)
	if err != nil {
		return nil, err
	}

	if len(req.Date) < 2 {
		return nil, fmt.Errorf("invalid date range")
	}

	if len(req.ZhibiaoArr) != 2 {
		return nil, fmt.Errorf("LTV requires exactly 2 events")
	}

	if req.WindowTime < 0 {
		return nil, fmt.Errorf("window time cannot be negative")
	}

	obj := &LTV{req: req}

	obj.sql, obj.args, err = utils.GetUserGroupSqlAndArgs(obj.req.UserGroup, obj.req.Appid)
	if err != nil {
		return nil, err
	}

	obj.revenueValueExpr, err = obj.getRevenueValueExpr()
	if err != nil {
		return nil, err
	}

	return obj, nil
}
