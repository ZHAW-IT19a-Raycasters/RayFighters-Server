package room

import (
	"context"
	"errors"
	"math"

	"github.zhaw.ch/raycasters/server/protocol/server"
	"github.zhaw.ch/raycasters/server/room/helpers"
)

const HubClientLimit = math.MaxUint8

// ErrHubIsFull is the error that is returned on an RegisterClient error.
var ErrHubIsFull = errors.New("the hub is currently filled to capacity")

// Hub represents a single room.
type Hub struct {
	// The channel over which a new connection
	// comes into.
	conn chan ClientMeta
	// The room id is used to uniquely id the room
	id int
	// The manager of the different client with the different connections.
	clientManager clientManager
	// The channel that signals the main loop to shutdown.
	running chan struct{}
	// Is the player counter, this variable counts the active
	// player connections, although that number is already availble
	// via the client map length, is would have to be accessed
	// from behind a lock. So to avoid the given overhead an
	// atomic implementation is used here.
	counter helpers.AtomicUint32
}

func NewHub(hubId int) *Hub {
	return &Hub{
		conn:    make(chan ClientMeta),
		id:      hubId,
		clientManager: newClientManager(),
		running: make(chan struct{}),
		counter: helpers.NewAtomicUint32(HubClientLimit),
	}
}

func (h *Hub) Size() uint32 {
	return h.counter.Value()
}

func (h *Hub) Full() bool {
	return h.counter.AtLimit()
}

func (h *Hub) RegisterClient(meta ClientMeta) error {
	// increment the counter
	err := h.counter.Inc()

	if err != nil {
		return err
	}

	h.conn <- meta

	return nil
}

func (h *Hub) UnregisterClient(meta ClientMeta) error {

	err := h.counter.Dec()

	if err != nil {
		return err
	}

	return nil
}

// TODO: Add broadcast function that can be called from the clients
// TODO: figure out if this is the right path to take
func (h *Hub) BroadCast(msg server.ServerOpcode) {
	h.clientManager.broadCast(msg, []byte{})
}

// Run is the main event loop of the hub.
func (h *Hub) Run() {
	for {
		select {
		case <-h.running:
			logger.Infof("Shutting down hub [%d]", h.id)
			return
		case data := <-h.conn:
			logger.Debug("Adding a new player to hub [%d]", h.id)
			h.clientManager.addClient(data)
		}
	}
}

// Shutdown will intialize a complete shutdown of any and all clients.
// With the cxt context a premature cancelation of all client
// connections can be started.
func (h *Hub) Shutdown(cxt context.Context) {
	waitCompletion := make(chan struct{})

	go func() {
		h.clientManager.shutdown(cxt)
		close(waitCompletion)
		// make sure that the main thread is closed
		close(h.running)
	}()

	select {
	case <-cxt.Done():
		logger.Infof("Hub [%d] was forcefully shutdown", h.id)
	case <-waitCompletion:
		logger.Infof("Hub [%d] was successfully shutdown", h.id)
	}
}
