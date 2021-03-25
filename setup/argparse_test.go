package setup

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	webconfig "github.zhaw.ch/raycasters/server/web/config"
)

func helper(t *testing.T, err error) {
	if err != nil {
		t.Errorf("Arguments are supposed to be valid\n%w", err)
	}
}

func helperNil(t *testing.T, err error, msg string) {
	if err == nil {
		t.Errorf("Arguments are supposed to be invalid\n%s", msg)
	}
}

func TestUsageMessage(t *testing.T) {
	os.Args = []string{"progname", "-p", "88", "-i", "localhost"}
	parser, err := argParse()
	helper(t, err)

	// force the parser to write the buffer
	buffer := &bytes.Buffer{}
	parser.out = buffer

	exp := "hallo"
	parser.Usage(errors.New(exp))

	res := buffer.String()

	if !strings.HasPrefix(res, exp) {
		t.Error("doesn't contain the expected value")
	}
}

func TestValidArguments(t *testing.T) {
	// save the old os arguments to the stack and write them back after
	// the test
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"progname", "-p", "88", "-i", "localhost"}
	parser, err := argParse()
	args := parser.Args()

	helper(t, err)

	err = webconfig.CheckIfIpValid(*args.Ip)
	helper(t, err)

	allowed := []int{}
	if !webconfig.CheckIfPortIsAllowed(*args.Port, allowed) {
		helper(t, fmt.Errorf("Portnum [%d]", *args.Port))
	}
}

func TestInvalidPortArgument(t *testing.T) {
	rawArgs := [][]string{
		{"progname", "-p", "666", "-i", "localhos"},
		{"progname", "-p", "666", "-i", "127.0.0.265"},
	}
	for _, rawArg := range rawArgs {

		parser, err := rawArgParse(rawArg)

		if parser.Args() != nil {
			helper(t, fmt.Errorf("On invalid argument args shoould be nil"))
		}

		helperNil(t, err, "Parsing of IP should have failed.")
	}
}

func TestInvalidIPArgument(t *testing.T) {
	rawArgs := [...]string{"progname", "-p", "-88", "-i", "localhost"}
	parser, err := rawArgParse(rawArgs[:])

	if parser.Args() != nil {
		helper(t, fmt.Errorf("On invalid argument args shoould be nil"))
	}
	helperNil(t, err, "Parsing of port should have failed.")
}

func TestInEmptyArgumentList(t *testing.T) {
	rawArgs := [...]string{"progname"}
	parser, err := rawArgParse(rawArgs[:])

	helper(t, err)
	args := parser.Args()

	if *args.Ip != "" {
		helper(t, fmt.Errorf("Ip address should be ''"))
	}

	if *args.Port != -1 {
		helper(t, fmt.Errorf("Default port address should be -1"))
	}
}
