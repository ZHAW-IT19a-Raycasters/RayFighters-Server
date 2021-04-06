package utils_test

import (
	"testing"

	"github.zhaw.ch/raycasters/server/utils"
)

func TestAtomicBoolValue(t *testing.T) {
	ab := utils.NewAtomicBool(false)
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

func TestAtomicBoolSetTrue(t *testing.T) {
	ab := utils.NewAtomicBool(false)

	if !ab.SetTrue() {
		t.Error("Should return true on the first set")
		return
	}

	if !ab.Get() {
		t.Error("Bool should be true after beeing set so")
		return
	}

	if ab.SetTrue() {
		t.Error("Should not return true on any set later")
		return
	}

	if !ab.Get() {
		t.Error("Bool should stil be true after beeing set so")
		return
	}
}

func TestAtomicBoolSetFalse(t *testing.T) {
	ab := utils.NewAtomicBool(false)

	if ab.SetFalse() {
		t.Error("Should return false as the intial condition is so")
		return
	}

	if ab.Get() {
		t.Error("Bool should be false after beeing set so")
		return
	}

	ab.Set(true)

	if !ab.Get() {
		t.Error("Bool should be true after beeing set so")
		return
	}

	if !ab.SetFalse() {
		t.Error("Should return true as the value was true")
		return
	}

	if ab.Get() {
		t.Error("Bool should be false after beeing set so")
		return
	}

	if ab.SetFalse() {
		t.Error("Should not return true since value should be false")
	}
}
func TestAtomicBoolMapBoolToInt(t *testing.T) {
	got := utils.MapBoolToInt(false)
	if got != utils.False {
		t.Error("False mapping from bool to int in the case false")
		return
	}

	got = utils.MapBoolToInt(true)
	if got != utils.True {
		t.Error("False mapping from bool to int in the case false")
	}
}

func TestMapIntToBool(t *testing.T) {
	got := utils.MapIntToBool(utils.True)
	if got != true {
		t.Error("False mapping from int to bool in the case of true")
		return
	}

	got = utils.MapIntToBool(utils.False)
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

	_ = utils.MapIntToBool(42)
}
