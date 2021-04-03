package room

import (
	"context"
	"errors"
	"sync"

	"github.zhaw.ch/raycasters/server/protocol"
	"github.zhaw.ch/raycasters/server/protocol/client"
	"github.zhaw.ch/raycasters/server/protocol/server"
	"github.zhaw.ch/raycasters/server/room/helpers"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.zhaw.ch/raycasters/server/log"
)

var logger = log.Log()

var ErrClientReaderAlreadyActive = errors.New("the client worker is already active")

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

type HubClientInteraction interface {
	// Get All Players from Hub
	GetAllPlayers() []server.Player
	// bool to check if player was hit
	// server.Player to know which player
	PlayerGotHit() (bool, server.Player)

	PlayersKD() []server.PlayerKD

	UnregisterClient(meta ClientMeta) error
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

type Client struct {
	// is the one channel that controlls the
	// input output connections
	running  chan struct{}
	hub      HubClientInteraction
	info     ClientGame
	meta     ClientInternalMeta
	status   ConnectionStatus
	isActive helpers.AtomicBool
	wg       *sync.WaitGroup
}

func NewClient(id uint8, meta ClientMeta) *Client {
	c := Client{
		running:  make(chan struct{}),
		info:     NewClientGame(NewCoordinate()),
		status:   Offline,
		isActive: helpers.NewAtomicBool(false),
		wg:       &sync.WaitGroup{},
	}

	// register connection
	c.registerConn(meta)

	return &c
}

func (c *Client) registerConn(conn ClientMeta) error {
	if c.status == Active || c.status == Registered {
		return ErrClientReaderAlreadyActive
	}
	c.meta = NewClientInternalMeta(c.meta.id, conn)

	return nil
}

func (c *Client) Run() error {
	// check if there already is an active worker
	if !c.isActive.SetTrue() {
		return ErrClientReaderAlreadyActive
	}
	c.status = Active

	go c.recvData()

	return nil
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

// registerTimeout will register the timeout in the timeout queue
func (c *Client) registerTimeout() {
	// TODO: Timer
	c.status = Timeout
}

// recvData is the actuall function used to run the client read loop
// it will also handle all the incomming opcode commands.
// Attention this function is a blocking call, so be carefull, while
// reading form it.
func (c *Client) recvData() {
	defer func() {
		// Error here is not important as the connections already died
		// before reaching this position.
		_ = c.meta.conn.Close()
		// set the isActive variable to false
		// result is not important as here the status
		// of the connection has to be per definition
		// false.
		_ = c.isActive.SetFalse()
	}()

readLoop:
	for {
		_, message, err := c.meta.conn.NextReader()

		if err != nil {
			// as per definition
			// Applications must break out of the application's read loop when this method returns
			// a non-nil error value. Errors returned from this method are permanent. Once this
			// method returns a non-nil error, all subsequent calls to this method return the same
			// error
			if err != nil {
				// if there was an unexpected close error
				// then the client has to be set to timeout
				if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
					c.registerTimeout()
					logger.Debugf("Client [%d] lost connection, setting to timeout", c.meta.id)
				} else if websocket.IsCloseError(err,
					websocket.CloseProtocolError, websocket.CloseNoStatusReceived,
					websocket.CloseInvalidFramePayloadData,
				) {
					c.registerTimeout()
					logger.Warnf("unexpected websocket clusure error received for client [%d] "+
						"- see https://rfc-editor.org/rfc/rfc6455.html#section-11.7 for more information\n[%s]", c.meta.id, err.Error())
				} else {
					// error is unknown so unregister the client
					c.status = ShuttingDown

					c.hub.UnregisterClient(c.meta.getMeta())

					logger.Warnf("unexpected error happend\n%s", err.Error())
				}

				break readLoop
			}
		}
		readData := [17]byte{}
		readBytes, err := message.Read(readData[:])

		if err != nil {
			logger.Warnf("unexpected read error received.\n[%s]", err.Error())
		}

		// handle the gotten data on an other go thread
		// as to not block the next readers
		go c.handleRead(readData[:readBytes])
	}
}

func (c *Client) SendData(message protocol.Bytefy) error {
	if !c.isActive.Get() {
		return ErrClientReaderAlreadyActive
	}
	cw, err := c.meta.conn.NextWriter(1)
	if err != nil {
		// return error
		_ = errors.New("something went wrong with creating NextWriter")
	} else {
		_, _ = cw.Write(message.Bytes())
	}
	return nil
}

func (c *Client) handleRead(t []byte) {
	switch client.Detect(t[0]) {
	case client.DisconnectOpcode:
		// Received a Disconnect request from client
		c.handleDisconnect(t)
	case client.UnknownOpcode:
		// Client received Unknown request from Server
		c.handleUnknownSent(t)
	case client.WhoAmIOpcode:
		// Received a WhoAmI request from client
		c.handleWhoAmI(t)
	case client.MovementOpcode:
		// Received a Movement request from client
		c.handleMovement(t)
	case client.ShootingOpcode:
		// Received a Shooting request from client
		c.handleShooting(t)
	case client.RequestOtherPlayersOpcode:
		// Received a RequestOtherPlayers request from client
		c.handleRequestOtherPlayer(t)
	case client.RequestPlayerKDOpcode:
		// Received a RequestOtherPlayers request from client
		c.handleRequestOtherPlayer(t)
	default:
		// Received Unknown Code from client
		c.handleUnknownReceived(t)
	}
}

func (c *Client) handleUnknownReceived(t []byte) {
	_, err := client.NewUnknownByte(t)
	if err != nil {
		logger.Info(err)
		newUnknown := server.NewUnknown()
		c.SendData(newUnknown)
	} else {
		newUnknown := server.NewUnknown()
		c.SendData(newUnknown)
	}
}

func (c *Client) handleDisconnect(t []byte) {
	_, err := client.NewDisconnectByte(t)
	if err != nil {
		logger.Info(err)
		newUnknown := server.NewUnknown()
		c.SendData(newUnknown)
	} else {
		newDisconnect := server.NewDisconnect()
		c.SendData(newDisconnect)
	}
}

func (c *Client) handleUnknownSent(t []byte) {
	logger.Error("Unknown Opcode was sent")
}

func (c *Client) handleWhoAmI(t []byte) {
	_, err := client.NewWhoAmIByte(t)
	if err != nil {
		logger.Info(err)
		newUnknown := server.NewUnknown()
		c.SendData(newUnknown)
	} else {
		newYouAreCalled := server.NewYouAreCalled(c.meta.id)
		c.SendData(newYouAreCalled)
	}
}

func (c *Client) handleMovement(t []byte) {
	movement, err := client.NewMovementByte(t)
	if err != nil {
		logger.Info(err)
		newUnknown := server.NewUnknown()
		c.SendData(newUnknown)
	} else {
		newMovementPlayer := server.NewMovementPlayer(c.meta.id, movement.CoorX(), movement.CoorY())
		c.SendData(newMovementPlayer)
	}
}

func (c *Client) handleShooting(t []byte) {
	_, err := client.NewShootingBytes(t)
	if err != nil {
		logger.Info(err)
		newUnknown := server.NewUnknown()
		c.SendData(newUnknown)
	} else {
		hit, player := c.hub.PlayerGotHit()
		if hit {
			newAnotherPlayerKilled := server.NewPlayerDeath(player.UserId(), c.meta.id)
			c.SendData(newAnotherPlayerKilled)
		}
	}
}

func (c *Client) handleRequestOtherPlayer(t []byte) {
	_, err := client.NewRequestOtherPlayersBytes(t)
	if err != nil {
		logger.Info(err)
		newUnknown := server.NewUnknown()
		c.SendData(newUnknown)
	} else {
		newPlayerInfo := server.NewSendAllPlayerInfo(c.hub.GetAllPlayers())
		c.SendData(newPlayerInfo)
	}
}

func (c *Client) handleRequestPlayerKD(t []byte) {
	_, err := client.NewRequestPlayerKDByte(t)
	if err != nil {
		logger.Info(err)
		newUnknown := server.NewUnknown()
		c.SendData(newUnknown)
	} else {
		newPlayersKD := server.NewPlayersKD(c.hub.PlayersKD())
		c.SendData(newPlayersKD)
	}
}
