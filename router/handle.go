package router

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rroy233/logger"
	"github.com/rroy233/tg-stickers-dl/config"
	"github.com/rroy233/tg-stickers-dl/db"
	"github.com/rroy233/tg-stickers-dl/handler"
	"github.com/rroy233/tg-stickers-dl/utils"
)

func Handle(update tgbotapi.Update) {
	logger.Info.Println(utils.JsonEncode(update))

	//command
	if update.Message != nil && update.Message.IsCommand() {
		logger.Info.Println(update.Message.Command())
		switch update.Message.Command() {
		case "help":
			handler.HelpCommand(update)
		case "getlimit":
			handler.GetLimitCommand(update)
		}
	}

	//Sticker message
	if update.Message != nil && update.Message.Sticker != nil {
		//判断是否达到限制
		if db.CheckLimit(update.Message.From.ID) == true {
			utils.SendPlainText(&update, fmt.Sprintf("普通用户每24h限制使用%d次，您已达到今日限制！！", config.Get().General.UserDailyLimit))
			return
		}
		handler.StickerMessage(update)
	}

	//inline query
	if update.CallbackQuery != nil {
		switch update.CallbackQuery.Data {
		case "DOWNLOAD_STICKERS_SET":
			//判断是否达到限制
			if db.CheckLimit(update.CallbackQuery.Message.From.ID) == true {
				utils.CallBackWithAlert(update.CallbackQuery.ID, fmt.Sprintf("普通用户每24h限制使用%d次，您已达到今日限制！！", config.Get().General.UserDailyLimit))
				return
			}
			handler.DownloadStickerSetQuery(update)
		}
	}
	return
}
