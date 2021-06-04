package client

import (
	"context"
	"errors"
	"time"

	"github.zhaw.ch/raycasters/server/utils"
)

// ClientTimeout says the maximal time a client
// may not have an active connection to the
// system, before beeing removed from it.
const ClientTimeout = 5 * time.Second

// getTime is the function that will return the
// current time.
// Attention this is used like this for testing
// purposes.
var getTime = time.Now

// ErrAlreadyTimeout is the error returned once
// a timeout would be repeatedly timeout.
var ErrAlreadyTimeout = errors.New("this client is already timed out")

// timerData is the struct holding all the
// information needed to handle all the
// timeouts
type timerData struct {
	id   uint8
	from time.Time
}

// newTimoutData will return a new TimerData instance.
func newTimoutData(id uint8) timerData {
	return timerData{
		id:   id,
		from: getTime(),
	}
}

// GetId will return the id, of the
// timed out client.
func (td timerData) GetItem() interface{} {
	return td.id
}

// GetTime will return the moment
// this timeout has been registered.
func (td timerData) GetTime() time.Time {
	return td.from
}

// ClientTimer is the timer queue used by the
// client.
type ClientTimer struct {
	alreadySend utils.AtomicBool
	id          uint8
	sender      func(data timerData)
}

// RegisterTimeout will register a new timeout
// if and only if there was no previously registered
// one.
func (ct *ClientTimer) RegisterTimeout() error {
	if !ct.alreadySend.SetTrue() {
		return ErrAlreadyTimeout
	}

	ct.sender(newTimoutData(ct.id))

	return nil
}

// Reset will reset the timer queue to
// be able to timeout once again.
func (ct *ClientTimer) Reset() bool {
	return ct.alreadySend.SetFalse()
}

// newClientTimer will create a new instance of
// a client timeout handler.
func newClientTimer(id uint8, sender func(timerData)) ClientTimer {
	return ClientTimer{
		alreadySend: utils.NewAtomicBool(false),
		id:          id,
		sender:      sender,
	}
}

// HubTimeoutHandler is the parent handler for any timeouts
// registered by clients.
type HubTimeoutHandler struct {
	callback  func(id uint8)
	baseTimer utils.BaseTimer
}

// NewHubTimer will create a new hub timer instance.
func NewHubTimer(bufferSize int, callback func(id uint8)) HubTimeoutHandler {
	ht := HubTimeoutHandler{
		callback: callback,
	}

	iCallback := func(elem interface{}) {
		// as this function can only be called
		// from the given context of the
		// ItemTimer the element will always
		// be a uint8
		id, _ := elem.(uint8)
		ht.callback(id)
	}

	ht.baseTimer = utils.NewBaseTimer(bufferSize, ClientTimeout, iCallback)

	return ht
}

// Run will run the timer instance
func (ht *HubTimeoutHandler) Run() error {
	return ht.baseTimer.Run()
}

// Shutdown will shutdown the timer
func (ht *HubTimeoutHandler) Shutdown(cxt context.Context) error {
	return ht.baseTimer.Shutdown(cxt)
}

// NewClientTimer will create a new client timer
// connected to this hub timer.
func (ht *HubTimeoutHandler) NewClientTimer(id uint8) ClientTimer {
	sender := func(td timerData) {
		ht.baseTimer.Enqueue(td)
	}
	return newClientTimer(id, sender)
}
