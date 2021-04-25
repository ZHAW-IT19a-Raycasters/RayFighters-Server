package mapstate

import (
	"errors"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

const sundownmap = `Sundown Resort
assets/sceneries/surreal_beach.jpg

#f07572;#a9ccd2
#f07572;32;0.5
32
22222211111666666666111111222222
20000200000000000000000000200002
20000000000000000000000000000002
20000200000000000000000000000002
20000200000030000030000000200002
20000200000030000030000000200002
20000200000030000030000000200002
10000200000030000030000000200001
10000200000030000030000000200001
10000000000030000030000000200001
10000200000030000030000000200001
12222200000030000030000000200001
60000000000030000030000000000006
60000000000030000030000000000006
60000333333330000033333333300006
60000000000000000000000000000006
60000000000000000000000000000006
60000333333333303333300333300006
60000000000000303000000000300006
10000000000000303000000000300001
10000200000000303000000000300001
10000200000000303000000000300001
10000200000000303000000000300001
10000200000000303000000000000001
10000200000000303000000000000001
20000200000000303000000000300002
20000200000000303000000000300002
20000200000000303000000000300002
20000200000000000000000000300002
20000200000003333333333333300002
20000200000000000000000000000002
22222211111666666666111111222222
6
1 = assets/walls/vertical_bars_pink.png
2 = assets/walls/slate_pink.jpg
3 = assets/walls/slate_purple.jpg
4 = assets/walls/slate_orange.jpg
5 = assets/walls/glass_pink.png
6 = assets/walls/glass_pink.png
`

func cleanUpTmpFile(name string) {
	if _, err := os.Stat(name); errors.Is(err, os.ErrExist) {
		os.Remove(name)
	}
}

func setupTmpFile(t *testing.T) (string, error) {
	tmpFile, err := ioutil.TempFile("", "sundown_resort*.txt")
	if err != nil {
		t.Errorf("on creating the tmp file, this happend [%s]", err.Error())
		return "", err
	}

	t.Cleanup(func() {
		cleanUpTmpFile(tmpFile.Name())
	})

	return tmpFile.Name(), nil
}

func writeTmpFile(name, input string) error {
	return ioutil.WriteFile(name, []byte(input), 0644)
}

func TestGameStateAddClientWithExistingID(t *testing.T) {
	file, err := setupTmpFile(t)
	if err != nil {
		t.Error("setupTmpFile failed")
		return
	}
	err = writeTmpFile(file, sundownmap)
	if err != nil {
		t.Error("writeTmpFile failed")
		return
	}
	mapController, err := NewMapController(file)
	if err != nil {
		t.Error("NewMapController returned error in TestAddClient")
		return
	}
	err = mapController.AddClient(1)

	if err != nil {
		t.Error("AddClient returned an error")
		return
	}

	err = mapController.AddClient(1)

	if err == nil {
		t.Error("2 Clients with the same id were created")
	} else if !errors.Is(err, ErrClientExists{}) {
		t.Error("Wrong error type returned")
	} else if !strings.HasSuffix(err.Error(), "[1] already exists") {
		t.Error("Wrong error suffix used")
	}
}

func TestMapControllerRemoveNonExistingClient(t *testing.T) {
	file, err := setupTmpFile(t)
	if err != nil {
		t.Error("setupTmpFile failed")
		return
	}
	err = writeTmpFile(file, sundownmap)
	if err != nil {
		t.Error("writeTmpFile failed")
		return
	}
	mapController, err := NewMapController(file)
	if err != nil {
		t.Error("NewMapController returned error in TestRemoveClient")
		return
	}

	err = mapController.RemoveClient(1)

	if err == nil {
		t.Error("RemoveClient should returned an error")
	}
}

func TestMapControllerRemoveClient(t *testing.T) {
	file, err := setupTmpFile(t)
	if err != nil {
		t.Error("setupTmpFile failed")
		return
	}
	err = writeTmpFile(file, sundownmap)
	if err != nil {
		t.Error("writeTmpFile failed")
		return
	}
	mapController, err := NewMapController(file)
	if err != nil {
		t.Error("NewMapController returned error in TestRemoveClient")
		return
	}

	err = mapController.AddClient(1)

	if err != nil {
		t.Error("AddClient returned an error")
		return
	}

	_, ok := mapController.getClient(1)
	if !ok {
		t.Error("GetClient returned an error")
		return
	}

	err = mapController.RemoveClient(1)
	if err != nil {
		t.Error("Remove Client returned error")
		return
	}

	_, ok = mapController.getClient(1)
	if ok {
		t.Error("Client wasn't removed")
	}
}

func TestMoveClientTest(t *testing.T) {
	file, err := setupTmpFile(t)
	if err != nil {
		t.Error("setupTmpFile failed")
		return
	}
	err = writeTmpFile(file, sundownmap)
	if err != nil {
		t.Error("writeTmpFile failed")
		return
	}
	mapController, err := NewMapController(file)
	if err != nil {
		t.Error("NewMapController returned error in TestMoveClient")
		return
	}

	err = mapController.AddClient(1)

	if err != nil {
		t.Error("AddClient returned an error")
		return
	}

	coordinate, err := mapController.MoveClient(1, 1, 1, 1)
	if err != nil {
		t.Error("MoveClient returned an error")
		return
	}

	if coordinate.X() != 1 || coordinate.Y() != 1 || coordinate.Angle() != 1 {
		t.Error("MoveClient didn't move the client")
	}
}

func TestMoveClientUnExistingIdTest(t *testing.T) {
	file, err := setupTmpFile(t)
	if err != nil {
		t.Error("setupTmpFile failed")
		return
	}
	err = writeTmpFile(file, sundownmap)
	if err != nil {
		t.Error("writeTmpFile failed")
		return
	}
	mapController, err := NewMapController(file)
	if err != nil {
		t.Error("NewMapController returned error in TestMoveClient")
		return
	}

	_, err = mapController.MoveClient(1, 1, 1, 1)
	if err == nil {
		t.Error("MoveClient should return an error")
		return
	}
}

func TestMoveClientFailTest(t *testing.T) {
	file, err := setupTmpFile(t)
	if err != nil {
		t.Error("setupTmpFile failed")
		return
	}
	err = writeTmpFile(file, sundownmap)
	if err != nil {
		t.Error("writeTmpFile failed")
		return
	}
	mapController, err := NewMapController(file)
	if err != nil {
		t.Error("NewMapController returned error in TestMoveClient")
		return
	}

	err = mapController.AddClient(1)

	if err != nil {
		t.Error("AddClient returned an error")
		return
	}

	_, err = mapController.MoveClient(1, 7, 0, 1)
	if err == nil {
		t.Error("MoveClient should have return an error")
		return
	}
}

func TestGetUnExistingClient(t *testing.T) {
	file, err := setupTmpFile(t)
	if err != nil {
		t.Error("setupTmpFile failed")
		return
	}
	err = writeTmpFile(file, sundownmap)
	if err != nil {
		t.Error("writeTmpFile failed")
		return
	}
	mapController, err := NewMapController(file)
	if err != nil {
		t.Error("NewMapController returned error in TestRemoveClient")
		return
	}

	_, ok := mapController.getClient(1)

	if ok {
		t.Error("Client shouldn't exist in GetUnExistingClient")
	}
}

func TestLoadFaultyMapController(t *testing.T) {
	_, err := NewMapController("thisreallyshouldtexist.txt")
	if err == nil {
		t.Error("Filename shouldn't exist if so try looking for it")
		return
	}
}

func TestSpawningPointAddClient(t *testing.T) {
	file, err := setupTmpFile(t)
	if err != nil {
		t.Error("setupTmpFile failed")
		return
	}
	err = writeTmpFile(file, sundownmap)
	if err != nil {
		t.Error("writeTmpFile failed")
		return
	}
	mapController, err := NewMapController(file)
	if err != nil {
		t.Error("NewMapController returned error in TestMoveClient")
		return
	}

	mapController.getNextInt = func(n int) int {return 4}

	err = mapController.AddClient(1)

	if err != nil {
		t.Error("AddClient returned an error")
		return
	}

	coordinate, ok := mapController.getClient(1)
	if !ok {
		t.Error("Client 1 wasn't created")
		return
	}

	if coordinate.X() != mapController.spawingarea[4].X() && coordinate.Y() != mapController.spawingarea[4].Y() {
		t.Error("Coordinates aren't the same")
	}
}
