package setup

import (
	"fmt"
	"io/ioutil"

	"github.com/pelletier/go-toml"
	webconfig "github.zhaw.ch/raycasters/server/web/config"
)

const (
	// Is the path to the default configuration file.
	Config = "config.toml"
)

// Is used to force this implementation, by using the internal
// values inside of checkifConfigIsLegal
type CheckConfig interface {
	// Will check if configuration is correct
	CheckConfiguration() error
}

// Represents a configuration error
type configurationError struct {
	section     string
	internalErr error
}

// The error interface
func (ce *configurationError) Error() string {
	return fmt.Sprintf("during parsing of the configuration file happend an error in the section:\n\t[%s]\n\t[%s]", ce.section, ce.internalErr)
}

// The error unwrap interface
func (ce *configurationError) Unwrap() error {
	return ce.internalErr
}

// Represents a collection of configuration, that is read out of the
// configuration file.
// As per definition if there is a config file in here
// It must implement the CheckConfig interface.
type Configuration struct {
	Web webconfig.WebConfig `toml:"Webconfig"`
}

// parses the configuration file and reads the data into a new configuration obj.
func configParse() (*Configuration, error) {
	data, err := ioutil.ReadFile(Config)
	if err != nil {
		return nil, &configurationError{internalErr: err, section: "reading"}
	}

	conf := Configuration{}

	err = toml.Unmarshal(data, &conf)

	if err != nil {
		return nil, &configurationError{internalErr: err, section: "parsing"}
	}

	return &conf, nil
}

// Will check that the configurated values are legal
func checkIfConfigIsLegal(conf *Configuration) error {
	err := conf.Web.CheckConfiguration()
	if err != nil {
		return &configurationError{section: "Webconfig", internalErr: err}
	}

	return nil
}
