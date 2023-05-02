package bot

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
	"time"
)

func Receive(app *appEnv) {
	ch := make(chan tgbotapi.Update)

	config := tgbotapi.NewUpdate(0)
	config.Timeout = 60

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
	}
}
