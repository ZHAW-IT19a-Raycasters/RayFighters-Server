package room

import "fmt"

// ErrHubExists is the error that will be returned
// if the given client does already exists.
type ErrHubExists struct {
	id int
}

// Error satisfies the error
func (err ErrHubExists) Error() string {
	return fmt.Sprintf("the given hub [%d] already exists", err.id)
}

// Is satisfies the error.Is configuration
func (err ErrHubExists) Is(target error) bool {
	_, ok := target.(ErrHubExists)
	return ok
}

// ErrHubNotExisting is the error that will be returned
// if the given client does already exists.
type ErrHubNotExisting struct {
	id int
}

// Error satisfies the error
func (err ErrHubNotExisting) Error() string {
	return fmt.Sprintf("the given hub [%d] does currently not exists", err.id)
}

// Is satisfies the error.Is configuration
func (err ErrHubNotExisting) Is(target error) bool {
	_, ok := target.(ErrHubNotExisting)
	return ok
}
