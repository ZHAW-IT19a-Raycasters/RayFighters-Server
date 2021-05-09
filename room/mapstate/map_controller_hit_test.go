package mapstate

import (
	"github.zhaw.ch/raycasters/server/gamemap"
	"math"
	"sync"
	"testing"
)

// NOTE:
// Positive directions:
//	- coordinates (x horizontally to the right, y vertically down);
//	- angle (with respect to x, clockwise)

func Test_isShotInDirectionOfTarget(t *testing.T) {
	type args struct {
		xTarget   float64
		yTarget   float64
		xShooter  float64
		yShooter  float64
		shotAngle float64
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "Hor. shot which hits", args: args{1.0, 0.0, 0.0, 0.0, 0.0}, want: true},
		{name: "Vert. shot which hits", args: args{1.0, 1.0, 0.0, 0.0, math.Pi / 2}, want: true},
		{name: "Diag. shot which hits", args: args{1.0, 1.0, 0.0, 0.0, math.Pi / 4}, want: true},
		{name: "Shooter and Target coords are equal", args: args{0.0, 0.0, 0.0, 0.0, math.Pi / 4}, want: false},
		{name: "Target is behind of shooter", args: args{0.0, 0.0, 1.0, 1.0, math.Pi / 4}, want: false},
		{name: "Target is in front of shooter", args: args{0.0, 0.0, 1.0, 1.0, math.Pi - math.Pi/4}, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isShotInDirectionOfTarget(tt.args.xTarget, tt.args.yTarget, tt.args.xShooter, tt.args.yShooter, tt.args.shotAngle); got != tt.want {
				t.Errorf("isShotInDirectionOfTarget() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Test if distance from point to the shot trajectory is calculated correctly
func Test_calcDistToShotTrajectory(t *testing.T) {
	type args struct {
		a       float64
		b       float64
		c       float64
		xTarget float64
		yTarget float64
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{name: "Point on the line", args: args{a: 1.0, b: -1.0, c: 0.0, xTarget: 1.0, yTarget: 1.0}, want: 0.0},
		{name: "ShotTraj: x-Axis, TargetPos: off the line", args: args{a: 0.0, b: 1.0, c: 0.0, xTarget: 1.0, yTarget: 2.0}, want: 2.0},
		{name: "ShotTraj: x-Axis, TargetPos: on the line", args: args{a: 0.0, b: 1.0, c: 0.0, xTarget: 1.0, yTarget: 0.0}, want: 0.0},
		{name: "ShotTraj: x-Axis, TargetPos: off the line", args: args{a: 0.0, b: 1.0, c: 0.0, xTarget: 0.0, yTarget: 5.0}, want: 5.0},
		{name: "ShotTraj: y-Axis, TargetPos: off the line", args: args{a: 1.0, b: 0.0, c: 0.0, xTarget: 6.0, yTarget: 100.0}, want: 6.0},
		{name: "ShotTraj: custom, TargetPos: off the line", args: args{a: 6.0, b: -5.0, c: 10.0, xTarget: -3.0, yTarget: 7.0}, want: 5.5056},
	}
	tol := 1e-4
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := calcDistToShotTrajectory(tt.args.a, tt.args.b, tt.args.c, tt.args.xTarget, tt.args.yTarget); math.Abs(got-tt.want) > tol {
				t.Errorf("calcDistToShotTrajectory() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getDistance(t *testing.T) {
	type args struct {
		x1 float64
		y1 float64
		x2 float64
		y2 float64
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{name: "point on point", args: args{1.0, 1.0, 1.0, 1.0}, want: 0.0},
		{name: "point on point", args: args{-7.0, -4.0, 17.0, 6.5}, want: 26.196},
	}
	tol := 1e-3
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getDistance(tt.args.x1, tt.args.y1, tt.args.x2, tt.args.y2); math.Abs(got-tt.want) > tol {
				t.Errorf("getDistance() = %v, want %v", got, tt.want)
			}
		})
	}
}
func TestMapController_GetNearestHit(t *testing.T) {
	shooterID := uint8(0)
	shooterX := 1.0
	shooterY := 1.0

	type fields struct {
		clients map[uint8]*gamemap.Coordinate
		map2d   gamemap.Map2d
	}
	type args struct {
		shooterID uint8
		shotAngle float64
		xShooter  float64
		yShooter  float64
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		want        bool
		wantID      uint8
		wantTargetX float64
		wantTargetY float64
	}{
		{name: "Someone got hit", fields: fields{
			clients: *CreateCustomClientMap([]uint8{0, 1}, []float64{shooterX, shooterY, 2.5, 1.0}),
			map2d:   gamemap.NewFlatMap2d(4),
		}, args: args{shooterID, 0, shooterX, shooterY}, want: true, wantID: 1, wantTargetX: 2.5, wantTargetY: 1.0},
		{name: "Nearest test: 2 Players inline", fields: fields{
			clients: *CreateCustomClientMap([]uint8{0, 6, 10}, []float64{shooterX, shooterY, 3.0, shooterY, 2.5, shooterY}),
			map2d:   gamemap.NewFlatMap2d(4),
		}, args: args{shooterID, 0, shooterX, shooterY}, want: true, wantID: 10, wantTargetX: 2.5, wantTargetY: shooterY},
		{name: "2 Players inline one behind", fields: fields{
			clients: *CreateCustomClientMap([]uint8{0, 6, 10, 77}, []float64{shooterX, shooterY, 3.0, shooterY, 2.5, shooterY, 0.0, shooterY}),
			map2d:   gamemap.NewFlatMap2d(4),
		}, args: args{shooterID, 0, shooterX, shooterY}, want: true, wantID: 10, wantTargetX: 2.5, wantTargetY: 1.0},
		{name: "Direction test: player behind is near", fields: fields{
			clients: *CreateCustomClientMap([]uint8{007, 13, 77}, []float64{shooterX, shooterY, 3.0, shooterY, 0.0, shooterY}),
			map2d:   gamemap.NewFlatMap2d(4),
		}, args: args{shooterID, 0, shooterX, shooterY}, want: true, wantID: 13, wantTargetX: 3.0, wantTargetY: 1.0},
		{name: "Direction test: one behind", fields: fields{
			clients: *CreateCustomClientMap([]uint8{0, 88}, []float64{shooterX, shooterY, 0.0, shooterY}),
			map2d:   gamemap.NewFlatMap2d(4),
		}, args: args{shooterID, 0, shooterX, shooterY}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := &MapController{
				clients: tt.fields.clients,
				map2d:   tt.fields.map2d,
			}
			got, got1, got2, got3 := mc.GetNearestHit(tt.args.shooterID, tt.args.shotAngle, tt.args.xShooter, tt.args.yShooter)
			if got != tt.want {
				t.Errorf("GetNearestHit() got = %v, wantHit %v", got, tt.want)
			}
			if got {
				// This part only if hit was positive
				if got1 != tt.wantID {
					t.Errorf("GetNearestHit() gotID = %v, wantID %v", got1, tt.wantID)
				}
				if got2 != tt.wantTargetX {
					t.Errorf("GetNearestHit() gotTargetX = %v, wantTargetX %v", got2, tt.wantTargetX)
				}
				if got3 != tt.wantTargetY {
					t.Errorf("GetNearestHit() gotTargetY = %v, wantTargetY %v", got3, tt.wantTargetY)
				}
			}
		})
	}
}

func TestMapController_GetNearestObstacleOnTrajectoryInShotDirection(t *testing.T) {
	flatMap := gamemap.NewFlatMap2d(4)
	map2d := gamemap.NewFlatMap2d(4)
	map2d.SetMapElement(2, 0, gamemap.Wall)
	map2d.SetMapElement(3, 3, gamemap.Wall)
	map2d.SetMapElement(0, 3, gamemap.Wall)
	type fields struct {
		clients map[uint8]*gamemap.Coordinate
		map2d   gamemap.Map2d
	}
	type args struct {
		shooterID uint8
		shotAngle float64
		rayStartX float64
		rayStartY float64
	}
	tests := []struct {
		name               string
		fields             fields
		args               args
		want_foundObstacle bool
		want_xObstacle     int
		want_yObstacle     int
	}{
		{name: "Shot in x-direction hits the wall", fields: fields{*CreateCustomClientMap([]uint8{0}, []float64{0.5, 0.5}), map2d}, args: args{0, 0, 0.5, 0.5}, want_foundObstacle: true, want_xObstacle: 2, want_yObstacle: 0},
		{name: "Diagonal shot hits the wall", fields: fields{*CreateCustomClientMap([]uint8{0}, []float64{1, 1}), map2d}, args: args{0, 0.7853, 1, 1}, want_foundObstacle: true, want_xObstacle: 3, want_yObstacle: 3},
		{name: "Shot in y-direction hits the wall ", fields: fields{*CreateCustomClientMap([]uint8{0}, []float64{0.5, 0.5}), map2d}, args: args{0, 1.5708, 0.5, 0.5}, want_foundObstacle: true, want_xObstacle: 0, want_yObstacle: 3},
		{name: "Flat map: no hit", fields: fields{*CreateCustomClientMap([]uint8{0}, []float64{1, 1}), flatMap}, args: args{0, 0.7853, 1, 1}, want_foundObstacle: false, want_xObstacle: 3, want_yObstacle: 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := &MapController{
				clients: tt.fields.clients,
				map2d:   tt.fields.map2d,
				lock:    &sync.RWMutex{},
			}
			got, got1, got2 := mc.GetNearestObstacleOnTrajectoryInShotDirection(tt.args.shooterID, tt.args.shotAngle, tt.args.rayStartX, tt.args.rayStartY)
			if got != tt.want_foundObstacle {
				t.Errorf("GetNearestObstacleOnTrajectoryInShotDirection() got = %v, foundObstacle %v", got, tt.want_foundObstacle)
			}
			if got {
				if got1 != tt.want_xObstacle {
					t.Errorf("GetNearestObstacleOnTrajectoryInShotDirection() got_xObstacle = %v, want_xObstacle %v", got1, tt.want_xObstacle)
				}
				if got2 != tt.want_yObstacle {
					t.Errorf("GetNearestObstacleOnTrajectoryInShotDirection() got_yObstacle = %v, want_yObstacle %v", got2, tt.want_yObstacle)
				}
			}
		})
	}
}
