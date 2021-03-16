package client

type ClientOpcode uint8

const (
	DisconnectOpcode          ClientOpcode = 0x00
	UnknownOpcode             ClientOpcode = 0x01
	MovementOpcode            ClientOpcode = 0x0A
	ShootingOpcode            ClientOpcode = 0x14
	RequestOtherPlayersOpcode ClientOpcode = 0x1E
)

var overview = [...]ClientOpcode{
	DisconnectOpcode,
	MovementOpcode,
	ShootingOpcode,
	RequestOtherPlayersOpcode,
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

type Disconnect struct{}

type Unknown struct{}

type Movement struct {
	coorX, coorY float64
}

func (m Movement) CoorX() float64 {
	return m.coorX
}

func (m Movement) CoorY() float64 {
	return m.coorY
}

type Shooting struct {
	angPl float64
}

func (s Shooting) AngPl() float64 {
	return s.angPl
}

type RequestOtherPlayers struct{}
