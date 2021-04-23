package utils

import (
	"context"
	"testing"
	"time"
)

type MockElemInterface struct {
	elem interface{}
	time time.Time
}

func (mi *MockElemInterface) GetItem() interface{} {
	return mi.elem
}

func (mi *MockElemInterface) GetTime() time.Time {
	return mi.time
}

const bufferSize = 4
const timeOutTime = 100 * time.Millisecond
const valueToReturn = 42

func TestBaseTimerRun(t *testing.T) {

	waiter := make(chan void)

	callback := func(elem interface{}) {
		close(waiter)
	}

	bt := NewBaseTimer(bufferSize, timeOutTime, callback)

	bt.Run()

	defer func() {
		bt.Shutdown(context.TODO())
	}()

	elem := MockElemInterface{
		elem: valueToReturn,
		time: time.Now(),
	}

	start := time.Now()

	bt.Enqueue(&elem)

	<-waiter

	if time.Since(start) <= timeOutTime {
		t.Error("the timer should be waiting at least the timeout time")
	}
}

func TestBaseTimerShutdown(t *testing.T) {

	waiter := make(chan void)

	callback := func(elem interface{}) {
		close(waiter)
	}

	// use a large timeout time
	bt := NewBaseTimer(bufferSize, timeOutTime, callback)

	err := bt.Shutdown(context.TODO())

	if err == nil {
		t.Error("There should be an error here")

	}

	err = bt.Run()

	if err != nil {
		t.Error("There should not be an error here")
		return
	}

	err = bt.Run()

	if err == nil {
		t.Error("There should be an error here")
		return
	}

	elem := MockElemInterface{
		elem: valueToReturn,
		time: time.Now(),
	}

	bt.Enqueue(&elem)

	<-waiter

	// forcefully increase the timeout time
	bt.timeoutTime = 1 * time.Second

	bt.Enqueue(&elem)

	cxt, cancel := context.WithCancel(context.Background())

	waiter = make(chan void)

	go func() {
		err = bt.Shutdown(cxt)

		if err != nil {
			t.Error("There should not be an error here, as the timer is currently running as expected.")
		}
		close(waiter)
	}()

	// directly cancel, so that sometimes an additional line may be
	// hit
	cancel()

	<-waiter
}
