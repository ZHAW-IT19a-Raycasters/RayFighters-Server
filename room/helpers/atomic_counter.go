package helpers

import (
	"errors"
	"sync/atomic"
)

// ErrAtomicAtLimit will be returned in the case that the counters limits
// have been reached.
var ErrAtomicAtLimit = errors.New("the atomic variable is all ready at it's limit")

// AtomicUint32 will store the atomic value, as
// well as the counters limit.
type AtomicUint32 struct {
	value uint32
	limit uint32
}

// NewAtomicUint32 will create an atomic counter struct.
func NewAtomicUint32(limit uint32) AtomicUint32 {
	return AtomicUint32{
		value: 0,
		limit: limit,
	}
}

// Value will return the current state of the variable.
func (counter *AtomicUint32) Value() uint32 {
	return atomic.LoadUint32(&counter.value)
}

// Limit will return the set limit.
func (counter *AtomicUint32) Limit() uint32 {
	return counter.limit
}

// AtLimit will return true, if the counter reached the
// given limit.
func (counter *AtomicUint32) AtLimit() bool {
	return atomic.CompareAndSwapUint32(&counter.value, counter.limit, counter.limit)
}

// AtZero will return true, if the counter is at it's
// lower limit a zero.
func (counter *AtomicUint32) AtZero() bool {
	return atomic.CompareAndSwapUint32(&counter.value, 0, counter.value)
}

// Inc will increment the counter by 1.
// This will return an error if the counter already was at
// it's limits.
func (counter *AtomicUint32) Inc() error {
	if counter.AtLimit() {
		return ErrAtomicAtLimit
	}
	atomic.AddUint32(&counter.value, 1)
	return nil
}

// Dec will decrement the counter by until 0.
// If the value already is zero, this function
// will return an error.
func (counter *AtomicUint32) Dec() error {
	if counter.AtZero() {
		return ErrAtomicAtLimit
	}
	// To subtract a signed positive constant value c from x, do
	// AddUint32(&x, ^uint32(c-1)).

	// c = 1 => c - 1 = 0
	atomic.AddUint32(&counter.value, ^uint32(0))
	return nil
}
