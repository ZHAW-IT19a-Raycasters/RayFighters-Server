package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"

	"github.zhaw.ch/raycasters/server/protocol/server"
	"github.zhaw.ch/raycasters/server/utils"

	"github.com/gorilla/websocket"
	"github.zhaw.ch/raycasters/server/log"
)

const (
	// MaxClientCommand is the prepared length of the opcode data slice
	// Attention this value has to be changed once the protocol changes.
	MaxClientCommand    int = 25
	defaultMessageType  int = websocket.BinaryMessage
	shutdownMessageType int = websocket.CloseMessage
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

// ErrTooLongMessage is the error returned when the s
type ErrTooLongMessage struct {
	id   uint8
	size int
}

// Error satisies the error interface
func (err ErrTooLongMessage) Error() string {
	return fmt.Sprintf("Non binary message type was received type : [%d]", err.id)
}

// Is satisfie the errors.Is definition
func (err ErrTooLongMessage) Is(target error) bool {
	_, ok := target.(ErrTooLongMessage)
	return ok
}

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

type ErrRecvHandler struct {
	inner error
}

func (err ErrRecvHandler) Error() string {
	return fmt.Sprintf("got a receiver error [%s]", err.inner.Error())
}

func (err ErrRecvHandler) Is(target error) bool {
	_, ok := target.(ErrRecvHandler)
	return ok
}

func (err ErrRecvHandler) Unwrap() error {
	return err.inner
}

type ErrRecvReader struct {
	inner error
}

func (err ErrRecvReader) Error() string {
	return fmt.Sprintf("got a receiver error [%s]", err.inner.Error())
}

func (err ErrRecvReader) Is(target error) bool {
	_, ok := target.(ErrRecvReader)
	return ok
}

func (err ErrRecvReader) Unwrap() error {
	return err.inner
}

// Client represents a client in it's full configuration.
// This struct stores all the information about a single
// client instance in a hub.
type Client struct {
	meta        clientInternalMeta
	queue       ClientQueue
	timoutQueue ClientTimer
	status      ConnectionStatus
	isActive    utils.AtomicBool
	wg          sync.WaitGroup
	lock        sync.RWMutex
}

// NewClient creates and sets up a new client from it's meta information.
func NewClient(id uint8, meta ClientMeta, queue ClientQueue, tq ClientTimer) *Client {
	c := Client{
		queue:       queue,
		status:      Offline,
		meta:        newClientInternalMeta(id, meta),
		timoutQueue: tq,
		isActive:    utils.NewAtomicBool(false),
		wg:          sync.WaitGroup{},
		lock:        sync.RWMutex{},
	}

	// register connection
	c.registerConn(meta, queue)

	return &c
}

// GetStatus will return the current connection status of the client
func (c *Client) GetStatus() ConnectionStatus {
	c.lock.RLock()
	defer c.lock.RUnlock()
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
func (c *Client) registerConn(conn ClientMeta, queue ClientQueue) error {
	// check if the receiving routine is running
	if c.isActive.Get() {
		logger.Debugf("client [%d] is already active", c.meta.id)
		return ErrClientWorkersAlreadyActive
	}

	// wait until there are no working connections
	// in case that there already was a timeout
	c.wg.Wait()
	logger.Debugf("registering connecteion [%d]", c.meta.id)

	c.lock.Lock()
	defer c.lock.Unlock()

	// make sure that there is no chance of this conn
	// already being registered
	if c.status == Registered {
		logger.Debugf("client [%d] is already registered", c.meta.id)
		return ErrClientWorkersAlreadyActive
	}

	c.meta = newClientInternalMeta(c.meta.id, conn)
	c.queue = queue
	c.status = Registered

	logger.Debugf("client connection [%d] was successfully armed", c.meta.id)

	// rearm the queue
	c.timoutQueue.Reset()

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
	_, isRegistered := c.CheckState(Registered)

	if !isRegistered {
		return ErrClientWorkersAlreadyActive
	}

	// setup the states needed to run the client
	// this has to run in this thread here or else
	// the counter might be wrong
	// using a function to ensure that the lock is dropped as fast
	// as possible
	c.wg.Add(2)

	c.lock.Lock()
	// setup the states
	c.status = Active
	c.lock.Unlock()

	go func() {
		err := c.recvData()
		c.wg.Done()

		if err != nil && c.recvConnectionErrHandler(err) {
			c.Shutdown(context.Background())
		}
	}()

	go func() {
		c.sendData()
		c.wg.Done()
	}()

	return nil
}

// CheckState will check if the current client state is equvalent
// to the expected value.
func (c *Client) CheckState(exp ConnectionStatus) (ConnectionStatus, bool) {
	// this write lock is needed in the later parts
	// of the processes
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.unsafeCheckState(exp)
}

// unsafeCheckState is the lock less equivalent of the
// unsafe check state.
// Attention this function assumes a client lock from the caller.
func (c *Client) unsafeCheckState(exp ConnectionStatus) (ConnectionStatus, bool) {
	return c.status, c.status == exp
}

// sendData is the actual function used to run the client send loop
// it will also handle all the outgoing data packages.
// Attention this function is a blocking call, so be careful, while
// reading form it.
// This function uses no locking as per the condition of the atomic
// condition this function will only ever be called a single time
// per run.
func (c *Client) sendData() {
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

		logger.Debugf("Sending message [% X], to [%d]", message, c.meta.GetUUID())

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

	return checkErr(err) && messageType == defaultMessageType
}

// recvDataHandler will handle the received data packages
func (c *Client) recvDataHandler(messageType int, message io.Reader) error {

	logger.Debugf("Received websocket message type [%d] from [%d]", messageType, c.meta.id)

	if messageType == websocket.CloseMessage {
		logger.Infof("the client with the id [%d] closed the connection.", c.meta.id)
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
	} else if readBytes > MaxClientCommand {
		return ErrTooLongMessage{
			id:   c.meta.GetId(),
			size: readBytes,
		}
	}

	// handle the gotten data on an other go thread
	// as to not block the next readers
	return c.queue.Enqueue(readData[:readBytes])
}

// recvConnectionErrHandler will handle the different error cases
// possibly received by the server from the front end
func (c *Client) recvConnectionErrHandler(err error) bool {
	if errors.Is(err, ErrRecvHandler{}) {
		// As this is not an error that needs
		// special attention simply return here.
		return false
	}
	// get inner
	err = errors.Unwrap(err)

	if websocket.IsCloseError(err, websocket.CloseAbnormalClosure) {
		// the case that there was an unexpected close event
		// such as internet connection problems and similar.
		// Given that case the client has to be set into its
		// timeout state, so that it can be removed if no
		// new connection can be established.
		logger.Warnf("Client [%d] lost connection, setting to timeout", c.meta.id)
		c.registerTimeout()
		return false
	}

	if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
		// the case that either
		// - the connection was correctly closed by the client side web socket
		// - the connection was incorrectly closed by a closed browser instance
		//   or the browser having navigated away from the page withe the connection.
		logger.Infof("Connection to client [%d] was shutdown by client", c.meta.id)
	} else {
		// every other unexpected case
		logger.Warnf("unexpected web socket closure error received for client [%d] "+
			"- see https://rfc-editor.org/rfc/rfc6455.html#section-11.7 for more "+
			"information\n[%s]", c.meta.id, err.Error())
	}

	return true
}

// recvData is the actual function used to run the client read loop
// it will also handle all the incoming opcode commands.
// Attention this function is a blocking call, so be careful, while
// reading form it.
func (c *Client) recvData() error {
	defer c.shutdownHandler()

	for {
		messageType, message, err := c.meta.conn.NextReader()
		// first check the error message as per standard
		// as in case of an error the messageType and
		// message will be their nulled values.

		// first check if client is still active, as there is the case where
		// the sender will shutdown after getting a disconnect opcode
		if !c.isActive.Get() {
			// simply shutdown
			logger.Debugf("shutting down the receiver of the client [%d]", c.meta.GetId())
			return nil
		}

		// as per definition
		// Applications must break out of the application's read loop when this method returns
		// a non-nil error value. Errors returned from this method are permanent. Once this
		// method returns a non-nil error, all subsequent calls to this method return the same
		// error
		if err != nil {
			return ErrRecvReader{
				inner: err,
			}
		}

		err = c.recvDataHandler(messageType, message)

		if err != nil {
			return ErrRecvHandler{
				inner: err,
			}
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

	logger.Debugf("shutting down client [%d]", c.meta.id)

	// Error here is not important as the connections already died
	// before reaching this position.
	err := c.meta.conn.Close()
	if err != nil {
		logger.Debugf("error happend here [%s] for client [%d]", err.Error(), c.meta.id)
	}

	// terminate the queue
	c.queue.Shutdown()
}

// Shutdown will shutdown the client and the queue.
// Attention this is a blocking call until the
// context expires.
func (c *Client) Shutdown(ctx context.Context) {
	_, isOffline := c.CheckState(Offline)

	if isOffline {
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

	// make a channel on which, we can wait on.
	runs := make(chan struct{})

	// Make a waiter that will (hopefully) shutdown soon.
	go func() {
		logger.Debugf("running shutdown procedure for client [%d]", c.meta.id)
		c.shutdownHandler()
		// wait until the work group reaches zero
		c.wg.Wait()

		logger.Debugf("waiting for client [%d] worker shutdown completed", c.meta.id)

		// toggle the completed shutdown channel
		close(runs)
	}()

	select {
	case <-runs:
		logger.Debugf("Shutdown of client [%d] completed", c.meta.eId)
	case <-ctx.Done():
		logger.Debugf("forced shutdown of client [%d], due to timeout", c.meta.eId)
	}
}
