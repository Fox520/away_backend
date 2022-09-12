package testhelper

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"

	fb_auth "firebase.google.com/go/auth"
	away_auth "github.com/Fox520/away_backend/auth"
)

var webApiKey string = ""
var MainUser *fb_auth.UserRecord
var OtherUser *fb_auth.UserRecord

const mainEmail string = "main@email.com"
const otherEmail string = "other@email.com"

// Returns token of a test user
func GetMainUserAuthToken() string {
	values := map[string]string{"email": MainUser.Email, "password": "123456", "returnSecureToken": "true"}
	json_data, err := json.Marshal(values)

	if err != nil {
		log.Fatal(err)
	}

	resp, err := http.Post("https://www.googleapis.com/identitytoolkit/v3/relyingparty/verifyPassword?key="+webApiKey, "application/json",
		bytes.NewBuffer(json_data))

	if err != nil {
		log.Fatal(err)
	}

	var res map[string]string

	json.NewDecoder(resp.Body).Decode(&res)
	return res["idToken"]
}

// Returns token of secondary test user
func GetOtherUserAuthToken() string {
	values := map[string]string{"email": OtherUser.Email, "password": "123456", "returnSecureToken": "true"}
	json_data, err := json.Marshal(values)

	if err != nil {
		log.Fatal(err)
	}

	resp, err := http.Post("https://www.googleapis.com/identitytoolkit/v3/relyingparty/verifyPassword?key="+webApiKey, "application/json",
		bytes.NewBuffer(json_data))

	if err != nil {
		log.Fatal(err)
	}

	var res map[string]string

	json.NewDecoder(resp.Body).Decode(&res)
	return res["idToken"]
}

func CreateUsers() {

	params := (&fb_auth.UserToCreate{}).Email(mainEmail).DisplayName("Alice").Password("123456").PhotoURL("https://www.example.com/photo.png").Disabled(false)
	u, err := away_auth.GetFirebaseAuthClient().CreateUser(context.Background(), params)
	if err != nil {
		log.Fatalf("error creating user 1: %v\n", err)
	}
	MainUser = u

	params2 := (&fb_auth.UserToCreate{}).Email(otherEmail).DisplayName("Jane").Password("123456").PhotoURL("https://www.example.com/photo.png").Disabled(false)
	u, err = away_auth.GetFirebaseAuthClient().CreateUser(context.Background(), params2)
	if err != nil {
		log.Fatalf("error creating user 2: %v\n", err)
	}
	OtherUser = u

}

func DeleteTestUsers() {
	if MainUser != nil {
		away_auth.GetFirebaseAuthClient().DeleteUser(context.Background(), MainUser.UID)

	} else {
		// Probably first run
		uid := getUserId(mainEmail, "123456")
		away_auth.GetFirebaseAuthClient().DeleteUser(context.Background(), uid)
	}
	if OtherUser != nil {
		away_auth.GetFirebaseAuthClient().DeleteUser(context.Background(), OtherUser.UID)

	} else {
		uid := getUserId(otherEmail, "123456")
		away_auth.GetFirebaseAuthClient().DeleteUser(context.Background(), uid)
	}
}

func getUserId(email string, password string) string {
	values := map[string]string{"email": email, "password": password, "returnSecureToken": "true"}
	json_data, err := json.Marshal(values)

	if err != nil {
		log.Fatal(err)
	}

	resp, err := http.Post("https://www.googleapis.com/identitytoolkit/v3/relyingparty/verifyPassword?key="+webApiKey, "application/json",
		bytes.NewBuffer(json_data))

	if err != nil {
		log.Fatal(err)
	}

	var res map[string]string

	json.NewDecoder(resp.Body).Decode(&res)
	return res["localId"]
}
