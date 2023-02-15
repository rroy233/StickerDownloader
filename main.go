package main

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rroy233/logger"
	"github.com/rroy233/tg-stickers-dl/config"
	"github.com/rroy233/tg-stickers-dl/db"
	"github.com/rroy233/tg-stickers-dl/handler"
	"github.com/rroy233/tg-stickers-dl/languages"
	"github.com/rroy233/tg-stickers-dl/router"
	"github.com/rroy233/tg-stickers-dl/utils"
	"os"
	"os/signal"
)

var bot *tgbotapi.BotAPI
var cancel context.CancelFunc
var stopCtx context.Context
var cancelCh chan int

func main() {

	//config
	config.Init()

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
			NotUseJson: true,
		})

	//language
	languages.Init()

	var err error
	bot, err = tgbotapi.NewBotAPI(config.Get().General.BotToken)
	if err != nil {
		logger.FATAL.Fatalln(err.Error())
	}

	//init
	utils.Init(bot)
	handler.Init(bot)
	config.Init()
	db.Init()

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
	signal.Notify(sigCh, os.Interrupt, os.Kill)
	<-sigCh

	Stop()
	logger.Info.Println(languages.Get(nil).System.StopRunning)

}

func Stop() {
	cancel()
	waitForDone(cancelCh)

	//clean temp files
	utils.CleanTmp()

	//close db
	db.Close()
}

func worker(stopCtx context.Context, uc tgbotapi.UpdatesChannel, cancelCh chan int) {
	for {
		select {
		case update := <-uc:
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
