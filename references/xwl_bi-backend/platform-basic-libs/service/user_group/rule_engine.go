package user_group

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/1340691923/xwl_bi/engine/db"
	"github.com/1340691923/xwl_bi/model"
	"github.com/1340691923/xwl_bi/platform-basic-libs/request"
	"github.com/1340691923/xwl_bi/platform-basic-libs/service/analysis"
	analysisUtils "github.com/1340691923/xwl_bi/platform-basic-libs/service/analysis/utils"
	"github.com/1340691923/xwl_bi/platform-basic-libs/util"
	"strconv"
	"strings"
	"time"
)

type UserGroupRuleContent struct {
	Version           int                        `json:"version"`
	Relation          string                     `json:"relation"`
	BehaviorCondition UserGroupBehaviorCondition `json:"behaviorCondition"`
	UserAttrCondition UserGroupUserAttrCondition `json:"userAttrCondition"`
	SequenceCondition map[string]interface{}     `json:"sequenceCondition"`
}

type UserGroupBehaviorCondition struct {
	Enabled  bool                    `json:"enabled"`
	Relation string                  `json:"relation"`
	Items    []UserGroupBehaviorRule `json:"items"`
}

type UserGroupUserAttrCondition struct {
	Enabled      bool                   `json:"enabled"`
	Relation     string                 `json:"relation"`
	Items        []UserGroupUserAttrRule `json:"items"`
	AttrFilter   request.AnalysisFilter `json:"attrFilter"`
	UserGroupIds []int                  `json:"userGroupIds"`
}

type UserGroupUserAttrRule struct {
	SourceType   string                 `json:"sourceType"`
	AttrFilter   request.AnalysisFilter `json:"attrFilter"`
	UserGroupIds []int                  `json:"userGroupIds"`
}

type UserGroupBehaviorRule struct {
	DateRange    UserGroupDateRange     `json:"dateRange"`
	BehaviorType string                 `json:"behaviorType"`
	EventName    string                 `json:"eventName"`
	Comparator   string                 `json:"comparator"`
	Value        interface{}            `json:"value"`
	RangeValue   []interface{}          `json:"rangeValue"`
	EventFilter  request.AnalysisFilter `json:"eventFilter"`
	UserFilter   request.AnalysisFilter `json:"userFilter"`
	UserGroupIds []int                  `json:"userGroupIds"`
}

type UserGroupDateRange struct {
	Mode  string                `json:"mode"`
	Start UserGroupDateEndpoint `json:"start"`
	End   UserGroupDateEndpoint `json:"end"`
}

type UserGroupDateEndpoint struct {
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
}

type eventCountRow struct {
	DistinctID string `db:"xwl_distinct_id"`
	EventCount int64  `db:"event_count"`
}

type distinctIDRow struct {
	DistinctID string `db:"xwl_distinct_id"`
}

type funnelSequenceEvalResult struct {
	GroupData map[string][]struct {
		LevelIndex int      `json:"level_index"`
		Count      int      `json:"count"`
		UI         []string `json:"ui"`
	} `json:"groupData"`
}

type traceSequenceEvalResult struct {
	TableRes []struct {
		Trace     string   `json:"trace"`
		UserCount uint64   `json:"user_count"`
		UI        []string `json:"ui"`
	} `json:"tableRes"`
}

func emptyAnalysisFilter() request.AnalysisFilter {
	return request.AnalysisFilter{
		FilterType: "COMPOUND",
		Filts:      []struct {
			FilterType string      `json:"filterType"`
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
	}
}

func hasAnalysisFilter(filter request.AnalysisFilter) bool {
	return len(filter.Filts) > 0
}

func normalizeRuleRelation(relation string) string {
	switch strings.ToUpper(strings.TrimSpace(relation)) {
	case "OR", "或":
		return "OR"
	default:
		return "AND"
	}
}

func newUserSetFromSlice(items []string) map[string]struct{} {
	res := make(map[string]struct{}, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		res[item] = struct{}{}
	}
	return res
}

func userSetToSlice(userSet map[string]struct{}) []string {
	res := make([]string, 0, len(userSet))
	for uid := range userSet {
		res = append(res, uid)
	}
	return res
}

func unionUserSet(left, right map[string]struct{}) map[string]struct{} {
	if left == nil && right == nil {
		return nil
	}
	res := make(map[string]struct{}, len(left)+len(right))
	for uid := range left {
		res[uid] = struct{}{}
	}
	for uid := range right {
		res[uid] = struct{}{}
	}
	return res
}

func intersectUserSet(left, right map[string]struct{}) map[string]struct{} {
	if left == nil {
		return right
	}
	if right == nil {
		return left
	}
	res := make(map[string]struct{})
	if len(left) > len(right) {
		left, right = right, left
	}
	for uid := range left {
		if _, ok := right[uid]; ok {
			res[uid] = struct{}{}
		}
	}
	return res
}

func subtractUserSet(base, removed map[string]struct{}) map[string]struct{} {
	if base == nil {
		return nil
	}
	if removed == nil {
		return base
	}
	res := make(map[string]struct{}, len(base))
	for uid := range base {
		if _, ok := removed[uid]; ok {
			continue
		}
		res[uid] = struct{}{}
	}
	return res
}

func combineUserSetsByRelation(relation string, sets ...map[string]struct{}) map[string]struct{} {
	if len(sets) == 0 {
		return nil
	}
	normalizedRelation := normalizeRuleRelation(relation)
	var res map[string]struct{}
	for _, item := range sets {
		if item == nil {
			continue
		}
		if res == nil {
			res = item
			continue
		}
		if normalizedRelation == "OR" {
			res = unionUserSet(res, item)
			continue
		}
		res = intersectUserSet(res, item)
	}
	return res
}

func parseInt64Value(value interface{}) (int64, error) {
	switch v := value.(type) {
	case nil:
		return 0, errors.New("空值")
	case int:
		return int64(v), nil
	case int32:
		return int64(v), nil
	case int64:
		return v, nil
	case float64:
		return int64(v), nil
	case float32:
		return int64(v), nil
	case json.Number:
		return v.Int64()
	case string:
		v = strings.TrimSpace(v)
		if v == "" {
			return 0, errors.New("空值")
		}
		return strconv.ParseInt(v, 10, 64)
	default:
		return 0, fmt.Errorf("无法识别的数值类型:%T", value)
	}
}

func evaluateCountComparator(count int64, comparator string, value interface{}, rangeValue []interface{}) bool {
	switch comparator {
	case "range":
		if len(rangeValue) < 2 {
			return false
		}
		startValue, err := parseInt64Value(rangeValue[0])
		if err != nil {
			return false
		}
		endValue, err := parseInt64Value(rangeValue[1])
		if err != nil {
			return false
		}
		return count >= startValue && count <= endValue
	case "!=", "<", "<=", ">", ">=", "=":
		targetValue, err := parseInt64Value(value)
		if err != nil {
			return false
		}
		switch comparator {
		case "!=":
			return count != targetValue
		case "<":
			return count < targetValue
		case "<=":
			return count <= targetValue
		case ">":
			return count > targetValue
		case ">=":
			return count >= targetValue
		default:
			return count == targetValue
		}
	default:
		targetValue, err := parseInt64Value(value)
		if err != nil {
			return false
		}
		return count == targetValue
	}
}

func parseDateValue(value interface{}) (time.Time, error) {
	switch v := value.(type) {
	case string:
		return util.Str2Time(v, util.TimeFormatDay2), nil
	default:
		return time.Time{}, errors.New("无法解析日期值")
	}
}

func normalizeDayStart(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func normalizeDayEnd(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 0, t.Location())
}

func (this *UserGroupService) resolveDateEndpoint(endpoint UserGroupDateEndpoint) (time.Time, error) {
	today := normalizeDayStart(time.Now())
	switch strings.TrimSpace(endpoint.Type) {
	case "date":
		return parseDateValue(endpoint.Value)
	case "today":
		return today, nil
	case "yesterday":
		return today.AddDate(0, 0, -1), nil
	case "relative_past_n_days":
		n, err := parseInt64Value(endpoint.Value)
		if err != nil {
			return time.Time{}, err
		}
		return today.AddDate(0, 0, -int(n)), nil
	default:
		return time.Time{}, fmt.Errorf("暂不支持的日期类型:%s", endpoint.Type)
	}
}

func (this *UserGroupService) resolveDateRange(dateRange UserGroupDateRange) (time.Time, time.Time, error) {
	startDate, err := this.resolveDateEndpoint(dateRange.Start)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	endDate, err := this.resolveDateEndpoint(dateRange.End)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	startDate = normalizeDayStart(startDate)
	endDate = normalizeDayEnd(endDate)
	if startDate.After(endDate) {
		return time.Time{}, time.Time{}, errors.New("行为规则开始时间不能大于结束时间")
	}
	return startDate, endDate, nil
}

func (this *UserGroupService) queryAllUsers() (map[string]struct{}, error) {
	sql := `select xwl_distinct_id from ` + analysisUtils.GetUserTableView(this.Appid, []string{})
	rows := []distinctIDRow{}
	if err := db.ClickHouseSqlx.Select(&rows, sql); err != nil {
		return nil, err
	}
	userIDs := make([]string, 0, len(rows))
	for _, row := range rows {
		userIDs = append(userIDs, row.DistinctID)
	}
	return newUserSetFromSlice(userIDs), nil
}

func (this *UserGroupService) queryUsersByFilter(filter request.AnalysisFilter) (map[string]struct{}, error) {
	if !hasAnalysisFilter(filter) {
		return nil, nil
	}
	whereSQL, args, cols, err := analysisUtils.GetWhereSql(filter)
	if err != nil {
		return nil, err
	}
	sql := `select xwl_distinct_id from ` + analysisUtils.GetUserTableView(this.Appid, cols) + ` where ` + whereSQL
	rows := []distinctIDRow{}
	if err := db.ClickHouseSqlx.Select(&rows, sql, args...); err != nil {
		return nil, err
	}
	userIDs := make([]string, 0, len(rows))
	for _, row := range rows {
		userIDs = append(userIDs, row.DistinctID)
	}
	return newUserSetFromSlice(userIDs), nil
}

func (this *UserGroupService) queryUserGroupSet(ids []int) (map[string]struct{}, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	sql, args, err := db.SqlBuilder.
		Select("id,user_list").
		From("user_group").
		Where(db.Eq{
			"create_by": this.ManagerID,
			"appid":     this.Appid,
			"id":        ids,
		}).
		ToSql()
	if err != nil {
		return nil, err
	}

	groups := []model.UserGroup{}
	if err := db.Sqlx.Select(&groups, sql, args...); err != nil {
		return nil, err
	}

	result := make(map[string]struct{})
	for _, group := range groups {
		userIDs, unzipErr := unzipUserIDs(group.UserList)
		if unzipErr != nil {
			return nil, unzipErr
		}
		for _, uid := range userIDs {
			result[uid] = struct{}{}
		}
	}
	return result, nil
}

func (this *UserGroupService) getCandidateUserSet(userFilter request.AnalysisFilter, userGroupIDs []int) (map[string]struct{}, error) {
	userFilterSet, err := this.queryUsersByFilter(userFilter)
	if err != nil {
		return nil, err
	}
	userGroupSet, err := this.queryUserGroupSet(userGroupIDs)
	if err != nil {
		return nil, err
	}
	if userFilterSet == nil && userGroupSet == nil {
		return this.queryAllUsers()
	}
	return intersectUserSet(userFilterSet, userGroupSet), nil
}

func (this *UserGroupService) queryEventRuleDoers(rule UserGroupBehaviorRule) (map[string]struct{}, error) {
	startDate, endDate, err := this.resolveDateRange(rule.DateRange)
	if err != nil {
		return nil, err
	}

	args := []interface{}{startDate.Format(util.TimeFormat), endDate.Format(util.TimeFormat), rule.EventName}
	sql := fmt.Sprintf("select xwl_distinct_id, count(*) as event_count from xwl_event%d where xwl_part_date >= toDateTime(?) and xwl_part_date <= toDateTime(?) and xwl_part_event = ?", this.Appid)

	if hasAnalysisFilter(rule.EventFilter) {
		whereSQL, whereArgs, _, eventErr := analysisUtils.GetWhereSql(rule.EventFilter)
		if eventErr != nil {
			return nil, eventErr
		}
		sql += " and " + whereSQL
		args = append(args, whereArgs...)
	}

	if hasAnalysisFilter(rule.UserFilter) {
		userWhereSQL, userArgs, cols, userErr := analysisUtils.GetWhereSql(rule.UserFilter)
		if userErr != nil {
			return nil, userErr
		}
		sql += " and xwl_distinct_id in (select xwl_distinct_id from " + analysisUtils.GetUserTableView(this.Appid, cols) + " where " + userWhereSQL + ")"
		args = append(args, userArgs...)
	}

	sql += " group by xwl_distinct_id"

	rows := []eventCountRow{}
	if err := db.ClickHouseSqlx.Select(&rows, sql, args...); err != nil {
		return nil, err
	}

	userSet := make(map[string]struct{})
	for _, row := range rows {
		if evaluateCountComparator(row.EventCount, rule.Comparator, rule.Value, rule.RangeValue) {
			userSet[row.DistinctID] = struct{}{}
		}
	}

	userGroupSet, err := this.queryUserGroupSet(rule.UserGroupIds)
	if err != nil {
		return nil, err
	}
	return intersectUserSet(userSet, userGroupSet), nil
}

func (this *UserGroupService) evaluateBehaviorRule(rule UserGroupBehaviorRule) (map[string]struct{}, error) {
	doerSet, err := this.queryEventRuleDoers(rule)
	if err != nil {
		return nil, err
	}

	if strings.ToLower(strings.TrimSpace(rule.BehaviorType)) != "not_done" {
		return doerSet, nil
	}

	candidateSet, err := this.getCandidateUserSet(rule.UserFilter, rule.UserGroupIds)
	if err != nil {
		return nil, err
	}
	if !evaluateCountComparator(0, rule.Comparator, rule.Value, rule.RangeValue) {
		return map[string]struct{}{}, nil
	}
	return subtractUserSet(candidateSet, doerSet), nil
}

func (this *UserGroupService) evaluateBehaviorCondition(condition UserGroupBehaviorCondition) (map[string]struct{}, error) {
	if !condition.Enabled || len(condition.Items) == 0 {
		return nil, nil
	}
	itemSets := make([]map[string]struct{}, 0, len(condition.Items))
	for _, item := range condition.Items {
		itemSet, err := this.evaluateBehaviorRule(item)
		if err != nil {
			return nil, err
		}
		itemSets = append(itemSets, itemSet)
	}
	return combineUserSetsByRelation(condition.Relation, itemSets...), nil
}

func (this *UserGroupService) evaluateUserAttrCondition(condition UserGroupUserAttrCondition) (map[string]struct{}, error) {
	if !condition.Enabled {
		return nil, nil
	}

	if len(condition.Items) > 0 {
		itemSets := make([]map[string]struct{}, 0, len(condition.Items))
		for _, item := range condition.Items {
			var itemSet map[string]struct{}
			var err error
			switch strings.TrimSpace(item.SourceType) {
			case "", "attr":
				itemSet, err = this.queryUsersByFilter(item.AttrFilter)
			case "user_group":
				itemSet, err = this.queryUserGroupSet(item.UserGroupIds)
			default:
				continue
			}
			if err != nil {
				return nil, err
			}
			itemSets = append(itemSets, itemSet)
		}
		return combineUserSetsByRelation(condition.Relation, itemSets...), nil
	}

	attrSet, err := this.queryUsersByFilter(condition.AttrFilter)
	if err != nil {
		return nil, err
	}
	userGroupSet, err := this.queryUserGroupSet(condition.UserGroupIds)
	if err != nil {
		return nil, err
	}
	if attrSet == nil && userGroupSet == nil {
		return nil, nil
	}
	return combineUserSetsByRelation(condition.Relation, attrSet, userGroupSet), nil
}

func (this *UserGroupService) evaluateSequenceCondition(sequenceCondition map[string]interface{}) ([]string, error) {
	if len(sequenceCondition) == 0 {
		return nil, errors.New("行为序列规则不能为空")
	}

	analysisType, _ := sequenceCondition["analysisType"].(string)
	switch analysisType {
	case "funnel":
		return this.evaluateFunnelSequenceCondition(sequenceCondition)
	case "trace":
		return this.evaluateTraceSequenceCondition(sequenceCondition)
	default:
		return nil, fmt.Errorf("暂不支持的行为序列类型:%s", analysisType)
	}
}

func (this *UserGroupService) evaluateFunnelSequenceCondition(sequenceCondition map[string]interface{}) ([]string, error) {
	formRaw, ok := sequenceCondition["form"]
	if !ok {
		return nil, errors.New("漏斗规则缺少分析表单")
	}
	contextRaw, ok := sequenceCondition["context"]
	if !ok {
		return nil, errors.New("漏斗规则缺少上下文")
	}

	formBytes, err := json.Marshal(formRaw)
	if err != nil {
		return nil, err
	}

	funnelReq := request.FunnelReqData{}
	if err := json.Unmarshal(formBytes, &funnelReq); err != nil {
		return nil, err
	}
	funnelReq.Appid = this.Appid

	contextBytes, err := json.Marshal(contextRaw)
	if err != nil {
		return nil, err
	}
	context := struct {
		GroupKey   string `json:"groupKey"`
		LevelIndex int    `json:"levelIndex"`
	}{}
	if err := json.Unmarshal(contextBytes, &context); err != nil {
		return nil, err
	}

	reqBytes, err := json.Marshal(funnelReq)
	if err != nil {
		return nil, err
	}
	i, err := analysis.NewAnalysisByCommand(analysis.FunnelComand, reqBytes)
	if err != nil {
		return nil, err
	}
	res, err := analysis.GetAnalysisRes(i)
	if err != nil {
		return nil, err
	}

	resultBytes, err := json.Marshal(res)
	if err != nil {
		return nil, err
	}
	result := funnelSequenceEvalResult{}
	if err := json.Unmarshal(resultBytes, &result); err != nil {
		return nil, err
	}

	groupRows, ok := result.GroupData[context.GroupKey]
	if !ok {
		return []string{}, nil
	}
	for _, row := range groupRows {
		if row.LevelIndex == context.LevelIndex {
			return row.UI, nil
		}
	}
	return []string{}, nil
}

func (this *UserGroupService) evaluateTraceSequenceCondition(sequenceCondition map[string]interface{}) ([]string, error) {
	formRaw, ok := sequenceCondition["form"]
	if !ok {
		return nil, errors.New("路径规则缺少分析表单")
	}
	contextRaw, ok := sequenceCondition["context"]
	if !ok {
		return nil, errors.New("路径规则缺少上下文")
	}

	formBytes, err := json.Marshal(formRaw)
	if err != nil {
		return nil, err
	}

	traceReq := request.TraceReqData{}
	if err := json.Unmarshal(formBytes, &traceReq); err != nil {
		return nil, err
	}
	traceReq.Appid = this.Appid

	contextBytes, err := json.Marshal(contextRaw)
	if err != nil {
		return nil, err
	}
	context := struct {
		Trace string `json:"trace"`
	}{}
	if err := json.Unmarshal(contextBytes, &context); err != nil {
		return nil, err
	}

	reqBytes, err := json.Marshal(traceReq)
	if err != nil {
		return nil, err
	}
	i, err := analysis.NewAnalysisByCommand(analysis.TraceComand, reqBytes)
	if err != nil {
		return nil, err
	}
	res, err := analysis.GetAnalysisRes(i)
	if err != nil {
		return nil, err
	}

	resultBytes, err := json.Marshal(res)
	if err != nil {
		return nil, err
	}
	result := traceSequenceEvalResult{}
	if err := json.Unmarshal(resultBytes, &result); err != nil {
		return nil, err
	}

	for _, row := range result.TableRes {
		if row.Trace == context.Trace {
			return row.UI, nil
		}
	}
	return []string{}, nil
}

func (this *UserGroupService) EvaluateRuleContent(rawRuleContent string) ([]string, error) {
	rawRuleContent = strings.TrimSpace(rawRuleContent)
	if rawRuleContent == "" {
		return nil, errors.New("分群规则不能为空")
	}

	ruleContent := UserGroupRuleContent{}
	if err := json.Unmarshal([]byte(rawRuleContent), &ruleContent); err != nil {
		return nil, err
	}

	if len(ruleContent.SequenceCondition) > 0 {
		return this.evaluateSequenceCondition(ruleContent.SequenceCondition)
	}

	behaviorSet, err := this.evaluateBehaviorCondition(ruleContent.BehaviorCondition)
	if err != nil {
		return nil, err
	}
	userAttrSet, err := this.evaluateUserAttrCondition(ruleContent.UserAttrCondition)
	if err != nil {
		return nil, err
	}

	hasBehavior := behaviorSet != nil
	hasUserAttr := userAttrSet != nil
	if !hasBehavior && !hasUserAttr {
		return nil, errors.New("分群规则不能为空")
	}

	var finalSet map[string]struct{}
	switch {
	case hasBehavior && hasUserAttr:
		finalSet = combineUserSetsByRelation(ruleContent.Relation, behaviorSet, userAttrSet)
	case hasBehavior:
		finalSet = behaviorSet
	default:
		finalSet = userAttrSet
	}

	return userSetToSlice(finalSet), nil
}
