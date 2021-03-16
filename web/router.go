package web

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	webconfig "github.zhaw.ch/raycasters/server/web/config"
)

// ZIEL: ein neuer ROUTER
type Server struct {
	endpoint *http.Server
}

func NewServer(conf *webconfig.WebConfig) *Server {
	router := mux.NewRouter()
	router.HandleFunc("/", handleRoot)
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
