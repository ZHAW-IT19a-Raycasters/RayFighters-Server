package gamemap

import (
	"fmt"

	"github.zhaw.ch/raycasters/server/log"
)

var logger = log.Log()

// Location is abstraction for game location.
type Location struct {
	dim   int
	map2d [][]MapElement
}

// Loads map from under given path and wraps it into "Location"
// Will panic if unable to load file.
func GameLocation(filePath string) (Location, error) {
	locationMap, err := LoadMap(filePath)
	if err != nil {
		logger.Errorf("unexpected error: %s", err.Error())
		return Location{}, fmt.Errorf("unexpected error: %s", err.Error())
	}
	return newLocation(len(locationMap), locationMap)
}

// newLocation create new Location, having a mDim^2 dimensiond lMap map.
// Throws errors if:
// - given mapDimension is not the same as given locationMap has
func newLocation(mDim int, lMap [][]MapElement) (Location, error) {
	return Location{dim: mDim, map2d: lMap}, nil
}

// Return map element for given coordinates x and y.
func (gamelocation *Location) GetLocationElement(x int, y int) (MapElement, error) {
	dim := gamelocation.dim
	if x >= dim || y >= dim {
		return 0, fmt.Errorf("index out of bound")
	}
	return gamelocation.map2d[x][y], nil
}

// Return map dimension
func (gamelocation *Location) GetMapDimension() int {
	return gamelocation.dim
}

// Compare two locations
func (l *Location) Compare(other *Location) bool {
	if l.GetMapDimension() != other.dim { // NOTE: because in the same package dimesion of "other" can be accessed
		return false
	}
	return equals2d(l.map2d, other.map2d)
}
