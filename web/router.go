package web

import (
	"context"
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

const (
	hubsCount           int           = 3
	websocketBufferSize int           = 1024
	serverTimeout       time.Duration = 15 * time.Second
)

var logger = log.Log()

// Server holds all endpoints for a server
type Server struct {
	endpoint *http.Server
	auth     authenticationManager
	conf     *webconfig.WebConfig
}

// ErrRoomIdUnavailable can be thrown when a room is requested that cannot currently be accessed.
var ErrRoomIdUnavailable = errors.New("no Room available for the given ID")

// NewServer registers routes for all exposed endpoints
// and attaches the appropriate routehandler functions
func NewServer(conf *webconfig.WebConfig) (*Server, error) {
	// instantiate hubs
	var hubs [hubsCount]*room.Hub
	var i int = 0
	for ; i < hubsCount; i++ {
		hubs[i] = room.NewHub(i)
		go hubs[i].Run()
	}

	auth, err := newAuthManager(conf.JWTConfig.RsaPrivateKey)

	if err != nil {
		return nil, err
	}

	// init global jwt decoding
	jwt.RegisterCustomField("uuid", uuid.UUID{})

	router := mux.NewRouter()

	// serve /auth without authentication middleware
	router.Path("/auth").
		Handler(http.HandlerFunc(auth.AuthConnection())).
		Methods(http.MethodPost, http.MethodOptions)

	// create a sub router for the blocked paths
	api := router.PathPrefix("/").Subrouter()
	api.Use(auth.Middleware)

	api.HandleFunc("/", serveRoot)

	wsPath := fmt.Sprintf("/ws/{%s:[0-9]+}", helpers.RoomIdParameter)

	api.HandleFunc(wsPath, func(w http.ResponseWriter, r *http.Request) {
		// TODO: get a hub reference from handler function via HubManager
		// handle with a wrapper lambda for now
		serveWs(hubs[:], w, r)
	})

	// handle CORS-requests
	// see https://de.wikipedia.org/wiki/Cross-Origin_Resource_Sharing
	// for more information about the topic
	corsHandler := cors.New(
		cors.Options{
			AllowedOrigins:   conf.CORS.Domains,
			AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodOptions},
			AllowCredentials: true,
		},
	)

	httpServer := &http.Server{
		Handler:      corsHandler.Handler(router),
		Addr:         conf.Address(),
		WriteTimeout: serverTimeout,
		ReadTimeout:  serverTimeout,
	}

	return &Server{
		endpoint: httpServer,
		auth:     auth,
		conf:     conf,
	}, nil
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

	logger.Debug("http connection received")
	roomID, err := helpers.ParseRoomID(r)

	if err != nil {
		logger.Error(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if roomID < 0 || roomID >= len(hubs) {
		logger.Debugf("Room [%d] not available", roomID)
		http.Error(w, ErrRoomIdUnavailable.Error(), http.StatusBadRequest)
		return
	}

	user, err := parseUserInformation(w, r)
	if err != nil {
		return
	}

	// this check is in theory not threadsafe as there is a lot time happing
	// until the connection build up is done, but this can handle some
	// wrong connection creations, especially under heavy load.
	if hubs[roomID].Full() {
		w.WriteHeader(http.StatusConflict)
		return
	}

	logger.Debug("correct token received upgrading the connection")

	// upgrade HTTP to Websocket connection
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  websocketBufferSize,
		WriteBufferSize: websocketBufferSize,
		CheckOrigin: func(r *http.Request) bool {
			// this is okay, as the CORS origin check is
			// done before the current connection ever
			// arrives to this function call
			return true
		},
	}

	logger.Debug("upgraded the connection")

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		// upgrader itself writes http error response
		logger.Error(err)
		return
	}

	cm := client.NewClientMeta(user.uid, user.name, conn)
	logger.Debugf("Created client meta for ID [% X]", cm.GetId())

	err = hubs[roomID].RegisterClient(cm)

	if err != nil {
		logger.Debug("Hub already filled up")
		w.WriteHeader(http.StatusForbidden)
		conn.Close()
		return
	}
	logger.Infof("user [%s] with uuid [%s] joined room [%d]", user.name, user.uid, roomID)
}
