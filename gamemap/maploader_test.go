package gamemap

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestLoadMapPanic(t *testing.T) {
	// Turn off the panic.
	defer func() { recover() }()
	LoadMap("NonExistingPath")
	// Never reaches here if `LoadMap` panics.
	t.Errorf("should have panicked")
}

func TestLoadMapNoPanic(t *testing.T) {
	fmt.Println(os.Getwd())
	defer func() {
		if recover() != nil {
			t.Errorf("should not have panicked")
		}
	}()
	file, err := ioutil.TempFile(".", "tempfile")
	if err != nil {
		t.Error(err)
		return
	}
	// "defer" stacks
	defer os.Remove(file.Name())
	if err != nil {
		t.Error(err)
		return
	}
	defer file.Close()
	if err != nil {
		t.Error(err)
		return
	}
	LoadMap(file.Name())
}
func mapReader(mapDim int, mapSlice []string) *strings.Reader {
	return strings.NewReader(mapString(mapDim, mapSlice))
}

// mapString converts given 1d slice into string, each position of given slice is interpreted as a new row.
// Is a helper function used in testing.
func mapString(mapDim int, mapSlice []string) string {
	var map2d bytes.Buffer
	for i := 0; i < headerLength-1; i++ {
		map2d.WriteString("\n")
	}

	map2d.WriteString(strconv.Itoa(mapDim))
	map2d.WriteString("\n")
	dim := len(mapSlice)
	for i := 0; i < dim; i++ {
		map2d.WriteString(mapSlice[i])
		if i != dim-1 {
			map2d.WriteString("\n")
		}
	}
	return map2d.String()
}

// Test if mapString produces expected result.
func TestMapString(t *testing.T) {
	expected := "\n\n\n\n2\n00\n0"
	actual := mapString(2, []string{"00", "0"})
	if expected != actual {
		t.Errorf("\nfalse result: (\nexpected:\n%s\nactual:\n%s", expected, actual)
	}
}

// Test reading given valid map, produces expected result.
func TestLoadMapValidMapOneSymbol(t *testing.T) {
	expectedMap := [][]MapElement{
		{Floor, Floor},
		{Floor, Floor},
	}
	actualMap, err := parseMap(mapReader(2, []string{"00", "00"}))
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}

	expected := true
	actual := equals2d(expectedMap, actualMap)
	if expected != actual {
		t.Errorf("false result: (expected: %v; actual: %v)", expected, actual)
		return
	}
}

func TestParseLine(t *testing.T) {
	someRune := rune('0')
	someElement, err := convert(someRune)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}
	expectedSlice := []MapElement{someElement, someElement, someElement}
	str := string(someRune)
	str += str + str
	actualSlice, err := parseLine(str)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}
	expected := true
	actual := equals1d(expectedSlice, actualSlice)
	if expected != actual {
		t.Errorf("false result: (expected: %v; actual: %v)", expected, actual)
	}
}

// Feeding empty string should return empty slice.
func TestParseLineEmptyLine(t *testing.T) {
	expectedSlice := []MapElement{}
	str := ""
	actualSlice, err := parseLine(str)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}
	expected := true
	actual := equals1d(expectedSlice, actualSlice)
	if expected != actual {
		t.Errorf("false result: (expected: %v; actual: %v)", expected, actual)
	}
}
