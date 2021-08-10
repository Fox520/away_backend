package testhelper

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
)

var webApiKey string = ""

// Returns token of a test user
func GetMainUserAuthToken() string {
	values := map[string]string{"email": "testing@email.com", "password": "123456", "returnSecureToken": "true"}
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
	values := map[string]string{"email": "other@email.com", "password": "123456", "returnSecureToken": "true"}
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
