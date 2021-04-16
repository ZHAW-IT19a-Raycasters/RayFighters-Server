package utils_test

import (
	"sync"
	"testing"

	"github.zhaw.ch/raycasters/server/utils"
)

const limit uint32 = 42

func TestSemaphoreBase(t *testing.T) {
	ac := utils.NewCountingSemaphore(limit)

	if ac.Value() != 0 {
		t.Error("Base value is incorrect.")
		return
	}

	if !ac.Empty() {
		t.Error("is zero should be false")
		return
	}

	if ac.Full() {
		t.Error("is filled should be false")
	}
}

func TestSemaphoreInc(t *testing.T) {
	ac := utils.NewCountingSemaphore(limit)

	err := ac.Acquire()

	if err != nil {
		t.Errorf("there was an error:\n%s", err.Error())
		return
	}

	val := ac.Value()

	if val != 1 {
		t.Error("value should be at one")
		return
	}

	// if ac.AtZero() {
	// 	t.Error("should not be zero")
	// }

	// if ac.AtLimit() {
	// 	t.Error("should not be filled up")
	// }

	for i := 0; i < int(limit)-1; i += 1 {
		err := ac.Acquire()
		if err != nil {
			t.Errorf("There was an error at position [%d]", i-1)
			return
		}
	}

	// if ac.AtZero() {
	// 	t.Error("should not be zero")
	// }

	// if !ac.AtLimit() {
	// 	t.Error("should be filled up")
	// }

	val = ac.Value()

	if val != int(limit) {
		t.Error("value should be at the given limit")
		return
	}

	// false increment
	err = ac.Acquire()

	if err == nil {
		t.Error("At the max inc there should have been an error.")
	}
}

func TestSemaphoreDec(t *testing.T) {
	ac := utils.NewCountingSemaphore(limit)

	ac.Acquire()

	val := ac.Value()

	if val != 1 {
		t.Error("value should be at one")
		return
	}

	err := ac.Release()

	if err != nil {
		t.Errorf("there was an error:\n%s", err.Error())
		return
	}

	val = ac.Value()

	if val != 0 {
		t.Error("value should be at one")
		return
	}
	if !ac.Empty() {
		t.Error("should not be zero")
	}

	if ac.Full() {
		t.Error("should not be filled up")
	}

	// below zero

	err = ac.Release()

	if err == nil {
		t.Error("As the value could have gone below zero, there should be an error here.")
	}
}

func TestSemaphoreFilledAndEmpty(t *testing.T) {
	ac := utils.NewCountingSemaphore(limit)

	for i := 0; i < int(limit); i += 1 {
		err := ac.Acquire()
		if err != nil {
			t.Errorf("There was an error at position [%d]", i)
			return
		}
	}

	if ac.Empty() {
		t.Error("should not be zero")
	}

	if !ac.Full() {
		t.Error("should be filled up")
	}

	// over filled so returns an error
	err := ac.Acquire()

	if err == nil {
		t.Error("There should be an error here telling us, that the counter is at it's limit")
		return
	}

	// decrement back to zero
	for i := int(limit - 1); i >= 0; i -= 1 {
		err := ac.Release()

		if err != nil {
			t.Errorf("There was an error at position [%d]", i)
			return
		}
	}

	if !ac.Empty() {
		t.Error("should be zero")
	}

	if ac.Full() {
		t.Error("should not be filled up")
	}
}

func TestSemaphoreFilledAndEmptyConcurrent(t *testing.T) {
	var wg sync.WaitGroup

	ac := utils.NewCountingSemaphore(limit)
	waiForDec := make(chan struct{})
	waitForAtLimit := make(chan struct{})

	go func() {
		// busy wait until filled up
		for !ac.Full() {
		}
		close(waitForAtLimit)
	}()

	for i := 0; i < int(ac.Limit()); i += 1 {
		wg.Add(1)
		go func(count int) {
			err := ac.Acquire()
			if err != nil {
				t.Errorf("There was an error at position [%d]", count)
				return
			}

			<-waiForDec

			err = ac.Release()
			if err != nil {
				t.Errorf("There was an error at position [%d]", count)
				return
			}
			wg.Done()
		}(i)
	}

	<-waitForAtLimit

	val := ac.Value()
	if val != int(ac.Limit()) {
		t.Error("Value should be at it's limit here")
		return
	}

	close(waiForDec)

	wg.Wait()

	val = ac.Value()
	if val != 0 {
		t.Error("Value should be at it's lower limit 0")
		return
	}
	if !ac.Empty() {
		t.Error("should be zero")
	}

	if ac.Full() {
		t.Error("should be not filled up")
	}
}
