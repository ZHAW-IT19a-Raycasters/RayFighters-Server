package controller

import "fmt"

// ErrClientNotTimedOut will be returned in cases
// where the specified client is not in a timed out
// state.
type ErrClientNotTimedOut struct {
	id uint8
}

// Error satisfies the error
func (err ErrClientNotTimedOut) Error() string {
	return fmt.Sprintf("the client [%d] is currently not timedout", err.id)
}

// Is satisfies the error.Is configuration
func (err ErrClientNotTimedOut) Is(target error) bool {
	_, ok := target.(ErrClientNotTimedOut)
	return ok
}

// ErrClientDoesNotExist will be returned in cases
// where the specified client does not yet exist.
type ErrClientDoesNotExist struct {
	id uint8
}

// Error satisfies the error
func (err ErrClientDoesNotExist) Error() string {
	return fmt.Sprintf("the given client [%d] does currently not exists", err.id)
}

// Is satisfies the error.Is configuration
func (err ErrClientDoesNotExist) Is(target error) bool {
	_, ok := target.(ErrClientDoesNotExist)
	return ok
}

// ErrClientExists is the error that will be returned
// if the given client does already exists.
type ErrClientExists struct {
	id uint8
}

// Error satisfies the error
func (err ErrClientExists) Error() string {
	return fmt.Sprintf("the given client [%d] already exists", err.id)
}

// Is satisfies the error.Is configuration
func (err ErrClientExists) Is(target error) bool {
	_, ok := target.(ErrClientExists)
	return ok
}
