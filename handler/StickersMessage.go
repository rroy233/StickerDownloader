package handler

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rroy233/logger"
	"github.com/rroy233/tg-stickers-dl/config"
	"github.com/rroy233/tg-stickers-dl/utils"
	"time"
)

func StickerMessage(update tgbotapi.Update) {

	oMsg := tgbotapi.NewMessage(update.Message.Chat.ID, "处理中...")
	oMsg.ReplyToMessageID = update.Message.MessageID
	msg, err := bot.Send(oMsg)
	if err != nil {
		logger.Error.Println("发送消息失败:", err)
		return
	}

	remoteFile, err := bot.GetFile(tgbotapi.FileConfig{
		FileID: update.Message.Sticker.FileID,
	})
	if err != nil {
		logger.Error.Println("获取文件失败:", err)
	}

	tempFilePath, err := utils.DownloadFile(remoteFile.Link(config.Get().General.BotToken))
	if err != nil {
		logger.Error.Println("下载文件失败:", err)
	}
	logger.Info.Println("已下载临时文件：", tempFilePath)

	//删除临时文件
	defer utils.RemoveFile(tempFilePath)

	//判断文件格式
	if utils.GetFileExtName(tempFilePath) != "webp" && utils.GetFileExtName(tempFilePath) != "webm" {
		utils.EditMsgText(update.Message.Chat.ID, msg.MessageID, "该表情不支持下载")
		return
	}

	outPath := fmt.Sprintf("./storage/tmp/convert_%d.gif", time.Now().UnixMicro())
	err = utils.Mp4ToGif(tempFilePath, outPath)
	if err != nil {
		logger.Error.Println("转换文件失败:", err)
		utils.EditMsgText(update.Message.Chat.ID, msg.MessageID, "转换文件失败")
		return
	}

	err = utils.SendFile(&update, outPath)
	if err != nil {
		logger.Error.Println("SendFile失败:", err)
		utils.EditMsgText(update.Message.Chat.ID, msg.MessageID, "发送文件失败")
		return
	}

	_, err = bot.Request(tgbotapi.NewEditMessageTextAndMarkup(update.Message.Chat.ID, msg.MessageID, "已完成转换！", tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("下载整套表情包", "DOWNLOAD_STICKERS_SET")),
	)))
	if err != nil {
		logger.Error.Println("删除消息失败:", err)
	}

	utils.RemoveFile(outPath)
	return
}
