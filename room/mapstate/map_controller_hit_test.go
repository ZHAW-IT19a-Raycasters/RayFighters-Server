package mapstate

import (
	"math"
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
	tol := 1e-4
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getDistance(tt.args.x1, tt.args.y1, tt.args.x2, tt.args.y2); math.Abs(got-tt.want) <= tol {
				t.Errorf("getDistance() = %v, want %v", got, tt.want)
			}
		})
	}
}
