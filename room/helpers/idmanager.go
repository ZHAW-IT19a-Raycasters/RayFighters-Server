package helpers

import (
	"errors"
	"math"
)

// IdManagerError is the error produced, if during a function call a possibly invalid state
// would have been reached.
var (
	ErrIdManagerFull  = errors.New("incorrect handling of the idManager, it is filled up")
	ErrIdManagerEmpty = errors.New("incorrect handling of the idManager, is is empty")
)

// IdCountLimit is the limit of possible id handled
// by the id manager.
const IdCountLimit int = math.MaxUint8 + 1

// IdManager is an abstraction over the standard library
// linked list implementation for uint8 values.
type IdManager struct {
	// The list of currently ids user ids
	// if this is empty all ids have been used.
	ids chan uint8
}

// Will create a new idManager and fills it up with all values for uint8.
func NewIdManager() IdManager {
	ids := make(chan uint8, IdCountLimit)
	for i := 0; i <= IdCountLimit-1; i += 1 {
		// make sure that only uint8 values are pushed into the
		// queue
		ids <- uint8(i)
	}
	return IdManager{
		ids: ids,
	}
}

// Will append the value to the back of the list.
// Can return an error, if the list is full.
func (m IdManager) Push(val uint8) error {
	select {
	case m.ids <- val:
		return nil
	default:
		return ErrIdManagerFull
	}
}

// Will return the currently first value of the queue.
// Attention: Will return an error if the queue is empty
// as well as a 0
func (m IdManager) Pop() (uint8, error) {
	select {
	case id := <-m.ids:
		return id, nil
	default:
		return 0, ErrIdManagerEmpty
	}
}

// Count will return the current amount of ids left.
func (m IdManager) Count() int {
	return len(m.ids)
}

// IsEmpty will check if there are still some ids left
func (m IdManager) IsEmpty() bool {
	return len(m.ids) == 0
}

// IsFull will check if there is still space for more ids.
func (m IdManager) IsFull() bool {
	return len(m.ids) == IdCountLimit
}
