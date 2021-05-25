// Package protocol contains all the abstractions needed
// to work with the server - client opcodes defined here:
// https://github.zhaw.ch/raycasters/docs/wiki/Protocol
package protocol

// Package is a readability abstraction.
// Used for clearer handling of the data
// packages that are received from or are
// send to the client.
type Package = []byte

// Is the interface that has to be implemented by the different opcode implementations.
type Bytefy interface {
	// Will return the by the protocol defined opcode representations.
	Bytes() Package
}
