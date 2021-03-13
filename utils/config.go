package utils

import (
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/pelletier/go-toml"
	"github.zhaw.ch/raycasters/server/web"
)

const (
	config = "config.toml"
)

type Configuration struct {
	Web web.WebConfig `toml:"Webconfig"`
}

func configParse() (*Configuration, error) {
	data, err := ioutil.ReadFile(config)
	if err != nil {
		res := errors.Unwrap(fmt.Errorf("During reading of the configuration file happend following error:\n%w", err))
		return nil, res
	}

	conf := Configuration{}

	toml.Unmarshal(data, &conf)

	return &conf, nil
}
