package bot

import (
	"flag"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
	"os"
)

type appEnv struct {
	bot            *tgbotapi.BotAPI
	tgAPIKey       string
	logger         *zap.Logger
	receiveTimeout int
	debug          bool
}

func (app *appEnv) fromArgs(args []string) error {
	fl := flag.NewFlagSet("tbot", flag.ContinueOnError)

	fl.StringVar(&app.tgAPIKey, "k", os.Getenv("TELEGRAM_API_KEY"), "Telegram API key")
	fl.StringVar(&app.tgAPIKey, "key", os.Getenv("TELEGRAM_API_KEY"), "Telegram API key")

	fl.IntVar(
		&app.receiveTimeout, "t", 60, "Receive message handler timeout",
	)
	fl.IntVar(
		&app.receiveTimeout, "timeout", 60, "Receive message handler timeout",
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

	if app.tgAPIKey == "" {
		fl.Usage()
		return fmt.Errorf("missing Telegram API key")
	}

	bot, err := tgbotapi.NewBotAPI(app.tgAPIKey)
	if err != nil {
		return fmt.Errorf("failed to create bot: %w", err)
	}
	app.bot = bot

	var config zap.Config
	if os.Getenv("GO_ENV") == "production" && !app.debug {
		config = zap.NewProductionConfig()
	} else {
		config = zap.NewDevelopmentConfig()
	}

	logger, err := config.Build()
	if err != nil {
		return err
	}

	switch fl.Arg(0) {
	case "send":
		logger = logger.With(zap.String("context", "send-cmd"))
	case "receive":
		logger = logger.With(zap.String("context", "receive-cmd"))
	}
	app.logger = logger

	return nil
}

func (app *appEnv) run() error {
	switch flag.Arg(0) {
	case "send":
		return app.Send()
	case "receive":
		return app.Receive()
	default:
		return fmt.Errorf("unknown command: %s", flag.Arg(0))
	}
}

func CLI(args []string) int {
	var app appEnv
	err := app.fromArgs(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "init error: %v\n", err)
		return 2
	}
	if err = app.run(); err != nil {
		fmt.Fprintf(os.Stderr, "runtime error: %v\n", err)
		return 1
	}
	return 0
}
