package gamelocation

import (
	"os"
	"path/filepath"
	"testing"

	"github.zhaw.ch/raycasters/server/gamelocation/maploader"
	"github.zhaw.ch/raycasters/server/gamelocation/utils"
)

// Test if passing correct parameters, will raise no erros.
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
	actual := utils.Compare2DMapElementSlices(expectedMap, loc.map2d)
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

//Check if reading bigger map will not raise an error.
func TestGetGameLocationValid(t *testing.T) {

	path, err := os.Getwd()
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}
	path = filepath.Join(path, "maploader/test/testmaps/testmap_07_valid_big.txt")
	file, err := os.Open(path)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}
	defer file.Close()
	_, err = maploader.ParseMap(path)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}
}

//Check if getter returns expected dimension.
func TestGetMapDimension(t *testing.T) {
	const expected int = 2
	locationMap := [][]utils.LocationElement{{utils.Floor, utils.Wall}, {utils.Floor, utils.Floor}}
	loc, err := getNewLocation(expected, locationMap)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}
	actual := loc.GetMapDimension()
	if actual != expected {
		t.Errorf("wrong dimesion: (expected: %d; actual: %d)", expected, loc.dimension)
		return
	}
}

//Check if getter returns expected map elment.
func TestGetLocationElement(t *testing.T) {
	const dim int = 2
	locationMap := [][]utils.LocationElement{{utils.Floor, utils.Wall}, {utils.Floor, utils.Floor}}
	loc, err := getNewLocation(dim, locationMap)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}
	expected := utils.Wall
	actual, err := loc.GetLocationElement(0, 1)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}
	if actual != expected {
		t.Errorf("wrong map element: (expected: %v; actual: %v)", expected, actual)
		return
	}
}

// Test if comparing two equal locations is correct.
func TestCompareSameLocations(t *testing.T) {
	const dim int = 2
	locationMap := [][]utils.LocationElement{{utils.Floor, utils.Wall}, {utils.Floor, utils.Floor}}
	loc1, err := getNewLocation(dim, locationMap)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}
	loc2, err := getNewLocation(dim, locationMap)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}
	expected := true
	actual := loc1.Compare(loc2)
	if expected != actual {
		t.Errorf("wrong resulst: (expected: %t; actual: %t)", expected, actual)
	}
}

// Test if comparing two distinct locations is correct.
func TestCompareDifferentLocations(t *testing.T) {
	const dim int = 2
	locationMap1 := [][]utils.LocationElement{{utils.Floor, utils.Wall}, {utils.Floor, utils.Floor}}
	loc1, err := getNewLocation(dim, locationMap1)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}
	locationMap2 := [][]utils.LocationElement{{utils.Wall, utils.Wall}, {utils.Floor, utils.Floor}}
	loc2, err := getNewLocation(dim, locationMap2)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}
	expected := false
	actual := loc1.Compare(loc2)
	if expected != actual {
		t.Errorf("wrong resulst: (expected: %t; actual: %t)", expected, actual)
	}
}
