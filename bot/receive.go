package bot

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
	"time"
)

func (app *appEnv) Receive() error {
	ch := make(chan tgbotapi.Update)

	config := tgbotapi.NewUpdate(0)
	config.Timeout = app.receiveTimeout

	go func() {
		for {
			updates, err := app.bot.GetUpdates(config)
			if err != nil {
				app.logger.Error("failed to get updates, retrying in 3 seconds...", zap.Error(err))
				time.Sleep(time.Second * 3)

				continue
			}

			for _, update := range updates {
				if update.UpdateID >= config.Offset {
					config.Offset = update.UpdateID + 1
					ch <- update
				}
			}
		}
	}()

	for update := range ch {
		fmt.Println("Received update")
		fmt.Println(update)

		// magic goes here
	}

	return nil
}
