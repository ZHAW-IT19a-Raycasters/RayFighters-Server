package webconfig

import "fmt"

// The configuration for the webserver
type WebConfig struct {
	// Is the given ip, this may be either localhost or any ip4 / ip6 value.
	Ip string
	// Is the given port, it is a unint16 as the largest possible port is 65,535.
	Port uint16
}

// The stringify interface
func (wc *WebConfig) String() string {
	return fmt.Sprintf("%s:%d", wc.Ip, wc.Port)
}

// The CheckIfValid interface
func (wc *WebConfig) CheckConfiguration() error {
	err := CheckIfIpValid(wc.Ip)

	if err != nil {
		return err
	}

	allowed := []int{}
	if CheckIfPortIsAllowed(int(wc.Port), allowed) {
		return nil
	}
	return invalidPortError{fmt.Sprintf("%d", wc.Port), true}
}

type JwtConfig struct {
	// RsaPrivateKeyLocation is the location where the rsa key lives in.
	RsaPrivateKeyLocation string
}
