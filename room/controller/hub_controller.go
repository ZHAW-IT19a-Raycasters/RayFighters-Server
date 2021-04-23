package controller

import (
	"context"
	"sync"

	"github.zhaw.ch/raycasters/server/protocol/server"
	"github.zhaw.ch/raycasters/server/utils"

	"github.zhaw.ch/raycasters/server/protocol"
	pclient "github.zhaw.ch/raycasters/server/protocol/client"
	"github.zhaw.ch/raycasters/server/room/client"
	"github.zhaw.ch/raycasters/server/room/helpers"
	"github.zhaw.ch/raycasters/server/room/items"
)

// HubController is the main brain of the hub
// it is soley responsible for all the interactions
// done by the user.
type HubController struct {
	// The manager of the different client with the different connections.
	clientManager client.ClientManager
	// connections stores the hub queues the soley conneciton
	// points to the clients.
	connections map[uint8]client.HubQueue
	// gameState stores the current state of the game
	// such as player kills and their deaths.
	gameState GameState
	// timeoutClientQueue stores the currently timedout clients
	// so that they can be removed once the timeout runs out.
	timeoutClientQueue client.HubTimer
	// timeoutItemQueue stores the currently timedout map elements
	// so that they can be respawned once the timeout runs out.
	timeoutItemQueue items.ItemTimer
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

// NewHubController will create a new controller
func NewHubController() *HubController {
	// TODO: add map implementation
	hc := HubController{
		clientManager: client.NewClientManager(),
		connections:   make(map[uint8]client.HubQueue),
		gameState:     NewGameState(),
		wg:            sync.WaitGroup{},
		lock:          sync.RWMutex{},
	}

	// TODO: implement item timer callbacks (blocked by missing map implementation)

	// client timer callbacks
	cTimerCallback := func(id uint8) {
		// error is not important for execution
		// as it only says if there was an
		// execution or not
		_ = hc.RemoveClientIfTimeout(id)
	}
	hc.timeoutClientQueue = client.NewHubTimer(helpers.IdCountLimit, cTimerCallback)

	return &hc
}

// RemoveClientIfTimeout will remove a client only
// if the client is timedout
func (hc *HubController) RemoveClientIfTimeout(id uint8) error {
	checkState := func() error {
		hc.lock.RLock()
		defer hc.lock.RUnlock()
		status, ok := hc.clientManager.GetClientStatus(id)
		if !ok {
			return ErrClientDoesNotExist{
				id: id,
			}
		} else if status != client.Timeout {
			return ErrClientNotTimedOut{
				id: id,
			}
		}
		return nil
	}

	err := checkState()
	if err != nil {
		return err
	}

	// actually timed out :(
	// remove the client but make sure for a
	// last time that there was no previous
	// removal during the moment since
	// releasing the lock.
	meta, ok := hc.clientManager.MapMeta(id)
	if !ok {
		return ErrClientDoesNotExist{
			id: id,
		}
	}
	return hc.removeClient(meta)
}

// Acquire will increment the counted value for the semaphore
func (hc *HubController) Acquire() error {
	return hc.semaphore.Acquire()
}

// Release will decrement the counterd value for the semaphore
func (hc *HubController) Release() error {
	return hc.semaphore.Release()
}

// Size will return the current client count
func (hc *HubController) Size() int {
	return hc.semaphore.Value()
}

// IsFull will check if there is still space for a new player
func (hc *HubController) IsFull() bool {
	return hc.semaphore.Full()
}

// ClientExists checks if the client has a place on the map or
// not.
func (hc *HubController) ClientExists(meta client.ClientMeta) bool {
	// check the meta information, so that
	// there never can be a double increment
	// per added client.
	hc.lock.RLock()
	defer hc.lock.RUnlock()
	_, ok := hc.clientManager.MapId(meta)
	return ok
}

// RegisterClient will add the client to the controller
func (hc *HubController) RegisterClient(meta client.ClientMeta) error {
	// TODO: register in the game state (if it is a new connection)

	// prepare the queues
	cqueue, hqueue := client.NewWorkQueue()

	// prepare messages without the new player information
	sendAllPLayerKd := hc.gameState.GetPlayersKD()
	sendAllPlayerInfo := server.NewSendAllPlayerInfo(hc.clientManager.GetPlayers())

	// directly add the client to the manager and
	// let it figure out, what to do here
	err := hc.clientManager.AddClient(meta, cqueue, hc.timeoutClientQueue.NewClientTimer)

	// something went wrong while adding the client
	if err != nil {
		return err
	}

	// get id this will return a valid value due
	// to adding above without any errors
	id, _ := hc.clientManager.MapId(meta)

	// we can ignore this error, as there is no
	// benefit of knowing again if the client id
	// already existed before.
	_ = hc.gameState.AddClient(id)

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

// runClientConnection is the main excecuting function for interacting with
// the client. It handles all the incommening opcodes received from the
// client.
func (hc *HubController) runClientConnection(hqueue client.HubQueue, id uint8) {
	var wg sync.WaitGroup
	defer wg.Wait()

	for {
		// if there was an error this means
		// that we can stop running this here
		data, err := hqueue.Dequeue()
		if err != nil {
			break
		} else if len(data) == 0 {
			logger.Error("Received no Data to Process")
			break
		}

		wg.Add(1)

		go func() {
			defer wg.Done()

			gotten := hc.handleRead(data, id)
			toSend := gotten.Bytes()
			// this means that the given opcode will not actually
			// result in a direct return message, this happens
			// with broadcast commands.
			if gotten == nil {
				return
			}
			hqueue.Enqueue(toSend)

			if toSend[0] == server.UnknownOpcode.Value() {
				// We ignore the ok because this function can't be called without the user already existing
				meta, _ := hc.clientManager.MapMeta(id)
				// We ignore the error here, as it is impossible per definition
				_ = hc.removeClient(meta)
				return
			}
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
func (hc *HubController) removeClient(meta client.ClientMeta) error {
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
	// as we do not have any type of limitation using
	// background here
	_ = hc.clientManager.RemoveClient(meta)

	hc.broadCast(server.NewPlayerLeft(id), helpers.NewIdSet())

	// decrement the counter
	// error here is impossible due to the checks above
	_ = hc.Release()

	return nil
}

// BroadCast will as the name implies broadcast a given
// byte interface to a given client.
func (hc *HubController) broadCast(msg protocol.Bytefy, except helpers.IdSet) {
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

// Shutdown will Shutdown the controller
func (hc *HubController) Shutdown(cxt context.Context) {
	// send shutdown message to all the clients
	hc.broadCast(server.NewDisconnect(), helpers.NewIdSet())

	waiter := make(chan struct{})
	wg := sync.WaitGroup{}

	// use complex way to avoid code repetition
	// thanks sonar lint alert
	shutdowns := [...]func(context.Context) error{
		func(icxt context.Context) error {
			// shutdown the item timer
			hc.clientManager.Shutdown(icxt)
			return nil
		},
		hc.timeoutClientQueue.Shutdown,
		hc.timeoutItemQueue.Shutdown,
	}

	wg.Add(len(shutdowns))

	shutdown := func(f func(context.Context) error) {
		defer wg.Done()
		// shutdown the timers
		f(cxt)
	}

	for _, f := range shutdowns {
		go shutdown(f)
	}

	go func() {
		wg.Wait()
		close(waiter)
	}()

	select {
	case <-waiter:
		// all good
		logger.Info("shutdown controller")
	case <-cxt.Done():
		logger.Info("forced shutdown of controller")
	}
}

// handleRead will as the name implies handle all the given reads.
func (hc *HubController) handleRead(data []byte, id uint8) protocol.Bytefy {
	// we can ommit the nil check here
	// as a len for a nil slice if
	// defined as 0.
	if len(data) == 0 {
		return hc.handleUnknown(id)
	}
	switch pclient.Detect(data[0]) {
	case pclient.DisconnectOpcode:
		// Received a Disconnect request from client
		return hc.handleDisconnect(data, id)
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
		return hc.handleRequestOtherPlayer(data)
	case pclient.RequestPlayerKDOpcode:
		// Received a RequestOtherPlayers request from client
		return hc.handleRequestPlayerKD(data)
	default:
		// Received Unknown Code from client
		return hc.handleUnknown(id)
	}
}

func (hc *HubController) handleUnknown(id uint8) protocol.Bytefy {
	logger.Errorf("Received Unknown Opcode from Client[%d]", id)
	return nil
}

func (hc *HubController) handleDisconnect(t []byte, id uint8) protocol.Bytefy {
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

func (hc *HubController) handleWhoAmI(t []byte, id uint8) protocol.Bytefy {
	_, err := pclient.NewWhoAmIByte(t)
	if err != nil {
		logger.Info(err)
		newUnknown := server.NewUnknown()
		return newUnknown
	}
	newYouAreCalled := server.NewYouAreCalled(id)
	return newYouAreCalled
}

func (hc *HubController) handleMovement(t []byte, id uint8) protocol.Bytefy {
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

func (hc *HubController) handleShooting(t []byte, id uint8) protocol.Bytefy {
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

func (hc *HubController) handleRequestOtherPlayer(t []byte) protocol.Bytefy {
	_, err := pclient.NewRequestOtherPlayersBytes(t)
	if err != nil {
		logger.Info(err)
		newUnknown := server.NewUnknown()
		return newUnknown
	}

	newPlayerInfo := server.NewSendAllPlayerInfo(hc.clientManager.GetPlayers())
	return newPlayerInfo
}

func (hc *HubController) handleRequestPlayerKD(t []byte) protocol.Bytefy {
	_, err := pclient.NewRequestPlayerKDByte(t)
	if err != nil {
		logger.Info(err)
		newUnknown := server.NewUnknown()
		return newUnknown
	}
	newPlayersKD := hc.gameState.GetPlayersKD()
	return newPlayersKD
}
