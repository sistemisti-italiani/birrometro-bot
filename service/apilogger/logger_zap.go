// +build zap

package apilogger

import (
	"go.uber.org/zap"
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

func NewApiLogger(cfg LogSettings) (*zap.Logger, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	zapcfg := zap.Config{
		Encoding: "console",
		InitialFields: map[string]interface{}{
			"hostname": hostname,
		},
	}

	// Init Logging

	// Setting output
	if cfg.Destination == "stdout" {
		zapcfg.OutputPaths = []string{"stdout"}
		zapcfg.ErrorOutputPaths = []string{"stdout"}
	} else if cfg.Destination == "stderr" {
		zapcfg.OutputPaths = []string{"stderr"}
		zapcfg.ErrorOutputPaths = []string{"stderr"}
	} else if cfg.Destination == "file" {
		zapcfg.OutputPaths = []string{"cfg.File"}
		zapcfg.ErrorOutputPaths = []string{"cfg.File"}
	}
	// TODO: implement others like ELK

	// Set logging level
	switch cfg.Level {
	case "trace":
		fallthrough
	case "debug":
		zapcfg.Level.SetLevel(zap.DebugLevel)
	default:
		fallthrough
	case "info":
		zapcfg.Level.SetLevel(zap.InfoLevel)
	case "warn":
		zapcfg.Level.SetLevel(zap.WarnLevel)
	case "error":
		zapcfg.Level.SetLevel(zap.ErrorLevel)
	}

	zapcfg.DisableCaller = !cfg.MethodName

	if cfg.JSON {
		zapcfg.Encoding = "json"
	}

	return zapcfg.Build()
}

func CloseLogger(logger *zap.Logger) {
	logger.Sync()
}
