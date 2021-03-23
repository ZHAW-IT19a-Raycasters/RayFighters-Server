package room

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"github.zhaw.ch/raycasters/server/log"
)

type ConnectionStatus uint

const (
	Active ConnectionStatus = iota
	Timeout
	Shuttingdown
	Offline
)

var logger = log.Log()

type Client struct {
	// is the one channel that controlls the
	// input output connections
	running chan struct{}
	info    ClientGame
	meta    ClientInternalMeta
	status  ConnectionStatus
	wg      *sync.WaitGroup
}

func NewClient(id uint8, meta ClientMeta) *Client {
	c := Client{
		running: make(chan struct{}),
		info:    NewClientGame(NewCoordinate()),
		meta:    NewClientInternalMeta(id, meta),
		status:  Offline,
		wg:      &sync.WaitGroup{},
	}
	return &c
}

func (c *Client) registerConn(conn ClientMeta) {
	c.meta = NewClientInternalMeta(c.meta.id, conn)
	c.status = Active
	// TODO: start receiving and sending here
}

func (c *Client) Shutdown(ctx context.Context) {
	// TODO: log message of shutdown

	logger.Debugf("Shutting down client [%d]", c.meta.eId)

	// set the state of the client to shutting down
	c.status = Shuttingdown
	// set the state of the client to offline
	defer func() { c.status = Offline }()

	// make a channel on which, we can wait on.
	runs := make(chan struct{})

	// Make a waiter that will (hopefully) shutdown soon.
	go func() {
		// close the running event handler
		close(c.running)
		// wait until the workgroup reaches zero
		c.wg.Wait()
		close(runs)
	}()

	select {
	case <-runs:
		// TODO: log the a clean shutdown was a success
		logger.Debugf("Shutdown of client [%d] complete", c.meta.eId)
	case <-ctx.Done():
		// TODO: log the a forced shutdown was made
		logger.Debugf("forced shutdown of client [%d], due to timeout", c.meta.eId)
	}
}

type ClientGame struct {
	pos    Coordinate
	kills  uint
	deaths uint
}

func NewClientGame(pos Coordinate) ClientGame {
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

func NewClientInternalMeta(id uint8, client ClientMeta) ClientInternalMeta {
	return ClientInternalMeta{
		id:   id,
		eId:  client.id,
		name: client.name,
		conn: client.conn,
	}
}
