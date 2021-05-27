package client

import (
	"context"
	"testing"
	"time"
)

func TestRespawn(t *testing.T) {
	const expId = 1
	// fix timeout (so that we don't wait multiple seconds)
	const clienttimeout = 50 * time.Millisecond

	queue := make(chan uint8)
	callback := func(id uint8) {
		queue <- id
	}

	ht := NewHubRespawnHandler(hubTimerSize, callback)
	ht.baseTimer.SetTimeout(clienttimeout)

	ht.Run()

	defer ht.Shutdown(context.TODO())

	ht.RegisterRespawn(expId)

	got := <-queue

	if got != expId {
		t.Error("Incorrect id used")
		return
	}

	// try to resend a timeout
	ht.RegisterRespawn(expId)

	got = <-queue

	if got != expId {
		t.Error("Incorrect id used")
		return
	}
}
