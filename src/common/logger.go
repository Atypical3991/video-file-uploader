package logger

import (
	"city_os/cmd/app/configs"
	log "github.com/sirupsen/logrus"
)

var Logger *log.Logger

// Initialising Logger  with set of flags/params from app config

func InitLogger() {
	if Logger == nil {
		Logger = &log.Logger{
			Formatter: &log.JSONFormatter{},
			Level:     configs.Config.GetLogLevel(),
		}
		if configs.Config.GetLogFileIO() != nil {
			Logger.Out = configs.Config.GetLogFileIO()
		}
	}
}
