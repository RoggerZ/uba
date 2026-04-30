package report

import (
	"sync"

	"github.com/1340691923/xwl_bi/engine/db"
	"github.com/1340691923/xwl_bi/engine/logs"
	"github.com/1340691923/xwl_bi/platform-basic-libs/my_error"
	"github.com/1340691923/xwl_bi/platform-basic-libs/service/myapp"
	"github.com/garyburd/redigo/redis"
	"go.uber.org/zap"
)

type tableIDFetcher func(key string) (string, error)

// TableIDResolver 负责管理“appid/appkey -> tableId”的本地旁路缓存。
//
// 设计原因：
// 1. 旧实现把缓存 map 和定时刷新散落成全局变量 + 自由函数，owner 不清晰。
// 2. 定时全量清理会把所有 key 一起打掉，命中率差，而且会让没有变化的 appid/appkey 也被迫重新查 Redis。
// 3. 改成旁路缓存后，只有 Redis `AppidToTableid` 发生写操作的那个 key 才会失效，更符合真实业务变更粒度。
//
// 示例：
// 1. Resolve("1001", "demo") 首次未命中缓存 -> 查 Redis -> 回填缓存
// 2. 同样的 appid/appkey 再次 Resolve -> 直接命中本地缓存
// 3. 如果后台调用 SetAppidToTableid("1001", "demo", 52)，myapp 会通知 resolver 只删除 "1001_xwl_demo" 这个本地 key
type TableIDResolver struct {
	fetch tableIDFetcher
	cache sync.Map
}

func NewTableIDResolver(fetch tableIDFetcher) *TableIDResolver {
	if fetch == nil {
		fetch = func(key string) (string, error) {
			conn := db.RedisPool.Get()
			defer conn.Close()
			return myapp.GetAppidToTableid(conn, key)
		}
	}

	return &TableIDResolver{
		fetch: fetch,
	}
}

var defaultTableIDResolver = NewTableIDResolver(nil)

func init() {
	myapp.RegisterAppTableIDCacheInvalidator(DefaultTableIDResolver().InvalidateByKey)
}

// DefaultTableIDResolver 返回 report_server 默认使用的 tableId resolver。
func DefaultTableIDResolver() *TableIDResolver {
	return defaultTableIDResolver
}

// Resolve 根据 appid/appkey 解析 tableId，并优先使用本地缓存。
//
// 数据流说明：
// 1. 先把 `appid + appkey` 通过统一 key builder 拼成 cacheKey。
// 2. 先查本地 cache，如果命中，直接返回。
// 3. 本地未命中时，再查 Redis 真值。
// 4. 查到后回填本地缓存，供下次命中。
//
// 示例：
// 1. 第一次 Resolve("1001", "demo") -> Redis 返回 "51" -> 本地缓存写入 "1001_xwl_demo" => "51"
// 2. 第二次 Resolve("1001", "demo") -> 直接返回本地缓存 "51"
func (r *TableIDResolver) Resolve(appid, appkey string) (string, error) {
	key := myapp.BuildAppTableIDKey(appid, appkey)

	// 第一步先查本地旁路缓存。
	//
	// 这样做的目的不是“替代 Redis”，而是让频繁重复访问的 appid/appkey
	// 不必每次都经过一次网络 hop。
	if cached, ok := r.cache.Load(key); ok {
		return cached.(string), nil
	}

	// 第二步查 Redis 真值。
	//
	// 这里的 fetch 是唯一可信来源：
	// 1. 本地缓存只是副本。
	// 2. 只有 Redis 里存在映射，才认为 tableId 合法。
	tableID, err := r.fetch(key)
	if err != nil {
		if err == redis.ErrNil {
			return "", my_error.NewBusiness(ERROR_TABLE, AppParmasErr)
		}
		logs.Logger.Error(
			"Resolve table id failed",
			zap.String("appid", appid),
			zap.String("appkey", appkey),
			zap.String("cache_key", key),
			zap.Error(err),
		)
		return "", my_error.NewBusiness(ERROR_TABLE, ServerErr)
	}

	// 第三步把 Redis 命中的值回填到本地旁路缓存。
	//
	// 示例：
	// 1. Redis 返回 key=1001_xwl_demo, value=51
	// 2. 本地缓存会同步保存这条映射
	r.cache.Store(key, tableID)
	return tableID, nil
}

// Invalidate 根据 appid/appkey 定向删除本地旁路缓存中的一个 key。
//
// 示例：
// 1. Invalidate("1001", "demo")
// 2. 实际删除的是本地缓存中的 "1001_xwl_demo"
func (r *TableIDResolver) Invalidate(appid, appkey string) {
	r.InvalidateByKey(myapp.BuildAppTableIDKey(appid, appkey))
}

// InvalidateByKey 直接按 cacheKey 删除本地缓存。
//
// 这个方法主要给旁路通知使用：
// 1. myapp 成功写入 Redis 后，不知道 resolver 内部结构。
// 2. 它只需要把发生变更的 cacheKey 回传过来。
// 3. resolver 自己负责删除对应本地缓存项。
func (r *TableIDResolver) InvalidateByKey(cacheKey string) {
	r.cache.Delete(cacheKey)
}

// Clear 仅用于测试或排障时手工清空所有本地缓存。
//
// 业务主流程不应该依赖这个函数做周期性全量失效。
func (r *TableIDResolver) Clear() {
	r.cache.Range(func(key, value interface{}) bool {
		r.cache.Delete(key)
		return true
	})
}
