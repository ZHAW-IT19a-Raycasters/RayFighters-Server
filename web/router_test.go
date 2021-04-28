package web

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
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
	router.HandleFunc("/auth", ServeAuthWithKey(privKey))
	router.HandleFunc("/ws/{roomID:[0-9]+}", func(rw http.ResponseWriter, r *http.Request) { serveWs([]*room.Hub{}, rw, r) })

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

	postBody, _ := json.Marshal(map[string]string{
		"username": "Phil",
	})

	//
	if _, err = client.Post(u.String()+"/auth", "application/json", bytes.NewBuffer(postBody)); err != nil {
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

	if err != nil {
		fmt.Print("Error during handshake", err)
	}
	defer c.Close()

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

func TestWsConnection(t *testing.T) {
	const hubId = 0
	hubs := make([]*room.Hub, 0, 1)
	hubs = append(hubs, room.NewHub(hubId))
	err := hubs[hubId].Run()

	if err != nil {
		t.Error("unable to start up the hub")
		return
	}

	// Need to create a router that we can pass the request through so that the vars will be added to the context
	router := mux.NewRouter()

	router.HandleFunc("/ws/{roomID:[0-9]+}", func(rw http.ResponseWriter, r *http.Request) {
		serveWs(hubs, rw, r)
	})

	// setup the server
	server := httptest.NewServer(router)
	defer server.Close()

	url, err := url.Parse(server.URL)
	if err != nil {
		t.Error("something went wrong with setting up the test server")
		return
	}
	url.Scheme = "ws"
	url.Path = "/ws/1"

	// start websocket client
	ws, resp, err := websocket.DefaultDialer.Dial(url.String(), nil)

	if err != nil {
		t.Errorf("there was an error while starting up the websocket [%s]", err.Error())
		return
	}
	if resp.Status != http.StatusText(http.StatusAccepted) || resp.Status != http.StatusText(http.StatusSwitchingProtocols) {
		t.Error("incorrect status was used")
		return
	}

	// receive the 3 information blocks
	sendInformation := make([][]byte, 0, 3)
	for i := 0; i < 3; i += 1 {
		messageType, data, err := ws.ReadMessage()
		if err != nil {
			t.Errorf("There was an error while reading the message [%s]", err.Error())
			return
		}

		if messageType != websocket.BinaryMessage {
			t.Error("incorrect message type used")
		}

		sendInformation = append(sendInformation, data)
	}

	// TODO: check out that the system works
}
