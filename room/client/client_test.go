package client

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/gorilla/websocket"
	pclient "github.zhaw.ch/raycasters/server/protocol/client"
	"github.zhaw.ch/raycasters/server/protocol/server"
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

func TestClientRecv(t *testing.T) {
	cq, hq := helpers.NewWorkQueue()
	cmeta, mock := getClientMeta(1)

	// make the connections blocking, so that we can run other
	// stuff while it waits
	funcTransferReader := make(chan func() (int, io.Reader, error), 10)
	funcTransferWriter := make(chan func(messageType int) (io.WriteCloser, error), 10)

	mock.funcNextReader = func() (int, io.Reader, error) {
		return (<-funcTransferReader)()
	}

	mock.funcNextWriter = func(messageType int) (io.WriteCloser, error) {
		return (<-funcTransferWriter)(messageType)
	}

	mock.funcCloser = func() error {
		// shutdown the resv as it is the only part of the system that may
		// do that.
		funcTransferReader <- func() (int, io.Reader, error) {
			return 0, nil, websocket.ErrCloseSent
		}

		// shutdown the client send, this has to be send first
		// as to hinder any race conditions.
		funcTransferWriter <- func(messageType int) (io.WriteCloser, error) {
			return nil, websocket.ErrCloseSent
		}
		return nil
	}

	client := NewClient(1, cmeta, cq)
	defer client.Shutdown(context.TODO())

	client.Run()

	if !client.isActive.Get() {
		t.Error("client should be running")
		return
	}

	// fill up the channel
	buffer := []byte{pclient.DisconnectOpcode.Value()}

	funcTransferReader <- func() (int, io.Reader, error) {
		return 0, bytes.NewReader(buffer), nil
	}

	data, err := hq.Dequeue()
	if err != nil {
		t.Error("There should not be an error here, as the channel should be opened")
		return
	}

	if !compareByteSlices(buffer, data, t) {
		return
	}

	// shutdown the client receive, this has to be send first
	// as to hinder any race conditions.
	funcTransferReader <- func() (int, io.Reader, error) {
		return 0, nil, websocket.ErrCloseSent
	}

	_, err = hq.Dequeue()

	if err == nil {
		t.Error("The channel should be closed here")
	}
}

func TestClientSend(t *testing.T) {
	cq, hq := helpers.NewWorkQueue()
	cmeta, mock := getClientMeta(1)

	// make the connections blocking, so that we can run other
	// stuff while it waits
	funcTransferReader := make(chan func() (int, io.Reader, error), 10)
	funcTransferWriter := make(chan func(messageType int) (io.WriteCloser, error), 10)

	mock.funcNextReader = func() (int, io.Reader, error) {
		t.Log("Reading from the read queue")
		return (<-funcTransferReader)()
	}

	mock.funcNextWriter = func(messageType int) (io.WriteCloser, error) {
		t.Log("Reading from the write queue")
		return (<-funcTransferWriter)(messageType)
	}

	mock.funcCloser = func() error {
		// shutdown the resv as it is the only part of the system that may
		// do that.
		funcTransferReader <- func() (int, io.Reader, error) {
			t.Log("read")
			return 0, nil, websocket.ErrCloseSent
		}

		// shutdown the client send, this has to be send first
		// as to hinder any race conditions.
		funcTransferWriter <- func(messageType int) (io.WriteCloser, error) {
			t.Log("wrote")
			return nil, websocket.ErrCloseSent
		}
		return nil
	}

	// send this first as to not have a race condition with
	// the enqueue
	funcTransferWriter <- func(messageType int) (io.WriteCloser, error) {
		if messageType != websocket.BinaryMessage {
			t.Errorf("Wrong message type used [%d] should be [%d]", messageType, websocket.BinaryMessage)
		}
		return mock.mwc, nil
	}

	client := NewClient(1, cmeta, cq)
	defer client.Shutdown(context.TODO())

	client.Run()

	// send the information to the server
	exp := []byte{server.DisconnectOpcode.Value()}
	err := hq.Enqueue(exp)

	if err != nil {
		t.Error("the queue should be open and receiving here")
		return
	}

	gotten := mock.mwc.Bytes()

	compareByteSlices(exp, gotten, t)
}
