package controller

import (
	"context"
	"errors"
	"math"
	"sync"

	"github.zhaw.ch/raycasters/server/room/client"
	"github.zhaw.ch/raycasters/server/utils"
)

// void is am empty struct with no memory
// footprint, this is used for signals
// counters.
type void struct{}

// HubClientLimit is the absolute upper limit supported
// by a single hub. This limitation comes from the
// decision to only support a player id of size uint8.
const HubClientLimit = math.MaxUint8

// workerCount is the ammount of workers the
// hub it self has at all times.
const workerCount = 1

var (
	// ErrHubIsFull is the error that is returned on an RegisterClient error.
	ErrHubIsFull = errors.New("the hub is currently filled to capacity")
	// ErrHubIsNotRunning is the error that as the name implies
	// means that the hub is currently not active.
	ErrHubIsNotRunning = errors.New("hub is not running")
	// ErrHubIsAlreadyRunning is the error returned if the
	// hub is already running.
	ErrHubIsAlreadyRunning = errors.New("hub is already running")
	// ErrHubIsStillStartingUp is the error returned if the
	// hub is still in the process of starting its
	// subsystems.
	ErrHubIsStillStartingUp = errors.New("hub is still starting up")
)

// Hub represents a single room.
// The hub controls the the individual connections between the
// different modules like the map or the client manager.
type Hub struct {
	// id is the designated hub identifier
	id int
	// controller is the struct responsible for running all the client opcodes
	// and handling all the io.
	controller *HubController
	// The channel over which a new connection
	// is added. This connections input can be accessed
	// via the `RegisterClient` function associated
	// with this type.
	input chan client.ClientMeta
	// The channel that signals the main loop to shutdown.
	shutdown chan void
	// If the current hub is running or not.
	isActive utils.AtomicBool
	// wg is the wait group used to see if the
	// workers are still running
	wg *sync.WaitGroup
}

// NewHub will initiate a new hub struct.
// The hubId needs to be unique over all
// the data structures.
func NewHub(hubId int) *Hub {
	h := &Hub{
		input:      make(chan client.ClientMeta),
		controller: NewHubController(),
		shutdown:   make(chan void),
		id:         hubId,
		isActive:   utils.NewAtomicBool(false),
		wg:         &sync.WaitGroup{},
	}
	return h
}

// IsAlive will check if the current hub is running or not.
func (h *Hub) IsAlive() bool {
	return h.isActive.Get()
}

// Size will return the current player count
func (h *Hub) Size() int {
	return h.controller.Size()
}

// Full will check if the hub is at it's set
// limitation.
func (h *Hub) Full() bool {
	return h.controller.IsFull()
}

// ClientState checks for the current client state.
// Attention this allows a meta object, where the
// connection is nil.
func (h *Hub) ClientState(meta client.ClientMeta) (client.ConnectionStatus, bool) {
	return h.controller.ClientState(meta)
}

// RegisterClient will use the meta information given
// by the parameter to check if the player
func (h *Hub) RegisterClient(meta client.ClientMeta) error {

	// if the client does not have an entry in the
	// manager, then the player count has to be
	// incremented.
	_, ok := h.ClientState(meta)
	if !ok {
		// increment the counter
		// if the counter already is maxed out
		// this will return an error
		err := h.controller.Acquire()

		if err != nil {
			// close the connection
			logger.Debug("hub already full")
			return ErrHubIsFull
		}

		logger.Debugf("adding client to hub [%d]", h.id)
	}

	// add the client to the queue to be added into the system
	h.input <- meta

	return nil
}

// Run is the main event loop of the hub.
// It handles all the clients added and removed
// from the hub instance.
// Attention this is a blocking function, so
// run this in a gorutine.
func (h *Hub) Run() error {
	// check if the hub is active
	if !h.isActive.SetTrue() {
		// do not restart the hub, if it's already running
		logger.Warnf("Hub [%d] is already running", h.id)
		return ErrHubIsAlreadyRunning
	}

	// add two runners
	h.wg.Add(workerCount)

	go h.runRegistration()

	return nil
}

func (h *Hub) runRegistration() {
	// set alive bool
	// reset the state of is active, on shutdown or
	// crash
	defer func() {
		h.wg.Done()
		h.isActive.SetFalse()
	}()

	register := func(meta client.ClientMeta) error {
		logger.Debugf("Adding a new player to hub [%d]", h.id)
		err := h.controller.RegisterClient(meta)
		if err != nil {
			meta.Close()
			logger.Debugf("unable to register client because [%s]", err.Error())
		}
		return err
	}

	for {
		select {
		case meta := <-h.input:
			if !h.isActive.Get() {
				meta.Close()
				return
			}
			register(meta)
		case <-h.shutdown:
			logger.Infof("Shutting down hub [%d]", h.id)
			return
		}
	}
}

// Shutdown will initialize a complete shutdown of any and all clients.
// With the cxt context a premature cancelation of all client
// connections can be started.
func (h *Hub) Shutdown(cxt context.Context) error {

	if !h.isActive.SetFalse() {
		return ErrHubIsNotRunning
	}

	waitCompletion := make(chan void)

	go func() {
		h.controller.Shutdown(cxt)

		// make sure that the main thread is closed
		close(h.shutdown)

		// wait until the main thread and the timer
		// have shutdown
		h.wg.Wait()

		// close blocker
		close(waitCompletion)
	}()

	select {
	case <-cxt.Done():
		logger.Infof("Hub [%d] was forcefully shutdown", h.id)
	case <-waitCompletion:
		logger.Infof("Hub [%d] was successfully shutdown", h.id)
	}

	return nil
}
