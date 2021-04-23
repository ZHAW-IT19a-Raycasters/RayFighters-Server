package room

import (
	"context"
	"errors"
	"math"
	"sync"
	"time"

	"github.zhaw.ch/raycasters/server/room/client"
	"github.zhaw.ch/raycasters/server/room/controller"
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

// ClientTimeout says the maximal time a client
// may not have an active connection to the
// system, before beeing removed from it.
const ClientTimeout = 5 * time.Second

// workerCount is the ammount of workers the
// hub it self has at all times.
const workerCount = 2

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
	controller *controller.HubController
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
		controller: controller.NewHubController(),
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

// RegisterClient will use the meta information given
// by the parameter to check if the player
func (h *Hub) RegisterClient(meta client.ClientMeta) error {

	// if the client does not have an entry in the
	// manager, then the player count has to be
	// incremented.
	if !h.controller.ClientExists(meta) {
		// increment the counter
		// if the counter already is maxed out
		// this will return an error
		err := h.controller.Acquire()

		if err != nil {
			return ErrHubIsFull
		}
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

	go h.runTimer()
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

	for {
		select {
		case meta := <-h.input:
			logger.Debugf("Adding a new player to hub [%d]", h.id)
			h.controller.RegisterClient(meta)
		case <-h.shutdown:
			logger.Infof("Shutting down hub [%d]", h.id)
			return
		}
	}
}

// runTimer is the timer instance that governs the hub
func (h *Hub) runTimer() {
	defer h.wg.Done()
	// create the timer without any time to wait
	timerHandler := time.NewTimer(0)
	// temporily save the current client to wait for
	var currentlyTimedOut client.TimerData

	for {

		// wait here if there is no other timeout currently
		select {
		case <-h.shutdown:
			// shutting down
			return
		case currentlyTimedOut = <-h.controller.TimerQueue():
			// set the timer to only wait for a moment
			diff := time.Until(currentlyTimedOut.GetTime().Add(ClientTimeout))
			timerHandler.Reset(diff)
		}

		// wait here until the timer is done for
		select {
		case <-h.shutdown:
			// shutting down
			return
		case <-timerHandler.C:
			// error irrelevant for system
			_ = h.controller.RemoveClientIfTimeout(currentlyTimedOut.GetId())
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
