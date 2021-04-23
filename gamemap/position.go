package gamemap

import (
	"math"
	"sync"
)

type MapCoordinate struct {
	x float64
	y float64
}

func NewMapCoordinate(x, y float64) MapCoordinate {
	return MapCoordinate{
		x: x,
		y: y,
	}
}

func (mc MapCoordinate) X() float64 {
	return mc.x
}

func (mc MapCoordinate) Y() float64 {
	return mc.y
}

type Coordinate struct {
	lock  *sync.RWMutex
	x     float64
	y     float64
	angle float64
}

func NewCoordinate() Coordinate {
	return Coordinate{
		x:     0,
		y:     0,
		angle: 0,
		lock:  &sync.RWMutex{},
	}
}

func (c *Coordinate) GetMapCoordinate() MapCoordinate {
	// x = 1.1 => 1.0
	c.lock.RLock()
	defer c.lock.RUnlock()
	return NewMapCoordinate(math.Floor(c.x), math.Floor(c.y))
}

func (c *Coordinate) X() float64 {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.x
}

func (c *Coordinate) Y() float64 {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.y
}

func (c *Coordinate) Angle() float64 {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.angle
}

func (c *Coordinate) SetX(x float64) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.x = x
}

func (c *Coordinate) SetY(y float64) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.y = y
}

func (c *Coordinate) SetAngle(angle float64) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.angle = angle
}
