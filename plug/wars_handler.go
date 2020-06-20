package plug

import (
	data "UwdBot/database"
	"UwdBot/sender"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type Wars struct {
	c *sender.Sender
}

func (w *Wars) Init(s *sender.Sender) {
	w.c = s
}

func (w *Wars) HandleCommands(msg *tgbotapi.Message, command string) {}

func (w *Wars) HandleRegisterCommands(msg *tgbotapi.Message, command string, user *data.User) {
	switch command {
	case "shop":
		go w.c.SendMarkdownReply(
			msg,
			w.GetShop(msg),
		)
	}
}

func (w *Wars) HandleCallbackQuery(update *tgbotapi.Update) {}

func (w *Wars) HandleAdminCommands(msg *tgbotapi.Message) {}

func (w *Wars) GetRegisteredCommands() []string {
	return []string{
		"shop",
	}
}
