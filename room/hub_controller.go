package room

import (
	"context"
	"sync"

	"github.zhaw.ch/raycasters/server/protocol/server"
	"github.zhaw.ch/raycasters/server/utils"

	"github.zhaw.ch/raycasters/server/protocol"
	pclient "github.zhaw.ch/raycasters/server/protocol/client"
	"github.zhaw.ch/raycasters/server/room/client"
	"github.zhaw.ch/raycasters/server/room/helpers"
)

type hubController struct {
	// The manager of the different client with the different connections.
	clientManager client.ClientManager
	// connections stores the hub queues the soley conneciton
	// points to the clients.
	connections map[uint8]helpers.HubQueue
	// gameState stores the current state of the game
	// such as player kills and their deaths.
	gameState GameState
	// Semaphore, counts the totally active
	// player connections, although that number is already available
	// via the client map length, is would have to be accessed
	// from behind a lock. And so creates an unneeded overhead
	// while asynchronously adding a new client. So to avoid the given
	// overhead an atomic counting semaphore implementation is here used.
	semaphore utils.CountingSemaphore
	// wg is the wait group used for counting the ammout
	// of active hub workers.
	wg sync.WaitGroup
	// lock is used to limit the access to the connections.
	// as well as to the client manager.
	lock sync.RWMutex
}

// acquire will increment the counted value for the semaphore
func (hc *hubController) acquire() error {
	return hc.semaphore.Acquire()
}

// release will decrement the counterd value for the semaphore
func (hc *hubController) release() error {
	return hc.semaphore.Release()
}

// Size will return the current client count
func (hc *hubController) Size() int {
	return hc.semaphore.Value()
}

// IsFull will check if there is still space for a new player
func (hc *hubController) IsFull() bool {
	return hc.semaphore.Full()
}

// newHubController will create a new controller
func newHubController() *hubController {
	// TODO: add map implementation
	return &hubController{
		clientManager: client.NewClientManager(),
		connections:   make(map[uint8]helpers.HubQueue),
		gameState:     NewGameState(),
		wg:            sync.WaitGroup{},
		lock:          sync.RWMutex{},
	}
}

// ClientExists checks if the client has a place on the map or
// not.
func (hc *hubController) ClientExists(meta client.ClientMeta) bool {
	// check the meta information, so that
	// there never can be a double increment
	// per added client.
	hc.lock.RLock()
	defer hc.lock.RUnlock()
	_, ok := hc.clientManager.MapId(meta)
	return ok
}

// registerClient will add the client to the controller
func (hc *hubController) registerClient(meta client.ClientMeta) error {
	// TODO: register in the game state (if it is a new connection)

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

	// send all information in the specified order
	// Attention although the order is specified
	// by the protocol the websocket protocol
	// does not garantie it.
	messages := []protocol.Bytefy{
		youAreCalled,
		sendAllPLayerKd,
		sendAllPlayerInfo,
	}

	for _, message := range messages {
		hqueue.Enqueue(message.Bytes())
	}

	// add the queue to the controller
	hc.lock.Lock()
	hc.connections[id] = hqueue
	hc.lock.Unlock()

	// run the client computations
	hc.wg.Add(1)

	go func() {
		defer hc.wg.Done()
		hc.runClientConnection(hqueue, id)
	}()

	// send the information about the client to all
	// clients. So as to make the new client spawn
	// in.
	// this will always return the value, because
	// the player was only just added above.
	player, _ := hc.clientManager.GetPlayer(id)
	playerJoined := server.NewPlayerJoined(player)

	except := helpers.NewIdSet()
	except.Add(id)
	hc.broadCast(playerJoined, except)

	// TODO: once map is finished update spawn for fitting coordinates
	// TODO: send last valid coordinates registered in case of
	// reconnecting to the server
	teleportPlayer := server.NewTeleportPlayer(id, 0, 0)
	hc.broadCast(teleportPlayer, helpers.NewIdSet())

	return nil
}

// runClientConnection
func (hc *hubController) runClientConnection(hqueue helpers.HubQueue, id uint8) {
	var wg sync.WaitGroup
	defer wg.Wait()

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
}

// removeClient is the actuall functionality, that is
// responsible for removing a given client from the
// hub.
// Attention this function assumes the client was already send or
// has send a disconnection message.
// Attention decrementing the client may only happen
// in a case, that there actually was a client removed.
func (hc *hubController) removeClient(meta client.ClientMeta) error {
	if !hc.ClientExists(meta) {
		return client.NewErrClientDoesntExist(meta)
	}

	// error is not needed as per definition the client has to exists
	// at this point
	id, _ := hc.clientManager.MapId(meta)

	// close queue
	// and initiate a shutdown of the client
	// side.
	q := hc.connections[id]
	q.Shutdown()

	// only lock here as the remove client function is a blocking call
	hc.lock.Lock()
	// remove queue
	delete(hc.connections, id)
	hc.lock.Unlock()

	// TODO: remove from map
	_ = hc.gameState.RemoveClient(id)

	// client exists so lock everything down to remove him
	_ = hc.clientManager.RemoveClient(meta)

	// decrement the counter
	// error here is imposible due to the checks above
	_ = hc.release()

	return nil
}

// BroadCast will as the name implies broadcast a given
// byte interface to a given client.
func (hc *hubController) broadCast(msg protocol.Bytefy, except helpers.IdSet) {
	// get the bytes first
	data := msg.Bytes()

	hc.lock.RLock()
	defer hc.lock.RUnlock()

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
