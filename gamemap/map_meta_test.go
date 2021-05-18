package gamemap

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestGetID(t *testing.T) {
	mm := MapMeta{
		id: 1,
		name: "desert.txt",
		filePath: "/maps/desert/",
	}

	if mm.GetId() != 1 {
		t.Error("MapMeta returned wrong ID")
	}
}

func TestGetName(t *testing.T) {
	mm := MapMeta{
		id: 1,
		name: "desert.txt",
		filePath: "/maps/desert/",
	}

	if mm.GetName() != "desert.txt" {
		t.Error("MapMeta returned wrong Name")
	}
}

func TestGetFilePath(t *testing.T) {
	mm := MapMeta{
		id: 1,
		name: "desert.txt",
		filePath: "/maps/desert/",
	}

	if mm.GetFilePath() != "/maps/desert/" {
		t.Error("MapMeta returned wrong FilePath")
	}
}

func TestLoadFile(t *testing.T) {
	file, err := ioutil.TempFile(".", "tempfile")
	if err != nil {
		t.Error(err.Error())
		return
	}

	mm := MapMeta{
		id: 1,
		name: file.Name(),
		filePath: "",
	}
	// "defer" stacks
	defer func() {
		os.Remove(file.Name())
		file.Close()
	}()

	expectedMap2d := newMap2d(2, [][]MapElement{{Floor, Floor}, {Floor, Floor}})
	file.WriteString(intoMapString(2, []string{"00", "00"}))
	actualMap2d, err := mm.LoadFile()
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	expected := true
	actual := expectedMap2d.Compare(&actualMap2d)
	if expected != actual {
		t.Errorf("false result: (expected: %v; actual: %v)", expected, actual)
	}
}