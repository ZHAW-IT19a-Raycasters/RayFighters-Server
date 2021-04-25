package mapstate

import "fmt"

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

// ErrWalkedIntoWall is the error that will be returned
// if the given client tries to walk into wall.
type ErrWalkedIntoWall struct {
	id uint8
}

// Error satisfies the error
func (err ErrWalkedIntoWall) Error() string {
	return fmt.Sprintf("the given client [%d] tried to walk into wall", err.id)
}

// Is satisfies the error.Is configuration
func (err ErrWalkedIntoWall) Is(target error) bool {
	_, ok := target.(ErrWalkedIntoWall)
	return ok
}

// ErrWalkedIntoWall is the error that will be returned
// if the given client tries to walk into wall.
type ErrClientDoesNotExist struct {
	id uint8
}

// Error satisfies the error
func (err ErrClientDoesNotExist) Error() string {
	return fmt.Sprintf("the given client [%d] tried to walk into wall", err.id)
}

// Is satisfies the error.Is configuration
func (err ErrClientDoesNotExist) Is(target error) bool {
	_, ok := target.(ErrClientDoesNotExist)
	return ok
}

