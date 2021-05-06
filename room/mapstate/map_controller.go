package mapstate

import (
	"math/rand"
	"sync"

	"github.zhaw.ch/raycasters/server/gamemap"
)

type MapController struct {
	clients     map[uint8]*gamemap.Coordinate
	spawingarea []gamemap.Coordinate
	map2d       gamemap.Map2d
	getNextInt  func(int) int
	lock        *sync.RWMutex
}

func NewMapController(filePath string) (MapController, error) {
	map2d, err := gamemap.LoadMap2d(filePath)
	if err != nil {
		return MapController{}, err
	}
	return MapController{
		spawingarea: map2d.GetLegalSpawns(),
		map2d:       map2d,
		clients:     make(map[uint8]*gamemap.Coordinate),
		getNextInt:  rand.Intn,
		lock:        &sync.RWMutex{},
	}, nil
}

// clientExists will check if the client already exists.
// Attention this function locks on its own.
func (mc *MapController) clientExists(id uint8) bool {
	mc.lock.RLock()
	defer mc.lock.RUnlock()
	_, ok := mc.clients[id]
	return ok
}

// AddClient will add a new client to the game map
func (mc *MapController) AddClient(id uint8) (gamemap.Coordinate, error) {

	ok := mc.clientExists(id)
	if ok {
		return gamemap.Coordinate{}, ErrClientExists{
			id: id,
		}
	}

	nextRand := mc.getNextInt(len(mc.spawingarea))
	spawn := mc.spawingarea[nextRand]
	coord := gamemap.NewCoordinate(spawn.X(), spawn.Y(), 0)
	pCoord := &coord

	mc.lock.Lock()
	defer mc.lock.Unlock()

	mc.clients[id] = pCoord
	return coord, nil
}

// RemoveClient will remove a client if possible
func (mc *MapController) RemoveClient(id uint8) error {

	ok := mc.clientExists(id)
	if !ok {
		return ErrClientDoesNotExist{
			id: id,
		}
	}

	mc.lock.Lock()
	defer mc.lock.Unlock()

	delete(mc.clients, id)
	return nil
}

// MoveClient will move the specified client to the given location
func (mc *MapController) MoveClient(id uint8, x, y, angle float64) (gamemap.Coordinate, error) {

	// no lock is needed here as
	// there the coordinate has
	// its own.
	client, ok := mc.clients[id]
	if !ok {
		return gamemap.Coordinate{}, ErrClientExists{
			id: id,
		}
	}

	coordinates := gamemap.NewCoordinate(x, y, angle)
	mapCoordinates := coordinates.GetMapCoordinate()
	element := mc.map2d.MapElementAt(mapCoordinates.X(), mapCoordinates.Y())

	if element == gamemap.Wall {
		return *client, ErrWalkedIntoWall{id: id}
	}

	client.SetX(x)
	client.SetY(y)
	client.SetAngle(angle)

	return coordinates, nil
}

// GetClient will return a the state of a specific client.
// And if the client could not be found false will be returned
func (mc *MapController) getClient(id uint8) (gamemap.Coordinate, bool) {
	mc.lock.RLock()
	defer mc.lock.RUnlock()
	c, ok := mc.clients[id]
	if !ok {
		return gamemap.Coordinate{}, ok
	}
	return *c, ok
}
