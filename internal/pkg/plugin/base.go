package plugin

import (
	"math/rand"
)

const (
	Sticker = 1
)

func generatePhrase(phrases []string) string {
	return phrases[rand.Intn(len(phrases))]
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

func generateSolved() string {
	phrases := []string{
		"Дядя, мы это уже решили!!",
		"Ну это уже решена чишо)",
		"ну что, это правильно!",
		"верно!",
		"Никита был бы доволен твоим интелектом!",
		"Верный ответ !",
	}
	return generatePhrase(phrases)
}

func generateWrong() string {
	phrases := []string{
		"ну близко, но не то",
		"я бы выбрал вариант выше, чем твой",
		"Это конечно кек. Но неверно",
		"Это неверно...",
		"УУУУУУУ нееее, не то...",
	}
	return generatePhrase(phrases)
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
