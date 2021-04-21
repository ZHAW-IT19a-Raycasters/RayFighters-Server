package gamemap

import (
	"errors"
)

// Map2d is structure used for representation of 2d squared location map.
type Map2d struct {
	dim   int
	map2d [][]MapElement
}

// Error codes returned by failures in gamemap.
var (
	ErrMap2dIndx  = errors.New("map2d: index out of bound")
	ErrMap2dNotSq = errors.New("map2d: invalid map dimension")
)

// LoadMap2d loads map from given path and wraps it into ingame abstraction.
func LoadMap2d(filePath string) (Map2d, error) {
	map2dSlice, err := LoadMap(filePath)
	if err != nil {
		logger.Error(err.Error())
		return Map2d{}, err
	}
	if !isSquared(map2dSlice) {
		return Map2d{}, ErrMap2dNotSq
	}
	return newMap2d(len(map2dSlice), map2dSlice)
}

// newMap2d is inpackage convinience function for creatin Map2d using passed parameters.
func newMap2d(mDim int, lMap [][]MapElement) (Map2d, error) {
	return Map2d{dim: mDim, map2d: lMap}, nil
}

// Return map element for given coordinates x and y.
func (map2d *Map2d) MapElementAt(x int, y int) (MapElement, error) {
	d := map2d.dim
	if x >= d || y >= d || x < 0 || y < 0 {
		return 0, ErrMap2dIndx
	}
	return map2d.map2d[x][y], nil
}

// GetMapDimension returns map dimension
func (map2d *Map2d) GetMapDimension() int {
	return map2d.dim
}

// Compare compares two Map2d structs.
// Only dimension and elements on corresponding positions are same.
func (l *Map2d) Compare(other *Map2d) bool {
	if l.GetMapDimension() != other.dim { // NOTE: because in the same package dimesion of "other" can be accessed
		return false
	}
	return equals2d(l.map2d, other.map2d)
}
