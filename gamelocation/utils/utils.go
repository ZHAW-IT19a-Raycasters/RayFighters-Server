package utils

// Enum used for mapping runes to internal codes.
type LocationElement int

const (
	Floor LocationElement = iota + 1
	Wall
	MedkitSpawn
	WeaponSpawn
)

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

	for i, _ := range sliceOne {
		if !Compare1DMapElementSlices(sliceOne[i], sliceTwo[i]) {
			return false
		}
	}
	return true
}
