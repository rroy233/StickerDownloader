package utils

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rroy233/StickerDownloader/config"
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

var ffmpegExecutablePath string
var rlottieExcutablePath string

func Init(api *tgbotapi.BotAPI) {
	//初始化rate Limiter
	//TG官方对message发送频率有限制
	//详见:https://core.telegram.org/bots/faq#broadcasting-to-users
	Limiter = ratelimit.New(30, ratelimit.Per(1*time.Second))

	bot = api
	initSender(3)

	var err error

	//folder check
	if IsExist("./storage") == false {
		err = os.Mkdir("./storage", 0755)
	}
	if IsExist("./storage/tmp") == false {
		err = os.Mkdir("./storage/tmp", 0755)
	}
	if IsExist("./ffmpeg") == false {
		err = os.Mkdir("./ffmpeg", 0755)
	}
	if config.Get().General.SupportTGSFile == true && IsExist("./lottie2gif") == false {
		err = os.Mkdir("./lottie2gif", 0755)
	}
	if err != nil {
		logger.FATAL.Println(err)
	}

	findFFmpeg()
	if config.Get().General.SupportTGSFile == true {
		findRlottie()
	}

	return
}

func findFFmpeg() {
	//find ffmpeg from system PATH
	paths := strings.Split(os.Getenv("PATH"), ":")
	for _, path := range paths {
		if IsExist(path+"/ffmpeg") == true {
			ffmpegExecutablePath = path + "/ffmpeg"
			return
		} else if IsExist(path+"/ffmpeg.exe") == true {
			ffmpegExecutablePath = path + "/ffmpeg.exe"
			return
		}
	}

	//Alerts users who have adopted older naming requirements
	if IsExist("./ffmpeg/"+getFfmpegFilename(false)) == true {
		if err := os.Rename("./ffmpeg/"+getFfmpegFilename(false), "./ffmpeg/"+getFfmpegFilename(true)); err != nil {
			logger.FATAL.Printf("Failed to rename '%s' into '%s', please rename it manually.", getFfmpegFilename(false), getFfmpegFilename(true))
		}
		logger.Info.Printf("%s has been renamed into '%s'!", getFfmpegFilename(false), getFfmpegFilename(true))
	}

	//find from StickerDownloader running folder
	if IsExist("./ffmpeg/"+getFfmpegFilename(true)) == false {
		logger.FATAL.Printf(languages.Get(nil).System.FfmpegNotExist, getFfmpegFilename(true))
	}
	ffmpegExecutablePath = "./ffmpeg/" + getFfmpegFilename(true)
}

func findRlottie() {
	//find from StickerDownloader running folder
	if IsExist("./lottie2gif/"+getRlottieFilename()) == false {
		logger.FATAL.Printf(languages.Get(nil).System.RlottieNotExist, "./lottie2gif/"+getRlottieFilename())
	}
	rlottieExcutablePath = "./lottie2gif/" + getRlottieFilename()
}
