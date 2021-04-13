package gamemap

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMap(t *testing.T) {
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

	if !Equal(expected_gmap, gotten_gmap.GMap, t) {
		t.Errorf("You are dumb")
		return
	}
}

func Equal(exp, act [][]MapElement, t *testing.T) bool {
	if len(exp) != len(act) {
		t.Errorf("Error: columns number missmatch. (Exp: %d, Act: %d)", len(exp), len(act))
		return false
	}
	if len(exp[0]) != len(act[0]) {
		t.Errorf("Error: rows number missmatch. (Exp: %d, Act: %d)", len(exp[0]), len(act[0]))
		return false
	}
	for rowIndex := 0; rowIndex < len(exp); rowIndex += 1 {
		for columnIndex, expEl := range exp[rowIndex] {
			if expEl != act[rowIndex][columnIndex] {
				t.Errorf("Error: wrong element. (Exp: %d, Act: %d)", expEl, act[rowIndex][columnIndex])
				return false
			}
		}
	}

	return true
}
