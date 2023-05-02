package bot

import (
	"flag"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
	"os"
	"time"
)

type appEnv struct {
	bot            *tgbotapi.BotAPI
	tgAPIKey       string
	logger         *zap.Logger
	receiveTimeout time.Duration
	debug          bool
}

func (app *appEnv) fromArgs(args []string) error {
	fl := flag.NewFlagSet("tbot", flag.ContinueOnError)

	fl.StringVar(&app.tgAPIKey, "k", os.Getenv("TELEGRAM_API_KEY"), "Telegram API key")
	fl.StringVar(&app.tgAPIKey, "key", os.Getenv("TELEGRAM_API_KEY"), "Telegram API key")

	fl.DurationVar(
		&app.receiveTimeout, "t", 60*time.Second, "Receive message handler timeout",
	)
	fl.DurationVar(
		&app.receiveTimeout, "timeout", 60*time.Second, "Receive message handler timeout",
	)

	fl.BoolVar(
		&app.debug, "d", false, "Enable debug mode",
	)
	fl.BoolVar(
		&app.debug, "debug", false, "Enable debug mode",
	)

	if err := fl.Parse(args); err != nil {
		return err
	}

	fl.Usage = func() {
		usage := `Usage: tgbot  [options]

Commands:
  send      Starts a pub/sub listener that consumes messages and sends them to a Telegram chat
  receive   Starts a tgbotapi listener that consumes *telegram* message events

Options:
  -k, --key string    API key (default: TELEGRAM_API_KEY environment variable)
  -t, --timeout int   Receive message handler timeout (default: 60)
  -d, --debug         Enable debug mode
  -h, --help          Show this help message
`
		fmt.Fprintf(os.Stderr, usage)
	}

	bot, err := tgbotapi.NewBotAPI(app.tgAPIKey)
	if err != nil {
		return err
	}
	app.bot = bot

	var config zap.Config
	if os.Getenv("GO_ENV") == "production" {
		config = zap.NewProductionConfig()
	} else {
		config = zap.NewDevelopmentConfig()
	}
	if app.debug {
		config = zap.NewDevelopmentConfig()
	}

	logger, err := config.Build()
	if err != nil {
		return err
	}
	app.logger = logger

	return nil
}

func (app *appEnv) run() error {
	return nil
}

func CLI(args []string) int {
	var app appEnv
	err := app.fromArgs(args)
	if err != nil {
		return 2
	}
	if err = app.run(); err != nil {
		fmt.Fprintf(os.Stderr, "Runtime error: %v\n", err)
		return 1
	}
	return 0
}
