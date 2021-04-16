package client

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.zhaw.ch/raycasters/server/protocol/server"

	"github.com/google/uuid"
	"github.zhaw.ch/raycasters/server/room/helpers"
)

// ClientShutdownTime is the upper limit for a client
// to remove the connection, before forced termination.
const ClientShutdownTime = 1 * time.Second

// ClientTimeoutTime is the upper limit for a client
// to reconnect after a undesired connection loss.
const ClientTimeoutTime = 5 * time.Second

// ErrClientDoesntExist represents a client that does not actually
// in the system exist.
type ErrClientDoesntExist struct {
	meta ClientMeta
}

func NewErrClientDoesntExist(meta ClientMeta) ErrClientDoesntExist {
	return ErrClientDoesntExist{
		meta,
	}
}

// Error is the function used to fulfill the error interface.
func (err ErrClientDoesntExist) Error() string {
	return fmt.Sprintf("client doesn't exist [%d] - [%s]", err.meta.id, err.meta.name)
}

// Is satisfies the error.Is requirements, so that
// the error above can be recognized correctly.
func (err ErrClientDoesntExist) Is(target error) bool {
	_, ok := target.(ErrClientDoesntExist)
	return ok
}

// ClientManager is the struct responsible for managing
// any and all client connections.
type ClientManager struct {
	// an overview of all the clients in the room currently.
	clients map[uint8]*Client
	// maps from the external UUID to the internally used
	mapping map[uuid.UUID]uint8
	// manages all the available ids
	idManager helpers.IdManager
	// Counts the amount of connections are currently open and
	// running.
	wg *sync.WaitGroup
	// The read writer lock used to block the client list on inserting
	// or removing a client.
	// Attention: This lock is only to be used via the internal functions.
	lock sync.RWMutex
}

// NewClientManager will initialize a new clientManager instance.
func NewClientManager() ClientManager {
	return ClientManager{
		clients:   make(map[uint8]*Client),
		mapping:   make(map[uuid.UUID]uint8),
		idManager: helpers.NewIdManager(),
		lock:      sync.RWMutex{},
		wg:        &sync.WaitGroup{},
	}
}

// Size will return the current client count.
// Attention although this is a thread safe
// locking functions it is not recommended
// to be used, as it forces the usage of
// a reader lock and that will slow down
// other write processes.
func (manager *ClientManager) Size() int {
	manager.lock.RLock()
	defer manager.lock.RUnlock()
	return len(manager.clients)
}

// MapId will map the client meta information with the internal value.
// If there is no such person this function will return false and
// a zero value.
// Attention this is thread safe, but fairly slow due to the lock.
func (manager *ClientManager) MapId(meta ClientMeta) (uint8, bool) {
	manager.lock.RLock()
	defer manager.lock.RUnlock()
	return manager.mapToInternalId(meta)
}

// MapMeta will map the client internal id with the correct meta
// information.
// If there is no such person this function will return false and
// a zero value.
// Attention this is thread safe, but fairly slow due to the lock.
func (manager *ClientManager) MapMeta(id uint8) (ClientMeta, bool) {
	manager.lock.RLock()
	defer manager.lock.RUnlock()
	return manager.mapToExternalMeta(id)
}

// mapToExternalId is the lock free version to the MapMeta function,
// it is intended to be used, with external locking in place as
// it itself does not lock, but assumes the caller already
// has one.
func (manager *ClientManager) mapToExternalMeta(id uint8) (ClientMeta, bool) {
	client, ok := manager.clients[id]
	if !ok {
		return ClientMeta{}, ok
	}
	return client.meta.GetMeta(), ok
}

// mapToInternalId is the lock free version to the MapId function,
// it is intended to be used, with external locking in place as
// it itself does not lock, but assumes the caller already
// has one.
func (manager *ClientManager) mapToInternalId(meta ClientMeta) (uint8, bool) {
	id, ok := manager.mapping[meta.id]
	return id, ok
}

// GetPlayers will return the senable player configuration
// and true if it exists, else it will return an empty
// server.Player struct and false.
func (manager *ClientManager) GetPlayer(id uint8) (server.Player, bool) {
	manager.lock.RLock()
	defer manager.lock.RUnlock()
	client, ok := manager.clients[id]
	if !ok {
		return server.Player{}, ok
	}
	return client.meta.GetPlayer(), ok
}

// GetPlayers will retuen a slice of all the currently
// registered players in the server.
func (manager *ClientManager) GetPlayers() []server.Player {
	manager.lock.RLock()
	defer manager.lock.RUnlock()

	size := len(manager.clients)
	players := make([]server.Player, 0, size)
	for _, client := range manager.clients {
		players = append(players, client.meta.GetPlayer())
	}
	return players
}

// AddClient will add a new client to the manager, if the client has
// once before been added, then the previous connection will be
// replaced, as per definition a client can only reconnect,
// over a new connection. If the client UUID already has a working
// connection, an error will be returned.
//
// If there is no more space left in the client manager
// this function will panic.
func (manager *ClientManager) AddClient(meta ClientMeta, queue helpers.ClientQueue) error {

	// lock the hub, so that writing is okay
	manager.lock.Lock()
	defer manager.lock.Unlock()

	// check if client ID already exists.
	if id, ok := manager.mapToInternalId(meta); ok {
		// the person already existed in the Hub
		// add him back
		if client, ok := manager.clients[id]; ok {
			// get client
			err := client.registerConn(meta, queue)

			if err != nil {
				return err
			}

			// start the client
			client.Run()

		} else {
			// There was a programming error, never is it allowed that the client
			// is empty.
			panic("On an existing UUID mapping did a client not exist.")
		}
	} else {
		// a new client
		// get the new ID for the client
		id, err := manager.idManager.Pop()

		// If the hub is filled up
		if err != nil {
			return err
		}

		// add a new client
		manager.wg.Add(1)

		// map the UUID to the internally used
		manager.mapping[meta.id] = id

		// create the client with the correct information
		client := NewClient(id, meta, queue)
		manager.clients[id] = client

		// start the client
		client.Run()
	}
	return nil
}

// RemoveClient will remove the given client from the manager, this will return an error
// if the given client does not exist.
func (manager *ClientManager) RemoveClient(meta ClientMeta) error {
	id, ok := manager.MapId(meta)
	if !ok {
		return NewErrClientDoesntExist(meta)
	}

	manager.lock.RLock()
	val := manager.clients[id]
	manager.lock.RUnlock()

	// make sure it is shutdown, at most wait for the timeout
	// afterwards kill it.
	cxt, cancel := context.WithTimeout(context.Background(), ClientShutdownTime)

	waiter := make(chan struct{})

	// run shutdown procedure, while preparing the rest of the system
	go func() {
		val.Shutdown(cxt)
		close(waiter)
	}()

	// this is run in an internal function, as to
	// keep the lock for as short as possible
	// due to the waiter possibly waiting for about 1s
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

	// remove client from manager
	manager.wg.Done()

	// lock the system and remove rest of stuff,
	// try to keep the lock for as short as possible
	manager.lock.Lock()
	defer manager.lock.Unlock()
	// free the id as the last point
	// at the end of the process
	// so to never have a double entry
	// with the same internal id
	//
	// add it back to the queue
	err := manager.idManager.Push(id)

	if err != nil {
		// if there was a panic here, there was a programming error
		// caused by incorrect id management
		str := fmt.Sprintf("id has already been freed.\n[%s]", err.Error())
		panic(str)
	}

	return nil
}

// Shutdown is a completed Shutdown of all the client manager
// there is no way to restart a manager, as to speed up the Shutdown
// the clients will only partially be unregistered.
func (manager *ClientManager) Shutdown(cxt context.Context) {
	// Will tell all the clients to shutdown in the given shutdown time.
	childCxt, chancelCxt := context.WithTimeout(cxt, ClientShutdownTime)

	// Remove from memory and close any open waiting channels.
	waiter := make(chan struct{})
	defer func() {
		chancelCxt()
	}()

	// start up waiter first
	go func() {
		// wait until all the workers have shutdown
		manager.wg.Wait()
		close(waiter)
	}()

	closeClient := func(val *Client) {
		val.Shutdown(childCxt)
		// Decrement the wait group here
		// as there there is no unregistration.
		manager.wg.Done()
	}

	// run shutdown on all clients
	for _, value := range manager.clients {
		// run shutdown in it's own threads
		// do not remove all the clients
		// as that would use up too much time
		// for not used data.
		go closeClient(value)
	}

	select {
	case <-waiter:
		logger.Debug("client manager shutdown complete")
	case <-cxt.Done():
		logger.Info("client manager was forcefully shutdown")
	}
}
