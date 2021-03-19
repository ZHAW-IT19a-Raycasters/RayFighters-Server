// Package server contains all the structs and parsing functionality
// needed for the server side raycaster protocol.
package server

import (
	"encoding/binary"
	"fmt"
	"math"
)

// A Server Opcode represents the internal representations
// of all the defined server side operational
// codes that control the game flow.
type ServerOpcode uint8

const (
	DisconnectOpcode          ServerOpcode = 0x64
	UnknownOpcode             ServerOpcode = 0x65
	MovementPlayerOpcode      ServerOpcode = 0x6E
	AnOtherPlayerKilledOpcode ServerOpcode = 0x78
	PlayerHimSelfKilledOpcode ServerOpcode = 0x79
	PlayerJoinedOpcode        ServerOpcode = 0x7A
	PlayerLeftOpcode          ServerOpcode = 0x7B
	SendAllPlayerInfoOpcode   ServerOpcode = 0x82
)

// Value returns the internal value representation as a uint8.
func (s ServerOpcode) Value() uint8 {
	// SAFETY: this is safe as the value type of
	// ServerOpcode is already based on the uint8
	// type.
	return uint8(s)
}

// Detect converts from the an external byte into the internally
// used ServerOpcode type.
//
// Will return Unknown when there is no mapping for the given value.
func Detect(opcode uint8) ServerOpcode {

	var overview = [...]ServerOpcode{
		DisconnectOpcode,
		UnknownOpcode,
		MovementPlayerOpcode,
		AnOtherPlayerKilledOpcode,
		PlayerHimSelfKilledOpcode,
		PlayerJoinedOpcode,
		PlayerLeftOpcode,
		SendAllPlayerInfoOpcode,
	}

	// cast into a comparable opcode type
	parsedOpcode := ServerOpcode(opcode)

	for _, element := range overview {
		if parsedOpcode == element {
			return element
		}
	}

	return UnknownOpcode
}

// Disconnect represents a server shutdown.
type Disconnect struct{}

// NewDiconnect creates a new disconnect instance.
func NewDisconnect() Disconnect {
	return Disconnect{}
}

// Bytes will turn the disconnect instance into a sendable byte stream.
func (d Disconnect) Bytes() []byte {
	return []byte{DisconnectOpcode.Value()}
}

// Unknown represents the fallback opcode in case there is a
// difference of protocol version or if there is an unknown
// opcode it self send.
type Unknown struct{}

// Will initialize the Unknown opcode struct.
func NewUnknown() Unknown {
	return Unknown{}
}

// Bytes will turn the unknown instance into a sendable byte stream.
func (u Unknown) Bytes() []byte {
	return []byte{UnknownOpcode.Value()}
}

// MovementPlayer represents a movement made by a specific player.
// It contains the user id and the respective offsets size the last
// time the command was send.
type MovementPlayer struct {
	userId       uint8
	coorX, coorY float64
}

// NewMovemementPlayer will construct a new player object.
func NewMovementPlayer(userId uint8, coorX, coorY float64) MovementPlayer {
	return MovementPlayer{
		userId: userId,
		coorX:  coorX,
		coorY:  coorY,
	}
}

// UserId will return the user whom has made the given movement.
func (mp MovementPlayer) UserId() uint8 {
	return mp.userId
}

// CoorX is the offset on the x axis done by the player.
func (mp MovementPlayer) CoorX() float64 {
	return mp.coorX
}

// CoorY is the offset on the y axis done by the player.
func (mp MovementPlayer) CoorY() float64 {
	return mp.coorY
}

// Bytes will turn the movement instance into a sendable byte stream.
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

// AnOtherPlayerKilled is the counter part to PlayerHimSelfKilledOpcode
// it is send to any other active clients in the room.
type AnOtherPlayerKilled struct {
	userId uint8
}

// NewAnotherPlayerKilled will initialize the instantiation of
// NewAnotherPlayerKilled.
func NewAnotherPlayerKilled(other byte) AnOtherPlayerKilled {
	return AnOtherPlayerKilled{
		userId: other,
	}
}

// UserId will return the userId of the player that gotten killed.
func (pk AnOtherPlayerKilled) UserId() uint8 {
	return pk.userId
}

// Bytes will turn the AnOtherPlayerKilled instance into a sendable
// byte stream.
func (pk AnOtherPlayerKilled) Bytes() []byte {
	return []byte{
		AnOtherPlayerKilledOpcode.Value(),
		pk.userId,
	}
}

// PlayerHimSelfKilled is the counterpart to AnOtherPlayerKilled
// as it deactivates the players abilities to move or shoot.
type PlayerHimSelfKilled struct{}

// NewPlayerHimSelfKilled will create the corresponding
// instance.
func NewPlayerHimSelfKilled() PlayerHimSelfKilled {
	return PlayerHimSelfKilled{}
}

// Bytes will turn the AnOtherPlayerKilled instance into a sendable
// byte stream.
func (p PlayerHimSelfKilled) Bytes() []byte {
	return []byte{
		PlayerHimSelfKilledOpcode.Value(),
	}
}

// Player is an internally used representation of the
// relative complex data type called player.
type Player struct {
	userId uint8
	name   string
}

// String will return a printable representation of any
// given player.
func (p Player) String() string {
	return fmt.Sprintf("Id: %d\n Name: %s", p.userId, p.name)
}

// NewPlayer will create a new player from the given
// input parameters.
//
// Attention this function can panic if the player
// name is larger then the larges uint16 value possible.
// (65'535)
func NewPlayer(userId uint8, name string) Player {
	if math.MaxUint16 <= len(name) {
		panic(fmt.Sprintf("Maximum name lenght is a %d", math.MaxUint16))
	}

	return Player{userId: userId, name: name}
}

// Will construct a new player from a byte array.
// Will return the offset where the player stops.
//
// Attention this function will not return an
// error if any of the given lenghts are incorrect,
// as this function should never actually be used
// on the server side, but is primarily used for
// testing purposes.
//
// Needed are:
//    - UserId           (1 byte)
//    - StringLength [n] (2 bytes)
//    - String           (n bytes)
func NewPlayerBytes(info []byte) (Player, int) {
	userId := info[0]
	lenName := binary.LittleEndian.Uint16(info[1:3])
	size := 1 + 2 + lenName
	name := string(info[3 : 3+lenName])
	return NewPlayer(userId, name), int(size)
}

// UserId is a getter for the internal user id.
func (p Player) UserId() uint8 {
	return p.userId
}

// Name is a getter of the player name.
func (p Player) Name() string {
	return p.name
}

// ByteSize is a helper that will return a players size,
// so that a correctly sized buffer can be preallocated.
func (p Player) ByteSize() int {
	return 1 + 2 + len(p.name)
}

// Bytes is wrapper function for PutBytes, where the byte
// representation is returned instead of written into.
func (p Player) Bytes() []byte {
	// UserId NameSize Name
	data := make([]byte, p.ByteSize())
	p.PutBytes(data)
	return data
}

// PutBytes will write the binary representation of the player into
// the `data` parameter, from the zero position on of the slice.
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

// PlayerJoined is the opcode that will be send in the case
// that a new player has joined a room. Just like the name
// implies.
type PlayerJoined struct {
	p Player
}

// NewPlayerJoined will create a new instance of the player joined
// type.
func NewPlayerJoined(p Player) PlayerJoined {
	return PlayerJoined{
		p: p,
	}
}

// Player is a getter for the player that has joined the room.
func (p PlayerJoined) Player() Player {
	return p.p
}

// Bytes will return the information needed to write the player
// information back.
func (p PlayerJoined) Bytes() []byte {
	size := 1                     // opcode
	size += p.Player().ByteSize() // Player Size

	data := make([]byte, size)

	data[0] = PlayerJoinedOpcode.Value()
	p.p.PutBytes(data[1:])

	return data
}

// PlayerLeft represents a player that left the game.
// This opcode would also be send after a disconnect
// timeout.
type PlayerLeft struct {
	userId uint8
}

// NewPlayerLeft create an opcode representing a player leaving the board.
func NewPlayerLeft(userId uint8) PlayerLeft {
	return PlayerLeft{
		userId: userId,
	}
}

// UserId of the player leaving is returned.
func (p PlayerLeft) UserId() uint8 {
	return p.userId
}

// Bytes will return the information needed to inform
// any clients in the room that the player left.
func (p PlayerLeft) Bytes() []byte {
	return []byte{
		PlayerLeftOpcode.Value(),
		p.userId,
	}
}

// SendAllPlayerInfo is a request that will have to return any and
// all player information requiered for operations.
type SendAllPlayerInfo struct {
	players []Player
}

// NewSendAllPlayerInfo will create the corresponding opcode type.
func NewSendAllPlayerInfo(players []Player) *SendAllPlayerInfo {
	return &SendAllPlayerInfo{
		players: players,
	}
}

// NewSendAllPlayerInfoBytes will read all the player
// information from the data stream and return a deserialized
// overview of all the players.
//
// This assumes the full binary stream to be used
// so with opcode.
//
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

// Players will return all the players in the current room.
func (p *SendAllPlayerInfo) Players() []Player {
	return p.players
}

// Bytes will return a correct byte stream with all the required
// player information.
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
//
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
