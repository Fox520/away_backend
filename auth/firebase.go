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
	// Initialise auth client

	opt := option.WithCredentialsFile("../config/serviceAccountKey.json")
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
