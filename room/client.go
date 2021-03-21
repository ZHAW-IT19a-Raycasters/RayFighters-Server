package room

import (
	"context"
	"sync"

	"github.com/gorilla/websocket"
)

type ConnectionStatus uint

const (
	Active ConnectionStatus = iota
	Timeout
	Shuttingdown
	Offline
)

type Client struct {
	// is the one channel that controlls the
	// input output connections
	running chan struct{}
	info    ClientGame
	meta    ClientMeta
	status  ConnectionStatus
	wg      *sync.WaitGroup
}

func NewClient(meta ClientMeta) *Client {
	c := Client{
		running: make(chan struct{}),
		info:    NewClientGame(NewCoordinate()),
		status:  Offline,
		wg:      &sync.WaitGroup{},
	}
	return &c
}

func (c *Client) registerConn(conn ClientMeta) {
	c.meta = conn
	c.status = Active
	// TODO: start receiving and sending here
}

func (c *Client) Shutdown(ctx context.Context) {
	// TODO: log message of shutdown

	// set the state of the client to shutting down
	c.status = Shuttingdown
	// set the state of the client to offline
	defer func() { c.status = Offline }()
	runs := make(chan struct{})

	// Make a waiter that will (hopefully) shutdown soon.
	go func() {
		close(c.running)
		c.wg.Wait()
		close(runs)
	}()

	select {
	case <-runs:
		// TODO: log the a clean shutdown was a success
	case <-ctx.Done():
		// TODO: log the a forced shutdown was made
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
	id   uint8
	name string
	conn *websocket.Conn
}
