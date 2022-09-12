package config

import (
	"log"
	"os"
	"sync"

	"github.com/spf13/viper"
)

var config *viper.Viper
var once sync.Once

func setup(env string) {

	var err error
	config = viper.New()
	config.SetConfigType("yaml")
	config.SetConfigName(env)
	config.AddConfigPath("../config/")
	config.AddConfigPath("config/")
	err = config.ReadInConfig()
	if err != nil {
		log.Fatal("error on parsing configuration file")
	}
}

func GetConfig() *viper.Viper {
	once.Do(func() {
		environment := os.Getenv("DEPLOYMENT_ENV")
		if environment == "" {
			environment = "development"
		}
		setup(environment)
	})
	return config
}
