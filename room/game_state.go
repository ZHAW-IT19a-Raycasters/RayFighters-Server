package room

import (
	"sync"

	"github.zhaw.ch/raycasters/server/gamemap"
	"github.zhaw.ch/raycasters/server/protocol/server"
	"github.zhaw.ch/raycasters/server/room/client"
)

type GameState struct {
	clients map[uint8]*client.ClientGame
	lock    *sync.RWMutex
}

func NewGameState() GameState {
	return GameState{
		clients: make(map[uint8]*client.ClientGame),
		lock:    &sync.RWMutex{},
	}
}

func (gs GameState) GetClient(id uint8) (*client.ClientGame, bool) {
	gs.lock.RLock()
	defer gs.lock.RUnlock()
	c, ok := gs.clients[id]
	return c, ok
}

func (gs GameState) GetPlayersKD() server.PlayersKD {
	gs.lock.RLock()
	defer gs.lock.RUnlock()
	playersKD := make([]server.PlayerKD, 0, len(gs.clients))
	for id, c := range gs.clients {
		playersKD = append(playersKD, server.NewPlayerKD(id, c.GetKills(), c.GetDeaths()))
	}
	return server.NewPlayersKD(playersKD)
}

func (gs *GameState) IncKill(id uint8) {
	gs.lock.RLock()
	defer gs.lock.RUnlock()
	gs.clients[id].IncKills()
}

func (gs *GameState) IncDeaths(id uint8) {
	gs.lock.RLock()
	defer gs.lock.RUnlock()
	gs.clients[id].IncDeaths()
}

func (gs *GameState) AddClient(id uint8) error {
	gs.lock.Lock()
	defer gs.lock.Unlock()
	_, ok := gs.clients[id]
	if ok {
		// TODO: FIND A MORE ACCURATE ERROR
		return ErrHubIsFull
	}
	c := client.NewClientGame(gamemap.NewCoordinate())

	gs.clients[id] = &c
	return nil
}

func (gs *GameState) RemoveClient(id uint8) error {
	gs.lock.Lock()
	defer gs.lock.Unlock()
	_, ok := gs.clients[id]
	if !ok {
		// TODO: FIND A MORE ACCURATE ERROR
		return ErrHubIsFull
	}
	delete(gs.clients, id)
	return nil
}
