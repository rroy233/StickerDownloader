package router

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rroy233/logger"
	"github.com/rroy233/tg-stickers-dl/config"
	"github.com/rroy233/tg-stickers-dl/db"
	"github.com/rroy233/tg-stickers-dl/handler"
	"github.com/rroy233/tg-stickers-dl/languages"
	"github.com/rroy233/tg-stickers-dl/utils"
)

func Handle(update tgbotapi.Update) {
	logger.Info.Println(utils.LogUserInfo(&update) + utils.JsonEncode(update))

	//command
	if update.Message != nil && update.Message.IsCommand() {
		logger.Info.Println(update.Message.Command())
		switch update.Message.Command() {
		case "start":
			handler.StartCommand(update)
		case "help":
			handler.HelpCommand(update)
		case "reload":
			handler.ReloadConfigCommand(update)
		case "getlimit":
			handler.GetLimitCommand(update)
		}
	}

	//Sticker message
	if update.Message != nil && update.Message.Sticker != nil {
		if db.CheckLimit(update.Message.From.ID) == true {
			utils.SendPlainText(&update, fmt.Sprintf(languages.Get().BotMsg.ErrReachLimit, config.Get().General.UserDailyLimit))
			return
		}
		handler.StickerMessage(update)
	}

	//inline query
	if update.CallbackQuery != nil {
		switch update.CallbackQuery.Data {
		case "DOWNLOAD_STICKERS_SET":
			if db.CheckLimit(update.CallbackQuery.Message.From.ID) == true {
				utils.CallBackWithAlert(update.CallbackQuery.ID, fmt.Sprintf(languages.Get().BotMsg.ErrReachLimit, config.Get().General.UserDailyLimit))
				return
			}
			handler.DownloadStickerSetQuery(update)
		}
	}
	return
}
