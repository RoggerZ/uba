package application

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/url"
	"time"

	"github.com/1340691923/xwl_bi/controller"
	"github.com/1340691923/xwl_bi/engine/db"
	"github.com/1340691923/xwl_bi/engine/logs"
	"github.com/1340691923/xwl_bi/model"
	"github.com/1340691923/xwl_bi/platform-basic-libs/rbac"
	"github.com/1340691923/xwl_bi/platform-basic-libs/sinker"
	"github.com/1340691923/xwl_bi/platform-basic-libs/util"
	clickhouse "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/tidwall/gjson"
	"go.uber.org/zap"

	"path/filepath"
	"strconv"
)

// InitLogs 初始化日志
func InitLogs() (fn func(), err error) {
	logger := logs.NewLog(
		logs.WithLogPath(filepath.Join(model.GlobConfig.Comm.Log.LogDir, model.CmdName)),
		logs.WithStorageDays(model.GlobConfig.Comm.Log.StorageDays),
		logs.WithLevel(model.GlobConfig.Comm.Log.Level),
	)
	logs.Logger, err = logger.InitLog()
	if err != nil {
		return
	}
	log.Println(fmt.Sprintf("日志组件初始化成功！日志所在目录：%v，保存天数为：%v", model.GlobConfig.Comm.Log.LogDir, model.GlobConfig.Comm.Log.StorageDays))
	fn = func() {}
	return
}

// InitMysql 初始化mysql连接
func InitMysql() (fn func(), err error) {
	config := model.GlobConfig.Comm.Mysql.Normalize()
	dbSource := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s",
		config.Username,
		config.Pwd,
		config.IP,
		config.Port,
		config.DbName)
	db.Sqlx, err = db.NewSQLXWithOptions(
		"mysql",
		dbSource,
		config.MaxOpenConns,
		config.MaxIdleConns,
		db.DBHealthOptions{
			Name:                   "mysql",
			Enabled:                config.HealthCheck.Enabled != nil && *config.HealthCheck.Enabled,
			PingInterval:           time.Duration(config.HealthCheck.PingIntervalSeconds) * time.Second,
			FailuresBeforeDegraded: config.HealthCheck.FailuresBeforeDegraded,
		},
	)
	if err != nil {
		return
	}

	log.Println(fmt.Sprintf("Mysql组件初始化成功！连接：%v，最大打开连接数：%v，最大等待连接数:%v",
		dbSource,
		config.MaxOpenConns,
		config.MaxIdleConns,
	))
	fn = func() {}
	return
}

// InitClickHouse 初始化mysql连接
func InitClickHouse() (fn func(), err error) {
	config := model.GlobConfig.Comm.ClickHouse
	dbSource := buildClickHouseStdDSN(
		config.IP,
		config.Port,
		config.DbName,
		config.Username,
		config.Pwd,
		config.GetMaxQuerySize(),
	)

	db.ClickHouseSqlx, err = db.NewSQLX(
		"clickhouse",
		dbSource,
		config.MaxOpenConns,
		config.MaxIdleConns,
	)

	if err != nil {
		return
	}
	db.ClickHouseNative, err = clickhouse.Open(&clickhouse.Options{
		Addr: []string{fmt.Sprintf("%s:%s", config.IP, config.Port)},
		Auth: clickhouse.Auth{
			Database: config.DbName,
			Username: config.Username,
			Password: config.Pwd,
		},
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},
		DialTimeout:     30 * time.Second,
		MaxOpenConns:    maxInt(config.MaxOpenConns, 1),
		MaxIdleConns:    maxInt(config.MaxIdleConns, 1),
		ConnMaxLifetime: time.Hour,
		Settings: clickhouse.Settings{
			"max_query_size": config.GetMaxQuerySize(),
		},
	})
	if err != nil {
		_ = db.ClickHouseSqlx.Close()
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err = db.ClickHouseNative.Ping(ctx); err != nil {
		_ = db.ClickHouseNative.Close()
		_ = db.ClickHouseSqlx.Close()
		return
	}
	log.Println(fmt.Sprintf("ClickHouse组件初始化成功！连接：%v，最大打开连接数：%v，最大等待连接数:%v",
		dbSource,
		config.MaxOpenConns,
		config.MaxIdleConns,
	))
	fn = func() {
		if db.ClickHouseNative != nil {
			_ = db.ClickHouseNative.Close()
		}
	}
	return
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func buildClickHouseStdDSN(host, port, database, username, password string, maxQuerySize int) string {
	query := url.Values{}
	query.Set("username", username)
	query.Set("password", password)
	query.Set("compress", "lz4")
	query.Set("max_query_size", fmt.Sprintf("%d", maxQuerySize))

	return (&url.URL{
		Scheme:   "clickhouse",
		Host:     fmt.Sprintf("%s:%s", host, port),
		Path:     "/" + database,
		RawQuery: query.Encode(),
	}).String()
}

// InitRedisPool 初始化redis
func InitRedisPool() (fn func(), err error) {
	config := model.GlobConfig.Comm.Redis

	db.RedisPool = db.NewRedisPool(config.Addr, config.Passwd, config.Db, config.MaxIdle, config.MaxActive)

	log.Println(fmt.Sprintf("Redis组件初始化成功！连接：%v，DB：%v，密码:%v MaxIdle:%v MaxActive:%v",
		config.Addr,
		config.Db,
		config.Passwd,
		config.MaxIdle,
		config.MaxActive,
	))
	fn = func() {}
	return
}

// InitTask 初始化项目启动任务
func InitTask() (fn func(), err error) {
	fn = func() {}
	return
}

// InitRbac 初始化项目启动任务
func InitRbac() (fn func(), err error) {
	config := model.GlobConfig.Comm.Mysql
	dbSource := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s",
		config.Username,
		config.Pwd,
		config.IP,
		config.Port,
		config.DbName)
	err = rbac.Run("mysql", dbSource)
	if err != nil {
		return
	}
	log.Println(fmt.Sprintf("Rbac组件初始化成功！连接：%v",
		dbSource,
	))
	return
}

// InitOpenWinBrowser 掉起浏览器
func InitOpenWinBrowser() (fn func(), err error) {
	config := model.GlobConfig
	if !config.Manager.DeBug {
		port := ":" + strconv.Itoa(int(config.Manager.Port))
		uri := fmt.Sprintf("%s%s", "http://127.0.0.1", port)
		_ = util.OpenWinBrowser(uri)
		log.Println(fmt.Sprintf("将打开浏览器！地址为：%v",
			uri,
		))
	}
	fn = func() {}
	return
}

// InitKafkaAsyncProduce 初始化kafka异步生产者
func InitKafkaAsyncProduce() (fn func(), err error) {
	config := model.GlobConfig.Comm.Kafka
	conn, err := db.NewKafkaAsyncProduce(config.Addresses, config.Username, config.Password)
	if err != nil {
		return
	}
	db.KafkaASyncProducer = conn
	fn = func() {
		log.Println("KafkaASyncProducer 关闭了")
		_ = db.KafkaASyncProducer.Close()
	}
	return
}

// InitKafkaSyncProduce 初始化kafka同步步生产者
func InitKafkaSyncProduce() (fn func(), err error) {
	config := model.GlobConfig.Comm.Kafka
	conn, err := db.NewKafkaSyncProduce(config.Addresses, config.Username, config.Password)
	if err != nil {
		return
	}
	db.KafkaSyncProducer = conn

	fn = func() {
		log.Println("KafkaSyncProducer 关闭了")
		_ = db.KafkaSyncProducer.Close()
	}

	return
}

func InitDebugSarama() (fn func(), err error) {
	debugSarama := sinker.NewKafkaSarama()
	err = debugSarama.Init(model.GlobConfig.Comm.Kafka, model.GlobConfig.Comm.Kafka.DebugDataTopicName, model.GlobConfig.Comm.Kafka.DebugDataGroup, func(msg model.InputMessage, markFn func()) {

		distinctId := gjson.GetBytes(msg.Value, "distinct_id").String()

		managerMap, ok := controller.ConnUUidMap.Load(distinctId)

		if ok {
			managerMap.(*controller.ManagerConnMap).Conns.Range(func(key, value interface{}) bool {

				if err := value.(*websocket.Conn).WriteJSON(map[string]interface{}{
					"code": 1,
					"data": util.Bytes2str(msg.Value),
				}); err != nil {
					if err == io.EOF {
						logs.Logger.Error("客户端已经断开WsSocket!", zap.Error(err))
					} else if err.Error() == "use of closed network connection" {
						logs.Logger.Error("服务端已经断开WsSocket!", zap.Error(err))
					} else {
						logs.Logger.Error("socket err!", zap.Error(err))
					}
					managerMap.(*controller.ManagerConnMap).DeleteConn(key.(string))
					controller.ConnUUidMap.Store(distinctId, managerMap)
				}
				return true
			})
		}

	}, func(generationID int32) {

	})
	if err != nil {
		return
	}

	go debugSarama.Run()

	log.Println(fmt.Sprintf("Debug数据消费者启动！"))

	fn = func() {
		_ = debugSarama.Stop()
	}
	return
}
