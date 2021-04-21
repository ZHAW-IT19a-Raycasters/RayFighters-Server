package gamemap

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

func TestLoadMapPanic(t *testing.T) {
	// Turn off the panic.
	defer func() { recover() }()
	LoadMap("NonExistingPath")
	// Never reaches here if `LoadMap` panics.
	t.Errorf("should have panicked")
}

func TestLoadMapNoPanic(t *testing.T) {
	fmt.Println(os.Getwd())
	defer func() {
		if recover() != nil {
			t.Errorf("should not have panicked")
		}
	}()
	file, err := ioutil.TempFile(".", "tempfile")
	if err != nil {
		t.Error(err)
		return
	}
	// "defer" stacks
	defer os.Remove(file.Name())
	if err != nil {
		t.Error(err)
		return
	}
	defer file.Close()
	if err != nil {
		t.Error(err)
		return
	}
	LoadMap(file.Name())
}
