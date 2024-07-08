package router

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rroy233/StickerDownloader/config"
	"github.com/rroy233/StickerDownloader/db"
	"github.com/rroy233/StickerDownloader/handler"
	"github.com/rroy233/StickerDownloader/languages"
	"github.com/rroy233/StickerDownloader/statistics"
	"github.com/rroy233/StickerDownloader/utils"
	"gopkg.in/rroy233/logger.v2"
	"runtime/debug"
	"strings"
	"time"
)

const (
	rateLimitShort = 2 * time.Second
	rateLimitLong  = 10 * time.Second
)

func Handle(update tgbotapi.Update) {
	//statistics
	statistics.Statistics.Record("MsgHandleTotalTimes", 1)
	statistics.Statistics.RecordUser(utils.MD5Short(fmt.Sprintf("%d", utils.GetUID(&update))))

	logger.Info.Println(utils.LogUserInfo(&update) + utils.JsonEncode(update))

	//recover
	defer func() {
		if r := recover(); r != nil {
			logger.Error.Printf("[APP CRUSHED]%v %s", r, string(debug.Stack()))
			utils.SendPlainText(&update, languages.Get(&update).BotMsg.ErrSysFailureOccurred)
		}
	}()

	//auto leave channel
	if update.ChannelPost != nil || update.EditedChannelPost != nil {
		handler.AutoLeave(update)
		return
	}
	//auto leave group
	if update.Message != nil && update.Message.LeftChatMember != nil && (update.Message.Chat.Type == "group" || update.Message.Chat.Type == "supergroup") {
		handler.AutoLeave(update)
		return
	}
	if update.MyChatMember != nil && (update.MyChatMember.Chat.Type == "channel" || update.MyChatMember.Chat.Type == "group" || update.MyChatMember.Chat.Type == "supergroup") {
		handler.AutoLeave(update)
		return
	}

	//command
	if update.Message != nil && update.Message.IsCommand() {
		logger.Info.Println(update.Message.Command())
		switch update.Message.Command() {
		case "start":
			handler.StartCommand(update)
		case "help":
			handler.HelpCommand(update)
		case "getlimit":
			handler.GetLimitCommand(update)
		case "invite":
			handler.InviteCommand(update)
		case "admin": //admin
			handler.AdminCommand(update)
		case "reload": //admin
			handler.ReloadConfigCommand(update)
		case "clearcache": //admin
			handler.ClearCacheCommand(update)
		case "statistics":
			handler.StatisticsCommand(update)
		default:
			return
		}
		statistics.Statistics.RecordCommand(update.Message.Command())
	}

	//add stickers url message
	// e.g. https://t.me/addstickers/xxx
	if update.Message != nil && strings.HasPrefix(update.Message.Text, "https://t.me/addstickers/") == true {
		handler.AddStickerUrlMessage(update)
		statistics.Statistics.Record("MsgStickerUrl", 1)
	}

	//Sticker message
	if update.Message != nil && update.Message.Sticker != nil {
		if db.CheckLimit(&update) == true {
			utils.SendPlainText(&update, fmt.Sprintf(languages.Get(&update).BotMsg.ErrReachLimit, config.Get().General.UserDailyLimit))
			return
		}
		//访问频率控制
		if limitLast := db.CheckUserRateLimit(utils.GetUID(&update), rateLimitShort); limitLast != -1 {
			utils.SendPlainText(&update, languages.Get(&update).BotMsg.ErrRateReachLimit)
			return
		}
		handler.StickerMessage(update)
		statistics.Statistics.Record("MsgStickerNum", 1)
	}

	//Animation message
	if update.Message != nil && update.Message.Animation != nil {
		if db.CheckLimit(&update) == true {
			utils.SendPlainText(&update, fmt.Sprintf(languages.Get(&update).BotMsg.ErrReachLimit, config.Get().General.UserDailyLimit))
			return
		}
		//访问频率控制
		if limitLast := db.CheckUserRateLimit(utils.GetUID(&update), rateLimitShort); limitLast != -1 {
			utils.SendPlainText(&update, languages.Get(&update).BotMsg.ErrRateReachLimit)
			return
		}
		handler.AnimationMessage(update)
		statistics.Statistics.Record("MsgAnimationNum", 1)
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
			//访问频率控制
			if limitLast := db.CheckUserRateLimit(utils.GetUID(&update), rateLimitLong); limitLast != -1 {
				utils.CallBackWithAlert(update.CallbackQuery.ID, languages.Get(&update).BotMsg.ErrRateReachLimit)
				return
			}
			handler.DownloadStickerSetQuery(update)
			statistics.Statistics.Record("MsgStickerSet", 1)
		case strings.HasPrefix(data, handler.QuitQueueCallbackQueryPrefix) == true:
			handler.QuitQueueQuery(update)
		}
	}
	return
}
