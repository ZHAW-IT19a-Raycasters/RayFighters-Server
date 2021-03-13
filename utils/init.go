package utils

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
		arparser.Usage(err)
		return nil
	}

	if err != nil {
		return nil
	}

	writeArgsToConfig(conf, args)

	return conf
}
