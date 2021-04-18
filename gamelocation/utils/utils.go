package utils

import "fmt"

// Enum used for mapping runes to internal codes.
type LocationElement int

const (
	Floor LocationElement = iota + 1
	Wall
	MedkitSpawn
	WeaponSpawn
)

// Maps golang "rune" into "LocationElement".
func GetLocationElement(r rune) (LocationElement, error) {
	switch {
	case r == '0':
		return Floor, nil
	case r == '+':
		return MedkitSpawn, nil
	case r == '*':
		return WeaponSpawn, nil
	case (r >= rune('1') && r <= rune('9')) || (r >= rune('A') && r <= rune('Z')):
		return Wall, nil
	default:
		return 0, fmt.Errorf("unsupported symbol: %v ", r)
	}
}

func Compare1DMapElementSlices(sliceOne, sliceTwo []LocationElement) bool {
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

func Compare2DMapElementSlices(sliceOne, sliceTwo [][]LocationElement) bool {
	// If one is nil, the other must also be nil.
	if (sliceOne == nil) != (sliceTwo == nil) {
		return false
	}
	if len(sliceOne) != len(sliceTwo) {
		return false
	}

	for i := range sliceOne {
		if !Compare1DMapElementSlices(sliceOne[i], sliceTwo[i]) {
			return false
		}
	}
	return true
}
