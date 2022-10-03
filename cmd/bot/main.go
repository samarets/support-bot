package main

import (
	"github.com/samarets/support-bot/cmd"
	"github.com/samarets/support-bot/internal/bot"
	"github.com/samarets/support-bot/internal/db"
	"github.com/samarets/support-bot/internal/log"
	"github.com/samarets/support-bot/internal/translations"
)

func main() {
	cfg, err := cmd.LoadConfig(".", ".env", "env")
	if err != nil {
		log.Error().Err(err).Send()
		panic(err)
	}

	translator, err := translations.NewTranslator("./locales", cfg.DefaultLocale, cfg.BotPrefix)
	if err != nil {
		log.Error().Err(err).Send()
		panic(err)
	}

	database, err := db.InitDB()
	if err != nil {
		log.Error().Err(err).Send()
		panic(err)
	}

	defer func(database *db.DB) {
		err := database.Close()
		if err != nil {
			log.Error().Err(err).Send()
			panic(err)
		}
	}(database)

	err = bot.InitBot(cfg.TelegramLoggerBotToken, cfg.TelegramAdminUserID, translator, database)
	if err != nil {
		log.Error().Err(err).Send()
		panic(err)
	}
}
