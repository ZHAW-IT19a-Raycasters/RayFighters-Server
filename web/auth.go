package web

import (
	"crypto/rand"
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
	JwtCookieName string = "token"
	JwtUserName   string = "username"
	JwtUUID       string = "uuid"
)

// Create a struct to read the username and password from the request body
type Credentials struct {
	Username string `json:"username"`
}

// holds information needed for validating provided authentication
type authenticationMiddleware struct {
	privateKey *rsa.PrivateKey
}

func newAuthMiddleware() (*authenticationMiddleware, error) {
	// generate private key
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		logger.Error("Failed to create the auth middlwares random key")
		return nil, err
	}
	return &authenticationMiddleware{
		privateKey: privKey,
	}, nil
}

// ServeAuthWithKey creates a closure with the given key and
// returns a route handler that
// creates a JWT, signs it with the provided key
// and adds it to the Response Header via Set-Cookie
func (auth *authenticationMiddleware) ServeAuthWithKey() (HandleFunc func(w http.ResponseWriter, r *http.Request)) {
	return func(w http.ResponseWriter, r *http.Request) {

		var creds Credentials
		err := parseToCredentialsFromJSON(&creds, r.Body)
		if err != nil {
			// If the structure of the body is wrong, return an HTTP error
			http.Error(w, err.Error(), http.StatusBadRequest)
			logger.Warnf("failed to decode json from request body: %s", err)
			return
		}

		cookie, err := auth.createCookie(&creds)
		if err != nil {
			// also wraps a json marshalling error
			logger.Warnf("failed to sign token: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		// w.Write(signed)
		// Finally, we request to set the client cookie to the JWT we just generated
		http.SetCookie(w, cookie)

	}
}

func (auth *authenticationMiddleware) createCookie(creds *Credentials) (*http.Cookie, error) {
	// create new client uuid
	uuidStr := uuid.NewString()

	// create JWT
	t := jwt.New()
	t.Set(jwt.IssuedAtKey, time.Now())
	t.Set(JwtUserName, creds.Username)
	t.Set(JwtUUID, uuidStr)

	// Signing the token (using raw rsa.PrivateKey)
	// also marshalls to JSON
	signed, err := jwt.Sign(t, jwa.RS512, auth.privateKey)
	if err != nil {
		return nil, err
	}

	// TODO expiry time?
	return &http.Cookie{
		Name:     JwtCookieName,
		Value:    string(signed[:]),
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	}, nil
}

// Middleware checks if user is authenticated via valid token
func (amw *authenticationMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Debug(r.Cookies())
		// get cookie from request
		jwtCookie, err := r.Cookie(JwtCookieName)
		if err != nil {
			logger.Warnf("no JWT found in request Cookies %s", err)
			http.Error(w, "no JWT found in Cookies", http.StatusUnauthorized)
			return
		}
		logger.Debug(*jwtCookie)

		token, err := jwt.ParseString(jwtCookie.Value, jwt.WithVerify(jwa.RS512, amw.privateKey.PublicKey))
		if err != nil {
			logger.Warnf("unauthenticated request %s", err)
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		user, userOk := token.Get(JwtUserName)
		uuid, uuidOk := token.Get(JwtUUID)
		if userOk && uuidOk {
			logger.Debugf("user %s with uuid %s authenticated. proceeding to %s", user, uuid, r.URL)
		}
		next.ServeHTTP(w, r)
	})
}

// runs validity checks against a received username
func validUsername(str string) (string, error) {
	// in utf-8 runes
	const maxUsernameLength = 255
	if !utf8.ValidString(str) {
		return "", fmt.Errorf("invalid utf-8 runes in username %s", str)
	}

	if utf8.RuneCountInString(str) > maxUsernameLength {
		return "", fmt.Errorf("username contains too many runes %d, %s", utf8.RuneCountInString(str), str)
	}
	escaped := html.EscapeString(str)
	return escaped, nil
}

// parse client's desired username from json body
func parseToCredentialsFromJSON(creds *Credentials, r io.Reader) error {
	err := json.NewDecoder(r).Decode(&creds)
	return err
}
