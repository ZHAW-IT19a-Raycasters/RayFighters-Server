package maploader_test

import (
	"bufio"
	"os"
	"path/filepath"
	"testing"

	"github.zhaw.ch/raycasters/server/gamelocation/maploader"
	"github.zhaw.ch/raycasters/server/gamelocation/utils"
)

/*
var expected_map = [][]gamelocation.Location{
	{gamelocation.Floor, gamelocation.Floor},
	{gamelocation.Floor, gamelocation.Floor},
}

var testFolderPath string = "maploader/test"

func TestValidMap(t *testing.T) {
	actual_map, err := loadFile("testmap_01_small_valid.txt")
	if err != nil {
		t.Errorf("Error: %s", err.Error())
		return
	}
	if status, err := compare2DSlices(expected_map, actual_map.GMap); !status {
		t.Errorf(err.Error())
		return
	}
}
func TestMapWrongColumnDimension(t *testing.T) {
	actual_map, err := loadFile("testmap_02_small_invalid_dimension_column.txt")
	if err != nil {
		t.Errorf("Error: %s", err.Error())
		return
	}

	if status, _ := compare2DSlices(expected_map, actual_map.GMap); status {
		t.Error("Expected: Column number missmatch.")
		return
	}
}
func TestMapWrongRowDimension(t *testing.T) {
	actual_map, err := loadFile("testmap_03_small_invalid_dimension_row.txt")
	if err != nil {
		t.Errorf("Error: %s", err.Error())
		return
	}

	if status, _ := compare2DSlices(expected_map, actual_map.GMap); status {
		t.Error("Expected: Row length missmatch.")
		return
	}
}

func TestWrongMapDimensionShortMap(t *testing.T) {
	actual_map, err := loadFile("testmap_04_invalid_dimension_rows_shorter.txt")
	if err != nil {
		t.Errorf("Error: %s", err.Error())
		return
	}

	if status, _ := compare2DSlices(expected_map, actual_map.GMap); status {
		t.Error("Expected: Row length missmatch.")
		return
	}
}

//TODO test get Map elemente

func loadFile(fileName string) (GameMap, error) {

	path, err := os.Getwd()
	if err != nil {
		return GameMap{}, fmt.Errorf("Error while getting the working directory:\n[%s]", err.Error())
	}
	path = filepath.Join(path, testFolderPath, fileName)
	actual_map, err := LoadMap(path)
	if err != nil {
		return GameMap{}, fmt.Errorf("Error: %s", err.Error())
	}
	return actual_map, nil
}
*/
func TestSkipLines(t *testing.T) {
	path, err := os.Getwd()
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}
	path = filepath.Join(path, "test/skiplines.txt")
	file, err := os.Open(path)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	err = maploader.SkipLines(scanner, 5)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}
	expected := "5"
	actual := scanner.Text()
	if expected != actual {
		t.Errorf("false result: (expected: %s; actual: %s)", expected, actual)
	}
}

func TestSkipLinesShortFile(t *testing.T) {
	path, err := os.Getwd()
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}
	path = filepath.Join(path, "test/skiplines_shortfile.txt")
	file, err := os.Open(path)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	err = maploader.SkipLines(scanner, 5)
	if err == nil {
		t.Errorf("no error was thrown")
	}
}

func TestReadMapDimension(t *testing.T) {
	path, err := os.Getwd()
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}
	path = filepath.Join(path, "test/readdimension.txt")
	file, err := os.Open(path)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	expected := 40
	actual, err := maploader.ReadMapDimension(scanner)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}
	if expected != actual {
		t.Errorf("false result: (expected: %d; actual: %d)", expected, actual)
	}
}

func TestReadMapDimensionNegativeNumber(t *testing.T) {
	path, err := os.Getwd()
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}
	path = filepath.Join(path, "test/readdimension_negative.txt")
	file, err := os.Open(path)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)

	_, err = maploader.ReadMapDimension(scanner)
	if err == nil {
		t.Errorf("no error was thrown")
		return
	}
}

func TestParseLine(t *testing.T) {
	someRune := rune('0')
	someElement, err := utils.GetLocationElement(someRune)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}
	expectedSlice := []utils.LocationElement{someElement, someElement, someElement}
	str := string(someRune)
	str += str + str
	actualSlice, err := maploader.ParseLine(str)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}
	expected := true
	actual := utils.Compare1DMapElementSlices(expectedSlice, *actualSlice)
	if expected != actual {
		t.Errorf("false result: (expected: %v; actual: %v)", expected, actual)
	}
}

// Feeding empty string should return empty slice.
func TestParseLineEmptyLine(t *testing.T) {
	expectedSlice := []utils.LocationElement{}
	str := ""
	actualSlice, err := maploader.ParseLine(str)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}
	expected := true
	actual := utils.Compare1DMapElementSlices(expectedSlice, *actualSlice)
	if expected != actual {
		t.Errorf("false result: (expected: %v; actual: %v)", expected, actual)
	}
}
