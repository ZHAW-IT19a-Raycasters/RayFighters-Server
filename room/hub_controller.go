package room

import "github.zhaw.ch/raycasters/server/utils"

type hubController struct {
	// The manager of the different client with the different connections.
	clientManager clientManager
	// Is the player counter, this variable counts the active
	// player connections, although that number is already available
	// via the client map length, is would have to be accessed
	// from behind a lock. And so creates an unneeded overhead
	// while asynchronously adding a new client. So to avoid the given
	// overhead an atomic implementation is here used.
	counter utils.Semaphore
}

func newHubController() *hubController {
	return &hubController{}
}
