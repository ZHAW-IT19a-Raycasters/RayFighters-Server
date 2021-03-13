package utils

import (
	"fmt"
	"os"

	"github.com/akamensky/argparse"
)

type args struct {
	// Will be "" if no ip has been given
	Ip *string
	// Will be -1 if no port has been given
	Port *int
}

type argParser struct {
	parser *argparse.Parser
}

func (p *argParser) Usage(msg interface{}) {
	fmt.Print(p.parser.Usage(msg))
}

func argParse() (*args, *argParser, error) {
	return rawArgParse(os.Args)
}

// is split into a sub function for testing puposes
func rawArgParse(arguments []string) (*args, *argParser, error) {
	parser := argparse.NewParser("RayServer", "Controlls all the rooms and interactions between the players.")
	// shutdown parser if not fullfilled
	parser.ExitOnHelp(true)

	// setup ip
	ipOpt := argparse.Options{
		Required: false,
		Validate: func(values []string) error {
			for _, rawVal := range values {
				err := checkIfIpValid(rawVal)
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
				err := checkIfPortValid(rawVal, []int{-1})
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
		return nil, &argParser{parser}, err
	}

	return &args{Ip: ip, Port: port}, &argParser{parser}, nil
}
