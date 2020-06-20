package plug

import (
	data "UwdBot/database"
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var (
	CHAT_ID   int64
	UserRanks = []Rank{
		{"Король", 1000, 1000},
		{"Депутат от народа", 0, 500},
		{"Зажиточный", 500, 300},
		{"Программист", 300, 300},
		{"Только что сдал ЕГЭ", 150, 50},
		{"Пельмень", 100, 100},
		{"Днарь", 0, 50},
		{"Изгой", 0, 0},
	}
)

type Rank struct {
	Rank       string
	Coins      int
	Reputation int
}

func (p *Profiler) SetChatID(ID int64) {
	CHAT_ID = ID
}

func getRank(user data.User) string {
	for _, rank := range UserRanks {
		if (rank.Coins <= user.Coins) && (rank.Reputation <= user.Reputation) {
			return rank.Rank
		}
	}
	return UserRanks[len(UserRanks)-1].Rank
}

func (p *Profiler) unregUser(msg *tgbotapi.Message) string {
	user := data.User{}
	user.DeleteUser(msg.From.ID)
	return "Ну заходи как нибудь еще, что делать..."
}

func (p *Profiler) showUserInfo(msg *tgbotapi.Message) string {
	var err error
	var user data.User
	user, err = user.FindUserByID(msg.From.ID)
	if err != nil {
		log.Println(err)
	}

	var repStat, coinsStat float32
	repStat, coinsStat, err = user.GetUserStatistics()
	if err != nil {
		log.Println(err)
	}

	rank := getRank(user)

	return fmt.Sprintf(
		"Привет ***@%s*** - ___%s___\nТвоя репутация: ***%d\n***💰: ***%d***\n\nТы на ***%d***%% круче остальных и на ***%d***%% богаче!",
		user.Username,
		rank,
		user.Reputation,
		user.Coins,
		int(repStat*100),
		int(coinsStat*100),
	)
}

func (p *Profiler) registerNewUser(msg *tgbotapi.Message) string {
	user := data.User{}
	count, err := user.CountUsersWithID(msg.From.ID)
	if err != nil {
		log.Panicln(err)
		return "Что то пошло не так..."
	}
	if count > 0 {
		return "Ты уже зарегистрирован!"
	}

	user.UserID = uint64(msg.From.ID)
	user.Username = msg.From.UserName
	_, err = user.CreateNewUser()

	if err != nil {
		return "Не удалось добавить. Попробуй позже..."
	}

	return "Вы успешно прошли регистрацию. /me"
}
