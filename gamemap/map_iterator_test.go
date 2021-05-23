package gamemap

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func compareMapMeta(expected, actual *MapMeta, t *testing.T) bool {
	if expected.GetId() != actual.GetId() {
		t.Log(expected.GetId())
		t.Log(actual.GetId())
		return true
	}
	if expected.GetFilePath() != actual.GetFilePath() {
		t.Log(expected.GetFilePath())
		t.Log(actual.GetFilePath())
		return true
	}
	if expected.GetName() != actual.GetName() {
		t.Log(expected.GetName())
		t.Log(actual.GetName())
		return true
	}
	return false
}

func TestGetMap(t *testing.T) {
	file, err := ioutil.TempFile(os.TempDir(), "tempfile")
	defer func() {
		os.Remove(file.Name())
		file.Close()
	}()

	if err != nil {
		t.Error("Error occurred in creating tempfile")
		return
	}

	file.WriteString("[" +
		"{" +
		"\"id\": 0," +
		"\"name\": \"Super gritty city backyard\"," +
		"\"filePath\": \"maps/city_yard.txt\"" +
		"}," +
		"{" +
		"\"id\": 1," +
		"\"name\": \"Desert\"," +
		"\"filePath\": \"maps/desert.txt\"" +
		"}]")

	mim, err := NewMapIteratorManager(file.Name())

	mi := mim.GetIterator()

	expected1 := MapMeta{
		id: 0,
		name: "Super gritty city backyard",
		filePath: filepath.Join(os.TempDir() + "/maps/city_yard.txt"),
	}

	expected2 := MapMeta{
		id: 1,
		name: "Desert",
		filePath: filepath.Join(os.TempDir() + "/maps/desert.txt"),
	}

	if compareMapMeta(&expected1, mi.GetMap(), t) {
		t.Error("Error in first Map Iteration")
		return
	}

	if compareMapMeta(&expected2, mi.GetMap(), t) {
		t.Error("Error in second Map Iteration")
		return
	}
}


func TestNewMapIteratorManager(t *testing.T) {
	file, err := ioutil.TempFile(os.TempDir(), "tempfile")
	defer func() {
		os.Remove(file.Name())
		file.Close()
	}()

	if err != nil {
		t.Error("Error occurred in creating tempfile")
		return
	}

	file.WriteString("[" +
		"{" +
			"\"id\": 0," +
			"\"name\": \"Super gritty city backyard\"," +
			"\"filePath\": \"maps/city_yard.txt\"" +
		"}," +
		"{" +
			"\"id\": 1," +
			"\"name\": \"Desert\"," +
			"\"filePath\": \"maps/desert.txt\"" +
		"}]")

	mim, err := NewMapIteratorManager(file.Name())

	if err != nil {
		t.Error("Error occurred in NewMapIteratorManager")
		return
	}

	expected1 := MapMeta{
		id: 0,
		name: "Super gritty city backyard",
		filePath: filepath.Join(os.TempDir() + "/maps/city_yard.txt"),
	}

	expected2 := MapMeta{
		id: 1,
		name: "Desert",
		filePath: filepath.Join(os.TempDir() + "/maps/desert.txt"),
	}

	if compareMapMeta(&expected1, <- mim.maps, t) {
		t.Error("Error in first Map Iteration")
		return
	}

	if compareMapMeta(&expected2, <- mim.maps, t) {
		t.Error("Error in second Map Iteration")
		return
	}

	if mim.GetFilePath() != file.Name() {
		t.Error("FilePath was set wrong in NewMapIterator")
	}
}

func TestWrongJson(t *testing.T) {
	file, err := ioutil.TempFile(os.TempDir(), "tempfile")
	defer func() {
		os.Remove(file.Name())
		file.Close()
	}()

	if err != nil {
		t.Error("Error occurred in creating tempfile")
		return
	}

	file.WriteString("[" +
		"{" +
		"\"id\": 0," +
		"\"name\": \"Super gritty city backyard\"," +
		"\"filePath\": \"maps/city_yard.txt\"" +
		"{" +
		"\"id\": 1," +
		"\"name\": \"Desert\"," +
		"\"filePath\": \"maps/desert.txt\"")

	_, err = NewMapIteratorManager(file.Name())

	if err == nil {
		t.Error("Unmarshal didn't detect a wrong JSON")
	}
}

func TestReadJSON(t *testing.T) {
	json := "[" +
		"{" +
		"\"id\": 0," +
		"\"name\": \"Super gritty city backyard\"," +
		"\"filePath\": \"maps/city_yard.txt\"" +
		"}," +
		"{" +
		"\"id\": 1," +
		"\"name\": \"Desert\"," +
		"\"filePath\": \"maps/desert.txt\"" +
		"}]"
	_, err := parseJSON([]byte(json), "Not/Real/Path")
	t.Log(err)
	if err != nil {
		t.Error("Expected an FilePath Error")
	}
}

func TestIsValid(t *testing.T) {
	err := IsValid("Not Real Dir")
	if err == nil {
		t.Error("IsValid should return Error and didn't")
	}
}
