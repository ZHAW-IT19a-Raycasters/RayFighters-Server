package mapstate

import (
	"math"
	"math/rand"
	"sync"

	"github.zhaw.ch/raycasters/server/log"

	"github.zhaw.ch/raycasters/server/gamemap"
)

var logger = log.Log()

const spawnOffset = 0.5

type iRand interface {
	Intn(n int) int
	Float64() float64
}

type sRand struct{}

func (r sRand) Intn(n int) int {
	return rand.Intn(n)
}

func (r sRand) Float64() float64 {
	return rand.Float64()
}

type MapController struct {
	clients     map[uint8]*gamemap.Coordinate
	spawingarea []gamemap.Coordinate
	map2d       gamemap.Map2d
	random      iRand
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
		random:      sRand{},
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

	coord := mc.unsafeRandomSpawn(id)

	return coord, nil
}

// unsafeRandomSpawn is the inner logic of the random spawn
// functionality, this function does not assume anything
// about the players existance or else, so be carefull
// when using it.
func (mc *MapController) unsafeRandomSpawn(id uint8) gamemap.Coordinate {
	nextRand := mc.random.Intn(len(mc.spawingarea))
	spawn := mc.spawingarea[nextRand]

	logger.Debugf("x: [%f] y: [%f]", spawn.X(), spawn.Y())
	// spawn.SetX(spawn.X() + spawnOffset)
	// spawn.SetY(spawn.Y() + spawnOffset)

	logger.Debugf("post x: [%f] y: [%f]", spawn.X(), spawn.Y())
	// coord := gamemap.NewCoordinate(spawn.X(), spawn.Y(), 0)
	coord := gamemap.NewCoordinate(spawn.X()+0.5, spawn.Y()+0.5, 0)
	pCoord := &coord

	mc.lock.Lock()
	defer mc.lock.Unlock()

	mc.clients[id] = pCoord

	return coord
}

// RandomSpawn will move a players location as to respawn it into the game.
func (mc *MapController) RandomSpawn(id uint8) (gamemap.Coordinate, error) {
	// no lock is needed here as
	// there the coordinate has
	// its own.
	_, ok := mc.clients[id]
	if !ok {
		return gamemap.Coordinate{}, ErrClientDoesNotExist{
			id: id,
		}
	}

	coord := mc.unsafeRandomSpawn(id)

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
		return gamemap.Coordinate{}, ErrClientDoesNotExist{
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

func (mc *MapController) GetPosition(userId uint8) (x float64, y float64, ok bool) {
	c, ok := mc.getClient(userId)
	if !ok {
		return 0, 0, ok
	}
	return c.X(), c.Y(), ok
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

// GetNearestHit returns uid of the nearest client, which may be hit by the shot, given there are no impenetrable obstacles on the way.
func (mc *MapController) GetNearestHit(shooterID uint8, shotAngle, xShooter, yShooter float64) (bool, uint8, float64, float64) {
	// calculate ray constants
	slope := math.Tan(shotAngle)
	a := -slope
	b := 1.0
	c := slope*xShooter - yShooter
	// defines target size
	margin := 0.15
	hit := false
	targetID := uint8(0)
	distToRay := 0.0
	xTarget := 0.0
	yTarget := 0.0
	distToShooter := math.MaxFloat64
	// We don't need lock because,
	// The iteration order over maps is not specified and is not guaranteed to be the same from one iteration
	// to the next. If a map entry that has not yet been reached is removed during iteration, the corresponding
	// iteration value will not be produced.
	// CHECK: https://golang.org/ref/spec#For_statements
	// Get the nearest target which intersects the shooting trajectory
	for id, coord := range mc.clients {
		if id == shooterID {
			// shooter itself should be skipped
			continue
		}
		x := coord.X()
		y := coord.Y()
		// if area around target intersects the ray -> hit
		distToRay = calcDistToShotTrajectory(a, b, c, x, y)
		if distToRay <= margin {
			// Invariant: target intersects the trajectory of shot

			if isShotInDirectionOfTarget(x, y, xShooter, yShooter, shotAngle) {
				// pick the nearest target
				d := getDistance(xShooter, yShooter, x, y)
				if distToShooter > d {
					hit = true
					// save the best result
					targetID = id
					distToShooter = d
					xTarget = x
					yTarget = y
				}
			}
		}
		logger.Debugf("Shot from (%f,%f) @ Target (%f,%f). Hit? %v. Delta trajectory: %f", xShooter, yShooter, xTarget, yTarget, hit, distToRay)
	}
	return hit, targetID, xTarget, yTarget
}

// getDistance returns euclidean distance between two points on the map
func getDistance(x1, y1, x2, y2 float64) float64 {
	dx := x2 - x1
	dy := y2 - y1
	return math.Sqrt(dx*dx + dy*dy)
}

// isShotInDirectionOfTarget checks if shot was in the direction of target
// Note: doesn't check if ray effectively hits the target
// Note: if (xS,yS) equal to (xTarget, yTarget) false is returned
func isShotInDirectionOfTarget(xTarget, yTarget, xShooter, yShooter, shotAngle float64) bool {
	// margin defines, when a-b is considered to be zero
	margin := 0.0000000000000001
	shotDirX := 1.0 * math.Cos(shotAngle)
	shotDirY := 1.0 * math.Sin(shotAngle)
	distX := xShooter - xTarget
	distY := yShooter - yTarget
	// shotDirX(Y) and distX(Y) have to point in diff. directions
	// check cases: diag. shot, horizontal shot, vertical shot
	if (shotDirX*distX < 0 && shotDirY*distY < 0) || (shotDirX-distX < margin && shotDirY*distY < 0) || (shotDirY-distY < margin && shotDirX*distX < 0) {
		return true
	}
	return false
}

// calcDistToShotTrajectory returns distance between target and an shooting trajectory defined by given parameters
// https://en.wikipedia.org/wiki/Distance_from_a_point_to_a_line
func calcDistToShotTrajectory(a, b, c, xTarget, yTarget float64) float64 {
	return math.Abs(a*xTarget+b*yTarget+c) / math.Sqrt(a*a+b*b)
}

// raycasting return true if some player was hit and uid of victim, else (false, 0) is returnedx
// shotAngle is in radians
func (mc *MapController) CheckShooting(shooterID uint8, shotAngle float64) (hit bool, victimID uint8) {
	logger.Debugf("Shot angle: %f°", 180/math.Pi*shotAngle)
	// y axes directed vertically downwards
	// x axes directed horizontally to the right
	// STEP 1: Get candidate

	// ray origin coordinates
	//rayStartX, rayStartY, _ := mc.GetPosition(shooterID)
	//hit, targetID, targetX, targetY := mc.GetNearestHit(shooterID, shotAngle, rayStartX, rayStartY)
	rayStartX, rayStartY, _ := mc.GetPosition(shooterID)
	logger.Debugf("Ray origin: x-%f, y-%f", rayStartX, rayStartY)
	hit, targetID, targetX, targetY := mc.GetNearestHit(shooterID, shotAngle, rayStartX, rayStartY)
	logger.Debugf("Shooter: %d Hit: %v", shooterID, hit)
	if !hit {
		return false, 0
	} else {
		foundObstacle, xObstacle, yObstacle := mc.GetNearestObstacleOnTrajectoryInShotDirection(shooterID, shotAngle, rayStartX, rayStartY)
		if !foundObstacle {
			// someone got hit :(
			return true, targetID
		} else {
			// some obstacle on the way was found
			// Check if obstacle lies
			// between target and shooter
			if isTargetProtected(rayStartX, float64(xObstacle), targetX) && isTargetProtected(rayStartY, float64(yObstacle), targetY) {
				// obstacle lies between shooter and target
				// target is safe
				// for now:)
				return false, 0
			}
		}
	}
	return true, targetID
}

// isTargetProtected verifies if for given parameters target will be protected by obstacle
// Example y:(2.5,2.0,2,5) -> true
func isTargetProtected(coordShooter, coordObstacle, coordTarget float64) bool {
	return math.Floor(coordShooter) <= coordObstacle && coordObstacle <= math.Floor(coordTarget) || math.Floor(coordTarget) <= coordObstacle && coordObstacle <= math.Floor(coordShooter)
}

func (mc *MapController) GetNearestObstacleOnTrajectoryInShotDirection(shooterID uint8, shotAngle, rayStartX, rayStartY float64) (foundObstacle bool, xObstacle int, yObstacle int) {
	// STEP 1: Initialization
	// maX(mapY) are used to the cell,
	// we are focused on.
	// Initialize them with coordinates of the cell,
	// where the shot was done.
	mapX, mapY, _ := mc.GetMapCoordinates(shooterID)
	mapDim := mc.map2d.GetMapDimension()
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

	// In each iteration step of algorithm
	// we walk 1 unit on x(or y)
	// The direction of "walk" is given by stepX(stepY) variables,
	// which are unchanged during the whole algorithm.
	var stepX int
	var stepY int
	//  STEP 2: Calculate steps and the ray length in the cell,
	// 			where the ray takes it origin,
	// 			until the first intersection with x-(or y-)grid line
	if rayDirX < 0 {
		// ray directed reversed to y-direction
		stepX = -1 // we go in negative x direction
		// (player posit. - curr cell coordinates ) x scaleFactor
		rayLengthX = (rayStartX - float64(mapX)) * rayLengthScaleFactorStepInX
	} else {
		// ray co-directed with y-direction
		stepX = 1 // we go in positive x direction
		// (coordinates of next cell - player posit.) x step
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
	// STEP 2: 	Finding nearest obstacle on the trajectory
	//			in direction of shooting.
	// In each step check if the cell is penetrable,
	// if not, we are done.
	for {
		// end condition is given by isMapBound()
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
		if mapX == mapDim || mapY == mapDim {
			// our of map bound, nothing found
			return false, -100, -100
		}
		// After each iteration check if any obstacle is on the way
		if !mc.IsPenetrable(mapX, mapY) {
			// found some obstacle
			// report
			return true, mapX, mapY
		}
	}
}

// CreateCustomClientMap is a helper function used in testing. Coordinates will be ordered to given id's
// Example uint8[2,4], float64[]{2.1,2.2,4.1,4.2} -> [2]->[2.1,2.2], [4]->[4.1,4.2]
func CreateCustomClientMap(ids []uint8, coords []float64) *map[uint8]*gamemap.Coordinate {
	clientsMap := make(map[uint8]*gamemap.Coordinate)
	for i := range ids {
		clientsMap[ids[i]] = gamemap.NewCustomCoordinate(coords[i*2], coords[i*2+1])
	}
	return &clientsMap
}
