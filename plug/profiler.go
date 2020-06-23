package plug

import (
	data "UwdBot/database"
	"fmt"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var (
	CHAT_ID   int64
	UserRanks = []Rank{
		{"Епископ", 1500},
		{"Владелец казино", 1300},
		{"Дальнобойщик", 1100},
		{"Король", 1000},
		{"Работает в шиномонтаже", 850},
		{"Депутат от народа", 700},
		{"Зажиточный", 500},
		{"Программист", 300},
		{"Только что сдал ЕГЭ", 150},
		{"Пельмень", 100},
		{"Днарь", 50},
		{"Изгой", 0},
	}
)

type Rank struct {
	Rank       string
	Reputation int
}

func (p *Profiler) SetChatID(ID int64) {
	CHAT_ID = ID
}

func GetMarkdownUsername(username string) string {
	return strings.ReplaceAll(username, "_", "\\_")
}

func getRank(user data.User) string {
	for _, rank := range UserRanks {
		if rank.Reputation <= user.Reputation {
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
		"***ЛИЧНАЯ КАРТОЧКА***\nПривет ***@%s*** - ___%s___\nТвоя репутация 👑: ***%d\n***Монеты💰: ***%d***\nБоевая мощь: ***%d***\nНа сегодня осталось ***%d*** единиц активности!\n\nТы на ***%d***%% круче остальных и на ***%d***%% богаче!",
		user.Username,
		rank,
		user.Reputation,
		user.Coins,
		user.WeaponsPower,
		user.Activity,
		int(repStat*100),
		int(coinsStat*100),
	)
}

func (p *Profiler) AddMoneyByUsername(money int, username string) string {
	user := data.User{}
	var err error
	user, err = user.FindUserByUsername(username)
	if err != nil {
		return fmt.Sprintf(
			"Пользоваателя %s не существует!",
			GetMarkdownUsername(username),
		)
	}
	user.AddMoney(money)
	return fmt.Sprintf(
		"Пользователю **@%s** начислено **%d💰**",
		GetMarkdownUsername(username),
		money,
	)
}

func (p *Profiler) registerNewUser(msg *tgbotapi.Message) string {
	user := data.User{}
	count, err := user.CountUsersWithID(msg.From.ID)
	if err != nil {
		return "Что то пошло не так..."
	}
	if count > 0 {
		return "Ты уже зарегистрирован!"
	}

	user.UserID = uint64(msg.From.ID)
	user.Username = msg.From.UserName
	if len(user.Username) == 0 {
		return "Для начала придумай себе nickname в Телеграме! "
	}
	_, err = user.CreateNewUser()

	if err != nil {
		return "Не удалось добавить. Попробуй позже..."
	}

	return "Вы успешно прошли регистрацию. /me"
}
