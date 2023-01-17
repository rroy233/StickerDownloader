package handler

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rroy233/logger"
	"github.com/rroy233/tg-stickers-dl/config"
	"github.com/rroy233/tg-stickers-dl/languages"
	"github.com/rroy233/tg-stickers-dl/utils"
)

var addStickersUrlPrefix = "https://t.me/addstickers/"

func AddStickerUrlMessage(update tgbotapi.Update) {
	userInfo := utils.GetLogPrefixMessage(&update) + "[AddStickerUrlMessage]"

	logger.Info.Println(userInfo + "Get text:" + update.Message.Text)

	if len(update.Message.Text) == len(addStickersUrlPrefix) {
		utils.SendPlainText(&update, languages.Get(&update).BotMsg.ErrFailedToDownload)
		return
	}

	stickerSet, err := bot.GetStickerSet(tgbotapi.GetStickerSetConfig{
		Name: update.Message.Text[len(addStickersUrlPrefix):],
	})
	if err != nil {
		logger.Error.Println(userInfo+"failed to GetStickerSet:", err)
		utils.SendPlainText(&update, languages.Get(&update).BotMsg.ErrFailedToDownload)
		return
	}

	//len equal to 0
	if len(stickerSet.Stickers) == 0 {
		logger.Error.Println(userInfo+"len(stickerSet.Stickers) == 0", utils.JsonEncode(stickerSet))
		utils.SendPlainText(&update, languages.Get(&update).BotMsg.ErrFailedToDownload)
		return
	}

	//try to download one
	remoteFile, err := bot.GetFile(tgbotapi.FileConfig{
		FileID: stickerSet.Stickers[0].FileID,
	})
	if err != nil {
		logger.Error.Println(userInfo+"failed to get file:", err)
	}
	tempFilePath, err := utils.DownloadFile(remoteFile.Link(config.Get().General.BotToken))
	if err != nil {
		logger.Error.Println(userInfo+"failed to download file:", err)
	}
	defer utils.RemoveFile(tempFilePath) //delete temp file
	//check file type
	if utils.GetFileExtName(tempFilePath) != "webp" && utils.GetFileExtName(tempFilePath) != "webm" {
		utils.SendPlainText(&update, languages.Get(&update).BotMsg.ErrStickerNotSupport)
		return
	}

	StickerMsg, err := bot.Send(tgbotapi.NewSticker(update.Message.Chat.ID, tgbotapi.FileID(stickerSet.Stickers[0].FileID)))
	if err != nil {
		logger.Error.Println(userInfo+"bot.Send error", err)
		utils.SendPlainText(&update, languages.Get(&update).BotMsg.ErrSysFailureOccurred)
		return
	}

	msgTpl := tgbotapi.NewMessage(update.Message.Chat.ID, languages.Get(&update).BotMsg.Processing)
	msgTpl.ReplyToMessageID = StickerMsg.MessageID
	replyMsg, err := bot.Send(msgTpl)
	if err != nil {
		logger.Error.Println(userInfo+"bot.Send error", err)
		utils.SendPlainText(&update, languages.Get(&update).BotMsg.ErrSysFailureOccurred)
		return
	}

	text := fmt.Sprintf(languages.Get(&update).BotMsg.StickersSetInfoFromUrl, stickerSet.Name, len(stickerSet.Stickers))
	_, err = bot.Request(tgbotapi.NewEditMessageTextAndMarkup(update.Message.Chat.ID, replyMsg.MessageID, text, tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(languages.Get(&update).BotMsg.DownloadStickerSet, DownloadStickerSetCallbackQuery)),
	)))
	if err != nil {
		logger.Error.Println(userInfo+"bot.Send error", err)
		utils.SendPlainText(&update, languages.Get(&update).BotMsg.ErrSysFailureOccurred)
		return
	}
	return
}
