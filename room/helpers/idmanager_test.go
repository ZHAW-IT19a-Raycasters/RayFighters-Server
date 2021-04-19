package helpers

import (
	"errors"
	"math"
	"testing"
)

func TestIdManagerPop(t *testing.T) {
	manager := NewIdManager()
	// check if the manager is full
	if !manager.IsFull() {
		t.Error("Manager should be full")
	}

	// check that there are exacly 256 elements there
	for i := 0; i <= math.MaxUint8; i += 1 {
		val, err := manager.Pop()
		if err != nil {

			t.Errorf("Undefined state at iteration [%d]\nError: [%s]", i, err.Error())
			return
		}
		if val != uint8(i) {
			t.Errorf("Wrong value at position [%d]\ngot: [%d]\nexp: [%d]", i, val, i)
			return
		}
	}
	if !manager.IsEmpty() {
		t.Error("Manager should be empty")
	}

	_, err := manager.Pop()
	if err == nil {
		t.Error("on the incorrect pop request, there should be an error")
	} else if !errors.Is(err, ErrIdManagerEmpty) {
		t.Error("incorrect error was send")
	}
}

func TestIdManagerPush(t *testing.T) {
	manager := NewIdManager()
	for i := 0; i <= math.MaxUint8; i += 1 {
		// empty out the manager
		_, _ = manager.Pop()
	}
	// fill up the manager
	for i := 0; i <= math.MaxUint8; i += 1 {
		// empty out the manager
		err := manager.Push(uint8(i))
		if err != nil {
			t.Errorf("The manager should still be fillable at position [%d]", i)
		}
	}

	// check if the manager is full
	if !manager.IsFull() {
		t.Error("Manager should be full")
		return
	}
	if manager.Count() != math.MaxUint8+1 {
		t.Errorf("Count should be at [%d] is at [%d]", math.MaxUint8+1, manager.Count())
		return
	}
	err := manager.Push(42)
	if err == nil {
		t.Errorf("The manager should not be fillable anymore.")
	} else if !errors.Is(err, ErrIdManagerFull) {
		t.Error("incorrect error was send")
	}
}

func TestIdManagerPopPush(t *testing.T) {
	manager := NewIdManager()
	for i := 0; i <= math.MaxUint8; i += 1 {
		// empty out the manager
		_, _ = manager.Pop()
	}
	// make sure the manager is empty
	if !manager.IsEmpty() {
		t.Error("Manager should be empty")
		return
	}
	const exp uint8 = 42

	err := manager.Push(exp)
	if err != nil {
		t.Errorf("The manager should be fillable.")
		return
	}
	val, err := manager.Pop()

	if err != nil {
		t.Errorf("The manager should not be empty.")
		return
	}

	if val != exp {
		t.Errorf("Value should be the same as the expected one")
		return
	}
}
