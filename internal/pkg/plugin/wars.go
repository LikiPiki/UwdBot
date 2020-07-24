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
	usersCountInTopList   = 10
	caravanPlayersToStart = 3
	// Arena constants
	arenaPlayersToStart = 2
	minArenaMoney       = 30
	minArenaReputation  = 5
	arenaRoundMaxTime   = 5
	// Shop constants
	shopItemsPerpage = 5
)

// Player struct for arena and caravans players
type Player struct {
	ID         uint64
	UserID     uint64
	Username   string
	Power      int
	Reputation int
	Coins      int
}

// Players - players array
type Players []Player

func checkPlayerByID(players Players, userID uint64) bool {
	for _, player := range players {
		if player.UserID == userID {
			return true
		}
	}
	return false
}

func (c *Players) getReputationAndCoins() (int, int) {
	var coins, reputation int
	for _, caravan := range *c {
		reputation += caravan.Reputation
		coins += caravan.Coins
	}
	coins = int((float32(coins)*0.1)/3 + 10)
	reputation = int((float32(reputation)*0.1)/3 + 3)
	return coins, reputation
}

func (c *Players) getPhraseAndIds() (string, []int) {
	playersPhrase := ""
	ids := make([]int, 0)

	for i, rob := range *c {
		playersPhrase += "@" + GetMarkdownUsername(rob.Username)
		ids = append(ids, int(rob.UserID))
		if i != (caravanPlayersToStart - 1) {
			playersPhrase += ", "
		}
	}
	return playersPhrase, ids
}

func getPlayersIDs(players Players) []int {
	ids := make([]int, 0)
	for _, pl := range players {
		ids = append(ids, int(pl.UserID))
	}
	return ids
}

func checkPlayersCount(players Players) int {
	count := 0
	for _, player := range players {
		if player.UserID != 0 {
			count++
		}
	}
	return count
}

func (w *Wars) RobCaravans(ctx context.Context, msg *tgbotapi.Message, user *database.User) string {
	robbersCount := checkPlayersCount(w.robbers)
	if robbersCount == caravanPlayersToStart {
		return "🐫🐪🐫"

	}
	if checkPlayerByID(w.robbers, uint64(msg.From.ID)) {
		return "Ты уже учавствуешь в набеге!"
	}
	w.robbers[robbersCount] = Player{
		user.ID, user.UserID, user.Username, user.WeaponsPower, user.Reputation, user.Coins,
	}
	robbersCount = checkPlayersCount(w.robbers)
	if robbersCount == caravanPlayersToStart {
		if !w.robberingProgress {
			go w.caravansStart(ctx, msg)
			return ""
		}
	}

	return fmt.Sprintf(
		"Для отправления каравана нужно еще %d грабителя!",
		caravanPlayersToStart-robbersCount,
	)
}

func (w *Wars) caravansStart(ctx context.Context, msg *tgbotapi.Message) {
	startPhrase := "Игроки: "
	playersPhrase, ids := w.robbers.getPhraseAndIds()
	startPhrase += playersPhrase
	reply := tgbotapi.NewMessage(
		msg.Chat.ID,
		fmt.Sprintf(
			"Игроки: %s начинают набег на караван. Посмотрим что у них выйдет\n\n_Это может занять какое то время!_",
			playersPhrase,
		),
	)
	reply.ParseMode = "markdown"
	reply.ReplyToMessageID = msg.MessageID

	if w.lastCaravanMessageWithCallback.From != nil {
		go func() {
			timer1 := time.NewTimer(time.Second * time.Duration(2))
			<-timer1.C

			if err := w.c.DeleteMessage(w.lastCaravanMessageWithCallback); err != nil {
				w.errors <- errors.Wrap(err, "cannot delete last caravan callbackMsg after 2 sec")
			}
			// clear last caravan message with callback button
			w.lastCaravanMessageWithCallback = &tgbotapi.Message{}
		}()
	}

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
				"Игрокам %s не удалось победить караван, их репутация упала на ***1*** балл",
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

	for i := 0; i < caravanPlayersToStart; i++ {
		w.robbers[i] = Player{}
	}

	w.robberingProgress = false
}

func (w *Wars) SendCaravanInvite(ctx context.Context, msg *tgbotapi.Message, user *database.User) {
	if w.robberingProgress {
		err := w.c.SendReply(
			msg,
			"🐫🐪🐫",
		)

		if err != nil {
			w.errors <- errors.Wrap(err, "cannot send robberingProgress from fastcaravan")
		}
		return
	}

	reply := tgbotapi.NewMessage(msg.Chat.ID, "Друзья, давайте грабить караван!")

	w.RobCaravans(ctx, msg, user)

	currentCaravanRobbers := checkPlayersCount(w.robbers)
	replyMarkup := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf(
					"Погнааале [ %d / %d ]",
					currentCaravanRobbers,
					caravanPlayersToStart,
				),
				"join",
			),
		),
	)
	reply.ReplyMarkup = replyMarkup

	lastCaravanMessage, err := w.c.Send(&reply)

	if err != nil {
		w.errors <- errors.Wrap(err, "cannot send caravan message")
		return
	}

	if w.lastCaravanMessageWithCallback.From != nil {
		if err := w.c.DeleteMessage(w.lastCaravanMessageWithCallback); err != nil {
			w.errors <- errors.Wrap(err, "cannot delete last caravan callback message")
		}
	}

	w.lastCaravanMessageWithCallback = lastCaravanMessage
}

func (w *Wars) HandleCaravanCallbackQuery(update *tgbotapi.Update) {
	if update.CallbackQuery != nil && (update.CallbackQuery.Data != "join") {
		return
	}

	user, err := w.db.UserStorage.FindUserByID(context.Background(), update.CallbackQuery.From.ID)
	if err != nil {
		// w.errors <- errors.Wrap(err, "cannot find user in fastcaravan")
		if err := w.c.SendInlineKeyboardReply(update.CallbackQuery, "Сначала зарегистрируйся"); err != nil {
			w.errors <- errors.Wrap(err, "cannot send inline keyboard reply from caravan callback")
			return
		}
		return
	}

	// костыль
	msg := update.CallbackQuery.Message
	msg.From.ID = update.CallbackQuery.From.ID

	robbersBefore := checkPlayersCount(w.robbers)
	reply := w.RobCaravans(
		context.Background(),
		msg,
		&user,
	)

	currentCaravanRobbers := checkPlayersCount(w.robbers)
	updatedMarkup := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf(
					"Погнааале [ %d / %d ]",
					currentCaravanRobbers,
					caravanPlayersToStart,
				),
				"join",
			),
		),
	)

	if (w.lastCaravanMessageWithCallback.From != nil) && (currentCaravanRobbers != robbersBefore) {
		edited, err := w.c.EditMessageMarkup(
			w.lastCaravanMessageWithCallback,
			&updatedMarkup,
		)

		if err != nil {
			w.errors <- errors.Wrap(err, "cannot edit last caravan message")
		} else {
			w.lastCaravanMessageWithCallback = &edited
		}
	}

	if err := w.c.SendInlineKeyboardReply(update.CallbackQuery, reply); err != nil {
		w.errors <- errors.Wrap(err, "cannot send inline keyboard reply from caravan callback")
		return
	}
}

func (w *Wars) GetTopPlayers(ctx context.Context, count int) string {
	result := "*ТОП ИГРОКОВ:*\n"
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
	usersCount, err := w.db.UserStorage.CountAllUsers(ctx)
	if err != nil {
		w.errors <- errors.Wrap(err, "cannot count all users in /top")
	}

	result += fmt.Sprintf(
		"\nВсего в участвуют: %d игроков!\n_Региструйся и победи всех_ */reg*",
		usersCount,
	)
	return result
}

func checkEqualKeyboards(k1 *tgbotapi.InlineKeyboardMarkup, k2 *tgbotapi.InlineKeyboardMarkup) bool {
	return true
}

func (w *Wars) SendShopWithKeyboard(ctx context.Context, msg *tgbotapi.Message, user *database.User) {
	weapons, err := w.db.WeaponStorage.GetWeaponsLimitOffset(ctx, shopItemsPerpage, 0)
	if err != nil {
		w.errors <- errors.Wrap(err, "cannot get weapons")
		return
	}

	if len(weapons) == 0 {
		if err := w.c.SendReply(msg, "Ты не можешь ничего купить, у тебя мало денег!"); err != nil {
			w.errors <- errors.Wrap(err, "cannot send can buy anything in shop")
		}
		return
	}

	reply := tgbotapi.NewMessage(msg.Chat.ID, "*Уютный shop 🛒 *")
	reply.ParseMode = "markdown"

	keyboard, err := w.GenerateWeaponsKeyboard(weapons, msg.From.ID, 0)
	if err != nil {
		w.errors <- errors.Wrap(err, "cannot generate shop weapons keyboard")
		return
	}

	reply.ReplyMarkup = keyboard
	_, err = w.c.Send(&reply)

	if err != nil {
		w.errors <- errors.Wrap(err, "cannot send newshop reply message")
	}
}

func (w *Wars) GenerateWeaponsKeyboard(weapons []database.Weapon, userID int, offset int) (tgbotapi.InlineKeyboardMarkup, error) {
	// count all weapons items to create pagination
	weaponsCount, err := w.db.WeaponStorage.GetWeaponsCount(context.Background())
	weaponsCount += (weaponsCount / 5) % 5
	keyboard := tgbotapi.InlineKeyboardMarkup{}
	if err != nil {
		return keyboard, err
	}

	for _, w := range weapons {
		var row []tgbotapi.InlineKeyboardButton
		buttonText := fmt.Sprintf("%s %d🏹️, %d💰\n", w.Name, w.Power, w.Cost)
		buttonClass := fmt.Sprintf("shop%d", w.ID)
		btn := tgbotapi.NewInlineKeyboardButtonData(buttonText, buttonClass)
		row = append(row, btn)
		keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, row)
	}
	var buttonNavRow []tgbotapi.InlineKeyboardButton

	// generate button <-, -> if need
	if offset != 0 {
		prevBtn := tgbotapi.NewInlineKeyboardButtonData("<- Сюда", fmt.Sprintf("p%d %d", offset, userID))
		buttonNavRow = append(buttonNavRow, prevBtn)
	}

	if weaponsCount-shopItemsPerpage > offset {
		nextBtn := tgbotapi.NewInlineKeyboardButtonData("Туда ->", fmt.Sprintf("n%d %d", offset, userID))
		buttonNavRow = append(buttonNavRow, nextBtn)
	}

	if len(buttonNavRow) > 0 {
		keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, buttonNavRow)
	}
	return keyboard, nil
}

func (w *Wars) HandleNewShopCallbackQuery(update *tgbotapi.Update) {
	ctx := context.Background()
	re := regexp.MustCompile(`shop(\d+)`)
	match := re.FindStringSubmatch(update.CallbackQuery.Data)

	if len(match) == 2 {
		weaponID, err := strconv.Atoi(match[1])
		if err != nil {
			w.errors <- errors.Wrap(err, "cannot convert buy match[2] to integer")
			return
		}

		reply := w.buyFromCallback(ctx, update.CallbackQuery.From.ID, weaponID)
		err = w.c.SendInlineKeyboardReply(update.CallbackQuery, reply)
		if err != nil {
			w.errors <- errors.Wrap(err, "cannot send inline callback reply from shop")
			return
		}
		return
	}

	user, err := w.db.UserStorage.FindUserByID(ctx, update.CallbackQuery.From.ID)
	if err != nil {
		if err := w.c.SendInlineKeyboardReply(update.CallbackQuery, "Ты не зарегистрирован!"); err != nil {
			w.errors <- errors.Wrap(err, "cannot send inline shop reply to not reg user")
		}
	}

	if user.Blacklist {
		if err := w.c.SendInlineKeyboardReply(update.CallbackQuery, "Ты забанен!"); err != nil {
			w.errors <- errors.Wrap(err, "cannot send inline shop reply to ban user")
		}
	}
	re = regexp.MustCompile(`([p|n])(\d+) (\d+)`)
	match = re.FindStringSubmatch(update.CallbackQuery.Data)

	if len(match) == 4 {
		direction := match[1]
		currentOffset, err := strconv.Atoi(match[2])
		if err != nil {
			w.errors <- errors.Wrap(err, "cannot parse regexp offset in inline button callback")
			return
		}
		userID, err := strconv.Atoi(match[3])
		if err != nil {
			w.errors <- errors.Wrap(err, "cannot parse regexp in userID inline button callback")
			return
		}

		if update.CallbackQuery.From.ID != userID {
			if err := w.c.SendInlineKeyboardReply(update.CallbackQuery, "Это не ваш открытый магазин, чтобы его листать!"); err != nil {
				w.errors <- errors.Wrap(err, "cannot send inline reply to callback query")
			}
			return
		}

		switch direction {
		case "n":
			currentOffset += shopItemsPerpage
		case "p":
			currentOffset -= shopItemsPerpage
			if currentOffset < 0 {
				currentOffset = 0
			}
		}

		weapons, err := w.db.WeaponStorage.GetWeaponsLimitOffset(ctx, shopItemsPerpage, currentOffset)
		if err != nil {
			w.errors <- errors.Wrap(err, "cannot get weapons with limit and offset")
			return
		}

		keyboard, err := w.GenerateWeaponsKeyboard(weapons, update.CallbackQuery.From.ID, currentOffset)
		if err != nil {
			w.errors <- errors.Wrap(err, "cannot generate shop weapons keyboard")
			return
		}

		_, err = w.c.EditMessageMarkup(update.CallbackQuery.Message, &keyboard)
		if err != nil {
			// Ignore this, because if fastclick button, happens lag
			// w.errors <- errors.Wrap(err, "cannot get weapons with limit and offset")
			// 5 min ban if spamming
			go w.getTimeBan(update.CallbackQuery.Message, update.CallbackQuery.From.UserName, 5)
			if err := w.c.DeleteMessage(update.CallbackQuery.Message); err != nil {
				w.errors <- errors.Wrap(err, "cannot delete message, for fastclicking")
			}

			return
		}
	}
}

func (w *Wars) getTimeBan(msg *tgbotapi.Message, username string, bantime int) {
	if err := w.db.UserStorage.SwitchBanUser(context.Background(), username, true); err != nil {
		w.errors <- errors.Wrap(err, "cannot ban user fastclicking")
	}

	err := w.c.SendMessageToUWDChat(
		fmt.Sprintf(
			"@%s, ты забанен на %d минут, за фасткликинг по кнопкам. Остудись, я напишу как разбаню тебя!",
			GetMarkdownUsername(username),
			bantime,
		),
	)

	if err != nil {
		w.errors <- errors.Wrap(err, "cannot send ban message to user fastclicking")
	}

	timer1 := time.NewTimer(time.Minute * time.Duration(bantime))
	<-timer1.C

	if err := w.db.UserStorage.SwitchBanUser(context.Background(), username, false); err != nil {
		w.errors <- errors.Wrap(err, "cannot ban user fastclicking")
	}

	err = w.c.SendMessageToUWDChat(
		fmt.Sprintf(
			"@%s, я тебя разбанил, возвращайся!",
			username,
		),
	)

	if err != nil {
		w.errors <- errors.Wrap(err, "cannot send unban message to user fastclicking")
	}
}

func (w *Wars) buyFromCallback(ctx context.Context, userID int, item int) string {
	user, err := w.db.UserStorage.FindUserByID(ctx, userID)
	if err != nil {
		return "Ошибочка вышла"
	}

	if user.ID == 0 {
		return "Сначала /reg"
	}

	weapon, err := w.db.WeaponStorage.GetWeaponsByID(ctx, item)
	if err != nil {
		return "Ошибочка вышла"
	}

	if user.Coins >= weapon.Cost {
		if err := w.db.UserStorage.DecreaseMoney(ctx, user.UserID, weapon.Cost); err != nil {
			w.errors <- errors.Wrap(err, "cannot decrease money")
			return "Не могу списать монеты"
		}

		if err := w.db.UserStorage.AddPower(ctx, int(user.UserID), weapon.Power); err != nil {
			w.errors <- errors.Wrap(err, "cannot add power")
			return "Не могу добавить мощности"
		}

		return fmt.Sprintf("Куплено: %s", weapon.Name)

	}

	return fmt.Sprintf("Вам не хватает %d💰!", weapon.Cost-user.Coins)
}

// Arena gameplay
func (w *Wars) RegisterToArena(ctx context.Context, msg *tgbotapi.Message, user *database.User) string {
	if w.arenaProgress {
		return "🥊💪🥊"
	}

	if user.Activity <= 0 {
		return "Не хватает очков активности!"
	}

	if checkPlayerByID(w.arenaPlayers, user.UserID) {
		return "Ты уже записался на арену!"
	}

	if err := w.db.UserStorage.DecreaseActivity(ctx, int(user.UserID)); err != nil {
		w.errors <- errors.Wrap(err, "cannot decrease arena activity")
		return ""
	}

	arenaPlayersCount := checkPlayersCount(w.arenaPlayers)
	w.arenaPlayers[arenaPlayersCount] = Player{
		user.ID, user.UserID, user.Username, user.WeaponsPower, user.Reputation, user.Coins,
	}

	arenaPlayersCount = checkPlayersCount(w.arenaPlayers)
	if arenaPlayersCount == arenaPlayersToStart {
		go w.startArenaFight(ctx, msg)
		return ""
	}

	return "Ты успешно записался на арену!"
}

func (w *Wars) startArenaFight(ctx context.Context, msg *tgbotapi.Message) {
	w.arenaProgress = true
	ids := getPlayersIDs(w.arenaPlayers)

	err := w.c.SendMarkdownReply(
		msg,
		fmt.Sprintf(
			"Начинаем бой между @%s, @%s!",
			GetMarkdownUsername(w.arenaPlayers[0].Username),
			GetMarkdownUsername(w.arenaPlayers[1].Username),
		),
	)
	if err != nil {
		w.errors <- errors.Wrap(err, "cannot send start arena message")
	}

	for i := range w.arenaPlayers {
		// generate +-50 percents to player power
		randPowerSupply := rand.Intn(2) + rand.Intn(51)/100
		w.arenaPlayers[i].Power *= randPowerSupply
	}

	winner, looser := Player{}, Player{}
	if w.arenaPlayers[0].Power > w.arenaPlayers[1].Power {
		winner = w.arenaPlayers[0]
		looser = w.arenaPlayers[1]
	}
	if w.arenaPlayers[1].Power > w.arenaPlayers[0].Power {
		winner = w.arenaPlayers[1]
		looser = w.arenaPlayers[0]
	}

	timeLeft := 1 + rand.Intn(arenaRoundMaxTime)
	timer1 := time.NewTimer(time.Minute * time.Duration(timeLeft))
	<-timer1.C

	if winner.ID != 0 {
		looser10PercentCoins := int(float32(looser.Coins) * 0.1)
		looser5PercentReputation := int(float32(looser.Reputation) * 0.05)
		earnMoney := minArenaMoney
		earnReputation := 5

		if looser10PercentCoins > minArenaMoney {
			earnMoney = looser10PercentCoins
		}
		if looser5PercentReputation > minArenaReputation {
			earnReputation = looser5PercentReputation
		}

		// Add reputation and coins to winner
		if err := w.db.UserStorage.AddMoney(ctx, winner.UserID, earnMoney); err != nil {
			w.errors <- errors.Wrap(err, "cant add money to arena winner")
			return
		}
		if err := w.db.UserStorage.AddReputation(ctx, winner.UserID, earnReputation); err != nil {
			w.errors <- errors.Wrap(err, "cant add reputation to arena winner")
			return
		}

		// Decrese money to looser
		decreaseMoney := earnMoney
		if looser.Coins < decreaseMoney {
			decreaseMoney = looser.Coins
		}
		if err := w.db.UserStorage.DecreaseMoney(ctx, looser.UserID, decreaseMoney); err != nil {
			w.errors <- errors.Wrap(err, "cant decrease money to arena looser")
			return
		}

		// SendReply to winner
		err = w.c.SendMarkdownReply(
			msg,
			fmt.Sprintf(
				"@*%s* победил в этом бое. Ему начислено *%d* монет и *%d* репутации. Проигравшему @%s снято *%d* монет.",
				winner.Username,
				earnMoney,
				earnReputation,
				GetMarkdownUsername(looser.Username),
				decreaseMoney,
			),
		)

		if err != nil {
			w.errors <- errors.Wrap(err, "cannt send message to winner")
			return
		}

	} else {
		drawMoney := w.arenaPlayers[0].Coins
		if w.arenaPlayers[1].Coins < drawMoney {
			drawMoney = w.arenaPlayers[1].Coins
		}
		drawMoney = int(float32(drawMoney) * 0.1)
		if drawMoney < minArenaMoney {
			drawMoney = minArenaMoney
		}
		// If draw add 10% (low level coins player) to users
		if err := w.db.UserStorage.AddMoneyToUsers(ctx, drawMoney, ids); err != nil {
			w.errors <- errors.Wrap(err, "cannot add money arena to users")
			return
		}

		replyString := fmt.Sprintf(
			"Бой: *@%s*, *@%s* был равным, им начислено *%d* монеток!",
			w.arenaPlayers[0].Username,
			w.arenaPlayers[1].Username,
			drawMoney,
		)
		if err := w.c.SendMarkdownReply(msg, replyString); err != nil {
			w.errors <- errors.Wrap(err, "cannot send draw arena message")
			return
		}
	}

	// Clean arena players to next round
	for i := 0; i < arenaPlayersToStart; i++ {
		w.arenaPlayers[i] = Player{}
	}

	w.arenaProgress = false
}
