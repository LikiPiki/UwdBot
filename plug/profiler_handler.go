package plug

import (
	data "UwdBot/database"
	"UwdBot/sender"
	"regexp"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type Profiler struct {
	c *sender.Sender
}

func (p *Profiler) Init(s *sender.Sender) {
	p.c = s
}

func (p *Profiler) HandleMessages(msg *tgbotapi.Message) {}

func (p *Profiler) HandleCommands(msg *tgbotapi.Message, command string) {
	switch command {
	case "reg":
		if CHAT_ID != msg.Chat.ID {
			p.c.SendReplyToMessage(msg, "Этот функционал не работет в этом чате")
			return
		}
		go p.c.SendReplyToMessage(
			msg,
			p.registerNewUser(msg),
		)
	}
}

func (p *Profiler) HandleRegisterCommands(msg *tgbotapi.Message, command string, user *data.User) {
	switch command {
	case "unreg":
		go p.c.SendReplyToMessage(
			msg,
			p.unregUser(msg),
		)
	case "me":
		go p.c.SendMarkdownReply(
			msg,
			p.showUserInfo(msg),
		)
	}
}

func (p *Profiler) HandleCallbackQuery(update *tgbotapi.Update) {
}

func (p *Profiler) HandleAdminCommands(msg *tgbotapi.Message) {
	re := regexp.MustCompile("^[a|A]ddmoney (\\d+) (\\w+)$")
	match := re.FindStringSubmatch(msg.Text)
	if len(match) == 3 {
		itemNumber, err := strconv.Atoi(match[1])
		if err != nil {
			p.c.SendMarkdownReply(msg, "Команда введена не верно, пробуй ``/addmoney 100 username``")
			return
		}
		p.c.SendMarkdownReply(
			msg,
			p.AddMoneyByUsername(itemNumber, match[2]),
		)
	}
}

func (p *Profiler) GetRegisteredCommands() []string {
	return []string{
		"unreg",
		"me",
	}
}
