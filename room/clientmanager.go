package room

import (
	"container/list"
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.zhaw.ch/raycasters/server/protocol/server"
	"github.zhaw.ch/raycasters/server/room/helpers"
)

// ClientShutdownTime is the upper limit for a client
// to remove the connection, before forced termination.
const ClientShutdownTime = 1 * time.Second

// ClientTimeoutTime is the upper limit for a client
// to reconnect after a undesired connection loss.
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
	// TODO: Check atomic counter for limit

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

		// If the hub is filled up
		if err != nil {
			return err
		}

		// add a new client
		manager.wg.Add(1)

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
	cxt, cancel := context.WithTimeout(context.Background(), ClientShutdownTime)

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

	// remove client from manager
	manager.wg.Done()

	// lock the system and remove rest of stuff,
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

	// run shutdown on all clients
	for _, value := range manager.clients {
		// run shutdown in it's own threads
		// do not remove all the clients
		// as that would use up too much time
		// for not used data.
		go func(val *Client) {
			val.Shutdown(childCxt)
			// Decrement the wait group here
			// as there is no degregisteation.
			manager.wg.Done()
		}(value)
	}

	select {
	case <-waiter:
		logger.Debug("client manager shutdown complete")
	case <-cxt.Done():
		logger.Info("client manager was forcefully shutdown")
	}
}

// broadCast will send the given message to all the connected clients.
// This function allows for a blacklist, so that the server does not send
// the broadcast to those clients.
func (manager *clientManager) broadCast(msg server.ServerOpcode, except []uint8) {
	// TODO: implement broadCast
	// This waits on final client implementation.

	// TODO: write benchmark for this
	// Attention this might be faster via a different implementation.

	// write except onto linked list, so that removing an element is in O(1)
	blackList := list.New()
	for blacklisted := range except {
		blackList.PushBack(blacklisted)
	}

	// use a reader lock for this, as any kind of write would be problematic.
	manager.lock.RLock()
	defer manager.lock.RUnlock()

	// loop over all the connected clients and internaly over the except
	for id, client := range manager.clients {
		// check if client is still active, if not ignore that client
		if client.status != Active {
			continue
		}

		for elem := blackList.Front(); elem != nil; elem = elem.Next() {
			if elem.Value.(uint8) != id {
				// TODO: send message to client

			} else {
				// Remove element from black list to speed up internal loop
				tmp := elem
				elem = elem.Next()
				blackList.Remove(tmp)
			}
		}
	}
}
