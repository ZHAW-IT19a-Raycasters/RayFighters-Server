package utils

// Will parse the arguments from both the configuration file as well as the commandline arguments
// If there was an error of any kind will print the given help message and return nil.
func Setup() *Configuration {
	// shutting of annoying error
	_, arparser, err := argParse()

	if err != nil {
		arparser.Usage(err)
		return nil
	}

	conf, err := configParse()

	if err != nil {
		arparser.Usage(err)
		return nil
	}

	if err != nil {
		arparser.Usage(err)
		return nil
	}

	return conf
}
