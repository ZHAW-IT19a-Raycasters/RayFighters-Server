package client

import (
	"fmt"

	"github.zhaw.ch/raycasters/server/utils"
)

// A Client Opcode represents the internal representations
// of all the defined client side operational
// codes that control the game flow.
type ClientOpcode uint8

const (
	DisconnectOpcode          ClientOpcode = 0x00
	UnknownOpcode             ClientOpcode = 0x01
	WhoAmIOpcode              ClientOpcode = 0x02
	MovementOpcode            ClientOpcode = 0x0A
	ShootingOpcode            ClientOpcode = 0x14
	RequestOtherPlayersOpcode ClientOpcode = 0x1E
	RequestPlayerKDOpcode     ClientOpcode = 0x1F
)

var Overview = [...]ClientOpcode{
	DisconnectOpcode,
	UnknownOpcode,
	WhoAmIOpcode,
	MovementOpcode,
	ShootingOpcode,
	RequestOtherPlayersOpcode,
	RequestPlayerKDOpcode,
}

// Value returns the internal value representation as a uint8.
func (o ClientOpcode) Value() uint8 {
	// SAFETY: This is safe as the original value of
	// ClientOpcode was a uint8 anyway.
	return uint8(o)
}

// Value will retrieve the Opcode and return the String value.
// If its and unknown opcode, I will panic.
func (o ClientOpcode) String() string {
	switch o {
	case DisconnectOpcode:
		return "Disconnect"
	case UnknownOpcode:
		return "Unknown"
	case WhoAmIOpcode:
		return "WhoAmI"
	case MovementOpcode:
		return "Movement"
	case ShootingOpcode:
		return "Shooting"
	case RequestOtherPlayersOpcode:
		return "RequestOtherPlayers"
	case RequestPlayerKDOpcode:
		return "RequestPlayerKD"
	default:
		// this panic should never be hit,
		// if it is hit it means there is some
		// illigal programming state.
		panic("Unknown client interface")
	}
}

// ByteSize returns the length of an opcode.
// If its and unknown opcode, I will panic.
func (o ClientOpcode) ByteSize() int {
	switch o {
	case DisconnectOpcode:
		// Opcode
		return 1
	case UnknownOpcode:
		// Opcode
		return 1
	case WhoAmIOpcode:
		// Opcode
		return 1
	case MovementOpcode:
		// Opcode + 3 * Float64
		return 1 + 3*8
	case ShootingOpcode:
		// Opcode + Flaot64
		return 1 + 8
	case RequestOtherPlayersOpcode:
		// Opcode
		return 1
	case RequestPlayerKDOpcode:
		// Opcode
		return 1
	default:
		// this panic should never be hit,
		// if it is hit it means there is some
		// illigal programming state.
		panic("Unknown client interface")
	}
}

// Detect converts from the an external byte into the internally
// used ServerOpcode type.
//
// Will return Unknown when there is no mapping for the given value.
func Detect(opcode uint8) ClientOpcode {
	parsedOpcode := ClientOpcode(opcode)

	for _, element := range Overview {
		if parsedOpcode == element {
			return element
		}
	}

	return UnknownOpcode
}

// ParsingError represents an Error that occurs trying to parse a faulty Opcode.
type ParsingError struct {
	opcode ClientOpcode
	msg    string
}

// Error will retrieve the ParsingError and return an Errormessage.
func (pe *ParsingError) Error() string {
	return fmt.Sprintf("Parsing error from %s [%s]", pe.opcode.String(), pe.msg)
}

// Disconnect represents a client shutdown.
type Disconnect struct{}

// NewDisconnect creates a new Disconnect instance.
func NewDisconnect() Disconnect {
	return Disconnect{}
}

// NewDisconnectByte returns a pointer from a new Disconnect object based on data.
// If an error occurs Nil will be returned.
// This should actually never be called, but is here for completion.
func NewDisconnectByte(data []byte) (*Disconnect, error) {
	// checking length etc
	// 1 Byte
	if err := checkOpcodeHelper(DisconnectOpcode, data); err != nil {
		return nil, err
	}

	val := NewDisconnect()
	return &val, nil
}

// Unknown represents the fallback opcode in case there is a
// difference of protocol version or if there is an unknown
// opcode it self send.
type Unknown struct{}

// Will initialize the Unknown opcode struct.
func NewUnknown() Unknown {
	return Unknown{}
}

// NewUnknownByte returns a pointer from a new Unknown object based on data.
// If an error occurs Nil will be returned.
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

// Disconnect represents a request for a Player.
type WhoAmI struct{}

// Will initialize the WhoAmI struct.
func NewWhoAmI() WhoAmI {
	return WhoAmI{}
}

// NewWhoAmIByte returns a pointer from a new WhoAmI object based on data.
// If an error occurs Nil will be returned.
func NewWhoAmIByte(data []byte) (*WhoAmI, error) {
	// checking length etc
	// 1 Byte
	if err := checkOpcodeHelper(WhoAmIOpcode, data); err != nil {
		return nil, err
	}

	val := NewWhoAmI()
	return &val, nil
}

// Movement represents a user Movement.
type Movement struct {
	coorX, coorY, anglPl float64
}

// CooX returns the Movements x Coordinate as float64.
func (m Movement) CoorX() float64 {
	return m.coorX
}

// CooX returns the Movements y Coordinate as float64.
func (m Movement) CoorY() float64 {
	return m.coorY
}

func (m Movement) AnglPl() float64 {
	return m.anglPl
}

// NewMovement constructs a new Movement object.
func NewMovement(coorX, coorY, anglPl float64) Movement {
	return Movement{
		coorX:  coorX,
		coorY:  coorY,
		anglPl: anglPl,
	}
}

// NewMovementByte returns a pointer from a new Movement object based on data
// If an error occurs Nil will be returned
func NewMovementByte(data []byte) (*Movement, error) {
	if err := checkOpcodeHelper(MovementOpcode, data); err != nil {
		return nil, err
	}

	coorX := utils.BinaryToFloat(data[1:9])
	coorY := utils.BinaryToFloat(data[9:17])
	anglPl := utils.BinaryToFloat(data[17:])

	val := NewMovement(coorX, coorY, anglPl)

	return &val, nil
}

// Shooting represents a shot fired by a player
type Shooting struct {
	angPl float64
}

// AngPl returns the Angle of the shot as float64
func (s Shooting) AngPl() float64 {
	return s.angPl
}

// NewShooting constructs a new Shooting object.
func NewShooting(angPl float64) Shooting {
	return Shooting{
		angPl: angPl,
	}
}

// NewShootingBytes returns a pointer from a new Shooting object based on data.
// If an error occurs Nil will be returned.
func NewShootingBytes(data []byte) (*Shooting, error) {
	// u8 f64
	err := checkOpcodeHelper(ShootingOpcode, data)

	if err != nil {
		return nil, err
	}

	// lenght is now correct
	angle := utils.BinaryToFloat(data[1:])

	val := NewShooting(angle)
	return &val, nil
}

// RequestOtherPlayers is the opcode that Requests the name and id of the other players.
type RequestOtherPlayers struct{}

// NewRequestOtherPlayers creates a new RequestOtherPlayers instance.
func NewRequestOtherPlayers() RequestOtherPlayers {
	return RequestOtherPlayers{}
}

// Disconnect represents a request for a Player.
type RequestPlayerKD struct{}

// Will initialize the WhoAmI struct.
func NewRequestPlayerKD() RequestPlayerKD {
	return RequestPlayerKD{}
}

// NewWhoAmIByte returns a pointer from a new WhoAmI object based on data.
// If an error occurs Nil will be returned.
func NewRequestPlayerKDByte(data []byte) (*RequestPlayerKD, error) {
	// checking length etc
	// 1 Byte
	if err := checkOpcodeHelper(RequestPlayerKDOpcode, data); err != nil {
		return nil, err
	}

	val := NewRequestPlayerKD()
	return &val, nil
}

// NewRequestOtherPlayersBytes returns a pointer from a new RequestOtherPlayers object based on data
// If an error occurs Nil will be returned
func NewRequestOtherPlayersBytes(data []byte) (*RequestOtherPlayers, error) {
	// checking length etc
	// 1 Byte
	if err := checkOpcodeHelper(RequestOtherPlayersOpcode, data); err != nil {
		return nil, err
	}

	val := NewRequestOtherPlayers()
	return &val, nil
}

// checkOpcodeHelper checks if data is a valid opcode.
// if its invalid a ParsingError will be returned.
// if its valid Nil will be returned.
func checkOpcodeHelper(opcode ClientOpcode, data []byte) error {
	// testing detect first, as to see if the opcode even is a opcode
	// or is unknown.

	if detected := Detect(data[0]); detected != opcode {
		return &ParsingError{
			opcode: opcode,
			msg: fmt.Sprintf("Incorrect Opcode used, should be [%s - %d] is [%s - %d]",
				opcode.String(),
				opcode,
				detected.String(),
				data[0],
			),
		}
	} else if len(data) != opcode.ByteSize() {
		return &ParsingError{
			opcode: opcode,
			msg:    fmt.Sprintf("must be only [%d] bytes long is [%d]", opcode.ByteSize(), len(data)),
		}
	}
	return nil
}
