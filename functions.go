package main

import (
	"fmt"
	"math/rand"
)

func generatePhrase(phrases []string) string {
	return phrases[rand.Intn(len(phrases))]
}

func generatePhraseWithUsername(username string, phrases []string) string {
	for i, phrase := range phrases {
		phrases[i] = fmt.Sprintf(phrase, username)
	}
	return generatePhrase(phrases)
}

func generateKek() string {
	phrases := []string{
		"Кек",
		"кпек",
		"пук",
		"КЕКУС",
		"КЕК",
	}
	return generatePhrase(phrases)
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

func GenerateRiot() (int, string) {
	phrases := []string{
		"Он нам не бонан!!!",
		"Бонан ЛОХ!",
		"@banannakryvay пишев ты!",
		"УУУУУ бонан самый худший админ",
	}
	stickers := []string{
		"CAADAgADBgADdPqvC4g0vr9WJeDGAg",
	}
	return GetStickerOrText(stickers, phrases)
}

func GetStickerOrText(stickers, phrases []string) (int, string) {
	chance := rand.Intn(2)
	if chance == Sticker {
		return chance, stickers[rand.Intn(len(stickers))]
	}
	return chance, phrases[rand.Intn(len(phrases))]
}
