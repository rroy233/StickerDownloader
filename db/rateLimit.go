package db

import (
	"fmt"
	"time"
)

// CheckUserRateLimit 检查用户访问频率
//
// 传入UID和最小允许访问间隔minInterval
//
// 返回int需要等待的时间(s)。若返回-1则无需等待即可放行，同时设置新的访问间隔minInterval
func CheckUserRateLimit(UID int64, minInterval time.Duration) int {
	key := fmt.Sprintf("%s:User_%d:RateLimit", ServicePrefix, UID)

	data := rdb.Get(ctx, key).Val()
	if data == "" {
		rdb.Set(ctx, key, time.Now().Unix(), minInterval)
		return -1
	}

	return int(rdb.TTL(ctx, key).Val() / time.Second)

}
