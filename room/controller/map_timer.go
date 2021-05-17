package controller

import (
	"context"
	"github.zhaw.ch/raycasters/server/utils"
	"time"
)

const MapTimeout = 5 * time.Minute

// getTime is the timer function used for
// mocking in tests.
var getTime = time.Now

// mapTimeData is the structure holding all the information
// needed to handle all the map timeouts used on map
// pickup.
type mapTimeData struct {
	time  time.Time
}

// newmapTimeData will create the new instance of
// sendable map data information.
func newmapTimeData() mapTimeData {
	return mapTimeData{
		time:  getTime(),
	}
}

// GetItem returns nil because no value is saved
func (it mapTimeData) GetItem() interface{} {
	// TODO: understand what to return here
	return nil
}

// GetTime will return the time when the object was
// created.
func (it mapTimeData) GetTime() time.Time {
	return it.time
}

// mapTimer is the data handler
type MapTimer struct {
	callback  func(elem interface{})
	baseTimer utils.BaseTimer
}

func NewMapTimer(mapBuffer int, callback func(elem interface{})) MapTimer {
	it := MapTimer{
		callback: callback,
	}

	iCallback := func(elem interface{}) {
		// as this function can only be called
		// from the given context of the
		// mapTimer the element will always
		// be a gamemap.MapCoordinate
		it.callback(elem)
	}

	it.baseTimer = utils.NewBaseTimer(mapBuffer, MapTimeout, iCallback)

	return it
}

// Enqueue will add the given map coordinate to the
// queue of data that is timed out.
func (t *MapTimer) Enqueue() {
	td := newmapTimeData()
	t.baseTimer.Enqueue(td)
}

// Run will activate and run the map timer
func (t *MapTimer) Run() error {
	return t.baseTimer.Run()
}

// Shutdown will gracefully shutdown the map timer or
// forcefully exit if the context is closed.
func (t *MapTimer) Shutdown(cxt context.Context) error {
	return t.baseTimer.Shutdown(cxt)
}
