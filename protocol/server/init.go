package server

type ServerOpcode uint8

const (
	Disconnect          ServerOpcode = 0x64
	UnknownOpcode       ServerOpcode = 0x65
	MovementPlayer      ServerOpcode = 0x6E
	AnOtherPlayerKilled ServerOpcode = 0x78
	PlayerHimSelfKilled ServerOpcode = 0x79
	NewPlayerJoined     ServerOpcode = 0x7A
	PlayerLeft          ServerOpcode = 0x7B
	SendAllPlayerInfo   ServerOpcode = 0x82
)

var overview = [...]ServerOpcode{
	Disconnect,
	UnknownOpcode,
	MovementPlayer,
	AnOtherPlayerKilled,
	PlayerHimSelfKilled,
	NewPlayerJoined,
	PlayerLeft,
	SendAllPlayerInfo,
}

func Detect(opcode uint8) ServerOpcode {
	parsedOpcode := ServerOpcode(opcode)

	for _, element := range overview {
		if parsedOpcode == element {
			return element
		}
	}

	return UnknownOpcode
}
