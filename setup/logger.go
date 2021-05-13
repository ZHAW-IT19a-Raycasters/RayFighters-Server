package setup

import (
    "github.zhaw.ch/raycasters/server/log"
)

// LoggerConfig
type LoggerConfig struct {
    Console ConsoleConfig
    File    FileConfig
}

func (lc LoggerConfig) CheckConfiguration() error {
    // no checks are needed for this struct 
    // as there are no filds.
    return nil
}


type ConsoleConfig struct {
    level string
}

func (cc ConsoleConfig) CheckConfiguration() error {
    _, err := log.LevelDetect(cc.level)
    return err
}

type FileConfig struct {
    level, filePath string
}

func (fc FileConfig) CheckConfiguration() error {
    _, err := log.LevelDetect(fc.level)

    // not checking the path because there
    // is no benefit to do it...
    return err
}

