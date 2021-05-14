package room

import (
	"github.zhaw.ch/raycasters/server/room/controller"
	"sync"
)

// HubManager is a manager of all Hubs.
// Is manages all running Hubs
type HubManager struct {
	hubs 	map[uint8]*controller.Hub
	lock    *sync.RWMutex
}

// NewHubManager will create a new manager
func NewHubManager() HubManager {
	return HubManager{
		hubs: 	 make(map[uint8]*controller.Hub),
		lock:    &sync.RWMutex{},
	}
}

// Create a new Hub
func (hm HubManager) AddHub(id uint8) error {
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
func (hm HubManager) RemoveHub(id uint8) error {
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
