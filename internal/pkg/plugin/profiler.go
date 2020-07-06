package plugin

import (
	"context"
	"fmt"
	"strings"

	"github.com/LikiPiki/UwdBot/internal/pkg/database"
	"github.com/pkg/errors"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type Rank struct {
	Rank       string
	Reputation int
}

func GetMarkdownUsername(username string) string {
	return strings.ReplaceAll(username, "_", "\\_")
}

func (p *Profiler) getRank(user database.User) string {
	for _, rank := range p.ranks {
		if rank.Reputation <= user.Reputation {
			return rank.Rank
		}
	}
	return p.ranks[len(p.ranks)-1].Rank
}

func (p *Profiler) unregUser(ctx context.Context, msg *tgbotapi.Message) (string, error) {
	if _, err := p.db.UserStorage.DeleteUser(ctx, msg.From.ID); err != nil {
		return "", errors.Wrap(err, "cannot unreg user")
	}

	return "Ну заходи как нибудь еще, что делать...", nil
}

func (p *Profiler) showUserInfo(ctx context.Context, msg *tgbotapi.Message) (string, error) {
	user, err := p.db.UserStorage.FindUserByID(ctx, msg.From.ID)
	if err != nil {
		return "", errors.Wrap(err, "cannot find user by id")
	}

	repStat, coinsStat, err := p.db.UserStorage.GetUserStatistics(ctx, user.Reputation, user.Coins)
	if err != nil {
		return "", errors.Wrap(err, "cannot get user stat")
	}

	rank := p.getRank(user)

	return fmt.Sprintf(
		`
***ЛИЧНАЯ КАРТОЧКА***
Привет ***@%s*** - ___%s___
Твоя репутация 👑: ***%d
***Монеты 💰: ***%d***
Боевая мощь 🏹: ***%d***
На сегодня осталось ***%d*** единиц активности!

Ты на ***%d***%% круче остальных и на ***%d***%% богаче!
`,
		user.Username,
		rank,
		user.Reputation,
		user.Coins,
		user.WeaponsPower,
		user.Activity,
		repStat*100,
		coinsStat*100,
	), nil
}

func (p *Profiler) AddMoneyByUsername(ctx context.Context, money int, username string) (string, error) {
	user, err := p.db.UserStorage.FindUserByUsername(ctx, username)
	if err != nil {
		return fmt.Sprintf(
			"Пользоваателя %s не существует!",
			GetMarkdownUsername(username),
		), nil
	}

	if err := p.db.UserStorage.AddMoney(ctx, user.ID, money); err != nil {
		return "", errors.Wrap(err, "cannot add money to user")
	}
	return fmt.Sprintf(
		"Пользователю **@%s** начислено **%d💰**",
		GetMarkdownUsername(username),
		money,
	), nil
}

func (p *Profiler) registerNewUser(ctx context.Context, msg *tgbotapi.Message) string {
	count, err := p.db.UserStorage.CountUsersWithID(ctx, msg.From.ID)
	if err != nil {
		return "Что то пошло не так..."
	}

	if count > 0 {
		return "Ты уже зарегистрирован!"
	}

	if len(msg.From.UserName) == 0 {
		return "Для начала придумай себе nickname в Телеграме! "
	}

	_, err = p.db.UserStorage.CreateNewUser(ctx, msg.From.UserName, uint64(msg.From.ID))
	if err != nil {
		return "Не удалось добавить. Попробуй позже..."
	}

	return "Вы успешно прошли регистрацию. /me"
}
