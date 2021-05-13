package web

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"io/ioutil"
	"net/http"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwt"
	"github.zhaw.ch/raycasters/server/utils"
)

const (
	JwtCookieName string = "token"
	JwtUserName   string = "username"
	JwtUUID       string = "uuid"
)

// ErrParseCookie is the error that happens on a parse error
type ErrParseCookie struct {
	logMsg, httpMsg string
	status          int
}

// Error satisfies the error interface
func (err ErrParseCookie) Error() string {
	return fmt.Sprintf("[%d] - [%s]", err.status, err.logMsg)
}

// ErrParseCookie is the error that happens on a parse error
type ErrNoCookie struct {
	logMsg, httpMsg string
	status          int
}

// Error satisfies the error interface
func (err ErrNoCookie) Error() string {
	return fmt.Sprintf("[%d] - [%s]", err.status, err.logMsg)
}

// ErrAuthManagerFailour is the error that is returned once
// there is an authentication error of any kind.
type ErrAuthManagerFailure struct {
	internal error
}

// Error satisfies the error interface
func (err ErrAuthManagerFailure) Error() string {
	return fmt.Sprintf("an error happend while creating the auth manager [%s]", err.Unwrap().Error())
}

// Is satisfies the errors.Is interface
func (err ErrAuthManagerFailure) Is(target error) bool {
	_, ok := target.(ErrAuthManagerFailure)
	return ok
}

// Unwrap satisfies the errors.Unwrap functionality.
func (err ErrAuthManagerFailure) Unwrap() error {
	return err.internal
}

// Create a struct to read the username and password from the request body
type Credentials struct {
	Username string `json:"username"`
}

// holds information needed for validating provided authentication
type authenticationManager struct {
	privateKeyPath string
	privateKey     *rsa.PrivateKey
}

// newAuthManager will create a new auth manager
// using either the given path or creating a new
// internally used RSA private key.
func newAuthManager(privateKeyPath string) (authenticationManager, error) {
	// special case where a new rsa key has to be generated
	if privateKeyPath == "" {
		logger.Debug("Generating a temporary RSA PRIVATE KEY")
		key, err := utils.GenerateRSA()
		if err != nil {
			return authenticationManager{},
				ErrAuthManagerFailure{
					internal: err,
				}
		}
		return authenticationManager{
			privateKeyPath: privateKeyPath,
			privateKey:     key,
		}, nil
	}

	logger.Debug("reading the RSA PRIVATE KEY")

	data, err := ioutil.ReadFile(privateKeyPath)

	if err != nil {
		return authenticationManager{}, ErrAuthManagerFailure{
			internal: err,
		}
	}
	key, err := utils.ParseRSA(data)

	if err != nil {
		return authenticationManager{}, ErrAuthManagerFailure{
			internal: err,
		}
	}

	return authenticationManager{
		privateKeyPath: privateKeyPath,
		privateKey:     key,
	}, nil
}

// AuthConnection creates a closure with the given key and
// returns a route handler that
// creates a JWT, signs it with the provided key
// and adds it to the Response Header via Set-Cookie
func (auth *authenticationManager) AuthConnection() (HandleFunc func(w http.ResponseWriter, r *http.Request)) {
	return func(w http.ResponseWriter, r *http.Request) {
		// check if there already is a valid JWT token
		err := auth.checkIfCookieValid(r)
		// not getting an error here means that the
		// user already is authenticated.
		if err == nil {
			return
		}

		var creds Credentials
		err = parseToCredentialsFromJSON(&creds, r.Body)

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

		// Finally, we request to set the client cookie to the JWT we just generated
		http.SetCookie(w, cookie)
	}
}

// checkIfCookieValid will check if there is a valid cookie
// if not this will return an error
func (amw *authenticationManager) checkIfCookieValid(r *http.Request) error {
	logger.Debug("Checking for a cookie")
	cookie, err := amw.findCookie(r)
	if err != nil {
		return nil
	}

	logger.Debug("cookie found")
	_, err = amw.parseCookie(cookie)
	if err != nil {
		logger.Debug("valid cookie found")
	}
	return err
}

// createCookie will create a standard issue JWT cookie.
func (auth *authenticationManager) createCookie(creds *Credentials) (*http.Cookie, error) {
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

	// TODO: expiry time?
	return &http.Cookie{
		Name:     JwtCookieName,
		Value:    string(signed[:]),
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	}, nil
}

// findCookie will look into the request header and
// search for the JWT cookie.
func (amw *authenticationManager) findCookie(r *http.Request) (*http.Cookie, error) {
	// get cookie from request
	jwtCookie, err := r.Cookie(JwtCookieName)
	if err != nil {
		return nil, ErrNoCookie{
			logMsg:  fmt.Sprintf("no JWT found in request Cookies %s", err.Error()),
			httpMsg: "no JWT found in Cookies",
			status:  http.StatusUnauthorized,
		}
	}
	return jwtCookie, nil
}

// parseCookie will parse the cookie if possible to
// check for its validity.
func (amw *authenticationManager) parseCookie(jwtCookie *http.Cookie) (jwt.Token, error) {

	logger.Debug(*jwtCookie)

	token, err := jwt.ParseString(jwtCookie.Value, jwt.WithVerify(jwa.RS512, amw.privateKey.PublicKey))

	if err != nil {
		return nil, &ErrParseCookie{
			logMsg:  fmt.Sprintf("unauthenticated request %s", err.Error()),
			httpMsg: "unauthenticated request",
			status:  http.StatusUnauthorized,
		}
	}
	return token, nil
}

// Middleware checks if user is authenticated via valid token
func (amw *authenticationManager) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := amw.findCookie(r)

		if err != nil {
			// we know we are talking about a specific error
			errHandler := err.(ErrNoCookie)
			logger.Info(errHandler.httpMsg)
			http.Error(w, errHandler.httpMsg, errHandler.status)
			return
		}

		token, err := amw.parseCookie(cookie)

		if err != nil {
			// we know we are talking about a specific error
			errHandler := err.(ErrParseCookie)
			logger.Info(errHandler.httpMsg)
			http.Error(w, errHandler.httpMsg, errHandler.status)
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

// userInfo is an internal struct used
// to transport the user information out
// of the parse funtion.
type userInfo struct {
	uid  uuid.UUID
	name string
}

// parseUserInformation will parse the user information
func parseUserInformation(w http.ResponseWriter, r *http.Request) (userInfo, error) {

	// get cookie from request
	jwtCookie, err := r.Cookie(JwtCookieName)
	if err != nil {
		logger.Debugf("no JWT token found in cookie")
		http.Error(w, "no JWT found in Cookies", http.StatusForbidden)
		return userInfo{}, nil
	}

	// unmarshal the token
	// error is not needed as it is directly send from here
	// and validated previously.
	token, _ := jwt.ParseString(jwtCookie.Value)
	usernameI, userOk := token.Get(JwtUserName)
	uuidI, uuidOk := token.Get(JwtUUID)

	if !(userOk && uuidOk) {
		logger.Debug("incorrect JWT token found in token cookie")
		http.Error(w, "incorrect JWT token found in Cookies", http.StatusBadRequest)

		return userInfo{}, nil
	}

	username, userOk := usernameI.(string)

	const unexpectedError = "unexpected error happened - please contact the service administrator"

	if !userOk {
		logger.Debugf("unable to string convert the user name")
		http.Error(w, unexpectedError, http.StatusInternalServerError)
		return userInfo{}, nil
	}

	// check for username length and escape it
	username, err = validUsername(username)

	if err != nil {
		logger.Debugf("unable to string convert the user name")
		http.Error(w, unexpectedError, http.StatusInternalServerError)
		return userInfo{}, err
	}

	// as the uuid might be both a string as well as a uuid unit
	// using e type switch here to figure its type out

	var uuidVal uuid.UUID

	switch parserd := uuidI.(type) {
	case string:
		uuidVal, err = uuid.Parse(parserd)
		if err != nil {
			logger.Debugf("unable to parse the uuid from [%s]", parserd)
			http.Error(w, unexpectedError, http.StatusInternalServerError)
			return userInfo{}, nil
		}
	case uuid.UUID:
		uuidVal = parserd
	default:
		logger.Debugf("unable to string convert the user uuid")
		http.Error(w, unexpectedError, http.StatusInternalServerError)
		return userInfo{}, nil
	}
	return userInfo{
		name: username,
		uid:  uuidVal,
	}, nil
}
