package analysis

import (
	"database/sql"
	"fmt"
	"strconv"

	"github.com/1340691923/xwl_bi/engine/db"
	"github.com/1340691923/xwl_bi/engine/logs"
	"github.com/1340691923/xwl_bi/platform-basic-libs/my_error"
	"github.com/1340691923/xwl_bi/platform-basic-libs/request"
	"github.com/1340691923/xwl_bi/platform-basic-libs/service/analysis/utils"
	jsoniter "github.com/json-iterator/go"
)

type Funnel struct {
	sql  string
	args []interface{}
	req  request.FunnelReqData
}

func (this *Funnel) buildWindowSql() (windowSql string, allArgs []interface{}, err error) {
	for _, zhibiao := range this.req.ZhibiaoArr {
		windowSql = windowSql + ","
		windowSql = windowSql + fmt.Sprintf(" xwl_part_event = '%v' ", zhibiao.EventName)

		if len(zhibiao.Relation.Filts) <= 0 {
			continue
		}

		windowSql = windowSql + " and "
		sql, args, _, buildErr := utils.GetWhereSql(zhibiao.Relation)
		if buildErr != nil {
			return "", nil, buildErr
		}

		allArgs = append(allArgs, args...)
		windowSql = windowSql + sql
	}

	return windowSql, allArgs, nil
}

func (this *Funnel) buildFilterSql() (whereFilterSql string, whereFilterArgs []interface{}, userFilterSql string, userFilterArgs []interface{}, err error) {
	if len(this.req.WhereFilterByUser.Filts) > 0 {
		var colArr []string
		var sql string
		sql, userFilterArgs, colArr, err = utils.GetWhereSql(this.req.WhereFilterByUser)
		if err != nil {
			return
		}
		userFilterSql = `and xwl_distinct_id in ( select xwl_distinct_id from ` + utils.GetUserTableView(this.req.Appid, colArr) + ` where ` + sql + ")"
	}

	whereFilterSql, whereFilterArgs, _, err = utils.GetWhereSql(this.req.WhereFilter)
	if err != nil {
		logs.Logger.Sugar().Errorf("req.WhereFilter", this.req.WhereFilter)
		return
	}

	whereFilterSql = whereFilterSql + this.sql
	return
}

func (this *Funnel) buildQueryContext() (startTime string, endTime string, windowSql string, sharedArgs []interface{}, whereFilterSql string, userFilterSql string, err error) {
	startTime = this.req.Date[0] + " 00:00:00"
	endTime = this.req.Date[1] + " 23:59:59"

	windowSql, sharedArgs, err = this.buildWindowSql()
	if err != nil {
		return
	}

	whereFilterArgs := []interface{}{}
	userFilterArgs := []interface{}{}
	whereFilterSql, whereFilterArgs, userFilterSql, userFilterArgs, err = this.buildFilterSql()
	if err != nil {
		return
	}

	sharedArgs = append(sharedArgs, whereFilterArgs...)
	sharedArgs = append(sharedArgs, this.args...)
	sharedArgs = append(sharedArgs, userFilterArgs...)
	return
}

func (this *Funnel) getResultTimeGroupSql() (groupSQL string, groupCol string) {
	switch this.req.ResultTimeFormat {
	case ByDay:
		return "time_group", "formatDateTime(xwl_part_date,'%Y-%m-%d') as time_group"
	case ByHour:
		return "time_group", "formatDateTime(xwl_part_date,'%Y-%m-%d %H:00') as time_group"
	case ByMinute:
		return "time_group", "formatDateTime(xwl_part_date,'%Y-%m-%d %H:%M') as time_group"
	case ByWeek:
		return "time_group", "concat(formatDateTime(toMonday(xwl_part_date),'%Y-%m-%d'),' ~ ',formatDateTime(addDays(toMonday(xwl_part_date), 6),'%Y-%m-%d')) as time_group"
	case Monthly:
		return "time_group", "formatDateTime(xwl_part_date,'%Y-%m') as time_group"
	default:
		return "", ""
	}
}

func (this *Funnel) GetExecSql() (SQL string, allArgs []interface{}, err error) {
	startTime, endTime, windowSql, sharedArgs, whereFilterSql, userFilterSql, err := this.buildQueryContext()
	if err != nil {
		return "", nil, err
	}

	SQL = `SELECT '总体' as groupkey,level_index,count(1) as count,groupUniqArray(xwl_distinct_id) as ui  FROM
			(
				SELECT  xwl_distinct_id,
					arrayWithConstant(windowFunnel_level, 1) levels,
					arrayJoin(arrayEnumerate( levels )) level_index
				  FROM (
					SELECT
					  xwl_distinct_id,
					  windowFunnel(` + strconv.Itoa(this.req.WindowTime) + `)(
						xwl_part_date
						` + windowSql + `
					  ) AS windowFunnel_level
					FROM  xwl_event` + strconv.Itoa(this.req.Appid) + `
					WHERE xwl_part_date >= toDateTime('` + startTime + `') and xwl_part_date <= toDateTime('` + endTime + `') and ` + whereFilterSql + ` ` + userFilterSql + `
					GROUP BY xwl_distinct_id
				)
			)
			group by level_index
			ORDER BY level_index limit 1000
	`

	allArgs = append(allArgs, sharedArgs...)

	if len(this.req.GroupBy) > 0 {
		groupSql := `SELECT toString(groupkey) as groupkey,level_index,count(1) as count ,groupUniqArray(xwl_distinct_id) as ui FROM
			(
				SELECT  xwl_distinct_id, groupkey,
					arrayWithConstant(windowFunnel_level, 1) levels,
					arrayJoin(arrayEnumerate( levels )) level_index
				  FROM (
					SELECT
					  xwl_distinct_id, ` + this.req.GroupBy[0] + ` groupkey,
					  windowFunnel(` + strconv.Itoa(this.req.WindowTime) + `)(
						xwl_part_date
						` + windowSql + `
					  ) AS windowFunnel_level
					  FROM xwl_event` + strconv.Itoa(this.req.Appid) + `
					  WHERE xwl_part_date >= toDateTime('` + startTime + `') and xwl_part_date <= toDateTime('` + endTime + `') and ` + whereFilterSql + ` ` + userFilterSql + `
					GROUP BY xwl_distinct_id,groupkey
				)
			)
			group by groupkey,level_index
			ORDER BY groupkey,level_index limit 1000
	`
		SQL = fmt.Sprintf("%s UNION ALL %s", SQL, groupSql)
		allArgs = append(allArgs, sharedArgs...)
	}

	return SQL, allArgs, nil
}

func (this *Funnel) GetTimeGroupExecSql() (SQL string, allArgs []interface{}, err error) {
	timeGroupSql, timeGroupCol := this.getResultTimeGroupSql()
	if timeGroupSql == "" || timeGroupCol == "" {
		return "", nil, nil
	}

	startTime, endTime, windowSql, sharedArgs, whereFilterSql, userFilterSql, err := this.buildQueryContext()
	if err != nil {
		return "", nil, err
	}

	SQL = `SELECT toString(time_group) as time_group,'总体' as groupkey,level_index,count(1) as count,groupUniqArray(xwl_distinct_id) as ui FROM
			(
				SELECT xwl_distinct_id,time_group,
					arrayWithConstant(windowFunnel_level, 1) levels,
					arrayJoin(arrayEnumerate(levels)) level_index
				FROM (
					SELECT
						xwl_distinct_id,
						` + timeGroupCol + `,
						windowFunnel(` + strconv.Itoa(this.req.WindowTime) + `)(
							xwl_part_date
							` + windowSql + `
						) AS windowFunnel_level
					FROM xwl_event` + strconv.Itoa(this.req.Appid) + `
					WHERE xwl_part_date >= toDateTime('` + startTime + `') and xwl_part_date <= toDateTime('` + endTime + `') and ` + whereFilterSql + ` ` + userFilterSql + `
					GROUP BY xwl_distinct_id,` + timeGroupSql + `
				)
			)
			GROUP BY time_group,level_index
			ORDER BY time_group,level_index limit 10000
	`

	allArgs = append(allArgs, sharedArgs...)

	if len(this.req.GroupBy) > 0 {
		groupSql := `SELECT toString(time_group) as time_group,toString(groupkey) as groupkey,level_index,count(1) as count,groupUniqArray(xwl_distinct_id) as ui FROM
			(
				SELECT xwl_distinct_id,time_group,groupkey,
					arrayWithConstant(windowFunnel_level, 1) levels,
					arrayJoin(arrayEnumerate(levels)) level_index
				FROM (
					SELECT
						xwl_distinct_id,
						` + timeGroupCol + `,
						` + this.req.GroupBy[0] + ` groupkey,
						windowFunnel(` + strconv.Itoa(this.req.WindowTime) + `)(
							xwl_part_date
							` + windowSql + `
						) AS windowFunnel_level
					FROM xwl_event` + strconv.Itoa(this.req.Appid) + `
					WHERE xwl_part_date >= toDateTime('` + startTime + `') and xwl_part_date <= toDateTime('` + endTime + `') and ` + whereFilterSql + ` ` + userFilterSql + `
					GROUP BY xwl_distinct_id,` + timeGroupSql + `,groupkey
				)
			)
			GROUP BY time_group,groupkey,level_index
			ORDER BY time_group,groupkey,level_index limit 10000
	`
		SQL = fmt.Sprintf("%s UNION ALL %s", SQL, groupSql)
		allArgs = append(allArgs, sharedArgs...)
	}

	return SQL, allArgs, nil
}

type FunnelRes struct {
	LevelIndex int      `json:"level_index" db:"level_index"`
	Count      int      `json:"count" db:"count"`
	UI         []string `json:"ui" db:"ui"`
}

type FunnelGroupRes struct {
	Groupkey   sql.NullString `json:"groupkey" db:"groupkey"`
	LevelIndex int            `json:"level_index" db:"level_index"`
	UI         []string       `json:"ui" db:"ui"`
	Count      int            `json:"count" db:"count"`
}

type FunnelTimeGroupRes struct {
	TimeGroup  sql.NullString `json:"time_group" db:"time_group"`
	Groupkey   sql.NullString `json:"groupkey" db:"groupkey"`
	LevelIndex int            `json:"level_index" db:"level_index"`
	UI         []string       `json:"ui" db:"ui"`
	Count      int            `json:"count" db:"count"`
}

func (this *Funnel) GetList() (interface{}, error) {
	sql, args, err := this.GetExecSql()
	if err != nil {
		return nil, err
	}

	logs.Logger.Sugar().Infof("SQL", sql, args)

	var funnelGroupResList []FunnelGroupRes
	if err := db.ClickHouseSqlx.Select(&funnelGroupResList, sql, args...); err != nil {
		return nil, err
	}

	groupData := map[string][]FunnelRes{}
	for _, v := range funnelGroupResList {
		if _, ok := groupData[v.Groupkey.String]; !ok {
			groupData[v.Groupkey.String] = []FunnelRes{}
		}
		groupData[v.Groupkey.String] = append(groupData[v.Groupkey.String], FunnelRes{
			LevelIndex: v.LevelIndex,
			Count:      v.Count,
			UI:         v.UI,
		})
	}

	timeGroupData := map[string]map[string][]FunnelRes{}
	timeSql, timeArgs, err := this.GetTimeGroupExecSql()
	if err != nil {
		return nil, err
	}
	if timeSql != "" {
		var funnelTimeGroupResList []FunnelTimeGroupRes
		if err := db.ClickHouseSqlx.Select(&funnelTimeGroupResList, timeSql, timeArgs...); err != nil {
			return nil, err
		}

		for _, v := range funnelTimeGroupResList {
			timeKey := v.TimeGroup.String
			groupKey := v.Groupkey.String
			if _, ok := timeGroupData[timeKey]; !ok {
				timeGroupData[timeKey] = map[string][]FunnelRes{}
			}
			if _, ok := timeGroupData[timeKey][groupKey]; !ok {
				timeGroupData[timeKey][groupKey] = []FunnelRes{}
			}
			timeGroupData[timeKey][groupKey] = append(timeGroupData[timeKey][groupKey], FunnelRes{
				LevelIndex: v.LevelIndex,
				Count:      v.Count,
				UI:         v.UI,
			})
		}
	}

	return map[string]interface{}{"groupData": groupData, "timeGroupData": timeGroupData}, nil
}

func NewFunnel(reqData []byte) (Ianalysis, error) {
	obj := &Funnel{}
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	err := json.Unmarshal(reqData, &obj.req)
	if err != nil {
		return nil, err
	}

	if len(obj.req.Date) < 2 {
		return nil, my_error.NewBusiness(ERROR_TABLE, TimeError)
	}
	if len(obj.req.ZhibiaoArr) > 30 {
		return nil, my_error.NewBusiness(ERROR_TABLE, ZhiBiaoNumError)
	}
	if obj.req.ResultTimeFormat == "" {
		obj.req.ResultTimeFormat = ByTotal
	}

	var T int
	switch obj.req.WindowTimeFormat {
	case "天":
		T = 60 * 60 * 24
	case "小时":
		T = 60 * 60
	case "分钟":
		T = 60
	case "秒":
		T = 1
	default:
		return nil, my_error.NewBusiness(ERROR_TABLE, TimeError)
	}
	obj.req.WindowTime = obj.req.WindowTime * T
	obj.sql, obj.args, err = utils.GetUserGroupSqlAndArgs(obj.req.UserGroup, obj.req.Appid)
	if err != nil {
		return nil, err
	}

	return obj, nil
}
