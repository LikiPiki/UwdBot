package main

import (
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
}

func (p *Poll) GetSuccess() string {
	return p.Data.Answers[p.Data.Valid]
}

func (p *Poll) Shuffle() {
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
