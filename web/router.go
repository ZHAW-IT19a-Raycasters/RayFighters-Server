package web

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.zhaw.ch/raycasters/server/log"
	"github.zhaw.ch/raycasters/server/room"
	"github.zhaw.ch/raycasters/server/web/helpers"

	webconfig "github.zhaw.ch/raycasters/server/web/config"
)

const hubsCount int = 3
var logger = log.Log()

// Server holds all endpoints for a server
type Server struct {
	endpoint *http.Server
	privateKey *rsa.PrivateKey
}

// NewServer registers routes for all exposed endpoints
// and attaches the appropriate routehandler functions
func NewServer(conf *webconfig.WebConfig) *Server {
	// instantiate hubs
	var hubs [hubsCount]*room.Hub
	var i int = 0
	for ; i < hubsCount; i++ {
		hubs[i] = room.NewHub(i)
		go hubs[i].Run()
	}

	// generate private key
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		fmt.Printf("failed to generate private key: %s\n", err)
	}

	amw := authenticationMiddleware{privKey}

	router := mux.NewRouter()
	// serve /auth without authentication middleware
	router.Path("/auth").Handler(http.HandlerFunc(ServeAuthWithKey(privKey))).Methods("POST")
	api := router.PathPrefix("/").Subrouter()
	api.HandleFunc("/", serveRoot)
	api.HandleFunc("/ws/{roomID:[0-9]+}", func(w http.ResponseWriter, r *http.Request) {
		// TODO: get a hub reference from handler function via HubManager
		// handle with a wrapper lambda for now
		serveWs(hubs[:], w, r)
	})

	
	api.Use(amw.Middleware)

	httpServer := &http.Server{Handler: router, Addr: conf.String(), WriteTimeout: 15 * time.Second, ReadTimeout: 15 * time.Second}
	
	server := &Server{endpoint: httpServer, privateKey: privKey}
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
	return p.endpoint.Shutdown(ctx)
}

// serveRoot serves a simple landing page.
func serveRoot(writer http.ResponseWriter, request *http.Request) {
	//ResponseWriter implements io.Writer
	//server answer to a client
	fmt.Fprintf(writer, "Hello from RayCasters!")
}

// serveWs upgrades HTTP request to WebSocket connection
// parses desired roomID from the request URL
// and registers the client to the hub
func serveWs(hubs []*room.Hub, w http.ResponseWriter, r *http.Request) {
	
	roomID, err := helpers.ParseRoomID(r)
	if err != nil {
		logger.Error(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if (roomID < 0 || roomID >= hubsCount) {
		err := fmt.Errorf("no room with id %d available",roomID)
		logger.Error(err)
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	// upgrade HTTP to Websocket connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error(err)
		return
	}

	logger.Infof("%s request to join room %d", r.RemoteAddr, roomID)

	// create and register client
	uuid := uuid.New()
	cm := room.NewClientMeta(uuid, "john dude", conn)

	hubs[roomID].RegisterClient(cm)
}
