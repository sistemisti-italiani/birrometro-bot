package bot

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/sistemisti-italiani/birrometro_bot/service/config"
	"github.com/sistemisti-italiani/birrometro_bot/service/database"
	"gopkg.in/tucnak/telebot.v2"
	"os"
	"syscall"
)

const (
	SYMBOL_BEER      = "🍺"
	SYMBOL_BEERS     = "🍻"
	SYMBOL_CHOCO     = "🍫"
	CHOCO_BEER_RATIO = 1 // how many choco stick we need to make a beer pint
)

// Register routes for APIs
func RegisterRoutes(bot *telebot.Bot, logger *logrus.Entry, db database.AppDatabase, cfg config.APPConfig, shutdown chan os.Signal) {
	// Main commands
	bot.Handle("/birra", birraHandler(bot, logger, db, cfg, shutdown))
	bot.Handle("/birre", birreHandler(bot, logger, db, cfg, shutdown))

	// Utilities
	bot.Handle("/id", func(m *telebot.Message) {
		err := updateChatUserInfo(db, m)
		if err != nil {
			logger.WithError(err).Error("error updating chat user info in /id")
			shutdown <- syscall.SIGTERM
			return
		}

		if m.Private() {
			_, err := bot.Send(m.Chat, fmt.Sprint(m.Sender.ID))
			if err != nil {
				logger.WithError(err).Error("error sending message in /id")
				shutdown <- syscall.SIGTERM
				return
			}
		}
	})
	bot.Handle("/chatid", func(m *telebot.Message) {
		err := updateChatUserInfo(db, m)
		if err != nil {
			logger.WithError(err).Error("error updating chat user info in /chatid")
			shutdown <- syscall.SIGTERM
			return
		}

		if !m.Private() {
			_, err = bot.Send(m.Chat, fmt.Sprint(m.Chat.ID))
			if err != nil {
				logger.WithError(err).Error("error sending message in /chatid")
				shutdown <- syscall.SIGTERM
				return
			}
		}
	})
	bot.Handle(telebot.OnAddedToGroup, func(m *telebot.Message) {
		err := db.AddOrUpdateChatInfo(m.Chat.ID, m.Chat.Title)
		if err != nil {
			logger.WithError(err).Error("error updating chat user info in OnAddedToGroup")
			shutdown <- syscall.SIGTERM
			return
		}
	})
	bot.Handle(telebot.OnUserJoined, func(m *telebot.Message) {
		err := updateChatUserInfo(db, m)
		if err != nil {
			logger.WithError(err).Error("error updating chat user info in OnUserJoined")
			shutdown <- syscall.SIGTERM
			return
		}
	})
	bot.Handle(telebot.OnNewGroupTitle, func(m *telebot.Message) {
		err := db.AddOrUpdateChatInfo(m.Chat.ID, m.Chat.Title)
		if err != nil {
			logger.WithError(err).Error("error updating chat user info in OnNewGroupTitle")
			shutdown <- syscall.SIGTERM
			return
		}
	})
}
