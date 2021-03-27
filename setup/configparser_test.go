package setup

import (
	"testing"
)

const fileName = "conf.toml"

func newConfigurationLegal() []byte {
	line := `[Webconfig]
	ip = "127.0.0.1"
	port = 44
	`
	return []byte(line)
}

func TestParseConfigLegal(t *testing.T) {
	rawConf := newConfigurationLegal()
	confParser := newConfigParser(fileName)

	// override the read function
	confParser.fileReader = func(name string) ([]byte, error) {
		return rawConf, nil
	}

	conf, err := confParser.Parse()

	if err != nil {
		t.Errorf("There should not be an error on this configuration file.\n%s", err.Error())
		return
	} else if conf == nil {
		t.Error("There should be a valid configuration file")
		return
	}
}
