package mapstate

import (
	"math"
	"math/rand"
	"sync"

	"github.zhaw.ch/raycasters/server/gamemap"
)

type MapController struct {
	clients     map[uint8]gamemap.Coordinate
	spawingarea []gamemap.Coordinate
	map2d       gamemap.Map2d
	getNextInt  func(int) int
	lock        *sync.RWMutex
}

func NewMapController(filePath string) (MapController, error) {
	map2d, err := gamemap.LoadMap2d(filePath)
	if err != nil {
		// TODO: Write wrapper for this error
		return MapController{}, err
	}
	return MapController{
		spawingarea: map2d.GetLegalSpawns(),
		map2d:       map2d,
		clients:     make(map[uint8]gamemap.Coordinate),
		getNextInt:  rand.Intn,
		lock:        &sync.RWMutex{},
	}, nil
}

func (mc *MapController) AddClient(id uint8) error {
	// TODO: Clean up
	mc.lock.Lock()
	defer mc.lock.Unlock()

	_, ok := mc.clients[id]

	if ok {
		return ErrClientExists{
			id: id,
		}
	}

	nextRand := mc.getNextInt(len(mc.spawingarea))
	spawn := mc.spawingarea[nextRand]
	mc.clients[id] = gamemap.NewCoordinate(spawn.X(), spawn.Y(), 0)
	return nil
}

// RemoveClient will remove a client if possible
func (mc *MapController) RemoveClient(id uint8) error {
	mc.lock.Lock()
	defer mc.lock.Unlock()
	_, ok := mc.clients[id]
	if !ok {
		return ErrClientDoesNotExist{
			id: id,
		}
	}
	delete(mc.clients, id)
	return nil
}

func (mc *MapController) MoveClient(id uint8, x, y, angle float64) (gamemap.Coordinate, error) {
	mc.lock.Lock()
	defer mc.lock.Unlock()
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
		return client, ErrWalkedIntoWall{id: id}
	}
	client.SetX(x)
	client.SetY(y)
	client.SetAngle(angle)

	mc.clients[id] = client

	return client, nil
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
	return c, ok
}

func (mc *MapController) GetPosition(userId uint8) (x float64, y float64, ok bool) {
	c, ok := mc.getClient(userId)
	if !ok {
		return 0, 0, ok
	}
	return c.Y(), c.Y(), ok
}

func (mc *MapController) GetMapCoordinates(userId uint8) (x int, y int, ok bool) {
	c, ok := mc.getClient(userId)
	if !ok {
		return 0, 0, ok
	}
	return c.GetMapCoordinate().X(), c.GetMapCoordinate().Y(), ok
}

func (mc *MapController) IsMapBound(x int, y int) bool {
	return mc.map2d.IsMapBoundary(x, y)
}

// IsPenetrable true if a bullet can pass through it
func (mc *MapController) IsPenetrable(x int, y int) bool {
	return mc.map2d.IsPenetrable(x, y)
}

// returns uid of the client which will be hit by the ray
func (mc *MapController) GetClientID(angle float64, rayStartX float64, rayStartY float64) (bool, uint8) {
	a := -math.Tan(angle)
	b := 1.0
	c := math.Tan(angle)*rayStartX - rayStartY
	margine := 0.00001
	for id, coord := range mc.clients {
		if math.Abs(calculateDistance(a, b, c, coord.X(), coord.Y())) < margine {
			return true, id
		}
	}
	return false, 0
}
func calculateDistance(a float64, b float64, c float64, x float64, y float64) float64 {
	return math.Abs(a*x+b*y+c) / math.Sqrt(a*a+b*b)
}
