package controller

import (
	"errors"
	"strings"
	"testing"
)

func compareByteArraysNoError(expectedData, gottenData []byte) bool {
	if len(expectedData) != len(gottenData) {
		return false
	}

	for i := range expectedData {
		exp := expectedData[i]
		got := gottenData[i]
		if exp != got {
			return false
		}
	}
	return true
}

func TestNewGameState(t *testing.T) {
	stateGame := NewGameState()
	err := stateGame.AddClient(1)

	if err != nil {
		t.Error("AddClient returned an error")
		return
	}

	client, ok := stateGame.GetClient(1)
	if !ok {
		t.Error("GetClient returned an error")
		return
	}

	if client == nil {
		t.Error("client is nil")
	}
}

func TestGameStateIncKill(t *testing.T) {
	const id uint8 = 1
	stateGame := NewGameState()
	err := stateGame.AddClient(id)

	if err != nil {
		t.Error("AddClient returned an error")
		return
	}

	client, ok := stateGame.GetClient(id)
	if !ok {
		t.Error("GetClient returned an error")
		return
	}

	if client == nil {
		t.Error("client is nil")
		return
	}

	killsBefore := client.GetKills()
	stateGame.IncKill(id)
	killsAfter := client.GetKills()

	if killsAfter-1 != killsBefore {
		t.Error("Amount of Kills wasn't increased")
	}
}

func TestGameStateIncDeaths(t *testing.T) {
	stateGame := NewGameState()
	err := stateGame.AddClient(1)

	if err != nil {
		t.Error("AddClient returned an error")
		return
	}

	client, ok := stateGame.GetClient(1)
	if !ok {
		t.Error("GetClient returned an error")
		return
	}

	if client == nil {
		t.Error("client is nil")
		return
	}

	killsBefore := client.GetDeaths()
	stateGame.IncDeaths(1)
	killsAfter := client.GetDeaths()
	if killsAfter-1 != killsBefore {
		t.Error("Amount of Deaths wasn't increased")
	}
}

func TestGameStateRemoveClient(t *testing.T) {
	stateGame := NewGameState()
	err := stateGame.AddClient(1)

	if err != nil {
		t.Error("AddClient returned an error")
		return
	}

	client, ok := stateGame.GetClient(1)
	if !ok {
		t.Error("GetClient returned an error")
		return
	}

	if client == nil {
		t.Error("client is nil")
		return
	}
	err = stateGame.RemoveClient(1)
	if err != nil {
		t.Error("Remove Client returned error")
		return
	}

	_, ok = stateGame.GetClient(1)
	if ok {
		t.Error("Client wasn't removed")
	}
}

func TestGameStateRemoveClientError(t *testing.T) {
	stateGame := NewGameState()
	err := stateGame.RemoveClient(1)
	if err == nil {
		t.Error("No Error was returned while trying to remove a client that doesn't exist")
	} else if !errors.Is(err, ErrClientDoesNotExist{}) {
		t.Error("Wrong error type returned")
	} else if !strings.HasSuffix(err.Error(), "[1] does currently not exists") {
		t.Error("Wrong error suffix used")
	}
}

func TestGameStateAddClientWithExistingID(t *testing.T) {
	stateGame := NewGameState()
	err := stateGame.AddClient(1)

	if err != nil {
		t.Error("AddClient returned an error")
		return
	}

	err = stateGame.AddClient(1)

	if err == nil {
		t.Error("2 Clients with the same id were created")
	} else if !errors.Is(err, ErrClientExists{}) {
		t.Error("Wrong error type returned")
	} else if !strings.HasSuffix(err.Error(), "[1] already exists") {
		t.Error("Wrong error suffix used")
	}
}

func TestGameStateGetPlayersKD(t *testing.T) {
	stateGame := NewGameState()
	err := stateGame.AddClient(1)
	if err != nil {
		t.Error("AddClient id 1 returned an error")
		return
	}
	err = stateGame.AddClient(2)
	if err != nil {
		t.Error("AddClient id 2 returned an error")
		return
	}
	stateGame.IncKill(2)
	stateGame.IncDeaths(1)
	stateGame.IncKill(2)
	stateGame.IncDeaths(1)
	stateGame.IncKill(1)
	stateGame.IncDeaths(2)

	expected1 := []byte{
		// PlayersKD Opcode
		0x83,
		// Amount of Players
		0x02,
		// Player with id 1
		0x01, 0x01, 0x00, 0x02, 0x00,
		// Player with id 2
		0x02, 0x02, 0x00, 0x01, 0x00,
	}

	expected2 := []byte{
		// PlayersKD Opcode
		0x83,
		// Amount of Players
		0x02,
		// Player with id 1
		0x02, 0x02, 0x00, 0x01, 0x00,
		// Player with id 2
		0x01, 0x01, 0x00, 0x02, 0x00,
	}

	playersKD := stateGame.GetPlayersKD()

	if !compareByteArraysNoError(expected1, playersKD.Bytes()) &&
		!compareByteArraysNoError(expected2, playersKD.Bytes()) {
		t.Error("Something went wrong with Byte generation")
	}

}
