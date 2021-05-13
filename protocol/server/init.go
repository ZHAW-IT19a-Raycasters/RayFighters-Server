// Package server contains all the structs and parsing functionality
// needed for the server side raycaster protocol.
package server

import (
	"encoding/binary"
	"fmt"
	"math"

	"github.zhaw.ch/raycasters/server/utils"
)

// A Server Opcode represents the internal representations
// of all the defined server side operational
// codes that control the game flow.
type ServerOpcode uint8

const (
	coordSize       = 8
	angleSize       = 8
	playerCoutnSize = 1
	stringCountSize = 2
	opcodeSize      = 1
	userIdSize      = 1
)

const (
	DisconnectOpcode        ServerOpcode = 0x64
	UnknownOpcode           ServerOpcode = 0x65
	MovementPlayerOpcode    ServerOpcode = 0x6E
	TeleportPlayerOpcode    ServerOpcode = 0x6F
	ShootingPlayerOpcode    ServerOpcode = 0x70
	PlayerDeathOpcode       ServerOpcode = 0x78
	PlayerJoinedOpcode      ServerOpcode = 0x7A
	PlayerLeftOpcode        ServerOpcode = 0x7B
	YouAreCalledOpcode      ServerOpcode = 0x7C
	SendAllPlayerInfoOpcode ServerOpcode = 0x82
	SendAllPlayerKDOpcode   ServerOpcode = 0x83
	MapChangeOpcode         ServerOpcode = 0x8C
)

var Overview = [...]ServerOpcode{
	DisconnectOpcode,
	UnknownOpcode,
	MovementPlayerOpcode,
	TeleportPlayerOpcode,
	ShootingPlayerOpcode,
	PlayerDeathOpcode,
	PlayerJoinedOpcode,
	PlayerLeftOpcode,
	YouAreCalledOpcode,
	SendAllPlayerInfoOpcode,
	SendAllPlayerKDOpcode,
	MapChangeOpcode,
}

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
	// cast into a comparable opcode type
	parsedOpcode := ServerOpcode(opcode)

	for _, element := range Overview {
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

// Disconnect represents a server shutdown.
type PlayerDeath struct {
	killedId uint8
	killerId uint8
}

func (pd PlayerDeath) KilledId() uint8 {
	return pd.killedId
}

func (pd PlayerDeath) KillerId() uint8 {
	return pd.killerId
}

// NewDisconnect creates a new disconnect instance.
func NewPlayerDeath(killedId uint8, killerId uint8) PlayerDeath {
	return PlayerDeath{
		killedId: killedId,
		killerId: killerId,
	}
}

func (pq PlayerDeath) ByteSize() uint {
	return opcodeSize + userIdSize + userIdSize
}

// Bytes will turn the disconnect instance into a sendable byte stream.
func (pd PlayerDeath) Bytes() []byte {
	// using the buffer write interface for simplisity
	buffer := make([]byte, pd.ByteSize())

	buffer[0] = PlayerDeathOpcode.Value()
	buffer[1] = pd.killedId
	buffer[2] = pd.killerId

	return buffer
}

// TeleportPlayer represents a movement made by a specific player.
// It contains the user id and the respective offsets size the last
// time the command was send.
type TeleportPlayer struct {
	userId       uint8
	coorX, coorY float64
}

// NewMovemementPlayer will construct a new player object.
func NewTeleportPlayer(userId uint8, coorX, coorY float64) TeleportPlayer {
	return TeleportPlayer{
		userId: userId,
		coorX:  coorX,
		coorY:  coorY,
	}
}

// UserId will return the user whom has made the given movement.
func (mp TeleportPlayer) UserId() uint8 {
	return mp.userId
}

// CoorX is the offset on the x axis done by the player.
func (mp TeleportPlayer) CoorX() float64 {
	return mp.coorX
}

// CoorY is the offset on the y axis done by the player.
func (mp TeleportPlayer) CoorY() float64 {
	return mp.coorY
}

func (mp TeleportPlayer) ByteSize() int {
	//     Op          + UserId     + CoorX + CoorY
	return opcodeSize + userIdSize + coordSize*2
}

// Bytes will turn the movement instance into a sendable byte stream.
func (mp TeleportPlayer) Bytes() []byte {

	// using the buffer write interface for simplisity
	buffer := make([]byte, mp.ByteSize())

	buffer[0] = TeleportPlayerOpcode.Value()
	buffer[1] = mp.UserId()

	// write floats into the bytes buffer
	data := [...]float64{mp.CoorX(), mp.CoorY()}
	for i, val := range data {
		offset := 2 + i*coordSize
		offsetEnd := offset + coordSize
		utils.FloatToBinary(val, buffer[offset:offsetEnd])
	}

	return buffer
}

// MovementPlayer represents a movement made by a specific player.
// It contains the user id and the respective offsets size the last
// time the command was send.
type MovementPlayer struct {
	userId                uint8
	coorX, coorY, anglePl float64
}

// NewMovemementPlayer will construct a new player object.
func NewMovementPlayer(userId uint8, coorX, coorY, anglePl float64) MovementPlayer {
	return MovementPlayer{
		userId:  userId,
		coorX:   coorX,
		coorY:   coorY,
		anglePl: anglePl,
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

func (mp MovementPlayer) AnglePl() float64 {
	return mp.anglePl
}

func (mp MovementPlayer) ByteSize() int {
	//     Op          + UserId     + CoorX + CoorY + AnglePl
	return opcodeSize + userIdSize + coordSize*2 + angleSize
}

// Bytes will turn the movement instance into a sendable byte stream.
func (mp MovementPlayer) Bytes() []byte {

	// using the buffer write interface for simplisity
	buffer := make([]byte, mp.ByteSize())

	buffer[0] = MovementPlayerOpcode.Value()
	buffer[1] = mp.UserId()

	// write floats into the bytes buffer
	data := [...]float64{mp.CoorX(), mp.CoorY(), mp.AnglePl()}
	for i, val := range data {
		offset := 2 + i*coordSize
		offsetEnd := offset + coordSize
		utils.FloatToBinary(val, buffer[offset:offsetEnd])
	}

	return buffer
}

type ShootingPlayer struct {
	userId uint8
	anglPl float64
}

func NewShootingPlayer(id uint8, angPl float64) ShootingPlayer {
	return ShootingPlayer{
		userId: id,
		anglPl: angPl,
	}
}

func (sp ShootingPlayer) UserId() uint8 {
	return sp.userId
}

func (sp ShootingPlayer) AnglePl() float64 {
	return sp.anglPl
}

func (sp ShootingPlayer) ByteSize() int {
	return opcodeSize + userIdSize + angleSize
}

func (sp ShootingPlayer) Bytes() []byte {
	data := make([]byte, sp.ByteSize())

	data[0] = ShootingPlayerOpcode.Value()
	data[1] = sp.UserId()
	utils.FloatToBinary(sp.AnglePl(), data[2:])
	return data
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
	return fmt.Sprintf("Id: %d\nName: %s", p.userId, p.name)
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
	size := opcodeSize + stringCountSize + lenName
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
	return opcodeSize + stringCountSize + len(p.name)
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

func (p PlayerJoined) ByteSize() int {
	return opcodeSize + p.Player().ByteSize()
}

// Bytes will return the information needed to write the player
// information back.
func (p PlayerJoined) Bytes() []byte {
	data := make([]byte, p.ByteSize())

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

type YouAreCalled struct {
	userId uint8
}

// UserId of the player leaving is returned.
func (yac YouAreCalled) UserId() uint8 {
	return yac.userId
}

func NewYouAreCalled(userId uint8) YouAreCalled {
	return YouAreCalled{
		userId: userId,
	}
}

func (yac YouAreCalled) Bytes() []byte {
	return []byte{
		YouAreCalledOpcode.Value(),
		yac.userId,
	}
}

type PlayerKD struct {
	id     uint8
	kills  uint16
	deaths uint16
}

type PlayersKD struct {
	playersKD []PlayerKD
}

func (pkd PlayerKD) UserId() uint8 {
	return pkd.id
}

func (pkd PlayerKD) Kills() uint16 {
	return pkd.kills
}

func (pkd PlayerKD) Deaths() uint16 {
	return pkd.deaths
}

func NewPlayerKD(id uint8, kills, deaths uint16) PlayerKD {
	return PlayerKD{
		id:     id,
		kills:  kills,
		deaths: deaths,
	}
}

func (p PlayersKD) ByteSize() int {
	return opcodeSize + playerCoutnSize + len(p.playersKD)*PlayerKD{}.ByteSize()
}

func (p PlayersKD) Bytes() []byte {
	b := make([]byte, p.ByteSize())
	b[0] = SendAllPlayerKDOpcode.Value()
	b[1] = byte(len(p.playersKD))

	for index, playerKD := range p.playersKD {
		playerKD.PutBytes(b[index*playerKD.ByteSize()+2:])
	}
	return b
}

func NewPlayersKD(players []PlayerKD) PlayersKD {
	return PlayersKD{
		playersKD: players,
	}
}

func (p PlayerKD) ByteSize() int {
	return 1 + 2 + 2
}

func (p PlayerKD) PutBytes(data []byte) {
	// this eliminates the bound check below
	// see https://go101.org/article/bounds-check-elimination.html
	// for more information on the topic.
	_ = data[p.ByteSize()-1]

	data[0] = p.id
	binary.LittleEndian.PutUint16(data[1:3], uint16(p.kills))
	binary.LittleEndian.PutUint16(data[3:5], uint16(p.deaths))
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

func (p *SendAllPlayerInfo) ByteSize() int {
	// precalculate bytesize
	size := 1 // opcode
	size += 1 // player count
	for _, player := range p.players {
		size += player.ByteSize()
	}

	return size
}

// Bytes will return a correct byte stream with all the required
// player information.
func (p *SendAllPlayerInfo) Bytes() []byte {

	// initialize buffer
	buffer := make([]byte, p.ByteSize())

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

// MapChange is the opcode that will tell the
// frontend what map is currently chosen
// and at what time it was chosen.
type MapChange struct {
	mapId    uint8
	unixTime int64
}

// NewMapChange will create a new map change opcode
func NewMapChange(mapId uint8, unixTime int64) MapChange {
	return MapChange{
		mapId:    mapId,
		unixTime: unixTime,
	}
}

func (p MapChange) ByteSize() int {
	// Opcode + MapId + time
	return 1 + 1 + 8
}

// Bytes will return a correct byte stream with all the required
// map change information.
func (m MapChange) Bytes() []byte {
	data := make([]byte, m.ByteSize())
	data[0] = MapChangeOpcode.Value()
	data[1] = m.mapId

	utils.Int64ToBinary(m.unixTime, data[2:])

	return data
}
