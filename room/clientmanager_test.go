package room

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"math"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.zhaw.ch/raycasters/server/protocol/server"
)

type MockedHubClientInteraction struct {
	funcAllPlayers       func() []server.Player
	funcPlayerGotHit     func() (bool, server.Player)
	funcPlayerKD         func() []server.PlayerKD
	funcUnregisterClient func(meta ClientMeta) error
}

func (m *MockedHubClientInteraction) GetAllPlayers() []server.Player {
	return m.funcAllPlayers()
}

func (m *MockedHubClientInteraction) PlayerGotHit() (bool, server.Player) {
	return m.funcPlayerGotHit()
}

func (m *MockedHubClientInteraction) PlayersKD() []server.PlayerKD {
	return m.funcPlayerKD()
}

func (m *MockedHubClientInteraction) UnregisterClient(meta ClientMeta) error {
	return m.funcUnregisterClient(meta)
}

func newMockedHubClientInteraction() MockedHubClientInteraction {
	return MockedHubClientInteraction{
		funcAllPlayers: func() []server.Player {
			return []server.Player{}
		},
		funcPlayerGotHit: func() (bool, server.Player) {
			return false, server.NewPlayer(0, "")
		},
		funcPlayerKD: func() []server.PlayerKD {
			return []server.PlayerKD{}
		},
		funcUnregisterClient: func(meta ClientMeta) error {
			return nil
		},
	}
}

type Mockable struct {
	mr             *MockedReader
	mwc            *MockedWriteCloser
	funcNextReader func() (int, io.Reader, error)
	funcNextWriter func(messageType int) (io.WriteCloser, error)
	funcCloser     func() error
}

func (m *Mockable) NextReader() (int, io.Reader, error) {
	return m.funcNextReader()
}

func (m *Mockable) NextWriter(messageType int) (io.WriteCloser, error) {
	return m.funcNextWriter(messageType)
}

func (m *Mockable) Close() error {
	// replace the next func reader and next writer with error returning versions
	m.funcNextReader = func() (int, io.Reader, error) {
		return 0, nil, websocket.ErrCloseSent
	}
	m.funcNextWriter = func(messageType int) (io.WriteCloser, error) {
		return nil, websocket.ErrCloseSent
	}
	return m.funcCloser()
}

func newMockable() Mockable {
	mr := newMockedReader()
	mwc := newMockedWriteCloser()
	pmr := &mr
	pmwc := &mwc
	return Mockable{
		mr:  pmr,
		mwc: pmwc,
		funcNextReader: func() (int, io.Reader, error) {
			return 0, nil, websocket.ErrCloseSent
		},
		funcNextWriter: func(messageType int) (io.WriteCloser, error) {
			return nil, websocket.ErrCloseSent
		},
		funcCloser: func() error {
			return nil
		},
	}
}

type MockedWriteCloser struct {
	buffer *bytes.Buffer
	cl     func() error
	*bufio.Writer
}

func newMockedWriteCloser() MockedWriteCloser {
	bff := new(bytes.Buffer)
	cl := func() error { return nil }
	bw := bufio.NewWriter(bff)
	return MockedWriteCloser{
		bff,
		cl,
		bw,
	}
}

func (mwc *MockedWriteCloser) Close() error {
	return mwc.cl()
}

type MockedReader struct {
	*bytes.Buffer
}

func newMockedReader() MockedReader {
	return MockedReader{
		new(bytes.Buffer),
	}
}

func getClientMeta(i int) ClientMeta {
	// as uuids should be unique, it makes no difference,
	// if a new one is called here.
	uuid := uuid.New()
	name := fmt.Sprintf("id: %d", i)
	mock := newMockable()
	return NewClientMeta(uuid, name, &mock)
}

func compareByteSlices(expeced, gotten []byte, t *testing.T) bool {
	if len(expeced) != len(gotten) {
		t.Errorf("incorrect byte slice lenght exp: [%d], got: [%d]", len(expeced), len(gotten))
		return false
	}

	for i := range expeced {
		exp := expeced[i]
		got := gotten[i]

		if exp != got {
			t.Errorf("Byte at position [%d] are not equal exp: [0x%X] - got: [0x%X]", i, exp, got)
			return false
		}
	}

	return true
}

func TestClientManagerAddClientSigle(t *testing.T) {
	hubMock := newMockedHubClientInteraction()
	cm := newClientManager(&hubMock)
	defer cm.shutdown(context.Background())

	// make sure size starts at 0
	if cm.size() != 0 {
		t.Error("client manager does not start at zero")
		return
	}

	clientMeta := getClientMeta(0)
	err := cm.addClient(clientMeta)
	if err != nil {
		t.Errorf("An unexpected error occuered, while adding a new client\n[%s]", err.Error())
		return
	}

	// check if the client was added
	lenClients := len(cm.clients)
	lenMapping := len(cm.mapping)
	if lenClients != 1 {
		t.Errorf("client size is not equals 1 -> [%d]", lenClients)
		return
	} else if lenClients != lenMapping {
		t.Errorf("map sizes are different client [%d] mapping [%d]", lenClients, lenMapping)
		return
	}

	iId, ok := cm.mapping[clientMeta.id]
	if !ok {
		t.Error("Client has not been mapped correctly, missing entry in id mapping.")
		return
	}

	client, ok := cm.clients[iId]

	if !ok {
		t.Error("Client has not been mapped correctly, missing entry in clients")
		return
	}

	if client.meta.eId != clientMeta.id {
		t.Error("Wrong uuid has been used, while mapping.")
	} else if client.meta.name != clientMeta.name {
		t.Error("Wrong name has been used, while mapping.")
	}
}

func TestClientManagerAddClientTwice(t *testing.T) {
	hubMock := newMockedHubClientInteraction()
	cm := newClientManager(&hubMock)
	defer cm.shutdown(context.Background())

	clientMeta := getClientMeta(0)

	err := cm.addClient(clientMeta)
	if err != nil {
		t.Errorf("An unexpected error occuered, while adding a new client\n[%s]", err.Error())
		return
	}

	// adding a connection, while client conn is still active
	mock := newMockable()
	clientMeta.conn = &mock

	err = cm.addClient(clientMeta)
	if err != nil {
		t.Error("no error expected here, as the client is timedout here")
		return
	}

	// force the client to be in a timeout state
	id, _ := cm.mapId(clientMeta)
	val := cm.clients[id]
	val.registerTimeout()

	err = cm.addClient(clientMeta)
	if err != nil {
		t.Errorf("An unexpected error occuered, while adding a new client\n[%s]", err.Error())
		return
	}

	// check if the client was added
	lenClients := len(cm.clients)
	lenMapping := len(cm.mapping)
	if lenClients != 1 {
		t.Errorf("client size is not equals 1 -> [%d]", lenClients)
		return
	} else if lenClients != lenMapping {
		t.Errorf("map sizes are different client [%d] mapping [%d]", lenClients, lenMapping)
		return
	}
}

func TestClientManagerMaxClients(t *testing.T) {
	hubMock := newMockedHubClientInteraction()
	cm := newClientManager(&hubMock)
	defer cm.shutdown(context.Background())

	// no errors here
	for i := 0; i <= math.MaxUint8; i += 1 {
		meta := getClientMeta(i)
		err := cm.addClient(meta)

		if err != nil {
			t.Errorf("Durring fill up of client manager, an error happend:\n[%s]", err.Error())
			return
		}

		// make sure that the cm size is the same as the i counter
		if cm.size() != (i + 1) {
			t.Errorf("There was a size difference at position [%d]", i)
			return
		}

	}
	// add an additional client
	meta := getClientMeta(256)
	err := cm.addClient(meta)
	if err == nil {
		t.Error("A client was added illigaly here.")
	}
}

func TestClientManagerRemoveClients(t *testing.T) {
	hubMock := newMockedHubClientInteraction()
	cm := newClientManager(&hubMock)
	defer cm.shutdown(context.Background())

	meta := getClientMeta(0)

	err := cm.removeClient(meta)
	if err == nil {
		t.Errorf("there should not be a client in the given map.\n[%s]", err.Error())
		return
	} else if !strings.HasPrefix(err.Error(), "client doesn't exist") {
		t.Errorf("Incorrect prefix has been used:\n[%s]", err.Error())
		return
	}

	cm.addClient(meta)

	if cm.size() != 1 {
		t.Error("size is not at 1")
		return
	}

	err = cm.removeClient(meta)
	if err != nil {
		t.Error("there should not be an error during removing a known client.")
		return
	}

	if cm.size() != 0 {
		t.Error("size is not at 0")
		return
	}

	// check that all the mappings return to normal

	if len(cm.mapping) != 0 {
		t.Error("mapping size is not at 0")
		return
	}

	if len(cm.clients) != 0 {
		t.Error("clients size is not at 0")
		return
	}

	if !cm.idManager.IsFull() {
		t.Error("id has not been returned to queue")
		return
	}
}

func TestClientManagerShutdown(t *testing.T) {
	hubMock := newMockedHubClientInteraction()
	cm := newClientManager(&hubMock)

	// no errors here
	for i := 0; i <= math.MaxUint8; i += 1 {
		// assume this works
		meta := getClientMeta(i)
		err := cm.addClient(meta)
		if err != nil {
			t.Error("There was an error while adding the clients")
			return
		}
	}

	// shutdown all the "workers"
	cm.shutdown(context.Background())

	// check that all the worker state are on shutdown
	for _, client := range cm.clients {
		// check the state of the client atomic bool
		if client.isActive.Get() {
			t.Errorf("client [%d] is still active", client.meta.id)
		}
		if client.status != Offline {
			t.Errorf("incorrect state has been used [%s]", client.status)
			return
		}
	}
	// TODO: test logger output - this needs code that
	// currently is not merged into this branch
}

func TestClientManagerForceShutdown(t *testing.T) {
	hubMock := newMockedHubClientInteraction()
	cm := newClientManager(&hubMock)

	// no errors here
	for i := 0; i <= math.MaxUint8; i += 1 {
		// assume this works
		meta := getClientMeta(i)
		_ = cm.addClient(meta)

	}

	// shutdown all the "workers"
	cxt, cancel := context.WithCancel(context.Background())
	cc := make(chan struct{})
	go func() {
		<-cc
		cm.shutdown(cxt)
	}()
	// direct shutdown
	cancel()
	close(cc)
	// TODO: test logger output - this needs code that
	// currently is not merged into this branch
}

func TestIncorrectAddClient(t *testing.T) {
	// TODO: fix this tests it hangs for ever

	// test for panic on logging in after lost connection
	// and incorrect removal of client data
	hubMock := newMockedHubClientInteraction()
	cm := newClientManager(&hubMock)

	defer func() {
		r := recover()
		if r == nil {
			t.Error("there should have been a panic")
		}
		cm.shutdown(context.Background())
	}()

	meta := getClientMeta(0)
	_ = cm.addClient(meta)

	id, _ := cm.mapId(meta)
	client := cm.clients[id]
	// shutdown the client
	client.registerTimeout()

	iId, _ := cm.mapId(meta)
	// forcefully remove a client from the internal client map...

	delete(cm.clients, iId)
	// decrement the client counter as well
	// or else this will hang forever
	cm.wg.Done()

	_ = cm.addClient(meta)
}

func TestIncorrectRemoveClient(t *testing.T) {
	// test for panic on logging in after lost connection
	// and incorrect removal of client data
	hubMock := newMockedHubClientInteraction()
	cm := newClientManager(&hubMock)
	defer func() {
		r := recover()
		if r == nil {
			t.Error("there should have been a panic")
		}
		cm.shutdown(context.Background())
	}()

	meta := getClientMeta(0)
	_ = cm.addClient(meta)

	id, _ := cm.mapId(meta)
	client := cm.clients[id]
	// shutdown the client
	client.registerTimeout()

	// forcefully push a value back onto the queue
	err := cm.idManager.Push(0)
	if err != nil {
		t.Errorf("There was an unexpected error on pushing onto the queue:\n%s", err.Error())
		return
	}

	_ = cm.removeClient(meta)
}

func TestBroadCast(t *testing.T) {

	// add test clients
	const ClientCount = 10
	clients := make([]ClientMeta, ClientCount)
	mocks := make([]*Mockable, ClientCount)

	hubMock := newMockedHubClientInteraction()
	manager := newClientManager(&hubMock)
	defer manager.shutdown(context.Background())

	for i := 0; i < ClientCount; i += 1 {
		clients[i] = getClientMeta(i)
		mocks[i] = clients[i].conn.(*Mockable)
		err := manager.addClient(clients[i])
		if err != nil {
			t.Errorf("error happend, while adding a new client.\n%s", err.Error())
			return
		}

		// do not start the clients
	}

	// broadCast this to all clients
	opcode := server.NewDisconnect()
	exp := opcode.Bytes()
	manager.broadCast(opcode, []byte{})

	// test that the broadcast was "send"
	for _, mock := range mocks {
		buff := mock.mwc.buffer.Bytes()
		if !compareByteSlices(exp, buff, t) {
			return
		}
	}
}
