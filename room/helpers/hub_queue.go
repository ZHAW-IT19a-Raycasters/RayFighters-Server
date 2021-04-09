package helpers

import (
	"errors"

	"github.zhaw.ch/raycasters/server/protocol"
	"github.zhaw.ch/raycasters/server/utils"
)

// QueueSize is the max buffer size of the queue
const QueueSize = 10

// ErrQueueClosed is the error that is returned, if the
// queue invariant is not fulfilled.
var ErrQueueClosed = errors.New("queue is closed")

// HubQueue is the connection used by the hub to talk to the
// client.
type HubQueue struct {
	packageQueue
}

// newHubQueue is the connections over which the hub can connect to the
// client.
func newHubQueue(in, out chan protocol.Package, running chan struct{}, active *utils.AtomicBool) HubQueue {
	return HubQueue{
		newPackageQueue(in, out, running, active),
	}
}

// ClientQueue is the connection used by the client to
// communicate with the server.
type ClientQueue struct {
	packageQueue
}

// newClientQeueue will create a new client queue
func newClientQeueue(in, out chan protocol.Package, running chan struct{}, active *utils.AtomicBool) ClientQueue {
	return ClientQueue{
		newPackageQueue(in, out, running, active),
	}
}

// packageQueue contains the two queues used for inter thread comminication.
type packageQueue struct {
	// running is used a a two way shutdown signal
	// and hinders eternal blocking on the channels.
	running  chan struct{}
	isActive *utils.AtomicBool
	inQueue  <-chan protocol.Package
	outQueue chan<- protocol.Package
}

// newPackageQueue will create a new package queue with which to
// communicate.
func newPackageQueue(in, out chan protocol.Package, running chan struct{}, active *utils.AtomicBool) packageQueue {
	return packageQueue{
		running:  running,
		isActive: active,
		inQueue:  in,
		outQueue: out,
	}
}

// Enqueue will put the package onto the out queue, so that the
// counter part can use, for work.
func (pq packageQueue) Enqueue(data protocol.Package) error {
	if !pq.isActive.Get() {
		return ErrQueueClosed
	}

	pq.outQueue <- data
	return nil
}

// Dequeue will block the gorutine until a new package from
// the paired until will be received.
// Attention use this function inside of a select, or
// else this function will block all your executations.
func (pq packageQueue) Dequeue() (protocol.Package, error) {
	select {
	case <-pq.running:
		return nil, ErrQueueClosed
	case d := <-pq.inQueue:
		return d, nil
	}
}

// Shutdown will shutdown the transmitions.
// Onces they shutdown there is no way to restart them.
func (pq packageQueue) Shutdown() {
	if !pq.isActive.SetFalse() {
		return
	}
	close(pq.running)
}

// NewWorkQueue will create a pair of hubs
func NewWorkQueue() (ClientQueue, HubQueue) {
	active := utils.NewAtomicBool(true)
	pactive := &active
	running := make(chan struct{})
	aChan := make(chan protocol.Package, QueueSize)
	bChan := make(chan protocol.Package, QueueSize)
	client := newClientQeueue(aChan, bChan, running, pactive)
	hub := newHubQueue(bChan, aChan, running, pactive)
	return client, hub
}
