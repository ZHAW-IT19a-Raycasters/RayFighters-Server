package items

import (
	"context"
	"time"

	"github.zhaw.ch/raycasters/server/gamemap"
	"github.zhaw.ch/raycasters/server/utils"
)

const ItemTimeout = 1 * time.Minute

// getTime is the timer function used for
// mocking in tests.
var getTime = time.Now

// timerQueue is an alias of the more complex channel
// needed for the item time data structures.
type timerQueue chan itemTimeData

// void is an emptry structure used
// for a a zero memory channel.
type void struct{}

// itemTimeData is the structure holding all the information
// needed to handle all the item timeouts used on item
// pickup.
type itemTimeData struct {
	coord gamemap.MapCoordinate
	time  time.Time
}

// GetCoord returns the coordinates at which the item
// used to exist at.
func (it itemTimeData) GetItem() interface{} {
	return it.coord
}

// GetTime will return the time when the object was
// created.
func (it itemTimeData) GetTime() time.Time {
	return it.time
}

// newItemTimeData will create the new instance of
// sendable item data information.
func newItemTimeData(coord gamemap.MapCoordinate) itemTimeData {
	return itemTimeData{
		coord: coord,
		time:  getTime(),
	}
}

// ItemTimer is the data handler
type ItemTimer struct {
	callback  func(coord gamemap.MapCoordinate)
	baseTimer utils.BaseTimer
}

func NewItemTimer(itemBuffer int, callback func(coord gamemap.MapCoordinate)) ItemTimer {
	it := ItemTimer{
		callback: callback,
	}

	iCallback := func(elem interface{}) {
		// as this function can only be called
		// from the given context of the
		// ItemTimer the element will always
		// be a gamemap.MapCoordinate
		coord, _ := elem.(gamemap.MapCoordinate)
		it.callback(coord)
	}

	it.baseTimer = utils.NewBaseTimer(itemBuffer, ItemTimeout, iCallback)

	return it
}

// Enqueue will add the given item coordinate to the
// queue of data that is timed out.
func (t *ItemTimer) Enqueue(coord gamemap.MapCoordinate) {
	td := newItemTimeData(coord)
	t.baseTimer.Enqueue(td)
}

// Run will activate and run the Item timer
func (t *ItemTimer) Run() error {
	return t.baseTimer.Run()
}

// Shutdown will gracefully shutdown the item timer or
// forcefully exit if the context is closed.
func (t *ItemTimer) Shutdown(cxt context.Context) error {
	return t.baseTimer.Shutdown(cxt)
}
