package plugin

import (
	"fmt"
	"regexp"

	"github.com/LikiPiki/UwdBot/internal/pkg/database"
	"github.com/LikiPiki/UwdBot/internal/pkg/sender"
	"github.com/pkg/errors"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type Base struct {
	c      *sender.Sender
	Videos []string
	errors chan error
}

func (b *Base) Init(s *sender.Sender, db *database.Database) {
	b.c = s
	b.Videos = make([]string, 0)
	b.errors = make(chan error)
}

func (b *Base) HandleMessages(msg *tgbotapi.Message) {}

func (b *Base) HandleCommands(msg *tgbotapi.Message, command string) {
	switch command {
	case "last":
		link, fl, err := b.getLastVideoLink()
		if err != nil {
			b.errors <- errors.Wrap(err, "cannot handle commands")
		}

		if fl {
			if err := b.c.SendReply(msg, fmt.Sprintf("Последнее видео: %s", link)); err != nil {
				b.errors <- errors.Wrap(err, "cannot handle commands")
			}
		}
	case "kek":
		go func() {
			if err := b.c.SendReply(msg, generateKek()); err != nil {
				b.errors <- errors.Wrap(err, "cannot send reply")
			}
		}()
	case "riot":
		messageType, sending := GenerateRiot()
		go func() {
			if err := b.c.SendStickerOrText(msg, messageType, sending); err != nil {
				b.errors <- errors.Wrap(err, "cannot send sticker or text")
			}
		}()
	}
}

func (b *Base) HandleRegisterCommands(*tgbotapi.Message, string, *database.User) {}

func (b *Base) HandleCallbackQuery(*tgbotapi.Update) {}

func (b *Base) HandleAdminCommands(msg *tgbotapi.Message) {
	messageTextBytes := []byte(msg.Text)
	regexSay := regexp.MustCompile(`@say ([^\n]*)`)
	indexes := regexSay.FindSubmatchIndex(messageTextBytes)

	if len(indexes) == 4 {
		go func() {
			b.errors <- b.c.SendMessageToUWDChat(msg.Text[indexes[2]:indexes[3]])
		}()
	}
}

func (b *Base) GetRegisteredCommands() []string {
	return []string{}
}

func (b *Base) Errors() <-chan error {
	return b.errors
}
