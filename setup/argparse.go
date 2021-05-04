package setup

import (
	"fmt"
	"io"
	"os"

	"github.com/akamensky/argparse"
	webconfig "github.zhaw.ch/raycasters/server/web/config"
)

// Represents the arguments utilized.
type commandlineArguments struct {
	// Will be "" if no ip has been given
	Ip *string
	// Will be -1 if no port has been given
	Port *int
}

// A wrapper for the argument parser
// is only used intern the package.
type argParser struct {
	out    io.Writer
	parser *argparse.Parser
	coms   *commandlineArguments
}

// newArgParser will create a new argument parser
func newArgParser(parser *argparse.Parser, coms *commandlineArguments) argParser {
	return argParser{
		parser: parser,
		out:    os.Stdout,
		coms:   coms,
	}
}

// Will print the usage error message to the commandline.
func (p *argParser) Usage(msg error) {
	fmt.Fprint(p.out, p.parser.Usage(msg))
}

// Args will return the commandlineArguments parsed out.
// Attention, if there was an error this will be nil
func (p *argParser) Args() *commandlineArguments {
	return p.coms
}

// Will parse the command line arguments themselfs.
func argParse() (argParser, error) {
	return rawArgParse(os.Args)
}

// Will parse any agument slice.
// This is split into a sub function for easier testing.
func rawArgParse(arguments []string) (argParser, error) {
	parser := argparse.NewParser("RayServer", "Controlls all the rooms and interactions between the players.")
	// shutdown parser if not fullfilled
	parser.ExitOnHelp(true)

	// setup ip
	ipOpt := argparse.Options{
		Required: false,
		Validate: func(values []string) error {
			for _, rawVal := range values {
				err := webconfig.CheckIfIpValid(rawVal)
				if err != nil {
					return fmt.Errorf("[-p|--port] %w", err)
				}
			}
			return nil
		},
		Help: "Sets the ip address",
	}

	ip := parser.String("i", "ip", &ipOpt)

	// setup port
	portOpt := argparse.Options{
		Required: false,
		Help:     "Sets the port",
		Default:  -1,
		Validate: func(values []string) error {
			for _, rawVal := range values {
				err := webconfig.CheckIfPortValid(rawVal, []int{-1})
				if err != nil {
					return fmt.Errorf("[-i|--ip] %w", err)
				}
			}
			return nil
		},
	}

	port := parser.Int("p", "port", &portOpt)

	err := parser.Parse(arguments)

	if err != nil {
		return newArgParser(parser, nil), err
	}

	return newArgParser(parser, &commandlineArguments{Ip: ip, Port: port}), nil
}
