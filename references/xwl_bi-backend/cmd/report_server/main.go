/*
**
该项目用于采集用户行为数据，以及埋点数据分析服务。
*/
package main

import (
	"flag"

	"github.com/1340691923/xwl_bi/application"
	"github.com/1340691923/xwl_bi/engine/logs"
	_ "github.com/ClickHouse/clickhouse-go/v2"
	_ "github.com/go-sql-driver/mysql"
	"go.uber.org/zap"
)

var (
	configFileDir  string
	configFileName string
	configFileExt  string
)

func init() {
	flag.StringVar(&configFileDir, "configFileDir", "config", "配置文件夹名")
	flag.StringVar(&configFileName, "configFileName", "config", "配置文件名")
	flag.StringVar(&configFileExt, "configFileExt", "json", "配置文件后缀")
}

// main 只负责 report_server 的启动装配，不再直接承载完整上报业务逻辑。
//
// 当前入口固定只做四件事：
// 1. 解析命令行并初始化基础依赖。
// 2. 构造 runtime，把 resolver、producer、HTTP handler 装配起来。
// 3. 启动后台循环和 HTTP 服务。
// 4. 在退出阶段统一停止后台循环并关闭服务。
func main() {
	flag.Parse()

	app := newReportApplication()
	if err := app.InitConfig().NotifyInitFnObservers().Error(); err != nil {
		logs.Logger.Error("数据系统 初始化失败", zap.Error(err))
		panic(err)
	}
	defer app.Close()

	runtimeState := newReportRuntime()
	runtimeState.startBackgroundLoops()
	runtimeState.run()
	runtimeState.waitForExit(app)
}

func newReportApplication() *application.App {
	return application.NewApp(
		"report_server",
		application.WithConfigFileDir(configFileDir),
		application.WithConfigFileName(configFileName),
		application.WithConfigFileExt(configFileExt),
		application.RegisterInitFnObserver(application.InitLogs),
		application.RegisterInitFnObserver(application.InitKafkaSyncProduce),
		application.RegisterInitFnObserver(application.InitKafkaAsyncProduce),
		application.RegisterInitFnObserver(application.InitRedisPool),
		application.RegisterInitFnObserver(application.InitMysql),
		application.RegisterInitFnObserver(application.InitClickHouse),
	)
}
