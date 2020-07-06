package plugin

import (
	"context"
	"log"
	"regexp"
	"strconv"

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
		{"Надмозг", 2_200},
		{"Епископ", 2_000},
		{"Владелец казино", 1_800},
		{"Посетитель казино", 1_600},
		{"Дальнобойщик", 1_400},
		{"Король", 1_500},
		{"Мимокрокодил", 1_300},
		{"Работает в шиномонтаже", 1_100},
		{"Депутат от народа", 1_000},
		{"Зажиточный", 700},
		{"Программист", 500},
		{"Только что сдал ЕГЭ", 400},
		{"Пельмень", 300},
		{"Днарь", 200},
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
	re := regexp.MustCompile("^[a|A]ddmoney (\\d+) (\\w+)$")
	match := re.FindStringSubmatch(msg.Text)
	if len(match) == 3 {
		itemNumber, err := strconv.Atoi(match[1])
		if err != nil {
			p.c.SendMarkdownReply(msg, "Команда введена не верно, пробуй ``/addmoney 100 username``")
			return
		}

		text, err := p.AddMoneyByUsername(context.Background(), itemNumber, match[2])
		if err != nil {
			p.errors <- errors.Wrap(err, "cannot add money by username")
			return
		}
		if err := p.c.SendMarkdownReply(
			msg,
			text,
		); err != nil {
			p.errors <- errors.Wrap(err, "cannot send MD reply")
		}
	}
}

func (p *Profiler) GetRegisteredCommands() []string {
	return []string{
		"unreg",
		"me",
	}
}
