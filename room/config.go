package room

import (
	"errors"
	"fmt"
	"os"

	"github.zhaw.ch/raycasters/server/gamemap"
)

// ErrMapConfigurationNotExist is the error returned, when
// a given path to a configuration file does not exist
type ErrMapConfigurationNotExist struct {
	path string
}

// Error satisfies the error interface
func (err ErrMapConfigurationNotExist) Error() string {
	return fmt.Sprintf("json map configuration file [%s] does not exist.", err.path)
}

// Is satisfies the erros.Is interface
func (err ErrMapConfigurationNotExist) Is(target error) bool {
	_, ok := target.(ErrMapConfigurationNotExist)
	return ok
}

// ErrMapConfigurationInvalid is the error returned, if
// there is a directoty instead of the json file present.
type ErrMapConfigurationInvalid struct {
	path string
}

// Error satisfies the error interface
func (err ErrMapConfigurationInvalid) Error() string {
	return fmt.Sprintf("json map configuration [%s] not a file.", err.path)
}

// Is satisfies the erros.Is interface
func (err ErrMapConfigurationInvalid) Is(target error) bool {
	_, ok := target.(ErrMapConfigurationInvalid)
	return ok
}

// MapConfig is the struct containing
// the path to the map overview
type MapConfig struct {
	JsonPath string
}

// CheckConfiguration satisfies the CheckIfValid
// interface.
func (mc MapConfig) CheckConfiguration() error {
	fileInfo, err := os.Stat(mc.JsonPath)
	if errors.Is(err, os.ErrNotExist) {
		return ErrMapConfigurationNotExist{
			path: mc.JsonPath,
		}
	}
	if fileInfo.IsDir() {
		return ErrMapConfigurationInvalid{
			path: mc.JsonPath,
		}
	}
	return gamemap.IsValid(mc.JsonPath)
}
