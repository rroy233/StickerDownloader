package db

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/rroy233/logger"
	"github.com/rroy233/tg-stickers-dl/config"
)

// var db *sqlx.DB
var rdb *redis.Client
var ctx = context.Background()

func Init() {
	//redis
	rdb = redis.NewClient(&redis.Options{
		Addr:     config.Get().Redis.Server + ":" + config.Get().Redis.Port,
		Password: config.Get().Redis.Password,
		DB:       config.Get().Redis.DB,
	})
	err := rdb.Ping(context.Background()).Err()
	if err != nil {
		logger.FATAL.Fatalln("[系统服务][异常]Redis启动失败:", err)
		return
	}
	logger.Info.Println("[系统服务][成功]Redis已连接")
	return
}
