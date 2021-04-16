package helpers

import "fmt"

// ErrIdAlready the id has already been added
type ErrIdAlready struct {
	id uint8
}

// Error is the error interface, so that the value can be returned
func (e ErrIdAlready) Error() string {
	return fmt.Sprintf("such id already available [%d]", e.id)
}

// Is satisfies the errors.Is definition
func (e ErrIdAlready) Is(target error) bool {
	_, ok := target.(ErrIdAlready)
	return ok
}

// ErrNoId the id does not exist in the set
type ErrNoId struct {
	id uint8
}

// Error is the error interface, so that the value can be returned
func (e ErrNoId) Error() string {
	return fmt.Sprintf("no such id available [%d]", e.id)
}

// Is satisfies the errors.Is definition
func (e ErrNoId) Is(target error) bool {
	_, ok := target.(ErrNoId)
	return ok
}

// void is an empty type, so that the set values do not
// use up any additional memory.
type void struct{}

// IdSet a set build for the client ids
type IdSet struct {
	data map[uint8]void
}

// NewIdSet creates a new set
func NewIdSet() IdSet {
	return IdSet{
		data: make(map[uint8]void),
	}
}

// Add adds a new id to the set, but will return an
// error if the id already existed
func (is *IdSet) Add(id uint8) error {
	if is.Has(id) {
		return ErrIdAlready{id}
	}

	is.data[id] = void{}

	return nil
}

// Remove will remove an id if it did
// exist else if will retun an error
func (is *IdSet) Remove(id uint8) error {
	if !is.Has(id) {
		return ErrNoId{id}
	}
	delete(is.data, id)
	return nil
}

// Has checks if the map has the id or not
func (is *IdSet) Has(id uint8) bool {
	_, ok := is.data[id]
	return ok
}
