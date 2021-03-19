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
	conn chan *Client
	// The room id
	id int
	// an overview of all the clients in the room currently.
	clients map[uint8]*Client
	// Counts the ammount of connections are currently open and
	// running.
	wg sync.WaitGroup
	// The read writer lock used to block the client list on inserting
	// a client.
	sync.RWMutex
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
	}()

	select {
	case <-cxt.Done():
		// TODO: add angry logging messages that
		// a forced shutdown was made.
	case <-waitCompletion:
		// TODO: add relaxed logging messages
	}
}
