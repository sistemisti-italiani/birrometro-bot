package bot

import (
	"github.com/sirupsen/logrus"
	"github.com/sistemisti-italiani/birrometro_bot/service/config"
	"github.com/sistemisti-italiani/birrometro_bot/service/database"
)

func Startup(logger *logrus.Entry, db database.AppDatabase, cfg config.APPConfig) error {
	logger.Info("local init OK")
	return nil
}
