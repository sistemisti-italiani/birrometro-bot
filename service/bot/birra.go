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

func birraHandler(bot *telebot.Bot, logger *logrus.Entry, db database.AppDatabase, cfg config.APPConfig,
	shutdown chan os.Signal) func(m *telebot.Message) {
	return func(m *telebot.Message) {
		err := updateChatUserInfo(db, m)
		if err != nil {
			logger.WithError(err).Error("error updating chat user info in /birra")
			shutdown <- syscall.SIGTERM
			return
		}

		if !isAuthorized(m, cfg) {
			return
		} else if m.Private() {
			_, err = bot.Send(m.Chat, "Oops, questo comando è disponibile solo nelle chat di gruppo!")
			if err != nil {
				logger.WithError(err).Error("error sending message in /birra")
				shutdown <- syscall.SIGTERM
				return
			}
		} else if m.ReplyTo == nil {
			_, err = bot.Send(m.Chat, "Usa questo comando citando un messaggio della persona a cui devi la birra!")
			if err != nil {
				logger.WithError(err).Error("error sending message in /birra")
				shutdown <- syscall.SIGTERM
				return
			}
		} else if m.ReplyTo.Sender.IsBot {
			_, err = bot.Send(m.Chat, "Non puoi dare una birra ad un bot -.-")
			if err != nil {
				logger.WithError(err).Error("error sending message in /birra")
				shutdown <- syscall.SIGTERM
				return
			}
		} else if m.ReplyTo.Sender.ID == m.Sender.ID {
			_, err = bot.Send(m.Chat, "Usa questo comando citando un messaggio della persona a cui devi la birra!")
			if err != nil {
				logger.WithError(err).Error("error sending message in /birra")
				shutdown <- syscall.SIGTERM
				return
			}
		} else {
			err := db.AddOrUpdateUserInfo(int64(m.ReplyTo.Sender.ID), m.ReplyTo.Sender.Username, m.ReplyTo.Sender.FirstName, m.ReplyTo.Sender.LastName)
			if err != nil {
				logger.WithError(err).Error("error updating user info in /birra")
				shutdown <- syscall.SIGTERM
				return
			}

			err = db.AddBeer(int64(m.Sender.ID), int64(m.ReplyTo.Sender.ID))
			if err != nil {
				logger.WithError(err).Error("error adding beer in /birra")
				shutdown <- syscall.SIGTERM
				return
			}

			var senderName = m.ReplyTo.Sender.Username
			if senderName == "" {
				senderName = m.Sender.FirstName + " " + m.Sender.LastName
			}
			_, err = bot.Send(m.Chat, fmt.Sprintf("Nuova birra per %s!", strings.TrimSpace(senderName)))
			if err != nil {
				logger.WithError(err).Error("error sending message in /birra")
				shutdown <- syscall.SIGTERM
				return
			}
		}
	}
}
