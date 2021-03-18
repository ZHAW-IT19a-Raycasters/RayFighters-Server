package client

import (
	"encoding/binary"
	"fmt"
	"math"
)

type ClientOpcode uint8

const (
	DisconnectOpcode          ClientOpcode = 0x00
	UnknownOpcode             ClientOpcode = 0x01
	MovementOpcode            ClientOpcode = 0x0A
	ShootingOpcode            ClientOpcode = 0x14
	RequestOtherPlayersOpcode ClientOpcode = 0x1E
)

var overview = [...]ClientOpcode{
	DisconnectOpcode,
	MovementOpcode,
	ShootingOpcode,
	RequestOtherPlayersOpcode,
}

func (o ClientOpcode) Value() uint8 {
	// SAFETY: This is safe as the original value of
	// ClientOpcode was a uint8 anyway.
	return uint8(o)
}

func (o ClientOpcode) String() string {
	switch o {
	case DisconnectOpcode:
		return "Disconnect"
	case UnknownOpcode:
		return "Unknown"
	case MovementOpcode:
		return "Movement"
	case ShootingOpcode:
		return "Shooting"
	case RequestOtherPlayersOpcode:
		return "RequestOtherPlayers"
	default:
		// this panic should never be hit,
		// if it is hit it means there is some
		// illigal programming state.
		panic("Unknown client interface")
	}
}

func (o ClientOpcode) ByteSize() int {
	switch o {
	case DisconnectOpcode:
		// Opcode
		return 1
	case UnknownOpcode:
		// Opcode
		return 1
	case MovementOpcode:
		// Opcode + 2 * Float64
		return 1 + 2*8
	case ShootingOpcode:
		// Opcode + Flaot64
		return 1 + 8
	case RequestOtherPlayersOpcode:
		// Opcode
		return 1
	default:
		// this panic should never be hit,
		// if it is hit it means there is some
		// illigal programming state.
		panic("Unknown client interface")
	}
}

func Detect(opcode uint8) ClientOpcode {
	parsedOpcode := ClientOpcode(opcode)

	for _, element := range overview {
		if parsedOpcode == element {
			return element
		}
	}

	return UnknownOpcode
}

type ParsingError struct {
	opcode ClientOpcode
	msg    string
}

func (pe *ParsingError) Error() string {
	return fmt.Sprintf("Parsing error from %s [%s]", pe.opcode.String(), pe.msg)
}

type Disconnect struct{}

func NewDisconnect() Disconnect {
	return Disconnect{}
}

// This should actually never be called, but is here for completion.
func NewDiconnectByte(data []byte) (*Disconnect, error) {
	// checking length etc
	// 1 Byte
	if err := checkOpcodeHelper(DisconnectOpcode, data); err != nil {
		return nil, err
	}

	val := NewDisconnect()
	return &val, nil
}

type Unknown struct{}

func NewUnknown() Unknown {
	return Unknown{}
}

// This should actually never be called, but is here for completion.
func NewUnknownByte(data []byte) (*Unknown, error) {
	// checking length etc
	// 1 Byte
	if err := checkOpcodeHelper(UnknownOpcode, data); err != nil {
		return nil, err
	}

	val := NewUnknown()
	return &val, nil
}

type Movement struct {
	coorX, coorY float64
}

func (m Movement) CoorX() float64 {
	return m.coorX
}

func (m Movement) CoorY() float64 {
	return m.coorY
}

func NewMovement(coorX, coorY float64) Movement {
	return Movement{
		coorX: coorX,
		coorY: coorY,
	}
}

func NewMovementByte(data []byte) (*Movement, error) {
	if err := checkOpcodeHelper(MovementOpcode, data); err != nil {
		return nil, err
	}

	coorX := BinaryToFloat(data[1:9])
	coorY := BinaryToFloat(data[9:])

	val := NewMovement(coorX, coorY)

	return &val, nil
}

type Shooting struct {
	angPl float64
}

func (s Shooting) AngPl() float64 {
	return s.angPl
}

func NewShooting(angPl float64) Shooting {
	return Shooting{
		angPl: angPl,
	}
}

func NewSchootingBytes(data []byte) (*Shooting, error) {
	// u8 f64
	err := checkOpcodeHelper(ShootingOpcode, data)

	if err != nil {
		return nil, err
	}

	// lenght is now correct
	angle := BinaryToFloat(data[1:])

	val := NewShooting(angle)
	return &val, nil
}

type RequestOtherPlayers struct{}

func NewRequestOtherPlayers() RequestOtherPlayers {
	return RequestOtherPlayers{}
}

func NewRequestOtherPlayersBytes(data []byte) (*RequestOtherPlayers, error) {
	// checking length etc
	// 1 Byte
	if err := checkOpcodeHelper(RequestOtherPlayersOpcode, data); err != nil {
		return nil, err
	}

	val := NewRequestOtherPlayers()
	return &val, nil
}

func checkOpcodeHelper(opcode ClientOpcode, data []byte) error {
	if len(data) != opcode.ByteSize() {
		return &ParsingError{
			opcode: opcode,
			msg:    fmt.Sprintf("must be only [%d] bytes long is [%d]", opcode.ByteSize(), len(data)),
		}
	} else if Detect(data[0]) != opcode {
		return &ParsingError{
			opcode: opcode,
			msg:    "Incorrect Opcode used.",
		}
	}
	return nil
}

// TODO: Refator into into utils
// Will convert a binary blob into a correct float64.
// Attention this assumes that the float starts from
// position 0 on and that the float is formatted
// according to the IEEE specification.
func BinaryToFloat(data []byte) float64 {

	// convert the float into the int64 representation
	intRep := binary.LittleEndian.Uint64(data)

	// convert the raw int representation to
	return math.Float64frombits(intRep)
}
