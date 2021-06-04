package testhelpers

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.zhaw.ch/raycasters/server/gamemap"
)

const json = `
[
	{
		"id": 0,
		"name": "Super gritty city backyard",
		"filePath": "city_yard.txt"
	}
]
`

const city_yard_name = "city_yard.txt"

const city_yard = `Super gritty city backyard
assets/sceneries/skyline.jpg

#303030;#010101
#000000;32;0.5
32
11111111111111111111111111111111
10000000000000000000000000000001
10222222222300000322222222002221
10200000000300000300000002000001
10200000000000000000000002222001
10200000000000000000000002002001
10200000000000000000000002002001
10200000300300300300300330002001
10000000355555555555555300002001
10200000300300300300300333333001
10200000000000000000000000000001
10000000000000000000000000000001
10000000000050000000000000000201
10200000000000000000000000000201
10200000000000000000000500000201
10200000000000000000000000000201
10252525200025252520000000000201
10200000000000000050000000000201
10200000000000000020000005000201
10200000000000000050000000000201
10000000000000000020000000000201
10000000000000000050000000000201
10200000000000000020000000000001
10200000000000000050000000000001
10252525252525250520000000000001
10200000000000020522222222222201
10200000000000020000000000000201
10200000000000020000000000000201
10200000000000000000000000000201
10220022222222222200000000022201
10000000000000020000000000000001
11111111111111111111111111111111
5
1 = assets/walls/fence_high.png
2 = assets/walls/concrete.jpg
3 = assets/walls/concrete_neon_pink.jpg
4 = assets/walls/concrete_neon_blue.jpg
5 = assets/walls/slate_pink.jpg
`

// errHandle will handle the given error and
// will stop a tests execution on an error.
func errHandle(err error, t *testing.T) {
	if err != nil {
		t.Errorf("something went wrong while setting up the temp dir [%s]", err.Error())
		t.SkipNow()
	}
}

// SetupMapIterator will create an initialized test MapIterator.
func SetupMapIterator(t *testing.T) gamemap.MapIteratorManager {
	pathJson, _ := SetupIndexJson(t)
	mapIter, err := gamemap.NewMapIteratorManager(pathJson)
	errHandle(err, t)
	return mapIter
}

// SetupIndexJson will create a testing index.json file.
func SetupIndexJson(t *testing.T) (string, string) {
	// setup main folder
	dir, err := ioutil.TempDir("", "maps*")
	errHandle(err, t)
	t.Cleanup(func() {
		os.RemoveAll(dir)
	})

	pathJson := filepath.Join(dir, "index.json")
	err = ioutil.WriteFile(pathJson, []byte(json), 0777)
	errHandle(err, t)

	pathCity := filepath.Join(dir, city_yard_name)

	err = ioutil.WriteFile(pathCity, []byte(city_yard), 0777)
	errHandle(err, t)

	return pathJson, pathCity
}
