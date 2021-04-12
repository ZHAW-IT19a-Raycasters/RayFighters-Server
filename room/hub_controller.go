package room

import (
	"context"
	"github.zhaw.ch/raycasters/server/protocol/server"
	"sync"

	"github.zhaw.ch/raycasters/server/protocol"
	pclient "github.zhaw.ch/raycasters/server/protocol/client"
	"github.zhaw.ch/raycasters/server/room/client"
	"github.zhaw.ch/raycasters/server/room/helpers"
)

type hubController struct {
	// The manager of the different client with the different connections.
	clientManager client.ClientManager
	connections   map[uint8]helpers.HubQueue
	wg            sync.WaitGroup
	lock          sync.RWMutex
}

func newHubController() *hubController {
	// TODO: add map implementation
	return &hubController{
		clientManager: client.NewClientManager(),
		connections:   make(map[uint8]helpers.HubQueue),
		wg:            sync.WaitGroup{},
		lock:          sync.RWMutex{},
	}
}

func (hc *hubController) clientExists(meta client.ClientMeta) bool {
	// check the meta information, so that
	// there never can be a double increment
	// per added client.
	hc.lock.RLock()
	defer hc.lock.RUnlock()
	_, ok := hc.clientManager.MapId(meta)
	return ok
}

func (hc *hubController) registerClient(meta client.ClientMeta) error {
	// TODO: implement

	// prepare the queues
	cqueue, hqueue := helpers.NewWorkQueue()

	// directly add the client to the manager and
	// let it figure out, what to do here
	err := hc.clientManager.AddClient(meta, cqueue)

	// something went wrong while adding the client
	if err != nil {
		return err
	}

	// get id
	// this will return a value due to adding above without any errors
	id, _ := hc.clientManager.MapId(meta)

	// add the queue to the controller
	hc.lock.Lock()
	defer hc.lock.Unlock()
	hc.wg.Add(1)
	hc.connections[id] = hqueue

	// run the client computations
	go func() {
		defer hc.wg.Done()
		for {
			// if there was an error this means
			// that we can stop running this here
			data, err := hqueue.Dequeue()
			if err != nil {
				break
			}

			go func() {
				gotten := hc.handleRead(data, id)
				hqueue.Enqueue(gotten.Bytes())
			}()
		}
	}()

	return nil
}

func (hc *hubController) removeClient(meta client.ClientMeta) error {
	// TODO: add implementation

	return nil
}

// BroadCast will as the name implies broadcast a given
// byte interface to a given client.

func (hc *hubController) broadCast(msg protocol.Bytefy) {
	// TODO: implement
}

func (hc *hubController) shutdown(cxt context.Context) {
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


func (hc *hubController) handleRead(data []byte, id uint8) protocol.Bytefy {
	switch pclient.Detect(data[0]) {
	case pclient.DisconnectOpcode:
		// Received a Disconnect request from client
		return hc.handleDisconnect(data, id)
	case pclient.UnknownOpcode:
		// hubController received Unknown request from Server
		hc.handleUnknownSent()
	case pclient.WhoAmIOpcode:
		// Received a WhoAmI request from client
		return hc.handleWhoAmI(data, id)
	case pclient.MovementOpcode:
		// Received a Movement request from client
		return hc.handleMovement(data, id)
	case pclient.ShootingOpcode:
		// Received a Shooting request from client
		return hc.handleShooting(data, id)
	case pclient.RequestOtherPlayersOpcode:
		// Received a RequestOtherPlayers request from client
		return hc.handleRequestOtherPlayer(data, id)
	case pclient.RequestPlayerKDOpcode:
		// Received a RequestOtherPlayers request from client
		return hc.handleRequestOtherPlayer(data, id)
	default:
		// Received Unknown Code from client
		return hc.handleUnknownReceived(data)
	}
	// TODO: change to actuall return
	return nil
}

func (hc *hubController) handleUnknownReceived(t []byte) protocol.Bytefy {
	_, err := pclient.NewUnknownByte(t)
	if err != nil {
		logger.Info(err)
		newUnknown := server.NewUnknown()
		return newUnknown
	} else {
		newUnknown := server.NewUnknown()
		return newUnknown
	}
}

func (hc *hubController) handleDisconnect(t []byte, id uint8) protocol.Bytefy {
	_, err := pclient.NewDisconnectByte(t)
	if err != nil {
		logger.Info(err)
		newUnknown := server.NewUnknown()
		return newUnknown
	} else {
		newDisconnect := server.NewDisconnect()

		hc.removeClient(hc.clientManager.GetClients()[id].GetInternalMeta().GetMeta())
		return newDisconnect
	}
}

func (hc *hubController) handleUnknownSent() {
	logger.Error("Unknown Opcode was sent")
}

func (hc *hubController) handleWhoAmI(t []byte, id uint8) protocol.Bytefy {
	_, err := pclient.NewWhoAmIByte(t)
	if err != nil {
		logger.Info(err)
		newUnknown := server.NewUnknown()
		return newUnknown
	} else {
		newYouAreCalled := server.NewYouAreCalled(id) // TODO: Find ID
		return newYouAreCalled
	}
}

func (hc *hubController) handleMovement(t []byte, id uint8) protocol.Bytefy {
	movement, err := pclient.NewMovementByte(t)
	if err != nil {
		logger.Info(err)
		newUnknown := server.NewUnknown()
		return newUnknown
	} else {
		newMovementPlayer := server.NewMovementPlayer(id, movement.CoorX(), movement.CoorY())
		return newMovementPlayer
	}
}

func (hc *hubController) handleShooting(t []byte, id uint8) protocol.Bytefy {
	_, err := pclient.NewShootingBytes(t)
	if err != nil {
		logger.Info(err)
		newUnknown := server.NewUnknown()
		return newUnknown
	} else {
		// TODO: IMPLEMENT SHOOTING FUNCTION WITH RAYCASTING
		newAnotherPlayerKilled := server.NewPlayerDeath(id, id)
		return newAnotherPlayerKilled
	}
}

func (hc *hubController) handleRequestOtherPlayer(t []byte, id uint8) protocol.Bytefy {
	_, err := pclient.NewRequestOtherPlayersBytes(t)
	if err != nil {
		logger.Info(err)
		newUnknown := server.NewUnknown()
		return newUnknown
	} else {
		var players []server.Player
		for _, c := range hc.clientManager.GetClients() {
			player := server.NewPlayer(id, c.GetInternalMeta().GetName())
			players = append(players, player)
		}
		newPlayerInfo := server.NewSendAllPlayerInfo(players)
		return newPlayerInfo
	}
}

func (hc *hubController) handleRequestPlayerKD(t []byte, id uint8) protocol.Bytefy {
	_, err := pclient.NewRequestPlayerKDByte(t)
	if err != nil {
		logger.Info(err)
		newUnknown := server.NewUnknown()
		return newUnknown
	} else {
		var playersKD []server.PlayerKD
		for _, c := range hc.clientManager.GetClients() {
			playerKD := server.NewPlayerKD(id, c.GetInfo().GetKills(), c.GetInfo().GetDeaths())
			playersKD = append(playersKD, playerKD)
		}
		newPlayersKD := server.NewPlayersKD(playersKD)
		return newPlayersKD
	}
}
