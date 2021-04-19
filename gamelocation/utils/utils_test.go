package utils_test

import (
	"testing"

	"github.zhaw.ch/raycasters/server/gamelocation/utils"
)

// Check if slices of same dimesions and having same elements are equal.
func Test1DMapElementSlicesSameDimesionSameElements(t *testing.T) {
	slice1 := []utils.LocationElement{utils.Floor, utils.Floor}
	slice2 := []utils.LocationElement{utils.Floor, utils.Floor}
	expected := true
	actual := utils.Compare1DMapElementSlices(slice1, slice2)
	if expected != actual {
		t.Errorf("false result: (expected: %t; actual: %t)", expected, actual)
	}
}

// Check if two empty-slices are equal.
func Test1DMapElementSlicesEmptySlices(t *testing.T) {
	slice1 := []utils.LocationElement{}
	slice2 := []utils.LocationElement{}
	expected := true
	actual := utils.Compare1DMapElementSlices(slice1, slice2)
	if expected != actual {
		t.Errorf("false result: (expected: %t; actual: %t)", expected, actual)
	}
}

// Check if two nil-slices are equal.
func Test1DMapElementSlicesNilSlices(t *testing.T) {
	var slice1 []utils.LocationElement
	var slice2 []utils.LocationElement
	expected := true
	actual := utils.Compare1DMapElementSlices(slice1, slice2)
	if expected != actual {
		t.Errorf("false result: (expected: %t; actual: %t)", expected, actual)
	}
}

// Check if slices having same dimesion but different elements are non-equal
func Test1DMapElementSlicesDifferentElementsSameDimensions(t *testing.T) {
	slice1 := []utils.LocationElement{utils.Floor, utils.Floor}
	slice2 := []utils.LocationElement{utils.Floor, utils.Wall}
	expected := false
	actual := utils.Compare1DMapElementSlices(slice1, slice2)
	if expected != actual {
		t.Errorf("false result: (expected: %t; actual: %t)", expected, actual)
	}
}

//Check if slices having different dimesions but same elemnents are non-equal.
func Test1DMapElementSlicesDifferentDimensionsSameElments(t *testing.T) {
	slice1 := []utils.LocationElement{utils.Floor, utils.Floor}
	slice2 := []utils.LocationElement{utils.Floor, utils.Floor, utils.Floor}
	expected := false
	actual := utils.Compare1DMapElementSlices(slice1, slice2)
	if expected != actual {
		t.Errorf("false result: (expected: %t; actual: %t)", expected, actual)
	}
}

//Check if slices having different dimesions and different elemnents are non-equal.
func Test1DMapElementSlicesDifferentDimensionsDifferentElments(t *testing.T) {
	slice1 := []utils.LocationElement{utils.Floor, utils.Floor}
	slice2 := []utils.LocationElement{utils.Wall, utils.Floor, utils.Floor}
	expected := false
	actual := utils.Compare1DMapElementSlices(slice1, slice2)
	if expected != actual {
		t.Errorf("false result: (expected: %t; actual: %t)", expected, actual)
	}
}

// Check if slices of same dimesions and having same elements are equal.
func Test2DMapElementSlicesSameDimesionSameElements(t *testing.T) {
	slice1 := [][]utils.LocationElement{{utils.Floor, utils.Floor}, {utils.Floor, utils.Floor}}
	slice2 := [][]utils.LocationElement{{utils.Floor, utils.Floor}, {utils.Floor, utils.Floor}}
	expected := true
	actual := utils.Compare2DMapElementSlices(slice1, slice2)
	if expected != actual {
		t.Errorf("false result: (expected: %t; actual: %t)", expected, actual)
	}
}

// Check if two empty-slices are equal.
func Test2DMapElementSlicesEmptySlices(t *testing.T) {
	slice1 := [][]utils.LocationElement{}
	slice2 := [][]utils.LocationElement{}
	expected := true
	actual := utils.Compare2DMapElementSlices(slice1, slice2)
	if expected != actual {
		t.Errorf("false result: (expected: %t; actual: %t)", expected, actual)
	}
}

// Check if two nil-slices are equal.
func Test2DMapElementSlicesNilSlices(t *testing.T) {
	var slice1 [][]utils.LocationElement
	var slice2 [][]utils.LocationElement
	expected := true
	actual := utils.Compare2DMapElementSlices(slice1, slice2)
	if expected != actual {
		t.Errorf("false result: (expected: %t; actual: %t)", expected, actual)
	}
}

// Check if slices having same dimesion but different elements are non-equal
func Test2DMapElementSlicesDifferentElementsSameDimensions(t *testing.T) {
	slice1 := [][]utils.LocationElement{{utils.Floor, utils.Floor}, {utils.Floor, utils.Floor}}
	slice2 := [][]utils.LocationElement{{utils.Floor, utils.Floor}, {utils.Floor, utils.Wall}}
	expected := false
	actual := utils.Compare2DMapElementSlices(slice1, slice2)
	if expected != actual {
		t.Errorf("false result: (expected: %t; actual: %t)", expected, actual)
	}
}

//Check if slices having different dimesions but same elemnents are non-equal.
func Test2DMapElementSlicesDifferentDimensionsSameElments(t *testing.T) {
	slice1 := [][]utils.LocationElement{{utils.Floor, utils.Floor}, {utils.Floor, utils.Floor}}
	slice2 := [][]utils.LocationElement{{utils.Floor, utils.Floor}, {utils.Floor, utils.Floor, utils.Floor}}
	expected := false
	actual := utils.Compare2DMapElementSlices(slice1, slice2)
	if expected != actual {
		t.Errorf("false result: (expected: %t; actual: %t)", expected, actual)
	}
}

//Check if slices having different dimesions and different elemnents are non-equal.
func Test2DMapElementSlicesDifferentDimensionsDifferentElments(t *testing.T) {
	slice1 := [][]utils.LocationElement{{utils.Floor, utils.Floor}, {utils.Floor, utils.Wall}}
	slice2 := [][]utils.LocationElement{{utils.Floor, utils.Wall}, {utils.Floor, utils.Floor, utils.Floor}}
	expected := false
	actual := utils.Compare2DMapElementSlices(slice1, slice2)
	if expected != actual {
		t.Errorf("false result: (expected: %t; actual: %t)", expected, actual)
	}
}

func TestGetLocationElementValidSymbol(t *testing.T) {
	expected := utils.Floor
	actual, err := utils.GetLocationElement(rune('0'))
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
	}
	if expected != actual {
		t.Errorf("false result: (expected: %v; actual: %v)", expected, actual)
	}
}

func TestGetLocationElementInvalidSymbol(t *testing.T) {
	_, err := utils.GetLocationElement(rune('/'))
	if err == nil {
		t.Error("no error was thrown")
	}
}

func TestGetLocationElementValidWallSymbol(t *testing.T) {
	expected := utils.Floor
	actual, err := utils.GetLocationElement(rune('0'))
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
	}
	if expected != actual || err != nil {
		t.Errorf("false result: (expected: %v; actual: %v)", expected, actual)
	}
}
