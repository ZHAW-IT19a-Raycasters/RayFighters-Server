package config

import (
	"fmt"
	"net"
	"strconv"
)

const (
	// Is the 127.0.0.1 valid alternative
	localhost string = "localhost"
)

// Represents an invalid port
type invalidPortError struct {
	port        string
	blackListed bool
}

// The error interface
func (err invalidPortError) Error() string {
	if err.blackListed {
		return fmt.Sprintf("Invalid value [%s] for port", err.port)
	}
	return fmt.Sprintf("bad port value [%s]", err.port)
}

// Checks if a port is valid or not
func CheckIfPortValid(rawValue string, allowed []int) error {
	// check if value can be turned into a integer
	val, err := strconv.Atoi(rawValue)
	if err != nil {
		return invalidPortError{port: rawValue}
	}

	if CheckIfPortIsAllowed(val, allowed) {
		return nil
	}

	return invalidPortError{port: rawValue, blackListed: true}
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

// Represents an invalid IP Error
type invalidIpError struct {
	falseIp string
}

func (err invalidIpError) Error() string {
	return fmt.Sprintf("is not a valid IP [%s]", err.falseIp)
}

func CheckIfIpValid(addr string) error {
	if addr == localhost {
		return nil
	}

	parsed := net.ParseIP(addr)

	// if valid ParseIP will return something
	if parsed != nil {
		return nil
	}
	return invalidIpError{falseIp: addr}
}
