package web

type WebConfig struct {
	Ip string
	// Is the given port, it is a unint16 as the largest possible port is 65,535.
	Port uint16
}
