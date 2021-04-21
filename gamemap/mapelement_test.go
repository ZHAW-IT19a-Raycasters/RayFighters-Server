package gamemap

import (
	"testing"
)

// Check if slices of same dimesions and having same elements are equal.
func TestEquals1dSameDimesionSameElms(t *testing.T) {
	slice1 := []MapElement{Floor, Floor}
	slice2 := []MapElement{Floor, Floor}
	expected := true
	actual := equals1d(slice1, slice2)
	if expected != actual {
		t.Errorf("false result: (expected: %t; actual: %t)", expected, actual)
	}
}

// Check if two empty-slices are equal.
func TestEquals1dEmptySlices(t *testing.T) {
	slice1 := []MapElement{}
	slice2 := []MapElement{}
	expected := true
	actual := equals1d(slice1, slice2)
	if expected != actual {
		t.Errorf("false result: (expected: %t; actual: %t)", expected, actual)
	}
}

// Check if two nil-slices are equal.
func TestEquals1dNilSlices(t *testing.T) {
	expected := true
	actual := equals1d(nil, nil)
	if expected != actual {
		t.Errorf("false result: (expected: %t; actual: %t)", expected, actual)
	}
}

func TestEquals1dOneNilSlice(t *testing.T) {
	expected := false
	actual := equals1d(nil, []MapElement{})
	if expected != actual {
		t.Errorf("false result: (expected: %t; actual: %t)", expected, actual)
	}
}

// Check if slices having same dimesion but different elements are non-equal
func TestEquals1dDiffElmsSameDimensions(t *testing.T) {
	slice1 := []MapElement{Floor, Floor}
	slice2 := []MapElement{Floor, Wall}
	expected := false
	actual := equals1d(slice1, slice2)
	if expected != actual {
		t.Errorf("false result: (expected: %t; actual: %t)", expected, actual)
	}
}

//Check if slices having different dimesions but same elemnents are non-equal.
func TestEquals1dDiffDimensionsSameElms(t *testing.T) {
	slice1 := []MapElement{Floor, Floor}
	slice2 := []MapElement{Floor, Floor, Floor}
	expected := false
	actual := equals1d(slice1, slice2)
	if expected != actual {
		t.Errorf("false result: (expected: %t; actual: %t)", expected, actual)
	}
}

//Check if slices having different dimesions and different elemnents are non-equal.
func TestEquals1dDiffDimensionsDiffElms(t *testing.T) {
	slice1 := []MapElement{MapElement(Floor), Floor}
	slice2 := []MapElement{Wall, Floor, Floor}
	expected := false
	actual := equals1d(slice1, slice2)
	if expected != actual {
		t.Errorf("false result: (expected: %t; actual: %t)", expected, actual)
	}
}

// Check if slices of same dimesions and having same elements are equal.
func TestEquals2dSameDimesionSameElms(t *testing.T) {
	slice1 := [][]MapElement{{Floor, Floor}, {Floor, Floor}}
	slice2 := [][]MapElement{{Floor, Floor}, {Floor, Floor}}
	expected := true
	actual := equals2d(slice1, slice2)
	if expected != actual {
		t.Errorf("false result: (expected: %t; actual: %t)", expected, actual)
	}
}

// Check if two empty-slices are equal.
func TestEquals2dEmptySlices(t *testing.T) {
	slice1 := [][]MapElement{}
	slice2 := [][]MapElement{}
	expected := true
	actual := equals2d(slice1, slice2)
	if expected != actual {
		t.Errorf("false result: (expected: %t; actual: %t)", expected, actual)
	}
}

// Check if two nil-slices are equal.
func TestEquals2dNilSlices(t *testing.T) {
	expected := true
	actual := equals2d(nil, nil)
	if expected != actual {
		t.Errorf("false result: (expected: %t; actual: %t)", expected, actual)
	}
}

// Check if two nil-slices are equal.
func TestEquals2dOneNilSlice(t *testing.T) {
	expected := false
	actual := equals2d(nil, [][]MapElement{})
	if expected != actual {
		t.Errorf("false result: (expected: %t; actual: %t)", expected, actual)
	}
}

// Check if slices having same dimesion but different elements are non-equal
func TestEquals2dDiffElmsSameDimensions(t *testing.T) {
	slice1 := [][]MapElement{{Floor, Floor}, {Floor, Floor}}
	slice2 := [][]MapElement{{Floor, Floor}, {Floor, Wall}}
	expected := false
	actual := equals2d(slice1, slice2)
	if expected != actual {
		t.Errorf("false result: (expected: %t; actual: %t)", expected, actual)
	}
}

//Check if slices having different dimesions but same elemnents are non-equal.
func TestEquals2dDiffRowDimensionsSameElms(t *testing.T) {
	slice1 := [][]MapElement{{Floor, Floor}, {Floor, Floor}}
	slice2 := [][]MapElement{{Floor, Floor}, {Floor, Floor, Floor}}
	expected := false
	actual := equals2d(slice1, slice2)
	if expected != actual {
		t.Errorf("false result: (expected: %t; actual: %t)", expected, actual)
	}
}

//Check if slices having different dimesions and different elemnents are non-equal.
func TestEquals2dDiffDimensionsDiffElms(t *testing.T) {
	slice1 := [][]MapElement{{Floor, Floor}, {Floor, Wall}}
	slice2 := [][]MapElement{{Floor, Wall}, {Floor, Floor, Floor}}
	expected := false
	actual := equals2d(slice1, slice2)
	if expected != actual {
		t.Errorf("false result: (expected: %t; actual: %t)", expected, actual)
	}
}

//Check if slices having different dimesions and different elemnents are non-equal.
func TestEquals2dDiffColDimensionsSameElms(t *testing.T) {
	slice1 := [][]MapElement{{Floor, Floor}, {Floor, Wall}}
	slice2 := [][]MapElement{{Floor, Floor}, {Floor, Floor}, {Floor}}
	expected := false
	actual := equals2d(slice1, slice2)
	if expected != actual {
		t.Errorf("false result: (expected: %t; actual: %t)", expected, actual)
	}
}

func TestConvertValidFloorSymbol(t *testing.T) {
	expected := Floor
	actual, err := convert(rune('0'))
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
	}
	if expected != actual {
		t.Errorf("false result: (expected: %v; actual: %v)", expected, actual)
	}
}

func TestConvertValidWallSymbol(t *testing.T) {
	expected := Wall
	actual, err := convert(rune('2'))
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
	}
	if expected != actual || err != nil {
		t.Errorf("false result: (expected: %v; actual: %v)", expected, actual)
	}
}

func TestConvertValidMKSymbol(t *testing.T) {
	expected := Medkit
	actual, err := convert(rune('+'))
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
	}
	if expected != actual || err != nil {
		t.Errorf("false result: (expected: %v; actual: %v)", expected, actual)
	}
}

func TestConvertValidWeSymbol(t *testing.T) {
	expected := Weapon
	actual, err := convert(rune('*'))
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
	}
	if expected != actual || err != nil {
		t.Errorf("false result: (expected: %v; actual: %v)", expected, actual)
	}
}

func TestConvertValidDecoSymbol(t *testing.T) {
	expected := Floor
	actual, err := convert(rune('~'))
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
	}
	if expected != actual {
		t.Errorf("false result: (expected: %v; actual: %v)", expected, actual)
	}
}

func TestConvertInvalidSymbol(t *testing.T) {
	_, err := convert(rune(' '))
	if err == nil {
		t.Error("an error should have been raised")
	}

}
