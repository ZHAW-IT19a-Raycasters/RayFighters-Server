package server

type ServerOpcode uint8

const (
	DisconnectOpcode          ServerOpcode = 0x64
	UnknownOpcode             ServerOpcode = 0x65
	MovementPlayerOpcode      ServerOpcode = 0x6E
	AnOtherPlayerKilledOpcode ServerOpcode = 0x78
	PlayerHimSelfKilledOpcode ServerOpcode = 0x79
	NewPlayerJoinedOpcode     ServerOpcode = 0x7A
	PlayerLeftOpcode          ServerOpcode = 0x7B
	SendAllPlayerInfoOpcode   ServerOpcode = 0x82
)

var overview = [...]ServerOpcode{
	DisconnectOpcode,
	UnknownOpcode,
	MovementPlayerOpcode,
	AnOtherPlayerKilledOpcode,
	PlayerHimSelfKilledOpcode,
	NewPlayerJoinedOpcode,
	PlayerLeftOpcode,
	SendAllPlayerInfoOpcode,
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

type Disconnect struct{}

type Unknown struct{}

type MovementPlayer struct {
	userId       uint8
	coorX, coorY float64
}

func (mp MovementPlayer) UserId() uint8 {
	return mp.userId
}

func (mp MovementPlayer) CoorX() float64 {
	return mp.coorX
}

func (mp MovementPlayer) CoorY() float64 {
	return mp.coorY
}

type AnOtherPlayerKilled struct {
	userId uint8
}

func (pk AnOtherPlayerKilled) UserId() uint8 {
	return pk.userId
}

type PlayerHimSelfKilled struct{}

type Player struct {
	userId uint8
	name   string
}

func (p Player) UserId() uint8 {
	return p.userId
}

func (p Player) Name() string {
	return p.name
}

type NewPlayerJoined struct {
	p Player
}

func (p NewPlayerJoined) Player() Player {
	return p.p
}

type PlayerLeft struct {
	userId uint8
}

func (p PlayerLeft) UserId() uint8 {
	return p.userId
}

type SendAllPlayerInfo struct {
	players []Player
}

func (p *SendAllPlayerInfo) Players() []Player {
	return p.players
}
