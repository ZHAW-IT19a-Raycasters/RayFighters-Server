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
	map2dSlice, err := loadmap(filePath)
	if err != nil {
		logger.Error(err.Error())
		return Map2d{}, err
	}
	if !isSquared(map2dSlice) {
		return Map2d{}, ErrMap2dNotSq
	}
	return newMap2d(len(map2dSlice), map2dSlice), nil
}

// newMap2d is inpackage convinience function for creatin Map2d using passed parameters.
func newMap2d(mDim int, lMap [][]MapElement) Map2d {
	return Map2d{dim: mDim, map2d: lMap}
}

// Return map element for given coordinates x and y.
func (map2d *Map2d) MapElementAt(x int, y int) MapElement {
	d := map2d.dim
	if x >= d || y >= d || x < 0 || y < 0 {
		return Floor
	}
	// as per definition seen in
	// https://github.zhaw.ch/raycasters/docs/wiki/Map
	return map2d.map2d[y][x]
}

// GetMapDimension returns map dimension
func (map2d *Map2d) GetMapDimension() int {
	return map2d.dim
}

// Compare compares two Map2d structs.
// Only dimension and elements on corresponding positions are same.
func (map2d *Map2d) Compare(other *Map2d) bool {
	if map2d.GetMapDimension() != other.dim { // NOTE: because in the same package dimesion of "other" can be accessed
		return false
	}
	return equals2d(map2d.map2d, other.map2d)
}

func (map2d *Map2d) GetLegalSpawns() []Coordinate {
	counter := 0
	for _, x := range map2d.map2d {
		for _, y := range x {
			if y != Wall {
				counter += 1
			}
		}
	}

	res := make([]Coordinate, 0, counter)
	for i, x := range map2d.map2d {
		for j, y := range x {
			if y != Wall {
				res = append(res, NewCoordinate(float64(i), float64(j), 0))
			}
		}
	}
	return res
}

func (map2d *Map2d) IsMapBoundary(x int, y int) bool {
	return x == (map2d.dim-1) || y == (map2d.dim-1)
}

// IsPenetrable true if a bullet can pass through it
func (map2d *Map2d) IsPenetrable(x int, y int) bool {
	return map2d.MapElementAt(x, y) != Wall
}
