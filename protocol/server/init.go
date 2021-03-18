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

func (s ServerOpcode) Value() uint8 {
	// SAFETY: this is safe as the value type of
	// ServerOpcode is already based on the uinit8
	// type.
	return uint8(s)
}

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

func NewDisconnect() Disconnect {
	return Disconnect{}
}

func (d Disconnect) Bytes() []byte {
	return []byte{DisconnectOpcode.Value()}
}

type Unknown struct{}

func NewUnknown() Unknown {
	return Unknown{}
}

func (u Unknown) Bytes() []byte {
	return []byte{UnknownOpcode.Value()}
}

type MovementPlayer struct {
	userId       uint8
	coorX, coorY float64
}

func NewMovementPlayer(userId uint8, coorX, coorY float64) MovementPlayer {
	return MovementPlayer{
		userId: userId,
		coorX:  coorX,
		coorY:  coorY,
	}
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

func (mp MovementPlayer) Bytes() []byte {
	size := 1
	size += 1
	size += 8
	size += 8

	// using the buffer write interface for simplisity
	buffer := make([]byte, size)

	buffer[0] = MovementPlayerOpcode.Value()
	buffer[1] = mp.userId

	// write floats into the bytes buffer
	FloatToBinary(mp.coorX, buffer[2:10])
	FloatToBinary(mp.coorY, buffer[10:18])

	return buffer
}

type AnOtherPlayerKilled struct {
	userId uint8
}

func NewAnotherPlayerKilled(other byte) AnOtherPlayerKilled {
	return AnOtherPlayerKilled{
		userId: other,
	}
}

func (pk AnOtherPlayerKilled) UserId() uint8 {
	return pk.userId
}

func (pk AnOtherPlayerKilled) Bytes() []byte {
	return []byte{
		AnOtherPlayerKilledOpcode.Value(),
		pk.userId,
	}
}

type PlayerHimSelfKilled struct{}

func NewPlayerHimSelfKilled() PlayerHimSelfKilled {
	return PlayerHimSelfKilled{}
}

func (p PlayerHimSelfKilled) Bytes() []byte {
	return []byte{
		PlayerHimSelfKilledOpcode.Value(),
	}
}

type Player struct {
	userId uint8
	name   string
}

func (p Player) String() string {
	return fmt.Sprintf("Id: %d\n Name: %s", p.userId, p.name)
}

func NewPlayer(userId uint8, name string) Player {
	if math.MaxInt16 <= len(name) {
		panic(fmt.Sprintf("Maximum name lenght is a %d", math.MaxInt16))
	}

	return Player{userId: userId, name: name}
}

// Will contruct a new player from a byte array.
// Will return the offset where the player stops.
func NewPlayerBytes(info []byte) (Player, int) {
	userId := info[0]
	lenName := binary.LittleEndian.Uint16(info[1:3])
	size := 1 + 2 + lenName
	name := string(info[3 : 3+lenName])
	return NewPlayer(userId, name), int(size)
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

// A wrapper function for PutBytes, where the byte representation is
// returned instead of written into.
func (p Player) Bytes() []byte {
	// UserId NameSize Name
	data := make([]byte, p.ByteSize())
	p.PutBytes(data)
	return data
}

// This function will write from the zero position on of the slice.
// This function will panic, if the data slice is not at least as long
// as the Player.ByteSize.
func (p Player) PutBytes(data []byte) {
	// this eliminates the bound check below
	// see https://go101.org/article/bounds-check-elimination.html
	// for more information on the topic.
	_ = data[p.ByteSize()-1]

	data[0] = p.userId
	binary.LittleEndian.PutUint16(data[1:3], uint16(len(p.name)))
	copy(data[3:], p.name)
}

type NewPlayerJoined struct {
	p Player
}

func NewNewPlayerJoined(p Player) NewPlayerJoined {
	return NewPlayerJoined{
		p: p,
	}
}

func (p NewPlayerJoined) Player() Player {
	return p.p
}

func (p NewPlayerJoined) Bytes() []byte {
	size := 1                     // opcode
	size += p.Player().ByteSize() // Player Size

	data := make([]byte, size)

	data[0] = NewPlayerJoinedOpcode.Value()
	p.p.PutBytes(data[1:])

	return data
}

type PlayerLeft struct {
	userId uint8
}

func NewPlayerLeft(userId uint8) PlayerLeft {
	return PlayerLeft{
		userId: userId,
	}
}

func (p PlayerLeft) UserId() uint8 {
	return p.userId
}

func (p PlayerLeft) Bytes() []byte {
	return []byte{
		PlayerLeftOpcode.Value(),
		p.userId,
	}
}

type SendAllPlayerInfo struct {
	players []Player
}

func NewSendAllPlayerInfo(players []Player) *SendAllPlayerInfo {
	return &SendAllPlayerInfo{
		players: players,
	}
}

// This assumes the full binary stream to be used
// so with opcode.
// Attention this is a function used for testing
// purposes only, so a multitude of security
// checks are not written.
func NewSendAllPlayerInfoBytes(data []byte) *SendAllPlayerInfo {
	// Op PlCount (UserId Name) * PlCount
	playerCount := data[1]
	players := make([]Player, 0, playerCount)

	// offset
	count := 2
	for count < len(data) {
		player, offset := NewPlayerBytes(data[count:])
		players = append(players, player)
		count += offset
	}

	return &SendAllPlayerInfo{
		players: players,
	}
}

func (p *SendAllPlayerInfo) Players() []Player {
	return p.players
}

func (p *SendAllPlayerInfo) Bytes() []byte {
	// precalculate bytesize
	size := 1 // opcode
	size += 1 // player count
	for _, player := range p.players {
		size += player.ByteSize()
	}
	// initialize buffer
	buffer := make([]byte, size)

	buffer[0] = SendAllPlayerInfoOpcode.Value()
	buffer[1] = uint8(len(p.players))
	count := 2 // index location inside the buffer
	// build up dynamic buffer content
	for _, player := range p.players {
		player.PutBytes(buffer[count : count+player.ByteSize()])
		count += player.ByteSize()
	}

	return buffer
}

// TODO: Refator into into utils
// Will insert the num into data, starting
// from the index 0 on. It will utilize
// the IEEE specification.
// Attention this function assumes that
// the data is already initialized.
func FloatToBinary(num float64, data []byte) {

	// convert the float into the int64 representation
	intRep := math.Float64bits(num)

	// write data back
	binary.LittleEndian.PutUint64(data, intRep)
}
