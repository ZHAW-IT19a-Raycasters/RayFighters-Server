package config

import "fmt"

type WebConfig struct {
	// Is the given ip, this may be either localhost or any ip4 / ip6 value.
	Ip string
	// Is the given port, it is a unint16 as the largest possible port is 65,535.
	Port uint16
}

func (wc *WebConfig) String() string {
	return fmt.Sprintf("%s:%d", wc.Ip, wc.Port)
}

func (wc *WebConfig) CheckConfiguration() error {

	return nil
}
