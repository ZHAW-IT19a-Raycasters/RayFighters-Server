package helpers

import "errors"

type empty struct{}
type semaphoreChan chan empty

var ErrSemaphoreAtLimit = errors.New("the atomic variable is all ready at it's limit")

type Semaphore struct {
	sem     semaphoreChan
	counter AtomicUint32
	limit   uint32
}

func NewSemaphore(limit uint32) Semaphore {
	return Semaphore{
		sem:     make(semaphoreChan, limit),
		counter: NewAtomicUint32(limit),
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

func (sem *Semaphore) AtZero() bool {
	return sem.counter.AtZero()
}

func (sem *Semaphore) AtLimit() bool {
	return sem.counter.AtLimit()
}

// Inc will increment the counter by 1.
// This will return an error if the counter already was at
// it's limits.
func (sem *Semaphore) Inc() error {
	select {
	case sem.sem <- empty{}:
		sem.counter.Inc()
		return nil
	default:
		return ErrSemaphoreAtLimit
	}
}

// Dec will decrement the counter by until 0.
// If there is no more already is zero, this function
// will return an error.
func (sem *Semaphore) Dec() error {
	select {
	case <-sem.sem:
		sem.counter.Dec()
		return nil
	default:
		return ErrSemaphoreAtLimit
	}
}
