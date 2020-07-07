package plugin

import (
	"context"

	"github.com/LikiPiki/UwdBot/internal/pkg/database"
	"github.com/LikiPiki/UwdBot/internal/pkg/sender"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/pkg/errors"
)

type Wars struct {
	c      *sender.Sender
	db     *database.Database
	errors chan error

	// Caravans
	robbers           Players
	robberingProgress bool

	// Arena
	arenaPlayers  Players
	arenaProgress bool
}

func (w *Wars) Init(s *sender.Sender, db *database.Database) {
	w.c = s
	w.db = db
	w.errors = make(chan error)
}

func (w *Wars) HandleMessages(msg *tgbotapi.Message) {
	w.HandleBuyItem(msg)
}

func (w *Wars) HandleCommands(msg *tgbotapi.Message, command string) {
	switch command {
	case "about":
		go w.c.SendReply(
			msg,
			"https://teletype.in/@likipiki/corovan",
		)
	}
}

func (w *Wars) HandleRegisterCommands(msg *tgbotapi.Message, command string, user *database.User) {
	switch command {
	case "arena":
		replyString := w.RegisterToArena(context.Background(), msg, user)
		if replyString == "" {
			return
		}

		if err := w.c.SendMarkdownReply(msg, replyString); err != nil {
			w.errors <- errors.Wrap(err, "cannot send arena reply")
		}

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
		"arena",
		"shop",
		"top",
		"caravan",
	}
}

func (w *Wars) Errors() <-chan error {
	return w.errors
}
