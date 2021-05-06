package mapstate

import (
	"math"
	"testing"
)

// NOTE:
// Positive directions:
//	- coordinates (x horizontally to the right, y vertically down);
//	- angle (with respect to x, clockwise)

// If shooter and target are having same coordinates, false is expected
func Test_isShotInDirectionOfTargetSameCoords(t *testing.T) {
	xTarget := 0.0
	yTarget := xTarget
	xShooter := 0.0
	yShooter := xShooter
	shotAngle := 0.2

	expected := false
	actual := isShotInDirectionOfTarget(xTarget, yTarget, xShooter, yShooter, shotAngle)
	if actual != expected {
		t.Errorf("Expected: %v Actual: %v ", expected, actual)
	}
}

// If the shot direction may possibly hit target, true is expected.
func Test_isShotInDirectionOfTargetHitDiagonal(t *testing.T) {
	xShooter := 0.0
	yShooter := 0.0
	xTarget := 1.0
	yTarget := 1.0
	shotAngle := math.Pi / 4

	expected := true
	actual := isShotInDirectionOfTarget(xTarget, yTarget, xShooter, yShooter, shotAngle)
	if actual != expected {
		t.Errorf("Expected: %v Actual: %v ", expected, actual)
	}
}

// If the shot direction may possibly hit target, true is expected.
func Test_isShotInDirectionOfTargetHitVertical(t *testing.T) {
	xShooter := 0.0
	yShooter := 0.0
	xTarget := 0.0
	yTarget := 1.0
	shotAngle := math.Pi / 2

	expected := true
	actual := isShotInDirectionOfTarget(xTarget, yTarget, xShooter, yShooter, shotAngle)
	if actual != expected {
		t.Errorf("Expected: %v Actual: %v ", expected, actual)
	}
}

// If the shot direction may possibly hit target, true is expected.
func Test_isShotInDirectionOfTargetHitHorizontal(t *testing.T) {
	xShooter := 0.0
	yShooter := 0.0
	xTarget := 1.0
	yTarget := 0.0
	shotAngle := 0.0

	expected := true
	actual := isShotInDirectionOfTarget(xTarget, yTarget, xShooter, yShooter, shotAngle)
	if actual != expected {
		t.Errorf("Expected: %v Actual: %v ", expected, actual)
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
