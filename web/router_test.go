package web

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.zhaw.ch/raycasters/server/room"
	"golang.org/x/net/publicsuffix"
)

func TestAPIRoot(t *testing.T) {
	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(serveRoot)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check the response body is what we expect.
	expected := `Hello from RayCasters!`
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}



// TestAuthorizedWebsocketEstablishment is the whole shebang in one test!
func TestAuthorizedWebsocketEstablishment(t *testing.T) {
	privKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	// set up a router for auth and ws
	router := mux.NewRouter()
	router.HandleFunc("/auth",ServeAuthWithKey(privKey))
	router.HandleFunc("/ws/{roomID:[0-9]+}", func(rw http.ResponseWriter, r *http.Request) {serveWs([]*room.Hub{} ,rw, r)})

	// Start our server
	ts := httptest.NewServer(router)
	defer ts.Close()

	u, err := url.Parse(ts.URL)
	if err != nil {
		log.Fatal(err)

	}

	// All users of cookiejar should import "golang.org/x/net/publicsuffix"
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		log.Fatal(err)
	}

	// create a client with our jar
	client := &http.Client{
		Jar: jar,
	}



	go func() {
		server.Start()
	}()
	defer server.ServerStop(context.TODO())


	postBody, _ := json.Marshal(map[string]string{
		"username":  "Phil",
	})

	// 
	if _, err = client.Post(u.String()+"/auth","application/json",bytes.NewBuffer(postBody)); err != nil {
		log.Fatal(err)
	}

	// we should have the auth token as cookie in the jar, mmmmmhhh cookies :)
	fmt.Println("After 1st request:")
	for _, cookie := range jar.Cookies(u) {
		fmt.Printf("  %s: %s\n", cookie.Name, cookie.Value)
	}

	// add our jar to the default dialer
	websocket.DefaultDialer.Jar = jar

	// change url to ws endpoint for same host
	u.Scheme = "ws"
	u.Path = "/ws/1"

	fmt.Printf("connecting to %s\n", u.String())

	c, res, err := websocket.DefaultDialer.Dial(u.String(), nil)
	defer c.Close()

	if err != nil {
		fmt.Print("Error during handshake", err)
	}

	if res.StatusCode != http.StatusSwitchingProtocols {
		t.Errorf("handler returned wrong status code during handshake: got %v want %v",
			res.StatusCode, http.StatusSwitchingProtocols)
	}
}

func TestWSBadRoomID(t *testing.T) {
	req := httptest.NewRequest("GET", "/ws/123", nil)
	rr := httptest.NewRecorder()
	// Need to create a router that we can pass the request through so that the vars will be added to the context
	router := mux.NewRouter()
	router.HandleFunc("/ws/{roomID:[0-9]+}", func(rw http.ResponseWriter, r *http.Request) { serveWs([]*room.Hub{}, rw, r) })
	router.ServeHTTP(rr, req)
	res := rr.Result()
	io.ReadAll(res.Body)
	if res.StatusCode != http.StatusConflict {
		t.Errorf("handler returned wrong status code for bad request: got %v want %v",
		res.StatusCode, http.StatusConflict)
	}
}

func TestWSBadProtocol(t *testing.T) {
	req := httptest.NewRequest("GET", "/ws/1", nil)
	rr := httptest.NewRecorder()
	// Need to create a router that we can pass the request through so that the vars will be added to the context
	router := mux.NewRouter()
	router.HandleFunc("/ws/{roomID:[0-9]+}", func(rw http.ResponseWriter, r *http.Request) { serveWs([]*room.Hub{}, rw, r) })
	router.ServeHTTP(rr, req)

	res := rr.Result()
	body, _ := io.ReadAll(res.Body)
	t.Log(string(body))
	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code for bad request: got %v want %v",
			res.StatusCode, http.StatusBadRequest)
	}
}
