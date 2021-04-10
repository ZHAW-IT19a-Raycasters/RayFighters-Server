package room

import (
	"context"
	"sync"

	"github.zhaw.ch/raycasters/server/protocol"
)

type hubController struct {
	// The manager of the different client with the different connections.
	clientManager clientManager
	lock          sync.RWMutex
}

func newHubController() *hubController {
	return &hubController{
		clientManager: newClientManager(),
		lock:          sync.RWMutex{},
	}
}

func (hc *hubController) clientExists(meta ClientMeta) bool {
	// check the meta information, so that
	// there never can be a double increment
	// per added client.
	hc.lock.RLock()
	defer hc.lock.RUnlock()
	_, ok := hc.clientManager.mapId(meta)
	return ok
}

func (hc *hubController) registerClient(meta ClientMeta) error {
	// TODO: implement

	return nil
}

func (hc *hubController) removeClient(meta ClientMeta) error {
	// TODO: add implementation

	return nil
}

// BroadCast will as the name implies broadcast a given
// byte interface to a given client.
func (h *hubController) broadCast(msg protocol.Bytefy) {
	h.clientManager.broadCast(msg)
}

func (h *hubController) shutdown(cxt context.Context) {
	// TODO:
}

// func (h *hubController) PlayerGotHit() (bool, server.Player) {
// 	// TODO: implement the player got hit interface
// 	// figure out if this is the correct path for
// 	// comminication
// 	return false, server.NewPlayer(0, "")
// }

// func (h *hubController) PlayersKD() []server.PlayerKD {
// 	// TODO: implement this
// 	// figure out if this is the correct path
// 	// for communication between the client
// 	// and the hub

// 	return []server.PlayerKD{}
// }

// // GetAllPlayers will generate a slice
// // of send able player data.
// func (h *hubController) GetAllPlayers() []server.Player {
// 	// lock the client manager, so that there is no
// 	// changes during this process.
// 	h.clientManager.lock.RLock()
// 	defer h.clientManager.lock.RUnlock()

// 	// pre-allocate all the players
// 	players := make([]server.Player, h.clientManager.size())

// 	i := 0 // add counter
// 	for userID, client := range h.clientManager.clients {
// 		players[i] = server.NewPlayer(userID, client.meta.name)
// 		i += 1
// 	}

// 	return players
// }
