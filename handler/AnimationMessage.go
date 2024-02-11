package handler

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rroy233/StickerDownloader/config"
	"github.com/rroy233/StickerDownloader/db"
	"github.com/rroy233/StickerDownloader/languages"
	"github.com/rroy233/StickerDownloader/utils"
	"gopkg.in/rroy233/logger.v2"
	"time"
)

func AnimationMessage(update tgbotapi.Update) {
	userInfo := utils.GetLogPrefixMessage(&update)

	oMsg := tgbotapi.NewMessage(update.Message.Chat.ID, languages.Get(&update).BotMsg.Processing)
	oMsg.ReplyToMessageID = update.Message.MessageID
	msg, err := utils.BotSend(oMsg)
	if err != nil {
		logger.Error.Println(userInfo+"failed to send msg:", err)
		return
	}

	//Enqueue
	qItem, quit := enqueue(&update, &msg)
	if quit == true {
		return
	}
	//Enqueue

	remoteFile, err := utils.BotGetFile(tgbotapi.FileConfig{
		FileID: update.Message.Animation.FileID,
	})
	if err != nil {
		logger.Error.Println(userInfo+"failed to get file:", err)
	}

	tempFilePath, err := utils.DownloadFile(remoteFile.Link(config.Get().General.BotToken))
	if err != nil {
		logger.Error.Println(userInfo+"failed to download file:", err)
	}

	logger.Info.Printf("%sGet Animation => %s", userInfo, tempFilePath)

	//delete temp file
	defer utils.RemoveFile(tempFilePath)

	//check file type
	if utils.GetFileExtName(tempFilePath) != "mp4" {
		utils.EditMsgText(update.Message.Chat.ID, msg.MessageID, languages.Get(&update).BotMsg.ErrStickerNotSupport)
		return
	}

	//path to save converted file
	outPath := fmt.Sprintf("./storage/tmp/convert_%d.gif", time.Now().UnixMicro())
	defer func() {
		utils.RemoveFile(outPath)
	}()

	//init convert task
	convertTask := utils.ConvertTask{
		InputFilePath:  tempFilePath,
		InputExtension: "mp4",
		OutputFilePath: outPath,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	err = convertTask.Run(ctx)
	cancel()
	if err != nil {
		logger.Error.Println(userInfo+"failed to convert:", err)
		utils.EditMsgText(update.Message.Chat.ID, msg.MessageID, languages.Get(&update).BotMsg.ErrConvertFailed)
		return
	}

	//Dequeue
	dequeue(qItem)
	//Dequeue

	_, err = utils.SendFileByPath(&update, outPath)
	if err != nil {
		logger.Error.Println(userInfo+"failed to SendFile:", err)
		utils.EditMsgText(
			update.Message.Chat.ID,
			msg.MessageID,
			fmt.Sprintf("%s(TelegramAPI:%s)", languages.Get(&update).BotMsg.ErrSendFileFailed, err.Error()),
		)
		return
	}

	//Consume the current user's daily limit
	if err = db.ConsumeLimit(&update); err != nil {
		logger.Error.Println(userInfo + err.Error())
	}

	utils.EditMsgText(update.Message.Chat.ID, msg.MessageID, languages.Get(&update).BotMsg.ConvertCompleted)
	if err != nil {
		logger.Error.Println(userInfo+"failed to delete msg:", err)
	}

	return
}
