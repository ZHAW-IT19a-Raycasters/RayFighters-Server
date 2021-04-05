package helpers

import (
	"errors"
	"sync/atomic"
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
func (sem *Semaphore) Empty() bool {
	return atomic.CompareAndSwapUint32(&sem.counter, 0, sem.counter)
}

// Full will return true, if the counter reached the
// given limit.
func (sem *Semaphore) Full() bool {
	return atomic.CompareAndSwapUint32(&sem.counter, sem.limit, sem.limit)
}

// Acquire will increment the counter by 1.
// This will return an error if the counter already was at
// it's limits.
func (sem *Semaphore) Acquire() error {
	select {
	case sem.sem <- empty{}:
		atomic.AddUint32(&sem.counter, 1)
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
		// To subtract a signed positive constant value c from x, do
		// AddUint32(&x, ^uint32(c-1)).
		atomic.AddUint32(&sem.counter, ^uint32(0))
		return nil
	default:
		return ErrSemaphoreAtLimit
	}
}
