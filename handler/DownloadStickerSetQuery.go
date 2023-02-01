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
const Hour = int64(3600)

type downloadTask struct {
	finished   int32
	failed     int32
	total      int32
	folderName string
}

func DownloadStickerSetQuery(update tgbotapi.Update) {
	userInfo := utils.GetLogPrefixCallbackQuery(&update)

	stickerSet, err := utils.BotGetStickerSet(tgbotapi.GetStickerSetConfig{
		Name: update.CallbackQuery.Message.ReplyToMessage.Sticker.SetName,
	})
	if err != nil {
		logger.Error.Println(userInfo+"DownloadStickerSetQuery-failed to GetStickerSet:", err)
		utils.CallBackWithAlert(update.CallbackQuery.ID, languages.Get(&update).BotMsg.ErrFailedToDownload)
	}

	//remove old msg to prevent frequent request
	if time.Now().Unix()-int64(update.CallbackQuery.Message.Date) < 48*Hour {
		utils.DeleteMsg(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID)
	} else {
		utils.EditMsgText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, update.CallbackQuery.Message.Text)
	}

	utils.CallBack(update.CallbackQuery.ID, "ok")

	oMsg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, languages.Get(&update).BotMsg.Processing)
	oMsg.ReplyToMessageID = update.CallbackQuery.Message.ReplyToMessage.MessageID
	msg, err := utils.BotSend(oMsg)
	if err != nil {
		logger.Error.Println(userInfo+"DownloadStickerSetQuery-failed to send <processing> msg:", err)
		utils.SendPlainText(&update, languages.Get(&update).BotMsg.ErrSysFailureOccurred)
		return
	}

	//Enqueue
	qItem, quit := enqueue(&update, &msg)
	if quit == true {
		return
	}
	//Enqueue

	//create temp folder
	folderPath := fmt.Sprintf("./storage/tmp/stickers_%d", time.Now().UnixMicro())
	err = os.Mkdir(folderPath, 0777)
	if err != nil || utils.IsExist(folderPath) == false {
		logger.Error.Println(userInfo+"DownloadStickerSetQuery-create folder failed:", err)
		utils.EditMsgText(update.CallbackQuery.Message.Chat.ID, msg.MessageID, languages.Get(&update).BotMsg.ErrFailed+"-1001")
		return
	}
	//delete temp folder
	defer func() {
		err = os.RemoveAll(folderPath)
		if err != nil {
			logger.Error.Println(userInfo+"DownloadStickerSetQuery-delete temp folder failed:", folderPath, err)
		}
	}()

	cancelCtx, cancel := context.WithCancel(context.Background())
	queue := make(chan tgbotapi.Sticker, 10)
	task := new(downloadTask)
	for i := 0; i < config.Get().General.DownloadWorkerNum; i++ {
		go downloadWorker(cancelCtx, queue, task)
	}
	task.total = int32(len(stickerSet.Stickers))
	task.folderName = folderPath
	timeStart := time.Now()
	go func() {
		for _, sticker := range stickerSet.Stickers {
			queue <- sticker
		}
	}()
	go func() {
		text := ""
		newText := ""
		for true {
			select {
			case <-cancelCtx.Done():
				return
			default:
				//update realtime progress
				newText = fmt.Sprintf(languages.Get(&update).BotMsg.DownloadingWithProgress, task.finished+task.failed, task.total)
				if text != newText {
					utils.EditMsgText(update.CallbackQuery.Message.Chat.ID, msg.MessageID, newText)
					text = newText
				}
				time.Sleep(1 * time.Second)
			}
		}
	}()

	//wait
	success := true
	for {
		if time.Now().Sub(timeStart).Seconds() > ProcessTimeout { //default 60s timeout
			success = false
			logger.Error.Println(userInfo+"DownloadStickerSetQuery-Task Timeout:", task)
			break
		}
		if task.finished+task.failed == task.total {
			break
		}
		time.Sleep(1 * time.Second)
	}
	cancel()
	if success == false {
		utils.EditMsgText(update.CallbackQuery.Message.Chat.ID, msg.MessageID, fmt.Sprintf(languages.Get(&update).BotMsg.ErrTimeout))
		return
	}

	//Dequeue
	dequeue(qItem)
	//Dequeue

	//start upload
	zipFilePath := fmt.Sprintf("./storage/tmp/%d.zip", time.Now().UnixMicro())
	err = utils.Compress(folderPath, zipFilePath)
	if err != nil {
		logger.Error.Println(userInfo+"DownloadStickerSetQuery-failed to compress files:", err)
		utils.EditMsgText(update.CallbackQuery.Message.Chat.ID, msg.MessageID, languages.Get(&update).BotMsg.ErrFailed+"-1002")
		return
	}
	utils.EditMsgText(update.CallbackQuery.Message.Chat.ID, msg.MessageID, fmt.Sprintf(languages.Get(&update).BotMsg.ConvertedWaitingUpload, task.finished, task.failed))

	//delete
	defer utils.RemoveFile(zipFilePath)

	fileStat, err := os.Stat(zipFilePath)
	if err != nil {
		logger.Error.Println(userInfo+"DownloadStickerSetQuery-failed to read zip file info:", err)
		utils.EditMsgText(update.CallbackQuery.Message.Chat.ID, msg.MessageID, languages.Get(&update).BotMsg.ErrFailed+"-1003")
		return
	}

	if fileStat.Size() > 50*MB {
		uploadTask := utils.NewUploadFile(zipFilePath, folderPath)

		//check third-Party service available
		thirdPartyAvailable := false
		if config.Get().General.UseExtFileHost {
			if uploadTask.CheckAvailable() == true {
				thirdPartyAvailable = true
			} else {
				logger.Info.Println(userInfo + "DownloadStickerSetQuery- third party NOT available！！！")
			}
			logger.Info.Println(userInfo + "DownloadStickerSetQuery-uploading(third party)")
			err = uploadTask.Upload2FileHost()
			if err != nil {
				logger.Error.Println(userInfo+"DownloadStickerSetQuery-failed to upload:", err)
				thirdPartyAvailable = false
			}
		}

		if config.Get().General.UseExtFileHost && thirdPartyAvailable == true {
			text := fmt.Sprintf(languages.Get(&update).BotMsg.UploadedThirdParty, stickerSet.Name, uploadTask.InfoRes.Data.File.Metadata.Size.Readable, uploadTask.InfoRes.Data.File.Url.Short)
			utils.EditMsgText(
				update.CallbackQuery.Message.Chat.ID, msg.MessageID,
				text,
				utils.EntityBold(text, stickerSet.Name),
				utils.EntityBold(text, uploadTask.InfoRes.Data.File.Metadata.Size.Readable),
			)
			logger.Info.Println(userInfo + "DownloadStickerSetQuery-upload (third party) successfully！！！")
		} else {
			// upload via Telegram separately
			err = uploadTask.UploadFragment(&update)
			if err != nil {
				logger.Error.Println(userInfo+"DownloadStickerSetQuery-failed to upload:", err)
				utils.EditMsgText(update.CallbackQuery.Message.Chat.ID, msg.MessageID, languages.Get(&update).BotMsg.ErrUploadFailed)
				return
			}

			text := fmt.Sprintf(languages.Get(&update).BotMsg.UploadedTelegram, stickerSet.Name, fileStat.Size()>>20)
			utils.EditMsgText(
				update.CallbackQuery.Message.Chat.ID,
				msg.MessageID,
				text,
				utils.EntityBold(text, stickerSet.Name),
				utils.EntityBold(text, fmt.Sprintf("%d", fileStat.Size()>>20)),
			)

			logger.Info.Println(userInfo + "DownloadStickerSetQuery-upload(Telegram-UploadFragment) successfully！！！")
		}

		uploadTask.Clean()
	} else {
		logger.Info.Println(userInfo + "DownloadStickerSetQuery-uploading(Telegram)")
		err = utils.SendFile(&update, zipFilePath)
		if err != nil {
			logger.Error.Println(userInfo+"DownloadStickerSetQuery-failed to upload:", err)
			utils.EditMsgText(update.CallbackQuery.Message.Chat.ID, msg.MessageID, languages.Get(&update).BotMsg.ErrUploadFailed)
			return
		}
		text := fmt.Sprintf(languages.Get(&update).BotMsg.UploadedTelegram, stickerSet.Name, fileStat.Size()>>20)
		utils.EditMsgText(
			update.CallbackQuery.Message.Chat.ID,
			msg.MessageID,
			text,
			utils.EntityBold(text, stickerSet.Name),
			utils.EntityBold(text, fmt.Sprintf("%d", fileStat.Size()>>20)),
		)
		logger.Info.Println(userInfo + "DownloadStickerSetQuery-upload(Telegram) successfully！！！")
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
			remoteFile, err := utils.BotGetFile(tgbotapi.FileConfig{
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
