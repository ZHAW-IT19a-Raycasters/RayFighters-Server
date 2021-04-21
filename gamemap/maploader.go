package gamemap

import (
	"errors"
	"io"
	"os"

	"github.zhaw.ch/raycasters/server/gamelocation/utils"
	"github.zhaw.ch/raycasters/server/log"
)

var logger = log.Log()

// Error codes returned by failures to parse a map.
var (
	ErrShortFile = errors.New("maploader: internal error")
)

func check(e error) {
	if e != nil {
		logger.Errorf(e.Error())
		panic(e)
	}
}

// LoadMap loads game map from the given file path.
// It panics if any problemns are encountered, while reading the file.
func LoadMap(filePath string) ([][]utils.LocationElement, error) {
	file, err := os.Open(filePath)
	check(err)
	defer file.Close()
	return parseMap(file)
}

func parseMap(file io.Reader) ([][]utils.LocationElement, error) {

	return [][]utils.LocationElement{}, nil
}
