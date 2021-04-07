package web

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwt"
)

// Create a struct to read the username and password from the request body
type Credentials struct {
	Username string `json:"username"`
}

// ServeAuthWithKey creates a closure with the given key and
// returns a route handler that
// creates a JWT, signs it with the provided key
// and adds it to the Response Header
func ServeAuthWithKey(privKey *rsa.PrivateKey) (HandleFunc func(w http.ResponseWriter, r *http.Request)) {
	return func(w http.ResponseWriter, r *http.Request) {
		// parse client's desired username from json body
		var creds Credentials
		err := json.NewDecoder(r.Body).Decode(&creds)
		if err != nil {
			// If the structure of the body is wrong, return an HTTP error
			w.WriteHeader(http.StatusBadRequest)
			fmt.Printf("failed to decode json: %s", err)
			return
		}

		// #TODO: check if username is already taken?
		fmt.Printf("%s has requested a JWT",creds.Username)

		// create new client
		uuidStr,_ := uuid.New().MarshalText()
	
		// create JWT
		t := jwt.New()
		t.Set(jwt.IssuedAtKey, time.Now())
		t.Set(`username`, creds.Username)
		t.Set(`uuid`, uuidStr)

		// Signing the token (using raw rsa.PrivateKey)
		// also marshalls to json
		signed, err := jwt.Sign(t, jwa.RS512, privKey)
		if err != nil {
			// also wraps a json marshalling error
			fmt.Printf("failed to sign token: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Finally, we set the client cookie for "token" as the JWT we just generated
		// we also set an expiry time which is the same as the token itself
		http.SetCookie(w, &http.Cookie{
			Name:    "token",
			Value:   string(signed[:]),
		})
		

		// enable HttpOnly
		// #TODO: frontend team will not be able to access it via clientside scripts! (security measure)
		// wait with enabling until we can make sure it works without clientside interaction

	}
}

// Middleware
		// parse and verify
		// 

type authenticationMiddleware struct {
	privateKey *rsa.PrivateKey
	// #TODO: later hold client manager to check if already logged in via uuid
}
// Middleware function, which will be called for each request
// checks if user is authenticated via valid token
// not authorized users requests get redirected to /auth
func (amw *authenticationMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		token, err := jwt.ParseHeader(r.Header,"Authorization",jwt.WithVerify(jwa.RS512, amw.privateKey.PublicKey))
		if err != nil {
			http.Error(w, "Forbidden", http.StatusForbidden)
		} else {
			if user, ok := token.Get(`username`); ok {
				fmt.Printf("Authenticated user %s\n", user)
			}
			next.ServeHTTP(w, r)
		}
	})
}