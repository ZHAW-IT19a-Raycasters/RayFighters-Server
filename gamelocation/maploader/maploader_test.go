package maploader

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.zhaw.ch/raycasters/server/gamelocation/utils"
)

var testFolderPath string = "test/testmaps/"

// Test if map in the given file, having more columns than given dimesion, will not throw an error.
func TestLoadMapLessRowsThanDimension(t *testing.T) {
	expectedMap := [][]utils.LocationElement{
		{utils.Floor, utils.Floor, utils.Floor},
		{utils.Floor, utils.Floor},
	}
	actualMap, err := loadTestFile("testmap_04_wrong_rdimension.txt")
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}
	expected := true
	actual := utils.Compare2DMapElementSlices(expectedMap, *actualMap)
	if expected != actual {
		t.Errorf("false result: (expected: %v; actual: %v)", expected, actual)
		return
	}
}

// Test if reading map-file, having illegal dimesion-number, will throw an error.
func TestLoadMapIllegalDimension(t *testing.T) {
	_, err := loadTestFile("testmap_06_wrong_neg_dimension.txt")
	if err == nil {
		t.Errorf("no error was thrown")
		return
	}
}

// Loads testfile.
// Throws error if:
// - can not build file path
// -  LoadMap() fails
func loadTestFile(fileName string) (*[][]utils.LocationElement, error) {
	path, err := os.Getwd()
	if err != nil {

		return &([][]utils.LocationElement{}), fmt.Errorf("unexpected error: %s", err.Error())
	}
	path = filepath.Join(path, testFolderPath, fileName)
	actualMap, err := ParseMap(path)
	if err != nil {
		return &([][]utils.LocationElement{}), fmt.Errorf("unexpected error: %s", err.Error())
	}
	return actualMap, nil
}

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
	err = skipLines(scanner, 5)
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
	err = skipLines(scanner, 5)
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
	actual, err := readMapDimension(scanner)
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

	_, err = readMapDimension(scanner)
	if err == nil {
		t.Errorf("no error was thrown")
		return
	}
}
