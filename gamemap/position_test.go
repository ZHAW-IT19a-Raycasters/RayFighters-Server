package gamemap_test

import (
	"testing"

	"github.zhaw.ch/raycasters/server/gamemap"
)

func TestCoordinate(t *testing.T) {
	// Attention this test does not prove anything
	// it is only ofr coverage
	coor := gamemap.NewCoordinate()
	exp := 44.4
	coor.SetAngle(exp)
	got := coor.Angle()

	if exp != got {
		t.Error("wrong value")
		return
	}

	coor.SetX(exp)
	got = coor.X()
	if exp != got {
		t.Error("wrong value")
		return
	}

	coor.SetY(exp)
	got = coor.Y()
	if exp != got {
		t.Error("wrong value")
		return
	}
}
