package protocol

// Is the interface that has to be implemented by the different opcode implementations.
type Bytefy interface {
	// Will return the by the protocol defined opcode representations.
	Bytes() []byte
}
