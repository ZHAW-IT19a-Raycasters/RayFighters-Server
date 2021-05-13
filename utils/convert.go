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

// Int64ToBinary will change the int64 into
// its binary format.
func Int64ToBinary(num int64, data []byte) {
	binary.LittleEndian.PutUint64(data, uint64(num))
}

// BinaryToInt64 will parse the binary into its
// integer format.
// Attention this is only used for testing.
func BinaryToInt64(data []byte) int64 {
	return int64(binary.LittleEndian.Uint64(data))
}
