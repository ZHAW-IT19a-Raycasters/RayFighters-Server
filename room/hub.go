package room

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.zhaw.ch/raycasters/server/room/helpers"
)

const ClientShudownTime = 1 * time.Second
const ClientTimeoutTime = 5 * time.Second

type ErrClientDoesntExist struct {
	meta ClientMeta
}

func (err ErrClientDoesntExist) Error() string {
	return fmt.Sprintf("client doesn't exist [%d] - [%s]", err.meta.id, err.meta.name)
}

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
		mapping:   make(map[uuid.UUID]uint8),
		idManager: helpers.NewIdManager(),
		wg:        &sync.WaitGroup{},
	}
}

func (manager *clientManager) size() int {
	return len(manager.clients)
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
// This will return an error, if there is no more space left in the given
// room. Attention this will not return an error if the given client
// already has a profile in the room.
func (manager *clientManager) addClient(meta ClientMeta) error {
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
		id, err := manager.idManager.Pop()

		if err != nil {
			return err
		}

		// map the UUID to the internaly used
		manager.mapping[meta.id] = id

		// create the client with the correct information
		client := NewClient(id, meta)
		manager.clients[id] = client
	}
	return nil
}

// removeClient will remove the given client from the manager, this will return an error
// if the given client does not exist.
func (manager *clientManager) removeClient(meta ClientMeta) error {
	id, ok := manager.mapId(meta)
	if !ok {
		// TODO: return a fitting error
		return ErrClientDoesntExist{meta}
	}
	val := manager.clients[id]

	// make sure it is shutdown, at most wait for the timout
	// afterwards kill it.
	cxt, cancel := context.WithTimeout(context.Background(), ClientShudownTime)

	waiter := make(chan struct{})

	// run shutdown procedure, while preparing the rest of the system
	go func() {
		val.Shutdown(cxt)
		close(waiter)
	}()

	// this is run in an internal function, as to
	// keep the lock for as short as possible
	remove := func() {
		// lock the system and remove stuff,
		// try to keep the lock for as short as possible
		manager.lock.Lock()
		defer manager.lock.Unlock()
		// remove from mapping
		delete(manager.mapping, meta.id)
		delete(manager.clients, id)
	}

	remove()

	<-waiter
	cancel()

	// lock the system and remove stuff,
	// try to keep the lock for as short as possible
	manager.lock.Lock()
	defer manager.lock.Unlock()
	// free the id as the last point, so that only at the end of the process
	//
	// add it back to the queue
	err := manager.idManager.Push(id)

	if err != nil {
		str := fmt.Sprintf(
			"if there was a panic here, there was a programming error"+
				"caused by incorrect id management.\n[%s]",
			err.Error(),
		)
		panic(str)
	}

	return nil
}

func (manager *clientManager) shutdown(cxt context.Context) {
	// Will tell all the clients to shutdown in the given shutdown time.
	childCxt, chancelCxt := context.WithTimeout(cxt, ClientShudownTime)

	// Remove from memory and close any open waiting channels.
	waiter := make(chan struct{})
	defer func() {
		chancelCxt()
	}()

	// start up waiter first
	go func() {
		manager.wg.Wait()
		close(waiter)
	}()

	// run shutdown on all clients
	for _, key := range manager.clients {
		// run shutdown in it's own threads
		// do not remove all the clients
		// as that would use up too much time
		// for not used data.
		go key.Shutdown(childCxt)
	}

	select {
	case <-waiter:
		logger.Debug("client manager shutdown complete")
	case <-cxt.Done():
		logger.Info("client manager was forcefully shutdown")
	}
}

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
	h.clientManager.lock.Lock()
	defer h.clientManager.lock.Unlock()
	// check that adding a client is safe
	if h.Full() {
		return ErrHubIsFull
	}

	// send meta to the hub
	h.conn <- meta

	return nil
}

// TODO: Add broadcast function that can be called from the clients
// func (h *Hub) broadCast(msg )

// Run is the main event loop of the hub.
// TODO: figure out if this approach makes sense
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
