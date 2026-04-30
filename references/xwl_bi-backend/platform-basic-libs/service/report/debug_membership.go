package report

import (
	"github.com/1340691923/xwl_bi/engine/db"
	"github.com/1340691923/xwl_bi/engine/logs"
	"github.com/1340691923/xwl_bi/platform-basic-libs/util"
	"github.com/garyburd/redigo/redis"
	"go.uber.org/zap"
)

type debugSetMembershipQuery func(key, distinctID string) (bool, error)

// DebugMembershipChecker 负责判断某个 distinct_id 是否属于 debug 设备集合。
//
// 旧实现里 IsDebugUser 和 CantInflowOfKakfa 完全重复。
// 这里把“是否命中 debug 设备集合”收口成单一职责点，避免重复 Redis 访问和重复日志分支。
type DebugMembershipChecker struct {
	query debugSetMembershipQuery
}

func NewDebugMembershipChecker(query debugSetMembershipQuery) *DebugMembershipChecker {
	if query == nil {
		query = func(key, distinctID string) (bool, error) {
			conn := db.RedisPool.Get()
			defer conn.Close()

			hit, err := redis.Int(conn.Do("SISMEMBER", key, distinctID))
			if err != nil {
				return false, err
			}
			return hit > 0, nil
		}
	}

	return &DebugMembershipChecker{query: query}
}

// IsDebugDevice 判断是否需要进入 debug 检查分支。
//
// 示例：
// 1. debug=0 -> 直接返回 false
// 2. debug=1 且 distinct_id 在 Redis 集合内 -> 返回 true
func (c *DebugMembershipChecker) IsDebugDevice(debug, distinctID, tableID string) bool {
	if !util.InstrArr([]string{DebugToDB, DebugNotToDB}, debug) {
		return false
	}

	key := BuildDebugDeviceIDKey(tableID)
	hit, err := c.query(key, distinctID)
	if util.FilterRedisNilErr(err) {
		logs.Logger.Error(
			"debug device lookup failed",
			zap.String("table_id", tableID),
			zap.String("distinct_id", distinctID),
			zap.String("redis_key", key),
			zap.Error(err),
		)
		return false
	}
	return hit
}
