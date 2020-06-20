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
				"Списано ***%d***💰, куплен(а): ___%s___!\n\n Прибавлено %d к боевой мощи!",
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
