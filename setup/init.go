package setup

import (
	"errors"
	"fmt"
)

// Will write back the parsed commandline parameters to the configuration data.
func writeArgsToConfig(conf *Configuration, args *args) {

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
	args, arparser, err := argParse()

	if err != nil {
		arparser.Usage(err)
		return nil
	}

	conf, err := configParse()

	if err != nil {
		err = fmt.Errorf("%s -> %w", err.Error(), errors.Unwrap(err))
		arparser.Usage(err)
		return nil
	}

	writeArgsToConfig(conf, args)

	err = checkIfConfigIsLegal(conf)

	if err != nil {
		arparser.Usage(err)
		return nil
	}

	return conf
}
