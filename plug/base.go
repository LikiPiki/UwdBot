package plug

import (
	"fmt"
	"log"
	"math/rand"

	"github.com/PuerkitoBio/goquery"
)

const (
	Sticker          = 1
	UWDChannelVideos = "https://www.youtube.com/user/uwebdesign/videos"
	LEN              = 20
)

func (b *Base) getLastVideoLink() (string, bool) {
	b.ParseVideos()
	if len(b.Videos) > 0 {
		return b.Videos[0], true
	}
	return "", false
}

func (b *Base) ParseVideos() {
	b.Videos = make([]string, 0)
	doc, err := goquery.NewDocument(UWDChannelVideos)
	if err != nil {
		log.Println(err)
	}
	parsed := make([]string, 0)
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		class, _ := s.Attr("class")
		link, _ := s.Attr("href")
		if class == "yt-uix-sessionlink yt-uix-tile-link  spf-link  yt-ui-ellipsis yt-ui-ellipsis-2" {
			video := fmt.Sprintf("youtube.com%s", link)
			parsed = append(parsed, video)
		}
	})
	b.Videos = parsed
}

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
