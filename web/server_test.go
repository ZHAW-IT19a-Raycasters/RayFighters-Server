package web

import (
	"testing"

	webconfig "github.zhaw.ch/raycasters/server/web/config"
)


func TestServerInstantiate(t *testing.T) {
	var conf = webconfig.WebConfig{
		Ip: "localhost",
		Port: 8081,
	}

	server := NewServer(&conf)
	if server.privateKey == nil {
		t.Errorf("private key creation failed")
	}
}
