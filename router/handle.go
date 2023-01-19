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
	"strings"
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
		case "invite":
			handler.InviteCommand(update)
		}
	}

	//add stickers url message
	// e.g. https://t.me/addstickers/xxx
	if update.Message != nil && strings.HasPrefix(update.Message.Text, "https://t.me/addstickers/") == true {
		handler.AddStickerUrlMessage(update)
	}

	//Sticker message
	if update.Message != nil && update.Message.Sticker != nil {
		if db.CheckLimit(&update) == true {
			utils.SendPlainText(&update, fmt.Sprintf(languages.Get(&update).BotMsg.ErrReachLimit, config.Get().General.UserDailyLimit))
			return
		}
		handler.StickerMessage(update)
	}

	//Animation message
	if update.Message != nil && update.Message.Animation != nil {
		if db.CheckLimit(&update) == true {
			utils.SendPlainText(&update, fmt.Sprintf(languages.Get(&update).BotMsg.ErrReachLimit, config.Get().General.UserDailyLimit))
			return
		}
		handler.AnimationMessage(update)
	}

	//inline query
	if update.CallbackQuery != nil {
		data := update.CallbackQuery.Data
		switch {
		case data == handler.DownloadStickerSetCallbackQuery:
			if db.CheckLimit(&update) == true {
				utils.CallBackWithAlert(update.CallbackQuery.ID, fmt.Sprintf(languages.Get(&update).BotMsg.ErrReachLimit, config.Get().General.UserDailyLimit))
				return
			}
			handler.DownloadStickerSetQuery(update)
		case strings.HasPrefix(data, handler.QuitQueueCallbackQueryPrefix) == true:
			handler.QuitQueueQuery(update)
		}
	}
	return
}
