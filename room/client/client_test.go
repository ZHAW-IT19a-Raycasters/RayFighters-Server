package client

import (
	"context"
	"errors"
	"testing"

	"github.zhaw.ch/raycasters/server/room/helpers"
)

func TestClientregisterConn(t *testing.T) {
	clientQueue, _ := helpers.NewWorkQueue()
	meta, _ := getClientMeta(42)
	client := NewClient(42, meta, clientQueue)
	defer client.Shutdown(context.TODO())

	// test get meta
	if meta.id != client.meta.getMeta().id {
		t.Error("wrong id returned")
		return
	}

	test := func() bool {
		err := client.registerConn(meta, clientQueue)

		if err == nil {
			t.Error("there should be an error here")
			return false
		} else if !errors.Is(err, ErrClientReaderAlreadyActive) {
			t.Error("wrong error returned")
			return false
		}
		return true
	}

	if !test() {
		return
	}

	// force state to be active
	client.isActive.SetTrue()

	test()

	// check the set true
	err := client.Run()
	if err == nil {
		t.Error("Should be an error here")
		return
	}

	client.Shutdown(context.TODO())
}
