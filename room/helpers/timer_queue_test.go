package helpers

import (
	"testing"
	"time"
)

func setupTimer(t *testing.T, f func() time.Time) {
	org := getTime
	getTime = f

	t.Cleanup(func() {
		getTime = org
	})
}

func setupTimerDefault(t *testing.T) time.Time {
	timeToTest := time.Now()

	setupTimer(t, func() time.Time {
		return timeToTest
	})
	return timeToTest
}

func TestTimer(t *testing.T) {
	const limit = 10
	const expId = 1
	expTime := setupTimerDefault(t)

	ht := NewHubTimer(limit)
	ct := ht.NewClientTimer(expId)
	ct.RegisterTimeout()

	got := <-ht.GetQueue()

	if expTime != got.GetTime() {
		t.Error("incorrect time was returned")
		return
	} else if got.GetId() != expId {
		t.Error("Incorrect id used")
	}
}
