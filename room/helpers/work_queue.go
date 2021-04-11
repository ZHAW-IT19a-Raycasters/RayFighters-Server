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
var ErrQueueIsDraining = errors.New("queue is already draining")
var ErrQueueEmpty = errors.New("queue is empty")

// HubQueue is the connection used by the hub to talk to the
// client.
type HubQueue struct {
	packageQueue
}

// ClientQueue is the connection used by the client to
// communicate with the server.
type ClientQueue struct {
	packageQueue
}

// packageQueue contains the two queues used for inter thread comminication.
type packageQueue struct {
	// running is used a a two way shutdown signal
	// and hinders eternal blocking on the channels.
	running    chan struct{}
	isActive   *utils.AtomicBool
	isDraining *utils.AtomicBool
	inQueue    <-chan protocol.Package
	outQueue   chan<- protocol.Package
}

// newPackageQueue will create a new package queue with which to
// communicate.
func newPackageQueue(in, out chan protocol.Package, running chan struct{}, active, drain *utils.AtomicBool) packageQueue {
	return packageQueue{
		running:    running,
		isActive:   active,
		isDraining: drain,
		inQueue:    in,
		outQueue:   out,
	}
}

// Enqueue will put the package onto the out queue, so that the
// counter part can use, for work.
func (pq packageQueue) Enqueue(data protocol.Package) error {
	if !pq.isActive.Get() {
		return ErrQueueClosed
	} else if pq.isDraining.Get() {
		return ErrQueueIsDraining
	}

	pq.outQueue <- data

	return nil
}

// Dequeue will block the gorutine until a new package from
// the paired until will be received.
// Attention use this function inside of a select, or
// else this function will block all your executations.
func (pq packageQueue) Dequeue() (protocol.Package, error) {
	if !pq.isActive.Get() {
		return nil, ErrQueueClosed
	} else if pq.isDraining.Get() {
		return nil, ErrQueueIsDraining
	}
	// check if the queue is still running and then
	select {
	case <-pq.running:
		return nil, ErrQueueClosed
	case d := <-pq.inQueue:
		return d, nil
	}
}

// DequeueDrain will remove all the operations from the queue and
// return them.
// Attention this is an exclusive access call, so that
// no other caller can use the queue while this is draining
func (pq packageQueue) DequeueDrain() ([]protocol.Package, error) {
	// get exclusive access to the queue
	if !pq.isDraining.SetTrue() {
		return nil, ErrQueueIsDraining
	}
	// set the system back to running
	defer pq.isActive.Set(false)

	// preallocate to the max queue size possible
	data := make([]protocol.Package, 0, QueueSize)

	// itterate over channel until empty
	for {
		select {
		case value := <-pq.inQueue:
			data = append(data, value)
		default:
			return data, nil
		}
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

// NewWorkQueue will create a pair of connections queues
func NewWorkQueue() (ClientQueue, HubQueue) {
	active := utils.NewAtomicBool(true)
	pactive := &active
	drain := utils.NewAtomicBool(false)
	pdrain := &drain

	running := make(chan struct{})

	aChan := make(chan protocol.Package, QueueSize)
	bChan := make(chan protocol.Package, QueueSize)

	client := ClientQueue{
		newPackageQueue(aChan, bChan, running, pactive, pdrain),
	}
	hub := HubQueue{
		newPackageQueue(bChan, aChan, running, pactive, pdrain),
	}
	return client, hub
}
