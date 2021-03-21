package helpers

import (
	"container/list"
	"fmt"
	"math"
)

// IdManagerError is the error produced, if during a function call a possibly invalid state
// would have been reached.
type IdManagerError struct {
	msg string
}

func NewIdManagerError(msg string) IdManagerError {
	return IdManagerError{
		msg: msg,
	}
}

// The error interface
func (m IdManagerError) Error() string {
	return fmt.Sprintf("Incorrect handling of the idManager id is %s", m.msg)
}

// IdManager is an abstraction over the standard library
// linkedlist implementation for uint8 values.
type IdManager struct {
	// The list of currently available user ids
	// if this is empty all ids have been used.
	available *list.List
}

// Will create a new idManager and fills it up with all values for uint8.
func NewIdManager() IdManager {
	ids := list.New()
	for i := 0; i <= math.MaxUint8; i += 1 {
		// make sure that only uint8 values are pushed into the
		// queue
		ids.PushBack(uint8(i))
	}
	return IdManager{
		available: ids,
	}
}

// Will append the value to the back of the list.
// Can return an error, if the list is full.
func (m IdManager) Push(val uint8) error {
	if m.IsFull() {
		return NewIdManagerError("full")
	}
	m.available.PushBack(val)
	return nil
}

// Will return the currently first value of the queue.
// Attention: Will return an error if the queue is empty
// as well as a 0
func (m IdManager) Pop() (uint8, error) {
	if m.IsEmpty() {
		return 0, NewIdManagerError("empty")
	}
	val := m.available.Front()
	// this is fast as the value will be in fist position anyway
	m.available.Remove(val)
	return val.Value.(uint8), nil
}

func (m IdManager) IsEmpty() bool {
	return m.available.Len() == 0
}

func (m IdManager) IsFull() bool {
	const max = math.MaxUint8 + 1
	return m.available.Len() == max
}
