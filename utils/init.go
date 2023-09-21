package utils

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rroy233/StickerDownloader/languages"
	"go.uber.org/ratelimit"
	"gopkg.in/rroy233/logger.v2"
	"os"
	"strings"
	"time"
)

var bot *tgbotapi.BotAPI
var loggerPrefix = "[utils]"
var Limiter ratelimit.Limiter
var isSystemFFmpegExist bool

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
	checkSystemFFmpeg()
	if isSystemFFmpegExist == false && IsExist("./ffmpeg/"+getFfmpegFilename()) == false {
		logger.FATAL.Printf(languages.Get(nil).System.FfmpegNotExist, getFfmpegFilename())
	}
	return
}

func checkSystemFFmpeg() {
	paths := strings.Split(os.Getenv("PATH"), ":")
	for _, path := range paths {
		if IsExist(path+"/ffmpeg") == true || IsExist(path+"/ffmpeg.exe") == true {
			isSystemFFmpegExist = true
			break
		}
	}
}
