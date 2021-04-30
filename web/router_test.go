package web

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.zhaw.ch/raycasters/server/room"
	"golang.org/x/net/publicsuffix"
)

func hubHelper(t *testing.T) []*room.Hub {
	hubs := make([]*room.Hub, 0, hubsCount)
	for i := 0; i < hubsCount; i += 1 {
		hub := room.NewHub(i)
		err := hub.Run()
		if err != nil {
			t.Errorf("unable to start up a hub [%s]", err.Error())
			return nil
		}
		t.Cleanup(func() {
			hub.Shutdown(context.TODO())
		})
		hubs = append(hubs, hub)
	}
	return hubs
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

// TestAuthorizedWebsocketEstablishment is the whole shebang in one test!
func TestAuthorizedWebsocketEstablishment(t *testing.T) {
	// set up a router for auth and ws
	router := mux.NewRouter()
	hubs := hubHelper(t)
	if hubs == nil {
		return
	}

	router.HandleFunc("/auth", authMIddleware.ServeAuthWithKey())
	router.HandleFunc("/ws/{roomID:[0-9]+}", func(rw http.ResponseWriter, r *http.Request) {
		serveWs(hubs, rw, r)
	})

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
	t.Log("After 1st request:")
	for _, cookie := range jar.Cookies(u) {
		t.Logf("  %s: %s\n", cookie.Name, cookie.Value)
	}

	// add our jar to the default dialer
	websocketDialer := websocket.DefaultDialer
	websocketDialer.Jar = jar

	// change url to ws endpoint for same host
	u.Scheme = "ws"
	u.Path = "/ws/1"

	t.Logf("connecting to %s\n", u.String())

	c, res, err := websocketDialer.Dial(u.String(), nil)

	if err != nil {
		t.Logf("Error during handshake [%s]", err.Error())
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

	hubs := hubHelper(t)
	if hubs == nil {
		return
	}
	router := mux.NewRouter()
	router.HandleFunc("/ws/{roomID:[0-9]+}", func(rw http.ResponseWriter, r *http.Request) {
		serveWs(hubs, rw, r)
	})
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
	hubs := hubHelper(t)
	if hubs == nil {
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

func TestWsConnection(t *testing.T) {
	const hubId = 0
	hubs := hubHelper(t)

	if hubs == nil {
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
	url.Path = "/ws/0"

	// start websocket client
	// create a header with a default JWT token
	urlPath := url.String()
	const name = "Cookie Smith"
	cookie, _ := authMIddleware.createCookie(&Credentials{
		Username: name,
	})

	// get a JWT token
	rq := make(http.Header)

	rq.Set("Cookie", cookie.String())
	dialer := websocket.DefaultDialer
	dialer.Jar.SetCookies(url, []*http.Cookie{cookie})

	ws, resp, err := dialer.Dial(urlPath, rq)

	if err != nil {
		t.Errorf("there was an error while starting up the websocket [%s]", err.Error())
		return
	}

	if resp.StatusCode != http.StatusSwitchingProtocols {
		t.Errorf("incorrect status was used exp : [%d] got: [%d]", http.StatusAccepted, resp.StatusCode)
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
