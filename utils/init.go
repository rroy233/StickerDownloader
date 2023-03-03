package utils

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rroy233/StickerDownloader/languages"
	"github.com/rroy233/logger"
	"go.uber.org/ratelimit"
	"os"
	"time"
)

var bot *tgbotapi.BotAPI
var loggerPrefix = "[utils]"
var Limiter ratelimit.Limiter

func Init(api *tgbotapi.BotAPI) {
	//初始化rate Limiter
	//TG官方对message发送频率有限制
	//详见:https://core.telegram.org/bots/faq#broadcasting-to-users
	Limiter = ratelimit.New(30, ratelimit.Per(1*time.Second))

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
		logger.FATAL.Printf(languages.Get(nil).System.FfmpegNotExist, getFfmpeg())
	}
	return
}
