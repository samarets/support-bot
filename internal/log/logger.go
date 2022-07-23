package log

import (
	"os"

	"github.com/rs/zerolog"
)

var DefaultLogger *zerolog.Logger

func init() {
	zerolog.LevelFieldName = "log_level"
	zerolog.TimestampFieldName = "timestamp"
	zerolog.TimeFieldFormat = "2006-01-02T15:04:05.000Z07:00"
	logger := zerolog.New(os.Stdout).With().Timestamp().Caller().Logger()
	DefaultLogger = &logger
}

func Info() *zerolog.Event {
	return DefaultLogger.Info()
}

func Error() *zerolog.Event {
	return DefaultLogger.Error()
}

func Warn() *zerolog.Event {
	return DefaultLogger.Warn()
}
