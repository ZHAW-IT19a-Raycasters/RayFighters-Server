package gamemap

import (
	"io/ioutil"
	"os"
	"reflect"
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
	expectedMap2d := newMap2d(2, [][]MapElement{{Floor, Floor}, {Floor, Floor}})
	file.WriteString(intoMapString(2, []string{"00", "00"}))
	actualMap2d, err := LoadMap2d(file.Name())
	if err != ErrorMapIsNotSquared && err != nil {
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
	if err != ErrorMapIsNotSquared && err != nil {
		t.Errorf(err.Error())
		return
	}
	if err != ErrorMapIsNotSquared {
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
	actualMap2d := newMap2d(expected, make([][]MapElement, expected))
	actual := actualMap2d.GetMapDimension()
	if expected != actual {
		t.Errorf("false result: (expected: %v; actual: %v)", expected, actual)
	}
}

//Check if getter returns expected map elment.
func TestGetLocationElementPositive(t *testing.T) {
	actualMap2d := newMap2d(4, [][]MapElement{{Floor, Floor}, {Floor, Wall}})
	expected := Wall
	actual := actualMap2d.MapElementAt(1, 1)
	if expected != actual {
		t.Errorf("false result: (expected: %v; actual: %v)", expected, actual)
	}
}

//Check if getter returns expected map elment.
func TestGetLocationElementInvalidCoor(t *testing.T) {
	actualMap2d := newMap2d(4, [][]MapElement{{Floor, Floor}, {Floor, Wall}})
	actual := actualMap2d.MapElementAt(-1, -1)
	if actual == Wall {
		t.Error("an should have been raised")
	}

}

// Test if comparing two equal map2d returns true.
func TestCompareEqualMap2ds(t *testing.T) {
	map2dOne := newMap2d(1, make([][]MapElement, 1))
	expected := true
	actual := map2dOne.Compare(&map2dOne)
	if expected != actual {
		t.Errorf("false result: (expected: %v; actual: %v)", expected, actual)
	}
}

// Test if comparing two map2d's having different dimesion will return false.
func TestCompareDifferentDimension(t *testing.T) {
	map2dOne := newMap2d(1, make([][]MapElement, 1))
	map2dTwo := newMap2d(2, make([][]MapElement, 2))
	expected := false
	actual := map2dOne.Compare(&map2dTwo)
	if expected != actual {
		t.Errorf("false result: (expected: %v; actual: %v)", expected, actual)
	}
}

func TestNewFlatMap2d(t *testing.T) {
	type args struct {
		mapDimension int
	}
	tests := []struct {
		name string
		args args
		want Map2d
	}{
		{name: "Map of dimension 4", args: args{4}, want: struct {
			dim   int
			map2d [][]MapElement
		}{dim: 4, map2d: [][]MapElement{{Floor, Floor, Floor, Floor}, {Floor, Floor, Floor, Floor}, {Floor, Floor, Floor, Floor}, {Floor, Floor, Floor, Floor}}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewFlatMap2d(tt.args.mapDimension); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewFlatMap2d() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMap2d_SetMapElement(t *testing.T) {
	type fields struct {
		dim   int
		map2d [][]MapElement
	}
	type args struct {
		x          int
		y          int
		mapElement MapElement
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   Map2d
	}{
		{name: "", fields: fields{4, [][]MapElement{{Floor, Floor, Floor, Floor}, {Floor, Floor, Floor, Floor}, {Floor, Floor, Floor, Floor}, {Floor, Floor, Floor, Floor}}}, args: args{2, 2, Wall}, want: newMap2d(4, [][]MapElement{{Floor, Floor, Floor, Floor}, {Floor, Floor, Floor, Floor}, {Floor, Floor, Wall, Floor}, {Floor, Floor, Floor, Floor}})},
		{name: "", fields: fields{3, [][]MapElement{{Floor, Floor, Floor}, {Floor, Floor, Floor}, {Floor, Floor, Floor}}}, args: args{2, 0, Wall}, want: newMap2d(3, [][]MapElement{{Floor, Floor, Wall}, {Floor, Floor, Floor}, {Floor, Floor, Floor}})},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			map2d := Map2d{
				dim:   tt.fields.dim,
				map2d: tt.fields.map2d,
			}
			if map2d.SetMapElement(tt.args.x, tt.args.y, tt.args.mapElement); !reflect.DeepEqual(map2d, tt.want) {
				t.Errorf("NewFlatMap2d() = %v, want %v", map2d, tt.want)
			}
		})
	}
}
