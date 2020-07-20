package bot

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/sistemisti-italiani/birrometro_bot/service/config"
	"github.com/sistemisti-italiani/birrometro_bot/service/database"
	"gopkg.in/tucnak/telebot.v2"
	"os"
	"strings"
	"syscall"
)

func birreHandler(bot *telebot.Bot, logger *logrus.Entry, db database.AppDatabase, cfg config.APPConfig,
	shutdown chan os.Signal) func(m *telebot.Message) {
	return func(m *telebot.Message) {
		err := updateChatUserInfo(db, m)
		if err != nil {
			logger.WithError(err).Error("error updating chat user info in /birre")
			shutdown <- syscall.SIGTERM
			return
		}

		if !isAuthorized(m, cfg) {
			return
		} else if m.Private() {
			_, err = bot.Send(m.Chat, "Oops, questo comando è disponibile solo nelle chat di gruppo!")
			if err != nil {
				logger.WithError(err).Error("error sending message in /birre")
				shutdown <- syscall.SIGTERM
				return
			}
		} else {
			beerDebts, err := db.BeerDebts(int64(m.Sender.ID))
			if err != nil {
				logger.WithError(err).Error("error getting beer debs infos in /birre")
				shutdown <- syscall.SIGTERM
				return
			}

			var msg strings.Builder
			msg.WriteString("Le birre che devi ")
			msg.WriteString(SYMBOL_BEERS)
			msg.WriteString(":\n\n")
			if len(beerDebts) == 0 {
				msg.WriteString("Nessuna!!!\n")
			}
			for uid, beers := range beerDebts {
				username, err := db.GetUserName(uid)
				if err != nil {
					logger.WithError(err).Error("error getting username in /birre")
					shutdown <- syscall.SIGTERM
					return
				}
				msg.WriteString(username)
				msg.WriteString(": ")
				for i := 0; i < beers; i++ {
					msg.WriteString(SYMBOL_BEER)
				}
				msg.WriteString(fmt.Sprintf(" (%d)\n", beers))
			}

			beerCreds, err := db.BeerCreds(int64(m.Sender.ID))
			if err != nil {
				logger.WithError(err).Error("error getting beer creds infos in /birre")
				shutdown <- syscall.SIGTERM
				return
			}

			msg.WriteString("\nLe birre che ti devono ")
			msg.WriteString(SYMBOL_BEERS)
			msg.WriteString(":\n\n")
			if len(beerCreds) == 0 {
				msg.WriteString("Nessuna :-(\n")
			}
			for uid, beers := range beerCreds {
				username, err := db.GetUserName(uid)
				if err != nil {
					logger.WithError(err).Error("error getting username in /birre")
					shutdown <- syscall.SIGTERM
					return
				}
				msg.WriteString(username)
				msg.WriteString(": ")
				for i := 0; i < beers; i++ {
					msg.WriteString(SYMBOL_BEER)
				}
				msg.WriteString(fmt.Sprintf(" (%d)\n", beers))
			}

			_, err = bot.Send(m.Chat, msg.String())
			if err != nil {
				logger.WithError(err).Error("error sending message in /birre")
				shutdown <- syscall.SIGTERM
				return
			}
		}
	}
}
