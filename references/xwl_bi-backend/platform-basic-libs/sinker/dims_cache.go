package sinker

import (
	"bytes"
	"database/sql"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/1340691923/xwl_bi/engine/db"
	"github.com/1340691923/xwl_bi/engine/logs"
	"github.com/1340691923/xwl_bi/platform-basic-libs/sinker/model"
	parser "github.com/1340691923/xwl_bi/platform-basic-libs/sinker/parse"
	"github.com/1340691923/xwl_bi/platform-basic-libs/util"
	"github.com/garyburd/redigo/redis"
	"github.com/jmoiron/sqlx"
	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

var (
	ErrTblNotExist       = errors.Errorf("table doesn't exist")
	selectSQLTemplate    = `select name, type, default_kind from system.columns where database = '%s' and table = '%s'`
	lowCardinalityRegexp = regexp.MustCompile(`LowCardinality\((.+)\)`)
)

const DimsHash = "dimsHash_"

var dimsCacheMap sync.Map

func GetDimsCachekey(database, table string) string {
	b := bytes.Buffer{}
	b.WriteString(DimsHash)
	b.WriteString(database)
	b.WriteString("_")
	b.WriteString(table)
	return b.String()
}

// ClearDimsCacheByTime 定时清理本地 + Redis 的 dims 缓存。
//
// 典型场景：
// 1. 表刚刚动态补列
// 2. 老缓存里还是旧列集合
// 3. 如果不清掉，后续写入仍然会按旧列顺序展开
func ClearDimsCacheByTime(clearTime time.Duration) {
	for {
		time.Sleep(clearTime)
		dimsCacheMap.Range(func(key, value interface{}) bool {
			ClearDimsCacheByRedis(key.(string))
			dimsCacheMap.Delete(key)
			return true
		})
	}
}

func ClearDimsCacheByTimeBylocal(clearTime time.Duration) {
	for {
		time.Sleep(clearTime)
		dimsCacheMap.Range(func(key, value interface{}) bool {
			ClearDimsCacheByRedis(key.(string))
			dimsCacheMap.Delete(key)
			return true
		})
	}
}

// ClearDimsCacheByTimeWithStop 允许调用方在进程退出时停止 dims 缓存清理循环。
//
// 设计原因：
// 1. report_server 和 sinker 都可能在后台启动这个循环。
// 2. 如果没有 stop 路径，进程退出阶段就无法表达“后台任务已经停止”。
func ClearDimsCacheByTimeWithStop(clearTime time.Duration, stop <-chan struct{}) {
	ticker := time.NewTicker(clearTime)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			dimsCacheMap.Range(func(key, value interface{}) bool {
				ClearDimsCacheByRedis(key.(string))
				dimsCacheMap.Delete(key)
				return true
			})
		case <-stop:
			return
		}
	}
}

// ClearDimsCacheByTimeBylocalWithStop 与 ClearDimsCacheByTimeWithStop 语义一致，
// 保留“仅本地循环调用”的命名习惯，方便渐进迁移旧调用方。
func ClearDimsCacheByTimeBylocalWithStop(clearTime time.Duration, stop <-chan struct{}) {
	ClearDimsCacheByTimeWithStop(clearTime, stop)
}

// ClearDimsCacheByRedis 负责删除 Redis 里的某个 dims 缓存键。
func ClearDimsCacheByRedis(key string) {
	redisConn := db.RedisPool.Get()
	defer redisConn.Close()

	_, err := redisConn.Do("unlink", key)
	if err != nil {
		if _, err = redisConn.Do("del", key); err != nil {
			logs.Logger.Error("err", zap.Error(err))
		}
	}
}

// ClearDimsCacheByKey 只删除当前进程内的本地 dims 缓存。
func ClearDimsCacheByKey(key string) {
	dimsCacheMap.Delete(key)
}

// GetDims 负责读取某张 ClickHouse 表的列定义，并带缓存。
//
// 查找顺序是：
// 1. 本地进程内缓存
// 2. Redis 缓存
// 3. ClickHouse system.columns
//
// 示例：
// 1. 第一次读 xwl_event51，缓存未命中 -> 查 CK -> 回填本地和 Redis
// 2. 后续再读 xwl_event51，直接走本地或 Redis
func GetDims(database, table string, excludedColumns []string, conn *sqlx.DB, onlyRedis bool) (dims []*model.ColumnWithType, err error) {
	dimsCachekey := GetDimsCachekey(database, table)
	if !onlyRedis {
		if cache, load := dimsCacheMap.Load(dimsCachekey); load {
			return cache.([]*model.ColumnWithType), nil
		}
	}

	var json = jsoniter.ConfigCompatibleWithStandardLibrary

	redisConn := db.RedisPool.Get()
	defer redisConn.Close()
	dimsBytes, redisErr := redis.Bytes(redisConn.Do("get", dimsCachekey))

	if redisErr == nil && len(dimsBytes) != 0 {
		dimsCache, err := util.GzipUnCompressByte(dimsBytes)
		if err == nil {
			jsonErr := json.Unmarshal(dimsCache, &dims)
			if jsonErr == nil {
				dimsCacheMap.Store(dimsCachekey, dims)
				return dims, nil
			}
			logs.Logger.Error("jsonErr", zap.Error(jsonErr))
		} else {
			logs.Logger.Error("GzipUnCompressByte Err", zap.Error(err))
		}
	} else if redisErr != redis.ErrNil {
		logs.Logger.Error("redisErr", zap.Error(redisErr))
	}

	var rs *sql.Rows
	if rs, err = conn.Query(fmt.Sprintf(selectSQLTemplate, database, table)); err != nil {
		return dims, errors.Wrapf(err, "")
	}
	defer rs.Close()

	var name, typ, defaultKind string
	for rs.Next() {
		if err = rs.Scan(&name, &typ, &defaultKind); err != nil {
			return dims, errors.Wrapf(err, "")
		}
		typ = lowCardinalityRegexp.ReplaceAllString(typ, "$1")
		if !util.InstrArr(excludedColumns, name) && defaultKind != "MATERIALIZED" {
			tp, nullable := parser.WhichType(typ)
			dims = append(dims, &model.ColumnWithType{Name: name, Type: tp, Nullable: nullable, SourceName: GetSourceName(name)})
		}
	}
	if len(dims) == 0 {
		return dims, errors.Wrapf(ErrTblNotExist, "%s.%s", database, table)
	}
	dimsCacheMap.Store(dimsCachekey, dims)

	res, _ := json.Marshal(dims)
	s, err := util.GzipCompressByte(res)
	if err != nil {
		return dims, err
	}
	_, err = redisConn.Do("SETEX", dimsCachekey, 60*60*6, s)
	return dims, err
}

// GetSourceName 用于把带点路径的列名转换成 parser 可识别的 sourceName。
//
// 示例：
// 1. 输入 "a.b"
// 2. 输出 "a\\.b"
func GetSourceName(name string) (sourcename string) {
	sourcename = strings.Replace(name, ".", "\\.", -1)
	return
}
