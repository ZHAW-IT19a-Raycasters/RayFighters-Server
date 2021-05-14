package room

import (
	"github.zhaw.ch/raycasters/server/room/controller"
	"sync"
)

// HubManager is a manager of all Hubs.
// Is manages all running Hubs
type HubManager struct {
	hubs 	map[int]*controller.Hub
	channel MapIteratorManager
	lock    *sync.RWMutex
}

// NewHubManager will create a new manager
func NewHubManager(filePath string) (HubManager, error) {
	mim, err := NewMapIteratorManager(filePath)
	if err != nil {
		return HubManager{}, err
	}
	return HubManager{
		hubs: 	 make(map[int]*controller.Hub),
		channel: mim,
		lock:    &sync.RWMutex{},
	}, nil
}

// Create a new Hub
func (hm HubManager) AddHub(id int) error {
	hm.lock.Lock()
	defer hm.lock.Unlock()

	_, ok := hm.hubs[id]

	if ok {
		return ErrHubExists{
			id: id,
		}
	}
	c := controller.NewHub(id)

	hm.hubs[id] = c
	return nil
}

// Remove Hub
func (hm HubManager) RemoveHub(id int) error {
	hm.lock.Lock()
	defer hm.lock.Unlock()
	_, ok := hm.hubs[id]
	if !ok {
		return ErrHubExists{
			id: id,
		}
	}
	delete(hm.hubs, id)
	return nil
}
