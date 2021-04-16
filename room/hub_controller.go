package room

import (
	"context"
	"sync"

	"github.zhaw.ch/raycasters/server/protocol/server"

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
	gameState     GameState
}

func newHubController() *hubController {
	// TODO: add map implementation
	return &hubController{
		clientManager: client.NewClientManager(),
		connections:   make(map[uint8]helpers.HubQueue),
		wg:            sync.WaitGroup{},
		lock:          sync.RWMutex{},
		gameState:     NewGameState(),
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
	// TODO: implement broadcast to all available players

	// prepare the queues
	cqueue, hqueue := helpers.NewWorkQueue()

	// prepare messages without the new player information
	sendAllPLayerKd := hc.gameState.GetPlayersKD()
	sendAllPlayerInfo := server.NewSendAllPlayerInfo(hc.clientManager.GetPlayers())

	// directly add the client to the manager and
	// let it figure out, what to do here
	err := hc.clientManager.AddClient(meta, cqueue)

	// something went wrong while adding the client
	if err != nil {
		return err
	}

	// get id this will return a valid value due
	// to adding above without any errors
	id, _ := hc.clientManager.MapId(meta)

	// initialize the queues with the start up messages
	youAreCalled := server.NewYouAreCalled(id)

	// TODO: once map is finished update spawn for fitting coordinates
	teleportPlayer := server.NewTeleportPlayer(id, 0, 0)

	// send all information in the specified order
	// Attention although the order is specified
	// by the protocol the websocket protocol
	// does not garantie it.
	messages := [...]protocol.Bytefy{
		youAreCalled,
		sendAllPLayerKd,
		sendAllPlayerInfo,
		teleportPlayer,
	}

	for _, message := range messages {
		hqueue.Enqueue(message.Bytes())
	}

	// add the queue to the controller
	hc.lock.Lock()
	hc.connections[id] = hqueue
	hc.lock.Unlock()

	hc.wg.Add(1)
	// run the client computations
	go func() {
		var wg sync.WaitGroup

		defer func() {
			wg.Wait()
			hc.wg.Done()
		}()

		for {
			// if there was an error this means
			// that we can stop running this here
			data, err := hqueue.Dequeue()
			if err != nil {
				break
			}

			wg.Add(1)

			go func() {
				defer wg.Done()

				gotten := hc.handleRead(data, id)
				// this means that the given opcode will not actually
				// result in a direct return message, this happens
				// with broadcast commands.
				if gotten == nil {
					return
				}

				hqueue.Enqueue(gotten.Bytes())
			}()
		}
	}()

	return nil
}

// removeClient will ad the name implies remove a client from the
// client.
// Attention this function assumes the client was already send or
// has send a disconnection message.
func (hc *hubController) removeClient(meta client.ClientMeta) error {
	// TODO: add implementation
	if !hc.clientExists(meta) {
		return client.NewErrClientDoesntExist(meta)
	}

	// client exists so lock everything down to remove him

	// error is not needed as per definition the client has to exists
	// at this point
	id, _ := hc.clientManager.MapId(meta)
	_ = hc.clientManager.RemoveClient(meta)

	// close queue
	q := hc.connections[id]
	q.Shutdown()

	// only lock here as the remove client function is a blocking call
	hc.lock.RLock()
	defer hc.lock.RUnlock()

	// TODO: remove from map
	_ = hc.gameState.RemoveClient(id)

	// remove queue
	delete(hc.connections, id)

	return nil
}

// BroadCast will as the name implies broadcast a given
// byte interface to a given client.
func (hc *hubController) broadCast(msg protocol.Bytefy, except helpers.IdSet) {
	// TODO: implement
	hc.lock.RLock()
	defer hc.lock.RUnlock()
	// get the bytes first
	data := msg.Bytes()

	for id, clientQueue := range hc.connections {
		if except.Has(id) {
			continue
		}
		clientQueue.Enqueue(data)
	}
}

// shutdown will shutdown the controller
func (hc *hubController) shutdown(cxt context.Context) {
	// send shutdown message to all the clients
	hc.broadCast(server.NewDisconnect(), helpers.NewIdSet())

	// shutdown the controller
	hc.clientManager.Shutdown(cxt)
}

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
	return nil
}

func (hc *hubController) handleUnknownReceived(t []byte) protocol.Bytefy {
	_, err := pclient.NewUnknownByte(t)
	if err != nil {
		logger.Info(err)
		newUnknown := server.NewUnknown()
		return newUnknown
	}
	newUnknown := server.NewUnknown()
	return newUnknown
}

func (hc *hubController) handleDisconnect(t []byte, id uint8) protocol.Bytefy {
	_, err := pclient.NewDisconnectByte(t)
	if err != nil {
		logger.Info(err)
		newUnknown := server.NewUnknown()
		return newUnknown
	}
	newDisconnect := server.NewDisconnect()
	// We ignore the ok because this function can't be called without the user already existing
	meta, _ := hc.clientManager.MapMeta(id)
	// We ignore the error here, as it is impossible per definition
	_ = hc.removeClient(meta)
	return newDisconnect
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
	}
	newYouAreCalled := server.NewYouAreCalled(id)
	return newYouAreCalled
}

func (hc *hubController) handleMovement(t []byte, id uint8) protocol.Bytefy {
	movement, err := pclient.NewMovementByte(t)
	if err != nil {
		logger.Info(err)
		newUnknown := server.NewUnknown()
		return newUnknown
	}
	// TODO: implement movement checking and in the negative case
	// return a teleport action.
	newMovementPlayer := server.NewMovementPlayer(id, movement.CoorX(), movement.CoorY(), movement.AnglPl())
	// don't send the movement information back to the actuall player
	// to avoid unneded information
	except := helpers.NewIdSet()
	except.Add(id)
	hc.broadCast(newMovementPlayer, except)

	// As the movement opcode does not actually requiere
	// an answer send nothing back.
	return nil
}

func (hc *hubController) handleShooting(t []byte, id uint8) protocol.Bytefy {
	shot, err := pclient.NewShootingBytes(t)
	if err != nil {
		logger.Info(err)
		newUnknown := server.NewUnknown()
		return newUnknown
	}
	// TODO: IMPLEMENT SHOOTING FUNCTION WITH RAYCASTING
	// newAnotherPlayerKilled := server.NewPlayerDeath(id, id)

	// dues to the not yet implemented raycasting feature
	// checking, who died is not possible, so simply propagade
	// the information.
	aShot := server.NewShootingPlayer(id, shot.AngPl())

	// broadcast the death to all players
	except := helpers.NewIdSet()
	except.Add(id)
	hc.broadCast(aShot, except)

	return nil
}

func (hc *hubController) handleRequestOtherPlayer(t []byte, id uint8) protocol.Bytefy {
	_, err := pclient.NewRequestOtherPlayersBytes(t)
	if err != nil {
		logger.Info(err)
		newUnknown := server.NewUnknown()
		return newUnknown
	}
	newPlayerInfo := server.NewSendAllPlayerInfo(hc.clientManager.GetPlayers())
	return newPlayerInfo
}

func (hc *hubController) handleRequestPlayerKD(t []byte, id uint8) protocol.Bytefy {
	_, err := pclient.NewRequestPlayerKDByte(t)
	if err != nil {
		logger.Info(err)
		newUnknown := server.NewUnknown()
		return newUnknown
	}
	newPlayersKD := hc.gameState.GetPlayersKD()
	return newPlayersKD
}
