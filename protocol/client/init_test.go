package client_test

import (
	"strings"
	"testing"

	"github.zhaw.ch/raycasters/server/protocol/client"
)

// using value 100 here as this is part of the server opcodes
// and not the client ones.
const fakeOpcode = client.ClientOpcode(100)

func TestDetect(t *testing.T) {

	for _, val := range client.Overview {
		got := client.Detect(val.Value())
		if val != got {
			t.Errorf("opcodes are different [%d] - [%d]", val, got)
		}
	}

	// Test unsupported opcode
	if client.Detect(fakeOpcode.Value()) != client.UnknownOpcode {
		t.Error("Did not get unknown opocode on false opcode.")
	}
}

func TestOpcodeString(t *testing.T) {
	exp := [...]string{
		"Disconnect",
		"Unknown",
		"WhoAmI",
		"AckMap",
		"Movement",
		"Shooting",
		"RequestOtherPlayers",
		"RequestPlayerKD",
	}
	for i, val := range client.Overview {
		if val.String() != exp[i] {
			t.Errorf("Name are not as expected.\n[%s] - [%s]", val.String(), exp[i])
			return
		}
	}
}
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

func TestShootingError(t *testing.T) {

	// flaot converted withhttp://www.binaryconvert.com/convert_double.html
	// 0x401B99999999999A
	// angle := 6.9
	data := [...]byte{
		client.ShootingOpcode.Value(),
		//, 0x40, 0x1B, 0x99, 0x99, 0x99, 0x99, 0x99, 0x9A
		0x9A, 0x99, 0x99, 0x99, 0x99, 0x99, 0x1B, 0x40,
	}

	// Testing case where bytesize is incorrect.
	_, err := client.NewShootingBytes(data[:len(data)-1])

	if err == nil {
		t.Error("Incorrect bytesize was accepted as correct")
		return
	} else if !strings.HasSuffix(err.Error(), "[must be only [9] bytes long is [8]]") {
		t.Error("Incorrect suffix has been used")
		return
	}

	// Testing case where the wrong opcode has been send.
	data[0] = fakeOpcode.Value()
	_, err = client.NewShootingBytes(data[:])

	str := "[Incorrect Opcode used, should be [Shooting - 20] is [Unknown - 100]]"

	if err == nil {
		t.Error("Incorrect opcode was accepted as correct")
		return
	} else if !strings.HasSuffix(err.Error(), str) {
		t.Error("Incorrect suffix has been used")
		return
	}
}
func TestMovement(t *testing.T) {
	// for manual float conversion
	// http://www.binaryconvert.com/result_double.html
	// 0x40 0x84 0xD0 0x00 0x00 0x00 0x00 0x00
	coorX := 666.0
	// 0x40 0x8F 0x3F 0xFD 0xF3 0xB6 0x45 0xA2
	coorY := 999.999

	// 6.90000000000000035527136788005E0
	// 0x40 0x1B 0x99 0x99 0x99 0x99 0x99 0x9A
	angl := 6.9

	data := [...]byte{
		client.MovementOpcode.Value(),
		// 0x40 0x84 0xD0 0x00 0x00 0x00 0x00 0x00
		0x00, 0x00, 0x00, 0x00, 0x00, 0xD0, 0x84, 0x40,
		// 0x40 0x8F 0x3F 0xFD 0xF3 0xB6 0x45 0xA2
		0xA2, 0x45, 0xB6, 0xF3, 0xFD, 0x3F, 0x8F, 0x40,
		// 0x40 0x1B 0x99 0x99 0x99 0x99 0x99 0x9A
		0x9A, 0x99, 0x99, 0x99, 0x99, 0x99, 0x1B, 0x40,
	}

	val, err := client.NewMovementByte(data[:])

	if err != nil {
		t.Error(err)
		return
	}

	if val.CoorX() != coorX {
		t.Error("incorrect coordinate X")
		return
	}

	if val.CoorY() != coorY {
		t.Error("incorrect coordinate Y")
		return
	}

	if val.AnglPl() != angl {
		t.Error("incorrect angle")
	}
}

func TestMovementError(t *testing.T) {
	// for manual float conversion
	// http://www.binaryconvert.com/result_double.html
	// 0x40 0x84 0xD0 0x00 0x00 0x00 0x00 0x00
	// coorX := 666.0
	// 0x40 0x8F 0x3F 0xFD 0xF3 0xB6 0x45 0xA2
	// coorY := 999.999
	// 0x40 0x1B 0x99 0x99 0x99 0x99 0x99 0x9A
	// angl := 6.9

	data := [...]byte{
		client.MovementOpcode.Value(),
		// 0x40 0x84 0xD0 0x00 0x00 0x00 0x00 0x00
		0x00, 0x00, 0x00, 0x00, 0x00, 0xD0, 0x84, 0x40,
		// 0x40 0x8F 0x3F 0xFD 0xF3 0xB6 0x45 0xA2
		0xA2, 0x45, 0xB6, 0xF3, 0xFD, 0x3F, 0x8F, 0x40,
		// 0x40 0x1B 0x99 0x99 0x99 0x99 0x99 0x9A
		0x9A, 0x99, 0x99, 0x99, 0x99, 0x99, 0x1B,
	}

	_, err := client.NewMovementByte(data[:len(data)-1])

	if err == nil {
		t.Error("Incorrect bytesize was accepted as correct")
		return
	} else if !strings.HasSuffix(err.Error(), "[must be only [25] bytes long is [23]]") {
		t.Error("Incorrect suffix has been used")
		return
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

func TestUnknownError(t *testing.T) {
	// correct version
	data := []byte{
		client.UnknownOpcode.Value(),
		1,
	}
	_, err := client.NewUnknownByte(data)
	if err == nil {
		t.Error("Incorrect bytesize was accepted as correct")
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

func TestDisconnectError(t *testing.T) {

	data := []byte{
		client.DisconnectOpcode.Value(), 44,
	}
	_, err := client.NewDisconnectByte(data)
	if err == nil {
		t.Error("Incorrect bytesize was accepted as correct")
		return
	} else if !strings.HasSuffix(err.Error(), "[must be only [1] bytes long is [2]]") {
		t.Error("Incorrect suffix has been used")
		return
	}
}

func TestWhoAmI(t *testing.T) {
	// correct version
	data := []byte{
		client.WhoAmIOpcode.Value(),
	}
	_, err := client.NewWhoAmIByte(data)
	if err != nil {
		t.Errorf("Correct version was flaged here\n%s", err)
		return
	}
}

func TestWhoAmIError(t *testing.T) {

	data := []byte{
		client.WhoAmIOpcode.Value(), 44,
	}
	_, err := client.NewWhoAmIByte(data)
	if err == nil {
		t.Error("Incorrect bytesize was accepted as correct")
		return
	} else if !strings.HasSuffix(err.Error(), "[must be only [1] bytes long is [2]]") {
		t.Error("Incorrect suffix has been used")
		return
	}
}

func TestRequestPlayerKD(t *testing.T) {
	// correct version
	data := []byte{
		client.RequestPlayerKDOpcode.Value(),
	}
	_, err := client.NewRequestPlayerKDByte(data)
	if err != nil {
		t.Errorf("Correct version was flaged here\n%s", err)
		return
	}
}

func TestRequestPlayerKDError(t *testing.T) {

	data := []byte{
		client.RequestPlayerKDOpcode.Value(), 44,
	}
	_, err := client.NewRequestPlayerKDByte(data)
	if err == nil {
		t.Error("Incorrect bytesize was accepted as correct")
		return
	} else if !strings.HasSuffix(err.Error(), "[must be only [1] bytes long is [2]]") {
		t.Error("Incorrect suffix has been used")
		return
	}
}

func TestPanicOnNotSupportedTypeByteSize(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Error("Did not panic on fake opcode bytesize")
		}
	}()

	_ = fakeOpcode.ByteSize()
}

func TestPanicOnNotSupportedTypeString(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Error("Did not panic on fake opcode string")
		}
	}()

	_ = fakeOpcode.String()
}
