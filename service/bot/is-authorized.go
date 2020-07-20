package bot

import (
	"github.com/sistemisti-italiani/birrometro_bot/service/config"
	"gopkg.in/tucnak/telebot.v2"
)

func isAuthorized(m *telebot.Message, cfg config.APPConfig) bool {
	return m.Private() || m.Chat != nil && (m.Chat.ID == cfg.Group || cfg.Group == 0)
}
