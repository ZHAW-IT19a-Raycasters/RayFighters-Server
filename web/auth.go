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

		// Finally, we request to set the client cookie to the JWT we just generated
		// #TODO expiry time?
		http.SetCookie(w, &http.Cookie{
			Name:    JWTCOOKIENAME,
			Value:   string(signed[:]),
		})

	}
}

type authenticationMiddleware struct {
	privateKey *rsa.PrivateKey
}
// Middleware checks if user is authenticated via valid token
// not authorized users requests get redirected to /auth
func (amw *authenticationMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// get cookie from request
		jwtCookie, err := r.Cookie(JWTCOOKIENAME)
		if err != nil{
			fmt.Printf("no JWT found in Cookies")
			http.Error(w, "no JWT found in Cookies", http.StatusUnauthorized)
		}

		fmt.Printf("%s found in cookies\n", jwtCookie.Name)

		token, err := jwt.ParseString(jwtCookie.Value,jwt.WithVerify(jwa.RS512, amw.privateKey.PublicKey))
		if err != nil {
			fmt.Printf("unauthenticated request")
			http.Error(w, "Forbidden", http.StatusUnauthorized)
		} else {
			if user, ok := token.Get(`username`); ok {
				fmt.Printf("Authenticated user %s\n", user)
			}
			if uuid, ok := token.Get(`uuid`); ok {
				fmt.Printf("with uuid %s\n", uuid)
			}
			next.ServeHTTP(w, r)
		}
	})
}

// runs various validity checks against a received username
func validUsername(str string) (string, error){
	// in utf-8 runes
	const maxUsernameLength = 255
	if !utf8.ValidString(str){
		return "", fmt.Errorf("invalid utf-8 runes in username %s", str)
	}
	escaped := html.EscapeString(str)
	if  utf8.RuneCountInString(escaped)>maxUsernameLength {
		return "", fmt.Errorf("username contains too many runes %d, %s",utf8.RuneCountInString(str), escaped)
	}
	return escaped, nil
}

func parseToCredentialsFromJSON(creds *Credentials, r io.Reader) error{
	// parse client's desired username from json body
	err := json.NewDecoder(r).Decode(&creds)
	return err
}