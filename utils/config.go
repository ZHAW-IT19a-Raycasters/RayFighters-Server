package utils

import (
	"fmt"
	"io/ioutil"

	"github.com/pelletier/go-toml"
	"github.zhaw.ch/raycasters/server/web"
)

const (
	config = "config.toml"
)

type configurationError struct {
	msg         string
	internalErr error
}

func (ce *configurationError) Error() string {
	return fmt.Sprintf("during %s of the configuration file happend an error", ce.msg)
}

func (ce *configurationError) Unwrap() error {
	return ce.internalErr
}

type Configuration struct {
	Web web.WebConfig `toml:"Webconfig"`
}

func configParse() (*Configuration, error) {
	data, err := ioutil.ReadFile(config)
	if err != nil {
		return nil, &configurationError{internalErr: err, msg: "reading"}
	}

	conf := Configuration{}

	err = toml.Unmarshal(data, &conf)

	if err != nil {
		return nil, &configurationError{internalErr: err, msg: "parsing"}
	}

	return &conf, nil
}
