package room

import "testing"

func TestNewGameState(t *testing.T) {
	stateGame := NewGameState()
	err := stateGame.AddClient(1)

	if err != nil {
		t.Error("AddClient returned an error")
	}

	client, ok := stateGame.GetClient(1)
	if !ok {
		t.Error("GetClient returned an error")
	}

	if client == nil {
		t.Error("client is nil")
	}
}

func TestGameStateIncKill(t *testing.T) {
	stateGame := NewGameState()
	err := stateGame.AddClient(1)

	if err != nil {
		t.Error("AddClient returned an error")
	}

	client, ok := stateGame.GetClient(1)
	if !ok {
		t.Error("GetClient returned an error")
	}

	if client == nil {
		t.Error("client is nil")
	}

	killsBefore := client.GetKills()
	stateGame.IncKill(1)
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
	}

	client, ok := stateGame.GetClient(1)
	if !ok {
		t.Error("GetClient returned an error")
	}

	if client == nil {
		t.Error("client is nil")
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
	}

	client, ok := stateGame.GetClient(1)
	if !ok {
		t.Error("GetClient returned an error")
	}

	if client == nil {
		t.Error("client is nil")
	}
	err = stateGame.RemoveClient(1)
	if err != nil {
		t.Error("Remove Client returned error")
	}

	client, ok = stateGame.GetClient(1)
	if ok {
		t.Error("Client wasn't removed")
	}
}

func TestGameStateRemoveClientError(t *testing.T) {
	stateGame := NewGameState()
	err := stateGame.RemoveClient(1)
	if err == nil {
		t.Error("No Error was returned while trying to remove a client that doesn't exist")
	}
}

func TestGameStateAddClientWithExistingID(t *testing.T) {
	stateGame := NewGameState()
	err := stateGame.AddClient(1)

	if err != nil {
		t.Error("AddClient returned an error")
	}

	err = stateGame.AddClient(1)

	if err == nil {
		t.Error("2 Clients with the same id were created")
	}
}

func TestGameStateGetPlayersKD(t *testing.T) {
	stateGame := NewGameState()
	err := stateGame.AddClient(1)
	if err != nil {
		t.Error("AddClient id 1 returned an error")
	}
	err = stateGame.AddClient(2)
	if err != nil {
		t.Error("AddClient id 2 returned an error")
	}
	stateGame.IncKill(2)
	stateGame.IncDeaths(1)
	stateGame.IncKill(2)
	stateGame.IncDeaths(1)
	stateGame.IncKill(1)
	stateGame.IncDeaths(2)

	expected1 := []byte {
		// PlayersKD Opcode
		0x83,
		// Amount of Players
		0x02,
		// Player with id 1
		0x01, 0x01, 0x00, 0x02, 0x00,
		// Player with id 2
		0x02, 0x02, 0x00, 0x01, 0x00,
	}

	expected2 := []byte {
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

	if !compareByteArrays(expected1, playersKD.Bytes(), t) && !compareByteArrays(expected2, playersKD.Bytes(), t) {
		t.Error("Something went wrong with Byte generation")
	}

}