package setup

import (
	"fmt"
	"os"

	"github.com/pelletier/go-toml"
)

// Is used to force this implementation, by using the internal
// values inside of checkifConfigIsLegal
type CheckConfig interface {
	// Will check if configuration is correct
	CheckConfiguration() error
}

// Represents a configuration error
type ErrConfiguration struct {
	section     string
	internalErr error
}

// The error interface
func (ce ErrConfiguration) Error() string {
	return fmt.Sprintf("during parsing of the configuration file happend an error in the section:\n\t[%s]\n\t[%s]", ce.section, ce.internalErr)
}

// The error unwrap interface
func (ce ErrConfiguration) Unwrap() error {
	return ce.internalErr
}

// The is configuration controller helps to check if the
// given error belongs to the written error type.
func (ce ErrConfiguration) Is(target error) bool {
	switch target.(type) {
	case ErrConfiguration:
		return true
	default:
		return false
	}
}

// configParser is contains the values responsible for
// parsing the configuration file.
type configParser struct {
	fileName   string
	fileReader func(name string) ([]byte, error)
}

// newConfigParser will create a new configuration parser
func newConfigParser(fileName string) configParser {
	return configParser{
		fileName:   fileName,
		fileReader: os.ReadFile,
	}
}

// Parse will as the name implies parse and create
// the configuration file.
func (cp configParser) Parse() (*Configuration, error) {
	data, err := cp.fileReader(cp.fileName)

	if err != nil {
		return nil, ErrConfiguration{internalErr: err, section: "reading"}
	}
	// create the configuration object
	conf := &Configuration{}

	// do the unmarshaling
	err = toml.Unmarshal(data, conf)

	if err != nil {
		return nil, ErrConfiguration{internalErr: err, section: "parsing"}
	}

	// do a primary check if the configuration is legal.
	err = CheckIfConfigIsLegal(conf)

	if err != nil {
		return nil, err
	}

	return conf, nil
}
