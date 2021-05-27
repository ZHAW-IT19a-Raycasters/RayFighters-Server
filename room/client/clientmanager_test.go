package client

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.zhaw.ch/raycasters/server/log"
	"github.zhaw.ch/raycasters/server/room/testhelpers"
)

func helpTimeout() (HubTimeoutHandler, func(id uint8) ClientTimer) {

	// empty callback as it make no sense in this context
	callback := func(id uint8) {}

	ht := NewHubTimer(hubTimerSize, callback)
	f := func(id uint8) ClientTimer {
		return ht.NewClientTimer(id)
	}
	return ht, f
}

func getClientMeta(i int) (ClientMeta, *testhelpers.MockedConnection) {
	// as uuids should be unique, it makes no difference,
	// if a new one is called here.
	uuid := uuid.New()
	name := fmt.Sprintf("id: %d", i)
	mock := testhelpers.NewMockedConnection()
	pmock := &mock
	return NewClientMeta(uuid, name, pmock), pmock
}

func compareByteSlices(expeced, gotten []byte, t *testing.T) bool {
	if gotten == nil {
		t.Error("gotten bytes are nil")
		return false
	}

	if expeced == nil {
		t.Error("expected bytes are nil")
		return false
	}

	if len(expeced) != len(gotten) {
		t.Errorf("incorrect byte slice length exp: [%d], got: [%d]", len(expeced), len(gotten))
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

func TestClientManagerAddClientSingle(t *testing.T) {

	cm := NewClientManager()
	defer cm.Shutdown(context.Background())

	// make sure size starts at 0
	if cm.Size() != 0 {
		t.Error("client manager does not start at zero")
		return
	}

	clientMeta, _ := getClientMeta(0)

	clientQueue, _ := NewWorkQueue()

	_, getClientTimer := helpTimeout()

	err := cm.AddClient(clientMeta, clientQueue, getClientTimer)

	if err != nil {
		t.Errorf("An unexpected error occurred, while adding a new client\n[%s]", err.Error())
		return
	}

	// check if the client was added
	lenClients := cm.Size()
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
	} else if client.GetStatus() != Active {
		t.Error("Wrong status for client")
	} else if gottenStatus, _ := cm.GetClientStatus(iId); client.GetStatus() != gottenStatus {
		t.Error("Wrong status for client")
	} else if _, ok := cm.GetClientStatus(iId + 1); ok {
		t.Error("there should not be an id of the given type")
	}
}

func TestClientManagerAddClientTwice(t *testing.T) {
	cm := NewClientManager()
	defer cm.Shutdown(context.Background())

	clientMeta, pmock := getClientMeta(0)
	// done, funcReader, _ := setupMock(pmock)

	clientQueue, _ := NewWorkQueue()

	_, getClientTimer := helpTimeout()

	err := cm.AddClient(clientMeta, clientQueue, getClientTimer)

	if err != nil {
		t.Errorf("An unexpected error occurred, while adding a new client\n[%s]", err.Error())
		return
	}

	// force the client to be in a timeout state
	id, _ := cm.MapId(clientMeta)
	val := cm.clients[id]

	// shutdown the system
	pmock.FuncReader <- func() (int, io.Reader, error) {
		return 0, nil, &websocket.CloseError{
			Code: websocket.CloseAbnormalClosure,
		}
	}

	// check that client is shutdown
	<-pmock.Done
	waiter := make(chan struct{})

	go func() {
		for val.isActive.Get() {
		}
		close(waiter)
	}()

	<-waiter

	if val.isActive.Get() {
		t.Error("client is still active")
		return
	} else if val.GetStatus() == Active {
		t.Error("client is still active")
		return
	} else if val.GetStatus() != Timeout {
		t.Error("client should be timed out")
		return
	}

	// adding a connection, while client conn is still active
	mock := testhelpers.NewMockedConnection()
	pmock = &mock

	pmock.Close()

	clientMeta.conn = pmock

	clientQueue, _ = NewWorkQueue()

	err = cm.AddClient(clientMeta, clientQueue, getClientTimer)
	if err != nil {
		t.Errorf("An unexpected error occurred, while adding a new client\n[%s]", err.Error())
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
	cm := NewClientManager()
	defer cm.Shutdown(context.Background())

	// no errors here
	for i := 0; i <= math.MaxUint8; i += 1 {
		clientQueue, _ := NewWorkQueue()
		meta, _ := getClientMeta(i)
		_, getClientTimer := helpTimeout()

		err := cm.AddClient(meta, clientQueue, getClientTimer)

		if err != nil {
			t.Errorf("During fill up of client manager, an error happened:\n[%s]", err.Error())
			return
		}

		// make sure that the cm size is the same as the i counter
		if cm.Size() != (i + 1) {
			t.Errorf("There was a size difference at position [%d]", i)
			return
		}

	}
	// add an additional client
	meta, _ := getClientMeta(256)
	clientQueue, _ := NewWorkQueue()
	_, getClientTimer := helpTimeout()

	err := cm.AddClient(meta, clientQueue, getClientTimer)
	if err == nil {
		t.Error("A client was added illegally here.")
	}
}

func TestClientManagerRemoveClients(t *testing.T) {
	cm := NewClientManager()
	defer cm.Shutdown(context.Background())

	meta, _ := getClientMeta(0)

	err := cm.RemoveClient(meta)
	if err == nil {
		t.Errorf("there should not be a client in the given map.\n[%s]", err.Error())
		return
	} else if !strings.HasPrefix(err.Error(), "client doesn't exist") {
		t.Errorf("Incorrect prefix has been used:\n[%s]", err.Error())
		return
	}

	clientQueue, _ := NewWorkQueue()
	_, getClientTimer := helpTimeout()

	_ = cm.AddClient(meta, clientQueue, getClientTimer)

	if cm.Size() != 1 {
		t.Error("size is not at 1")
		return
	}

	err = cm.RemoveClient(meta)
	if err != nil {
		t.Error("there should not be an error during removing a known client.")
		return
	}

	if cm.Size() != 0 {
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
	buffer := new(bytes.Buffer)
	// setup logger
	builder := log.NewBuilder()
	builder.RegisterHandler(log.NewWriterHandler(log.Debug, buffer))

	err := builder.Build()
	if err != nil {
		t.Error("There is a test that did not clean up the global logger after it")
		return
	}
	t.Cleanup(func() {
		builder.Reset()
	})

	cm := NewClientManager()

	// no errors here
	for i := 0; i <= math.MaxUint8; i += 1 {
		// assume this works
		meta, _ := getClientMeta(i)
		clientQueue, _ := NewWorkQueue()
		_, getClientTimer := helpTimeout()

		err := cm.AddClient(meta, clientQueue, getClientTimer)

		if err != nil {
			t.Error("There was an error while adding the clients")
			return
		}
	}

	// shutdown all the "workers"
	cm.Shutdown(context.Background())

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

	// check buffer
	data := buffer.Bytes()
	strData := string(data)
	strData = strings.TrimSpace(strData)
	if !strings.HasSuffix(strData, "shutdown complete") {
		t.Error("incorrect suffix")
	}
}

func TestClientManagerIncorrectAddClient(t *testing.T) {

	// test for panic on logging in after lost connection
	// and incorrect removal of client data
	cm := NewClientManager()

	defer func() {
		r := recover()
		if r == nil {
			t.Error("there should have been a panic")
		}
		cm.Shutdown(context.Background())
	}()

	meta, _ := getClientMeta(0)
	clientQueue, _ := NewWorkQueue()
	_, getClientTimer := helpTimeout()

	_ = cm.AddClient(meta, clientQueue, getClientTimer)

	id, _ := cm.MapId(meta)
	client := cm.clients[id]
	// shutdown the client
	client.registerTimeout()

	iId, _ := cm.MapId(meta)
	// forcefully remove a client from the internal client map...

	delete(cm.clients, iId)
	// decrement the client counter as well
	// or else this will hang forever
	cm.wg.Done()

	_ = cm.AddClient(meta, clientQueue, getClientTimer)
}

func TestClientManagerNotExistingClientRemoved(t *testing.T) {
	// test for panic on logging in after lost connection
	// and incorrect removal of client data
	cm := NewClientManager()
	defer cm.Shutdown(context.TODO())

	meta, _ := getClientMeta(0)
	err := cm.RemoveClient(meta)
	if err == nil {
		t.Error("There should be an error here, as the client was never added")
	} else if !errors.Is(err, ErrClientDoesntExist{}) {
		t.Error("Wrong error type used")
	} else if !strings.Contains(err.Error(), "client doesn't exist") {
		t.Error("incorrect returned string used")
	}
}

func TestClientManagerIncorrectRemoveClientID(t *testing.T) {
	// test for panic on logging in after lost connection
	// and incorrect removal of client data
	cm := NewClientManager()
	defer func() {
		r := recover()
		if r == nil {
			t.Error("there should have been a panic")
		}
		cm.Shutdown(context.Background())
	}()

	meta, _ := getClientMeta(0)
	cq, _ := NewWorkQueue()
	_, getClientTimer := helpTimeout()

	_ = cm.AddClient(meta, cq, getClientTimer)

	id, _ := cm.MapId(meta)
	client := cm.clients[id]
	// shutdown the client
	client.registerTimeout()

	// forcefully push a value back onto the queue
	err := cm.idManager.Push(0)
	if err != nil {
		t.Errorf("There was an unexpected error on pushing onto the queue:\n%s", err.Error())
		return
	}

	_ = cm.RemoveClient(meta)
}

func TestClientManagerMapping(t *testing.T) {

	cm := NewClientManager()

	meta, _ := getClientMeta(0)

	cq, _ := NewWorkQueue()
	_, getClientTimer := helpTimeout()

	_ = cm.AddClient(meta, cq, getClientTimer)

	defer cm.Shutdown(context.TODO())

	// get the id
	id, ok := cm.MapId(meta)

	if !ok {
		t.Error("the meta should be represented in the client manager")
		return
	}

	cimeta, ok := cm.MapMeta(id)

	if !ok {
		t.Error("the id should be represented in the client manager")
		return
	}

	if cimeta.GetId() != meta.GetId() {
		t.Error("incorrect external id was returned")
	}

	cm.RemoveClient(meta)

	_, ok = cm.MapMeta(id)

	if ok {
		t.Error("there should be no valid mapping for the given client")
	}
}

func TestClientManagerGetPlayer(t *testing.T) {
	cm := NewClientManager()

	meta, _ := getClientMeta(0)

	cq, _ := NewWorkQueue()

	_, getClientTimer := helpTimeout()

	_ = cm.AddClient(meta, cq, getClientTimer)

	id, _ := cm.MapId(meta)

	defer cm.Shutdown(context.TODO())

	player, ok := cm.GetPlayer(id)

	if !ok {
		t.Error("player should exist")
		return
	}

	if player.Name() != meta.GetName() {
		t.Error("Incorrect name used for player")
	} else if player.UserId() != id {
		t.Error("incorrect id was used for player")
	}

	cm.RemoveClient(meta)
	_, ok = cm.GetPlayer(id)
	if ok {
		t.Error("there should be no player by that id")
	}

}

func TestClientManagerGetPlayers(t *testing.T) {
	cm := NewClientManager()

	meta, _ := getClientMeta(0)

	cq, _ := NewWorkQueue()

	_, getClientTimer := helpTimeout()

	_ = cm.AddClient(meta, cq, getClientTimer)

	id, _ := cm.MapId(meta)

	defer cm.Shutdown(context.TODO())

	pls := cm.GetPlayers()

	if len(pls) != 1 {
		t.Error("wrong slice lenght for players")
		return
	}

	player := pls[0]
	if player.Name() != meta.GetName() {
		t.Error("Incorrect name used for player")
	} else if player.UserId() != id {
		t.Error("incorrect id was used for player")
	}

}
