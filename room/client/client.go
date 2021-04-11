package client

import (
	"context"
	"errors"
	"sync"

	"github.zhaw.ch/raycasters/server/gamemap"
	"github.zhaw.ch/raycasters/server/room/helpers"
	"github.zhaw.ch/raycasters/server/utils"

	"github.com/gorilla/websocket"
	"github.zhaw.ch/raycasters/server/log"
)

var logger = log.Log()

var ErrClientReaderAlreadyActive = errors.New("the client worker is already active")

// MaxClientCommand is the prepared lenght of the opcode data slice
// Attention this value has to be changed once the protocol changes.
const MaxClientCommand = 17

type Client struct {
	info     ClientGame
	meta     ClientInternalMeta
	queue    helpers.ClientQueue
	status   ConnectionStatus
	isActive utils.AtomicBool
	run      chan struct{}
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

// registerConn will register the new client with it's exclusive hub connection.
// Attention this function assumes that there already is a lock on the client
// object.
func (c *Client) registerConn(conn ClientMeta, queue helpers.ClientQueue) error {
	// check if the recv is running
	if c.isActive.Get() {
		return ErrClientReaderAlreadyActive
	}

	c.lock.Lock()
	defer c.lock.Unlock()

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

// sendData is the actuall function used to run the client send loop
// it will also handle all the outgoing data packages.
// Attention this function is a blocking call, so be carefull, while
// reading form it.
// This function uses no locking as per the condition of the atomic
// condition this function will only ever be called a single time
// per run.
func (c *Client) sendData() {
	defer c.shutdownHandler()

	checkErr := func(err error) bool {
		if err != nil {
			// try to register a timeout
			c.registerTimeout()

			// break out of the endless loop to a shutdown
			return false
		}
		return true
	}

	process := func(message []byte) bool {

		cw, err := c.meta.conn.NextWriter(websocket.BinaryMessage)

		if checkErr(err) {
			return false
		}

		_, err = cw.Write(message)

		if checkErr(err) {
			return false
		}

		// flush the message to the network
		err = cw.Close()

		return !checkErr(err)
	}
	// No checks for activation are needed
	// due to that this function will only
	// be called from the run function
	for {
		message, err := c.queue.Dequeue()

		if err != nil {
			// this means that we have to shutdown the system

			// due to how the client works this error can
			// be ignored as it would implie an other
			// drainer
			messages, _ := c.queue.DequeueDrain()
			for _, message = range messages {
				if !process(message) {
					return
				}
			}

			return
		}

		if !process(message) {
			return
		}
	}
}

// recvData is the actuall function used to run the client read loop
// it will also handle all the incomming opcode commands.
// Attention this function is a blocking call, so be carefull, while
// reading form it.
func (c *Client) recvData() {
	// TODO: refactor out of this function

	defer c.shutdownHandler()

	for {
		_, message, err := c.meta.conn.NextReader()

		// as per definition
		// Applications must break out of the application's read loop when this method returns
		// a non-nil error value. Errors returned from this method are permanent. Once this
		// method returns a non-nil error, all subsequent calls to this method return the same
		// error
		if err != nil {

			if websocket.IsCloseError(err, websocket.CloseAbnormalClosure) {
				logger.Warnf("Client [%d] lost connection, setting to timeout", c.meta.id)
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

		// prepare a zero sized array of size 17
		var readData [MaxClientCommand]byte
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
		err = c.queue.Enqueue(readData[:readBytes])
		if err != nil {
			return
		}
	}
}

// shutdownHandler will handle closing the connections
func (c *Client) shutdownHandler() {

	// only shutdown once
	if !c.isActive.SetFalse() {
		return
	}

	// Error here is not important as the connections already died
	// before reaching this position.
	_ = c.meta.conn.Close()

	// terminate the queue
	c.queue.Shutdown()
}

// Shutdown will shutdown the client and the queue.
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

	c.registerTimeout()

	c.shutdownHandler()
	c.lock.Lock()
	// set the state of the client to shutting down
	c.status = ShuttingDown
	c.lock.Unlock()

	// send the deconnect opcode
	// ignore error as it is not important
	// TODO: figure out what to do here...
	//_ = c.SendData(server.NewDisconnect())

	// make a channel on which, we can wait on.
	runs := make(chan struct{})

	// Make a waiter that will (hopefully) shutdown soon.
	go func() {

		// close the connection
		// error is irrelevant as we already send the
		// disconnect message
		c.meta.conn.Close()

		c.queue.Shutdown()

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

// registerTimeout will register the timeout in the timeout queue
func (c *Client) registerTimeout() {
	// using a full lock here as this functions will
	// be called by both the sender and the receiver.
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.status != Active {
		return
	}

	// TODO: Timer
	c.status = Timeout

	c.queue.Shutdown()
	// close the connection
	_ = c.meta.conn.Close()
	// set the activity state to false
	_ = c.isActive.SetFalse()
	// close queue
	c.queue.Shutdown()
}
