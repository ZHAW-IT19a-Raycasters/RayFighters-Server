package server_test

import (
	"testing"

	"github.zhaw.ch/raycasters/server/protocol/server"
)

func TestPlayerBytes(t *testing.T) {
	var id uint8 = 0x42     // 1  Byte
	name := "SomeoneRandom" // 13 Byte
	player := server.NewPlayer(id, name)

	expected := []byte{0x42, 13, 0, 83, 111, 109, 101, 111, 110, 101, 82, 97, 110, 100, 111, 109}
	gotten := player.Bytes()

	if len(expected) != len(gotten) {
		t.Errorf("The lenght of th e arrays are not the same")
		return
	}

	for i := range expected {
		if expected[i] != gotten[i] {
			t.Errorf("The byte at position [%d] should be [%d], but is [%d]", i, expected[i], gotten[i])
			return
		}
	}

}

func comparePlayers(got, exp server.Player, t *testing.T) bool {
	if exp.UserId() != got.UserId() {
		t.Errorf("The Id of the player is incorrect")
		return false
	}

	if exp.Name() != got.Name() {
		t.Errorf("The Name of the player is incorrect")
		return false
	}
	return true
}

func TestPlayerFromBytes(t *testing.T) {
	var id uint8 = 0x42     // 1  Byte
	name := "SomeoneRandom" // 13 Byte
	expectedPlayer := server.NewPlayer(id, name)
	gottenPlayer, _ := server.NewPlayerBytes(expectedPlayer.Bytes())
	comparePlayers(gottenPlayer, expectedPlayer, t)
}

func TestSendAllPlayerInfo(t *testing.T) {
	players := [...]server.Player{
		server.NewPlayer(1, "Hans"),
		server.NewPlayer(4, "Peter"),
		server.NewPlayer(52, "Meier"),
	}
	data := server.NewSendAllPlayerInfo(players[:])
	gottenData := data.Bytes()
	expectedData := []byte{
		byte(server.SendAllPlayerInfoOpcode),
		3,
	}

	// appending hans
	expectedData = append(expectedData, 1)
	expectedData = append(expectedData, byte(len("Hans")))
	expectedData = append(expectedData, 0)
	expectedData = append(expectedData, []byte("Hans")...)

	// appending Peter
	expectedData = append(expectedData, 4)
	expectedData = append(expectedData, byte(len("Peter")))
	expectedData = append(expectedData, 0)
	expectedData = append(expectedData, []byte("Peter")...)

	// append Meier
	expectedData = append(expectedData, 52)
	expectedData = append(expectedData, byte(len("Meier")))
	expectedData = append(expectedData, 0)
	expectedData = append(expectedData, []byte("Meier")...)

	if len(gottenData) != len(expectedData) {
		t.Errorf("The data lenght is different [%d] - [%d]", len(gottenData), len(expectedData))
		return
	}

	for i := range expectedData {
		if expectedData[i] != gottenData[i] {
			t.Errorf("the byted at position [%d] are not equals [%d] - [%d]", i, expectedData[i], gottenData[i])
			return
		}
	}

}

func TestPlayerFromBytesMuliple(t *testing.T) {
	players := [...]server.Player{
		server.NewPlayer(1, "Hans"),
		server.NewPlayer(4, "Peter"),
		server.NewPlayer(52, "Meier"),
	}

	sapi := server.NewSendAllPlayerInfo(players[:])

	gottenOpcode := server.NewSendAllPlayerInfoBytes(sapi.Bytes())
	gottenPlayers := gottenOpcode.Players()
	if len(gottenPlayers) != len(players) {
		t.Errorf("Array lenght is different [%d] - [%d]", len(gottenPlayers), len(players))
		return
	}

	for i := range gottenPlayers {
		if !comparePlayers(gottenPlayers[i], players[i], t) {
			t.Errorf("players are not equals\n%s\n%s\n", players[i], gottenPlayers[i])
			return
		}
	}
}

func TestMovementPlayerBytes(t *testing.T) {
	// used https://www.binaryconvert.com/result_double.html?
	// for manual convertion

	// writing out the values to make sure the
	// compiler will use the correct size.
	var userId uint8 = 0x44

	// 1.00000000000000005551115123126E-1
	// 0x3FB999999999999A
	var coorX float64 = 0.1
	// 0x3FC999999999999A
	// 2.00000000000000011102230246252E-1
	var coorY float64 = 0.2

	expectedData := [...]byte{
		server.MovementPlayerOpcode.Value(),
		userId,
		// 0x3F 0xB9 0x99 0x99 0x99 0x99 0x99 0x9A
		0x9A, 0x99, 0x99, 0x99, 0x99, 0x99, 0xB9, 0x3F,
		// 0x3F 0xC9 0x99 0x99 0x99 0x99 0x99 0x9A
		0x9A, 0x99, 0x99, 0x99, 0x99, 0x99, 0xC9, 0x3F,
	}
	gottenData := server.NewMovementPlayer(userId, coorX, coorY).Bytes()

	if len(expectedData) != len(gottenData) {
		t.Errorf("Different lenght of results [%d] - [%d]", len(expectedData), len(gottenData))
		return
	}

	for i := range expectedData {
		exp := expectedData[i]
		got := gottenData[i]
		if exp != got {

			t.Errorf("Bytes at position [%d] are different [%d] - [%d]", i, exp, got)
			return
		}
	}
}
