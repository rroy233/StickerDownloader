package utils

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

func EntityBold(text, boldPart string) tgbotapi.MessageEntity {
	offset, length := getPartIndex(text, boldPart)
	return tgbotapi.MessageEntity{
		Type:   "bold",
		Offset: offset,
		Length: length,
	}
}

func EntityUnderline(text, boldPart string) tgbotapi.MessageEntity {
	offset, length := getPartIndex(text, boldPart)
	return tgbotapi.MessageEntity{
		Type:   "underline",
		Offset: offset,
		Length: length,
	}
}

func EntityLink(text, part, url string) tgbotapi.MessageEntity {
	offset, length := getPartIndex(text, part)
	return tgbotapi.MessageEntity{
		Type:   "url",
		Offset: offset,
		Length: length,
		URL:    url,
	}
}

func EntityMention(text, part string) tgbotapi.MessageEntity {
	offset, length := getPartIndex(text, part)
	return tgbotapi.MessageEntity{
		Type:   "mention",
		Offset: offset,
		Length: length,
	}
}

func EntityTag(text, part string) tgbotapi.MessageEntity {
	offset, length := getPartIndex(text, part)
	return tgbotapi.MessageEntity{
		Type:   "hashtag",
		Offset: offset,
		Length: length,
	}
}

func EntityCode(text, part string) tgbotapi.MessageEntity {
	offset, length := getPartIndex(text, part)
	return tgbotapi.MessageEntity{
		Type:   "code",
		Offset: offset,
		Length: length,
	}
}

func EntityTextLink(text, part, url string) tgbotapi.MessageEntity {
	offset, length := getPartIndex(text, part)
	return tgbotapi.MessageEntity{
		Type:   "text_link",
		Offset: offset,
		Length: length,
		URL:    url,
	}
}
