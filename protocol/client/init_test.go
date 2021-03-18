package client_test

import (
	"testing"

	"github.zhaw.ch/raycasters/server/protocol/client"
)

func TestRequestOtherPlayers(t *testing.T) {
	// correct version
	data := []byte{
		client.RequestOtherPlayersOpcode.Value(),
	}

	_, err := client.NewRequestOtherPlayersBytes(data)
	if err != nil {
		t.Errorf("Correct version was flaged here\n%s", err)
		return
	}

	// incorrect size
	data = []byte{
		client.RequestOtherPlayersOpcode.Value(),
		42,
	}

	_, err = client.NewRequestOtherPlayersBytes(data)
	if err == nil {
		t.Error("Incorrect size version was not flaged here")
		return
	}

	// incorrect opcode
	data = []byte{
		client.DisconnectOpcode.Value(),
	}

	_, err = client.NewRequestOtherPlayersBytes(data)
	if err == nil {
		t.Error("Incorrect opcode version was not flaged here")
	}
}
