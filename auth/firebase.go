package auth

import (
	"context"
	"log"
	"os"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"google.golang.org/api/option"
)

var logger = log.New(os.Stderr, "auth: ", log.LstdFlags|log.Lshortfile)

func SetupFirebaseAuthClient() *auth.Client {
	serviceAccountKeyPath := os.Getenv("SERVICE_ACCOUNT_KEY_PATH")
	if serviceAccountKeyPath == "" {
		serviceAccountKeyPath = "/Users/thomas/Documents/projects/away_backend/config/serviceAccountKey.json" // "C:/Users/Asus/Documents/prog/away_backend/config/serviceAccountKey.json"
	}

	// Initialise auth client
	opt := option.WithCredentialsFile(serviceAccountKeyPath)
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		logger.Print("Error initialising app", err)
	}
	auth, err := app.Auth(context.Background())
	if err != nil {
		logger.Print("Firebase load error", err)
	}

	return auth
}
