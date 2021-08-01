package config

import "github.com/spf13/viper"

type Config struct {
	DBHost     string `mapstructure:"DB_HOST"`
	DBPassword string `mapstructure:"DB_PASSWORD"`
	DBName     string `mapstructure:"DB_NAME"`
	DBPort     string `mapstructure:"DB_PORT"`
}

func LoadConfig() (cfg Config, err error) {
	viper.SetConfigType("yaml")
	viper.SetConfigName("config")
	viper.AddConfigPath("../config")
	err = viper.ReadInConfig()

	if err != nil {
		return cfg, err
	}

	err = viper.Unmarshal(&cfg)
	return
}
