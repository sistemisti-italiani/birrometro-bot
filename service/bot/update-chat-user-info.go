package bot

import (
	"github.com/sistemisti-italiani/birrometro_bot/service/database"
	"gopkg.in/tucnak/telebot.v2"
)

// This function is used to sync the DB: as Telegram bots don't have the capability to ask for:
// - user infos (username, first name, last name) by user ID
// - group infos (title) by group ID
// - even in which group the bot is!
// we need to have a database with all infos
func updateChatUserInfo(db database.AppDatabase, m *telebot.Message) error {
	err := db.AddOrUpdateUserInfo(int64(m.Sender.ID), m.Sender.Username, m.Sender.FirstName, m.Sender.LastName)
	if err != nil {
		return err
	}
	if !m.Private() {
		err = db.AddOrUpdateChatInfo(m.Chat.ID, m.Chat.Title)
	}
	return err
}
