package runner

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/1340691923/xwl_bi/application"
	"github.com/1340691923/xwl_bi/cmd/sinker/action"
	"github.com/1340691923/xwl_bi/cmd/sinker/geoip"
	"github.com/1340691923/xwl_bi/engine/logs"
	"github.com/1340691923/xwl_bi/model"
	"github.com/1340691923/xwl_bi/platform-basic-libs/service/consumer_data"
	"github.com/1340691923/xwl_bi/platform-basic-libs/sinker"
	"github.com/1340691923/xwl_bi/platform-basic-libs/util"
	_ "github.com/ClickHouse/clickhouse-go/v2"
	_ "github.com/go-sql-driver/mysql"
	jsoniter "github.com/json-iterator/go"
	"go.uber.org/zap"
)

// Run 负责执行 sinker 的完整启动与退出流程。
func Run(configFileDir, configFileName, configFileExt string) {
	defer logPanicStack()

	app := newSinkerApplication(configFileDir, configFileName, configFileExt)
	if err := app.InitConfig().
		NotifyInitFnObservers().
		Error(); err != nil {
		logs.Logger.Error("Sinker 初始化失败", zap.Error(err))
		panic(err)
	}
	defer app.Close()

	geoResolver, err := geoip.NewGeoip()
	if err != nil {
		logs.Logger.Error("GeoIP 初始化失败", zap.Error(err))
		panic(err)
	}
	defer geoResolver.Close()

	startPprofServer(model.GlobConfig.Sinker.PprofHttpPort)

	runtimeState, err := newSinkerRuntime(geoResolver)
	if err != nil {
		logs.Logger.Error("创建 sinker 运行时失败", zap.Error(err))
		panic(err)
	}

	runtimeState.startBackgroundLoops()
	runtimeState.runConsumers()
	runtimeState.waitForExit(app)
}

func newSinkerApplication(configFileDir, configFileName, configFileExt string) *application.App {
	return application.NewApp(
		"sinker",
		application.WithConfigFileDir(configFileDir),
		application.WithConfigFileName(configFileName),
		application.WithConfigFileExt(configFileExt),
		application.RegisterInitFnObserver(application.InitLogs),
		application.RegisterInitFnObserver(application.InitMysql),
		application.RegisterInitFnObserver(application.InitClickHouse),
		application.RegisterInitFnObserver(application.InitRedisPool),
	)
}

// logPanicStack 在入口层统一打印 panic 栈，避免线上只看到进程退出。
func logPanicStack() {
	if r := recover(); r != nil {
		buf := make([]byte, 2048)
		n := runtime.Stack(buf, false)
		stackInfo := fmt.Sprintf("%s", buf[:n])
		logs.Logger.Sugar().Errorf("panic stack info %s", stackInfo)
		logs.Logger.Sugar().Errorf("---> consumer_data error: %v", r)
	}
}

func startPprofServer(port uint16) {
	if port == 0 {
		return
	}

	go func() {
		httpPort := ":" + strconv.Itoa(int(port))
		if err := http.ListenAndServe(httpPort, nil); err != nil {
			logs.Logger.Info("pprof server stopped", zap.Error(err))
		}
	}()

	log.Println(fmt.Sprintf("sinker 服务启动成功，性能检测入口为: http://127.0.0.1:%v", port))
}

type sinkerRuntime struct {
	realTimeWarehousing  *consumer_data.RealTimeWarehousing
	reportAcceptStatus   *consumer_data.ReportAcceptStatus
	reportData2CK        *consumer_data.ReportData2CK
	reportConsumerPool   *util.DynamicWorkerPool
	reportPersistPool    *util.DynamicWorkerPool
	adminServer          *sinkerAdminServer
	partsMonitor         *clickHousePartsMonitor
	historyReplayBlocker *historyReplayBlocker
	realTimeConsumer     *sinker.KafkaSarama
	reportConsumer       *sinker.KafkaSarama
	reportPipeline       *reportConsumerPipeline
	reportRateSampler    *consumerRateSampler
	realTimeRateSampler  *consumerRateSampler
	protector            *reportConsumptionProtector
	backgroundStop       chan struct{}
	backgroundOnce       sync.Once
}

func newSinkerRuntime(geoResolver *geoip.Geoip2) (*sinkerRuntime, error) {
	var (
		sinkerC                      = model.GlobConfig.Sinker.Normalize()
		json                         = jsoniter.ConfigCompatibleWithStandardLibrary
		decoder                      = newMessageDecoder(json)
		backgroundStop               = make(chan struct{})
		insertPressureSettings       = queryClickHouseInsertPressureSettings()
		datePartitionLimit           = insertPressureSettings.MaxPartitionsPerInsertBlock
		historyReplayBlockerVal      = newHistoryReplayBlocker(consumer_data.SidecarRetentionCutoff(time.Now().Local()))
		realTimePartsGuard           = consumer_data.NewPartsPressureGuard(consumer_data.TableNameRealTimeWarehousing)
		reportAcceptStatusPartsGuard = consumer_data.NewPartsPressureGuard(consumer_data.TableNameAcceptanceStatus)
		realTimeWarehousing          = consumer_data.NewRealTimeWarehousingWithPartitionLimit(sinkerC.RealTimeWarehousing, datePartitionLimit)
		reportAcceptStatus           = consumer_data.NewReportAcceptStatusWithPartitionLimit(sinkerC.ReportAcceptStatus, datePartitionLimit)
		reportData2CK                = consumer_data.NewReportData2CK(sinkerC.ReportData2CK)
		partsMonitor                 = newClickHousePartsMonitor(insertPressureSettings, map[string]*consumer_data.PartsPressureGuard{
			consumer_data.TableNameRealTimeWarehousing: realTimePartsGuard,
			consumer_data.TableNameAcceptanceStatus:    reportAcceptStatusPartsGuard,
		}, backgroundStop)
	)

	realTimeWarehousing.SetPartsPressureGuard(realTimePartsGuard)
	reportAcceptStatus.SetPartsPressureGuard(reportAcceptStatusPartsGuard)

	realtimeHandler := newRealTimeMessageHandler(realTimeWarehousing, historyReplayBlockerVal, partsMonitor.NotifyWrite)
	reportHandler := newReportMessageHandler(
		newGeoEnricher(geoResolver),
		newActionSchemaSynchronizer(action.AddTableColumn),
		metaEventRecorderFunc(action.AddMetaEvent),
		reportAcceptStatus,
		reportData2CK,
		historyReplayBlockerVal,
		partsMonitor.NotifyWrite,
	)

	reportConsumerPool, err := util.NewDynamicWorkerPool(buildReportConsumerPoolConfig(sinkerC.ReportConsumerPool))
	if err != nil {
		return nil, err
	}
	reportPersistPool, err := util.NewDynamicWorkerPool(buildReportPersistPoolConfig(sinkerC.ReportPersistPool))
	if err != nil {
		_ = reportConsumerPool.Close()
		return nil, err
	}
	reportAcceptStatus.SetAsyncExecutor(reportPersistPool)
	reportData2CK.SetAsyncExecutor(reportPersistPool)
	adminServer, err := newSinkerAdminServer(sinkerC)
	if err != nil {
		_ = reportPersistPool.Close()
		_ = reportConsumerPool.Close()
		return nil, err
	}

	var (
		realTimeConsumer = sinker.NewKafkaSarama()
		reportConsumer   = sinker.NewKafkaSarama()
		reportPipeline   = newReportConsumerPipeline(reportHandler, reportConsumerPool)
	)

	if err := initConsumer(realTimeConsumer, model.GlobConfig.Comm.Kafka.RealTimeDataGroup, decoder.Wrap(realtimeHandler), func(generationID int32) {}); err != nil {
		return nil, err
	}
	if err := initConsumer(reportConsumer, model.GlobConfig.Comm.Kafka.ReportData2CKGroup, decoder.Wrap(reportPipeline), reportPipeline.OnSessionCleanup); err != nil {
		_ = reportPipeline.Close()
		return nil, err
	}

	reportRateSampler, err := newConsumerRateSampler(
		model.GlobConfig.Comm.Kafka.ReportData2CKGroup,
		model.GlobConfig.Comm.Kafka.ReportTopicName,
		reportConsumer,
		model.GlobConfig.Comm.Kafka,
		sinkerC.Protection,
	)
	if err != nil {
		_ = reportPipeline.Close()
		_ = reportPersistPool.Close()
		_ = reportConsumerPool.Close()
		return nil, err
	}

	realTimeRateSampler, err := newConsumerRateSampler(
		model.GlobConfig.Comm.Kafka.RealTimeDataGroup,
		model.GlobConfig.Comm.Kafka.ReportTopicName,
		realTimeConsumer,
		model.GlobConfig.Comm.Kafka,
		sinkerC.Protection,
	)
	if err != nil {
		_ = reportRateSampler.Close()
		_ = reportPipeline.Close()
		_ = reportPersistPool.Close()
		_ = reportConsumerPool.Close()
		return nil, err
	}

	protector := newReportConsumptionProtector(
		sinkerC.Protection,
		reportConsumer,
		realTimeConsumer,
		reportRateSampler,
		realTimeRateSampler,
		reportPipeline,
		reportConsumerPool,
		reportPersistPool,
	)
	adminServer.SetProtectionController(protector)

	return &sinkerRuntime{
		realTimeWarehousing:  realTimeWarehousing,
		reportAcceptStatus:   reportAcceptStatus,
		reportData2CK:        reportData2CK,
		reportConsumerPool:   reportConsumerPool,
		reportPersistPool:    reportPersistPool,
		adminServer:          adminServer,
		partsMonitor:         partsMonitor,
		historyReplayBlocker: historyReplayBlockerVal,
		realTimeConsumer:     realTimeConsumer,
		reportConsumer:       reportConsumer,
		reportPipeline:       reportPipeline,
		reportRateSampler:    reportRateSampler,
		realTimeRateSampler:  realTimeRateSampler,
		protector:            protector,
		backgroundStop:       backgroundStop,
	}, nil
}

func initConsumer(consumer *sinker.KafkaSarama, consumerGroup string, handle func(model.InputMessage, func()), cleanupFn func(generationID int32)) error {
	return consumer.Init(
		model.GlobConfig.Comm.Kafka,
		model.GlobConfig.Comm.Kafka.ReportTopicName,
		consumerGroup,
		handle,
		cleanupFn,
	)
}

func (r *sinkerRuntime) startBackgroundLoops() {
	r.adminServer.Start()
	go action.MysqlConsumer()
	go sinker.ClearDimsCacheByTimeWithStop(time.Minute*30, r.backgroundStop)
	go consumeKafkaConsumerErrors("report-consumer", model.GlobConfig.Comm.Kafka.ReportTopicName, r.reportConsumer.Errors())
	go consumeKafkaConsumerErrors("realtime-consumer", model.GlobConfig.Comm.Kafka.ReportTopicName, r.realTimeConsumer.Errors())
	startSinkerWorkerPoolStatsLoops(r.reportConsumerPool, r.backgroundStop)
	startSinkerWorkerPoolStatsLoops(r.reportPersistPool, r.backgroundStop)
	startHistoryReplayBlockerLogLoop(r.historyReplayBlocker, r.backgroundStop)
	r.reportRateSampler.Start(r.backgroundStop)
	r.realTimeRateSampler.Start(r.backgroundStop)
	r.protector.Start(r.backgroundStop)
}

func (r *sinkerRuntime) runConsumers() {
	go r.reportConsumer.Run()
	go r.realTimeConsumer.Run()
}

// waitForExit 先停 consumer，再关闭 worker pool，最后冲刷批量器。
// 这里额外记录 worker pool close 之后的最终快照，避免 cleanup 中间态误导排查。
func (r *sinkerRuntime) waitForExit(app *application.App) {
	app.WaitForExitSign(
		func() {
			r.backgroundOnce.Do(func() {
				close(r.backgroundStop)
			})
			if err := r.adminServer.Stop(); err != nil {
				logs.Logger.Sugar().Infof("sinker admin server 停止失败: %v", err)
			}
			if err := r.reportRateSampler.Close(); err != nil {
				logs.Logger.Sugar().Infof("report rate sampler 停止失败: %v", err)
			}
			if err := r.realTimeRateSampler.Close(); err != nil {
				logs.Logger.Sugar().Infof("real time rate sampler 停止失败: %v", err)
			}
			if err := r.reportConsumer.Stop(); err != nil {
				logs.Logger.Sugar().Infof("reportData2CKSarama 停止失败: %v", err)
			}
			if err := r.realTimeConsumer.Stop(); err != nil {
				logs.Logger.Sugar().Infof("realTimeDataSarama 停止失败: %v", err)
			}
			pipelineCloseErr := r.reportPipeline.Close()
			if pipelineCloseErr != nil {
				logs.Logger.Sugar().Infof("report consumer worker pool 停止失败: %v", pipelineCloseErr)
			} else {
				r.reportPipeline.LogFinalSnapshot("report consumer final snapshot after worker pool close")
			}
			if err := r.reportPersistPool.Close(); err != nil {
				logs.Logger.Sugar().Infof("report persist worker pool 鍋滄澶辫触: %v", err)
			}
			r.realTimeWarehousing.BypassPartsPressureGuard(true)
			r.reportAcceptStatus.BypassPartsPressureGuard(true)
		},
		func() {
			if err := r.reportData2CK.FlushAll(); err != nil {
				logs.Logger.Sugar().Infof("清理 reportData2CK FlushAll 失败: %v", err)
			} else {
				logs.Logger.Sugar().Infof("清理 reportData2CK 完毕")
			}
		},
		func() {
			if err := r.realTimeWarehousing.FlushAll(); err != nil {
				logs.Logger.Sugar().Infof("清理 realTimeWarehousing 失败: %v", err)
			} else {
				logs.Logger.Sugar().Infof("清理 realTimeWarehousing 完毕")
			}
		},
		func() {
			if err := r.reportAcceptStatus.FlushAll(); err != nil {
				logs.Logger.Sugar().Infof("清理 reportAcceptStatus 失败: %v", err)
			} else {
				logs.Logger.Sugar().Infof("清理 reportAcceptStatus 完毕")
			}
		},
	)
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
