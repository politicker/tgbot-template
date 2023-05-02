package main

import (
	"github.com/politicker/telegram-bot-template/bot"
	"os"
)

func main() {
	os.Exit(bot.CLI(os.Args))
}
