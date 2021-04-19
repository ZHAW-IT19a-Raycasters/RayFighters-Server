package controller

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.zhaw.ch/raycasters/server/protocol/client"
	"github.zhaw.ch/raycasters/server/protocol/server"
	rClient "github.zhaw.ch/raycasters/server/room/client"
	"github.zhaw.ch/raycasters/server/room/helpers"
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

func TestHandleReadUnknownReceived(t *testing.T) {
	hub := NewHubController()
	request := []byte{
		client.UnknownOpcode.Value(),
	}
	actual := hub.handleRead(request, 1)
	expected := []byte{
		server.UnknownOpcode.Value(),
	}
	if !compareByteArrays(expected, actual.Bytes(), t) {
		t.Errorf("Happend in ReadUnknownReveived")
	}
}

func TestHandleReadWhoAmI(t *testing.T) {
	hub := NewHubController()
	request := []byte{
		client.WhoAmIOpcode.Value(),
	}
	actual := hub.handleRead(request, 1)
	expected := []byte{
		server.YouAreCalledOpcode.Value(),
		1,
	}
	if !compareByteArrays(expected, actual.Bytes(), t) {
		t.Errorf("Happend in WhoAmI")
	}
}

func TestHandleReadWhoamiError(t *testing.T) {
	hub := NewHubController()
	request := []byte{
		client.WhoAmIOpcode.Value(),
		0x01,
	}
	actual := hub.handleRead(request, 1)
	expected := []byte{
		server.UnknownOpcode.Value(),
	}
	if !compareByteArrays(expected, actual.Bytes(), t) {
		t.Errorf("Happend in WhoAmIError")
	}
}

func TestHandleReadMovement(t *testing.T) {
	hub := NewHubController()
	cQueue, hQueue := helpers.NewWorkQueue()
	const otherId uint8 = 2
	hub.connections[otherId] = hQueue

	_, hQueue = helpers.NewWorkQueue()
	const userId uint8 = 1
	hub.connections[userId] = hQueue

	request := []byte{
		client.MovementOpcode.Value(),
		0x00, 0x00, 0x00, 0x00, 0x00, 0xD0, 0x84, 0x40,
		0xA2, 0x45, 0xB6, 0xF3, 0xFD, 0x3F, 0x8F, 0x40,
		0x00, 0x00, 0x00, 0x00, 0x00, 0xC4, 0x72, 0x40,
	}
	actual := hub.handleRead(request, userId)
	expected := []byte{
		server.MovementPlayerOpcode.Value(),
		1,
		0x00, 0x00, 0x00, 0x00, 0x00, 0xD0, 0x84, 0x40,
		0xA2, 0x45, 0xB6, 0xF3, 0xFD, 0x3F, 0x8F, 0x40,
		0x00, 0x00, 0x00, 0x00, 0x00, 0xC4, 0x72, 0x40,
	}

	if actual != nil {
		t.Error("there should not be a returned value on movement player")
		return
	}

	gotten, err := cQueue.Dequeue()

	if err != nil {
		t.Errorf("An unexepected error happend\n[%s]", err.Error())
		return
	}

	if !compareByteArrays(expected, gotten, t) {
		t.Errorf("Happend in Movement")
	}
}

func TestHandleReadMovementError(t *testing.T) {
	hub := NewHubController()

	const userId uint8 = 1

	request := []byte{
		client.MovementOpcode.Value(),
		0x00, 0x00, 0x00, 0x00, 0x00, 0xD0, 0x84, 0x40,
		0xA2, 0x45, 0xB6, 0xF3, 0xFD, 0x3F, 0x8F, 0x40,
		0x00, 0x00, 0x00, 0x00, 0x00, 0xC4, 0x72,
	}
	actual := hub.handleMovement(request, userId)
	expected := []byte{
		server.UnknownOpcode.Value(),
	}

	if !compareByteArrays(expected, actual.Bytes(), t) {
		t.Errorf("Happend in Movement")
	}
}

// TODO: THIS TEST
func TestHandleReadDisconnect(t *testing.T) {
	hub := NewHubController()
	meta, _ := getClientMeta(1)

	hub.RegisterClient(meta)
	// hub.clientManager.GetClients()
	magic := hub.handleDisconnect([]byte{client.DisconnectOpcode.Value()}, 0)
	t.Log(magic)
}

func TestHandleReadShooting(t *testing.T) {
	hub := NewHubController()
	cQueue, hQueue := helpers.NewWorkQueue()
	const otherId uint8 = 2
	hub.connections[otherId] = hQueue

	_, hQueue = helpers.NewWorkQueue()
	const userId uint8 = 1
	hub.connections[userId] = hQueue

	request := []byte{
		client.ShootingOpcode.Value(),
		0x00, 0x00, 0x00, 0x00, 0x00, 0xD0, 0x84, 0x40,
	}
	actual := hub.handleRead(request, userId)
	expected := []byte{
		server.ShootingPlayerOpcode.Value(),
		1,
		0x00, 0x00, 0x00, 0x00, 0x00, 0xD0, 0x84, 0x40,
	}

	if actual != nil {
		t.Error("there should not be a returned value on movement player")
		return
	}

	gotten, err := cQueue.Dequeue()

	if err != nil {
		t.Errorf("An unexepected error happend\n[%s]", err.Error())
		return
	}

	if !compareByteArrays(expected, gotten, t) {
		t.Errorf("Happend in Shooting")
	}
}

func TestHandleReadShootingError(t *testing.T) {
	hub := NewHubController()
	const userId uint8 = 1

	request := []byte{
		client.ShootingOpcode.Value(),
		0x00, 0x00, 0x00, 0x00, 0x00, 0xD0, 0x84,
	}
	actual := hub.handleRead(request, userId)
	expected := []byte{
		server.UnknownOpcode.Value(),
	}

	if actual == nil {
		t.Error("there should not be a returned value on movement player")
		return
	}

	if !compareByteArrays(expected, actual.Bytes(), t) {
		t.Errorf("Happend in Movement")
	}
}

func TestHandleReadRequestOtherPlayer(t *testing.T) {
	hub := NewHubController()
	cQueue, hQueue := helpers.NewWorkQueue()
	const otherId uint8 = 2
	client2Meta, _ := getClientMeta(2)

	hub.RegisterClient(client2Meta)
	hub.connections[otherId] = hQueue

	_, hQueue = helpers.NewWorkQueue()
	const userId uint8 = 1

	client1Meta, _ := getClientMeta(1)

	hub.RegisterClient(client1Meta)
	hub.connections[userId] = hQueue

	request := []byte{
		client.RequestOtherPlayersOpcode.Value(),
	}
	actual := hub.handleRead(request, userId)
	expected := []byte{
		server.SendAllPlayerInfoOpcode.Value(),
		0x02,
	}

	// AppendPlayer2
	id2, _ := hub.clientManager.MapId(client2Meta)
	expected = append(expected, id2)
	expected = append(expected, byte(len(client2Meta.GetName())))
	expected = append(expected, 0)
	expected = append(expected, []byte(client2Meta.GetName())...)

	// AppendPlayer1
	id1, _ := hub.clientManager.MapId(client1Meta)
	expected = append(expected, id1)
	expected = append(expected, byte(len(client1Meta.GetName())))
	expected = append(expected, 0)
	expected = append(expected, []byte(client1Meta.GetName())...)

	if actual != nil {
		t.Error("there should not be a returned value on movement player")
		return
	}

	gotten, err := cQueue.Dequeue()
	gotten, err = cQueue.Dequeue()
	gotten, err = cQueue.Dequeue()

	if err != nil {
		t.Errorf("An unexepected error happend\n[%s]", err.Error())
		return
	}

	if !compareByteArrays(expected, gotten, t) {
		t.Errorf("Happend in Request other Players")
	}
}

func TestHandleReadRequestPlayerKD(t *testing.T) {
	hub := NewHubController()
	magic := hub.handleRequestPlayerKD([]byte{client.RequestPlayerKDOpcode.Value()})
	t.Log(magic)
}
