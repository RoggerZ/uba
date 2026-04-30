package router

import (
	. "github.com/1340691923/xwl_bi/controller"
	"github.com/1340691923/xwl_bi/middleware"
	"github.com/1340691923/xwl_bi/platform-basic-libs/api_config"
	"github.com/gofiber/fiber/v2"
)

func runCost(app *fiber.App) {
	c := api_config.NewApiRouterConfig()
	const AbsolutePath = "/api/cost"
	appG := app.Group(AbsolutePath).Use(middleware.FilterAppid)
	{
		appG = appG.Use(middleware.OperaterLog)

		c.MountApi(api_config.MountApiBasePramas{Remark: "新增渠道成本", AbsolutePath: AbsolutePath, RelativePath: "add"}, appG.(*fiber.Group), ChannelCostController{}.Add)
		c.MountApi(api_config.MountApiBasePramas{Remark: "修改渠道成本", AbsolutePath: AbsolutePath, RelativePath: "update"}, appG.(*fiber.Group), ChannelCostController{}.Update)
		c.MountApi(api_config.MountApiBasePramas{Remark: "删除渠道成本", AbsolutePath: AbsolutePath, RelativePath: "delete"}, appG.(*fiber.Group), ChannelCostController{}.Delete)
		c.MountApi(api_config.MountApiBasePramas{Remark: "渠道成本列表", AbsolutePath: AbsolutePath, RelativePath: "list"}, appG.(*fiber.Group), ChannelCostController{}.List)
	}
}
