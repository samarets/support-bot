package cmd

import (
	"fmt"

	"github.com/samarets/support-bot/internal/log"
	"github.com/spf13/viper"
)

type Config struct {
	TelegramLoggerBotToken string `mapstructure:"bot_token"`
	TelegramAdminUserID    int64  `mapstructure:"admin_user_id"`
	DefaultLocale          string `mapstructure:"default_locale"`
	BotPrefix              string `mapstructure:"bot_prefix"`
}

func LoadConfig(path, configFile, configType string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName(configFile)
	viper.SetConfigType(configType)

	viper.SetDefault("bot_token", "")
	viper.SetDefault("admin_user_id", 0)
	viper.SetDefault("default_locale", "uk-UA")
	viper.SetDefault("bot_prefix", "🤖")

	viper.AutomaticEnv()

	if err = viper.ReadInConfig(); err != nil {
		fmt.Println(err)
		log.Warn().Msg("used default configs")
	} else {
		log.Info().Msgf("loaded config from %s", configType)
	}

	err = viper.Unmarshal(&config)

	return config, err
}
