package web

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwt"
)


func TestAuthTokenIssueing(t *testing.T) {

	// generate key
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Errorf("failed to generate private key: %s\n", err)
		return
	}
	
	var payload []byte
	// Create signed payload
	token := jwt.New()
	token.Set(`foo`, `bar`)
	payload, err = jwt.Sign(token, jwa.RS256, privKey)
	if err != nil {
		t.Errorf("failed to generate signed payload: %s\n", err)
		return
	}


	{ // Parse signed payload, and perform (1) verification of the signature
		// and (2) validation of the JWT token
		// Validation can be performed in a separate step using `jwt.Validate`
		token, err := jwt.Parse(
			payload,
			jwt.WithValidate(true),
			jwt.WithVerify(jwa.RS256, &privKey.PublicKey),
		)
		if err != nil {
			t.Errorf("failed to parse JWT token: %s\n", err)
			return
		}
		buf, err := json.MarshalIndent(token, "", "  ")
		if err != nil {
			t.Errorf("failed to generate JSON: %s\n", err)
			return
		}
		fmt.Printf("%s\n", buf)
	}

	// 
}

func TestAuthRequestWithValidJson(t *testing.T) {
	postBody, _ := json.Marshal(map[string]string{
		"username":  "Phil",
	})
	req := httptest.NewRequest("POST", "/auth", bytes.NewBuffer(postBody))
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	// #TODO: generate key
    router.HandleFunc("/auth", ServeAuthWithKey(nil))
	router.ServeHTTP(rr, req)
	res := rr.Result()
	if res.StatusCode != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
		res.StatusCode, http.StatusOK)
	}
}


func TestAuthTokenResponseVaild(t *testing.T) {
	postBody, _ := json.Marshal(map[string]string{
		"username":  "Phil",
	})
	req := httptest.NewRequest("POST", "/auth", bytes.NewBuffer(postBody))
	rr := httptest.NewRecorder()

	// generate private key
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		fmt.Printf("failed to generate private key: %s\n", err)
	}

	router := mux.NewRouter()
	// #TODO: generate key
    router.HandleFunc("/auth", ServeAuthWithKey(privKey))
	router.ServeHTTP(rr, req)
	res := rr.Result()
	if res.StatusCode != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
		res.StatusCode, http.StatusOK)
	}
}