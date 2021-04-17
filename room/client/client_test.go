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

type blockChan chan struct{}

type chanFuncReader chan func() (int, io.Reader, error)

type chanFuncWriter chan func(messageType int) (io.WriteCloser, error)

func setupMock(mock *MockedConnection) (blockChan, chanFuncReader, chanFuncWriter) {

	done := make(blockChan)

	// make the connections blocking, so that we can run other
	// stuff while it waits
	funcTransferReader := make(chanFuncReader, 10)
	funcTransferWriter := make(chanFuncWriter, 10)

	mock.mwc.cl = func() error {
		close(done)
		return nil
	}

	mock.funcNextReader = func() (int, io.Reader, error) {
		select {
		case <-done:
			return 0, nil, websocket.ErrCloseSent
		case d := <-funcTransferReader:
			return d()
		}
	}

	mock.funcNextWriter = func(messageType int) (io.WriteCloser, error) {
		select {
		case <-done:
			return nil, websocket.ErrCloseSent
		case d := <-funcTransferWriter:
			return d(messageType)
		}
	}

	// override the closer, so that the channels don't get changed by it
	mock.funcCloser = func() error {
		select {
		case <-done:
		default:
			close(done)
		}
		return nil
	}
	return done, funcTransferReader, funcTransferWriter
}

func TestClientregisterConn(t *testing.T) {
	clientQueue, _ := helpers.NewWorkQueue()
	meta, _ := getClientMeta(42)
	client := NewClient(42, meta, clientQueue)
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
	cq, hq := helpers.NewWorkQueue()
	cmeta, mock := getClientMeta(1)

	_, funcTransferReader, _ := setupMock(mock)

	client := NewClient(1, cmeta, cq)
	defer client.Shutdown(context.TODO())

	client.Run()

	if !client.isActive.Get() {
		t.Error("client should be running")
		return
	}

	// fill up the channel
	buffer := []byte{pclient.RequestOtherPlayersOpcode.Value()}

	funcTransferReader <- func() (int, io.Reader, error) {
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

	done, _, funcTransferWriter := setupMock(mock)

	// send this first as to not have a race condition with
	// the enqueue
	funcTransferWriter <- func(messageType int) (io.WriteCloser, error) {
		// shutdown the client
		if messageType != websocket.BinaryMessage {
			t.Errorf("Wrong message type used [%d] should be [%d]", messageType, websocket.BinaryMessage)
		}
		return mock.mwc, nil
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

	client := NewClient(1, cmeta, cq)
	defer client.Shutdown(context.TODO())

	client.Run()

	<-done

	// wait until the client has shutdown
	// completely
	client.Shutdown(context.TODO())

	gotten := mock.mwc.Bytes()

	compareByteSlices(exp, gotten, t)
}
func TestClientSendShutdown(t *testing.T) {
	cq, hq := helpers.NewWorkQueue()
	cmeta, mock := getClientMeta(1)

	done, _, funcTransferWriter := setupMock(mock)

	// send this first as to not have a race condition with
	// the enqueue
	funcTransferWriter <- func(messageType int) (io.WriteCloser, error) {
		// shutdown the client
		if messageType != websocket.CloseMessage {
			t.Errorf("Wrong message type used [%d] should be [%d]", messageType, websocket.BinaryMessage)
		}
		return mock.mwc, nil
	}

	exp := []byte{server.DisconnectOpcode.Value()}
	err := hq.Enqueue(exp)

	if err != nil {
		t.Error("the queue should be open and receiving here")
		return
	}

	client := NewClient(1, cmeta, cq)
	defer client.Shutdown(context.TODO())

	client.Run()

	<-done

	// wait until the client has shutdown
	// completely
	client.Shutdown(context.TODO())

	gotten := mock.mwc.Bytes()

	compareByteSlices(exp, gotten, t)
}

func TestClientSendNextWriterError(t *testing.T) {
	cq, hq := helpers.NewWorkQueue()
	cmeta, mock := getClientMeta(1)

	done, _, _ := setupMock(mock)

	exp := []byte{server.PlayerJoinedOpcode.Value()}
	err := hq.Enqueue(exp)

	if err != nil {
		t.Error("the queue should be open and receiving here")
		return
	}

	client := NewClient(1, cmeta, cq)
	defer client.Shutdown(context.TODO())

	client.Run()

	close(done)

	// wait until the client has shutdown
	// completely
	client.Shutdown(context.TODO())
}
