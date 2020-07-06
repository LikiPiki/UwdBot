package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/LikiPiki/UwdBot/internal/pkg/database"
	"github.com/pkg/errors"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

const (
	pollApiUrl = "https://engine.lifeis.porn/api/millionaire.php"
)

func (m *Minigames) sendCasinoMiniGame(ctx context.Context, msg *tgbotapi.Message, user *database.User) {
	if user.Coins < 10 {
		if err := m.c.SendReplyToMessage(msg, "Слишком мало денег, накопи еще и приходи потом!"); err != nil {
			m.errors <- errors.Wrap(err, "cannot send reply")
		}

		return
	}

	if err := m.db.UserStorage.DecreaseMoney(ctx, user.ID, 10); err != nil {
		m.errors <- errors.Wrap(err, "cannot decrease money")
		return
	}

	miniGame, moneys := generateCasino()
	reply := tgbotapi.NewMessage(
		msg.Chat.ID,
		miniGame,
	)

	sent, err := m.c.Send(&reply)
	if err != nil {
		m.errors <- errors.Wrap(err, "cannot send reply")
		return
	}

	if (m.LastCasinoMessage.From != nil) && (sent.From.ID == m.LastCasinoMessage.From.ID) {
		if err := m.c.DeleteMessage(m.LastCasinoMessage); err != nil {
			m.errors <- errors.Wrap(err, "cannot delete message")
			return
		}
	}

	m.LastCasinoMessage = sent

	if moneys != 0 {
		if err := m.db.UserStorage.AddMoney(ctx, user.ID, moneys); err != nil {
			m.errors <- errors.Wrap(err, "cannot add money")
			return
		}

		if err := m.c.SendReplyToMessage(msg, fmt.Sprintf("Уважаемый, вы победили... + %d💰", moneys)); err != nil {
			m.errors <- errors.Wrap(err, "cannot send reply")
		}
	}
}

func generateCasino() (string, int) {
	icons := []string{
		"🚑", "🎡", "💊", "🐵", "🍒", "🍾", "🥒", "🦄",
	}
	iconsNum := make([]int, 3)
	var win string

	for i := 0; i < 3; i++ {
		iconsNum[i] = rand.Intn(len(icons))
		win = win + icons[iconsNum[i]]
	}
	// check winner
	if (iconsNum[0] == iconsNum[1]) && (iconsNum[1] == iconsNum[2]) {
		return win, 100
	}
	if (iconsNum[0] == iconsNum[1]) || (iconsNum[1] == iconsNum[2]) || (iconsNum[0] == iconsNum[2]) {
		return win, 30
	}
	return win, 0
}

// Polls function
type QuestionsData struct {
	Question string   `json:"question"`
	Valid    int      `json:"valid"`
	Users    []string `json:"users"`
	Answers  []string `json:"answers"`
	Time     time.Time
	Solved   bool
}

type Poll struct {
	Ok      bool          `json:"ok"`
	Data    QuestionsData `json:"data"`
	Message *tgbotapi.Message
	members map[string]int
}

func (m *Minigames) GetSuccess(poll Poll) string {
	return poll.Data.Answers[poll.Data.Valid]
}

func (p *Poll) HaveMember(name string) int {
	return p.members[name]
}

func (m *Minigames) GetPollResults(ctx context.Context, winner string, winnerID int, poll Poll) (string, error) {
	result := fmt.Sprintf(
		"`%s`\nПравильный ответ - __%s__.\nОтветил - @%s",
		poll.Message.Text,
		m.GetSuccess(poll),
		GetMarkdownUsername(winner),
	)

	user, err := m.db.UserStorage.FindUserByID(ctx, winnerID)
	if err != nil {
		return "", errors.Wrap(err, "cannot find user by id")
	}

	if user.ID > 0 {
		money := rand.Intn(11)
		result += fmt.Sprintf(
			" + %d💰",
			money,
		)
		if err := m.db.UserStorage.AddMoney(ctx, uint64(winnerID), money); err != nil {
			return "", errors.Wrap(err, "cannot add money to user")
		}
	}

	if len(poll.members) > 1 {
		result += fmt.Sprintf(
			"\nПытались: __%s__",
			GetMarkdownUsername(m.getAllMembersUsernamesString(winner, poll)),
		)
	}

	return result, nil
}

func (m *Minigames) getAllMembersUsernamesString(winner string, poll Poll) (result string) {
	first := true
	for username := range poll.members {
		if username == winner {
			continue
		}
		if !first {
			result += fmt.Sprintf(", %s", username)
		} else {
			result += username
			first = false
		}
	}
	return result
}

func (m *Minigames) AddMember(name string, poll Poll) {
	poll.members[name]++
}

func (m *Minigames) Shuffle(poll Poll) Poll {
	if len(poll.Data.Answers) == 0 {
		return Poll{}
	}

	valid := poll.Data.Answers[0]

	for i := 0; i < len(poll.Data.Answers); i++ {
		random := rand.Intn(len(poll.Data.Answers))
		tmp := poll.Data.Answers[random]
		poll.Data.Answers[random] = poll.Data.Answers[i]
		poll.Data.Answers[i] = tmp
	}

	for k, ans := range poll.Data.Answers {
		if ans == valid {
			poll.Data.Valid = k
		}
	}

	return poll
}

func (m *Minigames) UpdatePollMessage(id int, msg *tgbotapi.Message) error {
	if len(m.Polls) > id {
		m.Polls[id].Message = msg
	} else {
		return errors.Errorf("invalid id: %d", id)
	}

	return nil
}

func (m *Minigames) LoadPoll() (Poll, error) {
	req, err := http.NewRequest("GET", pollApiUrl, nil)
	if err != nil {
		return Poll{}, errors.Wrap(err, "cannot perform HTTP GET request")
	}

	var poll Poll
	poll.members = make(map[string]int)

	q := req.URL.Query()
	q.Add("count", "1")
	value := rand.Intn(3) + 1
	q.Add("q", strconv.Itoa(value))
	req.URL.RawQuery = q.Encode()

	resp, err := http.Get(req.URL.String())
	if err != nil {
		return Poll{}, errors.Wrap(err, "cannot perform HTTP GET request")
	}
	defer resp.Body.Close()

	jsonCode, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Poll{}, errors.Wrap(err, "cannot read HTTP body")
	}

	err = json.Unmarshal(jsonCode, &poll)
	if err != nil {
		return Poll{}, errors.Wrap(err, "cannot unmarshal json")
	}

	poll = m.Shuffle(poll)
	poll.Data.Time = time.Now()
	poll.Data.Solved = false

	return poll, nil
}

func (m *Minigames) GetPoll() (int, error) {
	poll, err := m.LoadPoll()
	if err != nil {
		return 0, errors.Wrap(err, "cannot load poll")
	}

	if len(m.Polls) < LEN {
		m.Polls = append(m.Polls, poll)
		return len(m.Polls) - 1, nil
	} else {
		for id, current := range m.Polls {
			if current.Data.Solved == true {
				m.Polls[id] = poll
				return id, nil
			}
		}
	}

	minTime, index := m.Polls[0].Data.Time, 0
	for id, current := range m.Polls {
		if current.Data.Time.Sub(minTime).String()[0] == '-' {
			minTime = current.Data.Time
			index = id
		}
	}

	m.Polls[index] = poll

	return index, nil
}

func (m *Minigames) CheckNumberQuestions(num, ans int) bool {
	if (len(m.Polls) > num) && (ans < 4) {
		return true
	}
	return false
}

func (m *Minigames) SolvePoll(num, ans int) {
	m.Polls[num].Data.Solved = true
}

func (m *Minigames) CheckPoll(num, ans int) (bool, bool) {
	if m.Polls[num].Data.Valid == ans {
		return true, m.Polls[num].Data.Solved
	}
	return false, m.Polls[num].Data.Solved
}
