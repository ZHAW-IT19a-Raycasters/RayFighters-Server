package room

import (
	"io"
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

func TestAddClientSigle(t *testing.T) {
	cm := newClientManager()

	uuid := uuid.New()
	name := "Hans"
	// Prepare the mocked data
	mock := Mockable{
		funcNextReader: func() (int, io.Reader, error) {

			return 0, nil, nil
		},
		funcNextWriter: func(messageType int) (io.WriteCloser, error) {

			return nil, nil
		},
	}

	clientMeta := NewClientMeta(uuid, name, &mock)
	cm.addClient(clientMeta)

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

	iId, ok := cm.mapping[uuid]
	if !ok {
		t.Error("Client has not been mapped correctly, missing entry in id mapping.")
		return
	}

	client, ok := cm.clients[iId]

	if !ok {
		t.Error("Client has not been mapped correctly, missing entry in clients")
		return
	}

	if client.meta.eId != uuid {
		t.Error("Wrong uuid has been used, while mapping.")
	}

	// create a new client, with the same id => count should stay the same

	mock = Mockable{
		funcNextReader: func() (int, io.Reader, error) {

			return 0, nil, nil
		},
		funcNextWriter: func(messageType int) (io.WriteCloser, error) {

			return nil, nil
		},
	}

}

func TestAddClientTwice(t *testing.T) {
	cm := newClientManager()

	uuid := uuid.New()
	name := "Hans"
	// Prepare the mocked data
	mock := Mockable{
		funcNextReader: func() (int, io.Reader, error) {

			return 0, nil, nil
		},
		funcNextWriter: func(messageType int) (io.WriteCloser, error) {

			return nil, nil
		},
	}

	clientMeta := NewClientMeta(uuid, name, &mock)
	cm.addClient(clientMeta)

	clientMeta = NewClientMeta(uuid, name, &mock)
	cm.addClient(clientMeta)

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
