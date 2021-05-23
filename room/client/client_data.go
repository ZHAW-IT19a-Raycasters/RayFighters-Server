package client

import (
	"github.zhaw.ch/raycasters/server/utils"
	"sync/atomic"

	"github.com/google/uuid"
	"github.zhaw.ch/raycasters/server/protocol/server"
)

// ConnectionStatus is an enum type used for showing the
// state of the
type ConnectionStatus uint

const (
	//Active represents the connections state, where
	// there is nominal client server interactions.
	Active ConnectionStatus = iota
	// Registered represents the state after
	// the connections, was added to the user.
	Registered
	// Timeout represents the state that happens
	// once there are connections errors between
	// client and server. It shows a temporary
	// waiting state between setting the client
	// offline and still being connected, giving
	// a change to reconnect to the server.
	//
	// Attention this state also means that the
	// previous connections is dead.
	Timeout
	// ShuttingDown is a temporary state where
	// the client connection is in while
	// waiting for the workers to shutdown
	ShuttingDown
	// Offline is the state, where the client
	// awaits deletion.
	Offline
)

// String satisfies the stringer interface and shows the string
// representations of the connection status.
func (c ConnectionStatus) String() string {
	vals := [...]string{
		"Active", "Registered", "Timeout", "ShuttingDown", "Offline",
	}
	return vals[c]
}

type Position struct {
	x float64
	y float64
}

// ClientGame represents the state of the client in the game
type ClientGame struct {
	kills     uint32
	deaths	  uint32
	IsDead	  utils.AtomicBool
}

// NewClientGame creates the default client object.
func NewClientGame() ClientGame {
	return ClientGame{
		kills:    0,
		deaths:   0,
		IsDead:	  utils.NewAtomicBool(false),
	}
}

// GetKills will return the current kill count
// of the client.
func (cg *ClientGame) GetKills() uint16 {
	return uint16(atomic.LoadUint32(&cg.kills))
}

// GetDeaths will reutrn the death count
// of the client.
func (cg *ClientGame) GetDeaths() uint16 {
	return uint16(atomic.LoadUint32(&cg.deaths))
}

// IncKills will increment the kill count
// of the player
func (cg *ClientGame) IncKills() {
	atomic.AddUint32(&cg.kills, 1)
}

// IncDeaths will increment the
// death counter by one
func (cg *ClientGame) IncDeaths() {
	atomic.AddUint32(&cg.deaths, 1)
}

// ClientMeta contains the shared meta information of a
// specific client. As well as the current connection used.
type ClientMeta struct {
	id   uuid.UUID
	name string
	conn ClientConnection
}

// NewClientMeta creates a new client meta object.
func NewClientMeta(id uuid.UUID, name string, conn ClientConnection) ClientMeta {
	return ClientMeta{
		id:   id,
		name: name,
		conn: conn,
	}
}

// GetId will reutrn the externally used
// id
func (cm ClientMeta) GetId() uuid.UUID {
	return cm.id
}

// GetName will return the client name
func (cm ClientMeta) GetName() string {
	return cm.name
}

func (cm *ClientMeta) Close() {
	cm.conn.Close()
}

// clientInternalMeta is the internal representation of a client meta
// it also contains the internal id of the client.
type clientInternalMeta struct {
	id   uint8
	eId  uuid.UUID
	name string
	conn ClientConnection
}

// newClientInternalMeta creates a new internal meta package
// from the client meta object.
func newClientInternalMeta(id uint8, client ClientMeta) clientInternalMeta {
	return clientInternalMeta{
		id:   id,
		eId:  client.GetId(),
		name: client.GetName(),
		conn: client.conn,
	}
}

// GetPlayer will return a sendable representation
//  of the client.
func (c clientInternalMeta) GetPlayer() server.Player {
	return server.NewPlayer(c.id, c.name)
}

// getMeta will return the meta information of the client.
func (c clientInternalMeta) GetMeta() ClientMeta {
	return NewClientMeta(c.GetUUID(), c.GetName(), c.conn)
}

// GetName will return the clients registered name
func (c clientInternalMeta) GetName() string {
	return c.name
}

// GetId will return the internal id of the client
func (c clientInternalMeta) GetId() uint8 {
	return c.id
}

// GetUUID will return the external unique id of the
// client.
func (c clientInternalMeta) GetUUID() uuid.UUID {
	return c.eId
}
