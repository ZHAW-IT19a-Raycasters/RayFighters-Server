package client

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/gorilla/websocket"
	pclient "github.zhaw.ch/raycasters/server/protocol/client"
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

func TestClientResv(t *testing.T) {
	cq, hq := helpers.NewWorkQueue()
	cmeta, mock := getClientMeta(1)

	funcTransfer := make(chan func() (int, io.Reader, error))

	mock.funcNextReader = func() (int, io.Reader, error) {
		return (<-funcTransfer)()
	}

	client := NewClient(1, cmeta, cq)
	defer client.Shutdown(context.TODO())
	client.Run()

	// fill up the channel
	buffer := []byte{pclient.DisconnectOpcode.Value()}
	funcTransfer <- func() (int, io.Reader, error) {
		return 0, bytes.NewReader(buffer), nil
	}
	funcTransfer <- func() (int, io.Reader, error) {
		return 0, nil, websocket.ErrCloseSent
	}

	data, _ := hq.Dequeue()
	compareByteSlices(buffer, data, t)
}
