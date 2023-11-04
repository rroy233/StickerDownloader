package db

import (
	_ "github.com/go-sql-driver/mysql"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type StickerItem struct {
	//telegram sticker type
	Info tgbotapi.Sticker `json:"info"`

	//path of local-cached file
	SavePath string `json:"save_path"`

	//file_id of converted sticker file
	ConvertedFileID string `json:"converted_file_id"`

	//time of caching
	SaveTimeStamp int64 `json:"save_time_stamp"`

	//extension of local-cached file
	FileExt string `json:"file_ext"`

	//md5 of local-cached file
	MD5 string `json:"md5"`

	//size of local-cached file
	Size int64 `json:"size"`
}
