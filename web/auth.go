package web

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"net/http"
)

// Create a struct to read the username and password from the request body
type Credentials struct {
	Username string `json:"username"`
}

// ServeAuthWithKey creates a closure with the given key and
// returns a route handler that
// creates a JWT, signs it and adds it to the Response Header
// via Set-Cookie. Also enables HttpOnly for security
func ServeAuthWithKey(privKey *rsa.PrivateKey) (HandleFunc func(w http.ResponseWriter, r *http.Request)) {
	return func(w http.ResponseWriter, r *http.Request) {
		// parse client's desired username from json body
		var creds Credentials
		err := json.NewDecoder(r.Body).Decode(&creds)
		if err != nil {
			// If the structure of the body is wrong, return an HTTP error
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// #TODO: check if username is already taken?
		fmt.Printf("%s has requested a JWT",creds.Username)
		
		// create JWT

		// sign JWT

		// create cookie

		// set cookie

		// enable HttpOnly

	}
}

// Middleware
		// parse and verify
		// token, err := jwt.ParseHeader(r.Header,"Authorization",jwt.WithVerify(jwa.RS512, &privKey.PublicKey))