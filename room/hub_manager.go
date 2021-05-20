package room

import (
	"context"
	"errors"
	"sync"

	"github.zhaw.ch/raycasters/server/gamemap"
	"github.zhaw.ch/raycasters/server/room/controller"
	"github.zhaw.ch/raycasters/server/utils"
)

var ErrHubUnavailable = errors.New("the given hub does not currnently exist")

var ErrHubManagerNotActive = errors.New("the hub manager is currently not active")

// HubManager is a manager of all Hubs.
// Is manages all running Hubs
type HubManager struct {
	hubs     map[int]*controller.Hub
	mapIter  gamemap.MapIteratorManager
	isActive utils.AtomicBool
	lock     *sync.RWMutex
}

// NewHubManager will create a new manager
func NewHubManager(filePath string) (HubManager, error) {
	mim, err := gamemap.NewMapIteratorManager(filePath)
	if err != nil {
		return HubManager{}, err
	}
	return HubManager{
		hubs:     make(map[int]*controller.Hub),
		mapIter:  mim,
		isActive: utils.NewAtomicBool(true),
		lock:     &sync.RWMutex{},
	}, nil
}

// Create a new Hub
func (hm *HubManager) AddHub(id int) error {
	if !hm.isActive.Get() {
		return ErrHubManagerNotActive
	}
	hm.lock.Lock()
	defer hm.lock.Unlock()

	_, ok := hm.hubs[id]

	if ok {
		return ErrHubExists{
			id: id,
		}
	}
	hub := controller.NewHub(id, hm.mapIter.GetIterator())

	hm.hubs[id] = hub

	return hub.Run()
}

// Remove Hub
func (hm *HubManager) RemoveHub(id int) error {
	if !hm.isActive.Get() {
		return ErrHubManagerNotActive
	}
	hm.lock.Lock()
	defer hm.lock.Unlock()
	_, ok := hm.hubs[id]
	if !ok {
		return ErrHubNotExisting{
			id: id,
		}
	}
	delete(hm.hubs, id)
	return nil
}

func (hm *HubManager) GetHub(id int) (*controller.Hub, error) {
	if !hm.isActive.Get() {
		return nil, ErrHubManagerNotActive
	}
	hm.lock.RLock()
	defer hm.lock.RUnlock()
	if id < 0 || id >= len(hm.hubs) {
		return nil, ErrHubUnavailable
	}

	return hm.hubs[id], nil
}

func (hm *HubManager) Shutdown(cxt context.Context) error {
	if !hm.isActive.SetFalse() {
		return ErrHubManagerNotActive
	}
	hm.lock.RLock()
	defer hm.lock.RUnlock()

	waiter := make(chan struct{})
	var wg sync.WaitGroup

	wg.Add(len(hm.hubs))

	go func() {
		wg.Wait()
		close(waiter)
	}()

	for _, hub := range hm.hubs {
		go func(h *controller.Hub) {
			defer wg.Done()

			h.Shutdown(cxt)
		}(hub)
	}

	select {
	case <-waiter:
		Logger.Info("All Hubs have finished shuttting down")
	case <-cxt.Done():
		Logger.Info("The Hub Manager has been forcefully shutdown")
	}
	return nil
}
