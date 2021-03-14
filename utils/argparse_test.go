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

func TestValidArgumentList(t *testing.T) {
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

func TestInValidArgumentList(t *testing.T) {
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
