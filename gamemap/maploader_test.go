package gamemap

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestValidMap(t *testing.T) {
	path, err := os.Getwd()
	if err != nil {
		t.Errorf("Error while getting the working directory:\n[%s]", err.Error())
		return
	}
	path = filepath.Join(path, "testmap_01_small_valid.txt")

	expected_gmap := [][]MapElement{
		{Floor, Floor, Floor, Floor, Floor},
		{Floor, Floor, Floor, Floor, Floor},
		{Floor, Floor, Floor, Floor, Floor},
		{Floor, Floor, Floor, Floor, Floor},
		{Floor, Floor, Floor, Floor, Floor},
	}

	gotten_gmap, err := MapParser(path)

	if err != nil {
		t.Errorf("Error: %s", err.Error())
		return
	}

	if status, err := Equal(expected_gmap, gotten_gmap.GMap, t); !status {
		t.Errorf(err.Error())
		return
	}
}
func TestUnvalidMapRowMissDimention(t *testing.T) {
	path, err := os.Getwd()
	if err != nil {
		t.Errorf("Error while getting the working directory:\n[%s]", err.Error())
		return
	}
	path = filepath.Join(path, "testmap_03_small_invalid_rowdim.txt")

	expected_gmap := [][]MapElement{
		{Floor, Floor, Floor, Floor, Floor},
		{Floor, Floor, Floor, Floor, Floor},
		{Floor, Floor, Floor, Floor, Floor},
		{Floor, Floor, Floor, Floor, Floor},
		{Floor, Floor, Floor, Floor, Floor},
	}

	gotten_gmap, err := MapParser(path)

	if err != nil {
		t.Errorf("Error: %s", err.Error())
		return
	}

	if status, _ := Equal(expected_gmap, gotten_gmap.GMap, t); status {
		t.Error("Expected: Row length missmatch.")
		return
	}
}
func TestUnvalidMapColumnsMissDimention(t *testing.T) {
	path, err := os.Getwd()
	if err != nil {
		t.Errorf("Error while getting the working directory:\n[%s]", err.Error())
		return
	}
	path = filepath.Join(path, "testmap_03_small_invalid_rowdim.txt")

	expected_gmap := [][]MapElement{
		{Floor, Floor, Floor, Floor, Floor},
		{Floor, Floor, Floor, Floor, Floor},
		{Floor, Floor, Floor, Floor, Floor},
		{Floor, Floor, Floor, Floor, Floor},
		{Floor, Floor, Floor, Floor, Floor},
	}

	gotten_gmap, err := MapParser(path)

	if err != nil {
		t.Errorf("Error: %s", err.Error())
		return
	}

	if status, _ := Equal(expected_gmap, gotten_gmap.GMap, t); status {
		t.Error("Expected: Columns number missmatch.")
		return
	}
}
func Equal(exp, act [][]MapElement, t *testing.T) (bool, error) {
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
