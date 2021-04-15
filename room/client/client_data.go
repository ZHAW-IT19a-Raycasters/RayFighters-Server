package client

import (
	"sync"

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

// ClientGame represents the state of the client in the game
type ClientGame struct {
	kills  uint16
	deaths uint16
	lock   *sync.RWMutex
}

// NewClientGame creates the default client object.
func NewClientGame() ClientGame {
	return ClientGame{
		kills:  0,
		deaths: 0,
		lock:   &sync.RWMutex{},
	}
}

func (cg ClientGame) GetKills() uint16 {
	cg.lock.RLock()
	defer cg.lock.RUnlock()
	return cg.kills
}

func (cg ClientGame) GetDeaths() uint16 {
	cg.lock.RLock()
	defer cg.lock.RUnlock()
	return cg.deaths
}

func (cg *ClientGame) IncKills() {
	cg.lock.Lock()
	defer cg.lock.Unlock()
	cg.kills += 1
}

func (cg *ClientGame) IncDeaths() {
	cg.lock.Lock()
	defer cg.lock.Unlock()
	cg.deaths += 1
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
		eId:  client.id,
		name: client.name,
		conn: client.conn,
	}
}

func (c clientInternalMeta) GetPlayer() server.Player {
	return server.NewPlayer(c.id, c.name)
}

// getMeta will return the meta information of the client.
func (c clientInternalMeta) GetMeta() ClientMeta {
	return NewClientMeta(c.GetUUID(), c.GetName(), c.conn)
}

func (c clientInternalMeta) GetName() string {
	return c.name
}

func (c clientInternalMeta) GetId() uint8 {
	return c.id
}

func (c clientInternalMeta) GetUUID() uuid.UUID {
	return c.eId
}
