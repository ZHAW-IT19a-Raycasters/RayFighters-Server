package utils

import (
	"encoding/binary"
	"math"
)

// Will convert a binary blob into a correct float64.
// Attention this assumes that the float starts from
// position 0 on and that the float is formatted
// according to the IEEE specification.
func BinaryToFloat(data []byte) float64 {

	// convert the float into the int64 representation
	intRep := binary.LittleEndian.Uint64(data)

	// convert the raw int representation to
	return math.Float64frombits(intRep)
}

// Will insert the num into data, starting
// from the index 0 on. It will utilize
// the IEEE specification.
// Attention this function assumes that
// the data is already initialized.
func FloatToBinary(num float64, data []byte) {

	// convert the float into the int64 representation
	intRep := math.Float64bits(num)

	// write data back
	binary.LittleEndian.PutUint64(data, intRep)
}
