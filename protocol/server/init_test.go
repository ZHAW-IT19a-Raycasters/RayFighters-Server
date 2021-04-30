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
		t.Errorf("Different lenght of results exp: [%d] - got: [%d]", len(expectedData), len(gottenData))
		return false
	}

	for i := range expectedData {
		exp := expectedData[i]
		got := gottenData[i]
		if exp != got {
			t.Errorf("Bytes at position [%d] are different exp: [%d] - got: [%d]", i, exp, got)
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

func TestPlayerString(t *testing.T) {
	var id uint8 = 0x42     // 1  Byte
	name := "SomeoneRandom" // 13 Byte
	expectedPlayer := server.NewPlayer(id, name)
	str := "Id: 66\nName: SomeoneRandom"
	if expectedPlayer.String() != str {
		t.Error("Incorrect player string generated")
	}
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

	// 0x4072C40000000000
	// 3.0025E2
	var anglePl float64 = 300.25

	expectedData := [...]byte{
		server.MovementPlayerOpcode.Value(),
		userId,
		// 0x3F 0xB9 0x99 0x99 0x99 0x99 0x99 0x9A
		0x9A, 0x99, 0x99, 0x99, 0x99, 0x99, 0xB9, 0x3F,
		// 0x3F 0xC9 0x99 0x99 0x99 0x99 0x99 0x9A
		0x9A, 0x99, 0x99, 0x99, 0x99, 0x99, 0xC9, 0x3F,
		// 0x40 0x72 0xC4 0x00 0x00 0x00 0x00 0x00
		0x00, 0x00, 0x00, 0x00, 0x00, 0xC4, 0x72, 0x40,
	}

	mp := server.NewMovementPlayer(userId, coorX, coorY, anglePl)
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
	} else if anglePl != mp.AnglePl() {
		t.Error("AngePL has been wrongly set")
	}
}

func TestTeleportPlayer(t *testing.T) {
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
		server.TeleportPlayerOpcode.Value(),
		userId,
		// 0x3F 0xB9 0x99 0x99 0x99 0x99 0x99 0x9A
		0x9A, 0x99, 0x99, 0x99, 0x99, 0x99, 0xB9, 0x3F,
		// 0x3F 0xC9 0x99 0x99 0x99 0x99 0x99 0x9A
		0x9A, 0x99, 0x99, 0x99, 0x99, 0x99, 0xC9, 0x3F,
	}

	mp := server.NewTeleportPlayer(userId, coorX, coorY)
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

func TestShootingPlayer(t *testing.T) {
	// used https://www.binaryconvert.com/result_double.html?
	// for manual convertion

	// writing out the values to make sure the
	// compiler will use the correct size.

	var userId uint8 = 0x44
	// 0x4072C40000000000
	// 3.0025E2
	var anglePl float64 = 300.25

	expectedData := [...]byte{
		server.ShootingPlayerOpcode.Value(),
		userId,
		// 0x40 0x72 0xC4 0x00 0x00 0x00 0x00 0x00
		0x00, 0x00, 0x00, 0x00, 0x00, 0xC4, 0x72, 0x40,
	}

	sp := server.NewShootingPlayer(userId, anglePl)

	data := sp.Bytes()

	compareByteArrays(expectedData[:], data, t)
}

func TestPlayerLeft(t *testing.T) {
	var userId uint8 = 0xFF
	expected := [...]byte{
		server.PlayerLeftOpcode.Value(),
		userId,
	}
	pl := server.NewPlayerLeft(userId)
	gotten := pl.Bytes()
	compareByteArrays(expected[:], gotten, t)

	// test getter for coverage
	if userId != pl.UserId() {
		t.Error("UserId was incorrectly set")
	}
}

func TestNewPlayerJoined(t *testing.T) {
	var userId uint8 = 0x42
	name := "ATotalyRandomName🤬"
	expected := [...]byte{
		server.PlayerJoinedOpcode.Value(),
		userId,
		byte(len(name)),
		0,
		// ATotallyRandomName🤬 -> converted with https://onlineutf8tools.com/convert-utf8-to-bytes
		0x41, 0x54, 0x6f, 0x74, 0x61, 0x6c, 0x79, 0x52, 0x61, 0x6e, 0x64, 0x6f, 0x6d, 0x4e, 0x61, 0x6d, 0x65, 0xf0, 0x9f, 0xa4, 0xac,
	}

	gotten := server.NewPlayerJoined(server.NewPlayer(userId, name)).Bytes()

	compareByteArrays(expected[:], gotten, t)
}

func TestPlayerDeath(t *testing.T) {
	var killedId byte = 0x0B
	var killerId byte = 0x0A

	exp := [...]byte{
		server.PlayerDeathOpcode.Value(),
		killedId,
		killerId,
	}

	pd := server.NewPlayerDeath(killedId, killerId)
	got := pd.Bytes()

	compareByteArrays(exp[:], got, t)

	// testing getters (for coverage)
	if killedId != pd.KilledId() {
		t.Error("killedId was set wrongly")
		return
	}

	if killerId != pd.KillerId() {
		t.Error("killerId was set wrongly")
	}
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
	// test given values
	opcodes := server.Overview

	for _, val := range opcodes {
		got := server.Detect(val.Value())
		if val != got {
			t.Errorf("opcodes are different [%d] - [%d]", val, got)
			return
		}
	}

	// test incorrect value
	// using 10 here as it is a client send opcode
	if server.UnknownOpcode != server.Detect(10) {
		t.Error("Wrong return value on fake opcode")
	}
}

func TestYouAreCalled(t *testing.T) {
	var userId byte = 0x0A

	exp := [...]byte{
		server.YouAreCalledOpcode.Value(),
		userId,
	}

	yac := server.NewYouAreCalled(userId)
	got := yac.Bytes()

	compareByteArrays(exp[:], got, t)

	// testing getters (for coverage)
	if userId != yac.UserId() {
		t.Error("UserId was set wrongly")
	}
}

func TestPlayerKD(t *testing.T) {
	var userId byte = 0x0A
	var kills uint16 = 0xFFFF
	var deaths uint16 = 0x0001
	exp := [...]byte{
		userId,
		0xFF, 0xFF,
		0x01, 0x00,
	}

	pkd := server.NewPlayerKD(userId, kills, deaths)
	data := make([]byte, pkd.ByteSize())
	pkd.PutBytes(data)

	compareByteArrays(exp[:], data, t)

	// testing getters (for coverage)
	if userId != pkd.UserId() {
		t.Error("UserId was set wrongly")
		return
	}
	if kills != pkd.Kills() {
		t.Error("Kills was set wrongly")
		return
	}
	if deaths != pkd.Deaths() {
		t.Error("Death was set wrongly")
	}
}

func TestPlayersKD(t *testing.T) {
	// Player 1
	var userId1 byte = 0x0A
	var kills1 uint16 = 0xFFFF
	var deaths1 uint16 = 0x0001

	// Player 2
	var userId2 byte = 0x0B
	var kills2 uint16 = 0x002F
	var deaths2 uint16 = 0x0100

	// Player 3
	var userId3 byte = 0x0C
	var kills3 uint16 = 0x1010
	var deaths3 uint16 = 0x0101

	pkd1 := server.NewPlayerKD(userId1, kills1, deaths1)
	pkd2 := server.NewPlayerKD(userId2, kills2, deaths2)
	pkd3 := server.NewPlayerKD(userId3, kills3, deaths3)

	data := [...]byte{
		server.SendAllPlayerKDOpcode.Value(),
		3,
		// Player 1
		userId1,
		0xFF, 0xFF,
		0x01, 0x00,
		// Player 2
		userId2,
		0x2F, 0x00,
		0x00, 0x01,
		// Player 3
		userId3,
		0x10, 0x10,
		0x01, 0x01,
	}

	apkd := server.NewPlayersKD([]server.PlayerKD{pkd1, pkd2, pkd3})
	got := apkd.Bytes()

	compareByteArrays(data[:], got, t)
}
