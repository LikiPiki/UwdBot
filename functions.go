package main

import (
	"fmt"
	"math/rand"
)

func genreatePhrase(phrases []string) string {
	return phrases[rand.Intn(len(phrases))]
}

func generatePhraseWithUsername(username string, phrases []string) string {
	for i, phrase := range phrases {
		phrases[i] = fmt.Sprintf(phrase, username)
	}
	return genreatePhrase(phrases)
}

func generateKek() string {
	phrases := []string{
		"Кек",
		"кпек",
		"пук",
		"КЕКУС",
		"КЕК",
	}
	return genreatePhrase(phrases)
}

func generateSolved(username string) string {
	phrases := []string{
		"@%s. Дядя, мы это уже решили!!",
		"@%s. Ну это уже решена чишо)",
	}
	return generatePhraseWithUsername(username, phrases)
}

func generateUserSolve(username string) string {
	phrases := []string{
		"@%s ну что, это правильно!",
		"@%s верно!",
		"@%s Никита был бы доволен твоим интелектом!",
		"@%s Верный ответ !",
	}
	return generatePhraseWithUsername(username, phrases)
}

func generateWrong(username string) string {
	phrases := []string{
		"@%s ну близко, но не то",
		"@%s я бы выбрал вариант выше, чем твой",
		"@%s Это конечно кек. Но неверно",
		"@%s Это конечно кек. Но неверно",
		"@%s Может стоит спросить у кого! Это не верно...",
		"@%s УУУУУУУУУУУУУУУУУУУУ нееее, не то...",
	}
	return generatePhraseWithUsername(username, phrases)
}
