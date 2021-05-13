package setup

import (
	"errors"
	"os"
	"path/filepath"

	"github.zhaw.ch/raycasters/server/log"
)

// ErrInavalidParent is the error returned if the
// parent is not a directory.
var ErrInavalidParent = errors.New("invalid file path was used, as the parent is not a directory")

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
	_, err := log.LevelDetect(fc.Level)

	if err != nil {
		return err
	}

	// not checking the file itself because there
	// is no benefit in doing so, as the
	// file might be newly created.

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
