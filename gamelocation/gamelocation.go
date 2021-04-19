package gamelocation

import (
	"fmt"

	"github.zhaw.ch/raycasters/server/gamelocation/maploader"
	"github.zhaw.ch/raycasters/server/gamelocation/utils"
	"github.zhaw.ch/raycasters/server/log"
)

var logger = log.Log()

// Location is abstraction for game location.
type Location struct {
	dimension   int
	locationMap [][]utils.LocationElement
}

/*
const (
	minMapDimesion  = 1
	maxMapDimension = 999
)
*/
// Loads map from under given path and wraps it into "Location",
func GetGameLocation(filePath string) (Location, error) {
	locationMap, err := maploader.LoadMap(filePath)
	if err != nil {
		logger.Errorf("unexpected error: %s", err.Error())
		return Location{}, fmt.Errorf("unexpected error: %s", err.Error())
	}

	return getNewLocation(len(*locationMap), *locationMap)
}

// Create new location, havin map with dimesion mapDimension^2.
// Throws errors if:
// - given mapDimension is not the same as given locationMap has
func getNewLocation(mapDimension int, locationMap [][]utils.LocationElement) (Location, error) {
	if len(locationMap) != mapDimension {
		return Location{}, fmt.Errorf("invalid map-column dimesion")
	}
	for rowIndex := 0; rowIndex < mapDimension; rowIndex += 1 {
		if len(locationMap[rowIndex][:]) != mapDimension {
			return Location{}, fmt.Errorf("invalid map-row dimesion")
		}
	}
	return Location{dimension: mapDimension, locationMap: locationMap}, nil
}

// Return map element for given coordinates x and y.
func (gamelocation *Location) GetLocationElement(x int, y int) (utils.LocationElement, error) {
	dim := gamelocation.dimension
	if x >= dim || y >= dim {
		return 0, fmt.Errorf("index out of bound")
	}
	return gamelocation.locationMap[x][y], nil
}

// Return map dimension
func (gamelocation *Location) GetMapDimension() int {
	return gamelocation.dimension
}

// Compare two locations
func (this Location) Compare(other Location) bool {
	if this.GetMapDimension() != other.dimension { // NOTE: because in the same package dimesion of "other" can be accessed
		return false
	}
	return utils.Compare2DMapElementSlices(this.locationMap, other.locationMap)
}
