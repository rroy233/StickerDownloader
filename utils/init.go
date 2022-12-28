package utils

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rroy233/logger"
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
		logger.FATAL.Println(getFfmpeg() + "不存在！！请到官网下载可执行文件并重命名放入./ffmpeg文件夹")
	}
	return
}
