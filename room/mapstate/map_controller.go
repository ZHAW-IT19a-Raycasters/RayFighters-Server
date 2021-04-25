package mapstate

import (
	"github.zhaw.ch/raycasters/server/gamemap"
	"sync"
)

type MapController struct {
	clients map[uint8]gamemap.Coordinate
	map2d gamemap.Map2d
	lock  *sync.RWMutex
}

func NewMapController(filePath string) (MapController, error) {
	map2d, err := gamemap.LoadMap2d(filePath)
	if err != nil {
		// TODO: Write wrapper for this error
		return MapController{}, err
	}
	return MapController{
		map2d:	 map2d,
		clients: make(map[uint8]gamemap.Coordinate),
		lock:  	 &sync.RWMutex{},
	}, nil
}

func (mc MapController) AddClient(id uint8) error {
	// TODO: Clean up
	mc.lock.Lock()
	defer mc.lock.Unlock()

	_, ok := mc.clients[id]

	if ok {
		return ErrClientExists{
			id: id,
		}
	}

	mc.clients[id] = gamemap.NewCoordinate(0, 0, 0)
	return nil
}

// RemoveClient will remove a client if possible
func (mc MapController) RemoveClient(id uint8) error {
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

func (mc MapController) MoveClient(id uint8, x, y, angle float64) (gamemap.Coordinate, error) {
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
