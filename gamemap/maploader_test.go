package gamemap

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"testing"
)

// intoMapReader converts given slice into proper map format an return a *string.Reader on it.
func intoMapReader(mapDim int, mapSlice []string) *strings.Reader {
	return strings.NewReader(intoMapString(mapDim, mapSlice))
}

// intoMapString converts given 1d slice into string, each position of given slice is interpreted as a new row.
// Is a helper function used in testing.
func intoMapString(mapDim int, mapSlice []string) string {
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

// Test if loading file from non-existing path will raise an error.
func TestLoadMapInvalidPath(t *testing.T) {
	_, err := LoadMap("NonExistingPath")
	if err == nil {
		t.Errorf("an error should have been raised")
	}
}

// Test if loading a file from existing path will raise no error.
func TestLoadMapValidPath(t *testing.T) {
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
	_, err = LoadMap(file.Name())
	if err == nil {
		t.Errorf("an error should have been raised")
	}
}

// Test if map in the given file, having more columns than given dimesion, will not raise an error.
func TestLoadMapLessRowsThanDimension(t *testing.T) {
	expectedMap := [][]MapElement{
		{Floor, Floor, Floor},
		{Floor, Floor},
	}
	actualMap, err := parseMap(intoMapReader(2, []string{"000", "00"}))
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

// Test if reading map-file, having negative dimesion-number, will raise an error.
func TestParseMapNegativeDimension(t *testing.T) {
	_, err := parseMap(intoMapReader(-3, []string{}))
	if err == nil {
		t.Errorf("an error should have been raised")
		return
	}
}

// Test if mapString produces expected result.
func TestMapString(t *testing.T) {
	expected := "\n\n\n\n2\n00\n0"
	actual := intoMapString(2, []string{"00", "0"})
	if expected != actual {
		t.Errorf("\nfalse result: (\nexpected:\n%s\nactual:\n%s", expected, actual)
	}
}

// Test reading given valid map, produces expected result.
func TestParseMapValidMapOneSymbol(t *testing.T) {
	expectedMap := [][]MapElement{
		{Floor, Floor},
		{Floor, Floor},
	}
	actualMap, err := parseMap(intoMapReader(2, []string{"00", "00"}))
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

// Test reading correct map, having multiple valid symbols, will raise no errors.
func TestParseMapValidMapMultipleSymbols(t *testing.T) {
	_, err := parseMap(intoMapReader(8, []string{
		"00123456",
		"789abcde",
		"fghjklmn",
		"opqrstuv",
		"wxyzABCD",
		"EFGHIJKL",
		"MNOPQRST",
		"UVWXYZ+*"}))
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}
}

// Test if reading a map, having less columns as given dimension, will raise an error.
func TestParseMapWrongLessColumnsThanDimension(t *testing.T) {
	_, err := parseMap(intoMapReader(2, []string{"00"}))
	if err == nil {
		t.Errorf("an error should have been raised")
		return
	}
	if err != ErrMapDimCol {
		t.Errorf("not an expected error was thrown")
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

// Test if skipping some numbers will return an expected result.
func TestSkipLines(t *testing.T) {

	scanner := bufio.NewScanner(intoMapReader(0, []string{"1", "2", "3", "4"}))
	err := skipLines(scanner, headerLength+3)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}
	scanner.Scan()
	expected := "4"
	actual := scanner.Text()
	if expected != actual {
		t.Errorf("false result: (expected: %s; actual: %s)", expected, actual)
	}
}

// Test if trying to skip more lines than file in total has will raise an error.
func TestSkipLinesShortFile(t *testing.T) {
	scanner := bufio.NewScanner(strings.NewReader(""))
	err := skipLines(scanner, 10)
	if err == nil {
		t.Errorf("an error should have been raised")
	}
}

// Test if reading a dimension from a file will return an expected rusult.
func TestReadMapDimension(t *testing.T) {
	scanner := bufio.NewScanner(strings.NewReader("40"))
	expected := 40
	actual, err := readMapDimension(scanner)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}
	if expected != actual {
		t.Errorf("false result: (expected: %d; actual: %d)", expected, actual)
	}
}

// Test if reading negative dimesion will raise an error
func TestReadMapDimensionNegativeNumber(t *testing.T) {
	scanner := bufio.NewScanner(strings.NewReader("-666"))
	_, err := readMapDimension(scanner)
	if err == nil {
		t.Errorf("an error should have been raised")
		return
	}
}
