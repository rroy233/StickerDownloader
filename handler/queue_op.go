package handler

import (
	"errors"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rroy233/StickerDownloader/db"
	"github.com/rroy233/StickerDownloader/languages"
	"github.com/rroy233/StickerDownloader/utils"
	"gopkg.in/rroy233/logger.v2"
	"time"
)

// 封装排队及等待操作
// *db.QItem 用于后续的弃权、出队操作
// bool 用于告知调用方是否要return结束
//
// 调用方法示例：
//
// qItem, quit := enqueue(&update, &msg)
// if quit == true {
// return
// }
func enqueue(update *tgbotapi.Update, queueEditMsg *tgbotapi.Message) (*db.QItem, bool) {
	oldMsgText := queueEditMsg.Text
	needRecover := false
	qItem, err := db.EnQueue(utils.GetUID(update))
	if err != nil {
		if errors.Is(err, db.ErrorQueueFull) {
			logger.Warn.Printf("[handler.enqueue]Queue is FULL! chatID:%d MsgID:%d", queueEditMsg.Chat.ID, queueEditMsg.MessageID)
			utils.EditMsgText(queueEditMsg.Chat.ID, queueEditMsg.MessageID, languages.Get(update).BotMsg.ErrSysBusy)
			return nil, true
		}
		utils.EditMsgText(queueEditMsg.Chat.ID, queueEditMsg.MessageID, languages.Get(update).BotMsg.ErrFailed)
		return nil, true
	}
	beginTime := time.Now()
	waitingNum := -2
	progressMsgInit := false
	for true {
		//timeout
		if time.Now().Sub(beginTime).Seconds() > float64(db.QueueTimeout) {
			utils.EditMsgText(queueEditMsg.Chat.ID, queueEditMsg.MessageID, languages.Get(update).BotMsg.ErrTimeout)
			return nil, true
		}
		//aborted by user or some else
		if qItem.IsAbort() == true || qItem.QueryFront() == -1 {
			utils.EditMsgText(queueEditMsg.Chat.ID, queueEditMsg.MessageID, languages.Get(update).BotMsg.QueueAborted)
			return nil, true
		}
		//stop waiting,start service
		if qItem.QueryFront() == 0 {
			break
		}
		//update waiting progress
		if progressMsgInit == false {
			//init msg and inline keyboard
			err = utils.BotRequest(tgbotapi.NewEditMessageTextAndMarkup(queueEditMsg.Chat.ID, queueEditMsg.MessageID, "Loading...",
				tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData(languages.Get(update).BotMsg.QueueAbortBtn, QuitQueueCallbackQueryPrefix+qItem.UUID),
					),
				),
			))
			if err != nil {
				logger.Error.Println(err)
			}
			progressMsgInit = true
			needRecover = true
		}
		if waitingNum != qItem.QueryFront() {
			waitingNum = qItem.QueryFront()
			utils.EditMsgTextAndMarkup(queueEditMsg.Chat.ID, queueEditMsg.MessageID, fmt.Sprintf(languages.Get(update).BotMsg.QueueProcess, waitingNum),
				tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData(languages.Get(update).BotMsg.QueueAbortBtn, QuitQueueCallbackQueryPrefix+qItem.UUID),
					),
				),
			)
			needRecover = true
		}
		time.Sleep(3 * time.Second)
	}

	//recover old msg text
	if needRecover == true {
		utils.EditMsgText(queueEditMsg.Chat.ID, queueEditMsg.MessageID, oldMsgText)
	}

	return qItem, false
}

// 封装出队操作
func dequeue(qItem *db.QItem) {
	qItem.Abort()
	if err := qItem.DeQueue(); err != nil {
		logger.Info.Println("qItem.DeQueue(),error", err)
	}
	return
}
