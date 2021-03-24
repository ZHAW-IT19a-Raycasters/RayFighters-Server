package room

import (
	"context"
	"errors"
	"time"

	"github.zhaw.ch/raycasters/server/protocol/server"
)

const ClientShudownTime = 1 * time.Second
const ClientTimeoutTime = 5 * time.Second

// HubBufferLimit is the amount of elements, that may be inside the
// client register channel.
const HubBufferLimit = 1

// ErrHubIsFull is the error that is returned on an RegisterClient error.
var ErrHubIsFull = errors.New("the hub is currently filled to capacity")

// Hub represents a single room.
type Hub struct {
	// The channel over which a new connection
	// comes into.
	conn chan ClientMeta
	// The room id
	id int
	// The manager of the different client with the different connections.
	clientManager clientManager
	// The channel that signals the main loop to shutdown.
	running chan struct{}
}

func NewHub(hubId int) (*Hub, chan<- ClientMeta) {
	hub := Hub{
		conn:    make(chan ClientMeta, HubBufferLimit),
		id:      hubId,
		running: make(chan struct{}),
	}
	return &hub, hub.conn
}

func (h *Hub) Size() int {
	return h.clientManager.size()
}

func (h *Hub) Full() bool {
	return h.clientManager.idManager.IsFull()
}

func (h *Hub) RegisterClient(meta ClientMeta) error {
	return h.clientManager.addClient(meta)
}

// TODO: Add broadcast function that can be called from the clients
func (h *Hub) BroadCast(msg server.ServerOpcode) {
	h.clientManager.broadCast(msg, []byte{})
}

// Run is the main event loop of the hub.
// TODO: figure out if this approach makes sense
// func (h *Hub) Run() {
// 	for {
// 		select {
// 		case <-h.running:
// 			logger.Infof("Shutting down hub [%d]", h.id)
// 			return
// 		case data := <-h.conn:
// 			logger.Debug("Adding a new player to hub [%d]", h.id)
// 			h.clientManager.addClient(data)
// 		}
// 	}
// }

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
