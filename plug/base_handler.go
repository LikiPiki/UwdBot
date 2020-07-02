package plug

import (
	data "UwdBot/database"
	"UwdBot/sender"
	"fmt"
	"regexp"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type Base struct {
	c      *sender.Sender
	Videos []string
}

func (b *Base) Init(s *sender.Sender) {
	b.c = s
}

func (b *Base) HandleMessages(msg *tgbotapi.Message) {}

func (b *Base) HandleCommands(msg *tgbotapi.Message, command string) {
	switch command {
	case "about":
		go b.c.SendReply(
			msg,
			"https://teletype.in/@likipiki/corovan",
		)
	case "last":
		link, fl := b.getLastVideoLink()
		if fl {
			b.c.SendReply(msg,
				fmt.Sprintf("Последнее видео: %s", link),
			)
		}
	case "kek":
		go b.c.SendReply(
			msg,
			generateKek(),
		)
	case "riot":
		messageType, sending := GenerateRiot()
		go b.c.SendStickerOrText(
			msg,
			messageType,
			sending,
		)
	}
}

func (b *Base) HandleRegisterCommands(msg *tgbotapi.Message, command string, user *data.User) {}

func (b *Base) HandleCallbackQuery(update *tgbotapi.Update) {}

func (b *Base) HandleAdminCommands(msg *tgbotapi.Message) {
	messageTextBytes := []byte(msg.Text)
	regexSay := regexp.MustCompile(`@say ([^\n]*)`)
	indexes := regexSay.FindSubmatchIndex(messageTextBytes)

	if len(indexes) == 4 {
		go b.c.SendMessageToUWDChat(msg.Text[indexes[2]:indexes[3]])
	}
}

func (b *Base) GetRegisteredCommands() []string {
	return []string{}
}
