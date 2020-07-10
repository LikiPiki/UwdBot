package plugin

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
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
func GetItalicUnderlineUsername(username string) string {
	return strings.ReplaceAll(username, "_", "_\\__")
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
*ЛИЧНАЯ КАРТОЧКА*
Привет *@%s* - _%s_
Твоя репутация 👑: *%d
*Монеты 💰: *%d*
Боевая мощь 🏹: *%d*
На сегодня осталось *%d* единиц активности!

Ты на *%.1f*%% круче остальных и на *%.1f*%% богаче!
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

	if err := p.db.UserStorage.AddMoney(ctx, user.UserID, money); err != nil {
		return "", errors.Wrap(err, "cannot add money to user")
	}
	return fmt.Sprintf(
		"Пользователю *@%s* начислено *%d💰*",
		username,
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

func (p *Profiler) HandleAdminRegexpCommands(msg *tgbotapi.Message) {
	// Add money case
	re := regexp.MustCompile(`^[a|A]ddmoney (\\d+) (\\w+)$`)
	match := re.FindStringSubmatch(msg.Text)
	if len(match) == 3 {
		itemNumber, err := strconv.Atoi(match[1])
		if err != nil {
			err = p.c.SendMarkdownReply(msg, "Команда введена не верно, пробуй ``/addmoney 100 username``")
			if err != nil {
				p.errors <- errors.Wrap(err, "cannot send wrong command reply")
			}
			return
		}

		text, err := p.AddMoneyByUsername(context.Background(), itemNumber, match[2])
		if err != nil {
			p.errors <- errors.Wrap(err, "cannot add money by username")
			return
		}
		if err := p.c.SendMarkdownReply(
			msg,
			text,
		); err != nil {
			p.errors <- errors.Wrap(err, "cannot send MD reply")
		}
		return
	}

	// Ban user
	re = regexp.MustCompile(`^[b|B]an (\\w+)`)
	match = re.FindStringSubmatch(msg.Text)

	if len(match) == 2 {
		if err := p.db.UserStorage.SwitchBanUser(context.Background(), match[1], true); err != nil {
			p.errors <- errors.Wrap(err, "cannot ban user")
		}

		reply := fmt.Sprintf("Пользователь @%s забанен!", match[1])
		if err := p.c.SendReply(msg, reply); err != nil {
			p.errors <- errors.Wrap(err, "cannot send reply")
		}

		return
	}

	// Unban user
	re = regexp.MustCompile(`^[u|U]nban (\\w+)`)
	match = re.FindStringSubmatch(msg.Text)

	if len(match) == 2 {
		if err := p.db.UserStorage.SwitchBanUser(context.Background(), match[1], false); err != nil {
			p.errors <- errors.Wrap(err, "cannot unban user")
		}

		reply := fmt.Sprintf("Пользователь @%s разабанен!", match[1])
		if err := p.c.SendReply(msg, reply); err != nil {
			p.errors <- errors.Wrap(err, "cannot send reply")
		}

		return
	}
}
