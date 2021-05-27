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
)

func TestClientRegisterConn(t *testing.T) {
	clientQueue, _ := NewWorkQueue()
	const id = 42
	meta, _ := getClientMeta(id)
	_, getClientTimer := helpTimeout()
	cT := getClientTimer(id)

	client := NewClient(id, meta, clientQueue, cT)
	defer client.Shutdown(context.TODO())

	// test get meta
	if meta.id != client.meta.GetMeta().id {
		t.Error("wrong id returned")
		return
	}

	test := func() bool {
		err := client.registerConn(meta, clientQueue)

		if err == nil {
			t.Error("there should be an error here")
			return false
		} else if !errors.Is(err, ErrClientWorkersAlreadyActive) {
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

	// this is primarily for coverage
	// check on the internal meta
	imeta := client.GetInternalMeta()

	if imeta.GetUUID() != meta.GetId() {
		t.Error("Incorrect id was used")
	} else if imeta.GetName() != meta.GetName() {
		t.Error("Incorrect name was returned")
	}
}

func TestClientRecv(t *testing.T) {
	const ID = 1
	cq, hq := NewWorkQueue()
	cmeta, mock := getClientMeta(ID)

	_, getClientTimer := helpTimeout()
	cT := getClientTimer(ID)
	client := NewClient(1, cmeta, cq, cT)
	defer client.Shutdown(context.TODO())

	client.Run()

	if !client.isActive.Get() {
		t.Error("client should be running")
		return
	}

	// fill up the channel
	buffer := []byte{pclient.RequestOtherPlayersOpcode.Value()}

	mock.FuncReader <- func() (int, io.Reader, error) {
		return websocket.BinaryMessage, bytes.NewReader(buffer), nil
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
	mock.FuncReader <- func() (int, io.Reader, error) {
		return 0, nil, websocket.ErrCloseSent
	}

	_, err = hq.Dequeue()

	if err == nil {
		t.Error("The channel should be closed here")
	}
}

func TestClientSend(t *testing.T) {
	const ID = 1
	cq, hq := NewWorkQueue()
	cmeta, mock := getClientMeta(ID)
	_, getClientTimer := helpTimeout()
	cT := getClientTimer(ID)

	// send this first as to not have a race condition with
	// the enqueue
	mock.FuncWriter <- func(messageType int) (io.WriteCloser, error) {
		// shutdown the client
		if messageType != websocket.BinaryMessage {
			t.Errorf("Wrong message type used [%d] should be [%d]", messageType, websocket.BinaryMessage)
		}
		mock.Close()
		return mock.Mwc, nil
	}

	// send the information to the server
	// opcode type is irrelevant with the
	// exception of the disconnect, as that
	// needs a different frame to send the
	// information to
	exp := []byte{server.PlayerDeathOpcode.Value()}
	err := hq.Enqueue(exp)

	if err != nil {
		t.Error("the queue should be open and receiving here")
		return
	}

	client := NewClient(ID, cmeta, cq, cT)

	client.Run()

	defer client.Shutdown(context.TODO())
	<-mock.Done

	gotten := mock.Mwc.Bytes()

	compareByteSlices(exp, gotten, t)
}

func TestClientSendShutdown(t *testing.T) {
	const ID = 1
	cq, hq := NewWorkQueue()
	cmeta, mock := getClientMeta(ID)
	_, getClientTimer := helpTimeout()
	cT := getClientTimer(ID)

	// send this first as to not have a race condition with
	// the enqueue
	mock.FuncWriter <- func(messageType int) (io.WriteCloser, error) {
		// shutdown the client
		if messageType != websocket.CloseMessage {
			t.Errorf("Wrong message type used [%d] should be [%d]", messageType, websocket.BinaryMessage)
		}
		return mock.Mwc, nil
	}

	exp := []byte{server.DisconnectOpcode.Value()}
	err := hq.Enqueue(exp)

	if err != nil {
		t.Error("the queue should be open and receiving here")
		return
	}

	client := NewClient(ID, cmeta, cq, cT)

	client.Run()

	<-mock.Done

	// wait until the client has shutdown
	// completely
	client.Shutdown(context.TODO())

	gotten := mock.Mwc.Bytes()

	compareByteSlices(exp, gotten, t)
}

func TestClientSendNextWriterError(t *testing.T) {
	const ID = 1
	cq, hq := NewWorkQueue()
	cmeta, mock := getClientMeta(ID)

	_, getClientTimer := helpTimeout()
	cT := getClientTimer(ID)

	exp := []byte{server.PlayerJoinedOpcode.Value()}
	err := hq.Enqueue(exp)

	if err != nil {
		t.Error("the queue should be open and receiving here")
		return
	}

	client := NewClient(ID, cmeta, cq, cT)
	defer client.Shutdown(context.TODO())

	client.Run()

	mock.Close()

	// wait until the client has shutdown
	// completely
	client.Shutdown(context.TODO())
}
