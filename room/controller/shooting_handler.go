package controller

import (
	"github.zhaw.ch/raycasters/server/protocol"
	pclient "github.zhaw.ch/raycasters/server/protocol/client"
	"github.zhaw.ch/raycasters/server/protocol/server"
	"github.zhaw.ch/raycasters/server/room/helpers"
	"github.zhaw.ch/raycasters/server/room/mapstate"
	"math"
)

func (hc *HubController) handleShooting(t []byte, shooterId uint8) protocol.Bytefy {
	shot, err := pclient.NewShootingBytes(t)
	shotAngle := shot.AngPl()
	if err != nil {
		logger.Info(err)
		newUnknown := server.NewUnknown()
		return newUnknown
	}
	hit, victimId := raycasting(&hc.mapController, shooterId, shotAngle)
	var death server.PlayerDeath
	aShot := server.NewShootingPlayer(shooterId, shotAngle)
	// add id's to be ignored by broadcasting
	ignoreId := helpers.NewIdSet()
	ignoreId.Add(shooterId)
	// inform the police and players about shooting
	hc.broadCast(aShot, ignoreId)
	// someone is dead :(
	if hit {
		death = server.NewPlayerDeath(shooterId, victimId)
		// broadcast the death to all players
		// ??? if i understand correctly everyone should be informed
		hc.broadCast(death, helpers.NewIdSet())
	}
	// As the shooting opcode does not require an answer to be sent back
	return nil
}

// raycasting return true if some player was hit and uid of victim, else (false, 0) is returned
// shotAngle is in radians
func raycasting(mc *mapstate.MapController, shooterId uint8, shotAngle float64) (bool, uint8) {
	// y axes directed vertically downwards
	// x axes directed horizontally to the right

	// ray origin coordinates
	rayStartX, rayStartY, _ := mc.GetPosition(shooterId)
	// current position on map
	mapX, mapY, _ := mc.GetMapCoordinates(shooterId)
	// projections of direction ray vector on x(or y) axes
	rayDirX := math.Sin(shotAngle) // vs. cosinus???
	rayDirY := math.Cos(shotAngle)

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
	// get coordinates of absraction
	// check if some is hit
	// compare coords of abstraction an player
	mapBoundaryReached := false
	for !mapBoundaryReached {
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

		if mc.IsMapBound(mapX, mapY) || !mc.IsPenetrable(mapX, mapY) {
			return false, 0
		}
		// check if the ray hit someone?
		// A check distances to all players if smaller than some value -> true
		// B go to cell get clients around check there distances to the line
		// you need line equation

		// for each client check if potentially it could be hit
		if mc.IsClientThere(mapX, mapY) {
			// calculate pricise distance
		}

	}
}
