package web

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	webconfig "github.zhaw.ch/raycasters/server/web/config"
)

// ZIEL: ein neuer ROUTER
type Server struct {
}

func ServingRoute(conf *webconfig.WebConfig) {
	router := mux.NewRouter()
	router.HandleFunc("/", handleRoot)
	http.ListenAndServe(conf.String(), router)
	//router.HandleFunc("/ws",)
}

func handleRoot(writer http.ResponseWriter, request *http.Request) {
	//ResponseWriter implements io.Writer
	fmt.Fprintf(writer, "Hello world from %s", request.URL.Path[1:]) //ich will alles ausser 0-te element
}
