package gamemap

import (
	"errors"
	"fmt"
)

// Error codes returned by failures in gamemap.
var (
	ErrMeWrSym = errors.New("mapelement: unsupported symbol")
)

// MapElement is enum representation of elements of game map.
type MapElement int

const (
	Floor MapElement = iota + 1
	Wall
	Medkit
	Weapon
)

// convert return corresponding MapElement for given rune.
func convert(r rune) (MapElement, error) {
	switch {
	case r == '0':
		return Floor, nil
	case r == '+':
		return Medkit, nil
	case r == '*':
		return Weapon, nil
	case (r >= rune('1') && r <= rune('9')) || (r >= rune('A') && r <= rune('Z')):
		return Wall, nil
	case (r >= rune('!') && r <= rune('~')):
		return Floor, nil
	default:
		return 0, fmt.Errorf("%s: %s ", ErrMeWrSym, string(r))
	}
}

// equals1d compares two 1d slices of MapElements.
// Returns true if either A: (both are nil) or B: (both have same length and same elements for each position).
func equals1d(sliceOne, sliceTwo []MapElement) bool {
	// If one is nil, the other must also be nil.
	if (sliceOne == nil) != (sliceTwo == nil) {
		return false
	}

	if len(sliceOne) != len(sliceTwo) {
		return false
	}

	for i, e := range sliceOne {
		if e != sliceTwo[i] {
			return false
		}
	}

	return true
}

// equals2d compares two 2d slices of MapElements.
// Returns true if either A: (both are nil) or B: (both have same dimensions and same elements for each position).
func equals2d(sliceOne, sliceTwo [][]MapElement) bool {
	// If one is nil, the other must also be nil.
	if (sliceOne == nil) != (sliceTwo == nil) {
		return false
	}

	if len(sliceOne) != len(sliceTwo) {
		return false
	}

	for i := range sliceOne {
		if !equals1d(sliceOne[i], sliceTwo[i]) {
			return false
		}
	}
	return true
}
