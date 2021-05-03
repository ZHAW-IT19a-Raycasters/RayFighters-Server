package webconfig

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strconv"

	"github.zhaw.ch/raycasters/server/utils"
)

const (
	// Is the 127.0.0.1 valid alternative
	localhost string = "localhost"
)

// ErrISDirectory is the error called if the file is a
// directory instead of a file.
var ErrISDirectory = errors.New("file is a directory")

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

func (err InvalidIpError) Error() string {
	return fmt.Sprintf("is not a valid IP [%s]", err.falseIp)
}

// InvalidRSAKeyPath is the error returend when there is
// no such RSA private key file found.
type InvalidRSAKeyPath struct {
	path     string
	internal error
}

func (err InvalidRSAKeyPath) Error() string {
	return fmt.Sprintf("file does not exist [%s] or [%s]", err.path, err.internal.Error())
}

func (err InvalidRSAKeyPath) Is(target error) bool {
	_, ok := target.(InvalidRSAKeyPath)
	return ok
}

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
	return isAllowed || val > 0
}

// CheckIfIpValid will check as the name implies
// if the ip is valid.
func CheckIfIpValid(addr string) error {
	if addr == localhost {
		return nil
	}

	parsed := net.ParseIP(addr)

	// if valid ParseIP will return something
	if parsed != nil {
		return nil
	}
	return InvalidIpError{falseIp: addr}
}

func CheckIfJWTValid(path string) error {

	// check if there are invalid path elements
	absPath, err := filepath.Abs(path)
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

	// check that it is a directory
	if fileInfo.IsDir() {
		return InvalidRSAKeyPath{
			path:     path,
			internal: ErrISDirectory,
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
