package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/go-telegram-bot-api/telegram-bot-api"
)

type QuestionsData struct {
	Question string   `json:"question"`
	Valid    int      `json:"valid"`
	Users    []string `json:"users"`
	Answers  []string `json:"answers"`
	Time     time.Time
	Solved   bool
}

type Poll struct {
	Ok      bool          `json:"ok"`
	Data    QuestionsData `json:"data"`
	Message *tgbotapi.Message
	members map[string]int
}

func (p *Poll) GetSuccess() string {
	return p.Data.Answers[p.Data.Valid]
}

func (p *Poll) HaveMember(name string) int {
	return p.members[name]
}

func (p *Poll) GetPollResults(winner string) string {
	result := fmt.Sprintf(
		"`%s`\nПравильный ответ - ___%s___.\nОтветил - @%s",
		p.Message.Text,
		p.GetSuccess(),
		winner,
	)
	if len(p.members) > 1 {
		result += fmt.Sprintf(
			"\nПытались: ___%s___",
			p.getAllMembersUsernamesString(winner),
		)
	}
	return result
}

func (p *Poll) getAllMembersUsernamesString(winner string) (result string) {
	first := true
	for username := range p.members {
		if username == winner {
			continue
		}
		if !first {
			result += fmt.Sprintf(", %s", username)
		} else {
			result += username
			first = false
		}
	}
	return result
}

func (p *Poll) AddMember(name string) {
	p.members[name]++
}

func (p *Poll) Shuffle() {
	if len(p.Data.Answers) == 0 {
		return
	}

	valid := p.Data.Answers[0]

	for i := 0; i < len(p.Data.Answers); i++ {
		random := rand.Intn(len(p.Data.Answers))
		tmp := p.Data.Answers[random]
		p.Data.Answers[random] = p.Data.Answers[i]
		p.Data.Answers[i] = tmp
	}

	for k, ans := range p.Data.Answers {
		if ans == valid {
			p.Data.Valid = k
		}
	}
}
