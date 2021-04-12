package client

import "testing"

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
