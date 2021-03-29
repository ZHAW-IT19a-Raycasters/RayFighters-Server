package helpers_test

import (
	"testing"

	"github.zhaw.ch/raycasters/server/room/helpers"
)

func TestAtomicBool(t *testing.T) {
	ab := helpers.NewAtomicBool(false)
	if ab.Get() != false {
		t.Error("Incorrect value has been set while initializing.")
		return
	}

	got := ab.Set(true)

	if got != false {
		t.Error("Incorrect value has been returned from the atomic bool.")
		return
	}

	if ab.Get() == false {
		t.Error("Incorrect value has been set")
		return
	}
}

func TestMapBoolToInt(t *testing.T) {
	got := helpers.MapBoolToInt(false)
	if got != helpers.False {
		t.Error("False mapping from bool to int in the case false")
		return
	}

	got = helpers.MapBoolToInt(true)
	if got != helpers.True {
		t.Error("False mapping from bool to int in the case false")
	}
}

func TestMapIntToBool(t *testing.T) {
	got := helpers.MapIntToBool(helpers.True)
	if got != true {
		t.Error("False mapping from int to bool in the case of true")
		return
	}

	got = helpers.MapIntToBool(helpers.False)
	if got != false {
		t.Error("False mapping from int to bool in the case of false")
		return
	}
}

func TestMapIntToBoolPanic(t *testing.T) {
	defer func() {
		res := recover()
		if res == nil {
			t.Error("No panic in case of false mapping")
		}
	}()

	_ = helpers.MapIntToBool(42)
}
