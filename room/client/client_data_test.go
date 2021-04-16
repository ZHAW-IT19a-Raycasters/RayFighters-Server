package client

import (
	"testing"

	"github.com/google/uuid"
)

func TestConnectionStatusString(t *testing.T) {
	// This test does not test anything interesting
	// but is only for coverage
	vals := [...]string{
		"Active", "Registered", "Timeout", "ShuttingDown", "Offline",
	}

	for i := 0; i < 4; i += 1 {
		internal := ConnectionStatus(i)
		if internal.String() != vals[i] {
			t.Error("Incorrect enum message")
			return
		}
	}
}

func TestClientGame(t *testing.T) {
	cg := NewClientGame()

	if cg.GetDeaths() != 0 {
		t.Error("Incorrect start value for deaths")
		return
	}

	if cg.GetKills() != 0 {
		t.Error("Incorrect start value for kills")
		return
	}

	cg.IncDeaths()

	if cg.GetDeaths() != 1 {
		t.Error("Incorrect value for deaths the client was killed once")
		return
	}

	cg.IncKills()
	cg.IncKills()
	cg.IncKills()

	if cg.GetKills() != 3 {
		t.Error("Incorrect start value for kills")
		return
	}
}

func TestInternalMeta(t *testing.T) {
	var id uint8 = 2
	uuid := uuid.New()
	name := "Hi"
	meta := NewClientMeta(uuid, name, nil)

	imeta := newClientInternalMeta(2, meta)

	if imeta.GetId() != id {
		t.Error("Incorrect is used")
		return
	}

	pl := imeta.GetPlayer()
	if pl.Name() != name {
		t.Error("wrong name propagated")
		return
	} else if pl.UserId() != id {
		t.Error("incorrect user id")
		return
	}
}
