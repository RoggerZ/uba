package main

import (
	"fmt"
	"log"
	"net/http/pprof"
	"sync"
	"time"

	"github.com/1340691923/xwl_bi/application"
	"github.com/1340691923/xwl_bi/controller"
	"github.com/1340691923/xwl_bi/engine/db"
	"github.com/1340691923/xwl_bi/engine/logs"
	"github.com/1340691923/xwl_bi/middleware"
	"github.com/1340691923/xwl_bi/model"
	"github.com/1340691923/xwl_bi/platform-basic-libs/service/report"
	"github.com/1340691923/xwl_bi/platform-basic-libs/sinker"
	sinkerModel "github.com/1340691923/xwl_bi/platform-basic-libs/sinker/model"
	"github.com/buaazp/fasthttprouter"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
	"go.uber.org/zap"
)

type reportRuntime struct {
	server         *fasthttp.Server
	backgroundStop chan struct{}
	backgroundOnce sync.Once
}

// newReportRuntime 负责把 report_server 运行期依赖完整装配出来。
//
// 这里要特别说明 `tableId` 缓存策略：
// 1. 当前不再采用“定时全量清空本地缓存”的方式。
// 2. 改为旁路缓存模型：Resolve 时查本地缓存，Redis 写入时只失效被修改的那个 key。
// 3. 因此 runtime 不再负责启动 tableId 定时刷新循环，只负责把 resolver 注入给业务 handler。
//
// 这样做的好处是：
// 1. 没有变化的 key 可以一直留在内存中，命中率更高。
// 2. 后台 `SetAppidToTableid/DeleteAppidToTableid` 只会失效对应 key，不会把所有 appid/appkey 一起打掉。
func newReportRuntime() *reportRuntime {
	tableIDResolver := report.DefaultTableIDResolver()
	debugMembership := report.NewDebugMembershipChecker(nil)
	debugProducer := report.NewDefaultDebugDataProducer()
	reportProducer := report.NewDefaultKafkaDataProducer()

	reportHandler := controller.NewReportHandler(controller.ReportHandlerDependencies{
		ResolveTableID: tableIDResolver.Resolve,
		BuildPayload:   report.DefaultPayloadBuilderRegistry().Build,
		IsDebugDevice:  debugMembership.IsDebugDevice,
		SendDebugData:  debugProducer.Send,
		SendReportData: reportProducer.Send,
		LoadDims: func(tableName string) ([]*sinkerModel.ColumnWithType, error) {
			return sinker.GetDims(model.GlobConfig.Comm.ClickHouse.DbName, tableName, []string{}, db.ClickHouseSqlx, true)
		},
	})

	router := newReportRouter(reportHandler)
	return &reportRuntime{
		server:         newReportServer(router.Handler),
		backgroundStop: make(chan struct{}),
	}
}

func newReportRouter(reportHandler fasthttp.RequestHandler) *fasthttprouter.Router {
	router := fasthttprouter.New()

	router.GET("/debug/pprof/", fasthttpadaptor.NewFastHTTPHandlerFunc(pprof.Index))
	router.GET("/debug/pprof/cmdline", fasthttpadaptor.NewFastHTTPHandlerFunc(pprof.Cmdline))
	router.GET("/debug/pprof/profile", fasthttpadaptor.NewFastHTTPHandlerFunc(pprof.Profile))
	router.GET("/debug/pprof/symbol", fasthttpadaptor.NewFastHTTPHandlerFunc(pprof.Symbol))
	router.GET("/debug/pprof/trace", fasthttpadaptor.NewFastHTTPHandlerFunc(pprof.Trace))
	router.GET("/debug/pprof/allocs", fasthttpadaptor.NewFastHTTPHandlerFunc(pprof.Handler("allocs").ServeHTTP))
	router.GET("/debug/pprof/block", fasthttpadaptor.NewFastHTTPHandlerFunc(pprof.Handler("block").ServeHTTP))
	router.GET("/debug/pprof/goroutine", fasthttpadaptor.NewFastHTTPHandlerFunc(pprof.Handler("goroutine").ServeHTTP))
	router.GET("/debug/pprof/heap", fasthttpadaptor.NewFastHTTPHandlerFunc(pprof.Handler("heap").ServeHTTP))
	router.GET("/debug/pprof/mutex", fasthttpadaptor.NewFastHTTPHandlerFunc(pprof.Handler("mutex").ServeHTTP))
	router.GET("/debug/pprof/threadcreate", fasthttpadaptor.NewFastHTTPHandlerFunc(pprof.Handler("threadcreate").ServeHTTP))

	router.POST("/test", func(ctx *fasthttp.RequestCtx) {
		ctx.WriteString(`{"code":0}`)
	})
	router.GET("/GetWordParse", controller.GetWordParse)
	router.POST(
		"/ingress/:typ/:appid/:appkey/:eventName/:debug",
		middleware.Cors(
			middleware.WechatSpider(
				middleware.ReportSyncMode(reportHandler),
			),
		),
	)

	return router
}

func newReportServer(handler fasthttp.RequestHandler) *fasthttp.Server {
	server := &fasthttp.Server{
		Handler: handler,
	}
	if model.GlobConfig.Report.ReadTimeout != 0 {
		server.ReadTimeout = time.Duration(model.GlobConfig.Report.ReadTimeout) * time.Second
	}
	if model.GlobConfig.Report.WriteTimeout != 0 {
		server.WriteTimeout = time.Duration(model.GlobConfig.Report.WriteTimeout) * time.Second
	}
	if model.GlobConfig.Report.MaxConnsPerIP != 0 {
		server.MaxConnsPerIP = model.GlobConfig.Report.MaxConnsPerIP
	}
	if model.GlobConfig.Report.MaxRequestsPerConn != 0 {
		server.MaxRequestsPerConn = model.GlobConfig.Report.MaxRequestsPerConn
	}
	if model.GlobConfig.Report.IdleTimeout != 0 {
		server.IdleTimeout = time.Duration(model.GlobConfig.Report.IdleTimeout) * time.Second
	}
	return server
}

// startBackgroundLoops 启动 report_server 依赖的后台循环。
//
// 当前固定包含：
// 1. Kafka async producer 错误消费。
// 2. dims 缓存清理循环。
//
// 注意：
// 1. tableId resolver 不再有“定时清空本地缓存”的后台循环。
// 2. 它现在是旁路缓存模式，靠 Redis 写操作触发对应 key 的定向失效。
func (r *reportRuntime) startBackgroundLoops() {
	go consumeAsyncProducerErrors(db.KafkaASyncProducer.Errors())
	go sinker.ClearDimsCacheByTimeBylocalWithStop(20*time.Second, r.backgroundStop)
}

func (r *reportRuntime) run() {
	go func() {
		port := fmt.Sprintf(":%v", model.GlobConfig.Report.ReportPort)
		logs.Logger.Sugar().Infof("service start")
		log.Println(fmt.Sprintf("上报服务启动成功 ,性能检测入口为: http://127.0.0.1:%v", model.GlobConfig.Report.ReportPort))
		if err := r.server.ListenAndServe(port); err != nil {
			logs.Logger.Error("service err", zap.Error(err))
			log.Panic(err)
		}
	}()
}

// waitForExit 统一处理 report_server 的退出收尾逻辑。
//
// 退出顺序固定为：
// 1. 先停后台循环，避免退出阶段仍有新的 dims 清理动作。
// 2. 再关闭 HTTP 服务。
//
// 这里不需要额外停止 tableId resolver：
// 1. resolver 已经没有后台 ticker。
// 2. 它只是一层本地旁路缓存，进程退出时会随进程内存一起回收。
func (r *reportRuntime) waitForExit(app *application.App) {
	app.WaitForExitSign(func() {
		r.backgroundOnce.Do(func() {
			close(r.backgroundStop)
		})

		logs.Logger.Sugar().Infof("数据上报服务停止中...")
		if err := r.server.Shutdown(); err != nil {
			logs.Logger.Sugar().Infof("数据上报服务停止失败 err", zap.Error(err))
		} else {
			logs.Logger.Sugar().Infof("数据上报服务停止成功...")
		}
	})
}
