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
func TestMapControllerIsShotInDirectionOfTargetSameCoords(t *testing.T) {
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
func TestMapControllerIsShotInDirectionOfTargetHitDiagonal(t *testing.T) {
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
func TestMapControllerIsShotInDirectionOfTargetHitVertical(t *testing.T) {
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
func TestMapControllerIsShotInDirectionOfTargetHitHorizontal(t *testing.T) {
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
