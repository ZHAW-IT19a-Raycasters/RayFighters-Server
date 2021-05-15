package controller

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"sync"

	"github.zhaw.ch/raycasters/server/room/mapstate"

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
	// mapManager is a Manager controlling the events of the map
	mapController mapstate.MapController
	// timeoutClientQueue stores the currently timedout clients
	// so that they can be removed once the timeout runs out.
	timeoutClientQueue client.HubTimeoutHandler
	// respawnHandler stores and later on handles
	// the currently killed players so that they
	// can be reborn once the timeout runs out.
	respawnHandler client.HubRespawnHandler
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

func getMap() string {
	_, filename, _, _ := runtime.Caller(0)
	p := filepath.Dir(filename)
	p = filepath.Dir(p)
	p = filepath.Dir(p)

	return filepath.Join(p, "maps", "desert.txt")
}

// NewHubController will create a new controller
func NewHubController() *HubController {
	// TODO: handle error
	// TODO: fix issue with file paths...
	// TODO: remove panics :(
	path := getMap()
	mapController, err := mapstate.NewMapController(path)
	if err != nil {
		panic(fmt.Sprintf("error happend while opening the map controller [%s]", err.Error()))
	}

	hc := HubController{
		clientManager: client.NewClientManager(),
		connections:   make(map[uint8]client.HubQueue),
		gameState:     NewGameState(),
		mapController: mapController,
		semaphore:     utils.NewCountingSemaphore(uint32(helpers.IdCountLimit)),
		wg:            sync.WaitGroup{},
		lock:          sync.RWMutex{},
	}

	// implement item timer callbacks (blocked by missing map implementation)
	// Attention this is not done anymore, but the comment is kept here as
	// a reminder for possible continued developement.

	// client timer callbacks
	cTimeoutCallback := func(id uint8) {
		// error is not important for execution
		// as it only says if there was an
		// execution or not
		_ = hc.removeClientIfTimeout(id)
	}

	hc.timeoutClientQueue = client.NewHubTimer(helpers.IdCountLimit, cTimeoutCallback)

	// TODO: handle error
	hc.timeoutClientQueue.Run()

	// client respawn callbacks
	cRespawnCallback := func(id uint8) {
		// error is not important for execution
		// as it only says if there was an
		// execution or not
		logger.Debugf("spawning player [%d] back into the map", id)

		_ = hc.respawnPlayer(id)
	}

	hc.respawnHandler = client.NewHubRespawnHandler(helpers.IdCountLimit, cRespawnCallback)

	// TODO: handle error
	hc.respawnHandler.Run()

	return &hc
}

func (hc *HubController) respawnPlayer(id uint8) error {
	logger.Debugf("trying to respawn player [%d]", id)
	// check if the player is stil active
	// only respawn if the client is still
	// active.
	err := hc.checkClientState(id, client.Active)

	if err != nil {
		logger.Errorf("something unexpected happend while checking for the player respawn state [%s]", err.Error())
		return err
	}

	// error not needed here as we previously already
	// checked if the player exists or not.
	coord, _ := hc.mapController.RandomSpawn(id)

	// teleport the player back into the game
	tp := server.NewTeleportPlayer(id, coord.X(), coord.Y())

	logger.Debugf("player [%d] teleported to x: [%f] y: [%f]", id, tp.CoorX(), tp.CoorY())

	hc.broadCast(tp, helpers.NewIdSet())

	return nil
}

func (hc *HubController) checkClientState(id uint8, expectedState client.ConnectionStatus) error {
	hc.lock.RLock()
	defer hc.lock.RUnlock()
	status, ok := hc.clientManager.GetClientStatus(id)
	if !ok {
		return ErrClientDoesNotExist{
			id: id,
		}
	} else if status != expectedState {
		return ErrClientUnexpectedState{
			state: expectedState,
			id:    id,
		}
	}
	return nil
}

// removeClientIfTimeout will remove a client only
// if the client is timedout
func (hc *HubController) removeClientIfTimeout(id uint8) error {

	err := hc.checkClientState(id, client.Timeout)

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

	hc.clientManager.SetClientState(meta, client.Offline)

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

// ClientState checks if the client has a place on the map or
// not.
func (hc *HubController) ClientState(meta client.ClientMeta) (client.ConnectionStatus, bool) {
	// check the meta information, so that
	// there never can be a double increment
	// per added client.
	c := client.Offline

	hc.lock.RLock()
	defer hc.lock.RUnlock()
	id, ok := hc.clientManager.MapId(meta)
	if !ok {
		return c, ok
	}

	// okay is not needed here as we know the clinet already exists.
	c, ok = hc.clientManager.GetClientStatus(id)
	return c, ok
}

// RegisterClient will add the client to the controller
func (hc *HubController) RegisterClient(meta client.ClientMeta) error {
	// TODO: register in the game state (if it is a new connection)
	logger.Debugf("registering client in hub [%s]", meta.GetId().String())

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

	coord, err := hc.mapController.AddClient(id)
	if err != nil {
		return err
	}

	logger.Debugf("Spawning player [%d] at x : [%f] y : [%f]", id, coord.X(), coord.Y())

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

	teleportPlayer := server.NewTeleportPlayer(id, coord.X(), coord.Y())
	hc.broadCast(teleportPlayer, helpers.NewIdSet())

	return nil
}

// runClientConnection is the main executing function for interacting with
// the client. It handles all the incoming opcodes received from the
// client.
func (hc *HubController) runClientConnection(hqueue client.HubQueue, id uint8) {
	var wg sync.WaitGroup

	rClient := func(id uint8) {
		// We ignore the ok because this function can't be called without the user already existing
		meta, ok := hc.clientManager.MapMeta(id)

		if !ok {
			return
		}
		// We ignore the error here, as it is impossible per definition
		_ = hc.removeClient(meta)
	}

	defer func() {
		wg.Wait()
		rClient(id)
	}()

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
			// this means that the given opcode will not actually
			// result in a direct return message, this happens
			// with broadcast commands.
			if gotten == nil {
				return
			}

			toSend := gotten.Bytes()
			hqueue.Enqueue(toSend)

			if toSend[0] == server.UnknownOpcode.Value() {
				return
			}
		}()
	}
}

// removeClient is the actual functionality, that is
// responsible for removing a given client from the
// hub.
// Attention this function assumes the client was already send or
// has send a disconnection message.
// Attention decrementing the client may only happen
// in a case, that there actually was a client removed.
func (hc *HubController) removeClient(meta client.ClientMeta) error {

	state, ok := hc.ClientState(meta)
	if !ok {
		return client.NewErrClientDoesntExist(meta)
	} else if state == client.Timeout {
		return nil
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
		hc.respawnHandler.Shutdown,
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
// id - client id
func (hc *HubController) handleRead(data []byte, id uint8) protocol.Bytefy {
	// we can omit the nil check here
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

func (hc *HubController) handleMovement(message []byte, id uint8) protocol.Bytefy {
	movement, err := pclient.NewMovementByte(message)
	if err != nil {
		logger.Infof("[%s] - message [% X]", err.Error(), message)
		newUnknown := server.NewUnknown()
		return newUnknown
	}

	coordinate, err := hc.mapController.MoveClient(id, movement.CoorX(), movement.CoorY(), movement.AnglPl())
	if err != nil {
		newTeleportPLayer := server.NewTeleportPlayer(id, coordinate.X(), coordinate.Y())
		return newTeleportPLayer
	}
	newMovementPlayer := server.NewMovementPlayer(id, coordinate.X(), coordinate.Y(), coordinate.Angle())
	// don't send the movement information back to the actual player
	// to avoid unneeded information
	except := helpers.NewIdSet()
	except.Add(id)
	hc.broadCast(newMovementPlayer, except)

	// As the movement opcode does not actually require
	// an answer send nothing back.
	return nil
}

func (hc *HubController) handleShooting(t []byte, shooterId uint8) protocol.Bytefy {
	shot, err := pclient.NewShootingBytes(t)
	if err != nil {
		logger.Info(err)
		newUnknown := server.NewUnknown()
		return newUnknown
	}

	shotAngle := shot.AngPl()
	hit, victimId := hc.mapController.CheckShooting(shooterId, shotAngle)

	aShot := server.NewShootingPlayer(shooterId, shotAngle)

	// add id's to be ignored by broadcasting
	ignoreId := helpers.NewIdSet()
	ignoreId.Add(shooterId)

	// inform the police and players about shooting
	hc.broadCast(aShot, ignoreId)

	// someone is dead :(
	if hit {
		death := server.NewPlayerDeath(shooterId, victimId)
		// broadcast the death to all players
		hc.broadCast(death, helpers.NewIdSet())

		// register the dead player for respawn
		hc.respawnHandler.RegisterRespawn(victimId)
	}
	// As the shooting opcode does not require an answer to be sent back
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
