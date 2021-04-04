package helpers

import "sync/atomic"

const (
	// False is the internal representation of
	// the mapping from int to false.
	False uint32 = 0
	// True is the internal representation of
	// the mapping from int to the true bool.
	// value.
	True uint32 = 1
)

// MapIntToBool will convert an uint32 to a boolean.
// Attention this function will panic, if the value
// is anything else then 1 or 0.
func MapIntToBool(value uint32) bool {
	if value == True {
		return true
	} else if value == False {
		return false
	}
	panic("There is no mapping for anything else then a 0 or 1")
}

// MapBoolToInt will convert the bool into the respective
// integer representation.
func MapBoolToInt(value bool) uint32 {
	if value {
		return True
	}
	return False
}

// AtomicBool is an atomic boolean value representation.
type AtomicBool struct {
	// As golang does not have an atomic boolean
	// type, is is mapped via value.
	// 1 is mapped to true
	// 0 is mapped to false
	value uint32
}

// NewAtomicBool will create an atomic bool.
func NewAtomicBool(value bool) AtomicBool {
	return AtomicBool{
		value: MapBoolToInt(value),
	}
}

// Get will return the current value of the atomic value.
func (counter *AtomicBool) Get() bool {
	return MapIntToBool(counter.value)
}

// Set will replace the current value and return
// the replaced one.
func (counter *AtomicBool) Set(value bool) bool {
	old := atomic.SwapUint32(&counter.value, MapBoolToInt(value))
	return MapIntToBool(old)
}

// SetTrue will set the value to true and if there was a change to the value
// will return true, else false.
func (counter *AtomicBool) SetTrue() bool {
	// if there was a swap => false to True
	// this will return true, else false
	return atomic.CompareAndSwapUint32(&counter.value, False, True)
}

// SetFalse will set the value to false and if there was a change to
// the value will return true, else false.
func (counter *AtomicBool) SetFalse() bool {
	// if there was a swap => false to True
	// this will return true, else false
	return atomic.CompareAndSwapUint32(&counter.value, True, False)
}
