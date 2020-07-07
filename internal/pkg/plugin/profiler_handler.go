package plugin

import (
	"context"
	"log"

	"github.com/LikiPiki/UwdBot/internal/pkg/database"
	"github.com/LikiPiki/UwdBot/internal/pkg/sender"
	"github.com/pkg/errors"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type Profiler struct {
	c      *sender.Sender
	ranks  []Rank
	db     *database.Database
	errors chan error
}

func (p *Profiler) Errors() <-chan error {
	return p.errors
}

func (p *Profiler) Init(s *sender.Sender, db *database.Database) {
	p.c = s
	p.ranks = []Rank{
		{"Надмозг", 800},
		{"Епископ", 700},
		{"Владелец казино", 650},
		{"Посетитель казино", 600},
		{"Дальнобойщик", 550},
		{"Король", 500},
		{"Мимокрокодил", 400},
		{"Работает в шиномонтаже", 350},
		{"Депутат от народа", 300},
		{"Зажиточный", 250},
		{"Программист", 200},
		{"Только что сдал ЕГЭ", 180},
		{"Пельмень", 150},
		{"Днарь", 120},
		{"Флексер", 100},
		{"Изгой", 0},
	}
	p.errors = make(chan error)
	p.db = db
}

func (p *Profiler) HandleMessages(msg *tgbotapi.Message) {}

func (p *Profiler) HandleCommands(msg *tgbotapi.Message, command string) {
	switch command {
	case "reg":
		if p.c.UWDChatID != msg.Chat.ID {
			if err := p.c.SendReplyToMessage(msg, "Этот функционал не работет в этом чате"); err != nil {
				p.errors <- errors.Wrap(err, "cannot send reply to message")
				return
			}
		}

		go func() {
			if err := p.c.SendReplyToMessage(
				msg,
				p.registerNewUser(context.Background(), msg),
			); err != nil {
				p.errors <- errors.Wrap(err, "cannot send reply to message")
			}
		}()
	}
}

func (p *Profiler) HandleRegisterCommands(msg *tgbotapi.Message, command string, user *database.User) {
	switch command {
	case "unreg":
		go func() {
			text, err := p.unregUser(context.Background(), msg)
			if err != nil {
				p.errors <- errors.Wrap(err, "cannot unreg user")
				return
			}

			if err := p.c.SendReplyToMessage(
				msg,
				text,
			); err != nil {
				p.errors <- errors.Wrap(err, "cannot send reply to message")
			}
		}()
	case "me":
		go func() {
			text, err := p.showUserInfo(context.Background(), msg)
			if err != nil {
				log.Println(err)
				p.errors <- errors.Wrap(err, "cannot get user info")
				return
			}

			if err := p.c.SendMarkdownReply(
				msg,
				text,
			); err != nil {
				p.errors <- errors.Wrap(err, "cannot send MD reply")
			}
		}()
	}
}

func (p *Profiler) HandleCallbackQuery(*tgbotapi.Update) {}

func (p *Profiler) HandleAdminCommands(msg *tgbotapi.Message) {
	p.HandleAdminRegexpCommands(msg)
}

func (p *Profiler) GetRegisteredCommands() []string {
	return []string{
		"unreg",
		"me",
	}
}
