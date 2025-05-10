package db

import (
	"errors"
	"fmt"
	tgbotapi "github.com/OvyFlash/telegram-bot-api"
	"github.com/go-redis/redis/v8"
	"github.com/rroy233/StickerDownloader/config"
	"github.com/rroy233/StickerDownloader/utils"
	"gopkg.in/rroy233/logger.v2"
	"strconv"
	"time"
)

// CheckLimit Determines if the user has reached today's limit
func CheckLimit(update *tgbotapi.Update) bool {
	UID := int64(0)
	if update.Message != nil {
		UID = update.Message.Chat.ID
	} else if update.CallbackQuery != nil {
		UID = update.CallbackQuery.Message.Chat.ID
	} else {
		return true
	}
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
	return false
}

// ConsumeLimit Consume the current user's daily limit
func ConsumeLimit(update *tgbotapi.Update) error {
	UID := int64(0)
	if update.Message != nil {
		UID = update.Message.Chat.ID
	} else if update.CallbackQuery != nil {
		UID = update.CallbackQuery.Message.Chat.ID
	} else {
		return errors.New("failed to get uid")
	}

	limit := rdb.Get(ctx, fmt.Sprintf("%s:UserLimit:%d", ServicePrefix, UID)).Val()
	if limit == "" {
		rdb.Set(ctx, fmt.Sprintf("%s:UserLimit:%d", ServicePrefix, UID), 1, 24*time.Hour)
		return nil
	}

	limitTimes, _ := strconv.Atoi(limit)
	rdb.Set(ctx, fmt.Sprintf("%s:UserLimit:%d", ServicePrefix, UID), limitTimes+1, redis.KeepTTL)
	return nil
}

// 获取该用户已使用的次数
func getUsed(UID int64) int {
	limit := rdb.Get(ctx, fmt.Sprintf("%s:UserLimit:%d", ServicePrefix, UID)).Val()
	if limit == "" {
		return -1
	}
	limitTimes, _ := strconv.Atoi(limit)
	return limitTimes
}

// GetLimit 获取该用户今日剩余可用次数
func GetLimit(UID int64) int {
	if UID == config.Get().General.AdminUID {
		return -1
	}
	limitTimes := getUsed(UID)
	if limitTimes == -1 {
		return config.Get().General.UserDailyLimit
	}
	return config.Get().General.UserDailyLimit - limitTimes
}

// RewardDailyOnce increases user's usage limit once per day (if not already rewarded)
func RewardDailyOnce(update *tgbotapi.Update, reward int) int {
	UID := utils.GetUID(update)

	rewardKey := fmt.Sprintf("%s:DailyRewarded:%d", ServicePrefix, UID)
	rewarded, err := rdb.Get(ctx, rewardKey).Result()
	if err == nil && rewarded == "1" {
		// Already rewarded today
		return 0
	}

	// Mark as rewarded today with 24h expiry
	rdb.Set(ctx, rewardKey, "1", 24*time.Hour)

	// Increase usage limit
	limitKey := fmt.Sprintf("%s:UserLimit:%d", ServicePrefix, UID)
	limit := rdb.Get(ctx, limitKey).Val()
	if limit == "" {
		rdb.Set(ctx, limitKey, reward, 24*time.Hour)
		return 0
	}

	limitTimes, _ := strconv.Atoi(limit)
	rdb.Set(ctx, limitKey, limitTimes+reward, 24*time.Hour)

	logger.Info.Printf("%s [RewardDailyOnce] ADD [%d] -> [%d]", utils.LogUserInfo(update), reward, limitTimes+reward)

	return reward
}
