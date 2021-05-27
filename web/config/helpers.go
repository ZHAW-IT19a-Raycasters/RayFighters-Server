package webconfig

import (
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"net"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.zhaw.ch/raycasters/server/utils"
)

// Is the 127.0.0.1 valid alternative
const localhost string = "localhost"

var (
	// getCurrentUser is extracted so that
	// it can get mocked in tests to achieve
	// always the same result.
	getCurrentUser = user.Current

	// getAbsPath is the function to get the
	// absolute file path from an initial string.
	// Attention this is for mocking only so as
	// for force an error here.
	getAbsPath = filepath.Abs

	// uriMatcher will match against the allowed
	// uri schemes used for cors
	uriMatcher = regexp.MustCompile("^https?$")
)

var (
	// ErrIsDirectory is the error called if the file is a
	// directory instead of a file.
	ErrIsDirectory = errors.New("file is a directory")

	// ErrLocalhostNotValidIp is the exception to the rule
	// of valid ip addresses for golang.
	ErrLocalhostNotValidIp = errors.New("localhost is not an allowed ip, please use 127.0.0.1 instead")

	// ErrEmptyCORSDomain is the error returned, when there where no
	// domains added to the CORS configuration.
	ErrEmptyCORSDomain = errors.New("at least one CORS domain has to be set")

	// ErrIncorrectCORSDomainScheme is the error returned once
	// the uri scheme is incorrect.
	ErrIncorrectCORSDomainScheme = errors.New("incorrect uri scheme used, we only support http or https")

	// ErrIncorrectCORSDomainScheme is the error returned once
	// the uri scheme is incorrect.
	ErrIncorrectCORSDomainPath = errors.New("incorrect uri path used, as per the CORS domain definition no paths are allowed")
)

// ErrInvalidCORSConfiguration is the wrapper error that
// handles all the CORS errors.
type ErrInvalidCORSConfiguration struct {
	internal error
}

// Error satisfies the error requirements.
func (err ErrInvalidCORSConfiguration) Error() string {
	return fmt.Sprintf("unable to use the CORS domains because of [%s]", err.Unwrap())
}

// Is satisfies the errors.Is requirements.
func (err ErrInvalidCORSConfiguration) Is(target error) bool {
	_, ok := target.(ErrInvalidCORSConfiguration)
	return ok
}

// Unwrap satisfies the errors.Unwrap requirements.
func (err ErrInvalidCORSConfiguration) Unwrap() error {
	return err.internal
}

// ErrUnableToParseUserHome is the error that
// is returned in the case that something went
// wrong while parsing a complicated user home
// path.
type ErrUnableToParseUserHome struct {
	path     string
	internal error
}

// Error satisfies the error requirements.
func (err ErrUnableToParseUserHome) Error() string {
	return fmt.Sprintf("unable to parse the path [%s] because of [%s]", err.path, err.internal)
}

// Is satisfies the errors.Is requirements.
func (err ErrUnableToParseUserHome) Is(target error) bool {
	_, ok := target.(ErrUnableToParseUserHome)
	return ok
}

// Unwrap satisfies the errors.Unwrap requirements.
func (err ErrUnableToParseUserHome) Unwrap() error {
	return err.internal
}

// Represents an invalid port
type InvalidPortError struct {
	port        string
	blackListed bool
}

// The error interface
func (err InvalidPortError) Error() string {
	if err.blackListed {
		return fmt.Sprintf("Invalid value [%s] for port", err.port)
	}
	return fmt.Sprintf("bad port value [%s]", err.port)
}

// Represents an invalid IP Error
type InvalidIpError struct {
	falseIp string
}

// The error interface
func (err InvalidIpError) Error() string {
	return fmt.Sprintf("is not a valid IP [%s]", err.falseIp)
}

// InvalidRSAKeyPath is the error returned when there is
// no such RSA private key file found.
type InvalidRSAKeyPath struct {
	path     string
	internal error
}

// The error interface
func (err InvalidRSAKeyPath) Error() string {
	return fmt.Sprintf("file [%s] is not a valid RSA private key -> [%s]", err.path, err.internal.Error())
}

// Is satisfies the errors.Is requirements.
func (err InvalidRSAKeyPath) Is(target error) bool {
	_, ok := target.(InvalidRSAKeyPath)
	return ok
}

// Unwrap satisfies the errors.Unwrap requirements.
func (err InvalidRSAKeyPath) Unwrap() error {
	return err.internal
}

// Checks if a port is valid or not
func CheckIfPortValid(rawValue string, allowed []int) error {
	// check if value can be turned into a integer
	val, err := strconv.Atoi(rawValue)
	if err != nil {
		return InvalidPortError{port: rawValue}
	}

	if CheckIfPortIsAllowed(val, allowed) {
		return nil
	}

	return InvalidPortError{port: rawValue, blackListed: true}
}

// Checks if the port is whitelisted
// As per definition the valid port range is 0 - 65535.
// Where as this function allows eceptions to be used
// mainly for defaults etc.
func CheckIfPortIsAllowed(val int, allowed []int) bool {
	isAllowed := false
	for _, listed := range allowed {
		// search if value is allowed
		if listed == val {
			isAllowed = true
			break
		}
	}
	// check if
	return isAllowed || (val >= 0 && val <= math.MaxUint16)
}

// CheckIfIpValid will check as the name implies
// if the ip is valid.
// Attention as "localhost" is not a valid
// ip as of golang it is not allowed to be
// used.
func CheckIfIpValid(addr string) error {
	if addr == localhost {
		return ErrLocalhostNotValidIp
	}

	parsed := net.ParseIP(addr)

	// if valid ParseIP will return something
	if parsed != nil {
		return nil
	}
	return InvalidIpError{falseIp: addr}
}

// fixUserHomePath will fix the given path
// in case that the path starts with a `~`
// Attention this function assumes a trimmed
// path.
func fixUserHomePath(path string) (string, error) {
	// check if we are talking to a home user
	if !strings.HasPrefix(path, "~") {
		return path, nil
	}

	usr, err := getCurrentUser()
	if err != nil {
		return "", ErrUnableToParseUserHome{
			path:     path,
			internal: err,
		}
	}

	dir := usr.HomeDir
	res := dir

	if strings.HasPrefix(path, "~/") {
		res = filepath.Join(res, path[2:])
	}

	return res, nil
}

// CheckIfRSAValid will check if the file given by the path
// is valid according to the internal requirements.
func CheckIfRSAValid(path string) error {
	// remove any excess
	path = strings.TrimSpace(path)

	// handle linux paths
	path, err := fixUserHomePath(path)

	if err != nil {
		return InvalidRSAKeyPath{
			path:     path,
			internal: err,
		}
	}

	// check if there are invalid path elements
	absPath, err := getAbsPath(path)
	if err != nil {
		return InvalidRSAKeyPath{
			path:     path,
			internal: err,
		}
	}

	// check if the file exists
	fileInfo, err := os.Stat(absPath)
	if err != nil {
		return InvalidRSAKeyPath{
			path:     path,
			internal: err,
		}
	}

	// check if the path is a directory
	if fileInfo.IsDir() {
		return InvalidRSAKeyPath{
			path:     path,
			internal: ErrIsDirectory,
		}
	}

	// we do not have to check the error, as we
	// previously checked if the file exists.
	rsaData, _ := ioutil.ReadFile(absPath)

	// check that is is a valid RSA file
	_, err = utils.ParseRSA(rsaData)

	if err != nil {
		return InvalidRSAKeyPath{
			path:     path,
			internal: err,
		}
	}

	return nil
}

// CheckIfCORSDomainValid check if the given
// domain is valid or not.
func CheckIfCORSDomainValid(domain string) error {

	// using the parse url request to check
	// if the given domain is valid or not
	uri, err := url.ParseRequestURI(domain)
	if err != nil {
		return ErrInvalidCORSConfiguration{
			internal: err,
		}
	}

	// check scheme
	if !uriMatcher.Match([]byte(uri.Scheme)) {
		return ErrIncorrectCORSDomainScheme
	}

	// check that there is no path used
	if strings.TrimSpace(uri.Path) != "" {
		return ErrIncorrectCORSDomainPath
	}

	return nil
}
