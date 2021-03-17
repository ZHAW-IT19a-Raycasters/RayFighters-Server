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

func NewSendAllPlayerInfo(players []Player) *SendAllPlayerInfo {
	return &SendAllPlayerInfo{players: players}
}

// This assumes the full binary stream to be used
// so with opcode.
// Attention this is a function used for testing 
// purposes only.
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
        
    return &SendAllPlayerInfo{players: players}
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
