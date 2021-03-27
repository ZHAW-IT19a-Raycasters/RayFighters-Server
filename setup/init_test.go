package setup

import (
	"bytes"
	"errors"
	"os"
	"strings"
	"testing"
)

func setupEnvironment(args []string, t *testing.T) {
	// setting up the callbacks needed to reset the system after each test
	orgParams := os.Args
	t.Cleanup(func() { os.Args = orgParams })
	os.Args = args
}

func setupParsers(t *testing.T) (*Setup, *argParser, *bytes.Buffer, *configParser, error) {
	setup := NewSetup()
	rawConf := newConfigurationLegal()
	confData, err := marshalConfiguration(rawConf)

	if err != nil {
		t.Errorf("marshalling on control failed\n%s", err.Error())
		return nil, nil, nil, nil, err
	}

	// using the helper functions from the other tests
	// setting up the systems
	aParser, buffer := setupArgParser(t)
	confParser := newConfigParser(fileName)

	// override the read function
	confParser.fileReader = func(name string) ([]byte, error) {
		return confData, nil
	}
	setup.args = func() (argParser, error) { return aParser, nil }

	setup.conf = func(s string) configParser { return confParser }

	return &setup, &aParser, buffer, &confParser, nil
}

func TestSetupSuccess(t *testing.T) {
	setupEnvironment([]string{"progname", "-p", "88", "-i", "localhost"}, t)

	setup, _, _, confParser, err := setupParsers(t)

	if err != nil {
		// error was already logged
		return
	}

	exp, err := confParser.Parse()

	if err != nil {
		t.Error("This should never be able to fail.")
		return
	}

	// fix the expected
	exp.Web.Ip = "localhost"
	exp.Web.Port = 88

	configuration := setup.Configuration()

	compareConfigs(*exp, *configuration, t)
}

func TestSetupArgsFailed(t *testing.T) {
	setupEnvironment([]string{"progname", "-p", "88", "-i", "localhost"}, t)

	setup, aParser, buffer, _, err := setupParsers(t)

	if err != nil {
		// error was already logged
		return
	}

	buffer.WriteByte(' ') // control

	// override the error
	setup.args = func() (argParser, error) { return *aParser, errors.New("Some error to be annoyed about") }

	configuration := setup.Configuration()

	if configuration != nil {
		t.Error("This should not be possible")
	}

	// check the line sent to the buffer
	res := buffer.String()

	if res == "" || res == " " {
		t.Error("usage string is empty")
	}
	if !strings.Contains(res, "usage") {
		t.Error("usage string has not been shown")
	}
}

func TestSetupConfFailed(t *testing.T) {
	setupEnvironment([]string{"progname", "-p", "88", "-i", "localhost"}, t)

	setup, _, buffer, confParser, err := setupParsers(t)

	if err != nil {
		// error was already logged
		return
	}

	buffer.WriteByte(' ') // control

	// override the reader to return garbage
	rawConf := newConfigurationLegal()
	rawConf.Web.Ip = "localho"
	confData, err := marshalConfiguration(rawConf)

	if err != nil {
		t.Errorf("marshalling failed\n%s", err.Error())
		return
	}

	confParser.fileReader = func(name string) ([]byte, error) {
		return confData, nil
	}

	configuration := setup.Configuration()

	if configuration != nil {
		t.Error("This should not be possible")
		return
	}

	// check the line sent to the buffer
	res := buffer.String()

	if res == "" || res == " " {
		t.Error("usage string is empty")
		return
	}

	if !strings.Contains(res, "->") {
		t.Error("error string has not been shown")
		return
	}

	if !strings.Contains(res, "usage") {
		t.Error("usage string has not been shown")
		return
	}
}
