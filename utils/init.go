package utils

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rroy233/logger"
	"github.com/rroy233/tg-stickers-dl/languages"
	"os"
)

var bot *tgbotapi.BotAPI
var loggerPrefix = "[utils]"

func Init(api *tgbotapi.BotAPI) {
	bot = api
	initSender(3)

	//file check
	if IsExist("./storage") == false {
		_ = os.Mkdir("./storage", 0755)
	}
	if IsExist("./storage/tmp") == false {
		_ = os.Mkdir("./storage/tmp", 0755)
	}
	if IsExist("./ffmpeg") == false {
		_ = os.Mkdir("./ffmpeg", 0755)
	}
	if IsExist("./ffmpeg/"+getFfmpeg()) == false {
		logger.FATAL.Printf(languages.Get().System.FfmpegNotExist, getFfmpeg())
	}
	return
}
