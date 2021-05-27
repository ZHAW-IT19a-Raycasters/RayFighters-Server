// Package setup is responsible for setting up
// system with the parameters gathers from outside.
package setup

// Setup is the struct that handles all the setup required
// by the system.
type Setup struct {
	args     func() (argParser, error)
	conf     func(filename string) configParser
	confFile string
}

// NewSetup will as the name implies create a new setup
// procedure.
func NewSetup() Setup {
	return Setup{
		args:     argParse,
		conf:     newConfigParser,
		confFile: Config,
	}
}

// writeArgsToConfig will write back the parsed command line parameters to the configuration data.
func writeArgsToConfig(conf *Configuration, args *commandlineArguments) {

	if *args.Ip != "" {
		conf.Web.Ip = *args.Ip
	}

	if *args.Port > -1 {
		conf.Web.Port = *args.Port
	}
}

// Configuration will parse the arguments from both the configuration file as well as
// the command line arguments. If there was an error of any kind will print the given
// help message and return nil.
func (s Setup) Configuration() *Configuration {
	argParser, err := s.args()

	if err != nil {
		argParser.Usage(err)
		return nil
	}

	confParser := s.conf(s.confFile)

	conf, err := confParser.Parse()

	if err != nil {
		argParser.Usage(err)
		return nil
	}

	writeArgsToConfig(conf, argParser.Args())

	return conf
}
