package gamemap

import (
	"bufio"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

// Test if loading file from non-existing path will raise an error.
func TestLoadMapInvalidPath(t *testing.T) {
	_, err := loadmap("NonExistingPath")
	if err == nil {
		t.Errorf("an error should have been raised")
	}
}

// Test if loading a file from existing path will raise no error.
func TestLoadMapValidPath(t *testing.T) {
	file, err := ioutil.TempFile(".", "tempfile")
	if err != nil {
		t.Error(err.Error())
	}
	// "defer" stacks
	defer os.Remove(file.Name())
	defer file.Close()
	_, err = loadmap(file.Name())
	if err == nil {
		t.Errorf("an error should have been raised")
	}
}

// Test if map in the given file, having more columns than given dimesion, will not raise an error.
func TestParseMapLessRowsThanDimension(t *testing.T) {
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

// Test if parsing a map file consistin of unsupported symbols will raise an error.
func TestParseMapWrongSymbols(t *testing.T) {
	_, err := parseMap(intoMapReader(2, []string{"  ", "  "}))
	if err == nil {
		t.Errorf("an error should have been raised")
		return
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

// Feeding empty string should return empty slice.
func TestParseLineInvalidSymobols(t *testing.T) {
	_, err := parseLine("  ")
	if err == nil {
		t.Errorf("an error should have been raises")
		return
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

// Test if reading non integer number will raise an error
func TestReadMapDimensionNonInt(t *testing.T) {
	scanner := bufio.NewScanner(strings.NewReader("muhaha"))
	_, err := readMapDimension(scanner)
	if err == nil {
		t.Errorf("an error should have been raised")
		return
	}
}

// Test if reading a dimension from a file with no lines left will raise an error.
func TestReadMapDimensionNoLinesToRead(t *testing.T) {
	scanner := bufio.NewScanner(strings.NewReader("40"))
	scanner.Scan()
	_, err := readMapDimension(scanner)
	if err == nil {
		t.Errorf("an error should have been raised")
		return
	}

}
