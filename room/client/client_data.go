package client

import (
	"github.com/google/uuid"
	"github.zhaw.ch/raycasters/server/gamemap"
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
	Registeredthe
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
	pos    gamemap.Coordinate
	kills  uint
	deaths uint
}

// NewClientGame creates the default client object.
func NewClientGame(pos gamemap.Coordinate) ClientGame {
	return ClientGame{
		pos:    pos,
		kills:  0,
		deaths: 0,
	}
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

// getMeta will return the meta information of the client.
func (c clientInternalMeta) getMeta() ClientMeta {
	return NewClientMeta(c.eId, c.name, c.conn)
}
