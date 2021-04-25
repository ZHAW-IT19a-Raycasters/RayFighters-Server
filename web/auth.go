package web

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwt"
)

const (
	JWTCOOKIENAME = "token"
)

// Create a struct to read the username and password from the request body
type Credentials struct {
	Username string `json:"username"`
}

// ServeAuthWithKey creates a closure with the given key and
// returns a route handler that
// creates a JWT, signs it with the provided key
// and adds it to the Response Header via Set-Cookie
func ServeAuthWithKey(privKey *rsa.PrivateKey) (HandleFunc func(w http.ResponseWriter, r *http.Request)) {
	return func(w http.ResponseWriter, r *http.Request) {
		
		var creds Credentials
		err := parseToCredentialsFromJSON(&creds, r.Body)
		if err != nil {
			// If the structure of the body is wrong, return an HTTP error
			http.Error(w, err.Error(), http.StatusBadRequest)
			logger.Warnf("failed to decode json from request body: %s", err)
			return
		}

		// create new client uuid
		uuidStr := uuid.New().String()
	
		// create JWT
		t := jwt.New()
		t.Set(jwt.IssuedAtKey, time.Now())
		t.Set(`username`, creds.Username)
		t.Set(`uuid`, uuidStr)

		// Signing the token (using raw rsa.PrivateKey)
		// also marshalls to JSON
		signed, err := jwt.Sign(t, jwa.RS512, privKey)
		if err != nil {
			// also wraps a json marshalling error
			logger.Warnf("failed to sign token: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Finally, we request to set the client cookie to the JWT we just generated
		// #TODO expiry time?
		http.SetCookie(w, &http.Cookie{
			Name:    JWTCOOKIENAME,
			Value:   string(signed[:]),
		})

	}
}

// holds information needed for validating provided authentication
type authenticationMiddleware struct {
	privateKey *rsa.PrivateKey
}

// Middleware checks if user is authenticated via valid token
func (amw *authenticationMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// get cookie from request
		jwtCookie, err := r.Cookie(JWTCOOKIENAME)
		if err != nil{
			http.Error(w, "no JWT found in Cookies", http.StatusUnauthorized)
			logger.Warnf("no JWT found in request Cookies %s", err)
			return
		}

		token, err := jwt.ParseString(jwtCookie.Value,jwt.WithVerify(jwa.RS512, amw.privateKey.PublicKey))
		if err != nil {
			logger.Warnf("unauthenticated request %s", err)
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		
		user, userOk := token.Get(`username`)
		uuid, uuidOk := token.Get(`uuid`)
		if (userOk && uuidOk) {
			logger.Infof("user %s with uuid %s authenticated. proceeding to %s", user, uuid, r.URL)
		}
		next.ServeHTTP(w, r)
	})
}

// runs validity checks against a received username
func validUsername(str string) (string, error){
	// in utf-8 runes
	const maxUsernameLength = 255
	if !utf8.ValidString(str){
		return "", fmt.Errorf("invalid utf-8 runes in username %s", str)
	}
	
	if  utf8.RuneCountInString(str)>maxUsernameLength {
		return "", fmt.Errorf("username contains too many runes %d, %s",utf8.RuneCountInString(str), str)
	}
	escaped := html.EscapeString(str)
	return escaped, nil
}

// parse client's desired username from json body
func parseToCredentialsFromJSON(creds *Credentials, r io.Reader) error{
	err := json.NewDecoder(r).Decode(&creds)
	return err
}