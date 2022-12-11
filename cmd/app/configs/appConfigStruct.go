package configs

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
)

type AppConfig struct {
	DB struct {
		URI      string
		PoolSize uint64
		DBs      struct {
			VideoCatalogueDB string
		}
		Collections struct {
			VideoCatalogueColl string
			VideoFilesColl     string
		}
	}
	Logger struct {
		OutFile string
		Level   string
	}
}

var Config *AppConfig

func (ac *AppConfig) GetLogLevel() log.Level {
	switch ac.Logger.Level {
	case "info":
		return log.InfoLevel
	case "debug":
		return log.DebugLevel
	case "error":
		return log.ErrorLevel
	default:
		return log.InfoLevel
	}

}

func (ac *AppConfig) GetLogFileIO() (file *os.File) {
	if ac.Logger.OutFile != "" {
		file, _ = os.OpenFile(ac.Logger.OutFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0777)
	}
	return
}

func LoadConfig() {
	// Loading application configs, it is env specific. Right now, there is only one type of config

	if Config == nil {
		viper.AddConfigPath("./cmd/app/configs")
		viper.SetConfigName("appConfig") // Register configs file name (no extension)
		viper.SetConfigType("json")      // Look for specific type
		if err := viper.ReadInConfig(); err != nil {
			log.Fatal("Config loading failed!!")
		}
		Config = &AppConfig{}
		Config.DB.URI = os.Getenv("MONGODB_URI")
		Config.DB.PoolSize = uint64(viper.Get("db.mongoDB.poolSize").(float64))
		Config.DB.Collections.VideoFilesColl = viper.Get("db.mongoDB.collections.videFilesCollection").(string)
		Config.DB.Collections.VideoCatalogueColl = viper.Get("db.mongoDB.collections.videoCatalogueCollection").(string)
		Config.DB.DBs.VideoCatalogueDB = viper.Get("db.mongoDB.dbs.videoCatalogueDB").(string)
		Config.Logger.OutFile = viper.Get("logger.outfile").(string)
		Config.Logger.Level = viper.Get("logger.level").(string)
	}

}
