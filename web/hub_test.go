package web

import (
	"testing"

	"github.com/gorilla/websocket"
)

func TestHubRegisterClient(t *testing.T) {
	hub := newHub()
	go hub.run()

	if len(hub.clients) != 0 {
		t.Errorf("Clinets map supposed to be empty")
	}

	client1 := &Client{hub: hub, conn: &websocket.Conn{}, send: make(chan []byte, 256)}
	client1.hub.register <- client1

	if len(hub.clients) != 1 {
		t.Errorf("Clients map supposed to contain a client")
	}

	if hub.clients[client1] == false {
		t.Errorf("Client registration failed!")
	}
}
