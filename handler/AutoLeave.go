package handler

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rroy233/StickerDownloader/utils"
	"gopkg.in/rroy233/logger.v2"
)

func AutoLeave(update tgbotapi.Update) {
	//get chatID
	var chatID int64
	var ChatUsername string
	if update.MyChatMember != nil && (update.MyChatMember.Chat.Type == "channel" || update.MyChatMember.Chat.Type == "group" || update.MyChatMember.Chat.Type == "supergroup") {
		//add to channel or group
		if update.MyChatMember.NewChatMember.Status == "left" && update.MyChatMember.NewChatMember.User.UserName == utils.BotGetSelf().UserName {
			logger.Info.Println("AutoLeave - Response Got!!!," + utils.JsonEncode(update))
			return
		}
		chatID = update.MyChatMember.Chat.ID
		ChatUsername = update.MyChatMember.Chat.UserName
	} else if update.ChannelPost != nil {
		//send channel post
		chatID = update.ChannelPost.SenderChat.ID
		ChatUsername = update.ChannelPost.SenderChat.UserName
	} else if update.EditedChannelPost != nil {
		//edit channel post
		chatID = update.EditedChannelPost.SenderChat.ID
		ChatUsername = update.EditedChannelPost.SenderChat.UserName
	} else if update.Message != nil && update.Message.LeftChatMember != nil && (update.Message.Chat.Type == "group" || update.Message.Chat.Type == "supergroup") {
		//group message
		if update.Message.LeftChatMember.UserName == utils.BotGetSelf().UserName {
			logger.Info.Println("AutoLeave - Response Got!!!," + utils.JsonEncode(update))
			return
		}
		chatID = update.Message.Chat.ID
		ChatUsername = update.Message.Chat.UserName
	} else {
		//unknown
		logger.Error.Printf("AutoLeave - Invalid Update:%s", utils.JsonEncode(update))
		return
	}

	//notice
	text := "StickerDownloader currently does not support groups/channels. For more details, please visit https://github.com/rroy233/StickerDownloader."
	msg := tgbotapi.NewMessage(chatID, text)
	nMsg, err := utils.BotSend(msg)
	if err != nil {
		logger.Error.Printf("AutoLeave - Failed to send notice:%s,%s", err.Error(), utils.JsonEncode(update))
	} else {
		logger.Info.Println("AutoLeave - Notice Sent!!" + utils.JsonEncode(nMsg))
	}

	cf := tgbotapi.LeaveChatConfig{
		ChatID:          chatID,
		ChannelUsername: ChatUsername,
	}
	err = utils.BotRequest(cf)
	if err != nil {
		logger.Error.Printf("AutoLeave - Send error:%s,%s", err.Error(), utils.JsonEncode(update))
		return
	}
	logger.Info.Println("AutoLeave - Request Sent!!!,", utils.JsonEncode(update))
}
