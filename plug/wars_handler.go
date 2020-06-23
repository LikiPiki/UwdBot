package plug

import (
	data "UwdBot/database"
	"UwdBot/sender"
	"regexp"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type Wars struct {
	c                 *sender.Sender
	robbers           CaravanRobbers
	robberingProgress bool
}

func (w *Wars) Init(s *sender.Sender) {
	w.c = s
}

func (w *Wars) HandleMessages(msg *tgbotapi.Message) {
	re := regexp.MustCompile("^[b|B]uy (\\d+)")
	match := re.FindStringSubmatch(msg.Text)
	if len(match) > 1 {
		itemNumber, err := strconv.Atoi(match[1])
		if err != nil {
			w.c.SendReplyToMessage(msg, "Не правильно указан номер товара")
			return
		}
		w.buyItem(itemNumber, msg)
	}
}

func (w *Wars) HandleCommands(msg *tgbotapi.Message, command string) {}

func (w *Wars) HandleRegisterCommands(msg *tgbotapi.Message, command string, user *data.User) {
	switch command {
	case "caravan":
		reply := w.RobCaravans(msg, user)
		if reply != "" {
			go w.c.SendMarkdownReply(
				msg,
				reply,
			)
		}
	case "shop":
		go w.c.SendMarkdownReply(
			msg,
			w.GetShop(msg),
		)
	case "top":
		go w.c.SendMarkdownReply(
			msg,
			w.GetTopPlayers(usersInTopList),
		)
	}
}

func (w *Wars) HandleCallbackQuery(update *tgbotapi.Update) {}

func (w *Wars) HandleAdminCommands(msg *tgbotapi.Message) {}

func (w *Wars) GetRegisteredCommands() []string {
	return []string{
		"shop",
		"top",
		"caravan",
	}
}
