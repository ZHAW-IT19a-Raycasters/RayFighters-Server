package webconfig

import (
	"os"
	"testing"
)

func TestWebconfAddr(t *testing.T) {
	exp := "127.0.0.1:666"
	wc := WebConfig{
		Ip:   "127.0.0.1",
		Port: 666,
	}

	got := wc.Address()

	if exp != got {
		t.Error("Address string should be the same")
	}
}

func TestCheckConfigWebConfValid(t *testing.T) {
	wc := WebConfig{
		Ip:   "127.0.0.1",
		Port: 666,
	}
	err := wc.CheckConfiguration()
	if err != nil {
		t.Error("There should not be an error here")
	}
}
func TestCheckConfigWebConfInvalidIP(t *testing.T) {
	wc := WebConfig{
		Ip:   "1270.0.1",
		Port: 666,
	}
	err := wc.CheckConfiguration()
	if err == nil {
		t.Error("There should be an error here")
	}
}

func TestCheckConfigWebConfInvalidPort(t *testing.T) {
	wc := WebConfig{
		Ip:   "127.0.0.1",
		Port: -1,
	}
	err := wc.CheckConfiguration()
	if err == nil {
		t.Error("There should be an error here")
	}
}

func TestCheckJwtConf(t *testing.T) {
	file := tmpFileSetup(t)
	err := os.WriteFile(file, []byte(RsaPrivateKey), 0666)
	if err != nil {
		t.Errorf("There was an error while writing the RSA data -> [%s]", err.Error())
		return
	}

	conf := JwtConfig{
		RsaPrivateKey: file,
	}

	err = conf.CheckConfiguration()
	if err != nil {
		t.Error("there should not be an error here, as the example is valid")
	}
}
