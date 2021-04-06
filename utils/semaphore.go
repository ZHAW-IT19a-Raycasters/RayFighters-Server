package utils

import (
	"errors"
)

type empty struct{}
type semaphoreChan chan empty

var ErrSemaphoreAtLimit = errors.New("the atomic variable is all ready at it's limit")

type Semaphore struct {
	sem     semaphoreChan
	counter uint32
	limit   uint32
}

func NewSemaphore(limit uint32) Semaphore {
	return Semaphore{
		sem:     make(semaphoreChan, limit),
		counter: 0,
		limit:   limit,
	}
}

// Value will return the current value of the semaphore
// Attention this is state nto check if there
// is still space left
func (sem *Semaphore) Value() int {
	return len(sem.sem)
}

// Limit will return the set limit.
func (sem *Semaphore) Limit() uint32 {
	return sem.limit
}

// Empty will return true, if the counter is at it's
// lower limit a zero.
// Attention this might not represents the actuall value
// due to non atomic operations.
func (sem *Semaphore) Empty() bool {
	return len(sem.sem) == 0
}

// Full will return true, if the counter reached the
// given limit.
// Attention this might not represents the actuall value
// due to non atomic operations.
func (sem *Semaphore) Full() bool {
	return len(sem.sem) == int(sem.limit)
}

// Acquire will increment the counter by 1.
// This will return an error if the counter already was at
// it's limits.
func (sem *Semaphore) Acquire() error {
	select {
	case sem.sem <- empty{}:
		return nil
	default:
		return ErrSemaphoreAtLimit
	}
}

// Release will decrement the counter by until 0.
// If there is no more already is zero, this function
// will return an error.
func (sem *Semaphore) Release() error {
	select {
	case <-sem.sem:
		return nil
	default:
		return ErrSemaphoreAtLimit
	}
}
