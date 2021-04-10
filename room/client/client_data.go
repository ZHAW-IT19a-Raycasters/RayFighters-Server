package client

import (
	"github.com/google/uuid"
	"github.zhaw.ch/raycasters/server/gamemap"
)

type ConnectionStatus uint

const (
	Active ConnectionStatus = iota
	Registered
	Timeout
	ShuttingDown
	Offline
)

func (c ConnectionStatus) String() string {
	vals := [...]string{
		"Active", "Registered", "Timeout", "ShuttingDown", "Offline",
	}
	return vals[c]
}

type ClientGame struct {
	pos    gamemap.Coordinate
	kills  uint
	deaths uint
}

func NewClientGame(pos gamemap.Coordinate) ClientGame {
	return ClientGame{
		pos:    pos,
		kills:  0,
		deaths: 0,
	}
}

type ClientMeta struct {
	id   uuid.UUID
	name string
	conn ClientConnection
}

func NewClientMeta(id uuid.UUID, name string, conn ClientConnection) ClientMeta {
	return ClientMeta{
		id:   id,
		name: name,
		conn: conn,
	}
}

type ClientInternalMeta struct {
	id   uint8
	eId  uuid.UUID
	name string
	conn ClientConnection
}

func (c ClientInternalMeta) getMeta() ClientMeta {
	return NewClientMeta(c.eId, c.name, c.conn)
}

func NewClientInternalMeta(id uint8, client ClientMeta) ClientInternalMeta {
	return ClientInternalMeta{
		id:   id,
		eId:  client.id,
		name: client.name,
		conn: client.conn,
	}
}
