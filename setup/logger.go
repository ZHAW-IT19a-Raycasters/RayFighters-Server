package setup

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.zhaw.ch/raycasters/server/log"
)

var (
	// ErrInavalidParent is the error returned if the
	// parent is not a directory.
	ErrInavalidParent = errors.New("invalid file path was used, as the parent is not a directory")

	// ErrEmptyPath is the error returned if the
	// given filepath is empty
	ErrEmptyPath = errors.New("the given path is empty")
)

// LoggerConfig contains the configurations
// information needed to setup the global
// logger.
type LoggerConfig struct {
	Console ConsoleConfig
	File    FileConfig
}

// CheckConfiguration satisfies the CheckConfig
// interface.
func (lc LoggerConfig) CheckConfiguration() error {
	// no checks are needed for this struct
	// as there are no filds.
	return nil
}

// ConsoleConfig is the configuration required
// for printing to the console.
type ConsoleConfig struct {
	Level string
}

// CheckConfiguration satisfies the CheckConfig
// interface.
func (cc ConsoleConfig) CheckConfiguration() error {
	// cleaning up the input
	cc.Level = cleanUpLevelStr(cc.Level)
	// empty line does not have to be checked on specifically
	// as it will be detected by the following function
	_, err := log.LevelDetect(cc.Level)
	return err
}

// FileConfig contains the file logger configuration
type FileConfig struct {
	Level, FilePath string
}

// CheckConfiguration satisfies the CheckConfig
// interface.
func (fc FileConfig) CheckConfiguration() error {
	// cleaning up the input
	fc.Level = cleanUpLevelStr(fc.Level)

	// empty line does not have to be checked on specifically
	// as it will be detected by the following function
	_, err := log.LevelDetect(fc.Level)

	if err != nil {
		return err
	}

	// not checking the file itself because there
	// is no benefit in doing so, as the
	// file might be newly created.
	fc.FilePath = strings.TrimSpace(fc.FilePath)
	if fc.FilePath == "" {

	}

	// check of the parent directory
	parent := filepath.Dir(fc.FilePath)

	// check if parent exists or not
	info, err := os.Stat(parent)

	if errors.Is(err, os.ErrNotExist) {
		return err
	}

	// check if the parent is even a directory or not
	if !info.IsDir() {
		return ErrInavalidParent
	}

	return nil
}

// cleanUpStr will clean up the input variables
func cleanUpLevelStr(s string) string {
	s = strings.TrimSpace(s)
	return strings.ToUpper(s)
}
