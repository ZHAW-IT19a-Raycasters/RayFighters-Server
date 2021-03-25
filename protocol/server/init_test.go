package server_test

import (
	"math"
	"strings"
	"testing"

	"github.zhaw.ch/raycasters/server/protocol/server"
)

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

func TestPlayerBytes(t *testing.T) {
	var id uint8 = 0x42     // 1  Byte
	name := "SomeoneRandom" // 13 Byte
	player := server.NewPlayer(id, name)

	expected := []byte{0x42, 13, 0, 83, 111, 109, 101, 111, 110, 101, 82, 97, 110, 100, 111, 109}
	gotten := player.Bytes()

	if !compareByteArrays(expected, gotten, t) {
		t.Errorf("happend with\n%s", player.String())
	}
}

func TestPlayerFromBytes(t *testing.T) {
	var id uint8 = 0x42     // 1  Byte
	name := "SomeoneRandom" // 13 Byte
	expectedPlayer := server.NewPlayer(id, name)
	gottenPlayer, _ := server.NewPlayerBytes(expectedPlayer.Bytes())
	comparePlayers(gottenPlayer, expectedPlayer, t)
}

func TestPlayerNameTooLong(t *testing.T) {
	defer func() {
		err := recover()
		if err == nil {
			t.Error("Did not panic as intended")
		}
	}()

	var userId byte = 0x66
	name := strings.Repeat("Long", math.MaxUint16)

	_ = server.NewPlayer(userId, name)
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

	compareByteArrays(expectedData, gottenData, t)
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
	mp := server.NewMovementPlayer(userId, coorX, coorY)
	gottenData := mp.Bytes()

	compareByteArrays(expectedData[:], gottenData, t)

	// make sure the values are correctly set
	// basically testing getters
	if userId != mp.UserId() {
		t.Error("UserId has been wrongly set")
	} else if coorX != mp.CoorX() {
		t.Error("CoorX has been wrongly set")
	} else if coorY != mp.CoorY() {
		t.Error("CoorY has been wrongly set")
	}
}

func TestPlayerLeft(t *testing.T) {
	var userId uint8 = 0xFF
	expected := [...]byte{
		server.PlayerLeftOpcode.Value(),
		userId,
	}
	gotten := server.NewPlayerLeft(userId).Bytes()
	compareByteArrays(expected[:], gotten, t)
}

func TestNewPlayerJoined(t *testing.T) {
	var userId uint8 = 0x42
	name := "ATotalyRandomName🤬"
	expected := [...]byte{
		server.PlayerJoinedOpcode.Value(),
		userId,
		byte(len(name)),
		0,
		// ATotalyRandomName🤬 -> converted with https://onlineutf8tools.com/convert-utf8-to-bytes
		0x41, 0x54, 0x6f, 0x74, 0x61, 0x6c, 0x79, 0x52, 0x61, 0x6e, 0x64, 0x6f, 0x6d, 0x4e, 0x61, 0x6d, 0x65, 0xf0, 0x9f, 0xa4, 0xac,
	}

	gotten := server.NewPlayerJoined(server.NewPlayer(userId, name)).Bytes()

	compareByteArrays(expected[:], gotten, t)
}

func TestPlayerHasBeenKilled(t *testing.T) {
	exp := []byte{
		server.PlayerHimSelfKilledOpcode.Value(),
	}
	got := server.NewPlayerHimSelfKilled().Bytes()

	compareByteArrays(exp, got, t)
}

func TestAnOtherPlayerKilled(t *testing.T) {
	var userId byte = 0x0A

	exp := [...]byte{
		server.AnOtherPlayerKilledOpcode.Value(),
		userId,
	}

	got := server.NewAnotherPlayerKilled(userId).Bytes()

	compareByteArrays(exp[:], got, t)
}

func TestUnknown(t *testing.T) {
	exp := [...]byte{
		server.UnknownOpcode.Value(),
	}

	got := server.NewUnknown().Bytes()
	compareByteArrays(exp[:], got, t)
}

func TestDisconnect(t *testing.T) {
	exp := [...]byte{
		server.DisconnectOpcode.Value(),
	}

	got := server.NewDisconnect().Bytes()
	compareByteArrays(exp[:], got, t)
}

func TestDetect(t *testing.T) {
	opcodes := [...]server.ServerOpcode{
		server.DisconnectOpcode,
		server.DisconnectOpcode,
		server.UnknownOpcode,
		server.MovementPlayerOpcode,
		server.AnOtherPlayerKilledOpcode,
		server.PlayerHimSelfKilledOpcode,
		server.PlayerJoinedOpcode,
		server.PlayerLeftOpcode,
		server.SendAllPlayerInfoOpcode,
	}

	for _, val := range opcodes {
		got := server.Detect(val.Value())
		if val != got {
			t.Errorf("opcodes are different [%d] - [%d]", val, got)
		}
	}
}
