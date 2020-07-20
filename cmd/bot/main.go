package main

import (
	"context"
	"expvar"
	"fmt"
	"github.com/ardanlabs/conf"
	"github.com/boltdb/bolt"
	"github.com/joho/godotenv"
	"github.com/pkg/errors"
	"github.com/sistemisti-italiani/birrometro_bot/service/apilogger"
	"github.com/sistemisti-italiani/birrometro_bot/service/bot"
	"github.com/sistemisti-italiani/birrometro_bot/service/config"
	"github.com/sistemisti-italiani/birrometro_bot/service/database"
	"gopkg.in/tucnak/telebot.v2"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
	_ "net/http/pprof" // #nosec G108
	"os"
	"os/signal"
	"syscall"
	"time"
)

// These two "variables" are modified at build-time
// APP_VERSION contains the app version (tag + commit after tag + current git ref)
var APP_VERSION = "devel"

// BUILD_DATE contains the timestamp of the build
var BUILD_DATE = "n/a"

// Main function
func main() {
	_ = godotenv.Load()
	if err := run(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "error: ", err)
		os.Exit(1)
	}
}

// Core function
// Loads the configuration, initialize log, default DB and auth handlers, debug web bot and main web bot
func run() error {
	// =========================================================================
	// Load Configuration and defaults

	// Create configuration defaults
	var cfg struct {
		Config struct {
			// YAML config file
			Path string `conf:"default:/conf/config.yml"`
		}
		Web struct {
			// Bot socket for webhook - if Bot->Webhook is not specified, poller will be used instead
			Bot string `conf:"default:localhost:3000"`

			// Debug socket
			Debug string `conf:"default:localhost:4000"`
		}
		Bot struct {
			// Public URL for webhook
			Webhook string `conf:"default:-"`

			// Poller timeout (if poller is used)
			PollerTimeout time.Duration `conf:"default:10s"`

			// Bot Token
			Token string `conf:"default:-"`

			// Shutdown context timeout
			ShutdownTimeout time.Duration `conf:"default:10s"`
		}
		DB struct {
			// Bolt DB file path
			Path string `conf:"default:-"`
		}
		Log apilogger.LogSettings
		App config.APPConfig
	}

	// Try to load configuration from environment variables and command line switches
	if err := conf.Parse(os.Args[1:], "CFG", &cfg); err != nil {
		if err == conf.ErrHelpWanted {
			usage, err := conf.Usage("CFG", &cfg)
			if err != nil {
				return errors.Wrap(err, "generating config usage")
			}
			fmt.Println(usage)
			return nil
		}
		return errors.Wrap(err, "parsing config")
	}

	// Override values from YAML if specified and if it exists (useful in k8s/compose)
	fp, err := os.Open(cfg.Config.Path)
	if err != nil && !os.IsNotExist(err) {
		fmt.Println("Can't read the config file, while it exists...")
	} else if err == nil {
		yamlFile, err := ioutil.ReadAll(fp)
		if err != nil {
			fmt.Printf("can't read config file: %v", err)
		}
		err = yaml.Unmarshal(yamlFile, &cfg)
		if err != nil {
			fmt.Printf("can't unmarshal config file: %v", err)
		}
		_ = fp.Close()
	}

	// =========================================================================
	// App Starting!

	// Init logging
	logger, err := apilogger.NewApiLogger(cfg.Log)
	if err != nil {
		panic(err)
	}

	// Print the build version for our logs. Also expose it under /debug/vars.
	logger.Infof("main : Started : Application initializing : version %q (%s)", APP_VERSION, BUILD_DATE)
	defer logger.Debugf("main : Completed")
	expvar.NewString("build").Set(APP_VERSION)
	expvar.NewString("build-date").Set(BUILD_DATE)

	// Print out the configuration
	out, err := conf.String(&cfg)
	if err != nil {
		return errors.Wrap(err, "generating config for output")
	}
	logger.Tracef("main : Config :\n%v\n", out)

	// =========================================================================
	// Start Database
	if cfg.DB.Path == "-" {
		logger.Error("DB path not specified, can't start")
		return errors.New("DB path not specified")
	}
	boltdb, err := bolt.Open(cfg.DB.Path, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		logger.WithError(err).Error("error opening Bolt DB")
		return err
	}

	db := database.NewAppDatabase(boltdb)
	err = db.Init()
	if err != nil {
		logger.WithError(err).Error("error init Bolt DB")
		return err
	}
	defer func() {
		_ = boltdb.Close()
	}()

	// =========================================================================
	// Make a channel to listen for an interrupt or terminate signal from the OS.
	// Use a buffered channel because the signal package requires it.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// =========================================================================
	// Start project specific items

	err = bot.Startup(logger, db, cfg.App)
	if err != nil {
		logger.WithError(err).Error("error launching bot startup")
		return errors.Wrap(err, "error launching bot startup")
	}

	// =========================================================================
	// Start Debug Service
	//
	// /debug/pprof - Added to the default mux by importing the net/http/pprof package.
	// /debug/vars - Added to the default mux by importing the expvar package.
	//
	// Not concerned with shutting this down when the application is shutdown.

	logger.Info("main : Started : Initializing debugging support")

	go func() {
		logger.Infof("main : Debug Listening %s", cfg.Web.Debug)
		logger.Infof("main : Debug Listener closed : %v", http.ListenAndServe(cfg.Web.Debug, http.DefaultServeMux))
	}()

	// Start Metrics service

	// =========================================================================
	// Start Bot Service

	logger.Info("main : Started : Initializing Bot")

	if cfg.Bot.Token == "-" {
		logger.WithError(err).Error("missing bot token")
		return errors.Wrap(err, "missing bot token")
	}

	var botSettings = telebot.Settings{
		Token: cfg.Bot.Token,
	}
	// Start webhook
	if cfg.Bot.Webhook != "-" {
		// Create routes
		router := telebot.Webhook{
			Listen: cfg.Web.Bot,
			Endpoint: &telebot.WebhookEndpoint{
				PublicURL: cfg.Bot.Webhook,
			},
		}

		botSettings.Poller = &router
	} else {
		botSettings.Poller = &telebot.LongPoller{Timeout: cfg.Bot.PollerTimeout}
	}

	b, err := telebot.NewBot(botSettings)
	if err != nil {
		logger.WithError(err).Error("can't start bot")
		return errors.Wrap(err, "can't start bot")
	}

	go func() {
		<-shutdown
		b.Stop()
	}()

	bot.RegisterRoutes(b, logger, db, cfg.App, shutdown)

	b.Start()

	// =========================================================================
	// Shutdown

	logger.Infof("main : Stopping : Start shutdown")

	// Give outstanding requests a deadline for completion.
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Bot.ShutdownTimeout)
	defer cancel()

	// Launching the shutdown generic handler
	err = bot.Shutdown(ctx)
	if err != nil {
		logger.Warningf("main : Graceful shutdown did not complete in %v : %v", cfg.Bot.ShutdownTimeout, err)
	}

	return err
}
