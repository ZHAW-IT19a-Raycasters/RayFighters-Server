package room

import (
	"context"
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
				gotten := hc.handleRead(data)
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
func (h *hubController) broadCast(msg protocol.Bytefy) {
	// TODO: implement
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

func (c *hubController) handleRead(data []byte) protocol.Bytefy {
	switch pclient.Detect(data[0]) {
	case pclient.DisconnectOpcode:
		// Received a Disconnect request from client
		c.handleDisconnect(data)
	case pclient.UnknownOpcode:
		// hubController received Unknown request from Server
		c.handleUnknownSent(data)
	case pclient.WhoAmIOpcode:
		// Received a WhoAmI request from client
		c.handleWhoAmI(data)
	case pclient.MovementOpcode:
		// Received a Movement request from client
		c.handleMovement(data)
	case pclient.ShootingOpcode:
		// Received a Shooting request from client
		c.handleShooting(data)
	case pclient.RequestOtherPlayersOpcode:
		// Received a RequestOtherPlayers request from client
		c.handleRequestOtherPlayer(data)
	case pclient.RequestPlayerKDOpcode:
		// Received a RequestOtherPlayers request from client
		c.handleRequestOtherPlayer(data)
	default:
		// Received Unknown Code from client
		c.handleUnknownReceived(data)
	}
	// TODO: change to actuall return
	return nil
}

func (c *hubController) handleUnknownReceived(t []byte) {
	_, err := pclient.NewUnknownByte(t)
	if err != nil {
		logger.Info(err)
		//newUnknown := server.NewUnknown()
		//c.SendData(newUnknown)
	} else {
		//newUnknown := server.NewUnknown()
		//c.SendData(newUnknown)
	}
}

func (c *hubController) handleDisconnect(t []byte) {
	_, err := pclient.NewDisconnectByte(t)
	if err != nil {
		logger.Info(err)
		//newUnknown := server.NewUnknown()
		//c.SendData(newUnknown)
	} else {
		//newDisconnect := server.NewDisconnect()
		//c.SendData(newDisconnect)

		//c.hub.UnregisterhubController(c.meta.getMeta())
	}
}

func (c *hubController) handleUnknownSent(t []byte) {
	logger.Error("Unknown Opcode was sent")
}

func (c *hubController) handleWhoAmI(t []byte) {
	_, err := pclient.NewWhoAmIByte(t)
	if err != nil {
		logger.Info(err)
		//newUnknown := server.NewUnknown()
		//c.SendData(newUnknown)
	} else {
		//newYouAreCalled := server.NewYouAreCalled(c.meta.id)
		//c.SendData(newYouAreCalled)
	}
}

func (c *hubController) handleMovement(t []byte) {
	_, err := pclient.NewMovementByte(t)
	if err != nil {
		logger.Info(err)
		//newUnknown := server.NewUnknown()
		//c.SendData(newUnknown)
	} else {
		//newMovementPlayer := server.NewMovementPlayer(c.meta.id, movement.CoorX(), movement.CoorY())
		//c.SendData(newMovementPlayer)
	}
}

func (c *hubController) handleShooting(t []byte) {
	_, err := pclient.NewShootingBytes(t)
	if err != nil {
		logger.Info(err)
		///newUnknown := server.NewUnknown()
		//c.SendData(newUnknown)
	} else {
		//hit, player := c.hub.PlayerGotHit()
		//if hit {
		//	newAnotherPlayerKilled := server.NewPlayerDeath(player.UserId(), c.meta.id)
		//	c.SendData(newAnotherPlayerKilled)
		//}
	}
}

func (c *hubController) handleRequestOtherPlayer(t []byte) {
	_, err := pclient.NewRequestOtherPlayersBytes(t)
	if err != nil {
		logger.Info(err)
		//newUnknown := server.NewUnknown()
		//c.SendData(newUnknown)
	} else {
		//newPlayerInfo := server.NewSendAllPlayerInfo(c.hub.GetAllPlayers())
		//c.SendData(newPlayerInfo)
	}
}

func (c *hubController) handleRequestPlayerKD(t []byte) {
	_, err := pclient.NewRequestPlayerKDByte(t)
	if err != nil {
		logger.Info(err)
		//newUnknown := server.NewUnknown()
		//c.SendData(newUnknown)
	} else {
		//newPlayersKD := server.NewPlayersKD(c.hub.PlayersKD())
		//c.SendData(newPlayersKD)
	}
}
