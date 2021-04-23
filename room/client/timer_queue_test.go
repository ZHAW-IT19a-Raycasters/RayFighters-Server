package client

import (
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

func setupTimerDefault(t *testing.T) time.Time {
	timeToTest := time.Now()

	setupTimer(t, func() time.Time {
		return timeToTest
	})
	return timeToTest
}

func TestTimer(t *testing.T) {
	const expId = 1
	expTime := setupTimerDefault(t)

	ht := NewHubTimer(hubTimerSize)
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

	got := <-ht.GetQueue()

	if expTime != got.GetTime() {
		t.Error("incorrect time was returned")
		return
	} else if got.GetId() != expId {
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

	got = <-ht.GetQueue()

	if expTime != got.GetTime() {
		t.Error("incorrect time was returned")
		return
	} else if got.GetId() != expId {
		t.Error("Incorrect id used")
		return
	}
}
