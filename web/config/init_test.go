package webconfig

import (
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
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

func TestCheckJwtConfNoRSA(t *testing.T) {
	conf := JwtConfig{
		RsaPrivateKey: "",
	}

	err := conf.CheckConfiguration()
	if err != nil {
		t.Error("there should not be an error here, as the example is valid")
	}
}
func TestFixLinuxPaths(t *testing.T) {
	const hiddenDirName = ".ssh"
	const rawfileName = "id_rsa"
	const file = "~/.ssh/id_rsa"

	tmpDir := tmpDirSetup(t)
	dirName := filepath.Join(tmpDir, hiddenDirName)
	// create the sub dir
	err := os.Mkdir(dirName, 0777)

	if err != nil {
		t.Errorf("error happend while creating the custom directory, [%s]", err.Error())
		return
	}

	fileName := filepath.Join(dirName, rawfileName)
	err = os.WriteFile(fileName, []byte(RsaPrivateKey), 0666)

	if err != nil {
		t.Errorf("error happend while writting to the rsa file [%s]", err.Error())
		return
	}

	orgGetCurrUser := getCurrentUser

	t.Cleanup(func() {
		getCurrentUser = orgGetCurrUser
	})

	getCurrentUser = func() (*user.User, error) {
		us := &user.User{
			HomeDir: tmpDir,
		}
		return us, nil
	}

	conf := JwtConfig{
		RsaPrivateKey: file,
	}

	err = conf.CheckConfiguration()
	if err != nil {
		t.Errorf("there should not be an error here, as the example is valid -> [%s]", err.Error())
	}
}

func TestFixLinuxPathsInvalidPath(t *testing.T) {
	// Attention this test is only for coverage purposes
	const hiddenDirName = ".ssh"
	const rawfileName = "id_rsa"
	const file = "~/.ssh/id_rsa"

	tmpDir := tmpDirSetup(t)
	dirName := filepath.Join(tmpDir, hiddenDirName)
	// create the sub dir
	err := os.Mkdir(dirName, 0777)

	if err != nil {
		t.Errorf("error happend while creating the custom directory, [%s]", err.Error())
		return
	}

	fileName := filepath.Join(dirName, rawfileName)
	err = os.WriteFile(fileName, []byte(RsaPrivateKey), 0666)

	if err != nil {
		t.Errorf("error happend while writting to the rsa file [%s]", err.Error())
		return
	}

	orgGetCurrUser := getCurrentUser

	t.Cleanup(func() {
		getCurrentUser = orgGetCurrUser
	})

	someAnnoyingError := fmt.Errorf("Some unbelivable error")

	getCurrentUser = func() (*user.User, error) {
		return nil, someAnnoyingError
	}

	conf := JwtConfig{
		RsaPrivateKey: file,
	}

	err = conf.CheckConfiguration()
	if err == nil {
		t.Errorf("there should be an error here, as the example is invalid")
	} else if !errors.Is(err, ErrUnableToParseUserHome{}) {
		t.Error("incorrect error type used")
	} else if !errors.Is(err, someAnnoyingError) {
		t.Error("the internal error is different then it should be")
	} else if !strings.Contains(err.Error(), "unable to parse") {
		t.Error("wrong error string")
	}
}

func TestCORSConfigValid(t *testing.T) {
	cors := []string{
		// this system
		"http://127.0.0.1:8080",
		"http://localhost:8080",
		// frontend
		"http://127.0.0.1:9000",
		"http://localhost:9000",
	}

	conf := CORSConfig{
		Domains: cors,
	}

	err := conf.CheckConfiguration()
	if err != nil {
		t.Error("there should not be an error on these valid domains")
	}
}

func TestCORSConfigEmptry(t *testing.T) {
	cors := []string{}

	conf := CORSConfig{
		Domains: cors,
	}

	err := conf.CheckConfiguration()
	if err == nil {
		t.Error("there should be an error on these invalid domains")
	}
}
func TestCORSConfigInvalid(t *testing.T) {
	cors := []string{
		"ftp://127.0.0.1:8080",
		"http://localhost:8080/Path",
	}

	conf := CORSConfig{
		Domains: cors,
	}

	err := conf.CheckConfiguration()
	if err == nil {
		t.Error("there should be an error on these invalid domains")
	}
}
