package client

type ClientOpcode uint8

const (
	Disconnect          ClientOpcode = 0x00
	UnknownOpcode       ClientOpcode = 0x01
	Movement            ClientOpcode = 0x0A
	Shooting            ClientOpcode = 0x14
	RequestOtherPlayers ClientOpcode = 0x1E
)

var overview = [...]ClientOpcode{
	Disconnect,
	Movement,
	Shooting,
	RequestOtherPlayers,
}

func Detect(opcode uint8) ClientOpcode {
	parsedOpcode := ClientOpcode(opcode)

	for _, element := range overview {
		if parsedOpcode == element {
			return element
		}
	}

	return UnknownOpcode
}
