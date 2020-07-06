package plugin

import (
	"context"
	"github.com/LikiPiki/UwdBot/internal/pkg/database"
	"github.com/LikiPiki/UwdBot/internal/pkg/sender"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/pkg/errors"
	"regexp"
	"strconv"
)

type Wars struct {
	c                 *sender.Sender
	robbers           CaravanRobbers
	robberingProgress bool
	db                *database.Database
	errors            chan error
}

func (w *Wars) Init(s *sender.Sender, db *database.Database) {
	w.c = s
	w.db = db
	w.errors = make(chan error)
}

func (w *Wars) HandleMessages(msg *tgbotapi.Message) {
	re := regexp.MustCompile("^[b|B]uy (\\d+)")
	match := re.FindStringSubmatch(msg.Text)
	if len(match) > 1 {
		itemNumber, err := strconv.Atoi(match[1])
		if err != nil {
			if err := w.c.SendReplyToMessage(msg, "Не правильно указан номер товара"); err != nil {
				w.errors <- errors.Wrap(err, "cannot send reply to message")
			}
			return
		}
		w.buyItem(context.Background(), itemNumber, msg)
	}
}

func (w *Wars) HandleCommands(*tgbotapi.Message, string) {}

func (w *Wars) HandleRegisterCommands(msg *tgbotapi.Message, command string, user *database.User) {
	switch command {
	case "caravan":
		reply := w.RobCaravans(context.Background(), msg, user)
		if reply != "" {
			go w.c.SendMarkdownReply(
				msg,
				reply,
			)
		}
	case "shop":
		go w.c.SendMarkdownReply(
			msg,
			w.GetShop(context.Background()),
		)
	case "top":
		go w.c.SendMarkdownReply(
			msg,
			w.GetTopPlayers(context.Background(), usersInTopList),
		)
	}
}

func (w *Wars) HandleCallbackQuery(*tgbotapi.Update) {}

func (w *Wars) HandleAdminCommands(*tgbotapi.Message) {}

func (w *Wars) GetRegisteredCommands() []string {
	return []string{
		"shop",
		"top",
		"caravan",
	}
}

func (w *Wars) Errors() <-chan error {
	return w.errors
}
