package room

import (
	"context"
	"errors"
	"math"

	"github.zhaw.ch/raycasters/server/log"
	"github.zhaw.ch/raycasters/server/room/client"
	"github.zhaw.ch/raycasters/server/utils"
)

var logger = log.Log()

// HubClientLimit is the absolute upper limit supported
// by a single hub. This limitation comes from the
// decision to only support a player id of size uint8.
const HubClientLimit = math.MaxUint8

// ErrHubIsFull is the error that is returned on an RegisterClient error.
var ErrHubIsFull = errors.New("the hub is currently filled to capacity")

// Hub represents a single room.
// The hub controls the the individual connections between the
// different modules like the map or the client manager.
type Hub struct {
	// id is the designated hub identifier
	id int
	// controller is the struct responsible for running all the client opcodes
	// and handling all the io.
	controller *hubController
	// Is the player counter, this variable counts the active
	// player connections, although that number is already available
	// via the client map length, is would have to be accessed
	// from behind a lock. And so creates an unneeded overhead
	// while asynchronously adding a new client. So to avoid the given
	// overhead an atomic implementation is here used.
	counter utils.Semaphore
	// The channel over which a new connection
	// is added. This connections input can be accessed
	// via the `RegisterClient` function associated
	// with this type.
	input chan client.ClientMeta
	// The channel over which a connections
	// are removed over. This
	output chan client.ClientMeta
	// The channel that signals the main loop to shutdown.
	shutdown chan struct{}
	// If the current hub is running or not.
	isActive utils.AtomicBool
}

// NewHub will initiate a new hub struct.
// The hubId needs to be unique over all
// the data structures.
func NewHub(hubId int) *Hub {
	h := &Hub{
		input:      make(chan client.ClientMeta),
		output:     make(chan client.ClientMeta),
		controller: newHubController(),
		shutdown:   make(chan struct{}),
		id:         hubId,
		counter:    utils.NewSemaphore(HubClientLimit),
		isActive:   utils.NewAtomicBool(false),
	}
	return h
}

// IsAlive will check if the current hub is running or not.
func (h *Hub) IsAlive() bool {
	return h.isActive.Get()
}

// Size will return the current player count
func (h *Hub) Size() int {
	return h.counter.Value()
}

// Full will check if the hub is at it's set
// limitation.
func (h *Hub) Full() bool {
	return h.counter.Full()
}

// RegisterClient will use the meta information given
// by the parameter to check if the player
func (h *Hub) RegisterClient(meta client.ClientMeta) error {

	// if the client does not have an entry in the
	// manager, then the player count has to be
	// incremented.
	if !h.controller.clientExists(meta) {
		// increment the counter
		// if the counter already is maxed out
		// this will return an error
		err := h.counter.Acquire()

		if err != nil {
			return ErrHubIsFull
		}
	}

	// add the client to the queue to be added into the system
	h.input <- meta

	return nil
}

// UnregisterClient will give the hub the command to remove
// a given client from the system.
func (h *Hub) UnregisterClient(meta client.ClientMeta) error {
	// Attention, this function does not actually decrement
	// the client counter. That has to happen at the very
	// last moment, after the client has been completely removed
	// from the hub.

	// check if the client exists.
	if !h.controller.clientExists(meta) {
		return client.NewErrClientDoesntExist(meta)
	}

	// client exists so give the removal command
	h.output <- meta
	return nil
}

// removeClient is the actuall functionality, that is
// responsible for removing a given client from the
// client manager and decrementing a client.
// Attention decrementing the client may only happen
// in a case, that there actually was a client removed.
func (h *Hub) removeClient(meta client.ClientMeta) {
	// TODO: Remove client from Map
	err := h.controller.removeClient(meta)
	// if there was an error here there was a
	// double remove called, in the time it took to
	// remove the specific client. This can only be
	// in case of hight load and bad programming.
	if err == nil {
		// error here is impossible given that
		// per client added to the hub, and it
		// is previously checked that the client
		// is only removed once, so the counter
		// is ever ony decremented once.
		_ = h.counter.Release()
	}
}

// Run is the main event loop of the hub.
// It handles all the clients added and removed
// from the hub instance.
// Attention this is a blocking function, so
// run this in a gorutine.
func (h *Hub) Run() {
	// check if the hub is active
	if !h.isActive.SetTrue() {
		// do not restart the hub, if it's already running
		logger.Warnf("Hub [%d] is already running", h.id)
		return
	}
	// set alive bool
	// reset the state of is active, on shutdown or
	// crash
	defer h.isActive.SetFalse()

	// TODO: move into hub controller
	for {
		select {
		case <-h.shutdown:
			logger.Infof("Shutting down hub [%d]", h.id)
			return
		case <-h.input:
			logger.Debugf("Adding a new player to hub [%d]", h.id)

			//h.clientManager.addClient(meta)
		case meta := <-h.output:
			logger.Debugf("Removing an old player from the hub [%d]", h.id)
			h.removeClient(meta)
		}
	}
}

// Shutdown will initialize a complete shutdown of any and all clients.
// With the cxt context a premature cancelation of all client
// connections can be started.
func (h *Hub) Shutdown(cxt context.Context) {
	waitCompletion := make(chan struct{})

	go func() {
		h.controller.shutdown(cxt)
		// make sure that the main thread is closed
		close(h.shutdown)
		// close blocker
		close(waitCompletion)
	}()

	select {
	case <-cxt.Done():
		logger.Infof("Hub [%d] was forcefully shutdown", h.id)
	case <-waitCompletion:
		logger.Infof("Hub [%d] was successfully shutdown", h.id)
	}
}
