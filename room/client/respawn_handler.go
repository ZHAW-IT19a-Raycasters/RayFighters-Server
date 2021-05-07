package client

import (
	"context"
	"time"

	"github.zhaw.ch/raycasters/server/utils"
)

// ClientTimeout says the maximal time a client
// is blocked from playing the game until a
// respawn happens.
const ClientRespawnTime = 5 * time.Second

// respawnData is the struct holding all
// the information needed to handle all
// respawn cycles.
type respawnData struct {
	id   uint8
	from time.Time
}

// newTimoutData will return a new TimerData instance.
func newRespawnData(id uint8) respawnData {
	return respawnData{
		id:   id,
		from: getTime(),
	}
}

// GetId will return the id, of the
// timed out client.
func (td respawnData) GetItem() interface{} {
	return td.id
}

// GetTime will return the moment
// this timeout has been registered.
func (td respawnData) GetTime() time.Time {
	return td.from
}

// HubTimeoutHandler is the parent handler for all
// respawn work requiered to be done.
type HubRespawnHandler struct {
	callback  func(id uint8)
	baseTimer utils.BaseTimer
}

// NewHubRespawnHandler will create a new respawn handler
func NewHubRespawnHandler(bufferSize int, callback func(id uint8)) HubRespawnHandler {
	rh := HubRespawnHandler{
		callback: callback,
	}

	iCallback := func(elem interface{}) {
		// as this function can only be called
		// from the given context of the
		// ItemTimer the element will always
		// be a uint8
		id, _ := elem.(uint8)
		rh.callback(id)
	}

	rh.baseTimer = utils.NewBaseTimer(bufferSize, ClientRespawnTime, iCallback)

	return rh
}

// RegisterRespawn will respawn the user called by the parameter id
// once its respawn time has run out.
func (rh *HubRespawnHandler) RegisterRespawn(id uint8) {
	data := newRespawnData(id)
	rh.baseTimer.Enqueue(data)
}

// Run will run the timer instance
func (rh *HubRespawnHandler) Run() error {
	return rh.baseTimer.Run()
}

// Shutdown will shutdown the timer
func (rh *HubRespawnHandler) Shutdown(cxt context.Context) error {
	return rh.baseTimer.Shutdown(cxt)
}
