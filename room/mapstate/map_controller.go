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
func (mc *MapController) GetHit(angle, rayStartX, rayStartY float64) (bool, uint8, float64, float64) {
	a := -math.Tan(angle)
	b := 1.0
	c := math.Tan(angle)*rayStartX - rayStartY
	margin := 0.00001
	hit := false
	victimId := uint8(0)
	distance := 0.0
	userX := 0.0
	userY := 0.0
	// We don't need lock because,
	// The iteration order over maps is not specified and is not guaranteed to be the same from one iteration
	// to the next. If a map entry that has not yet been reached is removed during iteration, the corresponding
	// iteration value will not be produced.
	// CHECK: https://golang.org/ref/spec#For_statements
	for id, coord := range mc.clients {
		userX = coord.X()
		userY = coord.Y()
		// TODO: Check if victim is behind shooter
		if (rayStartX-userX)*math.Cos(angle) > 0 && (rayStartY-userY)*rayStartY*math.Sin(angle) > 0 {
			//is behind
			// skip this one
			continue
		}
		if math.Abs(calculateDistance(a, b, c, userX, userY)) < margin {
			// TODO: I DON'T LIKE THIS FIND BETTER SOLUTION!!!!!
			if !hit {
				hit = true
				victimId = id
				distance = math.Sqrt(math.Abs(rayStartX-userX)*math.Abs(rayStartX-userX) +
					math.Abs(rayStartY-userY)*math.Abs(rayStartY-userY))
				continue
			}

			if distance > math.Sqrt(math.Abs(rayStartX-userX)*math.Abs(rayStartX-userX)+
				math.Abs(rayStartY-userY)*math.Abs(rayStartY-userY)) {
				victimId = id
				distance = math.Sqrt(math.Abs(rayStartX-userX)*math.Abs(rayStartX-userX) +
					math.Abs(rayStartY-userY)*math.Abs(rayStartY-userY))
			}
		}
	}
	return hit, victimId, userX, userY
}

func calculateDistance(a float64, b float64, c float64, x float64, y float64) float64 {
	return math.Abs(a*x+b*y+c) / math.Sqrt(a*a+b*b)
}

// raycasting return true if some player was hit and uid of victim, else (false, 0) is returned
// shotAngle is in radians
func (mc *MapController) CheckShooting(shooterId uint8, shotAngle float64) (bool, uint8) {
	// y axes directed vertically downwards
	// x axes directed horizontally to the right
	// STEP 1: Get candidate

	// ray origin coordinates
	rayStartX, rayStartY, _ := mc.GetPosition(shooterId)
	hit, candidateId, candX, candY := mc.GetHit(shotAngle, rayStartX, rayStartX)
	if !hit {
		return false, 0
	}

	// STEP2: CHECK if there is any obstacle in between

	// current position on map
	mapX, mapY, _ := mc.GetMapCoordinates(shooterId)
	// projections of direction ray vector on x(or y) axes
	rayDirX := math.Cos(shotAngle)
	rayDirY := math.Sin(shotAngle)

	// calculate ray length scaling factor when taking step in x(or y) direction
	//rayLengthScaleFactorStepInX := math.Sqrt(1  + (rayDirY/rayDirX) *(rayDirY/rayDirX))
	rayLengthScaleFactorStepInX := 1 / math.Abs(math.Cos(shotAngle))
	//rayLengthScaleFactorStepInY :=  math.Sqrt(1  + (rayDirX/rayDirY) *(rayDirX/rayDirY))
	rayLengthScaleFactorStepInY := 1 / math.Abs(math.Sin(shotAngle))
	// ray length in in x(or y) direction
	rayLengthX := 0.0 // if we make step in x direction
	rayLengthY := 0.0 // if we make step in y direction

	// we always walk 1 unit in the same x(und y) direction
	var stepX int
	var stepY int
	// calculate step and the ray length
	// until the first intersection with x-(or y-)grid line
	if rayDirX < 0 {
		// ray directed reversed to y-direction
		stepX = -1 // we go in negative x direction
		// (player posit. - curr cell coord ) x scaleFactor
		rayLengthX = (rayStartX - float64(mapX)) * rayLengthScaleFactorStepInX
	} else {
		// ray co-directed with y-direction
		stepX = 1 // we go in positive x direction
		// (coord of next cell - player posit.) x step
		rayLengthX = (float64(mapX+1.0) - rayStartX) * rayLengthScaleFactorStepInX
	}
	if rayDirY < 0 {
		// ray directed reversed to y-direction
		stepY = -1
		rayLengthY = (rayStartY - float64(mapY)) * rayLengthScaleFactorStepInY
	} else {
		// ray co-directed with y-direction
		stepY = 1
		rayLengthY = (float64(mapY+1.0) - rayStartY) * rayLengthScaleFactorStepInY
	}
	// get coordinates of obstacle
	// check if some is hit
	// compare coords of abstraction an player
	for {
		if rayLengthX < rayLengthY {
			// walk along x axis
			mapX += stepX
			// we move 1 unit
			// add to length 1 x scalingFactor
			rayLengthX += rayLengthScaleFactorStepInX
		} else {
			// walk along y axis
			mapY += stepY
			// we move 1 unit
			// add to length 1 x scalingFactor
			rayLengthY += rayLengthScaleFactorStepInY
		}

		// After each iteration check if there is obstacle

		if mc.IsMapBound(mapX, mapY) {
			// we reached the map bound
			// found no obstacle was on the way to candidate, therefore
			// the candidate is dead
			return true, candidateId
		}

		if !mc.IsPenetrable(mapX, mapY) {
			// found some obstacle
			// the victim is safe
			// for now
			break
		}
	}
	// Check if obstacle between candidate and shooter
	if inOrder(rayStartX, float64(mapX), candX) && inOrder(rayStartY, float64(mapY), candY) {
		return false, 0
	}
	return true, candidateId
}

// true if values are ordered as given
func inOrder(var1, var2, var3 float64) bool {
	return var1 < var2 && var2 < var3
}
