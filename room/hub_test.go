package room

import (
	"context"
	"testing"
)

func TestHub(t *testing.T) {
	const hubId = 1
	hub := NewHub(hubId)

	waitUntilActive := make(chan struct{})
	go func() {
		// bussy wait until hub is running
		for !hub.IsAlive() {
		}
		close(waitUntilActive)
	}()

	go hub.Run()

	// wait until active
	<-waitUntilActive
	// try to run the hub again
	// this should
	err := hub.Run()

	if err == nil {
		t.Error("hub is already running")
		return
	}

	hub.Shutdown(context.TODO())

	if hub.IsAlive() {
		t.Error("hub should not be active here")
	}
}
