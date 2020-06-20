package plug

import (
	data "UwdBot/database"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

const (
	poll_api_url = "https://engine.lifeis.porn/api/millionaire.php"
)

func (m *Minigames) sendCasinoMiniGame(msg *tgbotapi.Message, user *data.User) {
	if user.Coins < 10 {
		m.c.SendReplyToMessage(msg, "Слишком мало денег, накопи еще и приходи потом!")
		return
	}
	user.DecreaseMoney(10)
	miniGame, status := generateCasino()
	m.c.SendReply(msg, miniGame)
	if status {
		user.AddMoney(100)
		m.c.SendReplyToMessage(msg, "Уважаемый, вы победили... + 50💰")
	}
}

func generateCasino() (string, bool) {
	icons := []string{
		"🚑", "🎡", "💊", "🐵", "🍒", "🍾",
	}
	iconsNum := []int{0, 0, 0}
	var win string
	status := false
	for i := 0; i < 3; i++ {
		iconsNum[i] = rand.Intn(len(icons))
		win = win + icons[iconsNum[i]]
	}
	// check winner
	if (iconsNum[0] == iconsNum[1]) && (iconsNum[1] == iconsNum[2]) {
		status = true
	}
	return win, status
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

func (p *Poll) GetSuccess() string {
	return p.Data.Answers[p.Data.Valid]
}

func (p *Poll) HaveMember(name string) int {
	return p.members[name]
}

func (p *Poll) GetPollResults(winner string, winnerID int) string {
	result := fmt.Sprintf(
		"`%s`\nПравильный ответ - ___%s___.\nОтветил - @%s",
		p.Message.Text,
		p.GetSuccess(),
		winner,
	)

	var user data.User
	var err error

	user, err = user.FindUserByID(winnerID)
	if (err == nil) && (user.ID > 0) {
		money := rand.Intn(11)
		result += fmt.Sprintf(
			" + %d💰",
			money,
		)
		user.AddMoney(money)
	}

	if len(p.members) > 1 {
		result += fmt.Sprintf(
			"\nПытались: ___%s___",
			p.getAllMembersUsernamesString(winner),
		)
	}
	return result
}

func (p *Poll) getAllMembersUsernamesString(winner string) (result string) {
	first := true
	for username := range p.members {
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

func (p *Poll) AddMember(name string) {
	p.members[name]++
}

func (p *Poll) Shuffle() {
	if len(p.Data.Answers) == 0 {
		return
	}

	valid := p.Data.Answers[0]

	for i := 0; i < len(p.Data.Answers); i++ {
		random := rand.Intn(len(p.Data.Answers))
		tmp := p.Data.Answers[random]
		p.Data.Answers[random] = p.Data.Answers[i]
		p.Data.Answers[i] = tmp
	}

	for k, ans := range p.Data.Answers {
		if ans == valid {
			p.Data.Valid = k
		}
	}
}

// func (m *Minigames) IsAdmin(username string) bool {
// 	return a.admins[username]
// }

func (m *Minigames) UpdatePollMessage(id int, msg *tgbotapi.Message) {
	if len(m.Polls) > id {
		m.Polls[id].Message = msg
	} else {
		log.Println("Invalid id")
	}
}

func (m *Minigames) LoadPoll() Poll {
	req, err := http.NewRequest("GET", poll_api_url, nil)
	if err != nil {
		log.Println(err)
	}

	poll := Poll{}
	poll.members = make(map[string]int)

	q := req.URL.Query()
	q.Add("count", "1")
	value := rand.Intn(3) + 1
	q.Add("q", strconv.Itoa(value))
	req.URL.RawQuery = q.Encode()

	resp, err := http.Get(req.URL.String())

	if err != nil {
		log.Println(err)
	}

	defer resp.Body.Close()

	jsonCode, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
	}

	err = json.Unmarshal(jsonCode, &poll)
	if err != nil {
		log.Println(err)
	}
	poll.Shuffle()
	poll.Data.Time = time.Now()
	poll.Data.Solved = false
	return poll
}

func (m *Minigames) GetPoll() int {
	poll := m.LoadPoll()
	if len(m.Polls) < LEN {
		m.Polls = append(m.Polls, poll)
		return len(m.Polls) - 1
	} else {
		for id, current := range m.Polls {
			if current.Data.Solved == true {
				m.Polls[id] = poll
				return id
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
	return index
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
