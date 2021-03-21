package room

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.zhaw.ch/raycasters/server/room/helpers"
)

const ClientShudownTime = 1 * time.Second

type clientManager struct {
	// an overview of all the clients in the room currently.
	clients map[uint8]*Client
	// mapps from the external uuid to the internally used
	mapping map[uuid.UUID]uint8
	// manages all the available ids
	idManager helpers.IdManager
	// Counts the ammount of connections are currently open and
	// running.
	wg *sync.WaitGroup
	// The read writer lock used to block the client list on inserting
	// or removing a client.
	// Attention: This lock is only to be used via the internal functions.
	lock sync.RWMutex
}

func newClientManager() clientManager {

	return clientManager{
		clients:   make(map[uint8]*Client),
		idManager: helpers.NewIdManager(),
		wg:        &sync.WaitGroup{},
	}
}

// mapId will map the client meta information with the internal value.
// If there is no such person this function will return false and
// a zero value.
func (manager *clientManager) mapId(meta ClientMeta) (uint8, bool) {
	id, ok := manager.mapping[meta.id]
	return id, ok
}

// addClient will add a new client to the manager, if the client has
// once before been added, then the previous connection will be
// replaced.
func (manager *clientManager) addClient(meta ClientMeta) {
	// lock the hub, so that writing is okay
	manager.lock.Lock()
	defer manager.lock.Unlock()
	// check if client ID already exists.
	if id, ok := manager.mapId(meta); ok {
		// the person already exisited in the Hub
		// add him back
		if val, ok := manager.clients[id]; ok {
			// get client
			val.registerConn(meta)
		} else {
			// There was a programming error, never is it allowed that the client
			// is empty.
			panic("On an existing UUID mapping did a client not exist.")
		}
	} else {
		// a new client
		// get the new ID for the client
		id, _ := manager.idManager.Pop()

		// map the UUID to the internaly used
		manager.mapping[meta.id] = id

		// create the client with the correct information
		client := NewClient(id, meta)
		manager.clients[id] = client
	}
}

// removeClient will remove the given client from the manager, this will return an error
// if the given client does not exist.
func (manager *clientManager) removeClient(meta ClientMeta) {
	manager.lock.Lock()
	defer manager.lock.Unlock()
	id, ok := manager.mapId(meta)
	if !ok {
		// TODO: return a fitting error
		return
	}
	// make sure it is shutdown
	val := manager.clients[id]
	cxt, cancel := context.WithTimeout(context.Background(), ClientShudownTime)
	val.Shutdown(cxt)
	// wait until done
	<-cxt.Done()

	cancel()

	// remove from mapping
	delete(manager.mapping, meta.id)
	delete(manager.clients, id)
}

func (manager *clientManager) shutdown(cxt context.Context) {
	// Will tell all the clients to shutdown in the given shutdown time.
	childCxt, chancelCxt := context.WithTimeout(cxt, ClientShudownTime)

	// Remove from memory and close any open waiting channels.
	defer func() {
		chancelCxt()
	}()

	for _, key := range manager.clients {
		key.Shutdown(childCxt)
	}

	manager.wg.Wait()
}

// Hub represents a single room.
type Hub struct {
	// The channel overwhich a new connection
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
		conn: make(chan ClientMeta),
		id:   hubId,

		running: make(chan struct{}),
	}

	return &hub, hub.conn
}

// TODO: Add broadcast function that can be called from the clients
// func (h *Hub) broadCast(msg )

// Run is the main event loop of the hub.
// TODO: figure out if this approach makes sense
func (h *Hub) Run() {
Main:
	for {
		select {
		case <-h.running:
			break Main
		case data := <-h.conn:
			h.clientManager.addClient(data)
		}
	}
	// TODO: Shuttingdown log
}

// Shutdown will intialize a complete shutdown of any and all clients.
// With the cxt context a premature cancelation of all client
// connections can be started.
func (h *Hub) Shutdown(cxt context.Context) {
	waitCompletion := make(chan struct{})

	go func() {
		h.clientManager.shutdown(cxt)
		close(waitCompletion)
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
