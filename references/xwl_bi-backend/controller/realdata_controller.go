package controller

import (
	"errors"
	"github.com/1340691923/xwl_bi/engine/db"
	"github.com/1340691923/xwl_bi/engine/logs"
	"github.com/1340691923/xwl_bi/platform-basic-libs/jwt"
	"github.com/1340691923/xwl_bi/platform-basic-libs/request"
	"github.com/1340691923/xwl_bi/platform-basic-libs/response"
	"github.com/1340691923/xwl_bi/platform-basic-libs/service/debug_data"
	"github.com/1340691923/xwl_bi/platform-basic-libs/service/realdata"
	"github.com/1340691923/xwl_bi/platform-basic-libs/util"
	"github.com/gofiber/fiber/v2"
	"strconv"
	"strings"
	"time"
)

type RealDataController struct {
	BaseController
}

// List 查看实时数据列表
func (c RealDataController) List(ctx *fiber.Ctx) error {

	type ReqData struct {
		Appid    int    `json:"appid"`
		SearchKw string `json:"searchKw"`
		Date     string `json:"date"`
	}

	var reqData ReqData

	if err := ctx.BodyParser(&reqData); err != nil {
		return c.Error(ctx, err)
	}

	appid := strconv.Itoa(reqData.Appid)

	type Res struct {
		CreateTime   string    `json:"create_time" db:"-"`
		CreateTimeDb time.Time `json:"-" db:"create_time"`
		EventName    string    `json:"event_name" db:"event_name"`
		ReportData   string    `json:"report_data" db:"report_data"`
	}

	filterSql := ""

	date := strings.Split(reqData.Date, ",")

	args := []interface{}{appid}

	if len(date) == 2 {
		filterSql = filterSql + ` and event_time >= toDateTime(?) and event_time <=toDateTime(?) `
		args = append(args, date[0], date[1])
	}
	if strings.TrimSpace(reqData.SearchKw) != "" {
		filterSql = filterSql + ` and event_name like '%` + reqData.SearchKw + `%' `
	}
	sql := `select report_data,event_name,event_time as create_time from xwl_real_time_warehousing prewhere   table_id = ?    ` + filterSql + ` order by event_time desc limit 1000;`
	logs.Logger.Sugar().Infof("sql", sql, args)
	var res []Res
	err := db.ClickHouseSqlx.Select(&res, sql, args...)
	if err != nil {
		return c.Error(ctx, err)
	}
	for index := range res {
		res[index].CreateTime = res[index].CreateTimeDb.Format(util.TimeFormat)
	}

	return c.Success(ctx, response.SearchSuccess, map[string]interface{}{"list": res})
}

// FailDataList 错误数据列表
func (c RealDataController) FailDataList(ctx *fiber.Ctx) error {

	type ReqData struct {
		Appid   int `json:"appid"`
		Minutes int `json:"minutes"`
	}

	var reqData ReqData

	if err := ctx.BodyParser(&reqData); err != nil {
		return c.Error(ctx, err)
	}

	if reqData.Minutes == 0 {
		reqData.Minutes = 10
	}

	realDataService := realdata.RealDataService{}

	res, err := realDataService.FailDataList(reqData.Minutes, reqData.Appid)
	if err != nil {
		return c.Error(ctx, err)
	}

	return c.Success(ctx, response.SearchSuccess, map[string]interface{}{"list": res})
}

// FailDataDesc 抽样示例
func (c RealDataController) FailDataDesc(ctx *fiber.Ctx) error {

	type ReqData struct {
		StartTime     string `json:"start_time"`
		EndTime       string `json:"end_time"`
		Appid         int    `json:"appid"`
		ErrorReason   string `json:"error_reason"`
		ErrorHandling string `json:"error_handling"`
		ReportType    string `json:"report_type"`
	}
	var reqData ReqData

	if err := ctx.BodyParser(&reqData); err != nil {
		return c.Error(ctx, err)
	}

	startTime := reqData.StartTime
	endTime := reqData.EndTime
	appid := strconv.Itoa(reqData.Appid)
	errorReason := reqData.ErrorReason
	errorHandling := reqData.ErrorHandling
	reportType := reqData.ReportType

	realDataService := realdata.RealDataService{}

	res, err := realDataService.FailDataDesc(appid, startTime, endTime, errorReason, errorHandling, reportType)
	if err != nil {
		return c.Error(ctx, err)
	}

	return c.Success(ctx, response.SearchSuccess, map[string]interface{}{"data": res})
}

// ReportCount 查看所有上报数据情况
func (c RealDataController) ReportCount(ctx *fiber.Ctx) error {

	var err error

	var reqData request.ReportCountReq
	if err := ctx.BodyParser(&reqData); err != nil {
		return c.Error(ctx, err)
	}

	startTime := reqData.StartTime
	endTime := reqData.EndTime
	appid := strconv.Itoa(reqData.Appid)

	realDataService := realdata.RealDataService{}

	res, err := realDataService.ReportCount(appid, startTime, endTime)
	if err != nil {
		return c.Error(ctx, err)
	}

	return c.Success(ctx, response.SearchSuccess, map[string]interface{}{"list": res})
}

// EventFailDesc 事件失败详情
func (c RealDataController) EventFailDesc(ctx *fiber.Ctx) error {

	var reqData request.EventFailDescReq
	if err := ctx.BodyParser(&reqData); err != nil {
		return c.Error(ctx, err)
	}

	startTime := reqData.StartTime
	endTime := reqData.EndTime
	appid := strconv.Itoa(reqData.Appid)
	dataName := reqData.DataName

	realDataService := realdata.RealDataService{}

	res, err := realDataService.EventFailDesc(appid, startTime, endTime, dataName)
	if err != nil {
		return c.Error(ctx, err)
	}

	return c.Success(ctx, response.SearchSuccess, map[string]interface{}{"list": res})
}

// AddDebugDeviceID 添加DEBUG设备ID
func (c RealDataController) AddDebugDeviceID(ctx *fiber.Ctx) error {

	var reqData request.AddDebugDeviceIDReq
	if err := ctx.BodyParser(&reqData); err != nil {
		return c.Error(ctx, err)
	}

	appid := strconv.Itoa(reqData.Appid)
	remark := reqData.Remark
	deviceID := reqData.DeviceID

	if deviceID == "" {
		return c.Error(ctx, errors.New("设备ID不能为空"))
	}

	cc, _ := jwt.ParseToken(c.GetToken(ctx))

	debugData := debug_data.DebugData{}

	err := debugData.AddDebugDeviceID(appid, deviceID, remark, cc.UserID)

	if err != nil {
		return c.Error(ctx, err)
	}

	return c.Success(ctx, response.OperateSuccess, nil)
}

// DelDebugDeviceID 删除测试设备
func (c RealDataController) DelDebugDeviceID(ctx *fiber.Ctx) error {

	var reqData request.DelDebugDeviceIDReq
	if err := ctx.BodyParser(&reqData); err != nil {
		return c.Error(ctx, err)
	}

	appid := strconv.Itoa(reqData.Appid)

	deviceID := reqData.DeviceID

	cc, _ := jwt.ParseToken(c.GetToken(ctx))

	debugData := debug_data.DebugData{}

	err := debugData.DelDebugDeviceID(appid, deviceID, cc.UserID)

	if err != nil {
		return c.Error(ctx, err)
	}

	return c.Success(ctx, response.OperateSuccess, nil)
}

// DebugDeviceIDList 查看测试设备列表
func (c RealDataController) DebugDeviceIDList(ctx *fiber.Ctx) error {

	var reqData request.DebugDeviceIDListReq
	if err := ctx.BodyParser(&reqData); err != nil {
		return c.Error(ctx, err)
	}

	appid := reqData.Appid
	cc, _ := jwt.ParseToken(c.GetToken(ctx))

	debugData := debug_data.DebugData{}

	res, err := debugData.DebugDeviceIDList(appid, cc.UserID)

	if err != nil {
		return c.Error(ctx, err)
	}
	return c.Success(ctx, response.SearchSuccess, map[string]interface{}{"list": res})
}
