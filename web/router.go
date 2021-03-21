package web

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	webconfig "github.zhaw.ch/raycasters/server/web/config"
)

// ZIEL: ein neuer ROUTER
type Server struct {
	endpoint *http.Server
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func NewServer(conf *webconfig.WebConfig) *Server {
	//TODO replace
	hub := newHub()
	go hub.run()
	router := mux.NewRouter()
	router.HandleFunc("/", handleRoot)
	router.HandleFunc("/chat", serveChat)

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})

	httpServer := &http.Server{Handler: router, Addr: conf.String(), WriteTimeout: 15 * time.Second, ReadTimeout: 15 * time.Second}

	server := &Server{endpoint: httpServer}
	return server
}

func (p *Server) Start() error {
	return p.endpoint.ListenAndServe()
}

func ShutdownContext() context.Context {
	return context.Background()
}

func (p *Server) ServerStop(ctx context.Context) error {
	// TODO: Gracful shutdown
	return p.endpoint.Shutdown(ctx)
}

func handleRoot(writer http.ResponseWriter, request *http.Request) {
	//ResponseWriter implements io.Writer
	fmt.Fprintf(writer, "Hello world from %s", request.URL.Path[1:]) //ich will alles ausser 0-te element
}

func serveChat(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, "/web/chat.html")
}

// serveWs handles websocket requests from the peer.
func serveWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := &Client{hub: hub, conn: conn, send: make(chan []byte, 256)}
	client.hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
}
