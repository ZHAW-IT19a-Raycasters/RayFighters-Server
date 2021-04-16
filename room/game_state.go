package room

import (
	"fmt"
	"sync"

	"github.zhaw.ch/raycasters/server/protocol/server"
	"github.zhaw.ch/raycasters/server/room/client"
)

// ErrClientDoesNotExist will be returned in cases
// where the specified client does not yet exist.
type ErrClientDoesNotExist struct {
	id uint8
}

// Error satisfies the error
func (err ErrClientDoesNotExist) Error() string {
	return fmt.Sprintf("the given client [%d] already exists", err.id)
}

// Is satisfies the error.Is configuration
func (err ErrClientDoesNotExist) Is(target error) bool {
	_, ok := target.(ErrClientDoesNotExist)
	return ok
}

// ErrClientExists is the error that will be returned
// if the given client does already exists.
type ErrClientExists struct {
	id uint8
}

// Error satisfies the error
func (err ErrClientExists) Error() string {
	return fmt.Sprintf("the given client [%d] already exists", err.id)
}

// Is satisfies the error.Is configuration
func (err ErrClientExists) Is(target error) bool {
	_, ok := target.(ErrClientExists)
	return ok
}

// GameState represents the current state of the
// game board.
type GameState struct {
	clients map[uint8]*client.ClientGame
	lock    *sync.RWMutex
}

// NewGameState create a new GameState object
func NewGameState() GameState {
	return GameState{
		clients: make(map[uint8]*client.ClientGame),
		lock:    &sync.RWMutex{},
	}
}

// GetClient will return a the state of a specific client.
// And if the client could not be found false will be returned
func (gs GameState) GetClient(id uint8) (*client.ClientGame, bool) {
	gs.lock.RLock()
	defer gs.lock.RUnlock()
	c, ok := gs.clients[id]
	if !ok {
		return nil, ok
	}
	return c, ok
}

// GetPlayersKD will return the player kd of all the currently
// logged in players.
func (gs GameState) GetPlayersKD() server.PlayersKD {
	gs.lock.RLock()
	defer gs.lock.RUnlock()
	playersKD := make([]server.PlayerKD, 0, len(gs.clients))
	for id, c := range gs.clients {
		playersKD = append(playersKD, server.NewPlayerKD(id, c.GetKills(), c.GetDeaths()))
	}
	return server.NewPlayersKD(playersKD)
}

// IncKill will increment the kill counter
func (gs *GameState) IncKill(id uint8) {
	gs.lock.RLock()
	defer gs.lock.RUnlock()
	gs.clients[id].IncKills()
}

// IncDeaths will increment the death counter
func (gs *GameState) IncDeaths(id uint8) {
	gs.lock.RLock()
	defer gs.lock.RUnlock()
	gs.clients[id].IncDeaths()
}

// AddClient will add a new client if it does not
// yet exists else this function will return an
// error.
func (gs *GameState) AddClient(id uint8) error {
	gs.lock.Lock()
	defer gs.lock.Unlock()

	_, ok := gs.clients[id]

	if ok {
		return ErrClientExists{
			id: id,
		}
	}
	c := client.NewClientGame()

	gs.clients[id] = &c
	return nil
}

// RemoveClient will remove a client if possible
func (gs *GameState) RemoveClient(id uint8) error {
	gs.lock.Lock()
	defer gs.lock.Unlock()
	_, ok := gs.clients[id]
	if !ok {
		return ErrClientDoesNotExist{
			id: id,
		}
	}
	delete(gs.clients, id)
	return nil
}
