package setup

import (
	"os"
	"testing"
)

func TestLoggerConfig(t *testing.T) {
	// Attention this is only for coverage until there
	// will be some actuall fields
	lg := LoggerConfig{}
	err := lg.CheckConfiguration()

	if err != nil {
		t.Error("this function should only be a stub")
	}
}

func TestConsoleConfig(t *testing.T) {
	cc := ConsoleConfig{
		Level: "debug",
	}

	err := cc.CheckConfiguration()

	if err != nil {
		t.Error("There should not be an error here as the configuration is valid")
		return
	}

	cc.Level = "Nope"

	err = cc.CheckConfiguration()

	if err == nil {
		t.Error("There should be an error here as the configuration is invalid")
		return
	}
}

func TestFileConfigSuccess(t *testing.T) {
	tmpFile := tmpFileSetup(t)
	fc := FileConfig{
		Level:    "debug",
		FilePath: tmpFile,
	}

	err := fc.CheckConfiguration()

	if err != nil {
		t.Error("There should not be an error here as the configuration is valid")
	}
}

func TestFileConfigIncorrectLevel(t *testing.T) {
	tmpFile := tmpFileSetup(t)
	fc := FileConfig{
		Level:    "debu",
		FilePath: tmpFile,
	}

	err := fc.CheckConfiguration()

	if err == nil {
		t.Error("There should be an error here as the configuration is invalid")
	}
}

func TestFileConfigIsNotDir(t *testing.T) {
	tmpFile := tmpFileSetup(t)
	fc := FileConfig{
		Level:    "debug",
		FilePath: tmpFile + "/file",
	}

	err := fc.CheckConfiguration()

	if err == nil {
		t.Error("There should be an error here as the configuration is invalid")
	}
}

func TestFileConfigDoesNotExist(t *testing.T) {
	tmpFile := tmpFileSetup(t)
	os.Remove(tmpFile)

	fc := FileConfig{
		Level:    "debug",
		FilePath: tmpFile + "/file",
	}

	err := fc.CheckConfiguration()

	if err == nil {
		t.Error("There should be an error here as the configuration is invalid")
	}
}
