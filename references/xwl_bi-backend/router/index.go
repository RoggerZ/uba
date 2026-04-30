// 路由层
package router

import (
	. "github.com/1340691923/xwl_bi/controller"
	. "github.com/1340691923/xwl_bi/middleware"
	"github.com/1340691923/xwl_bi/views"
	. "github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	jsoniter "github.com/json-iterator/go"
)

const staticAssetMaxAgeSeconds = 60 * 60 * 24 * 30

func Init() *App {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	app := New(Config{
		AppName:     "铸龙-BI",
		JSONDecoder: json.Unmarshal,
		JSONEncoder: json.Marshal,
	})

	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestCompression,
	}))

	staticFileSystem := views.GetFileSystem()
	// static 和 vendor 下都是构建哈希文件或显式版本目录，可以交给浏览器/CDN 缓存。
	// 根路径继续不设置长缓存，避免 index.html 在发版后继续引用旧的 chunk 文件。
	app.Use("/static", filesystem.New(filesystem.Config{
		Root:       staticFileSystem,
		PathPrefix: "/static",
		MaxAge:     staticAssetMaxAgeSeconds,
	}))
	app.Use("/vendor", filesystem.New(filesystem.Config{
		Root:       staticFileSystem,
		PathPrefix: "/vendor",
		MaxAge:     staticAssetMaxAgeSeconds,
	}))
	app.Use("/", filesystem.New(filesystem.Config{
		Root: staticFileSystem,
	}))

	app.Use(
		cors.New(cors.Config{
			AllowOrigins: "*",
			AllowHeaders: "Origin, Content-Type, Accept, Authorization, X-Token",
			AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		}),
		pprof.New(),
	)

	app.Post("/api/gm_user/login", ManagerUserController{}.Login)
	routerWebsocket(app)
	app.Use(
		Timer,
		JwtMiddleware,
		Rbac,
	)

	return runRouterGroupFn(
		app,
		runOperaterLog,
		runGmUser,
		runRealData,
		runMetaData,
		runAnalysis,
		runPannel,
		runApp,
		runUserGroup,
		runCost,
	)
}

type routerGroupFn func(app *App)

func runRouterGroupFn(app *App, fns ...routerGroupFn) *App {
	for _, fn := range fns {
		fn(app)
	}
	return app
}
