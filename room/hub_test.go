package room

import (
	"fmt"
	"io"
	"math"
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

func TestAddClientSigle(t *testing.T) {
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

func TestAddClientTwice(t *testing.T) {
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

func TestMaxClients(t *testing.T) {
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

func TestRemoveClients(t *testing.T) {
	cm := newClientManager()

	meta := getClientMeta()

	cm.addClient(meta)

	if cm.size() != 1 {
		t.Error("size is not at 1")
		return
	}

	// // ignore the ok, as I know it exists
	// id, _ := cm.mapId(meta)
	// val := cm.clients[id]

	cm.removeClient(meta)

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
