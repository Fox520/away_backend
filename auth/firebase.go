package auth

import (
	"context"
	"log"
	"os"
	"sync"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"github.com/spf13/viper"
	"google.golang.org/api/option"
)

var logger = log.New(os.Stderr, "auth: ", log.LstdFlags|log.Lshortfile)

var client *auth.Client
var once sync.Once

func GetFirebaseAuthClient() *auth.Client {
	once.Do(func() {
		environment := os.Getenv("DEPLOYMENT_ENV")
		if environment == "" {
			environment = "development"
		}
		setup(environment)

		serviceAccountKeyPath := config.GetString("service_account_path")

		// Initialise auth client
		opt := option.WithCredentialsFile(serviceAccountKeyPath)
		app, err := firebase.NewApp(context.Background(), nil, opt)
		if err != nil {
			logger.Fatal("Error initialising app", err)
		}
		auth, err := app.Auth(context.Background())
		if err != nil {
			logger.Fatal("Firebase load error", err)
		}
		client = auth
	})

	return client
}

var config *viper.Viper

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
