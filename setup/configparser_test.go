package setup

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/pelletier/go-toml"
	webconfig "github.zhaw.ch/raycasters/server/web/config"
)

const fileName = "conf.toml"

func compareConfigs(exp, got Configuration, t *testing.T) bool {
	if got.Web.Ip != exp.Web.Ip {
		t.Error("The parsed ip is incorrect")
		return false
	}

	if got.Web.Port != exp.Web.Port {
		t.Error("The parsed port is incorrect")
		return false
	}

	return true
}

func newConfigurationLegal() Configuration {
	return Configuration{
		Web: webconfig.WebConfig{
			Ip:   "127.0.0.1",
			Port: 44,
			JWTConfig: webconfig.JwtConfig{
				RsaPrivateKey: ".",
			},
		},
	}
}

func marshalConfiguration(conf Configuration) ([]byte, error) {
	return toml.Marshal(conf)
}

func TestParseConfigLegal(t *testing.T) {
	rawConf := newConfigurationLegal()

	confData, err := marshalConfiguration(rawConf)

	if err != nil {
		t.Errorf("marshalling on control failed\n%s", err.Error())
		return
	}

	confParser := newConfigParser(fileName)

	// override the read function
	confParser.fileReader = func(name string) ([]byte, error) {
		return confData, nil
	}

	conf, err := confParser.Parse()

	if err != nil {
		t.Errorf("There should not be an error on this configuration file.\n%s", err.Error())
		return
	} else if conf == nil {
		t.Error("There should be a valid configuration file")
		return
	}

	compareConfigs(rawConf, *conf, t)
}

func TestParseConfigOsErr(t *testing.T) {
	// Any of these errors are unlikely as per definition
	// they only serve to get a higher coverage
	confParser := newConfigParser(fileName)

	// override the read function
	confParser.fileReader = func(name string) ([]byte, error) {
		return nil, os.ErrNotExist
	}

	conf, err := confParser.Parse()

	if err == nil {
		t.Error("There should be an OS Error here")
		return
	}

	// as configurationError is a error wrapper, this is important
	if !errors.Is(err, os.ErrNotExist) {
		// this is an impossible case
		t.Error("Wrong error given")
		return
	}

	if !strings.HasPrefix(err.Error(), "during parsing of the configuration file happend") {
		t.Error("incorrect prefix used for error")
		return
	}

	if conf != nil {
		// this is an impossible case
		t.Error("a configuration was returned")
		return
	}
}

func TestParserIncorrectConfig(t *testing.T) {
	rawConf := newConfigurationLegal()
	// using an illigal ip
	rawConf.Web.Ip = "localho"

	confData, err := marshalConfiguration(rawConf)

	if err != nil {
		t.Errorf("marshalling failed\n%s", err.Error())
		return
	}

	confParser := newConfigParser(fileName)

	// override the read function
	confParser.fileReader = func(name string) ([]byte, error) {
		return confData, nil
	}
	conf, err := confParser.Parse()

	if conf != nil {
		t.Error("there should not be a configuration returned here.")
		return
	}

	if err == nil {
		t.Error("an error should have been used here")
		return
	}

	if !errors.Is(err, ErrConfiguration{}) {
		t.Error("incorrect error has been used")
		return
	}
}

func TestParserIlligalConfig(t *testing.T) {

	confData := []byte("a random string")
	confParser := newConfigParser(fileName)

	// override the read function
	confParser.fileReader = func(name string) ([]byte, error) {
		return confData, nil
	}

	conf, err := confParser.Parse()

	if conf != nil {
		t.Error("there should not be a configuration returned here.")
		return
	}

	if err == nil {
		t.Error("an error should have been used here")
		return
	}

	if !errors.Is(err, ErrConfiguration{}) {
		t.Error("incorrect error has been used")
		return
	}
}
