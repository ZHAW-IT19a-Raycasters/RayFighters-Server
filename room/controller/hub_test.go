package controller

import (
	"context"
	"github.com/gorilla/websocket"
	"io"
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

	meta, mock := getClientMeta(1)
	err = hub.RegisterClient(meta)

	if err != nil {
		t.Error("RegisterClient returned an error")
	}

	if hub.Size() != 1 {
		t.Error("Hub size should be one here")
	}

	if hub.Full() {
		t.Error("Hub shouldn't be full at this point")
	}

	if !hub.IsAlive() {
		t.Error("hub should be active here")
		return
	}

	mock.FuncReader <- func() (int, io.Reader, error) {
		return 0, nil, &websocket.CloseError{
			websocket.CloseGoingAway,
			"I fly away",
		}
	}

	<- mock.Done

	hub.Shutdown(context.TODO())

	if hub.IsAlive() {
		t.Error("hub should not be active here")
	}
}
