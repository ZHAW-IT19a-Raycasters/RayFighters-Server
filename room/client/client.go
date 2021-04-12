package client

import (
	"context"
	"errors"
	"sync"

	"github.zhaw.ch/raycasters/server/gamemap"
	"github.zhaw.ch/raycasters/server/room/helpers"
	"github.zhaw.ch/raycasters/server/utils"

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

type ClientGame struct {
	pos    gamemap.Coordinate
	kills  uint16
	deaths uint16
}

func NewClientGame(pos gamemap.Coordinate) ClientGame {
	return ClientGame{
		pos:    pos,
		kills:  0,
		deaths: 0,
	}
}

func (cg ClientGame) GetPosition() gamemap.Coordinate {
	return cg.pos
}

func (cg ClientGame) GetKills() uint16 {
	return cg.kills
}

func (cg ClientGame) GetDeaths() uint16 {
	return cg.deaths
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

func (c ClientInternalMeta) GetMeta() ClientMeta {
	return NewClientMeta(c.eId, c.name, c.conn)
}

func (c ClientInternalMeta) GetName() string {
	return c.name
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
	info     ClientGame
	meta     ClientInternalMeta
	queue    helpers.ClientQueue
	status   ConnectionStatus
	isActive utils.AtomicBool
	wg       sync.WaitGroup
	lock     sync.RWMutex
}

func NewClient(id uint8, meta ClientMeta, queue helpers.ClientQueue) *Client {
	c := Client{
		info:     NewClientGame(gamemap.NewCoordinate()),
		queue:    queue,
		status:   Offline,
		isActive: utils.NewAtomicBool(false),
		wg:       sync.WaitGroup{},
		lock:     sync.RWMutex{},
	}

	// register connection
	c.registerConn(meta, queue)

	return &c
}

func (c Client) GetInternalMeta() ClientInternalMeta {
	return c.meta
}

func (c Client) GetInfo() ClientGame {
	return c.info
}

// registerConn will register the new client with it's exclusive hub connection.
// Attention this function assumes that there already is a lock on the client
// object.
func (c *Client) registerConn(conn ClientMeta, queue helpers.ClientQueue) error {
	// check if the recv is running
	if c.isActive.Get() {
		return ErrClientReaderAlreadyActive
	}

	// make sure that there is no chance of this conn
	// already beeing registered
	if c.status == Registered {
		return ErrClientReaderAlreadyActive
	}
	c.status = Registered
	c.meta = NewClientInternalMeta(c.meta.id, conn)
	c.queue = queue

	return nil
}

func (c *Client) Run() error {
	// check if there already is an active worker
	if !c.isActive.SetTrue() {
		return ErrClientReaderAlreadyActive
	}
	// setup the states needed to run the client
	// this has to run in this thread here or else
	// the counter might be wrong
	// using a function to ensure that the lock is dropped as fast
	// as possible
	c.wg.Add(2)

	setupStates := func() {
		c.lock.Lock()
		defer c.lock.Unlock()
		// setup the states
		c.status = Active
	}
	setupStates()

	go func() {
		c.recvData()
		c.wg.Done()
	}()
	go func() {
		c.sendData()
		c.wg.Done()
	}()

	return nil
}

func (c *Client) checkOffline() bool {
	// this write lock is needed in the later parts
	// of the processes
	c.lock.RLock()
	defer c.lock.RUnlock()
	// don't bother with running the function more then
	// once
	// unlock again here
	return c.status == Offline
}

func (c *Client) Shutdown(ctx context.Context) {
	if c.checkOffline() {
		return
	}

	defer func() {
		c.lock.Lock()
		c.status = Offline
		c.lock.Unlock()
	}()

	logger.Debugf("Shutting down client [%d]", c.meta.eId)

	c.lock.Lock()
	// set the state of the client to shutting down
	c.status = ShuttingDown
	c.lock.Unlock()

	// send the deconnect opcode
	// ignore error as it is not important
	//_ = c.SendData(server.NewDisconnect())

	// make a channel on which, we can wait on.
	runs := make(chan struct{})

	// Make a waiter that will (hopefully) shutdown soon.
	go func() {
		// close the connection
		// error is irrelevant as we already send the
		// disconnect message
		c.meta.conn.Close()

		// wait until the workgroup reaches zero
		c.wg.Wait()

		// toggle the completed shutdown cannel
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

func (c *Client) sendData() error {

	if !c.isActive.Get() {
		return ErrClientReaderAlreadyActive
	}

	for {
		message, err := c.queue.Dequeue()

		if err != nil {
			// this means that we have to shutdown the system
			return nil
		}
		cw, err := c.meta.conn.NextWriter(websocket.BinaryMessage)

		if err != nil {
			// TODO: create a fixed error
			// return error
			return errors.New("something went wrong with creating NextWriter")
		}
		_, _ = cw.Write(message)
	}
}

// registerTimeout will register the timeout in the timeout queue
func (c *Client) registerTimeout() {
	c.lock.Lock()
	defer c.lock.Unlock()
	// TODO: Timer
	c.status = Timeout

	c.queue.Shutdown()
	// close the connection
	_ = c.meta.conn.Close()
	// set the activity state to false
	_ = c.isActive.SetFalse()
}

// recvData is the actuall function used to run the client read loop
// it will also handle all the incomming opcode commands.
// Attention this function is a blocking call, so be carefull, while
// reading form it.
func (c *Client) recvData() {
	// This function needs no locking
	// as per the condition of the atomic
	// condition this function will only
	// ever be called a single time per
	// run
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

	for {
		_, message, err := c.meta.conn.NextReader()

		// as per definition
		// Applications must break out of the application's read loop when this method returns
		// a non-nil error value. Errors returned from this method are permanent. Once this
		// method returns a non-nil error, all subsequent calls to this method return the same
		// error
		if err != nil {

			// TODO: figure out how to handle the timeout
			if websocket.IsCloseError(err, websocket.CloseAbnormalClosure) {
				logger.Debugf("Client [%d] lost connection, setting to timeout", c.meta.id)
			} else if websocket.IsCloseError(err,
				websocket.CloseProtocolError, websocket.CloseNoStatusReceived,
				websocket.CloseInvalidFramePayloadData,
			) {
				logger.Warnf("unexpected websocket clusure error received for client [%d] "+
					"- see https://rfc-editor.org/rfc/rfc6455.html#section-11.7 for more information\n[%s]", c.meta.id, err.Error())
			} else {
				logger.Warnf("unexpected error happend\n%s", err.Error())
			}

			// if there was an unexpected close error
			// then the client has to be set to timeout
			c.registerTimeout()

			// break out of the endless loop directly to a shutdown
			return
		}

		readData := [17]byte{}
		readBytes, err := message.Read(readData[:])

		if readBytes == 0 {
			logger.Warn("empty package received")
			return
		}

		if err != nil {
			logger.Warnf("unexpected read error received.\n[%s]", err.Error())
		}

		// handle the gotten data on an other go thread
		// as to not block the next readers
		c.queue.Enqueue(readData[:])
	}
}
