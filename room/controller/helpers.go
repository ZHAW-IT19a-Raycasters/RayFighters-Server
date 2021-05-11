package controller

import (
	"fmt"

	"github.zhaw.ch/raycasters/server/room/client"
)

// ErrClientUnexpectedState will be returned in cases
// where the specified client is not in a timed out
// state.
type ErrClientUnexpectedState struct {
	state client.ConnectionStatus
	id    uint8
}

// Error satisfies the error
func (err ErrClientUnexpectedState) Error() string {
	return fmt.Sprintf("the client [%d] is currently not in the expected state [%d]", err.id, err.state)
}

// Is satisfies the error.Is configuration
func (err ErrClientUnexpectedState) Is(target error) bool {
	_, ok := target.(ErrClientUnexpectedState)
	return ok
}

// ErrClientDoesNotExist will be returned in cases
// where the specified client does not yet exist.
type ErrClientDoesNotExist struct {
	id uint8
}

// Error satisfies the error
func (err ErrClientDoesNotExist) Error() string {
	res := fmt.Sprintf("the given client [%d] does currently not exists", err.id)
	return res
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
