package utils

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rroy233/logger"
)

var msgQueue chan tgbotapi.Chattable

func initSender(senderNum int) {
	msgQueue = make(chan tgbotapi.Chattable, senderNum*2)
	for i := 0; i < senderNum; i++ {
		go sender()
	}
}

func addToSendQueue(msg tgbotapi.Chattable) {
	msgQueue <- msg
	return
}

func sender() {
	for {
		msg, _ := <-msgQueue
		rl.Take()
		_, err := bot.Request(msg)
		if err != nil {
			logger.Error.Println(loggerPrefix + "[sender]" + err.Error())
		}
	}
}
