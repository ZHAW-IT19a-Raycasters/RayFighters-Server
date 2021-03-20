package room

import (
	"container/list"
	"fmt"
	"math"
)

// idManagerError is the error produced, if during a function call a possibly invalid state
// would have been reached.
type idManagerError struct {
	msg string
}

func newIdManagerError(msg string) idManagerError {
	return idManagerError{
		msg: msg,
	}
}

// The error interface
func (m idManagerError) Error() string {
	return fmt.Sprintf("Incorrect handling of the idManager id is %s", m.msg)
}

// idManager is an abstraction over the standard library
// linkedlist implementation for uint8 values.
type idManager struct {
	// The list of currently available user ids
	// if this is empty all ids have been used.
	available *list.List
}

// Will create a new idManager and fills it up with all values for uint8.
func newIdManager() idManager {
	ids := list.New()
	for i := 0; i <= math.MaxUint8; i += 1 {
		// make sure that only uint8 values are pushed into the
		// queue
		ids.PushBack(uint8(i))
	}
	return idManager{
		available: ids,
	}
}

// Will append the value to the back of the list.
// Can return an error, if the list is full.
func (m idManager) push(val uint8) error {
	if m.isFull() {
		return newIdManagerError("full")
	}
	m.available.PushBack(val)
	return nil
}

// Will return the currently first value of the queue.
// Attention: Will return an error if the queue is empty
// as well as a 0
func (m idManager) pop() (uint8, error) {
	if m.isEmpty() {
		return 0, newIdManagerError("empty")
	}
	val := m.available.Front()
	return val.Value.(uint8), nil
}

func (m idManager) isEmpty() bool {
	return m.available.Len() > 0
}

func (m idManager) isFull() bool {
	return m.available.Len() <= math.MaxUint8
}
