package gamemap

import (
	"io/ioutil"
	"os"
	"testing"
)

// Test if passing path to valid map will deliver expected result.
func TestLoadMap2dValid(t *testing.T) {
	file, err := ioutil.TempFile(".", "tempfile")
	if err != nil {
		t.Error(err.Error())
		return
	}
	// "defer" stacks
	defer os.Remove(file.Name())
	defer file.Close()
	expectedMap2d, err := newMap2d(2, [][]MapElement{{Floor, Floor}, {Floor, Floor}})
	if err != nil {
		t.Error(err.Error())
		return
	}
	file.WriteString(intoMapString(2, []string{"00", "00"}))
	actualMap2d, err := LoadMap2d(file.Name())
	if err != ErrMap2dNotSq && err != nil {
		t.Errorf(err.Error())
		return
	}
	expected := true
	actual := expectedMap2d.Compare(&actualMap2d)
	if expected != actual {
		t.Errorf("false result: (expected: %v; actual: %v)", expected, actual)
	}
}

// Test if passing a path to  squared map will raise an error.
func TestLoadMap2dNotSquared(t *testing.T) {
	file, err := ioutil.TempFile(".", "tempfile")
	if err != nil {
		t.Error(err.Error())
		return
	}
	// "defer" stacks
	defer os.Remove(file.Name())
	defer file.Close()

	file.WriteString(intoMapString(2, []string{"00", "000"}))
	_, err = LoadMap2d(file.Name())
	if err != ErrMap2dNotSq && err != nil {
		t.Errorf(err.Error())
		return
	}
	if err != ErrMap2dNotSq {
		t.Error("an error should have been raised")
	}
}

// Test if passing invalid path will raise an error.
func TestLoadMap2dNonExistingPath(t *testing.T) {
	_, err := LoadMap2d("SomeNonExistingPath")
	if err == nil {
		t.Error("an error should have been raised")
	}
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
func TestGetLocationElementPositive(t *testing.T) {
	actualMap2d, err := newMap2d(4, [][]MapElement{{Floor, Floor}, {Floor, Wall}})
	if err != nil {
		t.Error(err)
	}
	expected := Wall
	actual, err := actualMap2d.MapElementAt(1, 1)
	if err != nil {
		t.Error(err)
	}
	if expected != actual {
		t.Errorf("false result: (expected: %v; actual: %v)", expected, actual)
	}
}

//Check if getter returns expected map elment.
func TestGetLocationElementInvalidCoor(t *testing.T) {
	actualMap2d, err := newMap2d(4, [][]MapElement{{Floor, Floor}, {Floor, Wall}})
	if err != nil {
		t.Error(err)
	}
	_, err = actualMap2d.MapElementAt(-1, -1)
	if err == nil {
		t.Error("an should have been raised")
	}

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
