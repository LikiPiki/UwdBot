package main

import (
	"fmt"
	"math/rand"
)

type QuestionsData struct {
	Question string   `json:"question"`
	Answers  []string `json:"answers"`
}

type Poll struct {
	Ok   bool          `json:"ok"`
	Data QuestionsData `json:"data"`
}

func (p *Poll) Shuffle() {
	for i := 0; i < len(p.Data.Answers); i++ {
		random := rand.Intn(len(p.Data.Answers))
		tmp := p.Data.Answers[random]
		p.Data.Answers[random] = p.Data.Answers[i]
		p.Data.Answers[i] = tmp
	}
}

func (p *Poll) Numerate() {
	for i := 0; i < len(p.Data.Answers); i++ {
		p.Data.Answers[i] = fmt.Sprintf("%d. %s", i+1, p.Data.Answers[i])
	}
}
