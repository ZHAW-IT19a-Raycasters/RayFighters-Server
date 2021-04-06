package room

import (
	"context"
	"fmt"
	"io"
	"math"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.zhaw.ch/raycasters/server/protocol/server"
)

func getClientMeta(i int) ClientMeta {
	// as uuids should be unique, it makes no difference,
	// if a new one is called here.
	uuid := uuid.New()
	name := fmt.Sprintf("id: %d", i)
	mock := newMockedConnection()
	return NewClientMeta(uuid, name, &mock)
}

func compareByteSlices(expeced, gotten []byte, t *testing.T) bool {
	if gotten == nil {
		t.Error("gotten bytes are nil")
		return false
	}

	if expeced == nil {
		t.Error("exptected bytes are nil")
		return false
	}

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

	// force the client to be in a timeout state
	id, _ := cm.mapId(clientMeta)
	val := cm.clients[id]
	val.Shutdown(context.Background())

	// adding a connection, while client conn is still active
	mock := newMockedConnection()
	clientMeta.conn = &mock

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

func TestClientManagerIncorrectAddClient(t *testing.T) {
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

func TestClientManagerIncorrectRemoveClient(t *testing.T) {
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

func TestClientManagerBroadCast(t *testing.T) {

	// add test clients
	const ClientCount = 10
	clients := make([]ClientMeta, ClientCount)
	mocks := make([]*MockedConnection, ClientCount)

	hubMock := newMockedHubClientInteraction()
	manager := newClientManager(&hubMock)

	for i := 0; i < ClientCount; i += 1 {
		clients[i] = getClientMeta(i)
		tmp := clients[i].conn.(*MockedConnection)
		mocks[i] = tmp
		// change the func in the mock to write into the mocked reader
		mocks[i].funcNextWriter = func(messageType int) (io.WriteCloser, error) {
			return tmp.mwc, nil
		}

		err := manager.addClient(clients[i])
		if err != nil {
			t.Errorf("error happend, while adding a new client.\n%s", err.Error())
			return
		}
	}

	manager.shutdown(context.Background())

	for _, mock := range mocks {
		mock.mwc.Reset()
	}

	// broadCast this to all clients
	// this is the case where all the clients are offline
	opcode := server.NewDisconnect()
	exp := opcode.Bytes()
	manager.broadCast(opcode)

	// test that the broadcast was not "send"
	// as all the clients are fully deactivated
	for _, mock := range mocks {
		buff := mock.mwc.Bytes()
		if len(buff) != 0 {
			t.Error("should be empty")
			return
		}
		mock.mwc.Reset()
	}

	// reactivate the clients for the test
	for _, client := range manager.clients {
		client.status = Active
	}

	manager.broadCast(opcode)

	// test that the broadcast was not "send"
	// as all the clients are not "isActive"
	for _, mock := range mocks {
		buff := mock.mwc.Bytes()
		if len(buff) != 0 {
			t.Error("should be empty")
			return
		}
		mock.mwc.Reset()
	}

	for _, client := range manager.clients {
		if !client.isActive.SetTrue() {
			t.Errorf("could not change the state of the isActive bool for [%d]", client.meta.id)
			return
		}
	}

	manager.broadCast(opcode)

	// test that the broadcast was "send"
	for _, mock := range mocks {
		buff := mock.mwc.Bytes()
		if !compareByteSlices(exp, buff, t) {
			return
		}
	}
}
