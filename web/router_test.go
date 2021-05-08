package web

import (
	"bytes"
	"context"
	"encoding/json"
	"github.zhaw.ch/raycasters/server/room/controller"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.zhaw.ch/raycasters/server/gamemap"
	"github.zhaw.ch/raycasters/server/protocol/client"
	oServer "github.zhaw.ch/raycasters/server/protocol/server"
	"github.zhaw.ch/raycasters/server/utils"
	"golang.org/x/net/publicsuffix"
)

func hubHelper(t *testing.T) ([]*controller.Hub, error) {
	hubs := make([]*controller.Hub, 0, hubsCount)
	for i := 0; i < hubsCount; i += 1 {
		hub := controller.NewHub(i)
		err := hub.Run()
		if err != nil {
			t.Errorf("unable to start up a hub [%s]", err.Error())
			return nil, err
		}
		t.Cleanup(func() {
			hub.Shutdown(context.TODO())
		})
		hubs = append(hubs, hub)
	}
	return hubs, nil
}

func setupDialer(url *url.URL, t *testing.T) *websocket.Dialer {

	// All users of cookiejar should import "golang.org/x/net/publicsuffix"
	jar, err := cookiejar.New(&cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	})
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
	url.Path = "/auth"
	if _, err = client.Post(url.String(), "application/json", bytes.NewBuffer(postBody)); err != nil {
		log.Fatal(err)
	}

	// we should have the auth token as cookie in the jar, mmmmmhhh cookies :)
	// t.Log("After 1st request:")
	// for _, cookie := range jar.Cookies(url) {
	// 	t.Logf("  %s: %s\n", cookie.Name, cookie.Value)
	// }

	// add our jar to the default dialer
	return &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 45 * time.Second,
		Jar:              jar,
	}
}

func setupTestServer(t *testing.T) ([]*controller.Hub, *httptest.Server, error) {
	hubs, err := hubHelper(t)

	if hubs == nil {
		return nil, nil, err
	}
	// Need to create a router that we can pass the request through so that the vars will be added to the context
	router := mux.NewRouter()

	auth, _ := newAuthManager("")

	router.HandleFunc("/auth", auth.ServeAuthWithKey())
	router.HandleFunc("/ws/{roomID:[0-9]+}", func(rw http.ResponseWriter, r *http.Request) {
		serveWs(hubs, rw, r)
	})

	// setup the server
	server := httptest.NewServer(router)
	t.Cleanup(server.Close)
	return hubs, server, nil
}

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

func TestWSBadRoomID(t *testing.T) {
	req := httptest.NewRequest("GET", "/ws/123", nil)
	rr := httptest.NewRecorder()
	// Need to create a router that we can pass the request through so that the vars will be added to the context

	hubs, err := hubHelper(t)
	if err != nil {
		return
	}
	router := mux.NewRouter()
	router.HandleFunc("/ws/{roomID:[0-9]+}", func(rw http.ResponseWriter, r *http.Request) {
		serveWs(hubs, rw, r)
	})
	router.ServeHTTP(rr, req)
	res := rr.Result()
	io.ReadAll(res.Body)
	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code for bad request: got %v want %v",
			res.StatusCode, http.StatusConflict)
	}
}

// TestAuthorizedWebsocketEstablishment is the whole shebang in one test!
func TestAuthorizedWebsocketEstablishment(t *testing.T) {
	_, server, err := setupTestServer(t)
	if err != nil {
		return
	}
	u, err := url.Parse(server.URL)
	if err != nil {
		log.Fatal(err)
	}

	dialer := setupDialer(u, t)

	// change url to ws endpoint for same host
	u.Scheme = "ws"
	u.Path = "/ws/1"

	t.Logf("connecting to %s\n", u.String())

	c, res, err := dialer.Dial(u.String(), nil)

	if err != nil {
		t.Logf("Error during handshake [%s]", err.Error())
	}
	defer c.Close()

	if res.StatusCode != http.StatusSwitchingProtocols {
		t.Errorf("handler returned wrong status code during handshake: got %v want %v",
			res.StatusCode, http.StatusSwitchingProtocols)
	}
}

func TestWSBadProtocol(t *testing.T) {
	req := httptest.NewRequest("GET", "/ws/1", nil)
	rr := httptest.NewRecorder()
	// Need to create a router that we can pass the request through so that the vars will be added to the context
	hubs, err := hubHelper(t)
	if err != nil {
		return
	}
	router := mux.NewRouter()
	router.HandleFunc("/ws/{roomID:[0-9]+}", func(rw http.ResponseWriter, r *http.Request) {
		serveWs(hubs, rw, r)
	})
	router.ServeHTTP(rr, req)

	res := rr.Result()
	body, _ := io.ReadAll(res.Body)

	sbody := string(body)

	if res.StatusCode != http.StatusForbidden {
		t.Errorf("handler returned wrong status code for bad request: got %v want %v",
			res.StatusCode, http.StatusForbidden)
	} else if !strings.Contains(sbody, "no JWT found in Cookies") {
		t.Error("incorrect error body returned")
	}
}

func TestWsFirstOpcodes(t *testing.T) {
	_, server, err := setupTestServer(t)
	if err != nil {
		return
	}

	url, err := url.Parse(server.URL)
	if err != nil {
		t.Error("something went wrong with setting up the test server")
		return
	}

	dialer := setupDialer(url, t)

	url.Scheme = "ws"
	url.Path = "/ws/0"
	urlPath := url.String()

	t.Logf("connecting to %s\n", url.String())

	ws, resp, err := dialer.Dial(urlPath, nil)

	if resp.StatusCode != http.StatusSwitchingProtocols {
		t.Errorf("incorrect status was used exp : [%d] got: [%d]", http.StatusAccepted, resp.StatusCode)
		return
	}

	if err != nil {
		t.Errorf("there was an error while starting up the websocket [%s]", err.Error())
		return
	}

	// receive the 4 spawn information blocks
	sendInformation := make([][]byte, 0, 4)
	for i := 0; i < 4; i += 1 {
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

	expOpcodes := map[byte]struct{}{
		oServer.YouAreCalledOpcode.Value():      {},
		oServer.SendAllPlayerInfoOpcode.Value(): {},
		oServer.SendAllPlayerKDOpcode.Value():   {},
		oServer.TeleportPlayerOpcode.Value():    {},
	}

	// check that the correct opcodes, where received
	for _, infoSend := range sendInformation {
		opcode := infoSend[0]
		_, ok := expOpcodes[opcode]
		if !ok {
			t.Errorf("Wrong opcode send [0x%X]", opcode)
			return
		}
	}
}

func TestWsMovement(t *testing.T) {
	_, server, err := setupTestServer(t)
	if err != nil {
		return
	}

	url, err := url.Parse(server.URL)
	if err != nil {
		t.Error("something went wrong with setting up the test server")
		return
	}

	dialer := setupDialer(url, t)

	url.Scheme = "ws"
	url.Path = "/ws/0"
	urlPath := url.String()

	t.Logf("connecting to %s\n", url.String())

	ws, resp, err := dialer.Dial(urlPath, nil)

	if resp.StatusCode != http.StatusSwitchingProtocols {
		t.Errorf("incorrect status was used exp : [%d] got: [%d]", http.StatusAccepted, resp.StatusCode)
		return
	}

	if err != nil {
		t.Errorf("there was an error while starting up the websocket [%s]", err.Error())
		return
	}

	// receive the 4 spawn information blocks
	for i := 0; i < 4; i += 1 {
		_, _, err := ws.ReadMessage()
		if err != nil {
			t.Errorf("There was an error while reading the message [%s]", err.Error())
			return
		}
	}

	// send the move message
	coord := gamemap.NewCoordinate(0.69, 1.111, 6.9)
	data := [3*8 + 1]byte{
		client.MovementOpcode.Value(),
	}
	for i, val := range [...]float64{coord.X(), coord.Y(), coord.Y()} {
		utils.FloatToBinary(val, data[1+i*8:1+i*8+8])
	}

	ws.WriteMessage(websocket.BinaryMessage, data[:])
}
