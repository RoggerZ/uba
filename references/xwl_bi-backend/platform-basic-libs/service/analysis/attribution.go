package analysis

import (
	"database/sql"
	"fmt"
	"sort"
	"strconv"
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

const (
	AttributionModelFirstTouch = "first_touch"
	AttributionModelLastTouch  = "last_touch"
	AttributionModelLinear     = "linear"
	AttributionModelUShape     = "u_shape"
	AttributionModelWShape     = "w_shape"
)

type Attribution struct {
	req request.AttributionReqData
}

type attributionEventRow struct {
	DistinctID       string         `db:"xwl_distinct_id"`
	EventTime        time.Time      `db:"xwl_part_date"`
	SourceKind       string         `db:"source_kind"`
	SourceIndex      int            `db:"source_index"`
	EventName        string         `db:"xwl_part_event"`
	LinkValue        sql.NullString `db:"link_value"`
	TouchGroupValue  sql.NullString `db:"touch_group_value"`
	ConversionCount  float64        `db:"conversion_count"`
	ConversionValue  float64        `db:"conversion_value"`
	SourceOrderValue int            `db:"source_order"`
}

type attributionTouchOccurrence struct {
	ID          string
	UserID      string
	SourceIndex int
	EventTime   time.Time
	GroupValue  string
	DateGroup   string
}

type attributionForwardOccurrence struct {
	EventTime time.Time
	LinkValue string
}

type attributionConversionOccurrence struct {
	UserID          string
	EventTime       time.Time
	LinkValue       string
	ConversionCount float64
	ConversionValue float64
	DateGroup       string
}

type attributionRowMetric struct {
	SortIndex       int
	DateGroup       string
	EventName       string
	GroupValue      string
	TouchCount      int
	TouchUsers      map[string]struct{}
	ValidTouchIDs   map[string]struct{}
	ValidTouchUsers map[string]struct{}
	ConversionUsers map[string]struct{}
	AttributedCount float64
	AttributedValue float64
}

type AttributionRowRes struct {
	SortIndex               int     `json:"-"`
	DateGroup               string  `json:"date_group"`
	EventName               string  `json:"event_name"`
	GroupValue              string  `json:"group_value"`
	TouchCount              int     `json:"touch_count"`
	TouchUserCount          int     `json:"touch_user_count"`
	ValidTouchCount         int     `json:"valid_touch_count"`
	ValidTouchUserCount     int     `json:"valid_touch_user_count"`
	ValidTouchRate          float64 `json:"valid_touch_rate"`
	AttributedCount         float64 `json:"attributed_count"`
	AttributedValue         float64 `json:"attributed_value"`
	AttributedUserCount     int     `json:"attributed_user_count"`
	ContributionRateByCount float64 `json:"contribution_rate_by_count"`
	ContributionRateByValue float64 `json:"contribution_rate_by_value"`
}

type AttributionSummaryRes struct {
	Model                       string  `json:"model"`
	GroupBy                     string  `json:"group_by"`
	WindowTime                  int     `json:"window_time"`
	WindowTimeFormat            string  `json:"window_time_format"`
	ValueField                  string  `json:"value_field"`
	HasValueField               bool    `json:"has_value_field"`
	TotalTouchCount             int     `json:"total_touch_count"`
	TotalTouchUserCount         int     `json:"total_touch_user_count"`
	ValidTouchCount             int     `json:"valid_touch_count"`
	ValidTouchUserCount         int     `json:"valid_touch_user_count"`
	TotalConversionCount        float64 `json:"total_conversion_count"`
	TotalConversionValue        float64 `json:"total_conversion_value"`
	AttributedConversionCount   float64 `json:"attributed_conversion_count"`
	AttributedConversionValue   float64 `json:"attributed_conversion_value"`
	UnattributedConversionCount float64 `json:"unattributed_conversion_count"`
	UnattributedConversionValue float64 `json:"unattributed_conversion_value"`
	AttributionCoverageRate     float64 `json:"attribution_coverage_rate"`
}

func NewAttribution(reqData []byte) (Ianalysis, error) {
	obj := &Attribution{}
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	if err := json.Unmarshal(reqData, &obj.req); err != nil {
		return nil, err
	}

	if len(obj.req.Date) < 2 {
		return nil, my_error.NewBusiness(ERROR_TABLE, TimeError)
	}
	if obj.req.ConversionEvent.EventName == "" {
		return nil, my_error.NewBusiness(ERROR_TABLE, EventNameEmptyError)
	}
	if len(obj.req.TouchArr) == 0 {
		return nil, my_error.NewBusiness(ERROR_TABLE, ZhiBiaoNumError)
	}
	if len(obj.req.GroupBy) > 1 {
		return nil, my_error.NewBusiness(ERROR_TABLE, GroupNumError)
	}
	if len(obj.req.GroupBy) == 1 && strings.TrimSpace(obj.req.GroupBy[0]) == "" {
		obj.req.GroupBy = []string{}
	}

	windowUnit, err := getAttributionWindowUnit(obj.req.WindowTimeFormat)
	if err != nil {
		return nil, err
	}
	obj.req.WindowTime = obj.req.WindowTime * windowUnit
	if obj.req.WindowTime <= 0 {
		return nil, my_error.NewBusiness(ERROR_TABLE, TimeError)
	}
	if obj.req.AttributionModel == "" {
		obj.req.AttributionModel = AttributionModelLastTouch
	}
	if obj.req.ConversionTimeFormat == "" {
		obj.req.ConversionTimeFormat = ByDay
	}

	return obj, nil
}

func getAttributionWindowUnit(windowTimeFormat string) (int, error) {
	switch windowTimeFormat {
	case "天":
		return 60 * 60 * 24, nil
	case "小时":
		return 60 * 60, nil
	case "分钟":
		return 60, nil
	case "秒":
		return 1, nil
	default:
		return 0, my_error.NewBusiness(ERROR_TABLE, TimeError)
	}
}

func (this *Attribution) GetExecSql() (SQL string, allArgs []interface{}, err error) {
	startTime := util.Str2Time(this.req.Date[0], util.TimeFormatDay2)
	endTime := util.Str2Time(this.req.Date[1], util.TimeFormatDay2).Add(23*time.Hour + 59*time.Minute + 59*time.Second)
	windowStart := startTime.Add(-time.Duration(this.req.WindowTime) * time.Second)

	globalFilterSql, globalArgs, err := buildAnalysisFilterSql(this.req.WhereFilter)
	if err != nil {
		return "", nil, err
	}
	userFilterSql, userFilterArgs, err := getUserfilterSqlArgs(this.req.WhereFilterByUser, this.req.Appid)
	if err != nil {
		return "", nil, err
	}
	userGroupSql, userGroupArgs, err := utils.GetUserGroupSqlAndArgs(this.req.UserGroup, this.req.Appid)
	if err != nil {
		return "", nil, err
	}

	var selectSqlList []string
	var argsList []interface{}

	conversionSql, conversionArgs, err := this.buildSelectSql(
		"conversion",
		0,
		this.req.ConversionEvent.EventName,
		this.req.ConversionEvent.Relation,
		startTime,
		endTime,
		globalFilterSql,
		globalArgs,
		userFilterSql,
		userFilterArgs,
		userGroupSql,
		userGroupArgs,
	)
	if err != nil {
		return "", nil, err
	}
	selectSqlList = append(selectSqlList, conversionSql)
	argsList = append(argsList, conversionArgs...)

	if this.req.ForwardEvent.EventName != "" {
		forwardSql, forwardArgs, forwardErr := this.buildSelectSql(
			"forward",
			0,
			this.req.ForwardEvent.EventName,
			this.req.ForwardEvent.Relation,
			windowStart,
			endTime,
			globalFilterSql,
			globalArgs,
			userFilterSql,
			userFilterArgs,
			userGroupSql,
			userGroupArgs,
		)
		if forwardErr != nil {
			return "", nil, forwardErr
		}
		selectSqlList = append(selectSqlList, forwardSql)
		argsList = append(argsList, forwardArgs...)
	}

	for index, touch := range this.req.TouchArr {
		touchSql, touchArgs, touchErr := this.buildSelectSql(
			"touch",
			index,
			touch.EventName,
			touch.Relation,
			windowStart,
			endTime,
			globalFilterSql,
			globalArgs,
			userFilterSql,
			userFilterArgs,
			userGroupSql,
			userGroupArgs,
		)
		if touchErr != nil {
			return "", nil, touchErr
		}
		selectSqlList = append(selectSqlList, touchSql)
		argsList = append(argsList, touchArgs...)
	}

	SQL = "SELECT * FROM (" + strings.Join(selectSqlList, " UNION ALL ") + ") ORDER BY xwl_distinct_id ASC, xwl_part_date ASC, source_order ASC, source_index ASC"
	return SQL, argsList, nil
}

func buildAnalysisFilterSql(filter request.AnalysisFilter) (string, []interface{}, error) {
	if len(filter.Filts) == 0 {
		return "1 = 1", nil, nil
	}
	sqlText, args, _, err := utils.GetWhereSql(filter)
	if err != nil {
		return "", nil, err
	}
	return sqlText, args, nil
}

func (this *Attribution) buildSelectSql(
	sourceKind string,
	sourceIndex int,
	eventName string,
	relation request.AnalysisFilter,
	startTime time.Time,
	endTime time.Time,
	globalFilterSql string,
	globalArgs []interface{},
	userFilterSql string,
	userFilterArgs []interface{},
	userGroupSql string,
	userGroupArgs []interface{},
) (string, []interface{}, error) {
	eventFilterSql, eventArgs, err := buildAnalysisFilterSql(relation)
	if err != nil {
		return "", nil, err
	}

	linkValueExpr := "'' AS link_value"
	if this.req.ForwardEvent.LinkField != "" && (sourceKind == "conversion" || sourceKind == "forward") {
		linkValueExpr = fmt.Sprintf("toString(ifNull(%s, '')) AS link_value", this.req.ForwardEvent.LinkField)
	}

	touchGroupExpr := "'' AS touch_group_value"
	if sourceKind == "touch" && len(this.req.GroupBy) > 0 {
		touchGroupExpr = fmt.Sprintf("toString(ifNull(%s, '')) AS touch_group_value", this.req.GroupBy[0])
	}

	conversionCountExpr := "toFloat64(0) AS conversion_count"
	conversionValueExpr := "toFloat64(0) AS conversion_value"
	if sourceKind == "conversion" {
		conversionCountExpr = "toFloat64(1) AS conversion_count"
		if this.req.ConversionEvent.ValueField != "" {
			conversionValueExpr = fmt.Sprintf("toFloat64(ifNull(%s, 0)) AS conversion_value", this.req.ConversionEvent.ValueField)
		} else {
			conversionValueExpr = "toFloat64(1) AS conversion_value"
		}
	}

	sourceOrder := 1
	switch sourceKind {
	case "touch":
		sourceOrder = 1
	case "forward":
		sourceOrder = 2
	case "conversion":
		sourceOrder = 3
	}

	sqlText := fmt.Sprintf(`
		SELECT
			xwl_distinct_id,
			xwl_part_date,
			'%s' AS source_kind,
			%d AS source_index,
			%d AS source_order,
			xwl_part_event,
			%s,
			%s,
			%s,
			%s
		FROM xwl_event%d
		WHERE xwl_part_date >= toDateTime('%s')
			AND xwl_part_date <= toDateTime('%s')
			AND (%s)
			AND xwl_part_event = '%s'
			AND (%s)%s%s
	`,
		sourceKind,
		sourceIndex,
		sourceOrder,
		linkValueExpr,
		touchGroupExpr,
		conversionCountExpr,
		conversionValueExpr,
		this.req.Appid,
		startTime.Format(util.TimeFormat),
		endTime.Format(util.TimeFormat),
		globalFilterSql,
		eventName,
		eventFilterSql,
		userGroupSql,
		userFilterSql,
	)

	args := append([]interface{}{}, globalArgs...)
	args = append(args, eventArgs...)
	args = append(args, userGroupArgs...)
	args = append(args, userFilterArgs...)

	return sqlText, args, nil
}

func (this *Attribution) GetList() (interface{}, error) {
	sqlText, args, err := this.GetExecSql()
	if err != nil {
		return nil, err
	}

	logs.Logger.Sugar().Infof("attribution sql", sqlText, args)

	var rows []attributionEventRow
	if err := db.ClickHouseSqlx.Select(&rows, sqlText, args...); err != nil {
		return nil, err
	}

	return this.buildResult(rows), nil
}

func (this *Attribution) buildResult(rows []attributionEventRow) map[string]interface{} {
	metricsMap := map[string]*attributionRowMetric{}
	totalTouchUsers := map[string]struct{}{}
	totalValidTouchIDs := map[string]struct{}{}
	totalValidTouchUsers := map[string]struct{}{}
	totalConversionUsers := map[string]struct{}{}

	totalTouchCount := 0
	totalConversionCount := 0.0
	totalConversionValue := 0.0
	attributedConversionCount := 0.0
	attributedConversionValue := 0.0

	currentUser := ""
	touches := []attributionTouchOccurrence{}
	forwards := []attributionForwardOccurrence{}
	conversions := []attributionConversionOccurrence{}
	touchCounter := 0

	flushUser := func() {
		if currentUser == "" {
			return
		}

		userAttributedCount, userAttributedValue := this.processUserEvents(
			touches,
			forwards,
			conversions,
			metricsMap,
			totalValidTouchIDs,
			totalValidTouchUsers,
			totalConversionUsers,
		)
		attributedConversionCount += userAttributedCount
		attributedConversionValue += userAttributedValue
	}

	for _, row := range rows {
		if currentUser != "" && currentUser != row.DistinctID {
			flushUser()
			touches = []attributionTouchOccurrence{}
			forwards = []attributionForwardOccurrence{}
			conversions = []attributionConversionOccurrence{}
			touchCounter = 0
		}
		currentUser = row.DistinctID

		switch row.SourceKind {
		case "touch":
			touchCounter++
			metric := this.ensureMetric(
				metricsMap,
				row.SourceIndex,
				this.formatDateGroup(row.EventTime),
				normalizeDisplayName(this.req.TouchArr[row.SourceIndex].EventNameDisplay, this.req.TouchArr[row.SourceIndex].EventName),
				normalizeGroupValue(row.TouchGroupValue),
			)
			metric.TouchCount++
			metric.TouchUsers[row.DistinctID] = struct{}{}
			totalTouchCount++
			totalTouchUsers[row.DistinctID] = struct{}{}
			touches = append(touches, attributionTouchOccurrence{
				ID:          fmt.Sprintf("%s_%d_%d", row.DistinctID, row.SourceIndex, touchCounter),
				UserID:      row.DistinctID,
				SourceIndex: row.SourceIndex,
				EventTime:   row.EventTime,
				GroupValue:  normalizeGroupValue(row.TouchGroupValue),
				DateGroup:   this.formatDateGroup(row.EventTime),
			})
		case "forward":
			forwards = append(forwards, attributionForwardOccurrence{
				EventTime: row.EventTime,
				LinkValue: normalizeNullString(row.LinkValue),
			})
		case "conversion":
			totalConversionCount += row.ConversionCount
			totalConversionValue += row.ConversionValue
			conversions = append(conversions, attributionConversionOccurrence{
				UserID:          row.DistinctID,
				EventTime:       row.EventTime,
				LinkValue:       normalizeNullString(row.LinkValue),
				ConversionCount: row.ConversionCount,
				ConversionValue: row.ConversionValue,
				DateGroup:       this.formatDateGroup(row.EventTime),
			})
		}
	}

	flushUser()

	rowsRes := make([]AttributionRowRes, 0, len(metricsMap))
	for _, metric := range metricsMap {
		validTouchCount := len(metric.ValidTouchIDs)
		touchUserCount := len(metric.TouchUsers)
		validTouchUserCount := len(metric.ValidTouchUsers)
		attributedUserCount := len(metric.ConversionUsers)
		validTouchRate := 0.0
		if metric.TouchCount > 0 {
			validTouchRate = float64(validTouchCount) / float64(metric.TouchCount)
		}

		contributionRateByCount := 0.0
		if totalConversionCount > 0 {
			contributionRateByCount = metric.AttributedCount / totalConversionCount
		}
		contributionRateByValue := 0.0
		if totalConversionValue > 0 {
			contributionRateByValue = metric.AttributedValue / totalConversionValue
		}

		rowsRes = append(rowsRes, AttributionRowRes{
			SortIndex:               metric.SortIndex,
			DateGroup:               metric.DateGroup,
			EventName:               metric.EventName,
			GroupValue:              metric.GroupValue,
			TouchCount:              metric.TouchCount,
			TouchUserCount:          touchUserCount,
			ValidTouchCount:         validTouchCount,
			ValidTouchUserCount:     validTouchUserCount,
			ValidTouchRate:          validTouchRate,
			AttributedCount:         metric.AttributedCount,
			AttributedValue:         metric.AttributedValue,
			AttributedUserCount:     attributedUserCount,
			ContributionRateByCount: contributionRateByCount,
			ContributionRateByValue: contributionRateByValue,
		})
	}

	sort.Slice(rowsRes, func(i, j int) bool {
		if rowsRes[i].DateGroup != rowsRes[j].DateGroup {
			return rowsRes[i].DateGroup > rowsRes[j].DateGroup
		}
		if rowsRes[i].SortIndex != rowsRes[j].SortIndex {
			return rowsRes[i].SortIndex < rowsRes[j].SortIndex
		}
		if this.req.ConversionEvent.ValueField != "" && rowsRes[i].AttributedValue != rowsRes[j].AttributedValue {
			return rowsRes[i].AttributedValue > rowsRes[j].AttributedValue
		}
		if rowsRes[i].AttributedCount != rowsRes[j].AttributedCount {
			return rowsRes[i].AttributedCount > rowsRes[j].AttributedCount
		}
		if rowsRes[i].TouchCount != rowsRes[j].TouchCount {
			return rowsRes[i].TouchCount > rowsRes[j].TouchCount
		}
		if rowsRes[i].EventName != rowsRes[j].EventName {
			return rowsRes[i].EventName < rowsRes[j].EventName
		}
		return rowsRes[i].GroupValue < rowsRes[j].GroupValue
	})

	unattributedCount := totalConversionCount - attributedConversionCount
	if unattributedCount < 0 {
		unattributedCount = 0
	}
	unattributedValue := totalConversionValue - attributedConversionValue
	if unattributedValue < 0 {
		unattributedValue = 0
	}

	coverageRate := 0.0
	if totalConversionCount > 0 {
		coverageRate = attributedConversionCount / totalConversionCount
	}

	groupByField := ""
	if len(this.req.GroupBy) > 0 {
		groupByField = this.req.GroupBy[0]
	}

	return map[string]interface{}{
		"rows": rowsRes,
		"summary": AttributionSummaryRes{
			Model:                       this.req.AttributionModel,
			GroupBy:                     groupByField,
			WindowTime:                  this.req.WindowTime,
			WindowTimeFormat:            this.req.WindowTimeFormat,
			ValueField:                  this.req.ConversionEvent.ValueField,
			HasValueField:               this.req.ConversionEvent.ValueField != "",
			TotalTouchCount:             totalTouchCount,
			TotalTouchUserCount:         len(totalTouchUsers),
			ValidTouchCount:             len(totalValidTouchIDs),
			ValidTouchUserCount:         len(totalValidTouchUsers),
			TotalConversionCount:        totalConversionCount,
			TotalConversionValue:        totalConversionValue,
			AttributedConversionCount:   attributedConversionCount,
			AttributedConversionValue:   attributedConversionValue,
			UnattributedConversionCount: unattributedCount,
			UnattributedConversionValue: unattributedValue,
			AttributionCoverageRate:     coverageRate,
		},
	}
}

func (this *Attribution) processUserEvents(
	touches []attributionTouchOccurrence,
	forwards []attributionForwardOccurrence,
	conversions []attributionConversionOccurrence,
	metricsMap map[string]*attributionRowMetric,
	totalValidTouchIDs map[string]struct{},
	totalValidTouchUsers map[string]struct{},
	totalConversionUsers map[string]struct{},
) (float64, float64) {
	if len(touches) == 0 || len(conversions) == 0 {
		return 0, 0
	}

	touchTimes := make([]time.Time, 0, len(touches))
	for _, touch := range touches {
		touchTimes = append(touchTimes, touch.EventTime)
	}

	userAttributedCount := 0.0
	userAttributedValue := 0.0
	hasForwardEvent := this.req.ForwardEvent.EventName != ""
	windowDuration := time.Duration(this.req.WindowTime) * time.Second

	for _, conversion := range conversions {
		windowStart := conversion.EventTime.Add(-windowDuration)
		effectiveStart := windowStart
		if hasForwardEvent {
			matchedForward, ok := findMatchedForward(forwards, conversion.EventTime, windowStart, conversion.LinkValue, this.req.ForwardEvent.LinkField != "")
			if !ok {
				if this.req.IncludeDirectConversion {
					this.addDirectConversionMetric(metricsMap, conversion, totalValidTouchIDs, totalValidTouchUsers, totalConversionUsers)
					userAttributedCount += conversion.ConversionCount
					userAttributedValue += conversion.ConversionValue
				}
				continue
			}
			effectiveStart = matchedForward.EventTime
		}

		windowTouches := getTouchesBetween(touches, touchTimes, effectiveStart, conversion.EventTime)
		if len(windowTouches) == 0 {
			if this.req.IncludeDirectConversion {
				this.addDirectConversionMetric(metricsMap, conversion, totalValidTouchIDs, totalValidTouchUsers, totalConversionUsers)
				userAttributedCount += conversion.ConversionCount
				userAttributedValue += conversion.ConversionValue
			}
			continue
		}

		userAttributedCount += conversion.ConversionCount
		userAttributedValue += conversion.ConversionValue
		totalConversionUsers[conversion.UserID] = struct{}{}

		weights := getAttributionWeights(this.req.AttributionModel, len(windowTouches))
		for index, touch := range windowTouches {
			weight := weights[index]
			if weight <= 0 {
				continue
			}

			displayName := normalizeDisplayName(this.req.TouchArr[touch.SourceIndex].EventNameDisplay, this.req.TouchArr[touch.SourceIndex].EventName)
			metric := this.ensureMetric(metricsMap, touch.SourceIndex, conversion.DateGroup, displayName, touch.GroupValue)
			metric.ValidTouchIDs[touch.ID] = struct{}{}
			metric.ValidTouchUsers[touch.UserID] = struct{}{}
			metric.ConversionUsers[conversion.UserID] = struct{}{}
			metric.AttributedCount += weight * conversion.ConversionCount
			metric.AttributedValue += weight * conversion.ConversionValue
			totalValidTouchIDs[touch.ID] = struct{}{}
			totalValidTouchUsers[touch.UserID] = struct{}{}
		}
	}

	return userAttributedCount, userAttributedValue
}

func (this *Attribution) addDirectConversionMetric(
	metricsMap map[string]*attributionRowMetric,
	conversion attributionConversionOccurrence,
	totalValidTouchIDs map[string]struct{},
	totalValidTouchUsers map[string]struct{},
	totalConversionUsers map[string]struct{},
) {
	directTouchID := fmt.Sprintf("direct_%s_%d", conversion.UserID, conversion.EventTime.UnixNano())
	metric := this.ensureMetric(metricsMap, len(this.req.TouchArr)+1, conversion.DateGroup, "直接转化", "直接转化")
	metric.TouchCount += int(conversion.ConversionCount)
	metric.TouchUsers[conversion.UserID] = struct{}{}
	metric.ValidTouchIDs[directTouchID] = struct{}{}
	metric.ValidTouchUsers[conversion.UserID] = struct{}{}
	metric.ConversionUsers[conversion.UserID] = struct{}{}
	metric.AttributedCount += conversion.ConversionCount
	metric.AttributedValue += conversion.ConversionValue
	totalValidTouchIDs[directTouchID] = struct{}{}
	totalValidTouchUsers[conversion.UserID] = struct{}{}
	totalConversionUsers[conversion.UserID] = struct{}{}
}

func getTouchesBetween(touches []attributionTouchOccurrence, touchTimes []time.Time, startTime time.Time, endTime time.Time) []attributionTouchOccurrence {
	if len(touches) == 0 {
		return nil
	}

	startIndex := sort.Search(len(touchTimes), func(i int) bool {
		return !touchTimes[i].Before(startTime)
	})
	endIndex := sort.Search(len(touchTimes), func(i int) bool {
		return touchTimes[i].After(endTime)
	})

	if startIndex >= endIndex {
		return nil
	}

	result := make([]attributionTouchOccurrence, 0, endIndex-startIndex)
	result = append(result, touches[startIndex:endIndex]...)
	return result
}

func findMatchedForward(
	forwards []attributionForwardOccurrence,
	conversionTime time.Time,
	windowStart time.Time,
	conversionLinkValue string,
	requireLink bool,
) (attributionForwardOccurrence, bool) {
	for index := len(forwards) - 1; index >= 0; index-- {
		current := forwards[index]
		if current.EventTime.After(conversionTime) {
			continue
		}
		if current.EventTime.Before(windowStart) {
			break
		}
		if requireLink && current.LinkValue != conversionLinkValue {
			continue
		}
		return current, true
	}
	return attributionForwardOccurrence{}, false
}

func getAttributionWeights(model string, count int) []float64 {
	if count <= 0 {
		return nil
	}

	weights := make([]float64, count)
	switch model {
	case AttributionModelFirstTouch:
		weights[0] = 1
	case AttributionModelLinear:
		value := 1 / float64(count)
		for index := range weights {
			weights[index] = value
		}
	case AttributionModelUShape:
		if count == 1 {
			weights[0] = 1
		} else if count == 2 {
			weights[0] = 0.5
			weights[1] = 0.5
		} else {
			weights[0] = 0.4
			weights[count-1] = 0.4
			middleWeight := 0.2 / float64(count-2)
			for index := 1; index < count-1; index++ {
				weights[index] = middleWeight
			}
		}
	case AttributionModelWShape:
		if count == 1 {
			weights[0] = 1
		} else if count == 2 {
			weights[0] = 0.5
			weights[1] = 0.5
		} else if count == 3 {
			weights[0] = 0.3
			weights[1] = 0.4
			weights[2] = 0.3
		} else {
			middleIndex := count / 2
			weights[0] = 0.3
			weights[middleIndex] = 0.3
			weights[count-1] = 0.3
			restWeight := 0.1 / float64(count-3)
			for index := 1; index < count-1; index++ {
				if index == middleIndex {
					continue
				}
				weights[index] = restWeight
			}
		}
	case AttributionModelLastTouch:
		fallthrough
	default:
		weights[count-1] = 1
	}

	return weights
}

func (this *Attribution) ensureMetric(metricsMap map[string]*attributionRowMetric, sourceIndex int, dateGroup string, eventName string, groupValue string) *attributionRowMetric {
	key := this.getMetricKey(sourceIndex, dateGroup, groupValue)
	if metric, found := metricsMap[key]; found {
		return metric
	}

	metricsMap[key] = &attributionRowMetric{
		SortIndex:       sourceIndex,
		DateGroup:       dateGroup,
		EventName:       eventName,
		GroupValue:      groupValue,
		TouchUsers:      map[string]struct{}{},
		ValidTouchIDs:   map[string]struct{}{},
		ValidTouchUsers: map[string]struct{}{},
		ConversionUsers: map[string]struct{}{},
	}
	return metricsMap[key]
}

func (this *Attribution) getMetricKey(sourceIndex int, dateGroup string, groupValue string) string {
	return strconv.Itoa(sourceIndex) + "|" + dateGroup + "|" + groupValue
}

func (this *Attribution) formatDateGroup(t time.Time) string {
	switch this.req.ConversionTimeFormat {
	case ByHour:
		return t.Format("2006-01-02 15:00")
	case Monthly:
		return t.Format("2006-01")
	case ByWeek:
		weekStart := t.AddDate(0, 0, -int((int(t.Weekday())+6)%7))
		weekEnd := weekStart.AddDate(0, 0, 6)
		return fmt.Sprintf("%s ~ %s", weekStart.Format(util.TimeFormatDay2), weekEnd.Format(util.TimeFormatDay2))
	case ByDay:
		fallthrough
	default:
		return t.Format(util.TimeFormatDay2)
	}
}

func normalizeNullString(value sql.NullString) string {
	if !value.Valid {
		return ""
	}
	return value.String
}

func normalizeGroupValue(value sql.NullString) string {
	groupValue := normalizeNullString(value)
	if strings.TrimSpace(groupValue) == "" {
		return "未设置"
	}
	return groupValue
}

func normalizeDisplayName(displayName string, eventName string) string {
	if strings.TrimSpace(displayName) != "" {
		return displayName
	}
	return eventName
}
