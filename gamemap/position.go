package gamemap

import (
	"math"
	"sync"
)

// MapCoordinate is a square in a map
// x and y are int because it
type MapCoordinate struct {
	x int
	y int
}

// NewMapCoordinate creates a New Coordinate Struct
func NewMapCoordinate(x, y float64) MapCoordinate {
	return MapCoordinate{
		x: int(x),
		y: int(y),
	}
}

// X return the x coordinate of a MapCoordinate
func (mc MapCoordinate) X() int {
	return mc.x
}

// Y return the y coordinate of a MapCoordinate
func (mc MapCoordinate) Y() int {
	return mc.y
}

// Coordinate saves a more precise position inside a map
// x and y are float64 so a position can be saves inside
// a MapCoordinate square
type Coordinate struct {
	x     float64
	y     float64
	angle float64
	lock  *sync.RWMutex
}

// NewCoordinate creates a Coordinate element
func NewCoordinate(x, y, angle float64) Coordinate {
	return Coordinate{
		x:     x,
		y:     y,
		angle: angle,
		lock:  &sync.RWMutex{},
	}
}

// NewCustomCoordinate is convenience function used for testing purposes
func NewCustomCoordinate(x, y float64) *Coordinate {
	return &Coordinate{
		x:    x,
		y:    y,
		lock: &sync.RWMutex{},
	}
}

// GetMapCoordinate return the MapCoordinate of a Coordinate
func (c *Coordinate) GetMapCoordinate() MapCoordinate {
	// x = 1.1 => 1.0
	c.lock.RLock()
	defer c.lock.RUnlock()
	return NewMapCoordinate(math.Floor(c.x), math.Floor(c.y))
}

// X return the x value of a Coordinate
func (c *Coordinate) X() float64 {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.x
}

// Y return the x value of a Coordinate
func (c *Coordinate) Y() float64 {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.y
}

// Angle return the angle value of a Coordinate
func (c *Coordinate) Angle() float64 {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.angle
}

// SetX stores the x value in a Coordinate
func (c *Coordinate) SetX(x float64) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.x = x
}

// SetY stores the x value in a Coordinate
func (c *Coordinate) SetY(y float64) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.y = y
}

// SetAngle stores the angle value in a Coordinate
func (c *Coordinate) SetAngle(angle float64) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.angle = angle
}
