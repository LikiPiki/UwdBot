package plugin

import (
	"context"
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"time"

	"github.com/LikiPiki/UwdBot/internal/pkg/database"
	"github.com/pkg/errors"

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

func (w *Wars) RobCaravans(ctx context.Context, msg *tgbotapi.Message, user *database.User) string {
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
			go w.caravansStart(ctx, msg)
			return ""
		}
	}

	return fmt.Sprintf(
		"Для отправления каравана нужно еще ***%d*** грабителя!",
		robCount-robbersCount,
	)
}

func (w *Wars) caravansStart(ctx context.Context, msg *tgbotapi.Message) {
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

	msgStart, err := w.c.Send(&reply)
	if err != nil {
		w.errors <- errors.Wrap(err, "cannot send message")
	}

	w.robberingProgress = true
	timeLeft := 1 + rand.Intn(10)
	timer1 := time.NewTimer(time.Minute * time.Duration(timeLeft))
	<-timer1.C
	if rand.Intn(2) == 0 {
		if err := w.db.UserStorage.AddMoneyToUsers(ctx, 10, ids); err != nil {
			w.errors <- errors.Wrap(err, "cannot add money to users")
			return
		}

		if err := w.db.UserStorage.AddReputationToUsers(ctx, 3, ids); err != nil {
			w.errors <- errors.Wrap(err, "cannot add reputation to users")
			return
		}

		err := w.c.SendMarkdownReply(
			msgStart,
			fmt.Sprintf(
				"Игрокам %s удалось одержать победу, им будет добавлено по **%d** монет и **%d** репутации",
				playersPhrase,
				10,
				3,
			),
		)
		if err != nil {
			w.errors <- errors.Wrap(err, "cannot send reply")
		}
	} else {
		if err := w.db.UserStorage.DecreaseReputationToUsers(ctx, 1, ids); err != nil {
			w.errors <- errors.Wrap(err, "cannot decrease money")
			return
		}

		err := w.c.SendMarkdownReply(
			msgStart,
			fmt.Sprintf(
				"Игрокам %s не удалось победить караван, их репутация упала на ***1*** бал",
				playersPhrase,
			),
		)
		if err != nil {
			w.errors <- errors.Wrap(err, "cannot send reply")
		}
	}

	// 7 % chance to find treasure
	treasureChance := rand.Intn(100)
	if treasureChance >= 92 {
		treasureCoins := 50 + rand.Intn(101)

		if err := w.db.UserStorage.AddMoneyToUsers(ctx, treasureCoins, ids); err != nil {
			w.errors <- errors.Wrap(err, "cannot add money to users")
			return
		}

		err = w.c.SendMarkdownReply(
			msgStart,
			fmt.Sprintf(
				"Игроки %s нашли мифическое сокровище старого фараона, им начислено по **%d**💰",
				playersPhrase,
				treasureCoins,
			),
		)
		if err != nil {
			w.errors <- errors.Wrap(err, "cannot send reply")
		}
	}

	for i := 0; i < robCount; i++ {
		w.robbers[i] = CaravanRobber{}
	}

	w.robberingProgress = false
}

func (w *Wars) GetTopPlayers(ctx context.Context, count int) string {
	result := "**ТОП ИГРОКОВ:**\n"
	users, err := w.db.UserStorage.GetTopUsers(ctx, count)
	if err != nil {
		w.errors <- errors.Wrap(err, "cannot get top users")
		return ""
	}

	for i, us := range users {
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

func (w *Wars) HandleBuyItem(msg *tgbotapi.Message) {
	re := regexp.MustCompile("^[b|B]uy (\\d+) ?(\\d+)?")
	match := re.FindStringSubmatch(msg.Text)

	if len(match) == 3 {
		count := 1

		if match[2] != "" {
			var err error
			count, err = strconv.Atoi(match[2])
			if err != nil {
				w.errors <- errors.Wrap(err, "cannot convert buy match[2] to integer")
			}
		}
		itemNumber, err := strconv.Atoi(match[1])
		if err != nil {
			if err := w.c.SendReplyToMessage(msg, "Не правильно указан номер товара"); err != nil {
				w.errors <- errors.Wrap(err, "cannot send reply to message")
			}
			return
		}

		w.buyItem(context.Background(), itemNumber, count, msg)
	}
}

func (w *Wars) GetShop(ctx context.Context) string {
	weapons, err := w.db.WeaponStorage.GetAllWeapons(ctx)
	if err != nil {
		w.errors <- errors.Wrap(err, "cannot get weapons")
		return ""
	}

	reply := "***Уютный shop 🛒 ***\n\n***Оружие:***\n"
	for _, w := range weapons {
		reply += fmt.Sprintf(
			"%d) ___%s___ %d🏹️, %d💰\n",
			w.ID,
			w.Name,
			w.Power,
			w.Cost,
		)
	}
	reply += "\n___Интересный стафф 🦄:___\nПоявится в скором времени...\n\n___Купить товар - реплай на сообщение buy номер товара___"
	return reply
}

func (w *Wars) buyItem(ctx context.Context, item int, count int, msg *tgbotapi.Message) {
	user, err := w.db.UserStorage.FindUserByID(ctx, msg.From.ID)
	if err != nil {
		w.errors <- errors.Wrap(err, "cannot find user")
		return
	}

	if user.ID == 0 {
		if err := w.c.SendReplyToMessage(msg, "Вы не зарегистрированы /reg"); err != nil {
			w.errors <- errors.Wrap(err, "cannot send reply")
		}
		return
	}

	weapon, err := w.db.WeaponStorage.GetWeaponsByID(ctx, item)
	if err != nil {
		w.errors <- errors.Wrap(err, "cannot get weapon by id")
		return
	}

	if user.Coins >= weapon.Cost*count {
		if err := w.db.UserStorage.DecreaseMoney(ctx, user.UserID, weapon.Cost*count); err != nil {
			w.errors <- errors.Wrap(err, "cannot decrease money")
			return
		}

		if err := w.db.UserStorage.AddPower(ctx, int(user.UserID), weapon.Power*count); err != nil {
			w.errors <- errors.Wrap(err, "cannot add power")
			return
		}

		var err error
		switch count {
		case 1:
			err = w.c.SendMarkdownReply(
				msg,
				fmt.Sprintf(
					"Списано ***%d***💰, куплен(а): ___%s___!\n\nПрибавлено %d 🏹 к боевой мощи!",
					weapon.Cost,
					weapon.Name,
					weapon.Power,
				),
			)
		default:
			err = w.c.SendMarkdownReply(
				msg,
				fmt.Sprintf(
					"Списано ***%d***💰, куплен(а):  ***%d x ***___%s___!\n\nПрибавлено %d 🏹 к боевой мощи!",
					weapon.Cost*count,
					count,
					weapon.Name,
					weapon.Power*count,
				),
			)
		}

		if err != nil {
			w.errors <- errors.Wrap(err, "cannot send reply")
		}

	} else {
		err := w.c.SendMarkdownReply(
			msg,
			fmt.Sprintf(
				"Вам не хватает ***%d***💰, чтобы купить ___%s___!",
				weapon.Cost*count-user.Coins,
				weapon.Name,
			),
		)
		if err != nil {
			w.errors <- errors.Wrap(err, "cannot send reply")
		}
	}
}
