package utils

import (
	"os"

	"github.com/akamensky/argparse"
)

type Args struct {
	Ip   *string
	Port *int
}

type ArgParser struct {
	parser *argparse.Parser
}

func (p *ArgParser) Help(msg string) {
	p.parser.Help(msg)
}

func ArgParse() (*Args, *ArgParser, error) {
	return internalArgParse(os.Args)
}

// is split into a sub function for testing puposes
func internalArgParse(args []string) (*Args, *ArgParser, error) {
	parser := argparse.NewParser("RayServer", "Controlls all the rooms and interactions between the players.")
	// shutdown parser if not fullfilled
	parser.ExitOnHelp(true)

	// setup ip
	ipOpt := argparse.Options{
		Required: false,
		Help:     "Sets the ip address",
	}

	ip := parser.String("i", "ip", &ipOpt)

	// setup port
	portOpt := argparse.Options{
		Required: false,
		Help:     "Sets the port",
	}

	port := parser.Int("p", "port", &portOpt)

	err := parser.Parse(args)

	if err != nil {
		return nil, &ArgParser{parser}, err
	}

	return &Args{Ip: ip, Port: port}, &ArgParser{parser}, nil
}
