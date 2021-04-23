package client

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.zhaw.ch/raycasters/server/room/helpers"
)

func setupTimer(t *testing.T, f func() time.Time) {
	org := getTime
	getTime = f

	t.Cleanup(func() {
		getTime = org
	})
}

const hubTimerSize = helpers.IdCountLimit

func TestTimer(t *testing.T) {
	const expId = 1
	// fix timeout (so that we don't wait multiple seconds)
	oldTimeout := ClientTimeout
	t.Cleanup(
		func() {
			ClientTimeout = oldTimeout
		},
	)

	ClientTimeout = 50 * time.Millisecond

	queue := make(chan uint8)
	callback := func(id uint8) {
		queue <- id
	}

	ht := NewHubTimer(hubTimerSize, callback)
	ht.Run()

	defer ht.Shutdown(context.TODO())

	ct := ht.NewClientTimer(expId)

	// try to reset here
	if ct.Reset() {
		t.Error("A reset should not be possible currently")
		return
	}

	err := ct.RegisterTimeout()
	if err != nil {
		t.Error("There should not be a timeout registered currently")
		return
	}

	got := <-queue

	if got != expId {
		t.Error("Incorrect id used")
		return
	}

	// try to resend a timeout
	err = ct.RegisterTimeout()
	if err == nil {
		t.Error("There should be an error here, as there was a timeout already registered")
		return
	} else if !errors.Is(err, ErrAlreadyTimeout) {
		t.Error("Wrong error type used")
		return
	}

	if !ct.Reset() {
		t.Error("Reset failed")
		return
	}

	// success again
	err = ct.RegisterTimeout()
	if err != nil {
		t.Error("There should not be a timeout registered currently")
		return
	}

	got = <-queue

	if got != expId {
		t.Error("Incorrect id used")
		return
	}
}
