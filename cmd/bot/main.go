package main

import (
	"github.com/samarets/support-bot/cmd"
	"github.com/samarets/support-bot/internal/bot"
	"github.com/samarets/support-bot/internal/db"
	"github.com/samarets/support-bot/internal/log"
)

func main() {
	cfg, err := cmd.LoadConfig(".", ".env", "env")
	if err != nil {
		log.Error().Err(err).Send()
		panic(err)
	}

	database, err := db.InitDB()
	if err != nil {
		panic(err)
	}

	defer func(database *db.DB) {
		err := database.Close()
		if err != nil {
			panic(err)
		}
	}(database)

	err = bot.InitBot(cfg.TelegramLoggerBotToken, database)
	if err != nil {
		panic(err)
	}
}
