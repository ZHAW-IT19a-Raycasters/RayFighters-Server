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
	// instantiate multiple hubs
	var hub [hubsCount]*(room.Hub)
	var i int = 0
	for ; i < hubsCount; i++ {
		hub[i] = room.NewHub(i)
		go hub[i].Run()
	}

	router := mux.NewRouter()
	router.HandleFunc("/", handleRoot)
	router.HandleFunc("/auth", serveAuth).Methods("POST")

	router.HandleFunc("/ws/{roomID:[0-9]+}", func(w http.ResponseWriter, r *http.Request) {
		// note: variables are strings
		vars := mux.Vars(r)

		// parse roomID from URL
		roomIDStr := vars["roomID"]
		roomID, err := strconv.Atoi(roomIDStr)
		if err != nil {
			logger.Errorf("couldn't parse roomID from URL: %v\n", err)
			return
		}
		if roomID >= hubsCount {
			logger.Errorf("request to join room with illegal room-ID \"%d\"", roomID)
			return
		}

		logger.Infof("%s request to join room %d", r.RemoteAddr, roomID)
		serveWs(hub[roomID], w, r)
	})
	router.HandleFunc("/chat", serveChat)

	httpServer := &http.Server{Handler: router, Addr: conf.String(), WriteTimeout: 15 * time.Second, ReadTimeout: 15 * time.Second}

	server := &Server{endpoint: httpServer}
	return server
}

// Start starts listening and serving the configured server.
// Blocking.
func (p *Server) Start() error {
	logger.Infof("Server listening on: %s", p.endpoint.Addr)
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

// handleRoot is the handler of "<server listening on>/".
func handleRoot(writer http.ResponseWriter, request *http.Request) {
	//ResponseWriter implements io.Writer
	//server answer to a client
	fmt.Print(request.URL)
	fmt.Fprintf(writer, "Hello from RayCasters!")
}

// serveChat is the handler of <server listening on>/chat.
// Used for testing puposes.
func serveChat(w http.ResponseWriter, r *http.Request) {
	logger.Infof("%s requested", r.URL)
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
		logger.Error(err)
		return
	}

	// create client meta
	uuid := uuid.New()
	cm := room.NewClientMeta(uuid, "john dude", conn)

	hub.RegisterClient(cm)
}

func serveAuth(w http.ResponseWriter, r *http.Request) {}

// create JWT and uuid here
