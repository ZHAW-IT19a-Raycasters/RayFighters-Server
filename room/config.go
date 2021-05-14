package room

import (
	"errors"
	"os"
)

// TODO: Comment

type MapConfig struct {
	JsonPath string
}

var ErrMapConfigurationNotExist = errors.New("Json Map Configuration does not exist.")
var ErrMapConfigurationInvalid = errors.New("Json Map Configuration not a file.")

func (mc MapConfig) CheckConfiguration() error {
	fileInfo, err := os.Stat(mc.JsonPath)
	if errors.Is(err, os.ErrNotExist) {
		return ErrMapConfigurationNotExist
	}
	if fileInfo.IsDir() {
		return ErrMapConfigurationInvalid
	}
	_, err = readJSON(mc.JsonPath)
	return err
}
