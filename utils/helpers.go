package utils

import (
	"fmt"
	"net"
	"strconv"
)

const (
	LOCALHOST = "localhost"
)

type invalidPort struct {
	port        string
	blackListed bool
}

func (err invalidPort) Error() string {
	if err.blackListed {
		return fmt.Sprintf("Invalid value [%s] for port", err.port)
	}
	return fmt.Sprintf("bad port value [%s]", err.port)
}

func checkIfPortValid(rawValue string, allowed []int) error {
	val, err := strconv.Atoi(rawValue)
	if err != nil {
		return invalidPort{port: rawValue}
	}
	isAllowed := false
	for _, listed := range allowed {
		// search if value is allowed
		if listed == val {
			isAllowed = true
			break
		}
	}
	// check if
	if isAllowed || val > 0 {
		return nil
	}
	return invalidPort{port: rawValue, blackListed: true}
}

type invalidIp struct {
	falseIp string
}

func (err invalidIp) Error() string {
	return fmt.Sprintf("is not a valid IP [%s]", err.falseIp)
}

func checkIfIpValid(addr string) error {
	if addr == LOCALHOST {
		return nil
	}

	parsed := net.ParseIP(addr)

	// if valid ParseIP will return something
	if parsed != nil {
		return nil
	}
	return invalidIp{falseIp: addr}
}
