package config

import (
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	DBHost            string `mapstructure:"DB_HOST"`
	DBPassword        string `mapstructure:"DB_PASSWORD"`
	DBName            string `mapstructure:"DB_NAME"`
	DBPort            string `mapstructure:"DB_PORT"`
	ELASTICSEARCH_URL string `mapstructure:"ELASTICSEARCH_URL"`
}

func LoadConfig() (cfg Config, err error) {
	folder := os.Getenv("CONFIG_FOLDER_PATH")
	if folder == "" {
		folder = "C:/Users/Asus/Documents/prog/away_backend/config"
	}
	viper.SetConfigType("yaml")
	viper.SetConfigName("config")
	viper.AddConfigPath(folder)
	err = viper.ReadInConfig()

	if err != nil {
		return cfg, err
	}

	err = viper.Unmarshal(&cfg)
	return
}
