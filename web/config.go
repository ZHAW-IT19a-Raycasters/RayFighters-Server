package web

import "fmt"

type WebConfig struct {
	Ip string
	// Is the given port, it is a unint16 as the largest possible port is 65,535.
	Port uint16
}

func (wc *WebConfig) String() string {
	return fmt.Sprintf("%s:%d", wc.Ip, wc.Port)
}
