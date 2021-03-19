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
	send   chan struct{}
	recv   chan struct{}
	wg     sync.WaitGroup
	status ConnectionStatus
	info   *ClientGame
	meta   *ClientMeta
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
		c.wg.Wait()
		close(runs)
	}()

	// close the channels, that the sender and
	// reader listen to.
	close(c.send)
	close(c.recv)

	select {
	case <-runs:
		// TODO: log the a clean shutdown was a success
	case <-ctx.Done():
		// TODO: log the a forced shutdown was made
	}
}

type ClientGame struct {
	pos    Coordinate
	meta   ClientMeta
	kills  uint
	deaths uint
}

type ClientMeta struct {
	id   uint8
	name string
	conn *websocket.Conn
}
