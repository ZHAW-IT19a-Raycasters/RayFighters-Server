package gamelocation

/*
import (
	"fmt"

	"github.zhaw.ch/raycasters/server/gamelocation/maploader"
	"github.zhaw.ch/raycasters/server/gamelocation/utils"
)

// Location is abstraction for game location.
type Location struct {
	dimension   int
	locationMap [][]utils.LocationElement
}

const (
	minMapDimesion  = 1
	maxMapDimension = 999
)

func GetGameLocation(filePath string) Location {
	dim := 6
	m, err := maploader.LoadMap(filePath)
	if err != nil {
		panic(err)
	}
	return newLocation(dim, m)
}

// Create new "mapDimension x mapDimension" location.
func newLocation(mapDimension int, loc [][]Location) Location {
	var locationMap = make([][]utils.LocationElement, 0, mapDimension)
	for rowIndex := 0; rowIndex < mapDimension; rowIndex += 1 {
		locationMap[rowIndex] = make([]utils.LocationElement, 0, mapDimension)
	}
	return Location{dimension: mapDimension, locationMap: locationMap}
}
func SetMapElemenent(x int, y int) {
	//TODO
}

// Return map element for given coordinates x and y.
// NOTE: Location is passed by value > copy
func (gamelocation Location) GetLocationElement(x int, y int) (utils.LocationElement, error) {
	dim := gamelocation.dimension
	if x >= dim || y >= dim {
		return 0, fmt.Errorf("GAMEMAP ERROR: index out of bound.")
	}
	return gamelocation.locationMap[x][y], nil
}

// Return map dimension
// NOTE: Location is passed by value > copy
func (gamelocation Location) GetMapDimension() int {
	return gamelocation.dimension
}

// Compare two locations
func (this Location) CompareTo(other Location) bool {
	if this.GetMapDimension() != other.dimension { // NOTE: because in the same package dimesion of "other" can be accessed
		return false
	}
	return utils.Compare2DMapElementSlices(this.locationMap, other.locationMap)
}
*/
