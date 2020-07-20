// +build !zap

package apilogger

import (
	"github.com/sirupsen/logrus"
	"os"
)

type LogSettings struct {
	Level            string `conf:"default:warn"`
	MethodName       bool   `conf:"default:false"`
	JSON             bool   `conf:"default:false"`
	Destination      string `conf:"default:stderr"` // Possible values: stderr, stdout, file, TODO
	File             string `conf:"default:/tmp/debug.log"`
	CombinedToStdout bool   `conf:"default:true"`
}

func NewApiLogger(cfg LogSettings) (*logrus.Entry, error) {
	// Init Logging
	logger := logrus.New()

	// Setting output
	if cfg.Destination == "stdout" {
		logger.SetOutput(os.Stdout)
	} else if cfg.Destination == "stderr" {
		logger.SetOutput(os.Stderr)
	} else if cfg.Destination == "file" {
		file, err := os.OpenFile(cfg.File, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
		if err == nil {
			logger.SetOutput(file)
		} else {
			logger.SetOutput(os.Stderr)
			logger.WithError(err).Error("Can't open log file for writing, using stderr")
		}
	}
	// TODO: implement others like ELK

	// Set logging level
	switch cfg.Level {
	case "trace":
		logger.SetLevel(logrus.TraceLevel)
	case "debug":
		logger.SetLevel(logrus.DebugLevel)
	case "warn":
		logger.SetLevel(logrus.WarnLevel)
	case "error":
		logger.SetLevel(logrus.ErrorLevel)
	case "info":
		fallthrough
	default:
		logger.SetLevel(logrus.InfoLevel)
	}

	if cfg.MethodName {
		logger.SetReportCaller(true)
	}

	if cfg.JSON {
		logger.SetFormatter(&logrus.JSONFormatter{})
	}

	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	return logger.WithFields(logrus.Fields{
		"hostname": hostname,
	}), nil
}

func CloseLogger(_ *logrus.Logger) {
	// Nothing to do here
}
