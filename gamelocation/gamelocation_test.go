package gamelocation

import (
	"testing"

	"github.zhaw.ch/raycasters/server/gamelocation/utils"
)

// Test if having correct parameters no errors will be thrown.
func TestGetNewLocationValid(t *testing.T) {
	const dim int = 2
	expectedMap := [][]utils.LocationElement{{utils.Floor, utils.Floor}, {utils.Floor, utils.Floor}}
	loc, err := getNewLocation(dim, expectedMap)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}
	if loc.dimension != dim {
		t.Errorf("wrong dimesion: (expected: %d; actual: %d)", dim, loc.dimension)
		return
	}
	expected := true
	actual := utils.Compare2DMapElementSlices(expectedMap, loc.locationMap)
	if expected != actual {
		t.Errorf("wrong dimesion: (expected: %d; actual: %d)", dim, loc.dimension)
		return
	}
}

// Test if passing incorrect map dimension will rais an error.
func TestGetNewLocationInvalidMapDimesion(t *testing.T) {
	const dim int = 2
	expectedMap := [][]utils.LocationElement{{utils.Floor, utils.Floor}, {utils.Floor}}
	_, err := getNewLocation(dim, expectedMap)
	if err == nil {
		t.Errorf("no error was thrown")
	}
}
