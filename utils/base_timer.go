package utils

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	// ErrTimerAlreadyRunning is returned in the case that
	// a double run is called on a specific timer.
	ErrTimerAlreadyRunning = errors.New("timer already active")

	// ErrTimerNotRunning si returned on a double shutdown.
	ErrTimerNotRunning = errors.New("timer not running")
)

// ElementInterface is the interface needed for satisfying
// the base timer abstraction.
// Please do only use this struture correctly abstracted.
// So that there can never actually be an incorrect
// casting.
type ElementInterface interface {
	GetItem() interface{}
	GetTime() time.Time
}

// BaseTimer is a base used for implementing
// dynamic timer structures.
// Attention never use this struct on it's
// own rather embed it in a wrapper type
// in which safe abstractions around the
// raw interfact channel and callback
// are build.
type BaseTimer struct {
	callback    func(elem interface{})
	queue       chan ElementInterface
	timeoutTime time.Duration
	isAlive     AtomicBool
	shutdown    chan void
	lock        *sync.RWMutex
	wg          *sync.WaitGroup
}

// NewBaseTimer will create a new base timer.
func NewBaseTimer(bufferSize int, timeoutTime time.Duration, callback func(elem interface{})) BaseTimer {
	return BaseTimer{
		callback:    callback,
		queue:       make(chan ElementInterface, bufferSize),
		timeoutTime: timeoutTime,
		isAlive:     NewAtomicBool(false),
		shutdown:    make(chan void),
		lock:        &sync.RWMutex{},
		wg:          &sync.WaitGroup{},
	}
}

// SetTimerTime will set the timer timeout duration
func (t *BaseTimer) SetTimeout(timeout time.Duration) {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.timeoutTime = timeout
}

// GetTimeout will return the timeout duration
func (t *BaseTimer) GetTimeout() time.Duration {
	t.lock.RLock()
	defer t.lock.RUnlock()
	return t.timeoutTime
}

// Enqueue will add the item to the queue
func (t *BaseTimer) Enqueue(item ElementInterface) {
	t.queue <- item
}

// Run will rum the timer handler.
func (t *BaseTimer) Run() error {
	if !t.isAlive.SetTrue() {
		return ErrTimerAlreadyRunning
	}

	t.wg.Add(1)
	go t.runTimer()

	return nil
}

// runTimer is the actuall function used for the timer.
// It handles runnin the callback function for the queued
// up element once the specified time has run out.
func (t *BaseTimer) runTimer() {
	defer t.wg.Done()
	// create the timer without any time to wait
	timerHandler := time.NewTimer(0)

	// make sure the timer is expired
	<-timerHandler.C

	// temporily save the current client to wait for
	var currentWaiter ElementInterface

	for {
		// wait here if there is no other timeout currently
		select {
		case <-t.shutdown:
			// shutting down
			return
		case currentWaiter = <-t.queue:
			// set the timer to only wait for a moment
			diff := time.Until(currentWaiter.GetTime().Add(t.GetTimeout()))

			// stop the timer here
			timerHandler.Stop()

			// reset will always return true here as
			// we are waiting every time for just a single
			// event
			timerHandler.Reset(diff)
		}

		// wait here until the timer is done for
		select {
		case <-t.shutdown:
			// shutting down
			return
		case <-timerHandler.C:
			t.callback(currentWaiter.GetItem())
		}
	}
}

// Shutdown will shutdown the main timer.
func (t *BaseTimer) Shutdown(cxt context.Context) error {
	if !t.isAlive.SetFalse() {
		return ErrTimerNotRunning
	}

	waiter := make(chan void)

	go func() {
		t.wg.Wait()
		close(waiter)
	}()

	close(t.shutdown)

	select {
	case <-cxt.Done():
		return nil
	case <-waiter:
		return nil
	}
}
