package myapp

import (
	"strings"
	"sync"

	"github.com/1340691923/xwl_bi/engine/db"
	"github.com/1340691923/xwl_bi/engine/logs"
	"github.com/garyburd/redigo/redis"
	"go.uber.org/zap"
)

const AppidToTableidHash = "AppidToTableid"

var (
	appTableIDInvalidatorMu sync.RWMutex
	appTableIDInvalidator   func(cacheKey string)
)

// BuildAppTableIDKey 统一构造 `appid/appkey -> tableId` 的 Redis 字段键。
//
// 为什么把它放在 myapp：
// 1. `AppidToTableid` Hash 的读写 owner 本身就在 myapp 包里。
// 2. report 领域层和后台管理写入层都依赖同一份 key 语义。
// 3. 这样 report 只需直接调用 myapp.BuildAppTableIDKey，不需要再绕一层额外目录。
//
// 示例：
// 1. appid="1001", appkey="demo" -> "1001_xwl_demo"
func BuildAppTableIDKey(appid, appkey string) string {
	var builder strings.Builder
	builder.Grow(len(appid) + len(appkey) + len("_xwl_"))
	builder.WriteString(appid)
	builder.WriteString("_xwl_")
	builder.WriteString(appkey)
	return builder.String()
}

// RegisterAppTableIDCacheInvalidator 注册“Redis 成功写入后，如何让本地旁路缓存失效”的回调。
//
// 这里由 myapp 提供注册点，而不是让 myapp 反向 import report：
// 1. myapp 负责 Redis 真值。
// 2. report 负责本地旁路缓存。
// 3. 通过注册回调桥接两者，可以避免包循环。
func RegisterAppTableIDCacheInvalidator(fn func(cacheKey string)) {
	appTableIDInvalidatorMu.Lock()
	defer appTableIDInvalidatorMu.Unlock()
	appTableIDInvalidator = fn
}

// NotifyAppTableIDCacheChanged 在 AppidToTableid 被写入后，通知本地旁路缓存删除对应 key。
//
// 这是旁路缓存模型的核心：
// 1. 只失效发生变化的那个 cacheKey。
// 2. 不做定时全量清空。
func NotifyAppTableIDCacheChanged(cacheKey string) {
	appTableIDInvalidatorMu.RLock()
	fn := appTableIDInvalidator
	appTableIDInvalidatorMu.RUnlock()
	if fn != nil {
		fn(cacheKey)
	}
}

func SetAppidToTableid(appid, appkey string, tableID int) (err error) {
	conn := db.RedisPool.Get()
	defer conn.Close()
	cacheKey := BuildAppTableIDKey(appid, appkey)
	_, err = conn.Do("hset", AppidToTableidHash, cacheKey, tableID)
	if err != nil {
		logs.Logger.Error("SetAppidToTableid err", zap.Error(err))
		return
	}
	NotifyAppTableIDCacheChanged(cacheKey)
	return
}

func GetAppidToTableid(conn redis.Conn, key string) (tableID string, err error) {
	tableID, err = redis.String(conn.Do("hget", AppidToTableidHash, key))
	return
}

func DeleteAppidToTableid(appid, appkey string) (err error) {
	conn := db.RedisPool.Get()
	defer conn.Close()
	cacheKey := BuildAppTableIDKey(appid, appkey)
	_, err = conn.Do("hdel", AppidToTableidHash, cacheKey)
	if err != nil {
		logs.Logger.Error("DeleteAppidToTableid", zap.Error(err))
		return
	}
	NotifyAppTableIDCacheChanged(cacheKey)
	return
}
