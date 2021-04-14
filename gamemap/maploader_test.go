package gamemap

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

var expected_map = [][]MapElement{
	{Floor, Floor, Floor, Floor, Floor},
	{Floor, Floor, Floor, Floor, Floor},
	{Floor, Floor, Floor, Floor, Floor},
	{Floor, Floor, Floor, Floor, Floor},
	{Floor, Floor, Floor, Floor, Floor},
}
var testFolderPath string = "/testmaps"

func TestValidMap(t *testing.T) {
	actual_map, err := loadMap("testmap_01_small_valid.txt")
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
	actual_map, err := loadMap("testmap_02_small_invalid_dimension_column.txt")
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
	actual_map, err := loadMap("testmap_03_small_invalid_dimension_row.txt")
	if err != nil {
		t.Errorf("Error: %s", err.Error())
		return
	}

	if status, _ := compare2DSlices(expected_map, actual_map.GMap); status {
		t.Error("Expected: Row length missmatch.")
		return
	}
}

func loadMap(fileName string) (GameMap, error) {

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

func compare2DSlices(exp, act [][]MapElement) (bool, error) {
	if len(exp) != len(act) {
		return false, fmt.Errorf("Error: columns number missmatch. (Exp: %d, Act: %d)", len(exp), len(act))
	}
	for rowIndex := 0; rowIndex < len(exp); rowIndex += 1 {

		if len(exp[rowIndex]) != len(act[rowIndex]) {
			return false, fmt.Errorf("Error: row %d length missmatch. (Exp: %d, Act: %d)", rowIndex, len(exp[rowIndex]), len(act[rowIndex]))
		}

		for columnIndex, expEl := range exp[rowIndex] {
			if expEl != act[rowIndex][columnIndex] {
				return false, fmt.Errorf("Error: wrong element. (Exp: %d, Act: %d)", expEl, act[rowIndex][columnIndex])
			}
		}
	}

	return true, nil
}
