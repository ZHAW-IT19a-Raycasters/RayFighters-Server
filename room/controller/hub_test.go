package controller

import (
	"context"
	"testing"

	"github.zhaw.ch/raycasters/server/room/testhelpers"
)

func TestHub(t *testing.T) {
	manager := testhelpers.SetupMapIterator(t)
	const hubId = 1
	hub := NewHub(hubId, manager.GetIterator())

	waitUntilActive := make(chan struct{})
	go func() {
		// bussy wait until hub is running
		for !hub.IsAlive() {
		}
		close(waitUntilActive)
	}()

	hub.Run()

	// wait until active
	<-waitUntilActive

	// try to run the hub again
	// this should
	err := hub.Run()

	if err == nil {
		t.Error("hub is already running")
		return
	}

	if !hub.IsAlive() {
		t.Error("hub should be active here")
		return
	}

	hub.Shutdown(context.TODO())

	if hub.IsAlive() {
		t.Error("hub should not be active here")
	}
}
