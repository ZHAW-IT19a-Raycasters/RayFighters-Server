package helpers

import (
	"errors"
	"time"

	"github.zhaw.ch/raycasters/server/utils"
)

// getTime is the function that will return the
// current time.
// Attention this is used like this for testing
// purposes.
var getTime = time.Now

// ErrAlreadyTimeout is the error returned once
// a timeout would be repeatedly timeout.
var ErrAlreadyTimeout = errors.New("this client is already timed out")

// TimerData is the struct holding all the
// information needed to handle all the
// timeouts
type TimerData struct {
	id   uint8
	from time.Time
}

// GetId will return the id, of the
// timed out client.
func (td *TimerData) GetId() uint8 {
	return td.id
}

// GetTime will return the moment
// this timeout has been registered.
func (td *TimerData) GetTime() time.Time {
	return td.from
}

// newTimeData will return a new TimerData instance.
func newTimeData(id uint8) TimerData {
	return TimerData{
		id:   id,
		from: getTime(),
	}
}

// ClientTimer is the timer queue used by the
// client.
type ClientTimer struct {
	alreadySend utils.AtomicBool
	id          uint8
	sender      chan<- TimerData
}

// RegisterTimeout will register a new timeout
// if and only if there was no prevouly registered
// one.
func (ct *ClientTimer) RegisterTimeout() error {
	if !ct.alreadySend.SetTrue() {
		return ErrAlreadyTimeout
	}

	ct.sender <- newTimeData(ct.id)
	return nil
}

// Reset will reset the timer queue to
// be able to timeout once again.
func (ct *ClientTimer) Reset() bool {
	return ct.alreadySend.SetFalse()
}

// newClientTimer will create a new instance of
// a client timeout handler.
func newClientTimer(id uint8, sender chan<- TimerData) ClientTimer {
	return ClientTimer{
		alreadySend: utils.NewAtomicBool(false),
		id:          id,
		sender:      sender,
	}
}

// HubTimer is the parent handler for any timeouts
// registered by clients.
type HubTimer struct {
	queue chan TimerData
}

// NewHubTimer will create a new hub timer instance.
func NewHubTimer(size uint) HubTimer {
	return HubTimer{
		queue: make(chan TimerData, size),
	}
}

// GetQueue will reuturn a receiver channel based
// on the main channel. This function can and
// should be used from inside a select statement.
func (ht HubTimer) GetQueue() <-chan TimerData {
	return ht.queue
}

// NewClientTimer will create a new client timer
// connected to this hub timer.
func (ht HubTimer) NewClientTimer(id uint8) ClientTimer {
	return newClientTimer(id, ht.queue)
}
