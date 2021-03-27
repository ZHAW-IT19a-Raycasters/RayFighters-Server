package setup

import (
	"errors"
	"fmt"
)

// Will write back the parsed commandline parameters to the configuration data.
func writeArgsToConfig(conf *Configuration, args *commandlineArguments) {

	if *args.Ip != "" {
		conf.Web.Ip = *args.Ip
	}

	if *args.Port > -1 {
		conf.Web.Port = uint16(*args.Port)
	}
}

// Will parse the arguments from both the configuration file as well as the commandline arguments
// If there was an error of any kind will print the given help message and return nil.
func Setup() *Configuration {
	// shutting of annoying error
	argParser, err := argParse()

	if err != nil {
		argParser.Usage(err)
		return nil
	}

	confParser := newConfigParser(Config)

	conf, err := confParser.Parse()

	if err != nil {
		err = fmt.Errorf("%s -> %w", err.Error(), errors.Unwrap(err))
		argParser.Usage(err)
		return nil
	}

	writeArgsToConfig(conf, argParser.Args())

	err = checkIfConfigIsLegal(conf)

	if err != nil {
		argParser.Usage(err)
		return nil
	}

	return conf
}
