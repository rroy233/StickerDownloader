package db

import (
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/rroy233/tg-stickers-dl/config"
	"strconv"
	"time"
)

const ServicePrefix = "StickerDl"

// CheckLimit 用户是否已达到今日限额
func CheckLimit(UID int64) bool {
	if UID == config.Get().General.AdminUID {
		return false
	}
	limit := rdb.Get(ctx, fmt.Sprintf("%s:UserLimit:%d", ServicePrefix, UID)).Val()
	if limit == "" {
		rdb.Set(ctx, fmt.Sprintf("%s:UserLimit:%d", ServicePrefix, UID), 1, 24*time.Hour)
		return false
	}

	limitTimes, _ := strconv.Atoi(limit)
	if limitTimes > config.Get().General.UserDailyLimit {
		return true
	}
	rdb.Set(ctx, fmt.Sprintf("%s:UserLimit:%d", ServicePrefix, UID), limitTimes+1, redis.KeepTTL)
	return false
}

// GetLimit 获取该用户今日可用次数
func GetLimit(UID int64) int {
	if UID == config.Get().General.AdminUID {
		return -1
	}
	limit := rdb.Get(ctx, fmt.Sprintf("%s:UserLimit:%d", ServicePrefix, UID)).Val()
	if limit == "" {
		return config.Get().General.UserDailyLimit
	}
	limitTimes, _ := strconv.Atoi(limit)
	return config.Get().General.UserDailyLimit - limitTimes
}
