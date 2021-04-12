package room

import (
	"github.zhaw.ch/raycasters/server/protocol/client"
	"testing"
)

func TestHandleRead(t *testing.T) {
	hub := newHubController()
}

func TestHandleUnknownReceived(t *testing.T) {
	hub := newHubController()
	magic := hub.handleUnknownReceived([]byte{client.UnknownOpcode.Value()})
}

func TestHandleDisconnect(t *testing.T) {
	hub := newHubController()
	magic := hub.handleDisconnect([]byte{client.DisconnectOpcode.Value()}, 1)
}

func TestHandleWhoAmI(t *testing.T) {
	hub := newHubController()
	magic := hub.handleWhoAmI([]byte{client.WhoAmIOpcode.Value()}, 1)
}

func TestHandleMovement(t *testing.T) {
	hub := newHubController()
	magic := hub.handleMovement([]byte{client.MovementOpcode.Value()}, 1)
}

func TestHandleShooting(t *testing.T) {
	hub := newHubController()
	magic := hub.handleShooting([]byte{client.ShootingOpcode.Value()}, 1)
}

func TestHandleRequestOtherPlayer(t *testing.T) {
	hub := newHubController()
	magic := hub.handleRequestOtherPlayer([]byte{client.RequestOtherPlayersOpcode.Value()}, 1)
}
func TestHandleRequestPlayerKD(t *testing.T) {
	hub := newHubController()
	magic := hub.handleRequestPlayerKD([]byte{client.RequestPlayerKDOpcode.Value()}, 1)
}