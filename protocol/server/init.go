package server

import (
	"encoding/binary"
	"fmt"
	"math"
)

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

func NewPlayer(userId uint8, name string) Player {
	if math.MaxInt16 <= len(name) {
		panic(fmt.Sprintf("Maximum name lenght is a %d", math.MaxInt16))
	}

	return Player{userId: userId, name: name}
}

func (p Player) UserId() uint8 {
	return p.userId
}

func (p Player) Name() string {
	return p.name
}

func (p Player) ByteSize() int {
	return 1 + 2 + len(p.name)
}

func (p Player) Bytes() []byte {
	// UserId NameSize Name
	data := make([]byte, p.ByteSize())
	p.PutBytes(data)
	return data
}

// This function will write from the zero position on of the slice.
func (p Player) PutBytes(data []byte) {
	// early bound check, figure out if this is okay...
	_ = data[p.ByteSize()-1]

	data[0] = p.userId
	binary.LittleEndian.PutUint16(data[1:3], uint16(len(p.name)))
	for i, val := range []byte(p.name) {
		data[i+3] = val
	}
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
