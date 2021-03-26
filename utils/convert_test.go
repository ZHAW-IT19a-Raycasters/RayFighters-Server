package utils_test

import (
	"testing"

	"github.zhaw.ch/raycasters/server/utils"
)

func compareByteArrays(expectedData, gottenData []byte, t *testing.T) bool {

	if len(expectedData) != len(gottenData) {
		t.Errorf("Different lenght of results [%d] - [%d]", len(expectedData), len(gottenData))
		return false
	}

	for i := range expectedData {
		exp := expectedData[i]
		got := gottenData[i]
		if exp != got {

			t.Errorf("Bytes at position [%d] are different [%d] - [%d]", i, exp, got)
			return false
		}
	}

	return true
}

func TestFloatToBinary(t *testing.T) {

	// 1.00000000000000005551115123126E-1
	// 0x3FB999999999999A
	var coorX float64 = 0.1
	expData := [...]byte{
		// 0x3F 0xB9 0x99 0x99 0x99 0x99 0x99 0x9A
		0x9A, 0x99, 0x99, 0x99, 0x99, 0x99, 0xB9, 0x3F,
	}
	gotData := make([]byte, 8)
	utils.FloatToBinary(coorX, gotData)

	compareByteArrays(expData[:], gotData, t)
}

func TestBinaryToFloat(t *testing.T) {

	// 1.00000000000000005551115123126E-1
	// 0x3FB999999999999A
	var expected float64 = 0.1

	rawData := [...]byte{
		// 0x3F 0xB9 0x99 0x99 0x99 0x99 0x99 0x9A
		0x9A, 0x99, 0x99, 0x99, 0x99, 0x99, 0xB9, 0x3F,
	}

	got := utils.BinaryToFloat(rawData[:])
	if got != expected {
		t.Errorf("incorrect conversion of number")
	}
}
