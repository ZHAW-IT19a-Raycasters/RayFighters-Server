package room

import (
	"context"
	"sync"
	"time"
)

const ClientShudownTime = 1 * time.Second

// Hub represents a single room.
type Hub struct {
	// The channel overwhich a new connection
	// comes into.
	conn chan ClientMeta
	// The room id
	id int
	// an overview of all the clients in the room currently.
	clients map[uint8]*Client
	// Counts the ammount of connections are currently open and
	// running.
	wg *sync.WaitGroup
	// The channel that signals the main loop to shutdown.
	running chan struct{}
	// The read writer lock used to block the client list on inserting
	// or removing a client.
	// Attention: This lock is only to be used via the internal functions.
	lock sync.RWMutex
}

func NewHub(hubId int) (*Hub, chan<- ClientMeta) {
	hub := Hub{
		conn:    make(chan ClientMeta),
		id:      hubId,
		clients: make(map[uint8]*Client),
		wg:      &sync.WaitGroup{},
		running: make(chan struct{}),
	}

	return &hub, hub.conn
}

func (h *Hub) addClient(meta ClientMeta) {
	// TODO: add the new client to the Hub
	// lock the hub, so that writing is okay
	h.lock.Lock()
	defer h.lock.Unlock()
	// check if client ID already exists.
	if val, ok := h.clients[meta.id]; ok {
		// the person already exisited in the Hub
		// add him back
		val.registerConn(meta)
	} else {
		// a new client
		client := NewClient(meta)
		h.clients[meta.id] = client
	}
}

// TODO: Add broadcast function that can be called from the clients
// func (h *Hub) broadCast(msg )

func (h *Hub) Run() {
Main:
	for {
		select {
		case <-h.running:
			break Main
		case data := <-h.conn:
			h.addClient(data)
		}
	}
	// TODO: Shuttingdown log
}

// Shutdown will intialize a complete shutdown of any and all clients.
// With the cxt context a premature cancelation of all client
// connections can be started.
func (h *Hub) Shutdown(cxt context.Context) {
	waitCompletion := make(chan struct{})
	// Will tell all the clients to shutdown in the given shutdown time.
	go func() {
		childCxt, chancelCxt := context.WithTimeout(cxt, ClientShudownTime)

		// Remove from memory and close any open waiting channels.
		defer func() {
			chancelCxt()
			close(waitCompletion)
		}()

		for _, key := range h.clients {
			key.Shutdown(childCxt)
		}

		// Will wait in here, so that the main room can be shutdown.
		h.wg.Wait()
		close(h.running)
	}()

	select {
	case <-cxt.Done():
		// TODO: add angry logging messages that
		// a forced shutdown was made.

		// make sure that the main thread is closed
		close(h.running)
	case <-waitCompletion:
		// TODO: add relaxed logging messages
	}
}
