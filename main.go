package main

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rroy233/StickerDownloader/config"
	"github.com/rroy233/StickerDownloader/db"
	"github.com/rroy233/StickerDownloader/handler"
	"github.com/rroy233/StickerDownloader/languages"
	"github.com/rroy233/StickerDownloader/router"
	"github.com/rroy233/StickerDownloader/statistics"
	"github.com/rroy233/StickerDownloader/utils"
	"gopkg.in/rroy233/logger.v2"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var bot *tgbotapi.BotAPI
var cancel context.CancelFunc
var stopCtx context.Context
var cancelCh chan int

func main() {
	//config
	config.Init()
	log.Println("[main]config=" + utils.JsonEncode(config.Get()))

	//logger
	logger.New(
		&logger.Config{
			StdOutput:      true,
			StoreLocalFile: true,
			StoreRemote:    config.Get().Logger.Report,
			RemoteConfig: logger.RemoteConfigStruct{
				RequestUrl: config.Get().Logger.ReportUrl,
				QueryKey:   config.Get().Logger.ReportQueryKey,
			},
		})

	//language
	languages.Init()

	var err error
	bot, err = tgbotapi.NewBotAPI(config.Get().General.BotToken)
	if err != nil {
		logger.FATAL.Fatalln(err.Error())
	}

	//init
	time.Local = time.FixedZone("CST", 8*3600)
	config.Init()
	rdb := db.Init()
	statistics.InitStatistic(rdb)
	utils.Init(bot)
	handler.Init(bot)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)
	stopCtx, cancel = context.WithCancel(context.Background())
	cancelCh = make(chan int, config.Get().General.WorkerNum)
	for i := 0; i < config.Get().General.WorkerNum; i++ {
		go worker(stopCtx, updates, cancelCh)
	}

	logger.Info.Println(languages.Get(nil).System.Running)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGKILL, syscall.SIGTERM)
	<-sigCh

	Stop()
	logger.Info.Println(languages.Get(nil).System.StopRunning)

}

func Stop() {
	cancel()
	waitForDone(cancelCh)

	//clean temp files
	utils.CleanTmp()

	//store statistics data into redis
	statistics.Statistics.Save()

	//close db
	db.Close()
}

func worker(stopCtx context.Context, uc tgbotapi.UpdatesChannel, cancelCh chan int) {
	for {
		select {
		case update := <-uc:
			utils.Limiter.Take()
			go router.Handle(update)
		case <-stopCtx.Done():
			cancelCh <- 1
			return
		}
	}
}

func waitForDone(cancelCh chan int) {
	num := 0
	for {
		if num == config.Get().General.WorkerNum {
			break
		}
		<-cancelCh
		num++
	}
}
