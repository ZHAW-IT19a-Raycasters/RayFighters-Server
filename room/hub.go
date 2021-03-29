package room

import (
	"context"
	"errors"
	"math"

	"github.zhaw.ch/raycasters/server/protocol"
	"github.zhaw.ch/raycasters/server/protocol/server"
	"github.zhaw.ch/raycasters/server/room/helpers"
)

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
	// The room id is used to uniquely identify the hub.
	id int
	// The manager of the different client with the different connections.
	clientManager clientManager
	// Is the player counter, this variable counts the active
	// player connections, although that number is already available
	// via the client map length, is would have to be accessed
	// from behind a lock. And so creates an unneeded overhead
	// while asynchronously adding a new client. So to avoid the given
	// overhead an atomic implementation is here used.
	counter helpers.AtomicUint32
	// The channel over which a new connection
	// is added. This connections input can be accessed
	// via the `RegisterClient` function associated
	// with this type.
	input chan ClientMeta
	// The channel over which a connections
	// are removed over. This
	output chan ClientMeta
	// The channel that signals the main loop to shutdown.
	shutdown chan struct{}
	// If the current hub is running or not.
	isActive helpers.AtomicBool
}

// NewHub will initiate a new hub struct.
// The hubId needs to be unique over all
// the data structures.
func NewHub(hubId int) *Hub {
	return &Hub{
		input:    make(chan ClientMeta),
		output:   make(chan ClientMeta),
		shutdown: make(chan struct{}),
		id:       hubId,
		counter:  helpers.NewAtomicUint32(HubClientLimit),
		isActive: helpers.NewAtomicBool(false),
	}
}

// IsAlive will check if the current hub is running or not.
func (h *Hub) IsAlive() bool {
	return h.isActive.Get()
}

// GetAllPlayers will generate a slice
// of send able player data.
func (h *Hub) GetAllPlayers() []server.Player {
	// lock the client manager, so that there is no
	// changes during this process.
	h.clientManager.lock.RLock()
	defer h.clientManager.lock.RUnlock()

	// pre-allocate all the players
	players := make([]server.Player, h.clientManager.size())

	i := 0 // add counter
	for userID, client := range h.clientManager.clients {
		players[i] = server.NewPlayer(userID, client.meta.name)
		i += 1
	}

	return players
}

// Size will return the current player count
func (h *Hub) Size() uint32 {
	return h.counter.Value()
}

// Full will check if the hub is at it's set
// limitation.
func (h *Hub) Full() bool {
	return h.counter.AtLimit()
}

// ClientExists will check if a specific client
// has an entry in the hub or not.
func (h *Hub) ClientExists(meta ClientMeta) bool {
	h.clientManager.lock.RLock()
	defer h.clientManager.lock.RUnlock()
	_, ok := h.clientManager.mapId(meta)
	return ok
}

// RegisterClient will use the meta information given
// by the parameter to check if the player
func (h *Hub) RegisterClient(meta ClientMeta) error {
	// check the meta information, so that
	// there never can be a double increment
	// per added client.
	h.clientManager.lock.RLock()
	defer h.clientManager.lock.RUnlock()

	// if the client does not have an entry in the
	// manager, then the counter may be incremented.
	if !h.ClientExists(meta) {
		// increment the counter
		// if the counter already is maxed out
		// this will return an error
		err := h.counter.Inc()
		if err != nil {
			return ErrHubIsFull
		}
	}

	h.input <- meta

	return nil
}

// UnregisterClient will give the hub the command to remove
// a given client from the system.
func (h *Hub) UnregisterClient(meta ClientMeta) error {
	// Attention, this function does not actually decrement
	// the client counter. That has to happen at the very
	// last moment, after the client has been completely removed
	// from the hub.

	// check if the client exists.
	if !h.ClientExists(meta) {
		return ErrClientDoesntExist{meta}
	}

	// client exists so give the removal command
	h.output <- meta
	return nil
}

// BroadCast will as the name implies broadcast a given
// byte interface to a given client.
func (h *Hub) BroadCast(msg protocol.Bytefy) {
	// TODO: Add broadcast function that can be called from the clients
	// TODO: figure out if this is the right path to take
	h.clientManager.broadCast(msg, []byte{})
}

// removeClient is the actuall functionality, that is
// responsible for removing a given client from the
// client manager and decrementing a client.
// Attention decrementing the client may only happen
// in a case, that there actually was a client removed.
func (h *Hub) removeClient(meta ClientMeta) {
	// TODO: Remove client from Map
	err := h.clientManager.removeClient(meta)
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
		h.counter.Dec()
	}
}

// Run is the main event loop of the hub.
// It handles all the clients added and removed
// from the hub instance.
func (h *Hub) Run() {
	// check if the hub is active
	if h.IsAlive() {
		// do not restart the hub, if it's already running
		logger.Warnf("Hub [%d] is already running", h.id)
		return
	}
	// set alive bool
	h.isActive.Set(true)
	// reset the state of is active, on shutdown or
	// crash
	defer h.isActive.Set(false)

	for {
		select {
		case <-h.shutdown:
			logger.Infof("Shutting down hub [%d]", h.id)
			return
		case meta := <-h.input:
			logger.Debugf("Adding a new player to hub [%d]", h.id)
			h.clientManager.addClient(meta)
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
		h.clientManager.shutdown(cxt)
		close(waitCompletion)
		// make sure that the main thread is closed
		close(h.shutdown)
	}()

	select {
	case <-cxt.Done():
		logger.Infof("Hub [%d] was forcefully shutdown", h.id)
	case <-waitCompletion:
		logger.Infof("Hub [%d] was successfully shutdown", h.id)
	}
}
