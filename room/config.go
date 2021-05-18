package room

import (
	"errors"
	"github.zhaw.ch/raycasters/server/gamemap"
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
	_, err = gamemap.readJSON(mc.JsonPath)
	return err
}
