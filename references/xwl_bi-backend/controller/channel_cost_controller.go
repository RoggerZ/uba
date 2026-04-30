package controller

import (
	"github.com/1340691923/xwl_bi/platform-basic-libs/request"
	"github.com/1340691923/xwl_bi/platform-basic-libs/response"
	"github.com/1340691923/xwl_bi/platform-basic-libs/service/cost"
	"github.com/gofiber/fiber/v2"
)

type ChannelCostController struct {
	BaseController
}

func (c ChannelCostController) Add(ctx *fiber.Ctx) error {
	var req request.ChannelCostAddReq
	if err := ctx.BodyParser(&req); err != nil {
		return c.Error(ctx, err)
	}
	service := cost.ChannelCostService{Ctx: ctx}
	if err := service.Add(req); err != nil {
		return c.Error(ctx, err)
	}
	return c.Success(ctx, response.OperateSuccess, nil)
}

func (c ChannelCostController) Update(ctx *fiber.Ctx) error {
	var req request.ChannelCostUpdateReq
	if err := ctx.BodyParser(&req); err != nil {
		return c.Error(ctx, err)
	}
	service := cost.ChannelCostService{Ctx: ctx}
	if err := service.Update(req); err != nil {
		return c.Error(ctx, err)
	}
	return c.Success(ctx, response.OperateSuccess, nil)
}

func (c ChannelCostController) Delete(ctx *fiber.Ctx) error {
	var req request.ChannelCostDeleteReq
	if err := ctx.BodyParser(&req); err != nil {
		return c.Error(ctx, err)
	}
	service := cost.ChannelCostService{Ctx: ctx}
	if err := service.Delete(req); err != nil {
		return c.Error(ctx, err)
	}
	return c.Success(ctx, response.DeleteSuccess, nil)
}

func (c ChannelCostController) List(ctx *fiber.Ctx) error {
	var req request.ChannelCostListReq
	if err := ctx.BodyParser(&req); err != nil {
		return c.Error(ctx, err)
	}
	service := cost.ChannelCostService{Ctx: ctx}
	list, err := service.List(req)
	if err != nil {
		return c.Error(ctx, err)
	}
	return c.Success(ctx, response.SearchSuccess, list)
}
