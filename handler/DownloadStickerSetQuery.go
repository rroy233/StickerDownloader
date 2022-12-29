package handler

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rroy233/logger"
	"github.com/rroy233/tg-stickers-dl/config"
	"github.com/rroy233/tg-stickers-dl/languages"
	"github.com/rroy233/tg-stickers-dl/utils"
	"os"
	"sync/atomic"
	"time"
)

const MB = 1 << 20

type downloadTask struct {
	finished   int32
	failed     int32
	total      int32
	folderName string
}

func DownloadStickerSetQuery(update tgbotapi.Update) {
	userInfo := utils.GetLogPrefixCallbackQuery(&update)

	stickerSet, err := bot.GetStickerSet(tgbotapi.GetStickerSetConfig{
		Name: update.CallbackQuery.Message.ReplyToMessage.Sticker.SetName,
	})
	if err != nil {
		logger.Error.Println(userInfo+"DownloadStickerSetQuery-failed to GetStickerSet:", err)
		utils.CallBackWithAlert(update.CallbackQuery.ID, languages.Get().BotMsg.ErrFailedToDownload)
	}

	utils.CallBack(update.CallbackQuery.ID, "ok")

	oMsg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, languages.Get().BotMsg.Processing)
	oMsg.ReplyToMessageID = update.CallbackQuery.Message.MessageID
	msg, err := bot.Send(oMsg)
	if err != nil {
		logger.Error.Println(userInfo+"DownloadStickerSetQuery-failed to send <processing> msg:", err)
		utils.SendPlainText(&update, languages.Get().BotMsg.ErrSysFailureOccurred)
		return
	}

	//create temp folder
	folderName := fmt.Sprintf("./storage/tmp/stickers_%s_%d", stickerSet.Name, time.Now().UnixMicro())
	err = os.Mkdir(folderName, 0777)
	if err != nil || utils.IsExist(folderName) == false {
		logger.Error.Println(userInfo+"DownloadStickerSetQuery-create folder failed:", err)
		utils.EditMsgText(update.CallbackQuery.Message.Chat.ID, msg.MessageID, languages.Get().BotMsg.ErrFailed+"-1001")
		return
	}
	//delete temp folder
	defer func() {
		err = os.RemoveAll(folderName)
		if err != nil {
			logger.Error.Println(userInfo+"DownloadStickerSetQuery-delete temp folder failed:", folderName, err)
		}
	}()

	cancelCtx, cancel := context.WithCancel(context.Background())
	queue := make(chan tgbotapi.Sticker, 10)
	task := new(downloadTask)
	for i := 0; i < config.Get().General.DownloadWorkerNum; i++ {
		go downloadWorker(cancelCtx, queue, task)
	}
	task.total = int32(len(stickerSet.Stickers))
	task.folderName = folderName
	timeStart := time.Now()
	go func() {
		for _, sticker := range stickerSet.Stickers {
			queue <- sticker
		}
	}()
	go func() {
		for true {
			select {
			case <-cancelCtx.Done():
				return
			default:
				//update realtime progress
				utils.EditMsgText(update.CallbackQuery.Message.Chat.ID, msg.MessageID, fmt.Sprintf(languages.Get().BotMsg.DownloadingWithProgress, task.finished+task.failed, task.total))
				time.Sleep(1 * time.Second)
			}
		}
	}()

	//wait
	success := true
	for {
		if time.Now().Sub(timeStart).Seconds() > 60 { //60s timeout
			success = false
			break
		}
		if task.finished+task.failed == task.total {
			break
		}
		logger.Debug.Println(userInfo+"DownloadStickerSetQuery-pending-used", time.Now().Sub(timeStart).Seconds())
		time.Sleep(1 * time.Second)
	}
	cancel()
	if success {
		zipFilePath := fmt.Sprintf("./storage/tmp/%s_%d.zip", stickerSet.Name, time.Now().UnixMicro())
		err = utils.Compress(folderName, zipFilePath)
		if err != nil {
			logger.Error.Println(userInfo+"DownloadStickerSetQuery-failed to compress files:", err)
			utils.EditMsgText(update.CallbackQuery.Message.Chat.ID, msg.MessageID, languages.Get().BotMsg.ErrFailed+"-1002")
			return
		}
		utils.EditMsgText(update.CallbackQuery.Message.Chat.ID, msg.MessageID, fmt.Sprintf(languages.Get().BotMsg.ConvertedWaitingUpload, task.finished, task.failed))

		//delete
		defer utils.RemoveFile(zipFilePath)

		fileStat, err := os.Stat(zipFilePath)
		if err != nil {
			logger.Error.Println(userInfo+"DownloadStickerSetQuery-failed to read zip file info:", err)
			utils.EditMsgText(update.CallbackQuery.Message.Chat.ID, msg.MessageID, languages.Get().BotMsg.ErrFailed+"-1003")
			return
		}
		if fileStat.Size() > 50*MB {
			logger.Info.Println(userInfo + "DownloadStickerSetQuery-uploading(third party)")
			uploadTask := utils.NewUploadFile(zipFilePath)
			err = uploadTask.Upload2FileHost()
			if err != nil {
				logger.Error.Println(userInfo+"DownloadStickerSetQuery-failed to upload:", err)
				utils.EditMsgText(update.CallbackQuery.Message.Chat.ID, msg.MessageID, languages.Get().BotMsg.ErrUploadFailed)
				return
			}
			text := fmt.Sprintf(languages.Get().BotMsg.UploadedThirdParty, stickerSet.Name, uploadTask.InfoRes.Data.File.Metadata.Size.Readable, uploadTask.InfoRes.Data.File.Url.Short)
			utils.EditMsgText(
				update.CallbackQuery.Message.Chat.ID, msg.MessageID,
				text,
				utils.EntityBold(text, stickerSet.Name),
				utils.EntityBold(text, uploadTask.InfoRes.Data.File.Metadata.Size.Readable),
			)
			logger.Info.Println(userInfo + "DownloadStickerSetQuery-upload (third party) successfully！！！")
		} else {
			logger.Info.Println(userInfo + "DownloadStickerSetQuery-uploading(Telegram)")
			err = utils.SendFile(&update, zipFilePath)
			if err != nil {
				logger.Error.Println(userInfo+"DownloadStickerSetQuery-failed to upload:", err)
				utils.EditMsgText(update.CallbackQuery.Message.Chat.ID, msg.MessageID, languages.Get().BotMsg.ErrUploadFailed)
				return
			}
			text := fmt.Sprintf(languages.Get().BotMsg.UploadedTelegram, stickerSet.Name, fileStat.Size()>>20)
			utils.EditMsgText(
				update.CallbackQuery.Message.Chat.ID,
				msg.MessageID,
				text,
				utils.EntityBold(text, stickerSet.Name),
				utils.EntityBold(text, fmt.Sprintf("%d", fileStat.Size()>>20)),
			)
			logger.Info.Println(userInfo + "DownloadStickerSetQuery-upload(Telegram) successfully！！！")
		}

	} else {
		utils.EditMsgText(update.CallbackQuery.Message.Chat.ID, msg.MessageID, fmt.Sprintf(languages.Get().BotMsg.ErrTimeout))
	}

	return
}

func downloadWorker(ctx context.Context, queue chan tgbotapi.Sticker, task *downloadTask) {
	var sticker tgbotapi.Sticker
	for {
		select {
		case <-ctx.Done():
			return
		case sticker = <-queue:
			i := task.finished + task.failed
			sum := task.total
			remoteFile, err := bot.GetFile(tgbotapi.FileConfig{
				FileID: sticker.FileID,
			})
			if err != nil {
				logger.Error.Printf("DownloadStickerSetQuery[%d/%d]-failed to get file:%s", i, sum, err.Error())
				atomic.AddInt32(&task.failed, 1)
				continue
			}

			tempFilePath, err := utils.DownloadFile(remoteFile.Link(config.Get().General.BotToken))
			if err != nil {
				logger.Error.Printf("DownloadStickerSetQuery[%d/%d]-failed to download:%s", i, sum, err.Error())
				atomic.AddInt32(&task.failed, 1)
				continue
			}
			logger.Info.Printf("DownloadStickerSetQuery[%d/%d]-temp file downloaded：%s\n", i, sum, tempFilePath)

			outFilePath := fmt.Sprintf("%s/%d.gif", task.folderName, i)
			err = utils.ConvertToGif(tempFilePath, outFilePath)
			utils.RemoveFile(tempFilePath)
			if err != nil {
				logger.Error.Printf("DownloadStickerSetQuery[%d/%d]-failed to convert：%s\n", i, sum, err.Error())
				atomic.AddInt32(&task.failed, 1)
				continue
			}
			atomic.AddInt32(&task.finished, 1)
		}
	}
}
