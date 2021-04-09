package helpers_test

import (
	"sync"
	"testing"

	"github.zhaw.ch/raycasters/server/protocol/client"
	"github.zhaw.ch/raycasters/server/protocol/server"
	"github.zhaw.ch/raycasters/server/room/helpers"
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

	_, h := helpers.NewWorkQueue()

	wg.Add(1)
	go func() {
		defer func() {
			// testing double shutdown
			h.Shutdown()
			wg.Done()
		}()
		_, err := h.Dequeue()
		if err == nil {
			t.Error("shutdown expected")
		}
	}()

	h.Shutdown()

	wg.Wait()
}

func TestCommunications(t *testing.T) {
	var wg sync.WaitGroup

	c, h := helpers.NewWorkQueue()

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
