package client

import (
	"errors"
	"sync"
	"testing"

	"github.zhaw.ch/raycasters/server/protocol"
	"github.zhaw.ch/raycasters/server/protocol/client"
	"github.zhaw.ch/raycasters/server/protocol/server"
)

func compareByte(expected, gotten []byte, t *testing.T) bool {
	lexp := len(expected)
	lgot := len(gotten)
	if lexp != lgot {
		t.Error("incorrect size")
		return false
	}

	for i := range expected {
		exp := expected[i]
		got := gotten[i]
		if exp != got {
			t.Errorf("incorrect bytes at position [%d] - exp: [0x%X] got: [0x%X]", i, exp, got)
			return false
		}
	}
	return true
}

func TestQueueShutdown(t *testing.T) {
	var wg sync.WaitGroup

	_, h := NewWorkQueue()
	blocker := make(chan struct{})
	wg.Add(1)
	go func() {
		defer func() {
			// testing double shutdown
			h.Shutdown()
			wg.Done()
		}()
		close(blocker)
		_, err := h.Dequeue()
		if err == nil {
			t.Error("shutdown expected")
		} else if !errors.Is(err, ErrQueueClosed) {
			t.Error("Wrong error used here")
		}
	}()
	<-blocker
	h.Shutdown()

	wg.Wait()

	// check that functions return correct error
	// these should not block, so they are run
	// here
	err := h.Enqueue([]byte{42})
	if err == nil {
		t.Error("There should be an error here")
		return
	} else if !errors.Is(err, ErrQueueClosed) {
		t.Error("The error should indicate that the queue is closed")
		return
	}

	_, err = h.Dequeue()
	if err == nil {
		t.Error("There should be an error here")
		return
	} else if !errors.Is(err, ErrQueueClosed) {
		t.Error("The error should indicate that the queue is closed")
		return
	}
}

func TestQueueCommunications(t *testing.T) {
	var wg sync.WaitGroup

	c, h := NewWorkQueue()

	request := []byte{client.DisconnectOpcode.Value()}
	response := []byte{server.DisconnectOpcode.Value()}

	// start the hub
	wg.Add(2)
	go func() {
		defer func() {
			wg.Done()
		}()

		data, err := h.Dequeue()
		if err != nil {
			t.Error("there should not be an error here, as the queue, should be open")
			h.Shutdown()
			return
		}

		if !compareByte(request, data, t) {
			h.Shutdown()
			return
		}

		err = h.Enqueue(response)
		if err != nil {
			t.Error("there should not be an error here, as the queue should still be open")
		}
	}()

	// start the client
	go func() {
		defer func() {
			c.Shutdown()
			wg.Done()
		}()

		err := c.Enqueue(request)
		if err != nil {
			t.Error("the queue should still be open here")
			return
		}

		data, err := c.Dequeue()
		if err != nil {
			t.Error("there should not be an error here, as the queue, should be open")
			return
		}

		if !compareByte(response, data, t) {
			return
		}
	}()

	wg.Wait()
}

func TestQueueDrain(t *testing.T) {

	c, h := NewWorkQueue()
	exp := make([]protocol.Package, 0, QueueSize)
	for i := 0; i < QueueSize; i += 1 {
		t := []byte{byte(i)}
		exp = append(exp, t)
		h.Enqueue(t)
	}

	got, err := c.DequeueDrain()

	if err != nil {
		t.Error("The queue should be active")
		return
	}

	lgot := len(got)
	lexp := len(exp)

	if lexp != lgot {
		t.Errorf("Incorrect data lenght exp: [%d] got:[%d]", lexp, lgot)
		return
	}

	for i := range got {
		if compareByte(exp[i], got[i], t) {
			return
		}
	}
}

func TestQueueDrainForce(t *testing.T) {
	// this test does not show anything speciall
	// it merely servers to get 100% coverage
	c, h := NewWorkQueue()

	// force the queue into draining mode
	c.isDraining.SetTrue()
	err := h.Enqueue([]byte{0})
	if err == nil {
		t.Error("Should be an error showing the draining error")
		return
	}

	_, err = c.Dequeue()
	if err == nil {
		t.Error("Should be an error showing the draining")
		return
	}

	_, err = c.DequeueDrain()
	if err == nil {
		t.Error("Should be an error showing the draining")
	}
}
