package helpers_test

import (
	"errors"
	"testing"

	"github.zhaw.ch/raycasters/server/room/helpers"
)

func TestIdSetAdd(t *testing.T) {
	const id uint8 = 66
	set := helpers.NewIdSet()

	// test contains
	if set.Has(id) {
		t.Error("set should not contain the id")
		return
	}

	err := set.Add(id)
	if err != nil {
		t.Error("id should not be in set yet")
		return
	}

	// test contains
	if !set.Has(id) {
		t.Error("set should contain the id")
		return
	}

	// test double add
	err = set.Add(id)
	if err == nil {
		t.Error("id already has been added to the set")
		return
	} else if !errors.Is(err, helpers.ErrIdAlready{}) {
		t.Error("wrong error returned")
		return
	} else if err.Error() != "such id already available [66]" {
		t.Error("Incorrect error message returned")
	}
}

func TestIdSetRemove(t *testing.T) {
	const id uint8 = 66
	set := helpers.NewIdSet()

	// test contains
	if set.Has(id) {
		t.Error("set should not contain the id")
		return
	}
	// error not important as this should work
	_ = set.Add(id)

	// test contains
	if !set.Has(id) {
		t.Error("set should contain the id")
		return
	}

	err := set.Remove(id)
	if err != nil {
		t.Error("as the set contains the id an error should have been returned here!")
		return
	}

	err = set.Remove(id)

	if err == nil {
		t.Error("set should contain the id")
		return
	} else if !errors.Is(err, helpers.ErrNoId{}) {
		t.Error("Incorrect error returned")
		return
	} else if err.Error() != "no such id available [66]" {
		t.Error("Incorrect error message returned")
	}

}
