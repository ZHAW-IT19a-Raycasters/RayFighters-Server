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

func comparePlayers(got, exp server.Player, t *testing.T) {
	if exp.UserId() != got.UserId() {
		t.Errorf("The Id of the player is incorrect")
        return
	}

	if exp.Name() != got.Name() {
		t.Errorf("The Name of the player is incorrect")
        return
	}
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
    expectedData := []byte {
        byte(server.SendAllPlayerInfoOpcode),
        3,
    }
    
    // appending hans
    expectedData = append(expectedData, 1)
    expectedData = append(expectedData, byte(len("Hans")))
    expectedData = append(expectedData, []byte("Hans")...)

   // appending Peter 
    expectedData = append(expectedData, 4)
    expectedData = append(expectedData, byte(len("Peter")))
    expectedData = append(expectedData, []byte("Peter")...)

    // append Meier
    expectedData = append(expectedData, 52)
    expectedData = append(expectedData, byte(len("Meier")))
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
	// players := [...]server.Player{
	// 	server.NewPlayer(1, "Hans"),
	// 	server.NewPlayer(4, "Peter"),
	// 	server.NewPlayer(52, "Meier"),
	// }

}
