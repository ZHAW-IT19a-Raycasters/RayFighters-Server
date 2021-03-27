package setup

import (
	webconfig "github.zhaw.ch/raycasters/server/web/config"
)

const (
	// Is the path to the default configuration file.
	Config = "config.toml"
)

// Represents a collection of configuration, that is read out of the
// configuration file.
// As per definition if there is a config file in here
// It must implement the CheckConfig interface.
type Configuration struct {
	Web webconfig.WebConfig `toml:"Webconfig"`
}

// checkIfConfigIsLegal will check that the configurated values are legal
func checkIfConfigIsLegal(conf *Configuration) error {
	err := conf.Web.CheckConfiguration()
	if err != nil {
		return &configurationError{section: "Webconfig", internalErr: err}
	}

	return nil
}
