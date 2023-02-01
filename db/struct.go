package db

import (
	_ "github.com/go-sql-driver/mysql"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type stickerItem struct {
	Info          tgbotapi.Sticker `json:"info"`
	SavePath      string           `json:"save_path"`
	SaveTimeStamp int64            `json:"save_time_stamp"`
	MD5           string           `json:"md5"`
	Size          int64            `json:"size"`
}
