package gamemap_test

import (
	"testing"

	"github.zhaw.ch/raycasters/server/gamemap"
)

func TestCoordinate(t *testing.T) {
	// Attention this test does not prove anything
	// it is only ofr coverage
	coor := gamemap.NewCoordinate(0, 0, 0)
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

func TestNewMapCoordinate(t *testing.T) {
	mc := gamemap.NewMapCoordinate(1.0, 2.0)
	if mc.X() != 1.0 {
		t.Error("NewMapCooridinate Stored X Cooridinate wrong")
		return
	}
	if mc.Y() != 2.0 {
		t.Error("NewMapCooridinate Stored Y Cooridinate wrong")
	}
}

func TestNewCustomCoordinate(t *testing.T) {
	mcc := gamemap.NewCustomCoordinate(1.5, 2.5)
	if mcc.GetMapCoordinate().X() != 1.0 {
		t.Error("NewMapCooridinate Stored X Cooridinate wrong")
		return
	}
	if mcc.GetMapCoordinate().Y() != 2.0 {
		t.Error("NewMapCooridinate Stored Y Cooridinate wrong")
	}
}
