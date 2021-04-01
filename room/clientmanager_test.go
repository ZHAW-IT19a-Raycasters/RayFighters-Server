package room

import (
	"context"
	"fmt"
	"io"
	"math"
	"strings"
	"testing"

	"github.com/google/uuid"
)

type Mockable struct {
	funcNextReader func() (int, io.Reader, error)
	funcNextWriter func(messageType int) (io.WriteCloser, error)
}

func (m *Mockable) NextReader() (int, io.Reader, error) {
	return m.funcNextReader()
}

func (m *Mockable) NextWriter(messageType int) (io.WriteCloser, error) {
	return m.funcNextWriter(messageType)
}

func newMockable() Mockable {
	return Mockable{
		funcNextReader: func() (int, io.Reader, error) {

			return 0, nil, nil
		},
		funcNextWriter: func(messageType int) (io.WriteCloser, error) {

			return nil, nil
		},
	}
}

func getClientMeta() ClientMeta {
	uuid := uuid.New()
	name := "Hans"
	// Prepare the mocked data
	mock := newMockable()

	return NewClientMeta(uuid, name, &mock)
}

func TestClientManagerAddClientSigle(t *testing.T) {
	cm := newClientManager()

	// make sure size starts at 0
	if cm.size() != 0 {
		t.Error("client manager does not start at zero")
		return
	}

	clientMeta := getClientMeta()
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
	cm := newClientManager()

	clientMeta := getClientMeta()
	err := cm.addClient(clientMeta)
	if err != nil {
		t.Errorf("An unexpected error occuered, while adding a new client\n[%s]", err.Error())
		return
	}

	mock := newMockable()
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
	getMeta := func(i int) ClientMeta {

		// as uuids should be unique, it makes no difference,
		// if a new one is called here.
		uuid := uuid.New()
		name := fmt.Sprintf("id: %d", i)
		mock := newMockable()

		return NewClientMeta(uuid, name, &mock)
	}

	cm := newClientManager()

	// no errors here
	for i := 0; i <= math.MaxUint8; i += 1 {
		meta := getMeta(i)
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
	meta := getMeta(256)
	err := cm.addClient(meta)
	if err == nil {
		t.Error("A client was added illigaly here.")
	}
}

func TestClientManagerRemoveClients(t *testing.T) {
	cm := newClientManager()

	meta := getClientMeta()

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
	getMeta := func(i int) ClientMeta {

		// as uuids should be unique, it makes no difference,
		// if a new one is called here.
		uuid := uuid.New()
		name := fmt.Sprintf("id: %d", i)
		mock := newMockable()

		return NewClientMeta(uuid, name, &mock)
	}

	cm := newClientManager()

	// no errors here
	for i := 0; i <= math.MaxUint8; i += 1 {
		// assume this works
		meta := getMeta(i)
		_ = cm.addClient(meta)

	}

	// shutdown all the "workers"
	cm.shutdown(context.Background())

	// check that all the worker state are on shutdown
	for _, client := range cm.clients {
		if client.status != Offline {
			t.Errorf("incorrect state has been used ")
		}
	}

	// TODO: test logger output - this needs code that
	// currently is not merged into this branch
}

func TestClientManagerForceShutdown(t *testing.T) {
	getMeta := func(i int) ClientMeta {

		// as uuids should be unique, it makes no difference,
		// if a new one is called here.
		uuid := uuid.New()
		name := fmt.Sprintf("id: %d", i)
		mock := newMockable()

		return NewClientMeta(uuid, name, &mock)
	}

	cm := newClientManager()

	// no errors here
	for i := 0; i <= math.MaxUint8; i += 1 {
		// assume this works
		meta := getMeta(i)
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
	// test for panic on logging in after lost connection
	// and incorrect removal of client data
	defer func() {
		r := recover()
		if r == nil {
			t.Error("there should have been a panic")
		}
	}()

	getMeta := func(i int) ClientMeta {

		// as uuids should be unique, it makes no difference,
		// if a new one is called here.
		uuid := uuid.New()
		name := fmt.Sprintf("id: %d", i)
		mock := newMockable()

		return NewClientMeta(uuid, name, &mock)
	}

	cm := newClientManager()
	meta := getMeta(0)
	_ = cm.addClient(meta)

	iId, _ := cm.mapId(meta)
	// forcefully remove a client from the internal client map...

	delete(cm.clients, iId)

	_ = cm.addClient(meta)
}

func TestIncorrectRemoveClient(t *testing.T) {
	// test for panic on logging in after lost connection
	// and incorrect removal of client data
	defer func() {
		r := recover()
		if r == nil {
			t.Error("there should have been a panic")
		}
	}()

	getMeta := func(i int) ClientMeta {

		// as uuids should be unique, it makes no difference,
		// if a new one is called here.
		uuid := uuid.New()
		name := fmt.Sprintf("id: %d", i)
		mock := newMockable()

		return NewClientMeta(uuid, name, &mock)
	}

	cm := newClientManager()
	meta := getMeta(0)
	_ = cm.addClient(meta)

	// forcefully push a value back onto the queue
	err := cm.idManager.Push(0)
	if err != nil {
		t.Errorf("There was an unexpected error on pushing onto the queue:\n%s", err.Error())
		return
	}

	_ = cm.removeClient(meta)
}
