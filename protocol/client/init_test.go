package client_test

import (
	"testing"

	"github.zhaw.ch/raycasters/server/protocol/client"
	"github.zhaw.ch/raycasters/server/utils"
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

func TestShooting(t *testing.T) {
	// flaot converted withhttp://www.binaryconvert.com/convert_double.html
	// 0x401B99999999999A
	angle := 6.9
	data := [...]byte{
		client.ShootingOpcode.Value(),
		//, 0x40, 0x1B, 0x99, 0x99, 0x99, 0x99, 0x99, 0x9A
		0x9A, 0x99, 0x99, 0x99, 0x99, 0x99, 0x1B, 0x40,
	}
	got, err := client.NewShootingBytes(data[:])

	if err != nil {
		t.Error(err)
		return
	}

	if got.AngPl() != angle {
		t.Errorf("Value was wrongly converted")
	}
}

func TestMovement(t *testing.T) {
	// for manual float conversion
	// http://www.binaryconvert.com/result_double.html
	// 0x40 0x84 0xD0 0x00 0x00 0x00 0x00 0x00
	coorX := 666.0
	// 0x40 0x8F 0x3F 0xFD 0xF3 0xB6 0x45 0xA2
	coorY := 999.999

	data := [...]byte{
		client.MovementOpcode.Value(),
		// 0x40 0x84 0xD0 0x00 0x00 0x00 0x00 0x00
		0x00, 0x00, 0x00, 0x00, 0x00, 0xD0, 0x84, 0x40,
		// 0x40 0x8F 0x3F 0xFD 0xF3 0xB6 0x45 0xA2
		0xA2, 0x45, 0xB6, 0xF3, 0xFD, 0x3F, 0x8F, 0x40,
	}

	val, err := client.NewMovementByte(data[:])

	if err != nil {
		t.Error(err)
		return
	}

	if val.CoorX() != coorX {
		t.Error("incorrect coordinate X")
	}

	if val.CoorY() != coorY {
		t.Error("incorrect coordinate Y")
	}
}

func TestUnknown(t *testing.T) {
	// correct version
	data := []byte{
		client.UnknownOpcode.Value(),
	}
	_, err := client.NewUnknownByte(data)
	if err != nil {
		t.Errorf("Correct version was flaged here\n%s", err)
		return
	}
}

func TestDisconnect(t *testing.T) {
	// correct version
	data := []byte{
		client.DisconnectOpcode.Value(),
	}
	_, err := client.NewDisconnectByte(data)
	if err != nil {
		t.Errorf("Correct version was flaged here\n%s", err)
		return
	}
}
func TestBinaryToFloat(t *testing.T) {

	// 1.00000000000000005551115123126E-1
	// 0x3FB999999999999A
	var expected float64 = 0.1

	rawData := [...]byte{
		// 0x3F 0xB9 0x99 0x99 0x99 0x99 0x99 0x9A
		0x9A, 0x99, 0x99, 0x99, 0x99, 0x99, 0xB9, 0x3F,
	}

	got := utils.BinaryToFloat(rawData[:])
	if got != expected {
		t.Errorf("incorrect conversion of number")
	}
}

func TestName(t *testing.T) {
	exp := [...]string{
		"Disconnect",
		"Unknown",
		"Movement",
		"Shooting",
		"RequestOtherPlayers",
	}
	for i, val := range client.Overview {
		if val.String() != exp[i] {
			t.Errorf("Name are not as expected.\n[%s] - [%s]", val.String(), exp[i])
			return
		}
	}
}
