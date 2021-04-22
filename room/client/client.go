package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"

	"github.zhaw.ch/raycasters/server/protocol/server"
	"github.zhaw.ch/raycasters/server/room/helpers"
	"github.zhaw.ch/raycasters/server/utils"

	"github.com/gorilla/websocket"
	"github.zhaw.ch/raycasters/server/log"
)

const (
	// MaxClientCommand is the prepared length of the opcode data slice
	// Attention this value has to be changed once the protocol changes.
	MaxClientCommand    = 17
	defaultMessageType  = websocket.BinaryMessage
	shutdownMessageType = websocket.CloseMessage
)

// logger is responsible for logging to the system logger.
var logger = log.Log()

var (
	// ErrClientWorkersAlreadyActive is the error that as the name implies
	// tries to show to the caller that the respective client already
	// has a running send and receiver pump thread.
	ErrClientWorkersAlreadyActive = errors.New("the client workers are already active")

	// ErrNoBytesRead is the error returned in case that the received
	// byte count it zero.
	ErrNoBytesRead = errors.New("empty packge received")

	// ErrClosedConnection is the error returned in case of
	// that the connections is closed by the client.
	ErrClosedConnection = errors.New("the client closed the connection")
)

// ErrNonBinaryMessage is the error returned when a non binary error happens
type ErrNonBinaryMessage struct {
	id int
}

// Error satisies the error interface
func (err ErrNonBinaryMessage) Error() string {
	return fmt.Sprintf("Non binary message type was received type : [%d]", err.id)
}

// Is satisfie the errors.Is definition
func (err ErrNonBinaryMessage) Is(target error) bool {
	_, ok := target.(ErrNonBinaryMessage)
	return ok
}

// Client represents a client in it's full configuration.
// This struct stores all the information about a single
// client instance in a hub.
type Client struct {
	meta     clientInternalMeta
	queue    helpers.ClientQueue
	status   ConnectionStatus
	isActive utils.AtomicBool
	wg       sync.WaitGroup
	lock     sync.RWMutex
}

// NewClient creates and sets up a new client from it's meta information.
func NewClient(id uint8, meta ClientMeta, queue helpers.ClientQueue) *Client {
	c := Client{
		queue:    queue,
		status:   Offline,
		meta:     newClientInternalMeta(id, meta),
		isActive: utils.NewAtomicBool(false),
		wg:       sync.WaitGroup{},
		lock:     sync.RWMutex{},
	}

	// register connection
	c.registerConn(meta, queue)

	return &c
}

// GetStatus will return the current connection status of the client
func (c *Client) GetStatus() ConnectionStatus {
	return c.status
}

// GetInternalMeta will return the internal version of the
// client meta information.
func (c *Client) GetInternalMeta() clientInternalMeta {
	return c.meta
}

// registerConn will register the new client with it's exclusive hub connection.
// Attention this function assumes that there already is a lock on the client
// object.
func (c *Client) registerConn(conn ClientMeta, queue helpers.ClientQueue) error {
	// check if the receiving routine is running
	if c.isActive.Get() {
		return ErrClientWorkersAlreadyActive
	}

	// wait until there are no working connections
	// in case that there already was a timeout
	c.wg.Wait()

	c.lock.Lock()
	defer c.lock.Unlock()

	// make sure that there is no chance of this conn
	// already being registered
	if c.status == Registered {
		return ErrClientWorkersAlreadyActive
	}

	c.meta = newClientInternalMeta(c.meta.id, conn)
	c.queue = queue
	c.status = Registered

	return nil
}

// Run starts the client send and receive pumps.
// It will return an error, if the function is
// called, while the pumps are already working.
// This function might block in case of
// reconnection, where a pump might still be
// cleaning up the buffers.
func (c *Client) Run() error {
	// check if there already is an active worker
	if !c.isActive.SetTrue() {
		return ErrClientWorkersAlreadyActive
	}

	// make sure that there is no chance of this conn
	// already being registered
	if !c.checkRegistered() {
		return ErrClientWorkersAlreadyActive
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

// checkOffline will check based on the
// current client state, if the client
// already as been set to offline.
func (c *Client) checkRegistered() bool {
	// this write lock is needed in the later parts
	// of the processes
	c.lock.RLock()
	defer c.lock.RUnlock()
	// don't bother with running the function more then
	// once
	// unlock again here
	return c.status == Registered
}

// checkOffline will check based on the
// current client state, if the client
// already as been set to offline.
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

// sendData is the actual function used to run the client send loop
// it will also handle all the outgoing data packages.
// Attention this function is a blocking call, so be careful, while
// reading form it.
// This function uses no locking as per the condition of the atomic
// condition this function will only ever be called a single time
// per run.
func (c *Client) sendData() {
	defer c.shutdownHandler()

	// No checks for activation are needed
	// due to that this function will only
	// be called from the run function
	for {
		message, err := c.queue.Dequeue()

		if err != nil {
			// the queue has shutdown => server is
			// shutting down

			// due to how the client works this error can
			// be ignored as it would imply an other
			// queue worker or drainer
			messages, _ := c.queue.DequeueDrain()
			for _, message := range messages {
				// if there was a false returned here
				// then this means the shutdown message
				// has been send
				if !c.sendDataHelper(message) {
					break
				}
			}
			return
		}

		if !c.sendDataHelper(message) {
			return
		}
	}
}

// sendDataHelper is the actuall function used to send data packages
// it checks for the type of message send and handles errors.
func (c *Client) sendDataHelper(message []byte) bool {
	// checkSendDataErr is a helper function responsible
	// for dealing with checking for any kind of errors
	// of the send data variety.
	checkErr := func(err error) bool {
		if err != nil {
			// try to register a timeout
			c.registerTimeout()

			// break out of the endless loop to a shutdown
			return false
		}
		return true
	}
	messageType := defaultMessageType
	// if there is a disconnect here
	if message[0] == server.DisconnectOpcode.Value() {
		messageType = shutdownMessageType
	}

	cw, err := c.meta.conn.NextWriter(messageType)

	if !checkErr(err) {
		return false
	}

	_, err = cw.Write(message)

	if !checkErr(err) {
		return false
	}

	// flush the message to the network
	err = cw.Close()

	return !checkErr(err) && messageType == defaultMessageType
}

// recvDataHandler will handle the received data pacakges
func (c *Client) recvDataHandler(messageType int, message io.Reader) error {

	logger.Debugf("Received message type [%d]", messageType)

	if messageType == websocket.CloseMessage {
		logger.Infof("the client with the id [%d] closed the connection.", messageType)
		return ErrClosedConnection
	} else if messageType != websocket.BinaryMessage {
		err := ErrNonBinaryMessage{
			id: messageType,
		}
		logger.Error(err.Error())
		return err
	}

	// prepare a zero sized array of size 17
	var readData [MaxClientCommand]byte
	readBytes, err := message.Read(readData[:])

	if err != nil {
		logger.Warnf("unexpected read error received.\n[%s]", err.Error())
		return err
	}

	if readBytes == 0 {
		return ErrNoBytesRead
	}

	// handle the gotten data on an other go thread
	// as to not block the next readers
	return c.queue.Enqueue(readData[:readBytes])
}

// recvData is the actual function used to run the client read loop
// it will also handle all the incoming opcode commands.
// Attention this function is a blocking call, so be careful, while
// reading form it.
func (c *Client) recvData() {
	defer c.shutdownHandler()

	for {
		messageType, message, err := c.meta.conn.NextReader()
		// first check the error message as per standard
		// as in case of an error the messageType and
		// message will be their nulled values.

		// as per definition
		// Applications must break out of the application's read loop when this method returns
		// a non-nil error value. Errors returned from this method are permanent. Once this
		// method returns a non-nil error, all subsequent calls to this method return the same
		// error
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
				logger.Infof("Connection to client [%d] was shutdown by client", c.meta.id)
			} else if websocket.IsCloseError(err, websocket.CloseAbnormalClosure) {
				logger.Warnf("Client [%d] lost connection, setting to timeout", c.meta.id)
				// if there was an unexpected close error
				// then the client has to be set to timeout
				c.registerTimeout()
			} else {
				logger.Warnf("unexpected web socket closure error received for client [%d] "+
					"- see https://rfc-editor.org/rfc/rfc6455.html#section-11.7 for more information\n[%s]", c.meta.id, err.Error())
			}
			// break out of the endless loop directly to a shutdown
			return
		}

		err = c.recvDataHandler(messageType, message)
		if err != nil {
			return
		}
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

	c.status = Timeout

	c.shutdownHandler()
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
// Attention this is a blocking call until the
// context expires.
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

	c.shutdownHandler()

	c.lock.Lock()
	// set the state of the client to shutting down
	c.status = ShuttingDown
	c.lock.Unlock()

	// make a channel on which, we can wait on.
	runs := make(chan struct{})

	// Make a waiter that will (hopefully) shutdown soon.
	go func() {
		// close the connection
		// error is irrelevant as we already send the
		// disconnect message
		_ = c.meta.conn.Close()

		c.queue.Shutdown()

		// wait until the work group reaches zero
		c.wg.Wait()

		// toggle the completed shutdown channel
		close(runs)
	}()

	select {
	case <-runs:
		logger.Debugf("Shutdown of client [%d] complete", c.meta.eId)
	case <-ctx.Done():
		logger.Debugf("forced shutdown of client [%d], due to timeout", c.meta.eId)
	}
}
