package web

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/rs/cors"
	"github.zhaw.ch/raycasters/server/log"
	"github.zhaw.ch/raycasters/server/room"
	"github.zhaw.ch/raycasters/server/room/client"
	"github.zhaw.ch/raycasters/server/web/helpers"

	webconfig "github.zhaw.ch/raycasters/server/web/config"
)

const hubsCount int = 3

var logger = log.Log()

// Server holds all endpoints for a server
type Server struct {
	endpoint *http.Server
}

// ErrRoomIdUnavailable can be thrown when a room is requested that cannot currently be accessed.
var ErrRoomIdUnavailable = errors.New("No Room available for the given ID")

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
		logger.Warnf("failed to generate private key: %s", err)
	}

	// init global jwt decoding
	jwt.RegisterCustomField("uuid", uuid.UUID{})

	amw := authenticationMiddleware{privKey}

	router := mux.NewRouter()

	// serve /auth without authentication middleware
	router.Path("/auth").
		Handler(http.HandlerFunc(ServeAuthWithKey(privKey))).
		Methods(http.MethodPost, http.MethodOptions)

	// create a sub router for the blocked paths
	api := router.PathPrefix("/").Subrouter()
	api.Use(amw.Middleware)
	api.HandleFunc("/", serveRoot)

	api.HandleFunc("/ws/{roomID:[0-9]+}", func(w http.ResponseWriter, r *http.Request) {
		// TODO: get a hub reference from handler function via HubManager
		// handle with a wrapper lambda for now
		serveWs(hubs[:], w, r)
	})

	// handle CORS-requests
	// see https://de.wikipedia.org/wiki/Cross-Origin_Resource_Sharing
	// for more information about the topic
	corsHandler := cors.Default()

	httpServer := &http.Server{
		Handler:      corsHandler.Handler(router),
		Addr:         conf.String(),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	return &Server{
		endpoint: httpServer,
	}
}

// Start starts listening and serving the configured server.
// Blocking.
func (p *Server) Start() error {
	logger.Infof("Server listening on: %s", p.endpoint.Addr)
	return p.endpoint.ListenAndServe()
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
	if roomID < 0 || roomID >= hubsCount {
		logger.Error(ErrRoomIdUnavailable)
		http.Error(w, ErrRoomIdUnavailable.Error(), http.StatusConflict)
		return
	}

	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	// upgrade HTTP to Websocket connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		// upgrader itself writes http error response
		logger.Error(err)
		return
	}

	// get cookie from request
	jwtCookie, err := r.Cookie(JWTCOOKIENAME)
	if err != nil {
		http.Error(w, "no JWT found in Cookies", http.StatusUnauthorized)
		return
	}

	// unmarshal the token
	token, err := jwt.ParseString(jwtCookie.Value)
	usernameI, userOk := token.Get(`username`)
	uuidI, uuidOk := token.Get(`uuid`)

	if userOk && uuidOk {
		uuid := uuidI.(uuid.UUID)
		username := fmt.Sprintf("%v", usernameI)
		cm := client.NewClientMeta(uuid, username, conn)
		hubs[roomID].RegisterClient(cm)
		logger.Infof("user %s with uuid %s joined room %d", username, uuid, roomID)
	}

}
