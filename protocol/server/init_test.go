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
		t.Fatal("The lenght of th e arrays are not the same")
	}

	for i := range expected {
		if expected[i] != gotten[i] {
			t.Errorf("The byte at position [%d] should be [%d], but is [%d]", i, expected[i], gotten[i])
		}
	}

}
