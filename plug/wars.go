package plug

import (
	"fmt"

	data "UwdBot/database"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

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
	reply += "\n___Интересный стафф:___\nПоявится в скором времени...\n\n___Купить товар /buy номер товара___"
	return reply
}
