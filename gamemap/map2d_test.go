package gamemap

import (
	"testing"
)

// Test if passing correct parameters, will raise no erros.
func TestGetNewLocationValid(t *testing.T) {

}

// Test if passing incorrect map dimension will rais an error.
func TestGetNewLocationInvalidMapDimesion(t *testing.T) {

}

//Check if reading bigger map will not raise an error.
func TestGetGameLocationValid(t *testing.T) {

}

//Check if getter returns expected dimension.
func TestGetMapDimension(t *testing.T) {
	expected := 2
	actualMap2d, err := newMap2d(expected, make([][]MapElement, expected))
	if err != nil {
		t.Error(err)
	}
	actual := actualMap2d.GetMapDimension()
	if expected != actual {
		t.Errorf("false result: (expected: %v; actual: %v)", expected, actual)
	}
}

//Check if getter returns expected map elment.
func TestGetLocationElement(t *testing.T) {

}

// Test if comparing two equal map2d returns true.
func TestCompareEqualMap2ds(t *testing.T) {
	map2dOne, err := newMap2d(1, make([][]MapElement, 1))
	if err != nil {
		t.Error(err)
	}
	expected := true
	actual := map2dOne.Compare(&map2dOne)
	if expected != actual {
		t.Errorf("false result: (expected: %v; actual: %v)", expected, actual)
	}
}

// Test if comparing two map2d's having different dimesion will return false.
func TestCompareDifferentDimension(t *testing.T) {
	map2dOne, err := newMap2d(1, make([][]MapElement, 1))
	if err != nil {
		t.Error(err)
	}
	map2dTwo, err := newMap2d(2, make([][]MapElement, 2))
	if err != nil {
		t.Error(err)
	}
	expected := false
	actual := map2dOne.Compare(&map2dTwo)
	if expected != actual {
		t.Errorf("false result: (expected: %v; actual: %v)", expected, actual)
	}
}
