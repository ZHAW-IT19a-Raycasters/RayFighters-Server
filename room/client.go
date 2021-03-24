package room

import (
	"context"
	"errors"
	"github.zhaw.ch/raycasters/server/protocol"
	"github.zhaw.ch/raycasters/server/protocol/client"
	"sync"

	"github.com/google/uuid"
	"github.zhaw.ch/raycasters/server/log"
)

var logger = log.Log()

type ConnectionStatus uint

const (
	Active ConnectionStatus = iota
	Timeout
	ShuttingDown
	Offline
)

func (c ConnectionStatus) String() string {
	vals := [...]string{
		"Active", "Timeout", "ShuttingDown", "Offline",
	}
	return vals[c]
}

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
	c.status = ShuttingDown
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

func (c *Client) ReadClient() () {
	for {
		_, message, err := c.meta.conn.NextReader()
		if err != nil {
			// Shutdown()
		}
		t := [17]byte{}
		readBytes, err := message.Read(t[:])
		if err != nil {
			// TODO: fail
		}
		c.handleRead(t[:readBytes])
	}
}

func (c *Client) sendData(message protocol.Bytefy) {
	cw, err := c.meta.conn.NextWriter(1)
	if err != nil {
		// return error
		_ = errors.New("something went wrong with creating NextWriter")
	} else {
		_, err = cw.Write(message.Bytes())
	}
}

func (c *Client) handleRead(t []byte) {
	switch client.Detect(t[0]) {
	case client.DisconnectOpcode:
		// Received a Disconnect request from client
		c.handleDisconnect(t)
	case client.UnknownOpcode:
		// Client received Unknown request from Server
		c.handleUnknown(t)
	case client.MovementOpcode:
		// Received a Movement request from client
		c.handleMovement(t)
	case client.ShootingOpcode:
		// Received a Shooting request from client
		c.handleShooting(t)
	case client.RequestOtherPlayersOpcode:
		// Received a RequestOtherPlayers request from client
		c.handleRequestOtherPlayer(t)
	default:
		// Received Unknown Code from client
		c.handleUnknownCode()
	}
}

func (c* Client) handleUnknownCode() {
	// c.sendData("error")
}

func (c *Client) handleDisconnect(t []byte) {
	unknown, err := client.NewUnknownByte(t)
	if err != nil {
		// c.sendData("error")
	} else {
		// TODO: SEND UNKNOWN TO FRONTEND
	}
}

func (c *Client) handleUnknown(t []byte) {
	unknown, err := client.NewUnknownByte(t)
	if err != nil {
		// c.sendData("error")
	} else {
		// TODO: SEND UNKNOWN TO FRONTEND
	}
}

func (c *Client) handleMovement(t []byte) {
	movement, err := client.NewMovementByte(t)
	if err != nil {
		// c.sendData(error)
	} else {
		// TODO: SEND MOVEMENT TO SERVER HUB
	}
}

func (c *Client) handleShooting(t []byte) {
	shooting, err := client.NewShootingBytes(t)
	if err != nil {
		// c.sendData("error")
	} else {
		// TODO: SEND SHOOTING TO SERVER HUB
	}
}

func (c *Client) handleRequestOtherPlayer(t []byte) {
	requestOtherPlayers, err := client.NewRequestOtherPlayersBytes(t)
	if err != nil {
		// c.sendData("error")
	} else {
		// TODO: SEND REQUESTOTHERPLAYERS TO FRONTEND
	}
}


