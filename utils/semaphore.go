package utils

import (
	"errors"
)

// void is an empty type that will not use up
// any memory.
type void struct{}

// semaphoreChan is an abstraction to simplify
// the written code. It represents a counting
// channel
type semaphoreChan chan void

// ErrSemaphoreAtLimit will be returned once there is
// no more space left for a new arrival
var ErrSemaphoreAtLimit = errors.New("the semaphore is all ready at it's limit")

// CountingSemaphore is the actually structure to interact
// with. It is a semaphore to be precise a counting
// semaphore.
// see https://en.wikipedia.org/wiki/Semaphore_(programming)
// for more information about semaphores.
type CountingSemaphore struct {
	// sem is the channel that is responsible for
	// counting the values
	sem semaphoreChan
	// limit is the set limit of value
	// up to which this semaphore can
	// in theory count.
	limit uint32
}

// NewCountingSemaphore will initialize a new counting
// semaphore.
func NewCountingSemaphore(limit uint32) CountingSemaphore {
	return CountingSemaphore{
		sem:   make(semaphoreChan, limit),
		limit: limit,
	}
}

// Value will return the current value of the semaphore
// Attention do not used this to check if there is
// is still space left. As the value might not be
// up to date.
func (sem *CountingSemaphore) Value() int {
	return len(sem.sem)
}

// Limit will return the set limit.
func (sem *CountingSemaphore) Limit() uint32 {
	return sem.limit
}

// Empty will return true, if the counter is at it's
// lower limit a zero.
// Attention this might not represents the actual value
// due to non atomic operations.
func (sem *CountingSemaphore) Empty() bool {
	return len(sem.sem) == 0
}

// Full will return true, if the counter reached the
// given limit.
// Attention this might not represents the actual value
// due to non atomic operations.
func (sem *CountingSemaphore) Full() bool {
	return len(sem.sem) == int(sem.limit)
}

// Acquire will increment the counter by 1.
// This will return an error if the counter already was at
// it's limits.
func (sem *CountingSemaphore) Acquire() error {
	select {
	case sem.sem <- void{}:
		return nil
	default:
		return ErrSemaphoreAtLimit
	}
}

// Release will decrement the counter by until 0.
// If there is no more already is zero, this function
// will return an error.
func (sem *CountingSemaphore) Release() error {
	select {
	case <-sem.sem:
		return nil
	default:
		return ErrSemaphoreAtLimit
	}
}
