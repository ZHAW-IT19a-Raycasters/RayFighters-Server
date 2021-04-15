package room

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.zhaw.ch/raycasters/server/protocol/client"
	"github.zhaw.ch/raycasters/server/protocol/server"
	rClient "github.zhaw.ch/raycasters/server/room/client"
)

func getClientMeta(i int) (rClient.ClientMeta, *MockedConnection) {
	// as uuids should be unique, it makes no difference,
	// if a new one is called here.
	uuid := uuid.New()
	name := fmt.Sprintf("id: %d", i)
	mock := NewMockedConnection()
	pmock := &mock
	return rClient.NewClientMeta(uuid, name, pmock), pmock
}

func compareByteArrays(expectedData, gottenData []byte, t *testing.T) bool {
	if len(expectedData) != len(gottenData) {
		t.Errorf("Different lenght of results [%d] - [%d]", len(expectedData), len(gottenData))
		return false
	}

	for i := range expectedData {
		exp := expectedData[i]
		got := gottenData[i]
		if exp != got {
			t.Errorf("Bytes at position [%d] are different [%d] - [%d]", i, exp, got)
			return false
		}
	}
	return true
}

func TestHandleReadWhoAmI(t *testing.T) {
	hub := newHubController()
	request := []byte{
		client.WhoAmIOpcode.Value(),
	}
	actual := hub.handleWhoAmI(request, 1)
	expected := []byte{
		server.YouAreCalledOpcode.Value(),
		1,
	}
	if !compareByteArrays(expected, actual.Bytes(), t) {
		t.Errorf("Happend in WhoAmI")
	}
}

func TestHandleReadMovement(t *testing.T) {
	hub := newHubController()
	request := []byte{
		client.MovementOpcode.Value(),
		0x00, 0x00, 0x00, 0x00, 0x00, 0xD0, 0x84, 0x40,
		0xA2, 0x45, 0xB6, 0xF3, 0xFD, 0x3F, 0x8F, 0x40,
	}
	actual := hub.handleRead(request, 1)
	expected := []byte{
		server.MovementPlayerOpcode.Value(),
		1,
		0x00, 0x00, 0x00, 0x00, 0x00, 0xD0, 0x84, 0x40,
		0xA2, 0x45, 0xB6, 0xF3, 0xFD, 0x3F, 0x8F, 0x40,
	}
	if !compareByteArrays(expected, actual.Bytes(), t) {
		t.Errorf("Happend in Movement")
	}
}

func TestHandleUnknownReceived(t *testing.T) {
	hub := newHubController()

	magic := hub.handleUnknownReceived([]byte{client.UnknownOpcode.Value()})
	t.Log(magic)
}

func TestHandleDisconnect(t *testing.T) {
	hub := newHubController()
	meta, _ := getClientMeta(1)

	hub.registerClient(meta)
	// hub.clientManager.GetClients()
	magic := hub.handleDisconnect([]byte{client.DisconnectOpcode.Value()}, 0)
	t.Log(magic)
}

func TestHandleWhoAmI(t *testing.T) {
	hub := newHubController()
	request := []byte{
		client.WhoAmIOpcode.Value(),
	}
	actual := hub.handleWhoAmI(request, 1)
	expected := []byte{
		server.YouAreCalledOpcode.Value(),
		1,
	}
	if !compareByteArrays(expected, actual.Bytes(), t) {
		t.Errorf("Happend in WhoAmI")
	}

}

func TestHandleMovement(t *testing.T) {
	hub := newHubController()
	request := []byte{
		client.MovementOpcode.Value(),
		0x00, 0x00, 0x00, 0x00, 0x00, 0xD0, 0x84, 0x40,
		0xA2, 0x45, 0xB6, 0xF3, 0xFD, 0x3F, 0x8F, 0x40,
	}
	actual := hub.handleMovement(request, 1)
	expected := []byte{
		server.MovementPlayerOpcode.Value(),
		1,
		0x00, 0x00, 0x00, 0x00, 0x00, 0xD0, 0x84, 0x40,
		0xA2, 0x45, 0xB6, 0xF3, 0xFD, 0x3F, 0x8F, 0x40,
	}
	if !compareByteArrays(expected, actual.Bytes(), t) {
		t.Errorf("Happend in Movement")
	}
}

func TestHandleShooting(t *testing.T) {
	hub := newHubController()
	magic := hub.handleShooting([]byte{client.ShootingOpcode.Value()}, 1)
	t.Log(magic)
}

func TestHandleRequestOtherPlayer(t *testing.T) {
	hub := newHubController()
	magic := hub.handleRequestOtherPlayer([]byte{client.RequestOtherPlayersOpcode.Value()}, 1)
	t.Log(magic)
}
func TestHandleRequestPlayerKD(t *testing.T) {
	hub := newHubController()
	magic := hub.handleRequestPlayerKD([]byte{client.RequestPlayerKDOpcode.Value()}, 1)
	t.Log(magic)
}
