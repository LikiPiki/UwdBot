package plug

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	data "UwdBot/database"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

const (
	usersInTopList = 10
	robCount       = 3
)

type CaravanRobber struct {
	ID         uint64
	UserID     uint64
	Username   string
	Power      int
	Reputation int
	Coins      int
}

type CaravanRobbers [robCount]CaravanRobber

func (c *CaravanRobbers) checkRobberById(userID uint64) bool {
	for _, caravan := range c {
		if caravan.UserID == userID {
			return true
		}
	}
	return false
}

func (c *CaravanRobbers) checkRobbersCount() int {
	count := 0
	for _, caravan := range c {
		if caravan.UserID != 0 {
			count++
		}
	}
	return count
}

func (c *CaravanRobbers) getReputationAndCoins() (int, int) {
	var coins, reputation int
	for _, caravan := range c {
		reputation += caravan.Reputation
		coins += caravan.Coins
	}
	coins = int((float32(coins)*0.1)/3 + 10)
	reputation = int((float32(reputation)*0.1)/3 + 3)
	return coins, reputation
}

func (w *Wars) RobCaravans(msg *tgbotapi.Message, user *data.User) string {
	robbersCount := w.robbers.checkRobbersCount()
	if robbersCount == robCount {
		return "🐫🐪🐫"
	}

	if w.robbers.checkRobberById(uint64(msg.From.ID)) {
		return "Ты уже учавствуешь в набеге!"
	}
	w.robbers[robbersCount] = CaravanRobber{
		user.ID, user.UserID, user.Username, user.WeaponsPower, user.Reputation, user.Coins,
	}
	robbersCount = w.robbers.checkRobbersCount()
	if robbersCount == robCount {
		if w.robberingProgress == false {
			go w.caravansStart(msg)
			return ""
		}
	}

	return fmt.Sprintf(
		"Для отправления каравана нужно еще ***%d*** грабителя!",
		robCount-robbersCount,
	)
}

func (w *Wars) caravansStart(msg *tgbotapi.Message) {
	startPhrase := "Игроки: "
	playersPhrase := ""
	ids := make([]int, 0)
	for i, rob := range w.robbers {
		playersPhrase += "@" + GetMarkdownUsername(rob.Username)
		ids = append(ids, int(rob.UserID))
		if i != (robCount - 1) {
			playersPhrase += ", "
		}
	}
	startPhrase += playersPhrase
	reply := tgbotapi.NewMessage(
		msg.Chat.ID,
		fmt.Sprintf(
			"Игроки: **%s** начинают набег на караван. **Посмотрим что у них выйдет**\n\n__Это может занять какое то время!__",
			playersPhrase,
		),
	)
	reply.ParseMode = "markdown"
	reply.ReplyToMessageID = msg.MessageID

	msgStart := w.c.Send(&reply)

	w.robberingProgress = true
	timeLeft := 1 + rand.Intn(10)
	timer1 := time.NewTimer(time.Minute * time.Duration(timeLeft))

	earnCoins, earnReputation := w.robbers.getReputationAndCoins()
	user := data.User{}
	<-timer1.C
	if rand.Intn(2) == 0 {
		user.AddMoneyToUsers(earnCoins, ids)
		user.AddReputationToUsers(earnReputation, ids)
		w.c.SendMarkdownReply(
			msgStart,
			fmt.Sprintf(
				"Игрокам %s удалось одержать победу, им будет добавлено по **%d** монет и **%d** репутации",
				playersPhrase,
				earnCoins,
				earnReputation,
			),
		)
	} else {
		user.DecreaseMoneyToUsers(10, ids)
		w.c.SendMarkdownReply(
			msgStart,
			fmt.Sprintf(
				"Игрокам %s не удалось победить караван, их репутация упала на ***10*** баллов",
				playersPhrase,
			),
		)
	}

	for i := 0; i < robCount; i++ {
		w.robbers[i] = CaravanRobber{}
	}

	w.robberingProgress = false
}

func (w *Wars) GetTopPlayers(count int) string {
	user := data.User{}
	result := "**ТОП ИГРОКОВ:**\n"
	topUsers, err := user.GetTopUsers(count)

	log.Println(err)

	if err != nil {
		return "Что то пошло не так..."
	}

	for i, us := range topUsers {
		result += fmt.Sprintf(
			"%d) %s: %d👑 %d💰\n",
			i+1,
			GetMarkdownUsername(us.Username),
			us.Reputation,
			us.Coins,
		)
	}

	result += "\n__Региструйся и победи всех__ **/reg**"
	return result
}

func (w *Wars) GetShop(msg *tgbotapi.Message) string {
	weap := data.Weapon{}
	weapons, err := weap.GetAllWeapons()
	if err != nil {
		return "Не удалось загрузить магазин..."
	}
	reply := "***Уютный shop 🛒 ***\n\n***Оружие:***\n"
	for _, w := range weapons {
		reply += fmt.Sprintf(
			"%d) ___%s___ %d🗡️, %d💰\n",
			w.ID,
			w.Name,
			w.Power,
			w.Cost,
		)
	}
	reply += "\n___Интересный стафф:___\nПоявится в скором времени...\n\n___Купить товар - реплай на сообщение buy номер товара___"
	return reply
}

func (w *Wars) buyItem(item int, msg *tgbotapi.Message) {
	var err error
	var user data.User
	user, err = user.FindUserByID(msg.From.ID)
	if err != nil {
		w.c.SendReplyToMessage(msg, "Вы не зарегистрированы /reg")
		return
	}
	var weapon data.Weapon
	weapon, err = weapon.GetWeaponsByID(item)
	if err != nil {
		w.c.SendReplyToMessage(msg, "Некоректный номер товара!")
		return
	}

	if user.Coins >= weapon.Cost {
		user.DecreaseMoney(weapon.Cost)
		user.AddPower(weapon.Power)
		w.c.SendMarkdownReply(
			msg,
			fmt.Sprintf(
				"Списано ***%d***💰, куплен(а): ___%s___!\n\nПрибавлено %d к боевой мощи!",
				weapon.Cost,
				weapon.Name,
				weapon.Power,
			),
		)
	} else {
		w.c.SendMarkdownReply(
			msg,
			fmt.Sprintf(
				"Вам не хватает ***%d***💰, чтобы купить ___%s___!",
				weapon.Cost-user.Coins,
				weapon.Name,
			),
		)
	}
}
