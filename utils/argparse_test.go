package utils

import (
	"fmt"
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

func TestValidArguments(t *testing.T) {
	rawArgs := [...]string{"progname", "-p", "88", "-i", "localhost"}
	args, _, err := rawArgParse(rawArgs[:])

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

		args, _, err := rawArgParse(rawArg)

		if args != nil {
			helper(t, fmt.Errorf("On invalid argument args shoould be nil"))
		}

		helperNil(t, err, "Parsing of IP should have failed.")
	}
}

func TestInvalidIPArgument(t *testing.T) {
	rawArgs := [...]string{"progname", "-p", "-88", "-i", "localhost"}
	args, _, err := rawArgParse(rawArgs[:])

	if args != nil {
		helper(t, fmt.Errorf("On invalid argument args shoould be nil"))
	}
	helperNil(t, err, "Parsing of port should have failed.")
}
func TestInEmptyArgumentList(t *testing.T) {
	rawArgs := [...]string{"progname"}
	args, _, err := rawArgParse(rawArgs[:])

	helper(t, err)

	if *args.Ip != "" {
		helper(t, fmt.Errorf("Ip address should be ''"))
	}

	if *args.Port != -1 {
		helper(t, fmt.Errorf("Default port address should be -1"))
	}
}
