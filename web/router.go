package web

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.zhaw.ch/raycasters/server/log"
	"github.zhaw.ch/raycasters/server/room"

	webconfig "github.zhaw.ch/raycasters/server/web/config"
)

const hubsCount int = 3

var logger = log.Log()

// Server holds all endpoints for a server
type Server struct {
	endpoint *http.Server
}

// NewServer registers routes for all exposed endpoints
// and attaches the appropriate handler functions
func NewServer(conf *webconfig.WebConfig) *Server {

	// implement only 1 hub for now
	//hub := room.NewHub(1)
	var hub [hubsCount]*(room.Hub)
	var i int = 0
	for ; i < hubsCount; i++ {
		hub[i] = room.NewHub(i)
		go hub[i].Run()
	}

	router := mux.NewRouter()
	router.HandleFunc("/", handleRoot)
	router.HandleFunc("/auth", serveAuth).Methods("POST")
	// TODO: improve regex for uint32
	router.HandleFunc("/ws/{roomID:[0-9]+}", func(w http.ResponseWriter, r *http.Request) {
		// parse roomID from URL
		vars := mux.Vars(r)
		roomIDStr := vars["roomID"]
		//roomID, err := strconv.ParseInt(roomIDStr, 10, 32)
		roomID, err := strconv.Atoi(roomIDStr)
		if err != nil {
			fmt.Printf("couldn't parse roomID from URL: %v\n", err)
			return
		}

		if roomID < hubsCount {
			logger.Infof("request to join room %d", roomID)
			serveWs(hub[roomID], w, r)
		}
	})
	router.HandleFunc("/chat", serveChat)

	httpServer := &http.Server{Handler: router, Addr: conf.String(), WriteTimeout: 15 * time.Second, ReadTimeout: 15 * time.Second}

	server := &Server{endpoint: httpServer}
	return server
}

// Start starts listening and serving the configured server.
// Blocking.
func (p *Server) Start() error {
	fmt.Println("Server listening on: ", p.endpoint.Addr)
	return p.endpoint.ListenAndServe()
}

// ShutdownContext returns server's shutdown context.
func ShutdownContext() context.Context {
	return context.Background()
}

// ServerStop will initialize a graceful shutdown of the server.
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
	http.ServeFile(w, r, "web/chat.html")
}

// serveWs handles websocket requests from the peer.
func serveWs(hub *room.Hub, w http.ResponseWriter, r *http.Request) {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	// create client meta
	uuid := uuid.New()
	cm := room.NewClientMeta(uuid, "john doe", conn)

	hub.RegisterClient(cm)
}

func serveAuth(w http.ResponseWriter, r *http.Request) {}

// create JWT and uuid here
